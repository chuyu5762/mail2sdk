// Package mail2sdk 提供 Mail2 临时邮箱系统的 Go SDK
//
// 这是一个单文件 SDK，用户只需复制此文件到项目中即可使用。
//
// 功能特性:
//   - 创建临时邮箱（支持 3 种模式 + 自动混用）
//   - 获取可用域名列表
//   - 指定域名或域名组创建邮箱
//   - 获取邮件列表
//   - 获取邮件详情（完整内容，支持用户自定义正则）
//   - 提取验证码（API 内置）
//   - 删除邮箱
//
// 使用示例:
//   mailbox, _ := mail2sdk.CreateMailbox(baseURL, apiKey, 1, "")
//   mails, _ := mail2sdk.GetMails(baseURL, apiKey, mailbox.Address)
//   code, _ := mail2sdk.ExtractCode(baseURL, apiKey, mailbox.Address, 5)
package mail2sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

// 版本信息
const Version = "2.0.0"

// 全局随机数生成器和域名选择器（线程安全）
var (
	rng            *rand.Rand
	rngOnce        sync.Once
	domainSelector *DomainSelector
	selectorOnce   sync.Once
)

// DomainSelector 域名选择器 - 使用轮询策略确保所有域名均匀使用
type DomainSelector struct {
	mu      sync.Mutex
	counters map[string]int // 每个域名的使用计数
}

// getRand 获取线程安全的随机数生成器
func getRand() *rand.Rand {
	rngOnce.Do(func() {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	})
	return rng
}

// getDomainSelector 获取全局域名选择器
func getDomainSelector() *DomainSelector {
	selectorOnce.Do(func() {
		domainSelector = &DomainSelector{
			counters: make(map[string]int),
		}
	})
	return domainSelector
}

// selectDomain 使用轮询策略选择域名（确保所有域名均匀使用）
//
// 策略：选择使用次数最少的域名，如果有多个最少使用的域名则随机选择一个
func (ds *DomainSelector) selectDomain(domains []string) string {
	if len(domains) == 0 {
		return ""
	}
	if len(domains) == 1 {
		return domains[0]
	}

	ds.mu.Lock()
	defer ds.mu.Unlock()

	// 初始化计数器（如果是新域名）
	for _, domain := range domains {
		if _, exists := ds.counters[domain]; !exists {
			ds.counters[domain] = 0
		}
	}

	// 找出使用次数最少的域名
	minCount := -1
	var candidates []string

	for _, domain := range domains {
		count := ds.counters[domain]
		if minCount == -1 || count < minCount {
			minCount = count
			candidates = []string{domain}
		} else if count == minCount {
			candidates = append(candidates, domain)
		}
	}

	// 从候选域名中随机选择一个
	selected := candidates[getRand().Intn(len(candidates))]

	// 增加使用计数
	ds.counters[selected]++

	return selected
}

// resetCounter 重置指定域名的计数（可选功能）
func (ds *DomainSelector) resetCounter(domain string) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	delete(ds.counters, domain)
}

// getStats 获取域名使用统计（内部使用）
func (ds *DomainSelector) getStats() map[string]int {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	stats := make(map[string]int)
	for k, v := range ds.counters {
		stats[k] = v
	}
	return stats
}

// GetDomainStats 获取域名使用统计（导出函数）
//
// 返回每个域名的使用次数，用于验证轮询策略的有效性
//
// 示例:
//   stats := mail2sdk.GetDomainStats()
//   for domain, count := range stats {
//       fmt.Printf("%s: %d 次\n", domain, count)
//   }
func GetDomainStats() map[string]int {
	return getDomainSelector().getStats()
}

// ResetDomainStats 重置所有域名的使用计数（导出函数）
//
// 用于清空计数器，重新开始计数
func ResetDomainStats() {
	ds := getDomainSelector()
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.counters = make(map[string]int)
}

// 邮箱生成模式常量
const (
	ModeAuto    = 0 // 自动混用（SDK 随机选择 random/chinese/english）
	ModeRandom  = 1 // 随机字符（如: bd4232）
	ModeChinese = 2 // 中文拼音（如: liufeng802）
	ModeEnglish = 3 // 英文名（如: lindaanderson）
)

// Mailbox 表示一个临时邮箱
type Mailbox struct {
	Address   string    `json:"email"`        // 邮箱地址
	Username  string    `json:"username"`     // 用户名
	Domain    string    `json:"domain"`       // 域名
	ExpiresAt time.Time `json:"expires_at"`   // 过期时间
	CreatedAt time.Time `json:"created_at"`   // 创建时间
}

// Mail 表示邮件基本信息
type Mail struct {
	ID         string    `json:"id"`          // 邮件 ID
	From       string    `json:"from"`        // 发件人
	Subject    string    `json:"subject"`     // 主题
	ReceivedAt time.Time `json:"received_at"` // 接收时间
}

// MailDetail 表示邮件完整详情
type MailDetail struct {
	ID       string    `json:"id"`           // 邮件 ID
	From     string    `json:"from"`         // 发件人
	To       []string  `json:"to"`           // 收件人列表
	Subject  string    `json:"subject"`      // 主题
	TextBody string    `json:"text_content"` // 纯文本内容（用户可自己写正则提取）
	HTMLBody string    `json:"html_content"` // HTML 内容（用户可自己写正则提取）
	ReceivedAt time.Time `json:"received_at"` // 接收时间
}

// CodeResult 表示验证码提取结果
type CodeResult struct {
	Code         string   `json:"code"`           // 提取到的验证码
	Found        bool     `json:"found"`          // 是否找到
	AllCodes     []string `json:"all_codes"`      // 所有找到的验证码
	CheckedMails int      `json:"checked_mails"`  // 检查的邮件数量
	LatestMailID string   `json:"latest_mail_id"` // 最新邮件 ID
}

// apiResponse 表示 API 标准响应
type apiResponse struct {
	Code int             `json:"code"` // 响应码
	Msg  string          `json:"msg"`  // 响应消息
	Data json.RawMessage `json:"data"` // 响应数据
}

// doRequest 执行 HTTP 请求的内部辅助函数
func doRequest(ctx context.Context, baseURL, apiKey, method, path string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body failed: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	fullURL := baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("User-Agent", fmt.Sprintf("Mail2SDK-Go/%s", Version))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response failed: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API error (status=%d): %s", resp.StatusCode, string(respBody))
	}

	if result == nil {
		return nil
	}

	var apiResp apiResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return fmt.Errorf("parse response failed: %w", err)
	}

	if apiResp.Code != 0 && apiResp.Code != 200 {
		return fmt.Errorf("API error (code=%d): %s", apiResp.Code, apiResp.Msg)
	}

	if len(apiResp.Data) > 0 {
		if err := json.Unmarshal(apiResp.Data, result); err != nil {
			return fmt.Errorf("parse data failed: %w", err)
		}
	}

	return nil
}

// filterDomains 过滤黑名单域名
//
// 参数:
//   domains: 原始域名列表
//   blacklist: 黑名单域名列表（支持子串匹配）
//
// 返回:
//   过滤后的域名列表
func filterDomains(domains []string, blacklist []string) []string {
	if len(blacklist) == 0 {
		return domains
	}

	filtered := make([]string, 0, len(domains))
	for _, domain := range domains {
		blocked := false
		for _, bl := range blacklist {
			if containsIgnoreCase(domain, bl) {
				blocked = true
				break
			}
		}
		if !blocked {
			filtered = append(filtered, domain)
		}
	}

	return filtered
}

// containsIgnoreCase 不区分大小写的字符串包含检查
func containsIgnoreCase(s, substr string) bool {
	// 简单的大小写转换
	s = toLower(s)
	substr = toLower(substr)
	return len(s) >= len(substr) && indexSubstring(s, substr) >= 0
}

// toLower 将字符串转为小写
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] >= 'A' && s[i] <= 'Z' {
			result[i] = s[i] + 32
		} else {
			result[i] = s[i]
		}
	}
	return string(result)
}

// indexSubstring 查找子串位置
func indexSubstring(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// GetDomains 获取所有可用域名列表
//
// 参数:
//   baseURL: API 基础地址（如: "https://mail.cwn.cc"）
//   apiKey: API 密钥
//
// 返回:
//   []string: 可用域名列表
//   error: 错误信息
//
// 示例:
//   domains, err := mail2sdk.GetDomains("https://mail.cwn.cc", "your-api-key")
func GetDomains(baseURL, apiKey string) ([]string, error) {
	ctx := context.Background()
	
	var result struct {
		Records []struct {
			Name    string `json:"name"`
			Enabled bool   `json:"enabled"`
		} `json:"records"`
	}

	if err := doRequest(ctx, baseURL, apiKey, "GET", "/api/domains", nil, &result); err != nil {
		return nil, err
	}

	domains := make([]string, 0, len(result.Records))
	for _, d := range result.Records {
		if d.Enabled {
			domains = append(domains, d.Name)
		}
	}

	return domains, nil
}

// CreateMailbox 创建临时邮箱
//
// 参数:
//   baseURL: API 基础地址
//   apiKey: API 密钥
//   mode: 生成模式 (0=自动混用, 1=随机, 2=中文, 3=英文)
//   domain: 指定域名（空字符串=""表示随机选择）
//   blacklist: 黑名单域名列表（可选，传 nil 表示不过滤）
//
// 返回:
//   *Mailbox: 邮箱信息
//   error: 错误信息
//
// 示例:
//   // 随机域名，随机字符
//   mailbox, _ := mail2sdk.CreateMailbox(baseURL, apiKey, 1, "", nil)
//   
//   // 指定域名，中文模式
//   mailbox, _ := mail2sdk.CreateMailbox(baseURL, apiKey, 2, "mail.btlcraft.eu.org", nil)
//   
//   // 自动混用模式，过滤 eu.org 和 edu.kg 域名
//   blacklist := []string{"eu.org", "edu.kg"}
//   mailbox, _ := mail2sdk.CreateMailbox(baseURL, apiKey, 0, "", blacklist)
func CreateMailbox(baseURL, apiKey string, mode int, domain string, blacklist []string) (*Mailbox, error) {
	ctx := context.Background()

	// 处理模式
	var apiMode string
	switch mode {
	case 0: // 自动混用
		modes := []string{"random", "chinese", "english"}
		apiMode = modes[getRand().Intn(3)]
	case 1:
		apiMode = "random"
	case 2:
		apiMode = "chinese"
	case 3:
		apiMode = "english"
	default:
		apiMode = "random"
	}

	// 如果没有指定域名但有黑名单，需要从可用域名中选择
	if domain == "" && len(blacklist) > 0 {
		allDomains, err := GetDomains(baseURL, apiKey)
		if err != nil {
			return nil, fmt.Errorf("获取域名列表失败: %w", err)
		}

		filtered := filterDomains(allDomains, blacklist)
		if len(filtered) == 0 {
			return nil, fmt.Errorf("黑名单过滤后没有可用域名")
		}

		// 使用轮询策略选择域名（确保所有域名均匀使用）
		domain = getDomainSelector().selectDomain(filtered)
	}

	// 构建请求体
	reqBody := map[string]interface{}{
		"mode": apiMode,
	}

	// 如果指定了域名
	if domain != "" {
		reqBody["domain"] = domain
	}

	var mailbox Mailbox
	if err := doRequest(ctx, baseURL, apiKey, "POST", "/api/mailbox", reqBody, &mailbox); err != nil {
		return nil, err
	}

	return &mailbox, nil
}

// CreateMailboxWithDomains 从指定域名组中随机选择一个创建邮箱
//
// 参数:
//   baseURL: API 基础地址
//   apiKey: API 密钥
//   mode: 生成模式 (0=自动混用, 1=随机, 2=中文, 3=英文)
//   domains: 域名数组，SDK 会随机选择一个
//   blacklist: 黑名单域名列表（可选，传 nil 表示不过滤）
//
// 返回:
//   *Mailbox: 邮箱信息
//   error: 错误信息
//
// 示例:
//   domains := []string{"mail.btlcraft.eu.org", "mail.ry.edu.kg"}
//   mailbox, _ := mail2sdk.CreateMailboxWithDomains(baseURL, apiKey, 1, domains, nil)
//   
//   // 使用黑名单过滤
//   blacklist := []string{"eu.org"}
//   mailbox, _ := mail2sdk.CreateMailboxWithDomains(baseURL, apiKey, 1, domains, blacklist)
func CreateMailboxWithDomains(baseURL, apiKey string, mode int, domains []string, blacklist []string) (*Mailbox, error) {
	if len(domains) == 0 {
		return CreateMailbox(baseURL, apiKey, mode, "", blacklist)
	}

	// 过滤黑名单域名
	filtered := filterDomains(domains, blacklist)
	if len(filtered) == 0 {
		return nil, fmt.Errorf("黑名单过滤后没有可用域名")
	}

	// 使用轮询策略选择域名（确保所有域名均匀使用）
	domain := getDomainSelector().selectDomain(filtered)

	return CreateMailbox(baseURL, apiKey, mode, domain, nil)
}

// GetMails 获取邮箱的邮件列表
//
// 参数:
//   baseURL: API 基础地址
//   apiKey: API 密钥
//   address: 邮箱地址
//
// 返回:
//   []Mail: 邮件列表
//   error: 错误信息
//
// 示例:
//   mails, err := mail2sdk.GetMails(baseURL, apiKey, "test@example.com")
func GetMails(baseURL, apiKey, address string) ([]Mail, error) {
	if address == "" {
		return nil, fmt.Errorf("address is required")
	}

	ctx := context.Background()
	path := fmt.Sprintf("/api/mailbox/%s/mails", url.PathEscape(address))

	var result struct {
		Count int    `json:"count"`
		Mails []Mail `json:"mails"`
	}

	if err := doRequest(ctx, baseURL, apiKey, "GET", path, nil, &result); err != nil {
		return nil, err
	}

	return result.Mails, nil
}

// GetMailDetail 获取邮件的完整详情
//
// 返回完整的邮件内容（TextBody 和 HTMLBody），用户可以自己编写正则表达式
// 来提取需要的内容（如链接、特定文本等）。
//
// 参数:
//   baseURL: API 基础地址
//   apiKey: API 密钥
//   address: 邮箱地址
//   mailID: 邮件 ID
//
// 返回:
//   *MailDetail: 邮件详情（包含完整的 TextBody 和 HTMLBody）
//   error: 错误信息
//
// 示例:
//   detail, _ := mail2sdk.GetMailDetail(baseURL, apiKey, address, mailID)
//   
//   // 用户可以自己写正则提取内容
//   re := regexp.MustCompile(`https://[^\s"<>]+`)
//   links := re.FindAllString(detail.HTMLBody, -1)
func GetMailDetail(baseURL, apiKey, address, mailID string) (*MailDetail, error) {
	if address == "" {
		return nil, fmt.Errorf("address is required")
	}
	if mailID == "" {
		return nil, fmt.Errorf("mailID is required")
	}

	ctx := context.Background()
	path := fmt.Sprintf("/api/mailbox/%s/mails/%s", url.PathEscape(address), url.PathEscape(mailID))

	var detail MailDetail
	if err := doRequest(ctx, baseURL, apiKey, "GET", path, nil, &detail); err != nil {
		return nil, err
	}

	return &detail, nil
}

// ExtractCode 提取验证码（使用 API 内置算法）
//
// API 会自动从邮件中提取 4-8 位数字验证码。
//
// 参数:
//   baseURL: API 基础地址
//   apiKey: API 密钥
//   address: 邮箱地址
//   maxMails: 最多检查的邮件数量（0 表示使用默认值 5）
//
// 返回:
//   *CodeResult: 验证码提取结果
//   error: 错误信息
//
// 示例:
//   result, err := mail2sdk.ExtractCode(baseURL, apiKey, address, 5)
//   if err == nil && result.Found {
//       fmt.Println("验证码:", result.Code)
//   }
func ExtractCode(baseURL, apiKey, address string, maxMails int) (*CodeResult, error) {
	if address == "" {
		return nil, fmt.Errorf("address is required")
	}

	ctx := context.Background()
	path := fmt.Sprintf("/api/mailbox/%s/code", url.PathEscape(address))

	if maxMails > 0 {
		path += "?max_mails=" + strconv.Itoa(maxMails)
	}

	var result CodeResult
	if err := doRequest(ctx, baseURL, apiKey, "GET", path, nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteMailbox 删除邮箱及其所有邮件
//
// 注意: 此操作不可逆！
//
// 参数:
//   baseURL: API 基础地址
//   apiKey: API 密钥
//   address: 邮箱地址
//
// 返回:
//   error: 错误信息
//
// 示例:
//   err := mail2sdk.DeleteMailbox(baseURL, apiKey, "test@example.com")
func DeleteMailbox(baseURL, apiKey, address string) error {
	if address == "" {
		return fmt.Errorf("address is required")
	}

	ctx := context.Background()
	path := fmt.Sprintf("/api/mailbox/%s", url.PathEscape(address))

	return doRequest(ctx, baseURL, apiKey, "DELETE", path, nil, nil)
}

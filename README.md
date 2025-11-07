# Mail2 SDK - Go 临时邮箱 SDK

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Version](https://img.shields.io/badge/version-2.0.0-orange)](https://github.com/chuyu5762/mail2sdk)

一个简洁易用的 Go 语言临时邮箱 SDK，用于与 Mail2 临时邮箱系统交互。

## 特性

- ✅ **单文件设计**：只需复制 `mail2sdk.go` 到项目即可使用
- ✅ **多种邮箱生成模式**：随机字符、中文拼音、英文名，支持自动混用
- ✅ **灵活的域名选择**：支持指定域名、域名组随机选择、黑名单过滤
- ✅ **智能轮询策略**：确保多个域名均匀使用，避免单一域名过载
- ✅ **完整的邮件操作**：创建邮箱、获取邮件、提取验证码、删除邮箱
- ✅ **验证码提取**：内置验证码提取功能，自动识别 4-8 位数字验证码
- ✅ **线程安全**：支持并发调用，内置锁机制保证数据一致性

## 快速开始

### 安装

#### 方式一：使用 go get（推荐）

```bash
go get github.com/chuyu5762/mail2sdk@latest
```

或指定版本：

```bash
go get github.com/chuyu5762/mail2sdk@v2.0.0
```

#### 方式二：直接复制文件

```bash
# 下载 SDK 文件
curl -O https://raw.githubusercontent.com/chuyu5762/mail2sdk/main/mail2sdk.go

# 或使用 wget
wget https://raw.githubusercontent.com/chuyu5762/mail2sdk/main/mail2sdk.go
```

### 基本使用

```go
package main

import (
    "fmt"
    "log"
    "github.com/chuyu5762/mail2sdk"
)

func main() {
    baseURL := "https://mail.cwn.cc"
    apiKey := "your-api-key"

    // 创建临时邮箱（随机字符模式）
    mailbox, err := mail2sdk.CreateMailbox(baseURL, apiKey, mail2sdk.ModeRandom, "", nil)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("邮箱地址: %s\n", mailbox.Address)

    // 获取邮件列表
    mails, err := mail2sdk.GetMails(baseURL, apiKey, mailbox.Address)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("收到 %d 封邮件\n", len(mails))

    // 提取验证码（检查最近 5 封邮件）
    codeResult, err := mail2sdk.ExtractCode(baseURL, apiKey, mailbox.Address, 5)
    if err != nil {
        log.Fatal(err)
    }
    if codeResult.Found {
        fmt.Printf("验证码: %s\n", codeResult.Code)
    }

    // 删除邮箱
    err = mail2sdk.DeleteMailbox(baseURL, apiKey, mailbox.Address)
    if err != nil {
        log.Fatal(err)
    }
}
```

## API 文档

### 邮箱生成模式

SDK 支持 4 种邮箱生成模式：

| 常量 | 值 | 说明 | 示例 |
|------|---|------|------|
| `ModeAuto` | 0 | 自动混用（SDK 随机选择） | - |
| `ModeRandom` | 1 | 随机字符 | bd4232@example.com |
| `ModeChinese` | 2 | 中文拼音 | liufeng802@example.com |
| `ModeEnglish` | 3 | 英文名 | lindaanderson@example.com |

### 核心函数

#### 1. CreateMailbox - 创建临时邮箱

```go
func CreateMailbox(baseURL, apiKey string, mode int, domain string, blacklist []string) (*Mailbox, error)
```

**参数：**
- `baseURL`：API 基础地址（如：`https://mail.cwn.cc`）
- `apiKey`：API 密钥
- `mode`：生成模式（0-3）
- `domain`：指定域名（空字符串表示随机选择）
- `blacklist`：黑名单域名列表（可选，传 `nil` 表示不过滤）

**返回：**
- `*Mailbox`：邮箱信息
- `error`：错误信息

**示例：**

```go
// 1. 随机域名 + 随机字符模式
mailbox, err := mail2sdk.CreateMailbox(baseURL, apiKey, mail2sdk.ModeRandom, "", nil)

// 2. 指定域名 + 中文拼音模式
mailbox, err := mail2sdk.CreateMailbox(baseURL, apiKey, mail2sdk.ModeChinese, "mail.btlcraft.eu.org", nil)

// 3. 自动混用模式 + 黑名单过滤
blacklist := []string{"eu.org", "edu.kg"}
mailbox, err := mail2sdk.CreateMailbox(baseURL, apiKey, mail2sdk.ModeAuto, "", blacklist)

// 4. 随机域名 + 英文名模式 + 黑名单过滤
blacklist := []string{"temp.com"}
mailbox, err := mail2sdk.CreateMailbox(baseURL, apiKey, mail2sdk.ModeEnglish, "", blacklist)
```

#### 2. CreateMailboxWithDomains - 从指定域名组创建邮箱

```go
func CreateMailboxWithDomains(baseURL, apiKey string, mode int, domains []string, blacklist []string) (*Mailbox, error)
```

**参数：**
- `baseURL`：API 基础地址
- `apiKey`：API 密钥
- `mode`：生成模式（0-3）
- `domains`：域名数组，SDK 会使用轮询策略选择
- `blacklist`：黑名单域名列表（可选）

**示例：**

```go
// 从指定域名组中选择（使用轮询策略确保均匀分布）
domains := []string{
    "mail.btlcraft.eu.org",
    "mail.ry.edu.kg",
    "temp.example.com",
}
mailbox, err := mail2sdk.CreateMailboxWithDomains(baseURL, apiKey, mail2sdk.ModeRandom, domains, nil)

// 使用黑名单过滤
blacklist := []string{"eu.org"}
mailbox, err := mail2sdk.CreateMailboxWithDomains(baseURL, apiKey, mail2sdk.ModeRandom, domains, blacklist)
```

#### 3. GetDomains - 获取可用域名列表

```go
func GetDomains(baseURL, apiKey string) ([]string, error)
```

**示例：**

```go
domains, err := mail2sdk.GetDomains(baseURL, apiKey)
if err != nil {
    log.Fatal(err)
}
for _, domain := range domains {
    fmt.Println(domain)
}
```

#### 4. GetMails - 获取邮件列表

```go
func GetMails(baseURL, apiKey, address string) ([]Mail, error)
```

**示例：**

```go
mails, err := mail2sdk.GetMails(baseURL, apiKey, mailbox.Address)
if err != nil {
    log.Fatal(err)
}

for _, mail := range mails {
    fmt.Printf("ID: %s\n", mail.ID)
    fmt.Printf("发件人: %s\n", mail.From)
    fmt.Printf("主题: %s\n", mail.Subject)
    fmt.Printf("接收时间: %s\n", mail.ReceivedAt)
    fmt.Println("---")
}
```

#### 5. GetMailDetail - 获取邮件详情

```go
func GetMailDetail(baseURL, apiKey, address, mailID string) (*MailDetail, error)
```

返回完整的邮件内容（包含纯文本和 HTML），用户可以使用正则表达式提取需要的内容。

**示例：**

```go
detail, err := mail2sdk.GetMailDetail(baseURL, apiKey, mailbox.Address, mailID)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("纯文本内容:\n%s\n", detail.TextBody)
fmt.Printf("HTML 内容:\n%s\n", detail.HTMLBody)

// 使用正则提取链接
import "regexp"
re := regexp.MustCompile(`https://[^\s"<>]+`)
links := re.FindAllString(detail.HTMLBody, -1)
for _, link := range links {
    fmt.Println("链接:", link)
}
```

#### 6. ExtractCode - 提取验证码

```go
func ExtractCode(baseURL, apiKey, address string, maxMails int) (*CodeResult, error)
```

使用 API 内置算法自动提取 4-8 位数字验证码。

**参数：**
- `maxMails`：最多检查的邮件数量（0 表示使用默认值 5）

**示例：**

```go
// 检查最近 5 封邮件
result, err := mail2sdk.ExtractCode(baseURL, apiKey, mailbox.Address, 5)
if err != nil {
    log.Fatal(err)
}

if result.Found {
    fmt.Printf("验证码: %s\n", result.Code)
    fmt.Printf("检查了 %d 封邮件\n", result.CheckedMails)
    fmt.Printf("最新邮件 ID: %s\n", result.LatestMailID)
    fmt.Printf("找到的所有验证码: %v\n", result.AllCodes)
} else {
    fmt.Println("未找到验证码")
}
```

#### 7. DeleteMailbox - 删除邮箱

```go
func DeleteMailbox(baseURL, apiKey, address string) error
```

**注意：** 此操作不可逆，会删除邮箱及其所有邮件。

**示例：**

```go
err := mail2sdk.DeleteMailbox(baseURL, apiKey, mailbox.Address)
if err != nil {
    log.Fatal(err)
}
fmt.Println("邮箱已删除")
```

### 数据结构

#### Mailbox - 邮箱信息

```go
type Mailbox struct {
    Address   string    `json:"email"`        // 邮箱地址
    Username  string    `json:"username"`     // 用户名
    Domain    string    `json:"domain"`       // 域名
    ExpiresAt time.Time `json:"expires_at"`   // 过期时间
    CreatedAt time.Time `json:"created_at"`   // 创建时间
}
```

#### Mail - 邮件基本信息

```go
type Mail struct {
    ID         string    `json:"id"`          // 邮件 ID
    From       string    `json:"from"`        // 发件人
    Subject    string    `json:"subject"`     // 主题
    ReceivedAt time.Time `json:"received_at"` // 接收时间
}
```

#### MailDetail - 邮件详情

```go
type MailDetail struct {
    ID         string    `json:"id"`           // 邮件 ID
    From       string    `json:"from"`         // 发件人
    To         []string  `json:"to"`           // 收件人列表
    Subject    string    `json:"subject"`      // 主题
    TextBody   string    `json:"text_content"` // 纯文本内容
    HTMLBody   string    `json:"html_content"` // HTML 内容
    ReceivedAt time.Time `json:"received_at"`  // 接收时间
}
```

#### CodeResult - 验证码提取结果

```go
type CodeResult struct {
    Code         string   `json:"code"`           // 提取到的验证码
    Found        bool     `json:"found"`          // 是否找到
    AllCodes     []string `json:"all_codes"`      // 所有找到的验证码
    CheckedMails int      `json:"checked_mails"`  // 检查的邮件数量
    LatestMailID string   `json:"latest_mail_id"` // 最新邮件 ID
}
```

## 高级功能

### 域名轮询策略

SDK 内置智能域名轮询策略，确保多个域名均匀使用，避免单一域名过载。

```go
// 创建多个邮箱，SDK 会自动均匀分配域名
domains := []string{"domain1.com", "domain2.com", "domain3.com"}

for i := 0; i < 10; i++ {
    mailbox, _ := mail2sdk.CreateMailboxWithDomains(baseURL, apiKey, mail2sdk.ModeRandom, domains, nil)
    fmt.Println(mailbox.Address)
}

// 查看域名使用统计
stats := mail2sdk.GetDomainStats()
for domain, count := range stats {
    fmt.Printf("%s: 使用了 %d 次\n", domain, count)
}

// 重置统计（可选）
mail2sdk.ResetDomainStats()
```

### 黑名单过滤

支持灵活的黑名单过滤，可以过滤特定后缀或域名：

```go
// 过滤 .eu.org 和 .edu.kg 后缀的域名
blacklist := []string{"eu.org", "edu.kg"}
mailbox, _ := mail2sdk.CreateMailbox(baseURL, apiKey, mail2sdk.ModeRandom, "", blacklist)

// 过滤特定域名
blacklist := []string{"temp.mail.com", "test.example.org"}
mailbox, _ := mail2sdk.CreateMailbox(baseURL, apiKey, mail2sdk.ModeRandom, "", blacklist)

// 组合使用域名组和黑名单
domains := []string{"mail1.com", "mail2.eu.org", "mail3.com"}
blacklist := []string{"eu.org"}
mailbox, _ := mail2sdk.CreateMailboxWithDomains(baseURL, apiKey, mail2sdk.ModeRandom, domains, blacklist)
// 最终只会从 mail1.com 和 mail3.com 中选择
```

### 自定义正则提取

除了内置的验证码提取功能，你也可以使用正则表达式提取自定义内容：

```go
import "regexp"

// 获取邮件详情
detail, _ := mail2sdk.GetMailDetail(baseURL, apiKey, mailbox.Address, mailID)

// 提取所有 URL
urlPattern := regexp.MustCompile(`https?://[^\s"<>]+`)
urls := urlPattern.FindAllString(detail.HTMLBody, -1)

// 提取特定格式的代码（如：CODE-123456）
codePattern := regexp.MustCompile(`CODE-\d{6}`)
code := codePattern.FindString(detail.TextBody)

// 提取邮箱地址
emailPattern := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
emails := emailPattern.FindAllString(detail.TextBody, -1)
```

## 实际应用场景

### 1. 自动化测试

```go
func TestUserRegistration(t *testing.T) {
    // 创建临时邮箱
    mailbox, _ := mail2sdk.CreateMailbox(baseURL, apiKey, mail2sdk.ModeRandom, "", nil)
    
    // 使用邮箱注册
    registerUser(mailbox.Address)
    
    // 等待验证邮件
    time.Sleep(5 * time.Second)
    
    // 提取验证码
    result, _ := mail2sdk.ExtractCode(baseURL, apiKey, mailbox.Address, 5)
    if !result.Found {
        t.Fatal("未收到验证码")
    }
    
    // 使用验证码完成注册
    verifyUser(result.Code)
    
    // 清理
    mail2sdk.DeleteMailbox(baseURL, apiKey, mailbox.Address)
}
```

### 2. 批量邮箱创建

```go
// 创建 100 个邮箱用于测试
domains := []string{"test1.com", "test2.com", "test3.com"}
mailboxes := make([]*mail2sdk.Mailbox, 0, 100)

for i := 0; i < 100; i++ {
    mailbox, err := mail2sdk.CreateMailboxWithDomains(
        baseURL, 
        apiKey, 
        mail2sdk.ModeAuto, 
        domains, 
        nil,
    )
    if err != nil {
        log.Printf("创建邮箱失败: %v", err)
        continue
    }
    mailboxes = append(mailboxes, mailbox)
    fmt.Printf("已创建: %s\n", mailbox.Address)
}

// 查看域名分布
stats := mail2sdk.GetDomainStats()
for domain, count := range stats {
    fmt.Printf("%s: %d 个邮箱 (%.2f%%)\n", domain, count, float64(count)/100*100)
}
```

### 3. 验证码监听

```go
// 轮询等待验证码
func waitForCode(address string, timeout time.Duration) (string, error) {
    deadline := time.Now().Add(timeout)
    ticker := time.NewTicker(3 * time.Second)
    defer ticker.Stop()
    
    for time.Now().Before(deadline) {
        result, err := mail2sdk.ExtractCode(baseURL, apiKey, address, 5)
        if err != nil {
            return "", err
        }
        
        if result.Found {
            return result.Code, nil
        }
        
        <-ticker.C
    }
    
    return "", fmt.Errorf("超时：未收到验证码")
}

// 使用示例
mailbox, _ := mail2sdk.CreateMailbox(baseURL, apiKey, mail2sdk.ModeRandom, "", nil)
sendVerificationEmail(mailbox.Address)

code, err := waitForCode(mailbox.Address, 2*time.Minute)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("收到验证码: %s\n", code)
```

## 错误处理

SDK 的所有函数都返回 `error`，建议进行适当的错误处理：

```go
mailbox, err := mail2sdk.CreateMailbox(baseURL, apiKey, mail2sdk.ModeRandom, "", nil)
if err != nil {
    // 处理错误
    log.Printf("创建邮箱失败: %v", err)
    return
}

// API 错误会包含详细信息
// 示例错误：
// - "API error (code=401): Invalid API key"
// - "API error (status=429): Too Many Requests"
// - "黑名单过滤后没有可用域名"
```

## 线程安全

SDK 内部使用了锁机制，所有函数都是线程安全的，可以在并发环境中使用：

```go
import "sync"

var wg sync.WaitGroup
domains := []string{"mail1.com", "mail2.com"}

// 并发创建 50 个邮箱
for i := 0; i < 50; i++ {
    wg.Add(1)
    go func(index int) {
        defer wg.Done()
        
        mailbox, err := mail2sdk.CreateMailboxWithDomains(
            baseURL, 
            apiKey, 
            mail2sdk.ModeRandom, 
            domains, 
            nil,
        )
        if err != nil {
            log.Printf("协程 %d 失败: %v", index, err)
            return
        }
        fmt.Printf("协程 %d: %s\n", index, mailbox.Address)
    }(i)
}

wg.Wait()
fmt.Println("完成")
```

## 性能建议

1. **复用 HTTP 客户端**：SDK 内部每次请求都会创建新的 HTTP 客户端。如果需要高性能，可以考虑修改 SDK 使用单例客户端。

2. **批量操作**：如果需要创建大量邮箱，建议使用并发（但注意 API 速率限制）。

3. **域名选择**：使用 `CreateMailboxWithDomains` 并提供多个域名可以提高成功率和分布均匀性。

4. **错误重试**：对于可能失败的操作（如网络问题），建议实现重试机制：

```go
func createMailboxWithRetry(baseURL, apiKey string, mode int, maxRetries int) (*mail2sdk.Mailbox, error) {
    var lastErr error
    for i := 0; i < maxRetries; i++ {
        mailbox, err := mail2sdk.CreateMailbox(baseURL, apiKey, mode, "", nil)
        if err == nil {
            return mailbox, nil
        }
        lastErr = err
        time.Sleep(time.Second * time.Duration(i+1)) // 指数退避
    }
    return nil, fmt.Errorf("重试 %d 次后失败: %w", maxRetries, lastErr)
}
```

## 常见问题

### 1. 如何获取 API Key？

访问 Mail2 系统的用户面板，登录后在 "API Keys" 页面创建新的 API Key。

### 2. API 有速率限制吗？

是的，每个 API Key 都有每日请求配额和并发限制。具体限制请查看你的 API Key 配置。

### 3. 邮箱会过期吗？

是的，临时邮箱会在创建后的一定时间内过期（由服务端配置）。可以通过 `mailbox.ExpiresAt` 查看过期时间。

### 4. 支持哪些邮箱域名？

使用 `GetDomains()` 函数可以获取当前可用的所有域名列表。

### 5. 验证码提取支持哪些格式？

内置提取功能支持 4-8 位纯数字验证码。如需提取其他格式，请使用 `GetMailDetail()` 获取邮件内容后自行编写正则表达式。

### 6. 可以接收附件吗？

当前版本不支持附件。如有需要，请联系服务提供商。

## 版本历史

### v1.0.0 (2025-11-07)
- ✅ 初始版本发布
- ✅ 支持邮箱创建（4 种模式）
- ✅ 支持域名选择和黑名单过滤
- ✅ 支持邮件获取和详情查看
- ✅ 支持验证码自动提取
- ✅ 支持邮箱删除
- ✅ 智能域名轮询策略
- ✅ 线程安全设计

## 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

## 相关链接

- 项目主页：https://github.com/chuyu5762/mail2sdk
- 问题反馈：https://github.com/chuyu5762/mail2sdk/issues
- Mail2 系统：https://mail.cwn.cc

---

**最后更新**：2025-11-07  
**版本**：v1.0.0

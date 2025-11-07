# 临时邮箱系统（Mail2）

一个基于 Go + Gin 框架的临时邮箱服务系统，支持 Casdoor 认证、API Key 管理、动态域名配置等功能。

## 项目简介

临时邮箱系统为用户提供临时、一次性的邮箱地址服务，用于接收验证码、测试邮件等场景。系统采用前后端分离架构，后端使用 Go 语言开发，集成 Casdoor 统一认证平台。

## 主要特性

- ✅ **邮箱管理**：创建、查询、删除临时邮箱，支持自定义有效期
- ✅ **邮件收发**：接收邮件并实时查询，支持邮件列表和详情查看
- ✅ **用户认证**：集成 Casdoor OAuth 2.0 授权码流程
- ✅ **API Key 管理**：为开发者提供 API Key 创建、管理和速率限制
- ✅ **域名动态配置**：支持多域名管理，动态启用/禁用域名
- ✅ **文件存储**：邮件正文存储于文件系统，数据库仅保存索引
- ✅ **速率限制**：每日请求配额和并发限制

## 技术栈

### 后端
- **语言**：Go 1.21+
- **框架**：Gin（HTTP 框架）
- **数据库**：MySQL 5.7+（通过 GORM 操作）
- **认证**：Casdoor（OAuth 2.0）
- **存储**：文件系统（邮件内容）

### 前端（规划中）
- **框架**：React + TypeScript
- **UI 库**：Ant Design Pro
- **状态管理**：Umi + Dva

## 快速开始

### 环境要求

- Go 1.21+
- MySQL 5.7+
- Git

### 安装步骤

1. **克隆仓库**
```bash
git clone https://cnb.cool/pu.ac.cn/mail2.git
cd mail2
```

2. **初始化数据库**
```bash
# 登录 MySQL
mysql -u root -p

# 执行初始化脚本
source backend/scripts/init_db.sql
```

3. **配置应用**
```bash
cd backend
cp configs/config.example.yaml configs/config.yaml
# 编辑 configs/config.yaml，配置数据库连接和 Casdoor 参数
```

4. **启动服务**

**Windows:**
```cmd
backend\scripts\start.bat
```

**Linux/Mac:**
```bash
chmod +x backend/scripts/start.sh
./backend/scripts/start.sh
```

5. **验证服务**
```bash
curl http://localhost:8888/healthz
```

### 更多文档

- 📖 [快速开始指南](test-docs/快速开始指南.md)（详细部署和测试说明）
- 📋 [开发文档](docs/临时邮箱系统开发文档.md)
- 📅 [开发计划](docs/临时邮箱系统开发计划.md)
- 🔧 [开发规范](AGENTS.md)

## 项目结构

```
mail2/
├── backend/              # 后端服务
│   ├── cmd/             # 程序入口
│   ├── configs/         # 配置文件
│   ├── internal/        # 内部代码
│   │   ├── app/        # 应用装配
│   │   ├── auth/       # Casdoor 认证
│   │   ├── config/     # 配置加载
│   │   ├── handlers/   # HTTP 处理器
│   │   ├── middleware/ # 中间件
│   │   ├── models/     # 数据模型
│   │   ├── repository/ # 数据访问层
│   │   └── service/    # 业务逻辑层
│   └── scripts/        # 脚本工具
├── docs/                # 项目文档
├── test-docs/           # 测试文档（不纳入版本控制）
├── test/                # 测试脚本（不纳入版本控制）
├── logs/                # 日志文件（不纳入版本控制）
├── storage/             # 邮件存储
└── bin/                 # 编译产物（不纳入版本控制）
```

## 核心 API

### 公共接口（需要 API Key）

- `GET /api/domains` - 获取域名列表
- `POST /api/mailbox` - 创建邮箱
- `GET /api/mailbox/:address/mails` - 获取邮件列表
- `GET /api/mailbox/:address/mails/:id` - 获取邮件详情
- `DELETE /api/mailbox/:address` - 删除邮箱

### 用户接口（需要 Casdoor 登录）

- `GET /user/api/mailboxes` - 获取我的邮箱列表
- `GET /user/api/mails` - 获取我的所有邮件
- `POST /user/api/api-keys` - 创建 API Key
- `GET /user/api/api-keys` - 获取 API Key 列表
- `DELETE /user/api/api-keys/:id` - 删除 API Key

详细 API 文档请参考项目文档。

## 开发规范

本项目严格遵循以下开发规范（详见 [AGENTS.md](AGENTS.md)）：

- ✅ **中文注释**：所有代码注释必须使用中文
- ✅ **DRY 原则**：禁止重复代码，3行以上重复必须提取
- ✅ **目录规范**：严格的目录结构和用途限制
- ✅ **Git 提交**：使用中文提交信息，遵循提交类型规范
- ✅ **自动提交**：完成功能后立即提交并推送

## 开发进度

截至 2025-10-31，项目已完成以下功能：

### 已完成 ✅
- [x] Go 后端服务框架
- [x] Casdoor 授权流程
- [x] MySQL 表结构设计与迁移
- [x] 邮箱生命周期管理
- [x] 邮件读取 API
- [x] API Key 管理与速率限制
- [x] 文件存储目录结构
- [x] 启动和测试脚本

### 进行中 🚧
- [ ] 域名动态配置后台
- [ ] 邮箱过期自动清理
- [ ] 真实邮件接收通道

### 计划中 📅
- [ ] Ant Design Pro 前端
- [ ] CI/CD Pipeline
- [ ] 监控与告警
- [ ] 性能优化与缓存

详细开发计划请查看 [开发计划文档](docs/临时邮箱系统开发计划.md)。

## 数据库设计

### 核心表结构

- **users**：用户表（UUID 主键，关联 Casdoor 用户）
- **domains**：域名表（支持动态启用/禁用）
- **mailboxes**：邮箱表（临时邮箱地址）
- **mail_indexes**：邮件索引表（邮件 ID 和文件路径）
- **api_keys**：API Key 表（用户 UUID 绑定）
- **rate_limits**：速率限制表（每日请求统计）

所有 ID 字段使用 UUID4 标准，确保全局唯一性。

## 配置说明

关键配置项（`backend/configs/config.yaml`）：

```yaml
server:
  addr: 0.0.0.0:8888      # 服务监听地址

mysql:
  dsn: root:password@tcp(127.0.0.1:3306)/mail2?charset=utf8mb4&parseTime=True&loc=Local

casdoor:
  endpoint: http://localhost:8000
  client_id: your_client_id
  client_secret: your_client_secret
  organization_name: built-in
  application_name: app-built-in
  redirect_url: http://localhost:8888/auth/casdoor/callback

filestorage:
  root: /path/to/storage   # 邮件存储路径
```

## 安全说明

- ⚠️ 邮件正文存储在文件系统，请确保目录权限安全
- ⚠️ API Key 仅在创建时返回明文，数据库存储哈希值
- ⚠️ 所有写操作需要 Token 或 API Key 认证
- ⚠️ 配置文件包含敏感信息，不应提交到 Git

## 贡献指南

1. Fork 本仓库
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 遵循项目开发规范（见 [AGENTS.md](AGENTS.md)）
4. 提交变更 (`git commit -m '新增: 添加某某功能'`)
5. 推送分支 (`git push origin feature/AmazingFeature`)
6. 创建 Pull Request

## 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

## 联系方式

- **项目仓库**：https://cnb.cool/pu.ac.cn/mail2
- **问题反馈**：通过 Issue 提交

---

**最后更新**：2025-10-31  
**版本**：v0.1.0-alpha

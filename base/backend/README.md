# Base 底座后端

基于 Go 1.21 + Gin + GORM + MySQL + Redis 实现。

## 快速开始

```bash
cd base/backend

# 1. 安装依赖
go mod tidy

# 2. 修改 config.yaml 中的 MySQL/Redis 配置
# 3. 确保 MySQL 数据库已存在（配置文件中的 dbname）

# 4. 启动服务
go run cmd/server/main.go
```

服务默认监听 `:8080`。首次启动会自动执行数据初始化：

- 超级管理员：`admin / admin123`
- 底座默认菜单（控制台、系统管理、消息管理、工作流管理及其子菜单）
- 默认超级管理员角色，并关联所有底座菜单与全部接口权限点
- 底座接口权限点（`base_permission`，按 `code` 幂等写入，新增接口只需在 `internal/seed/permission_seed.go` 补登记）

### 命令行参数

```bash
go run cmd/server/main.go -mock-data    # 额外插入模拟机构数据（仅本地测试）
```

### 手动初始化管理员

初始化超管接口（仅在平台还没有 admin 账号时创建，不提供重置密码——重置请由管理员登录后在「用户管理 → 重置密码」操作）：

```bash
POST /base/api/v1/auth/init
{ "password": "admin123" }
```

## 目录说明

```
backend/
├── cmd/server/          # 服务启动入口
├── config/
│   └── config.go        # 配置结构定义与加载
├── config.yaml          # 运行时配置文件（需自行修改数据库连接）
├── internal/
│   ├── adapter/         # 子应用注册表与统一代理
│   │   ├── controller.go   # /base/api/v1/app/:appCode/* 入口
│   │   ├── proxy.go        # 反向代理实现
│   │   └── registry.go     # 应用配置内存缓存
│   ├── controllers/     # HTTP 控制器
│   ├── middleware/      # JWT、租户、权限、审计、CORS
│   ├── models/          # GORM 数据模型
│   ├── routes/          # 路由注册
│   ├── seed/            # 默认数据初始化与模拟数据
│   └── service/         # 业务逻辑
└── pkg/                 # 公共包
    ├── db/              # MySQL 连接
    ├── jwt/             # JWT 生成与解析
    ├── notifier/        # 邮件等通知渠道
    ├── redis/           # Redis 连接
    ├── response/        # 统一响应封装
    ├── storage/         # 文件存储
    └── utils/           # 通用工具
```

## 配置文件

参考 `config.yaml`：

```yaml
server:
  port: 8080
  mode: debug          # debug / release
  # 可信反向代理：仅当请求来自这些地址时才信任 X-Forwarded-For / X-Real-IP
  trusted_proxies:
    - 127.0.0.1
  # 登录/验证码/初始化接口限流（每窗口允许的请求数 + 窗口秒数）
  login_rate_limit: 10
  login_rate_window_seconds: 60
  captcha_rate_limit: 30
  captcha_rate_window_seconds: 60
  init_rate_limit: 5
  init_rate_window_seconds: 60
  # 上传目录与单文件大小上限（MB）
  upload_dir: ./uploads
  max_upload_mb: 50

mysql:
  host: 127.0.0.1
  port: 3306
  user: root
  password: xxx
  dbname: caam_base
  charset: utf8mb4
  max_open: 100
  max_idle: 10

redis:
  addr: 127.0.0.1:6379
  password:
  db: 4

jwt:
  secret: base-platform-jwt-secret
  expire_hours: 8
  issuer: base-platform
```

## 核心接口

### 公开接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/base/api/v1/auth/login` | 登录（支持租户编码、验证码） |
| POST | `/base/api/v1/auth/init` | 首次部署时创建超级管理员（已存在则报错） |
| GET  | `/base/api/v1/auth/captcha` | 获取图形验证码（5 位数字+字母，返回 `captcha_id` / `captcha_img`） |
| GET  | `/base/api/v1/site-info` | 站点公开配置（`captchaEnabled`，登录页据此决定是否展示验证码） |
| GET  | `/base/api/v1/files/*key` | 文件公开访问（key 含日期目录，如 `20260101/xxx.png`） |

> 三个公开接口（`login` / `captcha` / `init`）均按真实客户端 IP 限流，超限返回 `code=429`，各自独立额度，可在 `server` 配置段调整。

### 登录后接口（均需要 JWT，并经过操作审计、接口权限校验）

| 分组 | 方法 | 路径 | 说明 |
|------|------|------|------|
| 认证 | GET/POST | `/base/api/v1/auth/info` / `change-password` / `logout` | 当前用户、修改密码（改密后旧 token 失效）、登出（token 进黑名单） |
| 仪表盘 | GET | `/base/api/v1/dashboard/stats` | 仪表盘统计 |
| 租户 | CRUD | `/base/api/v1/tenants` | 仅限超级管理员 |
| 系统设置 | GET/PUT/POST | `/base/api/v1/settings` | 仅平台超管；含邮件测试（保存后立即重载通知渠道） |
| 数据字典 | CRUD | `/base/api/v1/dicts` | 含字典项管理 |
| 应用 | CRUD | `/base/api/v1/apps` | 应用定义（平台级：租户可读，仅平台超管可增删改） |
| 应用实例 | CRUD | `/base/api/v1/app-instances` | 含我的应用列表 |
| 用户 | CRUD | `/base/api/v1/users` | 含分配角色、重置密码（需传入新密码，服务端按安全策略校验长度） |
| 角色 | CRUD | `/base/api/v1/roles` | 含分配菜单、分配权限 |
| 菜单 | CRUD | `/base/api/v1/menus/tree` | 菜单树 |
| 权限 | CRUD | `/base/api/v1/permissions/tree` | 权限树 |
| 流程角色 | CRUD | `/base/api/v1/workflow-roles` | 审批角色定义与成员：`GET /workflow-roles/user-options`（成员选项）、`POST /workflow-roles/:id/users`（覆盖式设置成员） |
| 流程定义 | CRUD | `/base/api/v1/workflows` | 含节点编排 `PUT /workflows/:id/nodes`（覆盖式保存）；`GET /workflows/options`（启用中的流程，白名单）、`GET /workflows/approver-options`（审批人候选） |
| 流程实例 | POST/GET/DELETE | `/base/api/v1/workflow-instances` | `POST` 发起（白名单）、`GET` 列表（非管理员强制只看自己发起的）、`GET /:id` 详情（含任务与流转日志）、`POST /:id/cancel` 撤销、`DELETE /:id` 删除（仅管理员） |
| 审批任务 | GET/POST | `/base/api/v1/workflow-tasks` | `GET` 我的待办/已办（`box=todo\|done`）、`POST /:id/approve` 通过、`POST /:id/reject` 驳回（均为白名单，归属校验在服务层） |
| 机构 | CRUD | `/base/api/v1/organizations/tree` | 机构树 |
| 消息 | CRUD | `/base/api/v1/messages` | 草稿（`POST /messages`、`PUT /messages/:id`、`POST /messages/:id/send`）、发送（`POST /messages/send`，支持 `templateCode` + `vars` 按模板渲染）、可用渠道（`GET /messages/channels`）、未读数、标记已读、全部已读；`box=inbox/sent` 区分收发 |
| 消息模板 | CRUD | `/base/api/v1/message-templates` | 模板管理 |
| 操作日志 | GET/POST/GET | `/base/api/v1/operation-logs` | 列表、删除、清空、导出 |
| 登录日志 | GET/POST/GET | `/base/api/v1/login-logs` | 列表、删除、清空、导出 |
| 文件 | POST/GET/DELETE | `/base/api/v1/files/*` | 上传、列表、删除 |
| 子应用代理 | ANY | `/base/api/v1/app/:appCode/*path` | 统一反向代理入口 |

## 安全与隔离

### 接口权限校验

- 所有登录接口默认挂载 `PermissionAuth` 中间件。
- 基于 `base_permission` 表的 `method` + `path` 进行匹配，支持 `:param` 通配符。
- 放行顺序：平台超管（`tenantID == 0`）→ 租户管理员（`is_admin`）→ 白名单接口 → 未分配权限的普通用户一律 403（白名单外不再放行 GET）；其余按权限点精确匹配，未命中返回 403。
- 详见 [docs/security.md](./docs/security.md)。

### 公开接口限流

- `RateLimitMiddleware(limit, window)`：基于 Redis 固定窗口、按「路由 + 真实客户端 IP」计数，Redis 不可用时退化为进程内限流。
- 已挂载：`POST /auth/login`（10/分钟）、`GET /auth/captcha`（30/分钟）、`POST /auth/init`（5/分钟），参数见 `config.yaml`。
- 客户端 IP 依赖 `server.trusted_proxies` 配置，部署在反向代理后必须填写代理地址。
- 详见 [docs/security.md](./docs/security.md)。

### 租户数据隔离

- 带 `tenant_id` 的模型，按 ID 操作时 service 层接收 `tenantID` 参数，普通租户自动追加 `WHERE tenant_id = ?`，超级管理员不过滤。
- 创建时统一使用 `controllers.resolveTenantID`：平台超管可通过请求体的 `tenantId` 指定目标租户，普通租户用户强制写入自身租户。
- 列表查询：普通租户只看本租户；平台超管传 `tenantId` 查询参数则过滤，不传则查看全部租户。
- `Tenant`、`App`、`Permission`、`Setting` 属于平台级资源，不执行租户隔离。
- 详见 [docs/security.md](./docs/security.md)。

## 子应用代理

- 应用在 `base_app` 表中注册后，内存注册表 `adapter.DefaultRegistry` 会自动重载。
- 代理路径：`/base/api/v1/app/{appCode}/{原应用API路径}`。
- 代理时会在请求头注入当前用户信息：
  - `X-Base-User-ID`
  - `X-Base-Username`
  - `X-Base-Tenant-ID`
- IFrame 类型应用不支持 API 代理。

## 工作流引擎

实现位置：`internal/service/workflow_service.go`（流程定义与节点）、`internal/service/workflow_engine_service.go`（引擎与查询）。

- **节点串行 + 或签**：节点按 `sort` 升序执行；一个节点若解析出多个审批人，则每人一条待办，任一人处理即视为节点完成，**同节点其余待办自动置为「已失效」**。
- **审批人解析**（`resolveApprovers`）：`role` → 流程角色成员（仅 `status=1` 用户）、`user` → 指定用户、`initiator` → 实例发起人。
- **自动通过**：节点解析不到有效审批人时直接跳过并写一条 `auto-pass` 日志，避免流程卡死。
- **推进**（`advanceWorkflow`）：审批通过后从当前节点顺序找下一个有审批人的节点；已无后续节点则实例置为「已通过」并记 `finish` 日志。
- **驳回**：实例置「已驳回」、结束时间落库，实例下所有未处理待办置为「已失效」。
- **撤销**：实例置「已撤销」，未处理待办失效；仅发起人本人或管理员可撤销。
- **删除**：仅管理员（`base:workflow-instance:delete`），软删除实例及其任务与流转日志。
- **可见性**：详情允许发起人、参与审批的人、管理员查看；列表对非管理员强制 `initiator_id = 自己`。
- **草稿/时间字段**：`end_at`、`handled_at`、`read_at`、`send_at` 均为 `*time.Time`，未发生时写入 `NULL`，避免 MySQL 严格模式（`NO_ZERO_DATE`）拒绝 `0000-00-00`。

## 数据模型

### 带租户隔离（含 `tenant_id`）

- `base_user`
- `base_role`
- `base_workflow_role`（流程角色，与系统角色解耦；成员关联表 `base_workflow_role_user`）
- `base_workflow`（流程定义）与 `base_workflow_node`（流程节点）
- `base_workflow_instance`（流程实例）、`base_workflow_task`（审批任务）、`base_workflow_log`（流转日志）
- `base_menu`
- `base_organization`
- `base_app_instance`
- `base_dict`
- `base_message`
- `base_message_template`
- `base_uploaded_file`
- `base_operation_log`
- `base_login_log`

> 流程角色（`base_workflow_role`，成员关联表 `base_workflow_role_user`）是「工作流管理」模块的基础数据：
> 只负责定义审批角色及其成员，供审批流程节点选择审批人使用。

### 平台级资源（不含 `tenant_id`）

- `base_app`
- `base_permission`
- `base_setting`
- `base_tenant`

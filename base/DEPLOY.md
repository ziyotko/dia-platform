# 底座平台（Base / business_base）部署说明

> 项目路径：`base/`｜前端部署子路径：`/business_base/`｜接口前缀：`/business_base/api`
> 部署拓扑：Nginx 承载前端静态资源并反代后端 → Go 后端（`0.0.0.0:8091`，配置项 `server.port`） → MySQL / Redis
> 定位：Base 是「管理后台底座」，自身不含业务；业务系统（门户、会员、应用等）以 **IFrame** 或 **API 代理** 方式接入。
> 文档与代码同步至：2026-09-24（数据库结构见 [backend/docs/database.md](backend/docs/database.md)，操作说明见 [frontend/docs/用户使用手册.md](frontend/docs/用户使用手册.md)）

## 目录结构

```
base/
├── backend/          # Go 后端（Gin + GORM），module: base
│   ├── main.go       # 唯一入口（与 config.yaml 同级，cmd/ 已废弃删除）
│   ├── config.yaml   # 唯一配置文件（viper 从当前工作目录读 ./config.yaml，必须在 backend 目录下启动）
│   ├── comp.bat      # 交叉编译脚本（set goos=linux + go build -o base main.go）
│   ├── docs/         # 文档（docs/database.md）
│   ├── uploads/      # 上传文件目录（运行时需可写）
│   └── internal/     # adapter / controllers / middleware / models / routes / seed / service
└── frontend/         # 底座管理后台（Vue3 + Vite + Element Plus）
    └── .env          # 部署子路径 VITE_BASE_PATH 与接口前缀 VITE_API_BASE_URL
```

---

## 一、环境依赖

| 组件 | 版本/说明 |
| --- | --- |
| Go | ≥ 1.27（`backend/go.mod` 声明 `go 1.27.1`），交叉编译 Linux 需 `CGO_ENABLED=0` |
| Node.js | ≥ 18（Vite 5） |
| MySQL | ≥ 8.0（InnoDB + utf8mb4），库名 `caam_base`，端口/账号按环境配置（本仓库当前为 `10.3.1.95:63400`） |
| Redis | 端口 6379，**单库**：`redis.db`（本仓库当前为 `1`）同时承载验证码、登录失败计数、限流计数、Token 吊销黑名单 |
| Nginx | 托管前端静态文件并反代 API；**需把 Nginx 自身地址填入 `server.trusted_proxies`**，否则限流会把所有用户算成同一个 IP |

> ⚠️ **程序不建库、不建账号**：请先手工创建数据库 `caam_base`（InnoDB + utf8mb4）并授予建表 / `ALTER`（加索引）权限。
> ⚠️ **Redis 必须可达**：`pkg/redis.Init` 会 `PING`，失败即 `Fatal` 退出进程（验证码、限流、Token 吊销都依赖 Redis）。

---

## 二、后端部署（backend）

### 1. 配置（config.yaml）

`backend/config.yaml` 全量说明（默认值由 `config.applyDefaults()` 兜底，缺失项不会导致启动失败）：

```yaml
server:
  port: 8091                              # 监听端口（改端口需同步 Nginx 与 frontend/vite.config.ts 的 dev 代理）
  mode: debug                             # debug / release；留空回退 release。debug 会打开 GORM SQL 日志
  api_prefix: /business_base/api          # HTTP 路由前缀（兜底值 config.DefaultAPIPrefix）
  trusted_proxies:                        # 可信反向代理地址；为空回退 [127.0.0.1]
    - 127.0.0.1                           # 生产必须填 Nginx 的 IP，否则限流按 RemoteAddr 统计
  login_rate_limit: 10                    # POST /auth/login    每窗口请求数
  login_rate_window_seconds: 60           #                     窗口秒数
  captcha_rate_limit: 30                  # GET /auth/captcha、GET /site-info
  captcha_rate_window_seconds: 60
  init_rate_limit: 5                      # POST /auth/init
  init_rate_window_seconds: 60
  upload_dir: ./uploads                   # 上传目录（相对进程工作目录；初始化时转绝对路径）
  max_upload_mb: 50                        # 单文件大小上限（MB），<=0 回退 50
  workflow_remind_interval_seconds: 600   # 工作流超时提醒扫描周期（秒）；显式配 0/负数表示关闭
mysql:
  host: 10.3.1.95
  port: 63400
  user: root
  password: '1qaz@WSX'                    # ⚠️ 明文密码，请按环境修改并限制文件权限
  dbname: caam_base
  charset: utf8mb4
  max_open: 100
  max_idle: 10
redis:
  addr: 127.0.0.1:6379
  password: ''
  db: 1
jwt:
  secret: base-platform-jwt-secret        # ⚠️ 生产必须换成 ≥32 位随机串
  expire_hours: 8                         # 登录态有效期（小时）
  issuer: base-platform
```

| 关键配置 | 作用 |
| --- | --- |
| `server.api_prefix` | **前缀唯一来源**：路由注册、权限白名单匹配、审计脱敏判定、文件 URL、子应用代理路径都通过 `pkg/permmatch.APIPrefix()` 取；改名只需改这里 + 前端 `.env` |
| `server.trusted_proxies` | 决定是否信任 `X-Forwarded-For` / `X-Real-IP`；仅在来源属于该列表时使用，否则一律 `RemoteAddr`（防伪造头绕过限流） |
| `server.upload_dir` / `max_upload_mb` | 上传落盘目录与单文件上限；控制器按 header 快速拒绝，存储层再用 `io.LimitReader` 兜底 |
| `server.workflow_remind_interval_seconds` | 工作流超时提醒后台轮询周期，`<= 0` 不启动提醒 |

### 2. 敏感信息（与 portal / member 的差异）

Base 后端 **没有** 环境变量注入通道，数据库密码与 JWT 密钥**直接写在 `config.yaml`**：

- 部署时请修改 `mysql.password` 与 `jwt.secret`，并让 `config.yaml` 仅对运行账号可读（如 `chmod 600 config.yaml`）。
- 不要将真实密码/密钥提交到仓库；仓库中的值是开发库占位。
- **更换 `jwt.secret` 会让所有已签发 Token 立即失效**（在线用户需重新登录），属于紧急踢人手段之一。

### 3. 编译

```bash
cd base/backend
# Windows 交叉编译 Linux 可执行文件（等价 comp.bat）
set goos=linux && go build -o base main.go
# 或直接在目标机编译：
go build -o base main.go
```

仓库根目录的批量脚本可与其他后端一起编译（默认 linux/amd64、`CGO_ENABLED=0` 静态二进制，编译后校验 ELF 架构）：

```powershell
powershell -ExecutionPolicy Bypass -File .\build-backends.ps1 -Only base -Vet
```

产物：`base/backend/base`（已被 `backend/.gitignore` 忽略，重建不产生 diff）。

### 4. 启动

```bash
cd base/backend
./base            # 必须在 backend 目录下启动，否则读不到 ./config.yaml
```

启动顺序（每一步失败都会 `Fatal` 退出，日志用 logrus 输出到 stdout）：

1. 读 `./config.yaml`（`config.Load` + `applyDefaults`）
2. `db.Init` → 连接 MySQL → `migrate()` → `normalizeLegacyData()`
   - `purgeLegacySoftDeletedRows()`：旧库 `deleted_at IS NOT NULL` 的残留行**物理删除**（幂等，新库直接跳过）
   - `dedupeUserUsernames()`：历史重复用户名整理（保 id 最小的一条，其余改名 `<原名>_dup<id>`），否则建唯一索引会失败
   - `AutoMigrate`（27 张表）+ `dropLegacyUserUsernameIndex()`（删历史单列索引）
   - `normalizeLegacyData()`：历史消息 `status 0/1` → `is_read=1` + `status=3`
3. `redis.Init`（PING）
4. `adapter.DefaultRegistry.Reload()`（应用注册表；失败仅告警）
5. `service.AuthService.EnsureSuperAdmin("admin123")` → 幂等创建 `admin / admin123`
6. `seed.Run()` → 清理废弃占位菜单 → 补种菜单 → 补种权限点 → 全部底座权限点授予 `super_admin` 角色
7. `service.ReloadNotifiers()` → 按数据库设置注册邮件 / 企业微信 / 短信渠道（未配置的不注册）
8. `go service.StartWorkflowReminder()` → 工作流超时提醒轮询
9. `routes.Register` → 监听端口

> 种子数据在每次启动时**幂等**执行，改代码后重启后端即可生效（`go run .` 亦可）。
> 菜单按 `(app_code, parent_id, name)` 补种：**已存在不覆盖**，保留管理员在后台的改名/调整，只补齐缺失项。
> 权限点按 `code` upsert（名称/类型/父子/方法/路径都会同步），并把 `app_code = base` 的全部权限点授予平台超管角色。

**默认账号**：`admin` / `admin123`（平台超级管理员，`tenant_id = 0`）。**上线后请立即登录「个人中心 → 修改密码」**。
修改密码后该用户此前签发的 Token 全部失效。

### 5. 命令行参数

```bash
./base -mock-data     # 额外插入 10 条模拟机构数据（仅本地测试；库中已有机构则跳过）
```

### 6. 手工初始化超管接口（可选）

仅当平台还没有 `admin` 账号时可用（幂等，已存在则报错；**不提供重置密码**，重置请由管理员在「用户管理 → 重置密码」操作）：

```bash
POST /business_base/api/auth/init
{ "password": "<≥8位新密码>" }
```

该接口限流 5 次/分钟。

### 7. 健康检查与统一响应

- 健康检查（公开、无需 Token）：`GET /business_base/api/site-info` → `{code:0,data:{captchaEnabled}}`
- **统一返回 HTTP 200 + 业务码**：`code=0` 成功、`400` 参数错误、`401` 未认证/Token 失效、`403` 无权限、`404` 不存在、`429` 触发限流、`500` 业务失败。
- 前端拦截器遇 `code=401` 会清理本地会话并跳转 `${BASE_URL}login`；遇其它非 0 码统一弹错误提示。
- 文件访问：`GET /business_base/api/files/*key`（公开，key 形如 `20260101/xxx.png`）。

### 8. 升级说明（无需手工 SQL）

- **菜单 / 权限点幂等补种**：新版本新增的菜单与权限点会在启动时自动补齐；平台超管角色自动获得新增权限点，升级后平台管理员不会被新权限点挡住。
- **废弃占位菜单自动清理**：历史遗留的 4 个工作流占位菜单（组件 `base/workflow/{model,instance,task,designer}/index.vue`）会被物理删除并解除角色菜单关联，避免点击落到 404。日志出现 `已清理 N 个历史遗留的占位菜单` 即已生效。
- **历史软删数据自动清理**：日志形如 `[migrate] <表> 表物理删除 N 条历史软删除记录`；**首次启动可能触发多轮重试**（父表被外键引用时需等子表先清空），最多 3 轮，仍失败只告警不阻断启动。
- **用户名去重 + 唯一索引**：启动时 `dedupeUserUsernames()` 先整理重复用户名（改名 `<原名>_dup<id>` 并打日志），再建唯一索引 `uk_base_user_tenant_username (tenant_id, username)`。
- **机构外键列**：`AutoMigrate` 会给 `base_user` 补 `organization_id` 列，无需手工处理。
- **限流 / 开关参数**：`config.yaml` 未配 `*_rate_limit` 时自动回退默认值，旧配置可直接使用。

### 9. 可选手工 SQL：清理历史 `deleted_at` 列

`AutoMigrate` **只加不删**：旧库仍会看到 `deleted_at` 列与 `idx_<表名>_deleted_at` 索引（新库没有，无害）。
如需彻底清理，先启动过一次新版后端（让启动期清理把软删行物理删除，**顺序不能颠倒**），再执行：

```sql
-- 1) 核对：哪些表还带 deleted_at
SELECT TABLE_NAME, INDEX_NAME FROM information_schema.STATISTICS
 WHERE TABLE_SCHEMA = DATABASE() AND COLUMN_NAME = 'deleted_at'
 GROUP BY TABLE_NAME, INDEX_NAME;

-- 2) 生成删列语句（索引名遵循 GORM 默认 idx_<表名>_deleted_at，缺失时整条会报 1091，可忽略）
SELECT CONCAT('ALTER TABLE `', TABLE_NAME, '` DROP INDEX `', INDEX_NAME, '`, DROP COLUMN `deleted_at`;')
  FROM information_schema.STATISTICS
 WHERE TABLE_SCHEMA = DATABASE() AND COLUMN_NAME = 'deleted_at'
 GROUP BY TABLE_NAME, INDEX_NAME;
```

> 执行前请先 `mysqldump` 备份；只删列/索引，各表其余字段与数据不受影响。

---

## 三、前端部署（frontend）

### 1. 构建

```bash
cd base/frontend
npm install
npm run build        # 等价 vue-tsc && vite build
```

也可用仓库根目录脚本（构建后自动校验 `dist/index.html` 的资源前缀与 `.env` 的 `VITE_BASE_PATH` 一致）：

```powershell
powershell -ExecutionPolicy Bypass -File .\build-frontends.ps1 -Only business_base -Install
```

构建产物：`base/frontend/dist/`

### 2. 关键配置

- `.env`（唯一环境文件，无 `.env.development` / `.env.production`）：

```
VITE_BASE_PATH=/business_base/
VITE_API_BASE_URL=/business_base/api
```

- `vite.config.ts` 用 `loadEnv` 读取上面两项：`base = VITE_BASE_PATH`，dev 代理键 = `VITE_API_BASE_URL`。
- **dev 代理目标写死在 `vite.config.ts`**（当前 `http://127.0.0.1:8091`），改后端 `server.port` 必须同步改这里。
- dev 端口 3000，访问 `http://localhost:3000/business_base/`。
- 路由基路径用 `createWebHistory(import.meta.env.BASE_URL)`、跳登录页用 `${import.meta.env.BASE_URL}login`——**不要硬编码子路径**，否则改名后会出现路由不匹配、登录跳转 404。
- 改部署路径需同步三处：前端 `.env` 的 `VITE_BASE_PATH` / `VITE_API_BASE_URL`、后端 `config.yaml` 的 `server.api_prefix`、Nginx 的 `location`。

### 3. 部署

将 `dist/` 内容复制到 Nginx 站点目录（如 `/var/www/business_base`）。

---

## 四、Nginx 反向代理示例

```nginx
server {
    listen 80;
    server_name base.example.com;

    # 管理后台静态资源（/business_base/ 下）
    location /business_base/ {
        root /var/www;                 # dist 需放在 /var/www/business_base/，保证 /business_base/ 命中 index.html
        try_files $uri $uri/ /business_base/index.html;
        index index.html;
    }

    # 后端 API（proxy_pass 后不要带 URI，否则会丢掉 /business_base/api 前缀）
    location /business_base/api/ {
        proxy_pass http://127.0.0.1:8091;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        client_max_body_size 60m;      # 需 ≥ server.max_upload_mb（默认 50MB），留出 multipart 余量
        proxy_read_timeout 120s;
        proxy_send_timeout 120s;
        # ⚠️ 子应用走 API 代理接入时，超时请按子应用最慢接口放宽
    }
}
```

> ⚠️ 经反向代理必须：① `server.trusted_proxies` 填 Nginx 地址；② 转发 `X-Real-IP` / `X-Forwarded-For`。
> 否则限流看到的都是 Nginx 的 IP，会误伤全站用户（且伪造头无法被识别）。

---

## 五、子应用接入（IFrame / API 代理）

### 1. 接入前准备

1. 在后台「系统管理 → 应用管理」注册应用，得到 `appCode`：
   - 接入类型 `type` 可选 `iframe` / `proxy`（微应用/qiankun 已移除，不再支持）。
   - `proxy`：填写 `后端入口` 与 `API 前缀`（代理时拼在业务路径前，如 `/api`，留空则直接转发）。
   - `iframe`：填写 `前端入口`。
2. 在「系统管理 → 应用实例」为对应租户**开通**该应用。
3. 在「系统管理 → 菜单管理」为该应用配置菜单（`appCode` 与注册时一致）。

### 2. IFrame 接入（最快）

- 子应用独立部署；菜单 `目标/打开方式 = IFrame` 时，底座用 `views/base/iframe/index.vue` 加载菜单「组件路径」中的地址。
- 底座自动在地址后追加查询参数（已有 query 会以 `&` 追加）：`base_token`、`user_id`、`username`、`tenant_id`。
- 子应用前端解析 `base_token`，调用底座 `GET /business_base/api/auth/info` 校验用户信息。
- 适合已有系统改动成本低的场景。

### 3. API 代理接入（推荐）

- 子应用后端无需暴露公网，由底座反向代理，代理路径：`/business_base/api/app/{appCode}/{原应用API路径}`。
- 实际转发地址 = `{后端入口}{API 前缀}{原应用API路径}`。
  例：`backendUrl=http://127.0.0.1:8081`、`API前缀=/api` 时，`/business_base/api/app/demo/users` → `http://127.0.0.1:8081/api/users`。
- 底座在请求头注入当前用户信息：`X-Base-User-ID`、`X-Base-Username`、`X-Base-Tenant-ID`（平台超管为 `0`）。
- IFrame 类型应用**不支持** API 代理。
- 代理入口鉴权口径：**已登录** + **调用方租户已开通并启用该应用**（平台超管放行）。未开通返回 `403 当前租户未开通该应用`。
- 若希望底座代管子应用接口权限：在「权限管理」登记 `appCode = 子应用编码` 的权限点（`method` + 转发到子应用的业务路径，如 `GET /users/:id`）。
  口径为**按应用启用**：该 app_code 有权限点时，未授权用户被底座拦下（`403 无权限访问该子应用接口`）；一个权限点都没登记则不做接口级校验，完全交给子应用。

### 4. 用户信息标准

| 来源 | 字段 | 说明 |
| --- | --- | --- |
| URL Query | `base_token` | 底座 JWT（仅 iframe 接入，由底座前端自动追加） |
| URL Query | `user_id` / `username` / `tenant_id` | 仅 iframe 接入，便于子应用展示与调试，可信度低于请求头 |
| Header | `X-Base-User-ID` | 用户 ID（代理接入） |
| Header | `X-Base-Username` | 用户名（代理接入） |
| Header | `X-Base-Tenant-ID` | 租户 ID（代理接入，平台超管为 `0`） |

### 5. 适配器示例

- Go / Gin 中间件示例见 `base/integration/adapter-go/middleware.go`：读取底座请求头并写入 gin 上下文，`r.Use(adapter.BaseAuthMiddleware())`。
- TypeScript 侧示例见 `base/integration/adapter-ts/types.ts`：定义用户类型与 `base_token` / 用户信息本地存取工具。
- 本仓库内的 `base/integration/` 只保留适配器示例代码，接入规范即本节内容。

### 6. 安全建议

- 生产环境使用 HTTPS。
- 子应用后端应校验 `X-Base-*` 请求头来源 IP 或签名，防止伪造。
- Token 有效期与底座 JWT 一致（默认 8 小时），子应用需处理 401 并跳转回底座登录页。
- 子应用本身仍需做权限校验，不可仅依赖底座请求头。

### 7. 与既有系统的对接示例（以 caam-portal 为例）

1. 保持 `portal` 的 frontend / backend 独立运行。
2. 底座「应用管理」注册：`code=caam-portal`、`type=proxy`、`backendUrl=http://127.0.0.1:8092`、`API 前缀=/business_portal/api`。
3. 「应用实例」中为目标租户开通该应用。
4. 「菜单管理」新增 `appCode=caam-portal` 的菜单，打开方式设为 `IFrame` 并填写门户前端地址（或按 proxy 方式配置 API 菜单）。
5. 用户即可从底座一站式进入门户。

---

## 六、部署顺序

1. 准备 MySQL / Redis，**手工创建数据库** `caam_base`（InnoDB + utf8mb4），确认账号有建表与 `ALTER`（加索引）权限。
2. 部署后端：改 `config.yaml`（端口、库连接、`jwt.secret`、`server.trusted_proxies` 填 Nginx IP）→ 确认 `config.yaml` 权限收紧 → 编译 → 启动 → 日志出现 `Base server listening on :8091`。
3. 健康检查：`GET /business_base/api/site-info` 返回 200 且 `code=0`。
4. 部署前端：构建 `dist` → 配置 Nginx（子路径与 `VITE_BASE_PATH` 一致）。
5. 登录 `admin / admin123` → 立即修改密码；按需创建租户、注册应用、开通实例、配置菜单与角色权限。
6. 配置「系统设置」：平台名称/版权、安全策略、邮件配置（可先「发送测试邮件」验证）、通知渠道（企微/短信）。
7. 接入第一个子应用并按需登记该应用编码的接口权限点，验证菜单跳转与代理请求。

---

## 七、回滚 / 升级

- **后端**：替换 `base` 可执行文件后重启；结构变更由 `AutoMigrate` 自动迁移（不会删列/删索引）。
- **前端**：替换 Nginx 目录下的 `dist` 内容；建议保留上一版本目录便于快速回退。
- **回滚注意**：新版本新增的表/列/索引无需回退；确需删表回滚请先备份。旧版本不认识新表不会报错，但新版本写入的数据在旧版本下可能不可见。
- **备份建议**：`mysqldump` 数据库 + `backend/uploads` 目录 + `backend/config.yaml`（含明文密码，注意保管）。
- 更换 `jwt.secret` 会使所有已签发 Token 立即失效（在线用户被强制重新登录）。
- 修改限流相关配置需重启后端生效。

---

## 八、常见问题（FAQ）

| 现象 | 原因 / 处理 |
| --- | --- |
| 启动即退出，日志 `加载配置失败` | 未在 `backend` 目录下启动（viper 读的是 `./config.yaml`） |
| 启动即退出，日志 `初始化Redis失败` | Redis 地址/密码/DB 不通或未启动；`redis.db` 需与其它模块区分（当前为 1） |
| 启动即退出，日志 `初始化数据库失败` | MySQL 不通、库不存在、账号无权限；日志若出现 `1451`（外键约束）说明历史软删清理某表失败，属告警级，可继续观察 |
| 服务起不来且日志提示唯一索引/重复用户 | 历史用户名重复导致建唯一索引失败；确认日志中的 `已重命名重复用户名` 行，再次启动即可完成建索引 |
| 登录/验证码返回「请求过于频繁，请稍后再试」（`code=429`） | 命中固定窗口限流（Redis `ratelimit:<路由>:<IP>:<窗口段>`）。优先确认 `server.trusted_proxies` 是否填了 Nginx IP，否则全站共用一个 IP 会被一起限流 |
| 多用户同时被登出、接口返回 `code=401` | Token 过期（`jwt.expire_hours`）或被吊销（登出 / 改密 / 更换 `jwt.secret`）；前端收到 401 自动跳登录页 |
| 用户「改完密码后其他设备被登出」 | 属正常安全机制：改密后写入 `auth:user:revoked-before:<uid>`，此前的 Token 全部失效 |
| 租户管理员打开「系统设置」报 403 / 页面提示仅超管可用 | 系统设置、应用定义、接口权限点为**平台级资源**，仅平台超管可写（设置项读也限超管）；这是设计行为 |
| 新接口只有超管/租户管理员能用，普通用户 403 | 新接口未在 `internal/seed/permission_seed.go` 的 `basePermissionSeeds` 登记权限点，或角色未在「角色管理 → 分配权限」中授权 |
| 上传大文件失败 | 单文件上限 `server.max_upload_mb`（默认 50MB）+ 扩展名白名单（jpg/jpeg/png/gif/pdf/doc/docx/xls/xlsx/ppt/pptx/txt/zip/rar/mp4）；同时确认 Nginx `client_max_body_size` ≥ 该值 |
| 上传报「非法路径 / 路径穿越」 | 文件名含 `../` 等危险字符会被拒（服务端会做文件名净化与目录内校验），改名后重试 |
| 删除菜单/角色/机构失败并提示存在关联 | 属删除保护：存在子节点或被引用时拒绝删除（报错含具体数量），需先处理下级数据 |
| 删除用户后同名重建「变成新记录、ID 变了」 | 本项目统一**物理删除**，唯一索引只被存活记录占用；这是设计行为（旧版「恢复原记录」语义已取消） |
| 工作流节点保存失败 | 该流程存在「审批中」的实例时拒绝保存节点配置，避免流程跑到一半结构变了 |
| 待办长时间无人处理 | 节点可配「超时提醒(分钟)」（0 = 不提醒）；后台按 `workflow_remind_interval_seconds` 扫描并给审批人发站内信，详情页以「催」标签展示 |
| 发送消息时只有「站内信」可选 | 站外渠道未配置：邮件/企微/短信需先在「系统设置」相应页签填写并保存（保存即重载渠道注册表） |
| 子应用菜单打开空白 / 报未配置地址 | 菜单「组件路径」需填写子应用 iframe 地址；`views/base/iframe/index.vue` 会提示「未配置子应用地址」 |
| 子应用代理返回 403 | ① 调用方租户未在「应用实例」开通该应用；② 该 app_code 已登记权限点但当前用户/角色未授权 |
| 页面接口前缀写错导致 404 | 前缀由 `server.api_prefix` 与前端 `VITE_API_BASE_URL` 双端决定，改子路径必须同时改这两处 + Nginx |

---

## 附录 A：安全与隔离实现说明

> 本附录面向开发/运维，说明底座在**接口级权限鉴权**与**租户数据隔离**方面的实现规范。

### A.1 接口级权限鉴权

实现位置：`internal/middleware/permission.go`。

**设计目标**

- 所有需要登录的接口，在 JWT 认证之后，额外校验当前用户是否拥有该接口的访问权限。
- 权限数据来自 `base_permission` 表，通过 `base_role_permission` 与 `base_user_role` 关联到用户。
- 前端返回的权限码仅用于 UI 控制，真正的鉴权在后端完成。

**匹配规则**：使用当前请求的 **HTTP Method + 路由模板** 匹配权限表的 `method` / `path`。

- 路由模板示例：`/business_base/api/users/:id`；
- 权限表 `path` 可配完整路径，也可配相对路径（如 `/users/:id`，种子即用相对路径）；
- 支持 `:param` 通配符，暂不支持 `*any`（子应用代理侧用 `pkg/permmatch` 支持 `*`）。

**放行策略（顺序）**

1. **平台超级管理员**（`tenantID == 0`）直接放行；
2. **租户管理员**（`base_user.is_admin = 1`）在本租户内直接放行；
3. **白名单接口**：用户基础信息、菜单、权限、改密、仪表盘、我的应用、未读消息数，以及工作流「属于自己的数据」类接口（见 A.2）；
4. **未分配任何权限的普通用户**：除白名单外一律 `403 未分配任何接口权限，请联系管理员在「角色管理 → 分配权限」中授权`；
5. 其余请求按权限点匹配，未命中返回 `403 无权限访问该接口`。

> ⚠️ 历史行为（已取消）：早期版本在用户「没有任何权限记录」时**整体放行**，且会放行所有 `GET`，导致零权限用户可读到 `/settings`（含 SMTP 密码）、`/users`、`/operation-logs` 等敏感数据。

**白名单（相对路径）**

- `GET /auth/info`、`GET /auth/menus`、`GET /auth/permissions`
- `POST /auth/change-password`、`POST /auth/logout`
- `GET /dashboard/stats`、`GET /app-instances/my`、`GET /messages/unread-count`
- `GET /workflows/options`、`POST|GET /workflow-instances`、`GET /workflow-instances/:id`、`POST /workflow-instances/:id/cancel`
- `GET /workflow-tasks`、`GET /workflow-tasks/approver-options`、`POST /workflow-tasks/:id/{approve,reject,transfer,add-approver}`

白名单键使用**相对路径**，匹配时通过 `permmatch.Candidates()` 逐个查表，因此写成完整路径也能命中。

**挂载方式**（`internal/routes/routes.go`）

```go
authorized := api.Group("", middleware.JWTAuth())
authorized.Use(middleware.OperationLog())
authorized.Use(middleware.PermissionAuth())
```

**权限配置**

- 权限种子在 `internal/seed/permission_seed.go`（幂等，按 `code` 写入/更新）：18 条 `type=menu` 分组节点 + 94 条 `type=api` 接口权限点（`method` + 相对路径）。
- 全部底座权限点自动授予平台超管角色（`tenant_id=0, code=super_admin`），保证升级后平台管理员不被拦截。
- 新增接口时在 `basePermissionSeeds` 补一条即可，无需手工建数据；普通用户角色的权限在「角色管理 → 分配权限」勾选。
- 手工新增权限记录字段含义：`app_code`（base 或子应用编码）、`code`（如 `base:user:create`）、`name`、`type`（menu/api/button）、`method`、`path`、`status`。

### A.2 工作流为何部分接口在白名单里

审批是「用户自己的事」：被分配了待办的普通用户必须能查看与处理，不应再依赖管理员额外分配接口权限。因此这些接口不做接口级校验，改用**归属校验**保证安全：

| 接口 | 归属校验（service 层） |
| --- | --- |
| 发起流程 | 只能发起本租户（平台超管例外）已启用的流程 |
| 我的申请 / 实例列表 | 非管理员强制 `initiator_id = 当前用户` |
| 实例详情 | 仅发起人、该实例的审批参与人、管理员可见 |
| 撤销 | 仅发起人本人或管理员，且实例必须处于「审批中」 |
| 待办列表 / 审批 / 驳回 | 按 `approver_id = 当前用户` 过滤，并校验任务归属与状态（不能重复处理、不能处理已结束流程） |
| 转办 / 加签 | 只能操作 `approver_id = 当前用户` 的待处理任务；目标必须是**同租户且启用**的用户（平台超管可选全部租户用户），跨租户/已停用/已是本节点待审批人会被拒绝 |
| 转办/加签候选人选项 | 只返回本租户启用用户的 `id/username/realName`，不返回密码等敏感字段 |

**流程定义（含节点编排）**属管理动作，仍走权限点 `base:workflow-def:*`；**删除流程实例**走 `base:workflow-instance:delete`（不在白名单）。

### A.3 工作流超时提醒（后台任务）

- 由 `service.StartWorkflowReminder` 在服务进程内按 `server.workflow_remind_interval_seconds`（默认 600 秒，`<= 0` 关闭）周期扫描，不对外暴露接口，无需额外调度器。
- 只扫描 `status = 待处理` 且所属实例仍在审批中、且 `TIMESTAMPADD(MINUTE, timeout_minutes, created_at) <= NOW()` 的待办，单轮上限 200 条。
- 提醒方式为**站内信**（`SendToUsers(0, "系统", tenantID, ...)`），收件人仅限该任务审批人本人；不涉及邮件/短信渠道，也不回写业务数据。
- 只更新任务的 `reminded_at` / `remind_count` 并写 `remind` 流转日志（同一超时周期内不重复提醒）。

### A.4 租户数据隔离

**核心原则**

- 每个租户只能操作自己租户下的数据；
- 平台超管（`tenantID == 0`）可跨租户管理；
- 按 ID 操作的 `Update` / `Delete` / `GetByID` 必须在 SQL 中追加 `WHERE tenant_id = ?`。

**Service 层模式**

```go
db := db.DB.Model(u)
if tenantID > 0 {
    db = db.Where("tenant_id = ?", tenantID)
}
return db.Updates(updates).Error
```

**写操作必须校验命中**（`internal/service/helper.go`）：GORM 的 `Updates` / `Delete` 命中 0 行时**不返回错误**，直接 `return res.Error` 会出现「提示保存成功但数据未变」。

- 更新前用 `ensureRecordExists(...)` 做存在性 + 归属校验；**不要用 `RowsAffected` 判断更新**（MySQL 默认返回「实际变更行数」，原样保存会得到 0，被误判为失败）。
- 删除后用 `ensureDeleteAffected(...)` 校验 `RowsAffected > 0`。

**Controller 层**：`tenantID := c.GetUint64("tenantID")` 后传入 service。

**已改造模块**（Update / Delete / GetByID 均已按租户隔离）：
User、Role、WorkflowRole、Workflow、WorkflowInstance（Delete 按租户 + Detail 限发起人/参与人/管理员）、Menu、Organization、AppInstance、Message、MessageTemplate、Dict、File。

> Role 的 `AssignMenus` / `AssignPermissions`、User 的 `AssignRoles` / `ResetPassword`、WorkflowRole 的 `AssignUsers`（成员必须与角色同租户）与 `ListUserOptions`、以及各模块 `List` / `Tree` 均已按当前租户过滤。

**List 查询与跨租户管理**

```go
if tenantID > 0 {
    query = query.Where("tenant_id = ?", tenantID)       // 普通租户：只能看本租户
} else if filterTenantID > 0 {
    query = query.Where("tenant_id = ?", filterTenantID) // 平台超管：按 tenantId 查询参数过滤
}                                                        // 平台超管不传 tenantId：查看全部租户
```

**创建时的目标租户（重要）**：统一使用 `controllers.resolveTenantID(c, requested)`——平台超管可用请求体 `tenantId` 指定目标租户（0 表示平台级）；普通租户用户**忽略请求中的 `tenantId`**，强制写入自身租户。
涉及接口：`POST /users`、`POST /roles`、`POST /app-instances`、`POST /dicts`、`POST /organizations`、`POST /message-templates`、`POST /workflow-roles`、`POST /workflows`。

**不需要租户隔离的模块**

- `Tenant`：本身即租户管理入口，路由已绑 `middleware.SuperAdminOnly()`，只有平台超管能访问；
- 平台级资源（无 `tenant_id` 字段）：`App`（应用定义）、`Permission`（权限点）、`Setting`（系统设置）。

**读取可见性口径**：字典 / 机构 / 消息模板 / 菜单 / 消息 = `tenant_id = 自身 OR tenant_id = 0`；用户 / 角色 / 文件 / 应用实例 = 严格 `tenant_id = 自身`；写操作（Update/Delete/GetByID）一律严格。

### A.5 登录验证码

实现：`internal/service/captcha_service.go`（与 portal / member / application 一致）。

- **样式**：`base64Captcha.NewDriverString`，5 位数字 + 大写字母，字符集 `234679ACDEFGHJKMNPQRTUVWXY`（已剔除 0/O、1/I/L、2/Z、5/S、8/B 等易混字符），带空心线/粘连线/正弦线干扰；图片 300×100，对应前端 150×50 容器 + `object-fit: cover`。
- **存储**：仅 Redis（key `captcha:<id>`，TTL 5 分钟），多实例部署校验一致。
- **校验**：一次性（无论对错校验后立即删除）、大小写不敏感、忽略首尾空格；`id` 或 `code` 为空直接失败。
- **接口**：`GET /business_base/api/auth/captcha` 返回 `{captcha_id, captcha_img}`；登录体使用 `captcha_id` + `captcha_code`。
- **开关**：由「系统设置 → 安全策略 → 登录验证码」控制（`settings.security.captchaEnabled`）。登录页先调 `GET /site-info` 拿 `captchaEnabled`，关闭时不渲染验证码框、不提交验证码字段，后端同样不校验。
- **失败处理**：验证码错误或密码错误均计入 `login_fail:<tenantID>:<username>`，阈值/窗口由安全策略决定（默认 5 次 / 30 分钟，TTL = 锁定时长）。

### A.6 安全策略（系统设置）

「系统设置 → 安全策略」写入 `base_setting`（`category = security`，值以字符串存储），由 `service.SettingsService{}.GetSecuritySettings()` 统一读取（带默认值与范围限制）：

| 设置项 | Key | 默认值 | 作用 |
| --- | --- | --- | --- |
| 登录验证码 | `captchaEnabled` | `true` | 关闭后登录接口不再校验验证码，登录页隐藏验证码框 |
| 登录失败锁定 | `loginLock` | `true` | 是否启用连续失败锁定 |
| 最大失败次数 | `maxFailCount` | `5` | 达到即锁定（范围 3~20） |
| 锁定时长(分钟) | `lockDuration` | `30` | 锁定时间，同时是失败计数 Redis key 的 TTL（范围 1~1440） |
| 密码最小长度 | `pwdMinLength` | `8` | 新增用户 / 重置密码 / 修改密码时服务端强校验（范围 6~32） |

**要点**

- 登录、验证码锁定、用户创建三个流程共用同一份配置，避免口径不一。
- **密码不存在隐式默认值**：新增用户必须显式填写初始密码，`POST /users/:id/reset-password` 也必须由管理员传入新密码（旧版会隐式用 `123456`，与最小长度策略自相矛盾）。
- 设置为空或非法时回退默认值（数值会被 clamp 到合法区间），旧环境无需预先写入任何记录。
- 前端「安全策略」表单必须与后端 key 对齐（`api/setting.ts` + `views/base/setting/index.vue`）。

### A.7 公开接口限流

实现：`internal/middleware/rate_limit.go`，基于 Redis 固定窗口。

| 接口 | 限制 | 配置项 |
| --- | --- | --- |
| `POST /business_base/api/auth/login` | 10 次/分钟 | `login_rate_limit` / `login_rate_window_seconds` |
| `GET /business_base/api/auth/captcha` | 30 次/分钟 | `captcha_rate_limit` / `captcha_rate_window_seconds` |
| `GET /business_base/api/site-info` | 30 次/分钟 | 同上（与验证码共用额度参数） |
| `POST /business_base/api/auth/init` | 5 次/分钟 | `init_rate_limit` / `init_rate_window_seconds` |

**要点**

- 计数 key：`ratelimit:<路由模板>:<客户端IP>:<窗口段>`，**每个接口各自独立配额**（portal 的实现不含路由段，同 IP 下各接口共用一个计数，base 未采纳）。
- 超限返回 `{code:429, message:"请求过于频繁，请稍后再试"}`。
- Redis 不可用时退化为**进程内固定窗口限流**（fail-closed），避免限流组件故障时被无限刷量。
- 客户端 IP 走 gin `ClientIP()`：仅当请求来自 `server.trusted_proxies` 时才信任 `X-Forwarded-For` / `X-Real-IP`，否则用 `RemoteAddr`。**部署在 Nginx 后必须填代理真实地址**。

### A.8 Token 吊销（登出 / 改密）

实现：`internal/service/token_service.go` + `internal/middleware/auth.go`，全部状态存 Redis，不新增数据表：

| Key | 写入时机 | 含义 |
| --- | --- | --- |
| `auth:token:revoked:<jti>` | `POST /auth/logout` | 该 token 已登出，TTL = 剩余有效期 |
| `auth:user:revoked-before:<userID>` | 修改密码成功后 | 该时刻之前签发的 token 全部失效，TTL = JWT 有效期 |

- `JWTAuth` 解析后校验 `jti` 黑名单与用户级失效时间，命中返回 `401 登录状态已失效，请重新登录`。
- Redis 不可用时 **fail-open**（放行并记 warn）：宁可短暂失去吊销能力，也不能让全部用户无法登录。
- 前端 `stores/user.ts` 的 `logout()` 会尽力调用登出接口（显式携带 token）后清理本地会话；拦截器遇 401 只调 `clearSession()`（**不要再调 logout()，会递归**）。

### A.9 写操作保护（删除 / 引用）

- **引用校验**：租户（用户/角色/应用实例/流程角色/流程定义/流程实例）、应用（应用实例）、机构（子机构）、菜单（子菜单）、权限（子权限）在存在下级或引用时拒绝删除，并返回具体数量。
- **级联清理**：删除用户清理 `base_user_role` / `base_workflow_role_user`；删除角色清理 `base_role_menu` / `base_role_permission` / `base_user_role`；删除菜单清理 `base_role_menu`；删除权限清理 `base_role_permission`；删除流程角色同时删 `base_workflow_role_user`。均在事务内完成。
- **账号保护**：不能删除当前登录用户；平台内置 `admin` 账号与 `tenant_id=0, code=super_admin` 角色不可删除。
- **应用实例唯一**：同一租户同一应用不可重复开通（应用层校验，删除后可重新开通）。
- **用户名唯一**：`uk_base_user_tenant_username (tenant_id, username)`——同租户内不可重名、跨租户可同名（与登录按「租户编码 + 用户名」定位一致）；service 层预检查只给友好提示，**并发由索引兜底**（MySQL 1062 转成「该租户下用户名已存在」，不暴露 SQL 细节）；启动时 `dedupeUserUsernames` 先整理历史重复数据，否则建索引会失败导致服务起不来。

### A.10 文件上传

- 单文件大小上限 `server.max_upload_mb`（默认 50）：控制器按 header 快速拒绝，存储层再用 `io.LimitReader` 流式兜底并删除超限的临时文件。
- 扩展名白名单：见 `FileService.IsAllowedType`（jpg/jpeg/png/gif/pdf/doc/docx/xls/xlsx/ppt/pptx/txt/zip/rar/mp4）。
- 文件名净化：`storage.sanitizeFileName` 去掉目录部分与危险字符/控制字符，避免 `../` 穿越与超长文件名。
- 存储根目录初始化时转**绝对路径**（`filepath.Abs`），读写（`Resolve`）与 `GET /files/*key` 都校验结果位于根目录内，不依赖进程工作目录。

### A.11 审计日志

- 含密码的接口（`POST /auth/change-password`、`POST /users/:id/reset-password`、`PUT /settings`）**不记录原始 body**，只写 `[敏感参数已脱敏]`。
- 请求体/响应体写入长度上限 4KB（超出截断并标注），避免大响应把审计表撑爆。
- 日志写入为**同步**执行（异步 goroutine 在进程退出/重启时会丢日志）；`module` 记录 gin 路由模板，`action` 记录 `METHOD 原始URL`。
- 导出 CSV 带 UTF-8 BOM 并按 RFC4180 转义（中文在 Excel 不乱码、字段含逗号不会错列）。

### A.12 平台级资源与子应用代理鉴权

**平台级资源只写于平台超管**（`middleware.SuperAdminOnly()`）：

| 资源 | 路由 | 说明 |
| --- | --- | --- |
| 系统设置 | `/settings`（GET/PUT/email/test） | SMTP、企微 webhook、短信网关、安全策略 |
| 应用定义 | `/apps`（POST/PUT/DELETE） | 租户仍可 `GET /apps` 查看可开通的应用 |
| 接口权限点 | `/permissions`（POST/PUT/DELETE） | 权限点是全局基线，租户管理员不得改动 |
| 菜单 | `/menus` | 平台内置菜单（tenant_id=0）对租户只读，写入限本租户（由 `tenant_id` 条件 + 命中校验保护） |

**通知渠道**

- 发送器注册表：`pkg/notifier`（email / wechat / sms），启动与「系统设置保存」后由 `service.ReloadNotifiers()` 重新注册。
- 未配置的渠道被 `Unregister`，`GET /messages/channels` 也就不返回，前端渠道下拉自然不出现。
- 短信采用「HTTP 网关」约定（`POST {gateway}` + `Authorization: Bearer` + `{to,sign,subject,content}`），底座不内置厂商 SDK，替换服务商只需改配置。
- 邮件/短信需选定具体收件人（不支持全员广播），企业微信为群推送。
- 注意：短信 / 企微渠道**没有** sender 概念，仅按通用约定推送。

**子应用代理鉴权**（`/business_base/api/app/:appCode/*path`，只挂 JWT，handler 内依次校验）：

1. 应用已注册且启用，且类型不是 iframe；
2. 调用方租户已开通并启用该应用（平台超管放行）；
3. **按应用启用**的接口权限：该 `app_code` 在 `base_permission` 中有登记时，用户需持有匹配权限点（路径匹配与 base 接口共用 `pkg/permmatch`，支持 `:param` / `*`）；未登记任何权限点时不校验。

> ❌ 新增子应用后若要做底座代管权限，请在「权限管理」登记 `appCode = 子应用编码` 的权限点。

### A.13 工作流引擎实现说明

实现位置：`internal/service/workflow_service.go`（定义与节点）、`internal/service/workflow_engine_service.go`（引擎与查询）。

- **节点串行 + 或签/会签**：节点按 `sort` 升序执行；一个节点解析出多个审批人时每人一条待办。`approve_mode = or`（默认）任一人处理即视为节点完成，**同节点其余待办自动失效**；`and`（会签）必须**所有待办都通过**才推进，任一驳回即节点驳回。审批方式会**快照**到任务行，定义后续改动不影响进行中的实例。
- **审批人解析**（`resolveApprovers`）：`role` → 流程角色成员（仅 `status=1` 用户）、`user` → 指定用户、`initiator` → 实例发起人。
- **自动通过**：节点解析不到有效审批人时直接跳过并写 `auto-pass` 日志，避免流程卡死。
- **推进**（`advanceWorkflow`）：通过后从当前节点顺序找下一个有审批人的节点；无后续节点则实例置「已通过」并写 `finish` 日志。
- **转办**：待办交接给同租户启用用户，原审批人不再持有该任务，写 `transfer` 日志。
- **加签**：为当前节点追加一名审批人（继承节点 `approve_mode`）形成临时会签，写 `add-approver` 日志。
- **超时提醒**：见 A.3。
- **驳回 / 撤销**：驳回后实例置「已驳回」，其余未处理待办置「已失效」；撤销仅发起人本人或管理员可执行。
- **删除**：仅管理员（`base:workflow-instance:delete`），物理删除实例及其任务与流转日志。
- **可见性**：详情允许发起人、参与审批的人、管理员查看；列表对非管理员强制 `initiator_id = 自己`。
- **节点改动保护**：`PUT /workflows/:id/nodes` 在存在进行中实例时拒绝保存；保存时归一化（`approve_mode` 非 `and` 一律按 `or`，`timeout_minutes` 限制 0~10080 分钟）。
- **时间字段**：`end_at`、`handled_at`、`read_at`、`send_at`、`reminded_at` 均为 `*time.Time`，未发生时写 `NULL`，避免 MySQL 严格模式（`NO_ZERO_DATE`）拒绝 `0000-00-00`。

### A.14 仪表盘口径

`service/dashboard_service.go` 的 `GetStats(tenantID, userID, isAdmin)`：

| 角色 | 资源概览 | 「我的」指标 | 趋势（近 7 天） |
| --- | --- | --- | --- |
| 平台超管 | 全平台：租户数/启用租户、应用数（iframe / proxy）、应用实例数、用户数（启用/停用/管理员）、角色数、机构数、消息数 | 未读、待我审批、我发起的进行中流程 | 新增用户 / 登录成功 / 消息 |
| 租户管理员（`is_admin`） | 本租户（机构/消息含 `tenant_id = 0` 的平台共享数据） | 同上 | 同上（限本租户） |
| 普通用户 | 不返回（保持 0） | 同上 | 不返回 |

- 趋势按天聚合最近 7 天（含今天），无数据补 0；驱动返回的日期可能带时分秒，统一截前 10 位。
- 今日统计：今日新增用户 / 今日登录 / 今日登录失败。

### A.15 开发 checklist

新增一个需要按 ID 操作的接口时，请确认：

- [ ] 模型是否包含 `tenant_id` 字段。
- [ ] `Update`/`Delete`/`GetByID` 是否接收并使用了 `tenantID` 参数。
- [ ] Controller 是否从 `c.GetUint64("tenantID")` 取当前租户并传入 service。
- [ ] 创建接口是否使用 `resolveTenantID(c, requested)`，而不是直接赋值 `c.GetUint64("tenantID")`。
- [ ] 判断「平台超管」是否统一使用 `models.IsPlatformTenant(tenantID)`，不要内联 `tenantID == 0`。
- [ ] 读取可见性是否符合约定：字典/机构/消息模板/菜单/消息为 `tenant_id = 自身 OR tenant_id = 0`，用户/角色/文件/应用实例严格等于自身。
- [ ] 列表接口：租户用户只看自己，超管默认全部并支持 `tenantId` 查询参数过滤。
- [ ] 跨租户请求是否返回「资源不存在」或 403，而不是泄露其它租户数据。
- [ ] 该接口是否已在 `internal/seed/permission_seed.go` 的 `basePermissionSeeds` 中登记。
- [ ] 若前端要控制按钮显隐，权限点 `code` 是否与 `userStore.can('base:xxx:yyy')` 完全一致（超管/租户管理员由 `can()` 自动放行）。
- [ ] 更新前是否校验记录存在/归属（`ensureRecordExists`），删除后是否校验命中行数（`ensureDeleteAffected`）。
- [ ] 新的删除接口是否需要引用校验与关联表清理（参考 Role / User / Menu / Permission 的实现）。
- [ ] 是否需要在 Redis 中做失效处理（如登出/改密后的 token 吊销）。
- [ ] 新增的是否为「平台级资源」：若是，写入接口应加 `middleware.SuperAdminOnly()`，并同步前端按钮的 `isSuperAdmin` 控制。
- [ ] 涉及关联关系时（如用户 → 机构）是否补充了删除前的引用校验。

### A.16 已知遗留（未实现 / 需注意）

- 短信 / 企业微信渠道无 sender（用 HTTP 网关 / 群机器人替代）；无 refresh token（8h JWT + 吊销已够）。
- 前端部分动作没有权限码（依赖后端归属校验或白名单）：文件页「预览」、消息页「查看」、流程实例页「发起流程」/「撤销」/「详情」、我的待办页「通过/驳回/加签/转办」、字典项「新增项」。
- `base_message_template` 的「短信 / 企微」渠道在前端模板表单未开放（渠道下拉仅 `in-app` / `email`）。
- 列表页统一固定每页 10 条，无每页条数选择与排序。

---

## 附录 B：后端接口清单

前缀统一为 `server.api_prefix`（默认 `/business_base/api`），下表以**相对路径**列出。

### 公开接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/auth/login` | 登录（支持租户编码、验证码），10 次/分钟 |
| POST | `/auth/init` | 首次部署创建超级管理员（已存在则报错），5 次/分钟 |
| GET | `/auth/captcha` | 图形验证码，返回 `captcha_id` / `captcha_img`，30 次/分钟 |
| GET | `/site-info` | 站点公开配置（`captchaEnabled`），30 次/分钟 |
| GET | `/files/*key` | 文件公开访问（key 含日期目录，如 `20260101/xxx.png`） |

### 登录后接口（JWT + 操作审计 + 接口权限校验）

| 分组 | 方法 | 路径 | 说明 |
| --- | --- | --- | --- |
| 认证 | GET/POST | `/auth/info`、`/auth/menus`、`/auth/permissions`、`/auth/change-password`、`/auth/logout` | 当前用户、菜单、权限标识、改密（旧 token 失效）、登出（token 进黑名单） |
| 仪表盘 | GET | `/dashboard/stats` | 资源概览 + 今日统计 + 「我的」+ 近 7 天趋势 |
| 租户 | CRUD | `/tenants` | 仅平台超管（路由级 `SuperAdminOnly`） |
| 系统设置 | GET/PUT/POST | `/settings`、`/settings/email/test` | 仅平台超管；保存后立即重载通知渠道 |
| 数据字典 | CRUD | `/dicts`、`/dicts/code/:code`、`/dicts/:id/items` | 字典与字典项 |
| 应用 | CRUD | `/apps` | 平台级：租户可读，仅平台超管可增删改 |
| 应用实例 | CRUD | `/app-instances`、`/app-instances/my` | 开通/编辑/删除、我的应用 |
| 用户 | CRUD | `/users`、`/users/:id/roles`、`/users/:id/reset-password` | 分配角色、重置密码（需传入新密码） |
| 角色 | CRUD | `/roles`、`/roles/:id/menus`、`/roles/:id/permissions` | 分配菜单、分配权限 |
| 菜单 | CRUD | `/menus`、`/menus/tree` | 菜单树 |
| 权限 | CRUD | `/permissions`、`/permissions/tree` | 平台级：仅平台超管可增删改 |
| 流程角色 | CRUD | `/workflow-roles`、`/workflow-roles/user-options`、`/workflow-roles/:id/users` | 审批角色定义与成员（覆盖式设置） |
| 流程定义 | CRUD | `/workflows`、`/workflows/:id/nodes`、`/workflows/options`、`/workflows/approver-options` | 含节点覆盖式保存与审批人候选 |
| 流程实例 | POST/GET/DELETE | `/workflow-instances`、`/workflow-instances/:id`、`/workflow-instances/:id/cancel` | 发起、列表（非管理员只看自己）、详情、撤销、删除（仅管理员） |
| 审批任务 | GET/POST | `/workflow-tasks`、`/workflow-tasks/approver-options`、`/workflow-tasks/:id/{approve,reject,transfer,add-approver}` | 待办/已办（`box=todo\|done`）与审批动作 |
| 机构 | CRUD | `/organizations`、`/organizations/tree` | 机构树 |
| 消息 | CRUD | `/messages`、`/messages/send`、`/messages/:id/send`、`/messages/channels`、`/messages/unread-count`、`/messages/read-all`、`/messages/:id/read` | 草稿/编辑/发送、可用渠道、未读数、已读 |
| 消息模板 | CRUD | `/message-templates` | 模板管理（`{{变量}}` 渲染） |
| 操作日志 | GET/POST | `/operation-logs`、`/operation-logs/delete`、`/operation-logs/clear`、`/operation-logs/export` | 列表、批量删除、清理（保留天数）、导出 CSV |
| 登录日志 | GET/POST | `/login-logs`、`/login-logs/delete`、`/login-logs/clear`、`/login-logs/export` | 同上 |
| 文件 | POST/GET/DELETE | `/files/upload`、`/files`、`/files/:id` | 上传、列表、删除 |
| 子应用代理 | ANY | `/app/:appCode/*path` | 统一反向代理入口（详见第五、A.12 节） |

---

## 附录 C：前端工程与开发说明

### C.1 技术栈与脚本

Vue 3.4 + Vite 5 + Element Plus 2.6（`locale: zhCn`，全局注册全部图标）+ Pinia 2.1 + TypeScript 5.4 + axios。

```bash
npm run dev      # 启动开发服务器（3000）
npm run build    # vue-tsc && vite build
npm run preview  # 预览生产构建产物
npm run lint     # ESLint 自动修复
```

> `package.json` 中 `name = business-base-frontend`。构建前置的类型检查失败会直接中断构建。

### C.2 目录说明

```
frontend/
├── src/
│   ├── api/            # 接口模块（auth/user/role/menu/permission/org/tenant/app/dict/file/log/login-log/message/setting/dashboard/workflow/workflowRole）
│   ├── components/     # Breadcrumb.vue / TenantSelect.vue / WorkflowDetail.vue / TrendCard.vue / StatCard.vue
│   ├── router/         # 动态路由生成（菜单驱动）
│   ├── stores/         # user.ts（登录态/菜单/权限/can()）、app.ts（我的应用）
│   ├── styles/         # 全局样式
│   ├── utils/          # request.ts（axios 封装）、tenantOptions.ts、appEntry.ts
│   └── views/
│       ├── base/       # 底座管理页面（dashboard / tenant / app / app-instance / user / organization /
│       │               #   role / menu / permission / dict / file / log / login-log / setting /
│       │               #   message / message-template / workflow-role / workflow/{definition,instance,task} / iframe）
│       ├── layout/     # 布局与弹出式菜单
│       ├── login/      # 登录页
│       ├── profile/    # 个人中心
│       └── error/      # 404
├── .env                # VITE_BASE_PATH / VITE_API_BASE_URL
├── vite.config.ts      # base、别名 @、dev 端口 3000、代理 {apiBase} → 127.0.0.1:8091
└── package.json
```

### C.3 请求封装约定

- `utils/request.ts`：`baseURL = VITE_API_BASE_URL || '/business_base/api'`，超时 30s，请求拦截自动加 `Authorization: Bearer <token>`。
- 响应拦截：`responseType === 'blob'`（导出）**直接返回原始数据**，跳过 `{code}` 校验；`code !== 0 && code !== 200` 统一 `ElMessage.error`；`code === 401` → `clearSession()` + 跳 `${BASE_URL}login`（**不要再调 logout()，会递归**）。
- 后端统一 `{code,message,data}`，失败也返回 HTTP 200，因此**不要用 HTTP 状态码判断业务成败**。

### C.4 动态路由

- 路由基路径 `createWebHistory(import.meta.env.BASE_URL)`；`constantRoutes` 只有 `/login`，其余由 `GET /auth/menus` 生成。
- 菜单 `type` 为 `directory` 生成嵌套路由（无组件）；`menu` 按 `component` 字段解析 `.vue`；
  `target === 'iframe'` 时用 `views/base/iframe/index.vue` 加载，地址存 `meta.url` 并自动追加 `base_token` 等参数。
- 路由守卫：`/auth/info` 失败 → 清会话回登录页；`/auth/menus` 失败 → 提示「菜单加载失败」但**继续进入系统**；**菜单为空 → 提示「当前账号未分配菜单权限，请联系管理员」并落到 `/profile`**。
- 因此：非管理员租户用户要看到菜单，必须由管理员在「角色管理」为其角色分配菜单（并分配接口权限，否则只有白名单接口可用）。
- 菜单勾选「页面缓存」后，Layout 用 `<keep-alive :include="cachedViewNames">` 保活；组件名由 `withComponentName()` 注入菜单名（`index.vue` 的推断名不唯一）。

### C.5 新增管理页面

1. 在 `src/views/base/` 下新建页面，如 `src/views/base/demo/index.vue`；
2. 「菜单管理」新增菜单，组件路径填 `base/demo/index.vue`，应用编码填 `base` 或对应子应用编码；
3. 后端在 `internal/seed/permission_seed.go` 登记接口权限点（否则只有超管/租户管理员可用）；
4. 给角色分配该菜单与接口权限，刷新即可生效。

### C.6 按钮级权限

- 页面里 `const can = (code: string) => userStore.can(code)`，模板 `v-if="can('base:user:delete')"`；
- `userStore.can(code)` = `userInfo.tenantId === 0 || userInfo.isAdmin || permissions.includes(code)`（与后端 `PermissionAuth` 同口径）；
- 平台级页面（应用管理、权限管理）额外 `canWrite = isSuperAdmin && can(code)`；
- 权限标识必须与 `permission_seed.go` 的 `code` 完全一致；租户管理页无权限点（仅超管可访问）。

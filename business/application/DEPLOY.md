# 项目申报评审系统（XXXX-application）部署说明

> 分支：`caam_release`（汽车版）／ `miic_release`（中心版）｜ 项目路径：`business/application/`
> 部署拓扑：Nginx 承载前端静态资源并反代后端 → Go 后端（默认 `0.0.0.0:8094`，配置项 `server.port`） → MySQL / Redis
> 说明：后端 = 申报评审 API；前端 = 「申报人端 + 管理端」同一个工程（两个登录入口，见第三节）。证书为数据记录 + 附件上传，不生成 PDF。
> 文档与代码同步至：2026-09-18（原 `README.md` 内容已并入本文档）

## 目录结构

```
business/application/
├── DEPLOY.md         # 本文档（部署 / 升级 / FAQ）
├── backend/          # Go 后端（Gin + GORM），module: application
│   ├── config.yaml   # 唯一配置文件（从进程工作目录读 ./config.yaml，须在 backend 目录下启动）
│   ├── docs/         # 数据库结构等文档（docs/database.md）
│   ├── logs/         # 日志目录（默认 logs/application.log）
│   ├── uploads/      # 上传文件目录（运行时需可写）
│   └── comp.bat      # Windows 交叉编译脚本（goos=linux，产物 ./application）
└── frontend/         # 申报人端 + 管理端（Vue3 + TS + Element Plus）
    ├── docs/         # 用户使用手册（docs/用户使用手册.md）
    └── .env          # 部署子路径 VITE_BASE_PATH 与接口前缀 VITE_API_BASE_URL
```

> 文档维护边界：`DEPLOY.md`（部署/升级/FAQ）、`backend/docs/database.md`（表结构、枚举、种子数据、约束口径）、`frontend/docs/用户使用手册.md`（界面操作与业务口径，章节与菜单目录一致）。改状态流转、上传限制、证书/公示口径时**三处都要同步**。

---

## 一、环境依赖

| 组件 | 版本/说明 |
| --- | --- |
| Go | ≥ 1.27（`backend/go.mod` 声明 `go 1.27.1`），交叉编译 Linux 需 `CGO_ENABLED=0` |
| Node.js | ≥ 18（Vite 5；`frontend/package.json`） |
| MySQL | ≥ 8.0（InnoDB + utf8mb4）库名 `XXXX_application`，当前仓库配置指向 `10.3.1.95:63400` 的 `caam_application`；**库必须事先建好**（程序只连库、不建库） |
| Redis | 端口 6379，**两个库分工固定**：`captcha_db`(2) 验证码、`anti_replay_db`(3) 限流计数；**Redis 不可达会直接 panic 退出** |
| Nginx | 托管前端静态文件并反代 API；**需把 Nginx 地址填入 `server.trusted_proxies`**，否则限流会把所有用户算成同一个 IP |

---

## 二、后端部署（backend）

### 1. 配置

编辑 `backend/config.yaml`（敏感项用环境变量覆盖，见下一节）：

- `server.port`: 监听端口，默认 `8094`（监听地址在代码中固定为 `0.0.0.0`）；改端口后需同步 Nginx 与 `frontend/vite.config.ts` 的 dev 代理目标
- `server.mode`: **当前仓库值为 `debug`，生产必须改为 `release`**（无环境变量可覆盖，需改配置文件）
- `server.api_prefix`: `/business_application/api`（必须与前端 `VITE_API_BASE_URL` 一致；留空回退 `/application/api`）
- `server.upload_dir_prefix`: `/business_application`（上传挂载 = `<该值>/uploads`，需与 `VITE_BASE_PATH` 去尾斜杠一致）
- `server.max_concurrent_ips`: 单个 IP 的并发请求上限（默认 100），超限返回 `code=429`
- `server.trusted_proxies`: **可信反向代理地址，生产必须填 Nginx 的 IP**（`c.ClientIP()` 取真实 IP 的依据；默认仅 `127.0.0.1`）
- `mysql`: host / port / user / `password`（**只填占位值 `APPLICATION_DB_PASSWORD`**）/ db_name / charset(`utf8mb4`) / max_open / max_idle
- `redis`: addr / password / `captcha_db`(2) / `anti_replay_db`(3)
- `jwt`: `secret`（占位值 `APPLICATION_JWT_SECRET`）/ `expire_hours`（默认 24）/ `issuer`（`caam-application`）
- `log`: `path`（默认 `logs/application.log`）/ `max_size`(100MB) / `max_backups`(30) / `max_age`(180 天)

> 注意：YAML 中不要出现重复 key（viper 解析会直接报错）。

### 2. 敏感信息用环境变量注入（必填）

数据库密码与 JWT 密钥已改为从环境变量读取，`config.yaml` 中仅保留占位值，请勿把真实密码/密钥提交到仓库。

| 环境变量 | 必填 | 说明 |
| --- | --- | --- |
| `APPLICATION_DB_PASSWORD` | 是 | 数据库密码，覆盖 `mysql.password` |
| `APPLICATION_JWT_SECRET` | 是 | JWT 密钥（建议 ≥32 位随机串），覆盖 `jwt.secret` |

```bash
# Linux / macOS
export APPLICATION_DB_PASSWORD='<数据库密码>'
export APPLICATION_JWT_SECRET='<≥32位随机密钥>'

# Windows PowerShell
$env:APPLICATION_DB_PASSWORD='<数据库密码>'
$env:APPLICATION_JWT_SECRET='<≥32位随机密钥>'

# Windows CMD
set APPLICATION_DB_PASSWORD=<数据库密码>
set APPLICATION_JWT_SECRET=<≥32位随机密钥>
```

> 未设置（或仍是占位值）时程序会打印告警后**直接退出**，生产环境必须注入。
> 更换 `APPLICATION_JWT_SECRET` 会使已签发的申报人 Token 与管理端 Token 全部失效。

### 3. 编译

```bash
cd business/application/backend
# Windows 本地交叉编译 Linux 可执行文件（等价 comp.bat）
set goos=linux && go build -o application main.go
# 或直接在目标机编译：
go build -o application main.go
```

仓库根目录的批量脚本可一并编译四个后端（默认 linux/amd64、`CGO_ENABLED=0` 静态二进制，编译后校验产物为对应架构的 ELF）：

```powershell
powershell -ExecutionPolicy Bypass -File .\build-backends.ps1 -Only application -Vet
```

### 4. 启动

```bash
./application
```

- 启动流程：读 `./config.yaml` → 注入环境变量（缺失直接退出） → 连 MySQL（**不建库**） → 连 Redis 两个库（Ping 失败 panic） → 初始化验证码存储 → 初始化 IP 并发限制器 → `AutoMigrate` 建表（14 张，`application_` 前缀，含专家库） → 写入种子数据 → 启动后台任务（每 10 分钟自动关闭报名期已过的批次） → 监听 `0.0.0.0:8094`
- 种子数据（幂等）：角色 `super_admin` / `manager` / `reviewer`（权限 upsert）、后台管理员账号、系统配置
- 默认后台账号：`admin` / `manager` / `reviewer`，初始密码 **`1qaz@WSX`**，**上线后立即修改**
- **两套账号互不通用**：后台账号在 `application_admins`（`POST /admin/login`，前端 `/business_application/admin/login`）；申报人账号在 `application_users`（自助注册 `POST /member/register`，`POST /member/login`）。后台 `admin` 在申报人登录页登录必然失败，属预期行为
- 静态资源上传目录：`./uploads`（后端以 `/business_application/uploads` 提供）
- 日志输出：`logs/application.log`（单文件 100MB、保留 30 个备份、180 天）
- 健康检查（公开接口，无需 Token）：`GET /business_application/api/captcha` 返回 `{"code":0,...}` 即正常
- 响应约定：**HTTP 状态恒为 200**，业务结果看 `code`：`0` 成功、`400` 参数错误、`401` 未登录/登录已过期、`403` 无操作权限、`404` 资源不存在、`429` 请求过于频繁、`500` 服务端错误

### 5. 上传与限流

- 上传扩展名白名单：`.pdf .doc .docx .xls .xlsx .ppt .pptx .txt .zip .rar .jpg .jpeg .png .gif`
- 落盘路径 `uploads/YYYYMMDD/<uuid>.<ext>`，接口返回带部署前缀的 `fileUrl`
- **后端硬限单个文件 ≤ 50MB**（`controllers/upload_controller.go`，超出返回 `code=400`「文件大小不能超过 50MB」）。Nginx `client_max_body_size` 需 ≥ 60m，留出 multipart 开销；配得比 50m 小会先在网关返回 413，看不到后端的业务提示
- 限流（Redis db 3 固定窗口，按真实客户端 IP 计数；不同接口使用不同 `scope`，额度互不影响；Redis 异常时退化为进程内限流 fail-closed）：

| 接口 | 额度 |
| --- | --- |
| `GET /captcha` | 30 次/分钟 |
| `POST /member/register` | 10 次/分钟 |
| `POST /member/login` | 10 次/分钟 |
| `POST /admin/login` | 10 次/分钟 |
| `POST /member/upload` | 20 次/分钟 |
| `POST /admin/upload` | 20 次/分钟 |

- 另有单 IP 并发请求上限 `server.max_concurrent_ips`（默认 100，**必须 > 0**：配成 0 会让所有请求都返回 429）；上述限制超限均返回 `code=429`「请求过于频繁，请稍后再试」

---

## 三、前端部署（frontend）

### 1. 构建

```bash
cd business/application/frontend
npm install
npm run build        # 等价 vue-tsc && vite build
```

也可用仓库根目录脚本（构建后自动校验 `dist/index.html` 里的资源前缀与 `VITE_BASE_PATH` 一致）：

```powershell
powershell -ExecutionPolicy Bypass -File .\build-frontends.ps1 -Only business_application
```

构建产物：`frontend/dist/`

### 2. 关键配置

- `.env`：`VITE_BASE_PATH=/business_application/`、`VITE_API_BASE_URL=/business_application/api`（当前仓库值）
- 若部署路径变化，改 `VITE_BASE_PATH` 后重新构建；同时同步 `backend/config.yaml` 的 `server.api_prefix`（`/xxxx/api`）与 `server.upload_dir_prefix`（`/xxxx`）
- dev 端口 `3003`；dev 代理键由 `.env` 推导（`[apiBase]` 与 `[basePath]/uploads`），**代理目标写死在 `frontend/vite.config.ts`**（当前 `http://127.0.0.1:8094`），改后端端口必须同步改这里
- 申报人端与管理端是**同一个工程**：申报人页面走 `PublicLayout`，管理端页面在 `/admin/*` 路由下并有独立登录页；两套 Token 分别存 `localStorage` 的 `application-member-token` / `application-admin-token`（`utils/request.ts` 按请求路径自动选择），返回 `code=401` 时按请求角色跳对应登录页（申报人 `/login`、管理端 `/admin/login`）并只清理本应用的键
- 依赖：Vue 3.4 / TypeScript 5.4 / Element Plus 2.6 / Pinia 2.1 / Vue Router 4.3 / ECharts 6 / Vite 5.2

### 3. 部署

将 `dist/` 内容复制到 Nginx 站点目录下的 `business_application/` 子目录（见下节示例）。

---

## 四、Nginx 反向代理示例

```nginx
server {
    listen 80;
    server_name application.example.com;

    # 前端静态资源（/business_application/ 下）
    location /business_application/ {
        root /var/www/caam-application;   # 需把 dist 内容放到 /var/www/caam-application/business_application/
        try_files $uri $uri/ /business_application/index.html;
        index index.html;
    }

    # 后端 API（proxy_pass 后不要带 URI，否则会丢掉 /business_application/api 前缀）
    location /business_application/api/ {
        proxy_pass http://127.0.0.1:8094;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        # 后端硬限单文件 50MB，网关留出 multipart 开销即可（建议 ≥60m）
        client_max_body_size 60m;
        proxy_read_timeout 120s;
        proxy_send_timeout 120s;
    }

    # 上传文件
    location /business_application/uploads/ {
        proxy_pass http://127.0.0.1:8094;
        # 也可改为 root /var/www/xxxxx-application-backend/uploads; 直接由 Nginx 托管（少一次反代）
    }
}
```

> 若经反向代理，务必：① `server.trusted_proxies` 填 Nginx 地址；② 转发 `X-Real-IP` / `X-Forwarded-For`。否则限流（含单 IP 并发上限）看到的是 Nginx 的 IP，会误伤全站用户。

---

## 五、部署顺序

1. 准备 MySQL（**先创建数据库 `xxxx_application`**，InnoDB + utf8mb4）与 Redis（可用 db 2/3）；确认数据库账号对该库有建表权限
2. 部署后端：改 `config.yaml`（`mode` 改 `release`、端口、库连接、`trusted_proxies` 填 Nginx IP）→ 注入 `APPLICATION_DB_PASSWORD` / `APPLICATION_JWT_SECRET` → 编译 → 启动 → 确认 `GET /business_application/api/captcha` 返回 200
3. 部署前端：构建 `dist` → 复制到 Nginx 站点目录 → reload Nginx
4. 首次进入管理端：用 `admin` / `1qaz@WSX` 登录并**立即修改密码**，维护申报类别、批次、专家库
5. 业务冒烟：申报人注册 → 提交申报 + 上传材料 → 管理端在线初审 → 分配评审人 → 评审人评分 → 确定/公示结果 → 颁发证书 → 申报人下载证书

---

## 六、回滚 / 升级

- 后端：替换可执行文件后重启（`AutoMigrate` 只增列不删列，可直接用旧版本回滚；确需删表回滚请先备份）
- 前端：替换 Nginx 目录下 `dist` 内容；改版本时建议保留旧目录以便快速回退
- 更换 `APPLICATION_JWT_SECRET` 会使两套 Token 立即失效（在线用户需重新登录）
- 备份建议：`mysqldump` 数据库 + `backend/uploads` 目录（申报材料、证书附件）+ `backend/config.yaml`（不含密码明文）

### 升级说明（2026-09-23，上线前安全加固）

本次涉及登录态与前端字段名，**先部署后端 → 再部署前端**，顺序反了会出现「注册 400 / 管理端 403」的短暂不一致。

1. **Token 立即失效**：JWT 现在带 `aud`（申报人 / 管理端两套互不通用）、校验签发者并锁定 HS256。旧版本签发的 Token 全部失效，所有在线用户需重新登录（等同「更换 JWT 密钥」的表现，无需改配置）。
2. **管理端鉴权收紧**：`/admin/*` 每次请求回查账号是否仍存在且启用，角色码为空直接 `401`；`GET /admin/dashboard` 需 `dashboard:view`，`POST /admin/upload` 需 `application:preliminary` 或 `certificate:manage`（评审人无上传权限，属预期）。
3. **注册强制验证码**：`POST /member/register` 新增必填 `captcha_id` + `captcha_code`（前端注册页已同步；旧前端或脚本调用会返回 `code=400`「请填写完整信息（含验证码）」）。
4. **新增限流**：注册 / 登录 / 上传见「二.5」表格；联调压测若遇 `429`，属限流生效，换 IP 或等待窗口结束即可。
5. **结果公示字段变化**：`GET /member/results` 返回条目改为白名单 `{id, title, status, batchId, batch{id,title}, userRealName, publishedAt}`，不再返回 `user` 对象（原先会把全体申报人的身份证号、手机号、邮箱一起下发）。自定义前端若读过 `user.realName`，请改用 `userRealName`。
6. **申报创建字段白名单**：`POST /member/applications` 只接受 `batchId / categoryId / title / projectBrief / content`，请求体里的 `status`、`publishedAt`、`totalScore`、`finalOpinion`、`id` 一律忽略（原先把整个模型绑定进请求体，可把伪造条目直接塞进结果公示）。
7. **时间按服务器本地时区解析**：批次起止时间等无时区串（`YYYY-MM-DD HH:mm:ss`）改用 `time.Local`（与 member / portal 一致）。**部署机时区必须是 `Asia/Shanghai`**，否则仍不会按北京时间入库。
8. **前端 401 处理**：按发起请求的角色分别跳 `/login`（申报人）或 `/admin/login`（管理端），且只清理本应用的 `localStorage` 键，不再 `localStorage.clear()`（原先会连带清掉同域部署的 portal / member / base 登录态，管理端还会被跳到申报人登录页）。

### 移除软删除残留（2026-09-22，无手工 SQL 也能跑）

后端已统一为**物理删除（硬删）**：模型不再内嵌 `gorm.DeletedAt`，代码里没有 `Unscoped()` 与
`deleted_at IS NULL`，删除接口直接 `DELETE`（无回收站/恢复语义，删除后不可恢复）。
唯一索引只被存活记录占用，删除后同用户名/同编码可以直接重新创建（是一条全新记录，ID 不复用）。

- 服务启动时 `pkg/db.purgeLegacySoftDeletedRows()` 会自动物理删除旧库中 `deleted_at IS NOT NULL` 的残留行，
  并打日志 `[migrate] <表> 表物理删除 N 条历史软删除记录`。原因：查询已不过滤 `deleted_at`，不清理这些行会重新"出现"
  在列表里。新库没有 `deleted_at` 列，自动跳过；幂等，可反复启动。
- `AutoMigrate` 不会删除已存在的列/索引，旧库的 `deleted_at` 列与 `idx_<表>_deleted_at` 索引会保留（无害）。
  如需彻底清掉，先用下面语句生成 SQL（某表若没有该索引，单独执行 `DROP COLUMN`）：

```sql
SELECT CONCAT('ALTER TABLE `', TABLE_NAME, '` DROP INDEX `', INDEX_NAME, '`, DROP COLUMN `deleted_at`;')
  FROM information_schema.STATISTICS
 WHERE TABLE_SCHEMA = DATABASE() AND COLUMN_NAME = 'deleted_at'
 GROUP BY TABLE_NAME, INDEX_NAME;
```

---

## 七、业务速览

### 全流程闭环（7 个环节）

| 环节 | 说明 | 申报人端 | 管理端 |
|------|------|----------|--------|
| 项目申报 | 浏览开放批次、填写项目信息、保存草稿并提交 | ✅ | 批次/类别维护 |
| 材料上传 | 草稿阶段上传 PDF/Word/图片等申报材料 | ✅ | 下载查看 |
| 在线初审 | 管理人对提交的申报进行通过/驳回并填写意见 | 进度可见 | ✅ |
| 专家评审 | 分配评审人、评分（0-100）与评审意见，自动汇总平均分 | 评分可见 | ✅ 分配/评审 |
| 进度查询 | 申报人查看流程时间线与各环节意见 | ✅ | - |
| 结果公示 | 通过/不通过结果确定后公示，发布公示公告 | ✅ 查看 | ✅ 发布 |
| 证书下载 | 对通过项目颁发证书，申报人可下载 | ✅ 下载 | ✅ 颁发 |

> 支撑模块：**专家库管理**——维护评审专家档案（专业领域、职称、单位、简介等），新增专家时自动创建评审人登录账号，支持启用/停用与评审任务统计。

### 角色

| 角色 | 说明 |
|------|------|
| 申报人（前端用户） | 注册/登录、申报项目、上传材料、查询进度、下载证书 |
| 管理人 `manager` | 维护类别/批次/专家库、在线初审、分配评审、确定结果、发布公示、颁发证书、管理申报人与通知 |
| 评审人 `reviewer` | 查看分配给自己的评审任务、在线评分并填写评审意见 |
| 超级管理员 `super_admin` | 全部权限，含账号与角色管理 |

### 申报状态流转

```
draft(草稿) → submitted(待初审) → preliminary_rejected(初审驳回)
                              ↘ under_review(待评审) → reviewed(评审完成)
                                                          → passed(通过) / rejected(不通过)
                                                          → published(已公示) → certified(已发证)
```

### API 前缀与分组

`/business_application/api`（由 `server.api_prefix` 决定）

- 公开接口：`/captcha`、`/member/register`（需验证码）、`/member/login`（需验证码）、`/admin/login`（需验证码）
- 申报人接口：`/member/*`（需申报人 JWT）
- 管理接口：`/admin/*`（需管理端 JWT + 细粒度权限校验；`GET /admin/dashboard` 需 `dashboard:view`，`POST /admin/upload` 需 `application:preliminary` 或 `certificate:manage`）

---

## 八、常见问题（FAQ）

| 现象 | 原因 / 处理 |
| --- | --- |
| 启动 panic `Failed to connect database` | 数据库不存在（程序不建库）、账号权限不足或网络不通；DSN 未设置 connect timeout，网络不通会等待较久 |
| 启动 panic `Failed to connect Redis (captcha / anti-replay)` | Redis 未启动，或 `redis.addr` / `captcha_db` / `anti_replay_db` 配置错误；本服务**强依赖 Redis** |
| 启动即退出并打印未设置数据库密码/JWT 密钥 | 未注入 `APPLICATION_DB_PASSWORD` / `APPLICATION_JWT_SECRET`，或仍是占位值 |
| 日志出现 `[GIN-debug] [WARNING] Running in "debug" mode` | `server.mode` 仍为 `debug`，生产环境改为 `release` |
| 接口返回 `code=429`「请求过于频繁，请稍后再试」 | 命中接口限流（见「二.5」：验证码 30/分、注册/登录 10/分、上传 20/分），或单 IP 并发超过 `max_concurrent_ips` |
| 上传返回「不支持的文件类型」 | 扩展名不在白名单（见「二.5」） |
| 上传大文件返回 413 | Nginx `client_max_body_size` 比后端硬限（50MB）还小，请求没到后端就被网关拒绝；调到 ≥60m |
| 注册接口返回 400「请填写完整信息（含验证码）」 | 未传 `captcha_id` / `captcha_code`（注册已强制验证码，见「二.5」与升级说明） |
| 页面能打开但接口 404 | `server.api_prefix` 与 `VITE_API_BASE_URL` 不一致，或 Nginx 反代路径写错（`proxy_pass` 带了 URI） |
| 所有用户被限流、日志里 IP 都是同一个 | `server.trusted_proxies` 未填 Nginx 地址，或 Nginx 未转发 `X-Real-IP` / `X-Forwarded-For` |
| 后台账号在申报人登录页登录失败 | 属预期：两套账号分别存于 `application_admins` / `application_users`，不通用 |
| 改了后端端口后 dev 前端请求不通 | `frontend/vite.config.ts` 的 proxy target 与 `server.port` 必须一致（当前 8094 / 3003） |

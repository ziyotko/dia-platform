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
- `server.upload_dir_prefix`: `/business_application`（历史 `fileUrl` 的前缀，用于把数据库里的 URL 归一化回 `uploads/` 路径；需与 `VITE_BASE_PATH` 去尾斜杠一致）
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
- 上传目录：`./uploads`（**不再静态托管**，只能通过带鉴权的 `GET /member/files`、`GET /admin/files` 读取，见「二.5」）
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
| `GET /member/files`、`GET /admin/files`（文件下载） | 120 次/分钟 |

- 另有单 IP 并发请求上限 `server.max_concurrent_ips`（默认 100，**必须 > 0**：配成 0 会让所有请求都返回 429）；上述限制超限均返回 `code=429`「请求过于频繁，请稍后再试」
- **材料数量配额**：单份申报最多 **20 份**材料（`service.MaxMaterialsPerApplication`，超出返回 `code=500`「材料数量不能超过 20 份」）；单文件大小与扩展名限制见上

### 上传文件的下载（鉴权，2026-09-23 第三批改动）

上传目录**不再对外静态托管**：原先 `main.go` 的 `r.Static(<upload_dir_prefix>/uploads, ./uploads)` 等于「知道 URL 就能下载」，而 URL 一旦外泄就永久可读。现在：

- 申报人：`GET /member/files?url=<fileUrl>`（带申报人 JWT），只能读**自己申报的材料**与**自己证书的附件**；
- 管理端：`GET /admin/files?url=<fileUrl>`（带管理端 JWT），管理人/超管可读任意文件，**评审人只能读分配给自己的申报的材料**；
- 越权返回 `code=403`「无操作权限」，路径不在 `uploads/` 内返回 `code=404`；`.zip/.rar` 以附件形式下发，其余内联预览；响应带 `X-Content-Type-Options: nosniff`。
- 前端已改为「带 token 的 blob 请求 + objectURL」打开文件（`utils/file.ts`），原来的 `window.open(fileUrl)` 不再使用。
- **Nginx 若还配了 `/uploads/` 直接指向磁盘目录，必须删除**：否则绕过上述鉴权（见「四、Nginx 反向代理示例」）。

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

    # 上传文件：**不要**用 location 直接指向磁盘目录（会绕过鉴权），
    # 也不要再单独配一个 location：文件现在由 API 下发（GET /business_application/api/member|admin/files）
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

### 升级说明（2026-09-23 · 第二批：数据完整性）

9. **评审人权限收敛**：`reviewer` 只保留 `dashboard:view` + `review:score`，去掉 `application:view` / `batch:view`（原先可翻看全部申报详情与他人评分意见）。评审人仍可正常使用「管理看板 / 我的评审 / 个人资料」；评审接口 `GET /admin/reviews`、`GET /admin/reviews/:id` 本身按本人任务自限。若你曾自定义过角色权限，请同步该变更（启动时 `seedRoles()` 会按代码覆盖角色表的 `permissions`）。
10. **评审人视图脱敏**：评审人拿到的申报只保留项目内容与材料，`totalScore` / `avgScore` / `finalOpinion` / `preliminaryOpinion` 一律置空，避免打分前被当前均分「锚定」。
11. **状态推进改条件更新**（并发保护）：申报的撤回/提交/初审/终审/公示/撤回结果、评审提交与汇总回写、批次发布/结束/进入评审/编辑、公示公告发布与删除、证书编辑，全部改为 `WHERE ... AND status = ?` 的条件更新并在 `RowsAffected = 0` 时提示「…状态已变更，请刷新后重试」。**重复点击、双人同时操作不再产生重复状态变更与重复通知**（前端会看到上述提示，刷新后重试即可）。
12. **证书颁发/作废事务化**：颁发证书现在在事务中先用「申报 `published` → `certified`」的条件更新抢占该行（兼作互斥锁），再加证书、回写状态；同一申报**不会因双击/重试产生两张有效证书**。作废证书与「申报退回已公示」同事务，不再出现「证书已作废但申报仍是已发证」。
13. **评审分配并发保护**：分配评审人的读-判-写改为单事务 + `SELECT ... FOR UPDATE` 行锁，并发保存不再插入重复的 `(application_id, reviewer_id)` 行（也就不会出现同一位专家被算两次平均分）。
14. **材料数量上限 20 份**（见「二.5」），超出时提示「材料数量不能超过 20 份」。改上限需同时改 `service.MaxMaterialsPerApplication` 与本文件/手册。

### 升级说明（2026-09-23 · 第三批：上线前收尾）

15. **上传文件不再公开**：`/uploads` 静态托管已移除，文件只能通过 `GET /member/files`、`GET /admin/files`（带 token）读取，详见「二.5」。**必须同时删除 Nginx 里直接指向 `uploads` 目录的 location**，否则鉴权形同虚设。历史 `fileUrl` 无需迁移（仍作为标识传给接口）。
16. **撤回初审**：新增 `POST /admin/applications/:id/revoke-preliminary`（权限 `application:preliminary`），把误通过的申报从「待评审」退回「待初审」；**仅在无人评分时可用**，会一并清空未评分的评审人。
17. **批次申报期可调整 / 可重开**：`PUT /admin/batches/:id` 现在允许「申报中」批次改申报起止时间；「已进入评审/已结束」的批次只能把**申报截止时间改到将来**（即延长申报期），项目类别始终锁定。新增 `POST /admin/batches/:id/reopen`（权限 `batch:publish`）把评审中/已结束的批次退回「申报中」，要求截止时间已在将来且**该批次无已公示/已发证记录**。
18. **最后一个超管保护**：`PUT /admin/admins/:id`、`DELETE /admin/admins/:id` 不再允许把系统里唯一一个「启用中的超级管理员」降级/停用/删除，否则报「系统必须保留至少一个启用中的超级管理员」。
19. **审计日志可读可导出**：系统日志页新增「详情」列与「导出 CSV」按钮（`GET /admin/audit-logs/export`，权限 `audit:view`，最多 10000 条，UTF-8 BOM 便于 Excel 打开）；模块下拉已补齐「专家库/专家评审/系统日志」。
20. **运维兜底**：`server.max_concurrent_ips` ≤ 0 时自动回退为 100 并打印告警（原先配 0 会让全站请求都 429）；`server.trusted_proxies` 写错时回退为「不信任任何代理」并告警；单 IP 并发计数归零后会自动从内存表移除。

### 手工 SQL（可选加固：唯一索引）

代码层已用事务 + 条件更新保证下列约束，索引是「数据库层兼底」。`AutoMigrate` **不会**创建它们（旧库可能已有重复行，贸然建索引会让启动失败），请确认无重复后手工执行：

```sql
-- 1) 评审分配：同一申报 + 同一位评审人只能一行
--    先去重（保留最小 id，被删掉的重复行不应有已评分记录，执行前请人工确认）
SELECT application_id, reviewer_id, COUNT(*) AS cnt, GROUP_CONCAT(id ORDER BY id) AS ids
  FROM application_review_assignments
 GROUP BY application_id, reviewer_id HAVING cnt > 1;

-- 仅当上面为 0 行时才执行：
ALTER TABLE `application_review_assignments`
  ADD UNIQUE KEY `uk_app_reviewer` (`application_id`, `reviewer_id`);

-- 2) 证书编号：未作废的证书编号唯一（作废后释放编号，因此不能建普通唯一索引）
--    MySQL 8.0.13+ 的函数索引：
SELECT cert_no, COUNT(*) AS cnt, GROUP_CONCAT(id ORDER BY id) AS ids
  FROM application_certificates WHERE status <> 'void'
 GROUP BY cert_no HAVING cnt > 1;

-- 仅当上面为 0 行时才执行：
ALTER TABLE `application_certificates`
  ADD UNIQUE KEY `uk_cert_no_active` ((IF(`status` = 'void', NULL, `cert_no`)));
```

> 低于 MySQL 8.0.13（不支持函数索引）时，改用生成列变体：
> `ALTER TABLE application_certificates ADD COLUMN cert_no_active varchar(64) GENERATED ALWAYS AS (IF(status='void', NULL, cert_no)) STORED, ADD UNIQUE KEY uk_cert_no_active (cert_no_active);`
> 建索引后，撞唯一键的请求会被翻译成「证书编号已存在」/「评审人重复分配，请刷新后重试」。

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
| 评审人 `reviewer` | 查看分配给自己的评审任务、在线评分并填写评审意见（看不到其他评审人的分数/意见与均分） |
| 超级管理员 `super_admin` | 全部权限，含账号与角色管理 |

### 申报状态流转

```
draft(草稿) ⇄ submitted(待初审) ⇄ under_review(待评审) → reviewed(评审完成)
                              ↘ preliminary_rejected(初审驳回)      ↘ passed(通过) / rejected(不通过)
                                                                     ↘ published(已公示) → certified(已发证)
```

- `submitted → draft`：申报人「撤回」（仅待初审可撤）；
- `under_review → submitted`：管理人「撤回初审」（`POST /admin/applications/:id/revoke-preliminary`，仅无人评分时）；
- `reviewed/under_review → passed/rejected`：确定结果；`published → passed`、`passed/rejected → reviewed`：撤回公示 / 撤回评审结果；
- `certified → published`：作废证书。

### 批次状态流转

```
draft(草稿) ──发布──► open(申报中) ──开始评审──► reviewing(评审中) ──结束──► closed(已结束)
                         ▲                        │                    │
                         └──── 重开申报（无公示记录且截止时间在将来） ───┘
```

- `open` 期间可改申报起止时间（延长/缩短）；`reviewing`/`closed` 只能把申报截止时间改到将来（延长申报期），项目类别一旦发布就不可改；
- 申报期到点后后台任务每 10 分钟自动把 `open` 批次转为 `reviewing`（日志中可见 `UPDATE ... SET status='reviewing'`）。

### API 前缀与分组

`/business_application/api`（由 `server.api_prefix` 决定）

- 公开接口：`/captcha`、`/member/register`（需验证码）、`/member/login`（需验证码）、`/admin/login`（需验证码）
- 申报人接口：`/member/*`（需申报人 JWT；`GET /member/files` 下载自己的材料/证书附件）
- 管理接口：`/admin/*`（需管理端 JWT + 细粒度权限校验；`GET /admin/dashboard` 需 `dashboard:view`，`POST /admin/upload` 需 `application:preliminary` 或 `certificate:manage`，`GET /admin/files` 下载文件（管理人全部/评审人仅本人任务），`GET /admin/audit-logs/export` 导出 CSV）

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
| 提示「…状态已变更，请刷新后重试」 | 并发/重复操作：另一个人（或另一个标签页）已把该申报/批次/证书推到下一步，刷新看最新状态即可（这是并发保护，不是故障） |
| 提示「该申报已颁发证书或状态已变更，请刷新后重试」 | 同一份申报被重复点击了「颁发证书」；刷新后若已发证，则在「证书管理」里查看 |
| 提示「材料数量不能超过 20 份」 | 单份申报的材料配额（见「二.5」），删除不需要的材料或调高 `MaxMaterialsPerApplication` |
| 下载材料/证书提示「无操作权限」 | 文件不属于当前账号：申报人只能看自己的材料与证书；评审人只能看分配给自己的申报材料（见「二.5」） |
| 下载材料/证书提示「文件不存在」或 404 | `url` 不在 `uploads/` 下（历史/外部地址），或磁盘上的文件已被删除 |
| 所有用户被限流、日志里 IP 都是同一个 | `server.trusted_proxies` 未填 Nginx 地址，或 Nginx 未转发 `X-Real-IP` / `X-Forwarded-For` |
| 后台账号在申报人登录页登录失败 | 属预期：两套账号分别存于 `application_admins` / `application_users`，不通用 |
| 改了后端端口后 dev 前端请求不通 | `frontend/vite.config.ts` 的 proxy target 与 `server.port` 必须一致（当前 8094 / 3003） |

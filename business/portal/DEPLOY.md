# 门户系统（XXXX-portal）部署说明

> 分支：`caam_release`（汽车版）／ `miic_release`（中心版）｜ 项目路径：`business/portal/`
> 部署拓扑：Nginx 承载静态资源并反代后端 → Go 后端（默认 `0.0.0.0:8092`，配置项 `server.port`） → MySQL / Redis
> 说明：本仓库仅含「管理后台」与「后端 API」；网站静态化由外部程序完成，产物由外部 Web 服务托管。
> 文档与代码同步至：2026-09-18

## 目录结构

```
business/portal/
├── backend/          # Go 后端（Gin），module: server
│   ├── config.yaml   # 唯一配置文件（viper 从当前工作目录读 ./config.yaml，须在 backend 目录下启动）
│   ├── docs/         # 数据库结构等文档（docs/database.md）
│   ├── uploads/      # 上传文件目录（运行时需可写）
│   └── logs/         # 日志目录（logrus + lumberjack，按天分文件）
└── frontend/         # 管理后台（Vue3 + Vite + Element Plus）
    └── .env          # 部署子路径 VITE_BASE_PATH 与接口前缀 VITE_API_BASE_URL
```

---

## 一、环境依赖

| 组件 | 版本/说明 |
| --- | --- |
| Go | ≥ 1.27（`backend/go.mod` 声明 `go 1.27.1`），交叉编译 Linux 需 `CGO_ENABLED=0` |
| Node.js | ≥ 18（Vite 5） |
| MySQL | ≥ 8.0（InnoDB + utf8mb4）库名 `XXXX_portal`，端口/账号按环境配置（本仓库当前为 `10.3.1.95:63400` 的 `caam_portal`） | 
| Redis | 端口 6379，**三个库分工固定**：`db`(6) 验证码（5 分钟 TTL）、`db1`(7) 防重放 nonce + 限流计数、`db2`(8) 设置/静态化参数缓存 |
| Nginx | 托管前端静态文件并反代 API；**需把 Nginx 自身地址填入 `server.trusted_proxies`**，否则限流/封禁会把所有用户算成同一个 IP |

---

## 二、后端部署（backend）

### 1. 配置

编辑 `backend/config.yaml`（敏感项用环境变量覆盖，见下一节）：

- `server.host` / `server.port`: 监听地址与端口，默认 `0.0.0.0:8092`；改端口后需同步 Nginx 与 `frontend/vite.config.ts` 的 dev 代理目标
- `server.mode`: 生产用 `release`（留空也会回退 release，不会输出调试信息）
- `server.api_prefix`: `/xxxxx/api`（与前端 `VITE_API_BASE_URL` 一致；本仓库当前值为 `/business_portal/api`）
- `server.upload_dir_prefix`: `/xxxxx`（上传挂载点前缀，需与 `VITE_BASE_PATH` 去尾斜杠一致；本仓库当前值为 `/business_portal`，文件从 `/business_portal/uploads/**` 提供）
- `server.allowed_origins`: 生产环境改为实际前端域名，不要用 `*`（dev 默认值需与 `frontend/vite.config.ts` 的 `server.port` 一致，当前为 `http://localhost:3000`、`http://127.0.0.1:3000`）
- `server.trusted_proxies`: **可信反向代理地址，生产必须填 Nginx 的 IP**，决定 `X-Forwarded-For`/`X-Real-IP` 是否被信任（留空回退为仅信任 `127.0.0.1`）。限流、防重放封禁都按真实客户端 IP 统计
- `server.max_concurrent_ips`: 单个 IP 的并发请求上限（默认 50）
- 限流（Redis db 7 固定窗口）：`analytics_rate_limit`/`analytics_rate_window_seconds`（站点分析，默认 60 次/60s）、`login_rate_limit`/`login_rate_window_seconds`（登录，默认 10 次/300s）、`captcha_rate_limit`/`captcha_rate_window_seconds`（验证码，默认 30 次/60s）
- 防重放：`replay_window_seconds`（时间戳新鲜度窗口，默认 120s）、`replay_max_fail`（同一 IP 窗口内失败次数阈值，默认 10）、`replay_ban_minutes`（达阈值后临时封禁分钟数，默认 15）
- `database`: host / port / username / `password`（**只填占位值 `PORTAL_DB_PASSWORD`**）/ dbname / charset(`utf8mb4`) / `loc: Asia/Shanghai` / 读写超时与连接池
- `redis`: host / port / password / **`db`=6 验证码、`db1`=7 防重放+限流、`db2`=8 缓存**
- `jwt.secret`（占位值 `PORTAL_JWT_SECRET`）、`jwt.expires_hour`（默认 24）
- `log.level` / `log.path`（默认 `./logs`；单文件 100MB、保留 180 天、自动压缩）

### 2. 敏感信息用环境变量注入（必填）

`PORTAL_DB_PASSWORD`（数据库密码）与 `PORTAL_JWT_SECRET`（JWT 密钥，≥32 位随机串）已改为从环境变量读取，`config.yaml` 中仅保留占位值，请勿把真实密码/密钥提交到仓库。

外部静态化程序的访问令牌同样走环境变量：后台「系统设置」里只填**环境变量名**（如 `CAAM_STATIC_TOKEN`），令牌值由部署时注入。

```bash
# Linux / macOS
export PORTAL_DB_PASSWORD='<数据库密码>'
export PORTAL_JWT_SECRET='<≥32位随机密钥>'

# Windows PowerShell
$env:PORTAL_DB_PASSWORD='<数据库密码>'
$env:PORTAL_JWT_SECRET='<≥32位随机密钥>'

# Windows CMD
set PORTAL_DB_PASSWORD=<数据库密码>
set PORTAL_JWT_SECRET=<≥32位随机密钥>
```

> 环境变量优先于 `config.yaml`。若未设置（或仍为占位值），程序会输出告警并直接退出，生产环境必须注入。

### 3. 编译

```bash
cd business/portal/backend
# Windows 本地交叉编译 Linux 可执行文件（等价 comp.bat）
set goos=linux && go build -o xxxx main.go
# 或直接在目标机编译：
go build -o xxxx main.go
```

仓库根目录的批量脚本可一并编译四个后端（默认 linux/amd64、`CGO_ENABLED=0` 静态二进制，编译后校验产物为对应架构的 ELF）：

```powershell
powershell -ExecutionPolicy Bypass -File .\build-backends.ps1 -Only portal -Vet
```

### 4. 启动

```bash
./xxxx
```

- 启动流程：读 `./config.yaml` → 注入环境变量（缺失直接退出） → 连 MySQL/Redis（3 个 db） → `AutoMigrate` 建表 → **幂等补齐文章搜索所需的 FULLTEXT 索引** → 播种默认角色/用户/菜单/菜单权限（幂等） → 加载静态化参数到缓存（db 8） → 监听端口
- 默认账号：`admin`（管理员，角色 ID=1），初始密码 `1qaz@WSX`，**上线后立即修改**
- 首次对已有大表补 FULLTEXT 索引（`ALTER TABLE article ADD FULLTEXT INDEX`，共 3 个：title/author/source）会重建索引并短时占锁；表很大时建议部署窗口内手动先建好，启动逻辑检测到索引存在会自动跳过
- 静态资源上传目录：`./uploads`（后端以 `/xxxx/uploads` 提供）；需保证运行账号对该目录有读写权限
- 日志输出：`./logs`（按天分文件，单文件 100MB、保留 180 天、自动压缩）
- 健康检查（公开接口，无需 Token、无需防重放头）：`GET /xxxxx/api/site-info`
- 接口统一返回 HTTP 200 + 业务码：`code=0` 成功、`code=1` 失败、`code=401` 未认证/Token 失效（前端据此自动跳登录页）

### 5. 静态化（外部程序联动）

- 后端通过 `POST /xxxxx/api/static/*` 转发到「外部静态化程序」，非本仓库代码
- 静态化程序的地址 / 令牌环境变量名 / 输出路径在后台「系统设置」中配置（存于数据库，缓存于 Redis db 8）；令牌值从「所填环境变量名」对应的环境变量读取，以 `Authorization: Bearer <token>` 发送
- 后端调用超时 120s（`services/static_program_client.go`），**Nginx 的 `proxy_read_timeout` 必须 ≥ 180s**，否则大页面生成会被网关截断成 504
- 部署前确认该程序可达，否则静态化请求返回 502

---

## 三、前端部署（frontend）

### 1. 构建

```bash
cd business/portal/frontend
npm install
npm run build        # 等价 vue-tsc && vite build
```

也可用仓库根目录脚本（构建后自动校验 `dist/index.html` 里的资源前缀与 `VITE_BASE_PATH` 一致，避免部署到错误子路径）：

```powershell
powershell -ExecutionPolicy Bypass -File .\build-frontends.ps1 -Only business_portal
```

构建产物：`frontend/dist/`

### 2. 关键配置

- `.env`：`VITE_BASE_PATH=/xxxxx/`（部署在 `/xxxxx/` 子路径下）、`VITE_API_BASE_URL=/xxxxx/api`（本仓库当前为 `/business_portal/` + `/business_portal/api`）
- 若部署路径变化，修改 `VITE_BASE_PATH` 后重新构建；同时在 `backend/config.yaml` 同步 `server.api_prefix`（`/xxxxx/api`）与 `server.upload_dir_prefix`（`/xxxxx`）
- dev 代理键由 `.env` 推导（`[apiBase]` 与 `[basePath]/uploads`），不要再手写成别的前缀——否则 dev 下请求不会被转发（历史问题：`/portal/api` 代理键对不上 `/business_portal/api`）
- dev 代理**目标地址写死在 `frontend/vite.config.ts`**（当前 `http://127.0.0.1:8092`），改了后端 `server.port` 必须同步改这里

### 3. 部署

将 `dist/` 内容复制到 Nginx 站点目录（如 `/var/www/xxxxx-portal`）。

---

## 四、Nginx 反向代理示例

```nginx
server {
    listen 80;
    server_name portal.example.com;

    # 管理后台静态资源（/xxxxx/ 下）
    location /xxxxx/ {
        root /var/www/caam-portal;   # 与 dist 对应，保证 /xxxxx/ 命中 index.html
        try_files $uri $uri/ /xxxxx/index.html;
        index index.html;
    }

    # 后端 API（proxy_pass 后不要带 URI，否则会丢掉 /xxxxx/api 前缀）
    location /xxxxx/api/ {
        proxy_pass http://127.0.0.1:8092;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        # 上传上限：视频 800MB、其它 50MB（与后端校验一致）
        client_max_body_size 800m;
        # 静态化调用后端超时 120s，网关留出余量
        proxy_read_timeout 180s;
        proxy_send_timeout 180s;
    }

    # 上传文件
    location /xxxxx/uploads/ {
        proxy_pass http://127.0.0.1:8092;
        # 也可改为 root /var/www/xxxxx-backend/uploads; 直接由 Nginx 托管（少一次反代）
    }
}
```

> 若经反向代理，务必：① `server.trusted_proxies` 填 Nginx 地址；② 转发 `X-Real-IP` / `X-Forwarded-For`。否则限流与防重放封禁看到的是 Nginx 的 IP，会误伤全站用户。

> 网站静态化产物（首页/栏目/详情页）由外部静态化程序产出，另配 Nginx 站点（或独立 Web 服务）托管，域名/路径需与后台「静态化访问地址」一致。

---

## 五、部署顺序

1. 准备 MySQL、Redis，并创建数据库 `xxxxx_portal`（InnoDB + utf8mb4）；确认数据库账号对目标库有建表与加索引（`ALTER`）权限
2. 部署后端：改 `config.yaml`（端口、库连接、`server.trusted_proxies` 填 Nginx IP）→ 注入 `PORTAL_DB_PASSWORD` / `PORTAL_JWT_SECRET`（以及外部静态化令牌环境变量）→ 编译 → 启动 → 确认 `GET /xxxxx/api/site-info` 返回 200
3. 部署前端：构建 `dist` → 配置 Nginx 指向（子路径与 `VITE_BASE_PATH` 一致）
4. 配置外部静态化程序地址/令牌环境变量名/输出路径，验证静态化接口（如首页生成）正常
5. 健康检查：访问后台页面、登录（验证码与限流生效）、上传（含大文件）、发起一次静态化任务确认无 502/504

---

## 六、回滚 / 升级

- 后端：替换 `xxxxx` 可执行文件后重启（结构变更会自动迁移表；新增的 FULLTEXT 索引也会在启动时幂等补齐）
- 前端：替换 Nginx 目录下 `dist` 内容；改版本时建议保留旧目录以便快速回退
- 更换 `PORTAL_JWT_SECRET` 会使所有已签发 Token 立即失效（在线用户需重新登录）
- 用旧版本回滚时，新版本已新增的表/列/索引无需回退（AutoMigrate 不会删列）；确需删表回滚请先备份
- 建议备份：`mysqldump` 数据库 + `backend/uploads` 目录 + `backend/config.yaml`（不含密码明文）

---

## 七、常见问题（FAQ）

| 现象 | 原因 / 处理 |
| --- | --- |
| 文章查询报 `Can't find FULLTEXT index matching the column list` | `article` 表缺少 `MATCH ... AGAINST` 所需的 FULLTEXT 索引。当前版本启动时会自动补齐 `idx_article_title_fulltext` / `idx_article_author_fulltext` / `idx_article_source_fulltext`；若日志出现 `[migrate] 创建 FULLTEXT 索引 ... 失败`，检查数据库账号是否具备 `ALTER` 权限 |
| 接口返回「缺少防重放攻击请求头」/「请求已过期，请重新发送」 | 请求未带 `X-Request-Timestamp`（毫秒时间戳）与 `X-Request-Nonce`（8–128 位）头。浏览器端由前端 `utils/request.ts` 统一添加；脚本或第三方对接需自行实现（已认证请求还需 `X-Request-Signature`） |
| 登录/验证码返回「请求过于频繁，请稍后再试」 | 命中固定窗口限流或防重放临时封禁（Redis db 7）。优先确认 `server.trusted_proxies` 是否填了 Nginx IP，否则全站共用一个 IP 会误伤合法用户 |
| 多用户同时被登出、接口返回 `code=401` | Token 过期（`jwt.expires_hour`）或 `PORTAL_JWT_SECRET` 被更换；前端收到 401 会自动跳登录页 |
| 静态化请求 502 / 504 | 502 = 外部静态化程序不可达（检查「系统设置」中的地址与网络连通）；504 = 生成耗时超过网关超时（后端客户端 120s，Nginx `proxy_read_timeout` 建议 ≥180s） |
| 上传大文件失败 | 后端限制：视频目录 800MB、其它目录 50MB；同时确认 Nginx `client_max_body_size` ≥ 该值 |
| 启动即退出并打印 `程序退出` | 未注入 `PORTAL_DB_PASSWORD` 或 `PORTAL_JWT_SECRET`（仍为占位值），或 `./config.yaml` 不在进程工作目录下 |
| 端口被占用启动失败 | `server.port` 默认 8092；被占用时改配置或释放端口，并同步 Nginx 与 `frontend/vite.config.ts`（dev 代理） |

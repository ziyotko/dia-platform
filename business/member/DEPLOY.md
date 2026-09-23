# 会员系统（XXXX-member）部署说明

> 分支：`caam_release`（汽车版）／ `miic_release`（中心版）｜ 项目路径：`business/member/`
> 部署拓扑：Nginx 承载前端静态资源并反代后端 → Go 后端（默认 `0.0.0.0:8093`，配置项 `server.port`） → MySQL / Redis
> 说明：本仓库含「会员中心（对外公开页 + 会员专区）」与「管理后台（`/admin/*` 路由，同一个前端工程）」以及后端 API；证书 PDF 由后端纯 Go 生成（需中文字体）。
> 文档与代码同步至：2026-09-18

## 目录结构

```
business/member/
├── backend/          # Go 后端（Gin），module: member
│   ├── config.yaml   # 唯一配置文件（从进程工作目录读 ./config.yaml，须在 backend 目录下启动）
│   ├── logs/         # 日志目录（lumberjack 滚动，默认 logs/member.log）
│   ├── uploads/      # 上传文件目录（运行时需可写）
│   └── comp.bat      # Windows 交叉编译脚本（goos=linux，产物 ./member）
└── frontend/         # 会员中心 + 管理后台（Vue3 + Vite + Element Plus）
    └── .env          # 部署子路径 VITE_BASE_PATH 与接口前缀 VITE_API_BASE_URL
```

---

## 一、环境依赖

| 组件 | 版本/说明 |
| --- | --- |
| Go | ≥ 1.27（`backend/go.mod` 声明 `go 1.27.1`），交叉编译 Linux 需 `CGO_ENABLED=0`（证书 PDF 为纯 Go 实现，无 CGO 依赖） |
| Node.js | ≥ 18（Vite 5；`frontend/package.json`） |
| MySQL | ≥ 8.0（InnoDB + utf8mb4）库名 `XXXX_member`，当前仓库配置指向 `10.3.1.95:63400` 的 `caam_member`；**库必须事先建好**（程序只连库、不建库） |
| Redis | 端口 6379，**两个库分工固定**：`captcha_db`(4) 验证码、`anti_replay_db`(5) 限流计数；**Redis 不可达时启动直接失败**（Ping 失败即 Fatal） |
| 中文字体 | 生成证书 PDF 必需 TTF/OTF 中文字体（Linux 需自行安装，见「二.5」）；不支持 TTC 字体集合 |
| Nginx | 托管前端静态文件并反代 API；**需把 Nginx 地址填入 `server.trusted_proxies`**，否则限流会把所有用户算成同一个 IP |

---

## 二、后端部署（backend）

### 1. 配置

编辑 `backend/config.yaml`（敏感项用环境变量覆盖，见下一节）：

- `server.host` / `server.port`: 监听地址与端口，默认 `0.0.0.0:8093`；改端口后需同步 Nginx 与 `frontend/vite.config.ts` 的 dev 代理目标
- `server.mode`: 生产用 `release`（可用 `MEMBER_MODE` 覆盖；留空自动回退 `release`）
- `server.api_prefix`: `/business_member/api`（必须与前端 `VITE_API_BASE_URL` 一致；留空回退 `/member/api`）
- `server.upload_dir_prefix`: `/business_member`（上传挂载 = `<该值>/uploads`，需与 `VITE_BASE_PATH` 去尾斜杠一致）
- `server.allowed_origins`: 允许的前端来源（生产填实际域名，不要用 `*`；当前 dev 值 `http://localhost:5173`、`http://127.0.0.1:5173`）
- `server.trusted_proxies`: **可信反向代理地址，生产必须填 Nginx 的 IP**（决定 `X-Forwarded-For`/`X-Real-IP` 是否被信任；默认仅 `127.0.0.1`）
- `server.max_concurrent_ips`: 单个 IP 的并发请求上限（默认 100）
- `mysql`: host / port / user / `password`（**只填占位值 `MEMBER_DB_PASSWORD`**）/ db_name / charset(`utf8mb4`) / max_open / max_idle
- `redis`: addr / password / `captcha_db`(4) / `anti_replay_db`(5)
- `jwt`: `secret`（占位值 `MEMBER_JWT_SECRET`）/ `expire_hours`（默认 24）/ `issuer`（`caam-member`）
- `log`: `path`（默认 `logs/member.log`）/ `max_size`(100MB) / `max_backups`(30) / `max_age`(180 天)
- `certificate.font_path`: 证书 PDF 中文字体路径，留空按系统常见路径自动查找（可用 `MEMBER_CERT_FONT` 覆盖）

> 注意：YAML 中不要出现重复 key（viper 解析会直接报错）。

### 2. 敏感信息用环境变量注入（必填）

| 环境变量 | 必填 | 说明 |
| --- | --- | --- |
| `MEMBER_DB_PASSWORD` | 是 | 数据库密码，覆盖 `mysql.password` |
| `MEMBER_JWT_SECRET` | 是 | JWT 密钥（建议 ≥32 位随机串），覆盖 `jwt.secret` |
| `MEMBER_MODE` | 否 | 覆盖 `server.mode`（如 `debug`） |
| `MEMBER_CERT_FONT` | 否 | 覆盖 `certificate.font_path`（证书中文字体） |

```bash
# Linux / macOS
export MEMBER_DB_PASSWORD='<数据库密码>'
export MEMBER_JWT_SECRET='<≥32位随机密钥>'

# Windows PowerShell
$env:MEMBER_DB_PASSWORD='<数据库密码>'
$env:MEMBER_JWT_SECRET='<≥32位随机密钥>'

# Windows CMD
set MEMBER_DB_PASSWORD=<数据库密码>
set MEMBER_JWT_SECRET=<≥32位随机密钥>
```

> 未设置（或仍是占位值）时程序会打印告警后**直接退出**，生产环境必须注入。

### 3. 编译

```bash
cd business/member/backend
# Windows 本地交叉编译 Linux 可执行文件（等价 comp.bat）
set goos=linux && go build -o member main.go
# 或直接在目标机编译：
go build -o member main.go
```

仓库根目录的批量脚本可一并编译四个后端（默认 linux/amd64、`CGO_ENABLED=0` 静态二进制，编译后校验产物为对应架构的 ELF）：

```powershell
powershell -ExecutionPolicy Bypass -File .\build-backends.ps1 -Only member -Vet
```

### 4. 启动

```bash
./member
```

- 启动流程：读 `./config.yaml` → 注入环境变量（缺失直接退出） → 连 MySQL（**不建库**） → 连 Redis 两个库（Ping 失败直接退出） → 初始化 IP 并发限制器 → `AutoMigrate` 建表（会员/申请/证书/会费/机构/公告/文章/系统配置/会员等级/变更记录/操作日志等） → 写入种子数据 → 监听端口
- 种子数据（幂等，仅当对应表为空时写入）：管理员账号、默认机构（总会 + 分会）、文章分类、系统配置（站点名称/银行账户/联系方式/备案号等 10 项）、示例公告、4 个默认会员等级（会员单位/理事单位/副理事长单位/理事长单位）
- 默认管理员：用户名 `admin`，初始密码 **`Abcd@1234`**，**上线后立即修改**；后台重置他人密码也重置为该默认密码
- 静态资源上传目录：`./uploads`（后端以 `/business_member/uploads` 提供，危险扩展名强制以附件下载）
- 日志输出：`logs/member.log`（单文件 100MB、保留 30 个备份、180 天）
- 健康检查（公开接口）：`GET /business_member/api/site-info`
- 响应约定：**HTTP 状态恒为 200**，业务结果看 `code`：`0` 成功、`400` 参数错误、`401` 未登录/Token 失效、`403` 需要管理员权限、`404` 不存在、`429` 请求过于频繁、`500` 服务端错误

### 5. 证书 PDF 中文字体（Linux 部署必做）

生成/补生成会员证书 PDF 会走 `pkg/certpdf`（fpdf + gofpdi 叠加模板页），中文字体缺少时接口直接报错「未找到可用的中文证书字体（TTF/OTF）」。

- 三种来源（按优先级）：`certificate.font_path` → 环境变量 `MEMBER_CERT_FONT` → 系统常见路径自动查找
- 自动查找的 Linux 路径：`/usr/share/fonts/opentype/noto/NotoSansCJKsc-Regular.otf`、`/usr/share/fonts/truetype/wqy/wqy-microhei.ttf`、`/usr/share/fonts/truetype/arphic/uming.ttf`
- 安装示例：Debian/Ubuntu `apt install fonts-noto-cjk`（或 `fonts-wqy-microhei`）；CentOS/RHEL `yum install wqy-microhei-fonts`
- **仅支持 TTF/OTF**；`.ttc` 字体集合会被拒绝，需先用 `fonttools` 之类的工具拆出单个 TTF

### 6. 限流与防刷

统一走 Redis 固定窗口（`anti_replay_db`=5，Redis 异常时退化为进程内计数）：

| 接口 | 限制 |
| --- | --- |
| `GET /captcha` | 30 次/分钟 |
| `POST /auth/register` | 10 次/分钟 |
| `POST /auth/check-exists` | 30 次/分钟 |
| `POST /auth/login` | 10 次/分钟 |
| `POST /upload` | 20 次/分钟 |
| `GET /charter`（章程 PDF 下载） | 30 次/分钟 |
| 单 IP 并发请求 | `server.max_concurrent_ips`（默认 100），超限返回 `code=429`「请求过于频繁，请稍后再试」 |

上传体积上限：会员中心 `POST /upload` 与票据 PDF 均 10MB，章程 PDF 20MB。

---

## 三、前端部署（frontend）

### 1. 构建

```bash
cd business/member/frontend
npm install
npm run build        # 等价 vue-tsc && vite build
```

也可用仓库根目录脚本（构建后自动校验 `dist/index.html` 里的资源前缀与 `VITE_BASE_PATH` 一致）：

```powershell
powershell -ExecutionPolicy Bypass -File .\build-frontends.ps1 -Only business_member
```

构建产物：`frontend/dist/`

### 2. 关键配置

- `.env`：`VITE_BASE_PATH=/business_member/`、`VITE_API_BASE_URL=/business_member/api`（当前仓库值）
- 若部署路径变化，改 `VITE_BASE_PATH` 后重新构建；同时同步 `backend/config.yaml` 的 `server.api_prefix`（`/xxxx/api`）与 `server.upload_dir_prefix`（`/xxxx`）
- dev 代理键由 `.env` 推导（`[apiBase]` 与 `[basePath]/uploads`），**代理目标写死在 `frontend/vite.config.ts`**（当前 `http://127.0.0.1:8093`），改后端端口必须同步改这里
- dev 端口：`3001`（`vite.config.ts` 的 `server.port`）
- 前端依赖：Vue 3.4 / Element Plus 2.6 / Pinia / ECharts 6 / wangEditor 5（章程富文本）
- 会员中心与管理后台是**同一个工程**：对外页面走 `PublicLayout`，后台页面路由以 `/admin/*` 开头（登录后按 `is_admin` 显示入口），不需要部署两个站点
- 现象说明：应用内路由 path 仍以 `/member/...` 开头（如 `/member/dashboard`），叠加 base 后实际 URL 为 `/business_member/member/dashboard`，属既有约定，不影响使用

### 3. 部署

将 `dist/` 内容复制到 Nginx 站点目录下的 `business_member/` 子目录（见下节示例）。

---

## 四、Nginx 反向代理示例

```nginx
server {
    listen 80;
    server_name member.example.com;

    # 前端静态资源（/business_member/ 下）
    location /business_member/ {
        root /var/www/caam-member;   # 需把 dist 内容放到 /var/www/caam-member/business_member/
        try_files $uri $uri/ /business_member/index.html;
        index index.html;
    }

    # 后端 API（proxy_pass 后不要带 URI，否则会丢掉 /business_member/api 前缀）
    location /business_member/api/ {
        proxy_pass http://127.0.0.1:8093;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        # 最大上传：章程 PDF 20MB（其余 10MB）
        client_max_body_size 25m;
        # 证书 PDF 生成（含批量补生成）耗时较长，网关留出余量
        proxy_read_timeout 180s;
        proxy_send_timeout 180s;
    }

    # 上传文件
    location /business_member/uploads/ {
        proxy_pass http://127.0.0.1:8093;
        # 也可改为 root /var/www/xxxxx-member-backend/uploads; 直接由 Nginx 托管（少一次反代）
    }
}
```

> 若经反向代理，务必：① `server.trusted_proxies` 填 Nginx 地址；② 转发 `X-Real-IP` / `X-Forwarded-For`。否则限流（含单 IP 并发上限）看到的是 Nginx 的 IP，会误伤全站用户。

---

## 五、部署顺序

1. 准备 MySQL（**先创建数据库 `xxxx_member`**，InnoDB + utf8mb4）与 Redis（可用 db 4/5）；确认数据库账号对该库有建表权限
2. Linux 目标机安装中文字体（见「二.5」），否则证书 PDF 生成会失败
3. 部署后端：改 `config.yaml`（端口、库连接、`trusted_proxies` 填 Nginx IP）→ 注入 `MEMBER_DB_PASSWORD` / `MEMBER_JWT_SECRET`（可选 `MEMBER_CERT_FONT`）→ 编译 → 启动 → 确认 `GET /business_member/api/site-info` 返回 200
4. 部署前端：构建 `dist` → 复制到 Nginx 站点目录 → reload Nginx，确认页面与登录正常
5. 首次进入后台：用 `admin` / `Abcd@1234` 登录并**立即修改密码**，检查机构树、会员等级、会费标准、证书模板等基础数据
6. 业务冒烟：注册 → 提交入会申请 → 后台审批 → 会费确认 → 生成证书（验证中文字体与 PDF 下载）→ 章程富文本/PDF 维护

---

## 六、回滚 / 升级

- 后端：替换可执行文件后重启（`AutoMigrate` 只增列不删列，可直接用旧版本回滚；确需删表回滚请先备份）
- 前端：替换 Nginx 目录下 `dist` 内容；改版本时建议保留旧目录以便快速回退
- 更换 `MEMBER_JWT_SECRET` 会使所有已签发 Token 立即失效（在线用户需重新登录）
- 备份建议：`mysqldump` 数据库 + `backend/uploads` 目录（含证书 PDF、证书模板、章程 PDF、票据、文章配图）+ `backend/config.yaml`（不含密码明文）

### 升级说明（2026-09-23，无需手工 SQL）

- **入会审批**：改为**事务内条件更新**（`WHERE id=? AND status='pending_review'`），并发重复点「通过」时后提交者报「该申请已被处理，请刷新后重试」；「已有已通过申请」校验移入事务。
- **入会审批缺等级直接报错**：总会（或申请机构）未配置「关联等级」时不再生成 `level_id=0` 的费用记录，而是返回「机构尚未配置会员等级，请先在后台「组织机构」中…关联会员等级，再审核通过」。**升级后请先确认总会的关联等级已配置**，否则审批会被拒（历史上会产生无法缴费的坑）。
- **入会审批同年费用记录幂等**：该会员当年已有费用记录时跳过创建并记 Warn（避免与下面的唯一索引冲突）。
- **会费确认口径统一**：首次置为「已缴费」时统一写入 `paid_at` 与 `confirmed_at`；「确认缴费」现在也允许对**未缴费**记录使用（线下收款登记实收金额、填 0 即免缴），`paid_amount` 支持请求显式传入。
- **操作日志**：`params` 写库前对口令类字段脱敏为 `***`；截断改为按**字符**（原先按字节切中文会产生非法 UTF-8，审计记录会静默写库失败）。
- **上传白名单**：新增 `.zip` / `.rar`（入会申请页一直提示可打包上传，此前会被 400 拒绝）；zip/rar 与危险扩展名一样强制以附件下载。
- **前端**：修复公告「置顶」开关不生效（字段名与后端不一致）；后台会费统计卡改为后端聚合（不再按当前页计算）；会费页切换筛选自动回到第 1 页；编辑费用时正确回填等级名；「系统管理」隐藏 `charter_file`；退出登录/注册不再 `localStorage.clear()`（避免清掉同域 portal/application 的登录态）。

### 手工 SQL（可选加固）

`member_fee_records` 的 `(member_id, year)` 唯一索引需**手工创建**（`AutoMigrate` 不建索引）。不建也能跑（服务层已在事务内加行锁判断重复），建索引是并发双击的数据库级兜底。

```sql
USE caam_member;   -- 必须先切库：DELETE ... JOIN 未选库会报 1046 No database selected

-- 1) 预检：是否存在同一会员同年度的重复记录
SELECT COUNT(*) AS dup_groups, IFNULL(SUM(c-1), 0) AS extra_rows FROM (
  SELECT member_id, year, COUNT(*) c
  FROM member_fee_records GROUP BY member_id, year HAVING c > 1
) t;

-- 2) 有重复时先清理（每组保留 id 最小的一条）
DELETE f FROM member_fee_records f
JOIN (
  SELECT MIN(id) AS keep_id, member_id, year
  FROM member_fee_records GROUP BY member_id, year HAVING COUNT(*) > 1
) d ON d.member_id = f.member_id AND d.year = f.year
WHERE f.id > d.keep_id;

-- 3) 创建唯一索引（若已存在会报 1061，可忽略）
ALTER TABLE member_fee_records ADD UNIQUE INDEX uk_member_year (member_id, year);
```

核查：`SHOW INDEX FROM member_fee_records WHERE Key_name = 'uk_member_year';`

---

## 七、常见问题（FAQ）

| 现象 | 原因 / 处理 |
| --- | --- |
| 启动即退出，日志 `Failed to connect to captcha Redis / anti-replay Redis` | Redis 未启动或 `redis.addr`/`captcha_db`/`anti_replay_db` 配置错误；本服务**强依赖 Redis**，Ping 失败直接退出 |
| 启动即退出并打印 `[WARN] 未设置数据库密码/JWT 密钥` + `程序退出` | 未注入 `MEMBER_DB_PASSWORD` / `MEMBER_JWT_SECRET`，或仍是占位值 |
| `Failed to connect to database` | 数据库不存在（程序不建库）、账号权限不足或网络不通；DSN 未设置 connect timeout，网络不通时会等待较久才报错 |
| 证书生成失败：`未找到可用的中文证书字体（TTF/OTF）` | 安装 CJK 字体（`fonts-noto-cjk` / `wqy-microhei-fonts`）或配置 `certificate.font_path` / `MEMBER_CERT_FONT`；`.ttc` 不支持 |
| 接口返回 `code=429`「请求过于频繁，请稍后再试」 | 命中登录/注册/上传等固定窗口限流，或单 IP 并发超过 `max_concurrent_ips` |
| 页面能打开但接口 404 / 跨域失败 | `server.api_prefix` 与 `VITE_API_BASE_URL` 不一致；或 Nginx 反代路径写错（`proxy_pass` 带了 URI）；或来源不在 `server.allowed_origins` |
| 所有用户被限流/封禁、日志里 IP 都是同一个 | `server.trusted_proxies` 未填 Nginx 地址，或 Nginx 未转发 `X-Real-IP` / `X-Forwarded-For` |
| 上传返回「文件大小不能超过 10MB」 | 会员中心上传上限 10MB（章程 PDF 为 20MB）；同时确认 Nginx `client_max_body_size` ≥ 该值 |
| 忘记 admin 密码 | 用另一管理员账号在后台重置（重置为默认密码 `Abcd@1234`），或直接处理数据库（密码为 bcrypt 哈希，无法反推） |
| 改了后端端口后 dev 前端请求不通 | `frontend/vite.config.ts` 的 proxy target 与 `server.port` 必须一致（当前 8093 / 3001） |

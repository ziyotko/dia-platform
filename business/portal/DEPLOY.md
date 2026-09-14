# 门户系统（XXXX-portal）部署说明

> 分支：`caam_release`（汽车版）／ `miic_release`（中心版）｜ 项目路径：`business/portal/`
> 部署拓扑：Nginx 承载静态资源并反代后端 → Go 后端(8084) → MySQL / Redis
> 说明：本仓库仅含「管理后台」与「后端 API」；网站静态化由外部程序完成，产物由外部 Web 服务托管。

## 目录结构

```
business/portal/
├── backend/    # Go 后端（Gin），module: server
└── frontend/   # 管理后台（Vue3 + Vite + Element Plus）
```

---

## 一、环境依赖

| 组件 | 版本/说明 |
| --- | --- |
| Go | ≥ 1.26（`backend/go.mod`） |
| Node.js | ≥ 18（Vite 5） |
| MySQL | ≥ 8.0 库名 `XXXX_portal`，端口默认 3305 修改与后端配置一致即可 | 
| Redis | 端口 6379，使用 db 9 / 10 / 11 |
| Nginx | 用于托管前端静态文件并反代 API |

---

## 二、后端部署（backend）

### 1. 配置

编辑 `backend/config.yaml`（或用环境变量覆盖）：

- `server.port`: 8084（对外可改，注意与 Nginx 一致）
- `server.mode`: 生产用 `release`
- `server.api_prefix`: `/xxxxx/api`（与前端一致即可，前端依赖）
- `server.allowed_origins`: 生产环境改为实际前端域名，不要用 `*`（开发默认值需与 `frontend/vite.config.ts` 的 `server.port` 一致，当前为 `http://localhost:3000`、`http://127.0.0.1:3000`）
- `database` / `redis`: 按实际环境修改

### 2. 敏感信息用环境变量注入（必填）

`PORTAL_DB_PASSWORD`（数据库密码）与 `PORTAL_JWT_SECRET`（JWT 密钥，≥32 位随机串）已改为从环境变量读取，`config.yaml` 中仅保留占位值，请勿把真实密码/密钥提交到仓库。

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

### 4. 启动

```bash
./xxxx
```

- 启动自动 `AutoMigrate` 建表，并初始化默认角色 / 用户 / 菜单（默认用户admin，初始密码1qaz@WSX）
- 静态资源上传目录：`./uploads`（后端以 `/xxxx/uploads` 提供）
- 日志输出：`./logs`

### 5. 静态化（外部程序联动）

- 后端通过 `POST /xxxxxx/api/static/*` 转发到「外部静态化程序」，非本仓库代码
- 静态化程序的地址 / 令牌环境变量名 / 输出路径在后台「系统设置」中配置（存于数据库），输出站点（如 `dist/caam-home`）由外部 Web 服务器托管
- 部署前确认该程序可达，否则静态化请求返回 502

---

## 三、前端部署（frontend）

### 1. 构建

```bash
cd business/portal/frontend
npm install
npm run build        # 等价 vue-tsc && vite build
```

构建产物：`frontend/dist/`

### 2. 关键配置

- `.env`：`VITE_BASE_PATH=/xxxxx/`（部署在 `/xxxxx/` 子路径下）、`VITE_API_BASE_URL=/xxxxx/api`
- 若部署路径变化，修改 `VITE_BASE_PATH` 后重新构建

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

    # 后端 API
    location /xxxxx/api/ {
        proxy_pass http://127.0.0.1:8084;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }

    # 上传文件
    location /xxxxx/uploads/ {
        proxy_pass http://127.0.0.1:8084;
    }
}
```

> 网站静态化产物（首页/栏目/详情页）由外部静态化程序产出，另配 Nginx 站点（或独立 Web 服务）托管，域名/路径需与后台「静态化访问地址」一致。

---

## 五、部署顺序

1. 准备 MySQL、Redis，并创建数据库 `xxxxx_portal`
2. 部署后端：改配置 → 注入环境变量 → 编译 → 启动 → 确认 `/xxxxx/api` 可访问
3. 部署前端：构建 `dist` → 配置 Nginx 指向
4. 配置外部静态化程序地址/令牌/输出路径，验证静态化接口（如首页生成）正常
5. 健康检查：访问后台页面、登录、上传、发起一次静态化任务确认无 502

---

## 六、回滚 / 升级

- 后端：替换 `xxxxx` 可执行文件后重启（结构变更会自动迁移表）
- 前端：替换 Nginx 目录下 `dist` 内容；改版本时建议保留旧目录以便快速回退

# 子应用接入标准

本目录提供已有 `frontend + backend` 系统接入 Base 底座平台的规范、示例代码与最佳实践。

## 接入前准备

1. 在 Base 后台「应用管理」中注册应用，获取 `appCode`。
   - 接入类型 `type` 可选：`iframe` / `proxy`。
   - 若使用 API 代理，填写 `backendUrl` 与 `API 前缀`（如 `/api`，代理时会拼在路径前）。
   - 若使用 IFrame，填写 `frontendUrl`。
2. 在「应用实例」中为对应租户开通该应用。
3. 在「菜单管理」中为该应用配置菜单（`appCode` 与注册时一致）。

## 接入方式

### 1. IFrame 接入（最快）

- 应用独立部署。
- Base 前端通过 iframe 加载应用前端地址（菜单 `target = iframe` 时，地址取自菜单的组件/路径字段）。
- 底座在 iframe 地址后自动追加查询参数（已有 query 会以 `&` 追加）：`base_token`、`user_id`、`username`、`tenant_id`。
- 子应用前端解析 `base_token`，调用底座 `GET /business_base/api/auth/info` 校验用户信息。
- 适合已有系统改动成本低的场景。

### 2. API 代理接入（推荐）

- 应用后端无需暴露公网，由底座反向代理。
- 底座在请求头中注入当前用户信息：
  - `X-Base-User-ID`
  - `X-Base-Username`
  - `X-Base-Tenant-ID`
- 应用后端读取请求头完成鉴权与数据隔离。
- 代理路径：`/business_base/api/app/{appCode}/{原应用API路径}`，实际转发地址为
  `{backendUrl}{API前缀}{原应用API路径}`（例如 `backendUrl=http://127.0.0.1:8081`、`API前缀=/api` 时，
  `/business_base/api/app/demo/users` 会转发到 `http://127.0.0.1:8081/api/users`）。
- IFrame 类型应用不支持 API 代理。
- 代理入口的鉴权口径：**已登录** + **调用方租户已开通并启用该应用**（在「应用实例」中开通；平台超管放行）。
  未开通时返回 `403 当前租户未开通该应用`。子应用自身的业务权限仍由子应用负责。
- 若希望底座代管子应用的接口权限：在「权限管理」中登记 `appCode = 子应用编码` 的权限点
  （`method` + 相对路径，如 `GET /users/:id`，`path` 为转发到子应用的业务路径）。
  口径为**按应用启用**：该 app_code 有权限点时，未授权用户访问会被底座拦下（`403 无权限访问该子应用接口`）；
  没有登记任何权限点时不做接口级校验，完全交给子应用。平台超管与租户管理员在各自范围内放行。

> 微应用（qiankun 等）接入方式暂未实现，已从应用类型中移除，请使用 iframe 或 proxy 接入。

## 用户信息标准

子应用通过以下方式识别登录用户：

| 来源 | 字段 | 说明 |
|------|------|------|
| URL Query | base_token | 底座 JWT（仅 iframe 接入，由底座前端自动追加） |
| URL Query | user_id / username / tenant_id | 仅 iframe 接入，便于子应用快速展示与调试，可信度低于请求头 |
| Header | X-Base-User-ID | 用户 ID（代理接入） |
| Header | X-Base-Username | 用户名（代理接入） |
| Header | X-Base-Tenant-ID | 租户 ID（代理接入，平台超管为 `0`） |

## 适配器示例

### 后端适配器（Go / Gin）

参考 [adapter-go/middleware.go](adapter-go/middleware.go)：Gin 中间件读取底座请求头并设置上下文。

```go
import "base/integration/adapter-go"

r.Use(adapter.BaseAuthMiddleware())
```

### 前端适配器（TypeScript）

参考 [adapter-ts/types.ts](adapter-ts/types.ts)：定义用户类型与 token/用户信息本地存取工具。

```ts
import { getBaseToken, setBaseToken, getBaseUser, setBaseUser } from './adapter-ts'
```

## 安全建议

- 生产环境使用 HTTPS。
- 子应用后端应校验 `X-Base-*` 请求头来源 IP 或签名，防止伪造。
- Token 有效期与底座 JWT 一致，子应用需处理 401 并跳转回底座登录页。
- 子应用本身仍需做权限校验，不可仅依赖底座请求头。

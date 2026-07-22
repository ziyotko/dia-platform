# 子应用接入标准

本目录提供已有 `frontend + backend` 系统接入 Base 底座平台的规范、示例代码与最佳实践。

## 接入前准备

1. 在 Base 后台「应用管理」中注册应用，获取 `appCode`。
   - 接入类型 `type` 可选：`iframe` / `proxy` / `micro`。
   - 若使用 API 代理，填写 `backendUrl`。
   - 若使用 IFrame，填写 `frontendUrl`。
2. 在「应用实例」中为对应租户开通该应用。
3. 在「菜单管理」中为该应用配置菜单（`appCode` 与注册时一致）。

## 接入方式

### 1. IFrame 接入（最快）

- 应用独立部署。
- Base 前端通过 iframe 加载应用前端地址。
- URL 中追加底座传参：`?base_token=xxx&tenant_id=xxx&user_id=xxx&username=xxx`。
- 子应用前端解析 token，调用底座 `/base/api/v1/auth/info` 校验用户信息。
- 适合已有系统改动成本低的场景。

### 2. API 代理接入（推荐）

- 应用后端无需暴露公网，由底座反向代理。
- 底座在请求头中注入当前用户信息：
  - `X-Base-User-ID`
  - `X-Base-Username`
  - `X-Base-Tenant-ID`
- 应用后端读取请求头完成鉴权与数据隔离。
- 代理路径：`/base/api/v1/app/{appCode}/{原应用API路径}`
- IFrame 类型应用不支持 API 代理。

### 3. 微应用/路由接入（体验最佳）

- 子应用按微前端规范导出生命周期（如 qiankun）。
- 子应用菜单、路由注册到底座菜单表。
- 底座负责容器渲染与路由切换。
- 当前底座前端已预留 iframe 容器，微应用接入可按相同菜单机制扩展。

## 用户信息标准

子应用通过以下方式识别登录用户：

| 来源 | 字段 | 说明 |
|------|------|------|
| URL Query | base_token | 底座 JWT |
| Header | X-Base-User-ID | 用户 ID |
| Header | X-Base-Username | 用户名 |
| Header | X-Base-Tenant-ID | 租户 ID |

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

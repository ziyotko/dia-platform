# Base 底座系统

Base 是一个可落地的「管理后台底座平台」，用于将任意已有的 `frontend + backend` 系统作为子应用统一接入、管理、运营。

## 定位

- 底座平台本身提供：统一登录、租户、应用注册、用户/角色/权限、菜单、审计日志、系统设置。
- 已有系统（子应用）按标准接入后，底座负责鉴权、菜单聚合、路由分发、权限校验、日志审计。
- 未来可通过插件/适配器机制横向扩展新能力，无需改动底座核心。

## 技术栈

- 后端：Go 1.21 + Gin + GORM + MySQL + Redis
- 前端：Vue 3 + Vite + Element Plus + Pinia + TypeScript

## 目录结构

```
base/
├── backend/          # 底座后端
├── frontend/         # 底座前端
├── integration/      # 子应用接入标准与 SDK/适配器示例
└── README.md
```

## 快速开始

### 1. 启动后端

```bash
cd base/backend
# 修改 config/config.yaml 中的 MySQL/Redis 配置
go mod tidy
go run cmd/server/main.go
```

服务默认监听 `:8080`，首次启动会自动创建：
- 超级管理员：admin / admin123
- 底座默认菜单（控制台、系统管理、租户、应用、用户、角色、菜单、设置）

### 2. 启动前端

```bash
cd base/frontend
npm install
npm run dev
```

默认访问 `http://localhost:3000/base/`，登录 admin / admin123。

## 核心概念

| 概念 | 说明 |
|------|------|
| 租户（Tenant） | 平台中的组织单位，用户、应用实例均归属租户 |
| 应用（App） | 一个可被接入的子系统，记录名称、编码、前端入口、后端入口、接入方式 |
| 应用实例（AppInstance） | 应用在租户下的开通实例，控制哪些租户能用哪些应用 |
| 用户（User） | 底座平台用户，支持多租户上下文 |
| 角色（Role） | 角色绑定权限点和菜单 |
| 菜单（Menu） | 底座菜单 + 各应用上报/注册的菜单树 |
| 适配器（Adapter） | 子应用按标准实现适配器，底座即可识别并代理/跳转 |

## 接入方式

1. **IFrame 接入**：子应用独立部署，底座通过 iframe 嵌入，统一传 token/租户信息。
2. **API 代理接入**：子应用后端被底座反向代理，底座在请求头中注入用户信息。
3. **微应用/路由接入**：子应用前端以微应用方式注册到底座，底座聚合菜单与路由。

详见 [integration/README.md](./integration/README.md)。

## 与现有 caam-portal 整合建议

1. 保持现有 `frontend/` 和 `backend/` 不变，独立运行。
2. 在 Base 后台「应用管理」注册：`code=caam-portal`、`type=proxy`、`backendUrl=http://localhost:8080`。
3. 在「应用实例」中为租户开通该应用。
4. 在「菜单管理」添加 `appCode=caam-portal` 的菜单，target 设为 `iframe` 并填写现有 frontend 地址。
5. 用户即可从 Base 平台一站式访问 caam-portal。

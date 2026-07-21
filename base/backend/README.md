# Base 底座后端

基于 Go + Gin + GORM + MySQL/Redis 实现。

## 快速开始

```bash
cd base/backend
go mod tidy

# 修改 config/config.yaml 中的数据库配置

# 首次运行前初始化超级管理员（可选，也可调用 /auth/init）

# 启动服务
go run cmd/server/main.go
```

## 目录说明

```
backend/
├── cmd/server/      # 启动入口
├── config/          # 配置
├── internal/
│   ├── adapter/     # 子应用代理与注册表
│   ├── controllers/ # HTTP 控制器
│   ├── middleware/  # JWT、租户、审计、CORS
│   ├── models/      # GORM 模型
│   ├── routes/      # 路由注册
│   └── service/     # 业务逻辑
└── pkg/             # 公共包（db、redis、jwt、response、utils）
```

## 核心接口

- `POST /base/api/v1/auth/login` 登录
- `GET  /base/api/v1/auth/info` 当前用户信息
- `GET  /base/api/v1/auth/menus` 当前用户菜单树
- `GET  /base/api/v1/auth/permissions` 当前用户权限
- `/base/api/v1/tenants` 租户 CRUD
- `/base/api/v1/apps` 应用 CRUD
- `/base/api/v1/app-instances` 应用实例 CRUD
- `/base/api/v1/users` 用户 CRUD
- `/base/api/v1/roles` 角色 CRUD
- `/base/api/v1/menus` 菜单 CRUD
- `/base/api/v1/permissions` 权限 CRUD
- `/base/api/v1/app/:appCode/*path` 子应用统一代理入口

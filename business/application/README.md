# 项目申报评审系统 (Online Application & Review System)

全流程线上申报评审系统：实现「项目申报 → 材料上传 → 在线初审 → 专家评审 → 进度查询 → 结果公示 → 证书下载」全流程线上闭环。

## 项目结构

```
business/application/
├── backend/          # Go 后端 API (Gin + GORM)
│   ├── main.go
│   ├── config.yaml
│   ├── config/       # 配置加载
│   ├── internal/
│   │   ├── controllers/   # 控制器
│   │   ├── models/        # 数据模型
│   │   ├── service/       # 业务服务
│   │   ├── routes/        # 路由注册
│   │   ├── middleware/    # 认证/权限/CORS/日志/IP限制
│   │   └── seed/          # 种子数据（角色/账号/系统配置）
│   └── pkg/          # 工具包 (db, jwt, redis, response, captcha, utils)
│
└── frontend/         # Vue 3 + TypeScript 前端
    ├── index.html
    ├── vite.config.ts
    └── src/
        ├── api/      # API层 (auth, member, admin)
        ├── stores/   # Pinia 状态管理 (user, admin, app)
        ├── router/   # Vue Router
        ├── views/    # 页面（申报人端 + 管理端）
        ├── utils/    # axios 封装、状态常量
        └── styles/   # 全局样式
```

## 快速启动

### 后端

```bash
cd backend

# 1. 修改 config.yaml 中的数据库配置（MySQL）
# 2. 确保 MySQL 中已创建 caam_application 数据库
# 3. 启动服务（端口 8087）
go mod tidy
go run main.go

# 默认管理账号：admin / manager / reviewer，初始密码 1qaz@WSX
```

### 前端

```bash
cd frontend

# 安装依赖
npm install

# 开发模式（端口 3003）
npm run dev

# 生产构建
npm run build
```

## 全流程闭环（7 个环节）

| 环节 | 说明 | 申报人端 | 管理端 |
|------|------|----------|--------|
| 项目申报 | 浏览开放批次、填写项目信息、保存草稿并提交 | ✅ | 批次/类别维护 |
| 材料上传 | 草稿阶段上传 PDF/Word/图片等申报材料 | ✅ | 下载查看 |
| 在线初审 | 管理人对提交的申报进行通过/驳回并填写意见 | 进度可见 | ✅ |
| 专家评审 | 分配评审人、评分（0-100）与评审意见，自动汇总平均分 | 评分可见 | ✅ 分配/评审 |
| 进度查询 | 申报人查看流程时间线与各环节意见 | ✅ | - |
| 结果公示 | 通过/不通过结果确定后公示，发布公示公告 | ✅ 查看 | ✅ 发布 |
| 证书下载 | 对通过项目颁发证书，申报人可下载 | ✅ 下载 | ✅ 颁发 |

## 角色

| 角色 | 说明 |
|------|------|
| 申报人（前端用户） | 注册/登录、申报项目、上传材料、查询进度、下载证书 |
| 管理人 manager | 维护类别/批次、在线初审、分配评审、确定结果、发布公示、颁发证书、管理申报人与通知 |
| 评审人 reviewer | 查看分配给自己的评审任务、在线评分并填写评审意见 |
| 超级管理员 super_admin | 全部权限，含账号与角色管理 |

## API 前缀

`/application/api`

- 公开接口：`/captcha`、`/member/register`、`/member/login`、`/admin/login`
- 申报人接口：`/member/*`（需 JWT）
- 管理接口：`/admin/*`（需 Admin JWT，细粒度权限校验）

## 申报状态流转

```
draft(草稿) → submitted(待初审) → preliminary_rejected(初审驳回)
                              ↘ under_review(待评审) → reviewed(评审完成)
                                                          → passed(通过) / rejected(不通过)
                                                          → published(已公示) → certified(已发证)
```

## 技术栈

- **后端**：Go 1.26, Gin, GORM, MySQL, Redis, JWT, bcrypt, Viper, logrus
- **前端**：Vue 3.4, TypeScript 5.4, Element Plus 2.6, Pinia 2.1, Vue Router 4.3, Vite 5.2
- **数据库**：MySQL (`caam_application`), Redis（验证码 + 防重放）

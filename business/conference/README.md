# 会议管理系统 (Conference Management System)

## 项目结构

```
business/conference/
├── backend/          # Go 后端 API (Gin + GORM)
│   ├── main.go
│   ├── config.yaml
│   ├── config/       # 配置加载
│   ├── internal/
│   │   ├── controllers/   # 14 控制器
│   │   ├── models/        # 15 数据模型
│   │   ├── service/       # 13 业务服务
│   │   ├── routes/        # 路由注册
│   │   ├── middleware/    # 认证/权限/CORS/日志/IP限制
│   │   └── seed/          # 种子数据
│   └── pkg/          # 工具包 (db, jwt, redis, response, utils, captcha)
│
└── frontend/         # Vue 3 + TypeScript 前端
    ├── index.html
    ├── vite.config.ts
    └── src/
        ├── api/      # API层 (auth, member, admin)
        ├── stores/   # Pinia状态管理 (user, admin, app)
        ├── router/   # Vue Router
        ├── views/
        │   ├── layout/    # 布局组件
        │   ├── login/     # 登录/注册
        │   ├── member/    # 会议前台 (10页面)
        │   └── admin/     # 管理后台 (16页面)
        ├── utils/    # axios封装
        └── styles/   # 全局样式
```

## 快速启动

### 后端

```bash
cd backend

# 1. 修改 config.yaml 中的数据库配置
# 2. 确保 MySQL 中已创建 caam_conference 数据库
# 3. 启动服务 (端口 8086)
go run main.go

# 默认管理员: admin / admin123
```

### 前端

```bash
cd frontend

# 安装依赖
npm install

# 开发模式 (端口 3002)
npm run dev

# 生产构建
npm run build
```

## 核心模块 (13个)

| 模块 | 说明 | 后台 | 前台 |
|------|------|------|------|
| 会议管理 | 创建/编辑/关闭会议，配置议程嘉宾 | ✅ | ✅ 浏览/查看 |
| 报名管理 | 报名/审核/候补机制 | ✅ 审核/管理 | ✅ 报名/取消 |
| 签到管理 | 二维码签到/线上自动签到 | ✅ 统计/管理 | ✅ 签到/查看 |
| 投票表决 | 单选/多选/等额/差额，匿名/实名 | ✅ 创建/计票 | ✅ 投票/查看 |
| 收费退费 | 在线支付/退款/开票 | ✅ 台账/审批 | ✅ 支付/退款 |
| 直播录播 | RTMP直播/HLS播放/VOD回放 | ✅ 控制/上传 | ✅ 观看/留言 |
| 问卷调研 | 单选/多选/简答问卷 | ✅ 创建/统计 | ✅ 填写 |
| 学分管理 | 签到自动发放/手动调整 | ✅ 管理/统计 | ✅ 查询 |
| 档案归档 | 自动归档所有会议数据 | ✅ 管理/导出 | - |
| 多角色权限 | 5种角色，数据隔离 | ✅ 权限控制 | - |
| 通知管理 | 站内信/自动通知 | ✅ 发送 | ✅ 接收 |
| 数据看板 | 图表统计/Excel导出 | ✅ | ✅ 简版 |
| 系统日志 | 操作日志/不可篡改 | ✅ 查看 | - |

## 管理角色 (5种)

| 角色 | 权限范围 |
|------|----------|
| 超级管理员 | 全部权限 |
| 会务管理员 | 会议/报名/签到/直播/问卷/档案 |
| 财务 | 订单/收费/退费/开票 |
| 分会管理员 | 仅本分会数据 |
| 监事 | 只读权限 |

## API 前缀

`/conference/api`

- 公开接口: `/captcha`, `/member/register`, `/member/login`, `/admin/login`
- 会员接口: `/member/*` (需JWT)
- 管理接口: `/admin/*` (需Admin JWT)

## 技术栈

- **后端**: Go 1.26.5, Gin, GORM, MySQL, Redis, JWT, bcrypt, Viper, logrus
- **前端**: Vue 3.4, TypeScript 5.4, Element Plus 2.6, Pinia 2.1, Vue Router 4.3, ECharts 6.1, Axios, Vite 5.2
- **数据库**: MySQL (caam_conference), Redis (captcha + anti-replay)

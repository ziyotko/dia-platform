# Base 底座前端

基于 Vue 3.4 + Vite 5 + Element Plus 2.6 + Pinia 2.1 + TypeScript 5.4 实现。

## 快速开始

```bash
cd base/frontend
npm install
npm run dev
```

默认端口 `3000`，开发服务器代理 `/base/api` 到 `http://127.0.0.1:8080`。
访问地址：`http://localhost:3000/base/`，登录 admin / admin123。

## 脚本说明

```bash
npm run dev      # 启动开发服务器
npm run build    # 类型检查 + 生产构建
npm run preview  # 预览生产构建产物
npm run lint     # ESLint 自动修复
```

## 目录说明

```
frontend/
├── src/
│   ├── api/            # API 接口模块（与后端 /base/api/v1 对应）
│   ├── components/     # 公共组件
│   │   └── Breadcrumb.vue
│   ├── router/         # 动态路由生成
│   ├── stores/         # Pinia 状态
│   │   ├── app.ts      # 应用级状态（我的应用列表）
│   │   └── user.ts     # 用户、菜单、权限、登录态
│   ├── styles/         # 全局样式
│   ├── utils/          # 工具函数（Axios 封装等）
│   └── views/          # 页面
│       ├── base/       # 底座管理页面
│       │   ├── app/
│       │   ├── app-instance/
│       │   ├── dashboard/
│       │   ├── dict/
│       │   ├── file/
│       │   ├── iframe/         # iframe 子应用容器
│       │   ├── log/
│       │   ├── login-log/
│       │   ├── menu/
│       │   ├── message/
│       │   ├── message-template/
│       │   ├── organization/
│       │   ├── permission/     # 接口权限管理
│       │   ├── role/
│       │   ├── setting/
│       │   ├── tenant/
│       │   ├── user/
│       │   └── workflow-role/  # 流程角色（审批角色 + 成员配置）
│       ├── error/      # 404
│       ├── layout/     # 布局与弹出式菜单
│       ├── login/      # 登录页
│       └── profile/    # 个人中心
├── vite.config.ts      # Vite 配置（含代理）
└── package.json
```

## 开发配置

[vite.config.ts](vite.config.ts) 关键配置：

```ts
export default defineConfig({
  plugins: [vue()],
  base: '/base/',
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  },
  server: {
    port: 3000,
    proxy: {
      '/base/api': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: true
      }
    }
  }
})
```

## 动态路由

菜单从后端 `/base/api/v1/auth/menus` 获取后，在 [src/router/index.ts](src/router/index.ts) 中按 `type` 生成路由：

- `directory`：生成嵌套路由（无组件，用于组织子菜单）。
- `menu`：绑定 `component` 字段对应的 `.vue` 文件。
- `target === 'iframe'`：使用 `views/base/iframe/index.vue` 容器加载外部地址，URL 记录在 `meta.url`。

登录成功后，后端返回的菜单树会触发 `addDynamicRoutes`，自动注入到底座布局下。

## 布局与菜单

- [src/views/layout/index.vue](src/views/layout/index.vue)：整体布局，左侧为 64px 图标栏，鼠标悬停/点击时右侧弹出 220px 菜单面板。
- [src/views/layout/components/SubMenu.vue](src/views/layout/components/SubMenu.vue)：递归菜单渲染。
- 顶部区域保留面包屑、消息通知、用户头像下拉。

## 新增管理页面

1. 在 `src/views/base/` 下新建页面，例如 `src/views/base/demo/index.vue`。
2. 在 Base 后台「菜单管理」新增菜单，组件路径填写 `base/demo/index.vue`，`appCode` 填写 `base` 或对应子应用编码。
3. 给角色分配该菜单和对应接口权限后刷新即可生效。

## 子应用页面

若菜单 `target` 为 `iframe`，前端会使用 `views/base/iframe/index.vue` 加载 `component` 字段中的外部 URL，并可通过 URL Query 向子应用传递 `base_token` 等信息。

# Base 底座前端

基于 Vue 3 + Vite + Element Plus + Pinia + TypeScript 实现。

## 快速开始

```bash
cd base/frontend
npm install
npm run dev
```

默认端口 3000，代理到 `http://127.0.0.1:8080`。

## 目录说明

```
frontend/
├── src/
│   ├── api/         # API 接口
│   ├── components/  # 公共组件
│   ├── router/      # 动态路由
│   ├── stores/      # Pinia 状态
│   ├── utils/       # 工具函数
│   └── views/       # 页面
│       ├── base/    # 底座管理页面
│       ├── layout/  # 布局组件
│       ├── login/   # 登录页
│       └── profile/ # 个人中心
```

## 动态路由

菜单从后端获取后，按 `type` 生成目录或菜单路由：

- `directory`：生成嵌套路由
- `menu`：绑定组件
- `target=iframe`：使用 iframe 容器加载外部地址

## 新增管理页面

1. 在 `src/views/base/` 下新建页面。
2. 在 Base 后台「菜单管理」新增菜单，组件路径填写 `base/xxx/index.vue`。
3. 给角色分配该菜单后刷新即可生效。

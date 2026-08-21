import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'
import type { MenuItem } from '@/api/menus'

// 预加载所有视图组件（供动态路由匹配使用）
const componentModules = import.meta.glob('/src/views/**/*.vue')
// 同时兼容 @/ 前缀的查找
const componentModulesAlias = import.meta.glob('@/views/**/*.vue')

// 公共路由（无需权限）
export const constantRoutes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/index.vue'),
    meta: { public: true }
  }
]

// 404 路由，等动态路由加载完再注册，避免优先匹配到根路径 /
const notFoundRoute = {
  path: '/:pathMatch(.*)*',
  name: 'NotFound',
  component: () => import('@/views/error/404.vue')
}

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: constantRoutes
})

/**
 * 根据后端返回的组件路径加载组件
 * 支持格式：@/views/xxx/xxx.vue 或 xxx/xxx（自动补全）
 */
function loadComponent(componentPath: string) {
  if (!componentPath) return undefined
  const trimmed = componentPath.trim()
  let path = trimmed

  if (trimmed.startsWith('@/')) {
    path = trimmed.replace('@/', '/src/')
  } else if (trimmed.startsWith('/src/')) {
    path = trimmed
  } else if (trimmed.startsWith('views/')) {
    path = `/src/${trimmed}`
  } else {
    const cleanPath = trimmed.replace(/^\/+/, '')
    const suffix = cleanPath.endsWith('.vue') ? '' : '.vue'
    path = `/src/views/${cleanPath}${suffix}`
  }

  const aliasPath = path.replace('/src/', '@/')
  const module = componentModules[path] || componentModulesAlias[path] || componentModules[aliasPath] || componentModulesAlias[aliasPath]

  if (!module) {
    console.warn(`[动态路由] 组件未找到: ${componentPath} (尝试: ${path}, ${aliasPath})`)
    return undefined
  }
  return module as () => Promise<any>
}

/**
 * 将后端菜单树转换为 Vue Router 路由配置
 */
export function generateRoutes(menus: MenuItem[]): any[] {
  const routes: any[] = []

  for (const menu of menus) {
    // 跳过禁用的菜单
    if (menu.status === 0) continue

    const route: any = {
      path: menu.path,
      name: menu.name,
      meta: {
        title: menu.name,
        icon: menu.icon,
        type: menu.type
      }
    }

    if (menu.type === 'directory') {
      // 目录：有子菜单则递归处理，过滤后无子路由则跳过
      if (menu.children && menu.children.length > 0) {
        route.children = generateRoutes(menu.children)
      }
      if (!route.children || route.children.length === 0) continue
    } else if (menu.type === 'menu') {
      // 菜单：绑定组件，找不到组件则跳过该路由
      const comp = loadComponent(menu.component)
      if (!comp) continue
      route.component = comp
    }

    routes.push(route)
  }

  return routes
}

/**
 * 将生成的路由动态添加到 Layout 下
 */
function findFirstValidRoute(routes: any[], parentPath = ''): string | undefined {
  for (const route of routes) {
    const fullPath = parentPath + (route.path.startsWith('/') ? route.path : '/' + route.path)
    if (route.component) return fullPath
    if (route.children) {
      const child = findFirstValidRoute(route.children, fullPath)
      if (child) return child
    }
  }
  return undefined
}

export function addDynamicRoutes(menus: MenuItem[]) {
  // 先移除旧的动态路由（如果存在）
  const oldLayout = router.getRoutes().find(r => r.name === 'Layout')
  if (oldLayout) {
    router.removeRoute('Layout')
  }

  const children = generateRoutes(menus)

  // 固定路由：个人中心
  children.push({
    path: '/profile',
    name: 'Profile',
    component: () => import('@/views/profile/index.vue'),
    meta: { title: '个人中心', icon: 'User' }
  })

  const redirectPath = findFirstValidRoute(children)

  const layoutRoute: any = {
    path: '/',
    name: 'Layout',
    component: () => import('@/views/layout/index.vue'),
    children
  }

  if (redirectPath) {
    layoutRoute.redirect = redirectPath
  }

  router.addRoute(layoutRoute)

  // 动态路由注册完成后再注册 404，确保 / 能匹配到 Layout 的 redirect
  const oldNotFound = router.getRoutes().find(r => r.name === 'NotFound')
  if (oldNotFound) {
    router.removeRoute('NotFound')
  }
  router.addRoute(notFoundRoute)

  return children
}

// 路由守卫
router.beforeEach(async (to, _from, next) => {
  const userStore = useUserStore()

  // 公共页面直接放行
  if (to.meta.public) {
    next()
    return
  }

  // 未登录跳转到登录页
  if (!userStore.isLoggedIn) {
    next('/login')
    return
  }

  // 已登录但未获取过动态路由，先获取菜单并生成路由
  if (!userStore.hasFetchedMenus) {
    try {
      await userStore.fetchUserMenusAndGenerateRoutes()
      // 重新导航到目标路由，确保新添加的路由生效
      next({ ...to, replace: true })
    } catch {
      next('/login')
    }
    return
  }

  next()
})

export default router

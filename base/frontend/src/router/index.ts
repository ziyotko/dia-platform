import { createRouter, createWebHistory } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import type { Menu } from '@/api/menu'

const componentModules = import.meta.glob('/src/views/**/*.vue')
const componentModulesAlias = import.meta.glob('@/views/**/*.vue')

const constantRoutes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/index.vue'),
    meta: { public: true }
  }
]

const notFoundRoute = {
  path: '/:pathMatch(.*)*',
  name: 'NotFound',
  component: () => import('@/views/error/404.vue')
}

const router = createRouter({
  history: createWebHistory('/base/'),
  routes: constantRoutes
})

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
  return (
    componentModules[path] ||
    componentModulesAlias[path] ||
    componentModules[aliasPath] ||
    componentModulesAlias[aliasPath]
  ) as (() => Promise<any>) | undefined
}

/**
 * 给路由组件注入 name（= 菜单名 = 路由名）。
 * 页面组件文件名都是 index.vue，默认推断出的组件名不唯一，无法用于 keep-alive 的 include，
 * 因此在这里统一把菜单名写成组件名，让 Layout 能按菜单精确控制哪些页面需要缓存。
 */
function withComponentName(loader: () => Promise<any>, name: string) {
  return () => loader().then((mod: any) => ({ ...(mod?.default || mod), name }))
}

export function generateRoutes(menus: Menu[]): any[] {
  const routes: any[] = []
  for (const menu of menus) {
    if (menu.status === 0) continue
    const route: any = {
      path: menu.path,
      name: menu.name,
      meta: {
        title: menu.name,
        icon: menu.icon,
        target: menu.target,
        permission: menu.permission
      }
    }
    if (menu.type === 'directory') {
      if (menu.children && menu.children.length > 0) {
        route.children = generateRoutes(menu.children)
      }
      if (!route.children || route.children.length === 0) continue
    } else if (menu.type === 'menu') {
      if (menu.target === 'iframe') {
        route.component = () => import('@/views/base/iframe/index.vue')
        route.meta.url = menu.component
      } else {
        if (!menu.component) continue
        const comp = loadComponent(menu.component)
        if (!comp) continue
        route.component = withComponentName(comp, menu.name)
      }
    }
    routes.push(route)
  }
  return routes
}

export function addDynamicRoutes(menus: Menu[]) {
  const oldLayout = router.getRoutes().find((r) => r.name === 'Layout')
  if (oldLayout) {
    router.removeRoute('Layout')
  }

  const children = generateRoutes(menus)
  children.push({
    path: '/profile',
    name: 'Profile',
    component: () => import('@/views/profile/index.vue'),
    meta: { title: '个人中心', icon: 'User' }
  })

  const layoutRoute: any = {
    path: '/',
    name: 'Layout',
    component: () => import('@/views/layout/index.vue'),
    children
  }

  const first = findFirstValidRoute(children)
  if (first) {
    layoutRoute.redirect = first
  }

  router.addRoute(layoutRoute)

  const oldNotFound = router.getRoutes().find((r) => r.name === 'NotFound')
  if (oldNotFound) {
    router.removeRoute('NotFound')
  }
  router.addRoute(notFoundRoute)

  return children
}

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

router.beforeEach(async (to, _from, next) => {
  const userStore = useUserStore()

  if (to.meta.public) {
    next()
    return
  }

  if (!userStore.isLoggedIn) {
    next('/login')
    return
  }

  if (!userStore.hasFetchedMenus) {
    try {
      await userStore.fetchUserInfo()
    } catch (error) {
      // 登录态已失效：清空本地会话并回到登录页
      console.error('[base] 获取用户信息失败', error)
      userStore.clearSession()
      next({ path: '/login', replace: true })
      return
    }

    try {
      await userStore.fetchPermissions()
    } catch (error) {
      // 权限标识获取失败不阻断进入系统：按钮级控制降级为仅超管/管理员可见
      console.error('[base] 获取权限标识失败', error)
    }

    let menus: Menu[] = []
    try {
      menus = await userStore.fetchUserMenusAndGenerateRoutes()
    } catch (error) {
      // 菜单接口异常不应把用户卡在登录页：进入系统并提示
      console.error('[base] 获取用户菜单失败', error)
      ElMessage.warning('菜单加载失败，请联系管理员')
      menus = []
    }

    addDynamicRoutes(menus || [])
    const target = findFirstValidRoute(generateRoutes(menus || []))

    if (!target) {
      // 已登录但没有任何菜单权限（例如新租户用户未分配角色/菜单）：
      // 进入系统并落到「个人中心」，避免停在登录页看起来像登录失败
      ElMessage.warning('当前账号未分配菜单权限，请联系管理员')
      next({ path: '/profile', replace: true })
      return
    }

    if (to.path === '/') {
      next({ path: target, replace: true })
    } else {
      next({ ...to, replace: true })
    }
    return
  }

  next()
})

export default router

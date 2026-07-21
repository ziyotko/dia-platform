import { createRouter, createWebHistory } from 'vue-router'
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
        route.component = comp
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
      const menus = await userStore.fetchUserMenusAndGenerateRoutes()
      addDynamicRoutes(menus)
      const first = findFirstValidRoute(generateRoutes(menus))
      if (to.path === '/' && first) {
        next({ path: first, replace: true })
      } else {
        next({ ...to, replace: true })
      }
    } catch {
      next('/login')
    }
    return
  }

  next()
})

export default router

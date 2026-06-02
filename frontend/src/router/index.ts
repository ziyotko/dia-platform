import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/login/index.vue'),
      meta: { public: true }
    },
    {
      path: '/',
      name: 'Layout',
      component: () => import('@/views/layout/index.vue'),
      redirect: '/dashboard',
      children: [
        {
          path: 'dashboard',
          name: 'Dashboard',
          component: () => import('@/views/dashboard/index.vue'),
          meta: { title: '欢迎首页', icon: 'HomeFilled' }
        },
        {
          path: 'system',
          name: 'System',
          meta: { title: '系统管理', icon: 'Setting' },
          children: [
            {
              path: 'users',
              name: 'Users',
              component: () => import('@/views/system/users.vue'),
              meta: { title: '用户管理', icon: 'UserFilled' }
            },
            {
              path: 'roles',
              name: 'Roles',
              component: () => import('@/views/system/roles.vue'),
              meta: { title: '角色管理', icon: 'Avatar' }
            },
            {
              path: 'menus',
              name: 'Menus',
              component: () => import('@/views/system/menus.vue'),
              meta: { title: '菜单管理', icon: 'Menu' }
            },
            {
              path: 'logs',
              name: 'Logs',
              component: () => import('@/views/system/logs.vue'),
              meta: { title: '操作日志', icon: 'List' }
            }
          ]
        },
        {
          path: 'profile',
          name: 'Profile',
          component: () => import('@/views/profile/index.vue'),
          meta: { title: '个人中心', icon: 'User' }
        },
        {
          path: 'settings',
          name: 'Settings',
          component: () => import('@/views/settings/index.vue'),
          meta: { title: '系统设置', icon: 'Setting' }
        },
        {
          path: 'content',
          name: 'Content',
          redirect: '/content/article',
          meta: { title: '内容管理', icon: 'Document' },
          children: [
            {
              path: 'column',
              name: 'Column',
              component: () => import('@/views/content/column.vue'),
              meta: { title: '栏目', icon: 'Memo' }
            },
            {
              path: 'article',
              name: 'Article',
              component: () => import('@/views/content/article.vue'),
              meta: { title: '文章管理', icon: 'Document' }
            },
            {
              path: 'category',
              name: 'Category',
              component: () => import('@/views/content/category.vue'),
              meta: { title: '分类管理', icon: 'Folder' }
            },
            {
              path: 'tag',
              name: 'Tag',
              component: () => import('@/views/content/tag.vue'),
              meta: { title: '标签管理', icon: 'PriceTag' }
            },
            {
              path: 'comment',
              name: 'Comment',
              component: () => import('@/views/content/comment.vue'),
              meta: { title: '评论管理', icon: 'ChatDotSquare' }
            },
            {
              path: 'ad',
              name: 'Ad',
              component: () => import('@/views/content/ad.vue'),
              meta: { title: '广告管理', icon: 'Promotion' }
            },
            {
              path: 'link',
              name: 'Link',
              component: () => import('@/views/content/link.vue'),
              meta: { title: '友链管理', icon: 'Link' }
            }
          ]
        }
      ]
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'NotFound',
      component: () => import('@/views/error/404.vue')
    }
  ]
})

router.beforeEach((to, _from, next) => {
  const userStore = useUserStore()
  if (!to.meta.public && !userStore.isLoggedIn) {
    next('/login')
  } else {
    next()
  }
})

export default router
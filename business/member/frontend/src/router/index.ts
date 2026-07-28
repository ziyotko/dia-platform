import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'

const router = createRouter({
  history: createWebHistory('/member/'),
  routes: [
    {
      path: '/',
      component: () => import('@/views/layout/PublicLayout.vue'),
      children: [
        {
          path: '',
          name: 'Home',
          component: () => import('@/views/home/index.vue'),
          meta: { public: true }
        },
        {
          path: 'login',
          name: 'Login',
          component: () => import('@/views/login/index.vue'),
          meta: { public: true }
        },
        {
          path: 'register',
          name: 'Register',
          component: () => import('@/views/register/index.vue'),
          meta: { public: true }
        },
        {
          path: 'reset-password',
          name: 'ResetPassword',
          component: () => import('@/views/login/ResetPassword.vue'),
          meta: { public: true }
        },
        {
          path: 'announcements',
          name: 'Announcements',
          component: () => import('@/views/announcements/index.vue'),
          meta: { public: true }
        },
        {
          path: 'announcements/:id',
          name: 'AnnouncementDetail',
          component: () => import('@/views/announcements/detail.vue'),
          meta: { public: true }
        },
        {
          path: '/:pathMatch(.*)*',
          name: 'NotFound',
          component: () => import('@/views/error/404.vue'),
          meta: { public: true }
        }
      ]
    },
    {
      path: '/member',
      component: () => import('@/views/layout/MemberLayout.vue'),
      redirect: '/member/dashboard',
      children: [
        { path: 'dashboard', name: 'Dashboard', component: () => import('@/views/member/Dashboard.vue') },
        { path: 'profile', name: 'Profile', component: () => import('@/views/member/Profile.vue') },
        { path: 'applications', name: 'Applications', component: () => import('@/views/member/Applications.vue') },
        { path: 'fees', name: 'Fees', component: () => import('@/views/member/Fees.vue') },
        { path: 'certificates', name: 'Certificates', component: () => import('@/views/member/Certificates.vue') },
        { path: 'organizations', name: 'Organizations', component: () => import('@/views/member/Organizations.vue') },
        { path: 'articles', name: 'Articles', component: () => import('@/views/member/Articles.vue') },
        { path: 'messages', name: 'Messages', component: () => import('@/views/member/Messages.vue') },
        { path: 'service', name: 'Service', component: () => import('@/views/member/Service.vue') }
      ]
    },
    {
      path: '/admin',
      component: () => import('@/views/layout/AdminLayout.vue'),
      redirect: '/admin/dashboard',
      meta: { admin: true },
      children: [
        { path: 'dashboard', name: 'AdminDashboard', component: () => import('@/views/admin/Dashboard.vue') },
        { path: 'members', name: 'AdminMembers', component: () => import('@/views/admin/Members.vue') },
        { path: 'applications', name: 'AdminApplications', component: () => import('@/views/admin/Applications.vue') },
        { path: 'fees', name: 'AdminFees', component: () => import('@/views/admin/Fees.vue') },
        { path: 'certificates', name: 'AdminCertificates', component: () => import('@/views/admin/Certificates.vue') },
        { path: 'organizations', name: 'AdminOrganizations', component: () => import('@/views/admin/Organizations.vue') },
        { path: 'member-levels', name: 'AdminMemberLevels', component: () => import('@/views/admin/MemberLevels.vue') },
        { path: 'messages', name: 'AdminMessages', component: () => import('@/views/admin/Messages.vue') },
        { path: 'articles', name: 'AdminArticles', component: () => import('@/views/admin/Articles.vue') },
        { path: 'announcements', name: 'AdminAnnouncements', component: () => import('@/views/admin/Announcements.vue') }
      ]
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'NotFound',
      component: () => import('@/views/error/404.vue'),
      meta: { public: true }
    }
  ]
})

router.beforeEach((to, _from, next) => {
  const userStore = useUserStore()

  if (to.meta.public) {
    next()
    return
  }

  if (!userStore.isLoggedIn) {
    next('/login')
    return
  }

  if (to.meta.admin && !userStore.isAdmin) {
    next('/member/dashboard')
    return
  }

  next()
})

export default router

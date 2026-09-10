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
        { path: 'dashboard', name: 'Dashboard', meta: { title: '会员首页' }, component: () => import('@/views/member/Dashboard.vue') },
        { path: 'profile', name: 'Profile', meta: { title: '我的资料' }, component: () => import('@/views/member/Profile.vue') },
        { path: 'applications', name: 'Applications', meta: { title: '我的申请' }, component: () => import('@/views/member/Applications.vue') },
        { path: 'fees', name: 'Fees', meta: { title: '会费管理' }, component: () => import('@/views/member/Fees.vue') },
        { path: 'certificates', name: 'Certificates', meta: { title: '我的证书' }, component: () => import('@/views/member/Certificates.vue') },
        { path: 'organizations', name: 'Organizations', meta: { title: '加入信息' }, component: () => import('@/views/member/Organizations.vue') },
        { path: 'articles', name: 'Articles', meta: { title: '我的文章' }, component: () => import('@/views/member/Articles.vue') },
        { path: 'messages', name: 'Messages', meta: { title: '会员留言' }, component: () => import('@/views/member/Messages.vue') },
        { path: 'service', name: 'Service', meta: { title: '服务中心' }, component: () => import('@/views/member/Service.vue') }
      ]
    },
    {
      path: '/admin',
      component: () => import('@/views/layout/AdminLayout.vue'),
      redirect: '/admin/dashboard',
      meta: { admin: true },
      children: [
        { path: 'dashboard', name: 'AdminDashboard', meta: { title: '管理首页' }, component: () => import('@/views/admin/Dashboard.vue') },
        { path: 'members', name: 'AdminMembers', meta: { title: '会员管理' }, component: () => import('@/views/admin/Members.vue') },
        { path: 'member-level-changes', name: 'AdminMemberLevelChanges', meta: { title: '会籍变更记录' }, component: () => import('@/views/admin/MemberLevelChanges.vue') },
        { path: 'applications', name: 'AdminApplications', meta: { title: '入会审核' }, component: () => import('@/views/admin/Applications.vue') },
        { path: 'fees', name: 'AdminFees', meta: { title: '会费管理' }, component: () => import('@/views/admin/Fees.vue') },
        { path: 'certificates', name: 'AdminCertificates', meta: { title: '证书管理' }, component: () => import('@/views/admin/Certificates.vue') },
        { path: 'organizations', name: 'AdminOrganizations', meta: { title: '组织机构' }, component: () => import('@/views/admin/Organizations.vue') },
        { path: 'member-levels', name: 'AdminMemberLevels', meta: { title: '会员等级' }, component: () => import('@/views/admin/MemberLevels.vue') },
        { path: 'fee-standards', name: 'AdminFeeStandards', meta: { title: '会费标准' }, component: () => import('@/views/admin/FeeStandards.vue') },
        { path: 'messages', name: 'AdminMessages', meta: { title: '会员留言' }, component: () => import('@/views/admin/Messages.vue') },
        { path: 'articles', name: 'AdminArticles', meta: { title: '文章管理' }, component: () => import('@/views/admin/Articles.vue') },
        { path: 'announcements', name: 'AdminAnnouncements', meta: { title: '公告管理' }, component: () => import('@/views/admin/Announcements.vue') },
        { path: 'system-config', name: 'AdminSystemConfig', meta: { title: '系统管理' }, component: () => import('@/views/admin/SystemConfig.vue') },
        { path: 'profile', name: 'AdminProfile', meta: { title: '我的资料' }, component: () => import('@/views/admin/Profile.vue') }
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

  if (to.meta.admin && userStore.userInfo !== null && !userStore.isAdmin) {
    next('/member/dashboard')
    return
  }

  next()
})

export default router

import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { loginRouteLocation } from '@/utils/returnUrl'
import { useSiteStore } from '@/stores/site'
import { ElMessage } from 'element-plus'

const router = createRouter({
  // 基路径跟随 vite base（由 .env 的 VITE_BASE_PATH 决定），避免与部署子路径不一致
  history: createWebHistory(import.meta.env.BASE_URL),
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
          path: 'announcements',
          name: 'Announcements',
          redirect: '/member/news'
        },
        {
          path: 'announcements/:id',
          name: 'AnnouncementDetail',
          redirect: '/member/news'
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
        { path: 'fees', name: 'Fees', meta: { title: '会费管理', requiresFees: true }, component: () => import('@/views/member/Fees.vue') },
        { path: 'certificates', name: 'Certificates', meta: { title: '我的证书' }, component: () => import('@/views/member/Certificates.vue') },
        { path: 'organizations', name: 'Organizations', meta: { title: '加入信息' }, component: () => import('@/views/member/Organizations.vue') },
        { path: 'articles', name: 'Articles', meta: { title: '我的文章' }, component: () => import('@/views/member/Articles.vue') },
        { path: 'news', name: 'MemberNews', meta: { title: '会员资讯' }, component: () => import('@/views/member/News.vue') }
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
        { path: 'profile-changes', name: 'AdminProfileChanges', meta: { title: '资料变更记录' }, component: () => import('@/views/admin/ProfileChanges.vue') },
        { path: 'applications', name: 'AdminApplications', meta: { title: '入会审核' }, component: () => import('@/views/admin/Applications.vue') },
        { path: 'fees', name: 'AdminFees', meta: { title: '会费管理', requiresFees: true }, component: () => import('@/views/admin/Fees.vue') },
        { path: 'certificates', name: 'AdminCertificates', meta: { title: '证书管理' }, component: () => import('@/views/admin/Certificates.vue') },
        { path: 'organizations', name: 'AdminOrganizations', meta: { title: '组织机构' }, component: () => import('@/views/admin/Organizations.vue') },
        { path: 'member-levels', name: 'AdminMemberLevels', meta: { title: '会员等级' }, component: () => import('@/views/admin/MemberLevels.vue') },
        { path: 'fee-standards', name: 'AdminFeeStandards', meta: { title: '会费标准', requiresFees: true }, component: () => import('@/views/admin/FeeStandards.vue') },
        { path: 'articles', name: 'AdminArticles', meta: { title: '文章管理' }, component: () => import('@/views/admin/Articles.vue') },
        { path: 'announcements', name: 'AdminAnnouncements', meta: { title: '公告管理' }, component: () => import('@/views/admin/Announcements.vue') },
        { path: 'system-config', name: 'AdminSystemConfig', meta: { title: '系统管理' }, component: () => import('@/views/admin/SystemConfig.vue') },
        { path: 'admin-accounts', name: 'AdminAccounts', meta: { title: '管理权限', superAdmin: true }, component: () => import('@/views/admin/AdminAccounts.vue') },
        { path: 'operation-logs', name: 'AdminOperationLogs', meta: { title: '操作日志' }, component: () => import('@/views/admin/OperationLogs.vue') },
        { path: 'profile', name: 'AdminProfile', meta: { title: '我的资料' }, component: () => import('@/views/admin/Profile.vue') }
      ]
    }
  ]
})

router.beforeEach(async (to, _from, next) => {
  const userStore = useUserStore()

  if (to.meta.public) {
    next()
    return
  }

  if (!userStore.isLoggedIn) {
    // 带上原目标地址，登录成功后自动跳回（见 @/utils/returnUrl）
    next(loginRouteLocation(to.fullPath))
    return
  }

  // userInfo 不持久化，刷新后需重新拉取，否则用户名显示、会员状态
  // （如证书下载）以及管理员守卫都会失效
  if (!userStore.userInfo) {
    try {
      await userStore.fetchUserInfo()
    } catch {
      // 复用 store 的统一清理（token/userInfo/menus + localStorage 键），
      // 避免这里漏清新增的会话键
      userStore.clearSession()
      next(loginRouteLocation(to.fullPath))
      return
    }
  }

  if (to.meta.admin && !userStore.isAdmin) {
    next('/member/dashboard')
    return
  }

  if (to.meta.superAdmin && !userStore.isSuperAdmin) {
    next('/admin/dashboard')
    return
  }

  if (to.meta.requiresFees) {
    const siteStore = useSiteStore()
    await siteStore.load()
    if (!siteStore.feesEnabled) {
      ElMessage.warning('会费功能已关闭')
      next(userStore.isAdmin ? '/admin/dashboard' : '/member/dashboard')
      return
    }
  }

  next()
})

export default router

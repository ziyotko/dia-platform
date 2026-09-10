import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useAdminStore } from '@/stores/admin'

const routes: RouteRecordRaw[] = [
  // Public routes
  {
    path: '/',
    component: () => import('@/views/layout/PublicLayout.vue'),
    children: [
      { path: '', redirect: '/login' },
      { path: 'login', name: 'Login', component: () => import('@/views/login/Login.vue'), meta: { title: '申报人登录', public: true } },
      { path: 'register', name: 'Register', component: () => import('@/views/login/Register.vue'), meta: { title: '申报人注册', public: true } },
      { path: 'admin/login', name: 'AdminLogin', component: () => import('@/views/admin/AdminLogin.vue'), meta: { title: '管理端登录', public: true } },
    ]
  },
  // Applicant routes (申报人)
  {
    path: '/member',
    component: () => import('@/views/layout/MemberLayout.vue'),
    children: [
      { path: '', redirect: '/member/dashboard' },
      { path: 'dashboard', name: 'MemberDashboard', component: () => import('@/views/member/Dashboard.vue'), meta: { title: '申报中心' } },
      { path: 'batches', name: 'MemberBatches', component: () => import('@/views/member/BatchList.vue'), meta: { title: '可申报项目' } },
      { path: 'applications', name: 'MemberApplications', component: () => import('@/views/member/ApplicationList.vue'), meta: { title: '我的申报' } },
      { path: 'applications/create', name: 'MemberApplicationCreate', component: () => import('@/views/member/ApplicationForm.vue'), meta: { title: '项目申报' } },
      { path: 'applications/:id', name: 'MemberApplicationDetail', component: () => import('@/views/member/ApplicationDetail.vue'), meta: { title: '申报详情' } },
      { path: 'announcements', name: 'MemberAnnouncements', component: () => import('@/views/member/AnnouncementList.vue'), meta: { title: '结果公示' } },
      { path: 'certificates', name: 'MemberCertificates', component: () => import('@/views/member/CertificateList.vue'), meta: { title: '我的证书' } },
      { path: 'notifications', name: 'MemberNotifications', component: () => import('@/views/member/NotificationList.vue'), meta: { title: '通知中心' } },
      { path: 'profile', name: 'MemberProfile', component: () => import('@/views/member/Profile.vue'), meta: { title: '个人资料' } },
    ]
  },
  // Admin routes (管理人 / 评审人)
  {
    path: '/admin',
    component: () => import('@/views/layout/AdminLayout.vue'),
    children: [
      { path: '', redirect: '/admin/dashboard' },
      { path: 'dashboard', name: 'AdminDashboard', component: () => import('@/views/admin/Dashboard.vue'), meta: { title: '管理看板' } },
      { path: 'categories', name: 'AdminCategories', component: () => import('@/views/admin/CategoryList.vue'), meta: { title: '类别管理' } },
      { path: 'batches', name: 'AdminBatches', component: () => import('@/views/admin/BatchList.vue'), meta: { title: '批次管理' } },
      { path: 'applications', name: 'AdminApplications', component: () => import('@/views/admin/ApplicationList.vue'), meta: { title: '申报管理' } },
      { path: 'applications/:id', name: 'AdminApplicationDetail', component: () => import('@/views/admin/ApplicationDetail.vue'), meta: { title: '申报详情' } },
      { path: 'reviews', name: 'AdminReviews', component: () => import('@/views/admin/ReviewList.vue'), meta: { title: '评审管理' } },
      { path: 'reviews/:id', name: 'AdminReviewScore', component: () => import('@/views/admin/ReviewScore.vue'), meta: { title: '专家评审' } },
      { path: 'announcements', name: 'AdminAnnouncements', component: () => import('@/views/admin/AnnouncementList.vue'), meta: { title: '结果公示' } },
      { path: 'certificates', name: 'AdminCertificates', component: () => import('@/views/admin/CertificateList.vue'), meta: { title: '证书管理' } },
      { path: 'notifications', name: 'AdminNotifications', component: () => import('@/views/admin/NotificationList.vue'), meta: { title: '通知管理' } },
      { path: 'users', name: 'AdminUsers', component: () => import('@/views/admin/UserList.vue'), meta: { title: '申报人管理' } },
      { path: 'admins', name: 'AdminAdmins', component: () => import('@/views/admin/AdminList.vue'), meta: { title: '账号管理' } },
      { path: 'audit', name: 'AdminAudit', component: () => import('@/views/admin/AuditLog.vue'), meta: { title: '系统日志' } },
      { path: 'profile', name: 'AdminProfile', component: () => import('@/views/admin/Profile.vue'), meta: { title: '个人资料' } },
    ]
  },
  // 404
  { path: '/:pathMatch(.*)*', name: 'NotFound', component: () => import('@/views/error/NotFound.vue'), meta: { title: '404' } }
]

const router = createRouter({
  history: createWebHistory('/application/'),
  routes
})

router.beforeEach((to, _from, next) => {
  const userStore = useUserStore()
  const adminStore = useAdminStore()

  if (to.meta.public) return next()

  if (to.path.startsWith('/admin')) {
    if (!adminStore.isLoggedIn) return next('/admin/login')
  } else if (to.path.startsWith('/member')) {
    if (!userStore.isLoggedIn) return next('/login')
  }

  next()
})

export default router

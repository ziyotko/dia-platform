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
      { path: 'login', name: 'Login', component: () => import('@/views/login/Login.vue'), meta: { title: '会员登录', public: true } },
      { path: 'register', name: 'Register', component: () => import('@/views/login/Register.vue'), meta: { title: '会员注册', public: true } },
      { path: 'admin/login', name: 'AdminLogin', component: () => import('@/views/admin/AdminLogin.vue'), meta: { title: '管理员登录', public: true } },
    ]
  },
  // Member routes
  {
    path: '/member',
    component: () => import('@/views/layout/MemberLayout.vue'),
    children: [
      { path: '', redirect: '/member/dashboard' },
      { path: 'dashboard', name: 'MemberDashboard', component: () => import('@/views/member/Dashboard.vue'), meta: { title: '会员中心' } },
      { path: 'meetings', name: 'MeetingList', component: () => import('@/views/member/MeetingList.vue'), meta: { title: '会议列表' } },
      { path: 'meetings/:id', name: 'MeetingDetail', component: () => import('@/views/member/MeetingDetail.vue'), meta: { title: '会议详情' } },
      { path: 'registrations', name: 'Registrations', component: () => import('@/views/member/Registrations.vue'), meta: { title: '我的报名' } },
      { path: 'orders', name: 'Orders', component: () => import('@/views/member/Orders.vue'), meta: { title: '我的订单' } },
      { path: 'votes', name: 'VoteList', component: () => import('@/views/member/VoteList.vue'), meta: { title: '投票表决' } },
      { path: 'live/:meetingId', name: 'LiveRoom', component: () => import('@/views/member/LiveRoom.vue'), meta: { title: '会议直播' } },
      { path: 'surveys', name: 'SurveyList', component: () => import('@/views/member/SurveyList.vue'), meta: { title: '问卷调研' } },
      { path: 'credits', name: 'CreditCenter', component: () => import('@/views/member/CreditCenter.vue'), meta: { title: '学分中心' } },
      { path: 'notifications', name: 'Notifications', component: () => import('@/views/member/Notifications.vue'), meta: { title: '通知中心' } },
      { path: 'profile', name: 'MemberProfile', component: () => import('@/views/member/Profile.vue'), meta: { title: '个人资料' } },
    ]
  },
  // Admin routes
  {
    path: '/admin',
    component: () => import('@/views/layout/AdminLayout.vue'),
    children: [
      { path: '', redirect: '/admin/dashboard' },
      { path: 'dashboard', name: 'AdminDashboard', component: () => import('@/views/admin/Dashboard.vue'), meta: { title: '管理后台' } },
      { path: 'meetings', name: 'AdminMeetingList', component: () => import('@/views/admin/MeetingList.vue'), meta: { title: '会议管理' } },
      { path: 'meetings/create', name: 'AdminMeetingCreate', component: () => import('@/views/admin/MeetingForm.vue'), meta: { title: '创建会议' } },
      { path: 'meetings/:id/edit', name: 'AdminMeetingEdit', component: () => import('@/views/admin/MeetingForm.vue'), meta: { title: '编辑会议' } },
      { path: 'meetings/:id', name: 'AdminMeetingDetail', component: () => import('@/views/admin/MeetingDetail.vue'), meta: { title: '会议详情' } },
      { path: 'registrations', name: 'AdminRegistrations', component: () => import('@/views/admin/RegistrationList.vue'), meta: { title: '报名管理' } },
      { path: 'sign-in', name: 'AdminSignIn', component: () => import('@/views/admin/SignInManagement.vue'), meta: { title: '签到管理' } },
      { path: 'votes', name: 'AdminVotes', component: () => import('@/views/admin/VoteList.vue'), meta: { title: '投票管理' } },
      { path: 'votes/create', name: 'AdminVoteCreate', component: () => import('@/views/admin/VoteForm.vue'), meta: { title: '创建投票' } },
      { path: 'votes/:id/edit', name: 'AdminVoteEdit', component: () => import('@/views/admin/VoteForm.vue'), meta: { title: '编辑投票' } },
      { path: 'finance', name: 'AdminFinance', component: () => import('@/views/admin/FinanceList.vue'), meta: { title: '财务管理' } },
      { path: 'finance/ledger', name: 'AdminLedger', component: () => import('@/views/admin/Ledger.vue'), meta: { title: '财务台账' } },
      { path: 'live', name: 'AdminLive', component: () => import('@/views/admin/LiveManagement.vue'), meta: { title: '直播管理' } },
      { path: 'surveys', name: 'AdminSurveys', component: () => import('@/views/admin/SurveyList.vue'), meta: { title: '问卷管理' } },
      { path: 'surveys/create', name: 'AdminSurveyCreate', component: () => import('@/views/admin/SurveyForm.vue'), meta: { title: '创建问卷' } },
      { path: 'surveys/:id/edit', name: 'AdminSurveyEdit', component: () => import('@/views/admin/SurveyForm.vue'), meta: { title: '编辑问卷' } },
      { path: 'surveys/:id/results', name: 'AdminSurveyResults', component: () => import('@/views/admin/SurveyResults.vue'), meta: { title: '问卷结果' } },
      { path: 'credits', name: 'AdminCredits', component: () => import('@/views/admin/CreditManagement.vue'), meta: { title: '学分管理' } },
      { path: 'archives', name: 'AdminArchives', component: () => import('@/views/admin/ArchiveList.vue'), meta: { title: '档案归档' } },
      { path: 'archives/:id', name: 'AdminArchiveDetail', component: () => import('@/views/admin/ArchiveDetail.vue'), meta: { title: '档案详情' } },
      { path: 'notifications', name: 'AdminNotifications', component: () => import('@/views/admin/NotificationList.vue'), meta: { title: '通知管理' } },
      { path: 'users', name: 'AdminUsers', component: () => import('@/views/admin/UserList.vue'), meta: { title: '用户管理' } },
      { path: 'audit', name: 'AdminAudit', component: () => import('@/views/admin/AuditLog.vue'), meta: { title: '系统日志' } },
      { path: 'profile', name: 'AdminProfile', component: () => import('@/views/admin/Profile.vue'), meta: { title: '个人资料' } },
    ]
  },
  // 404
  { path: '/:pathMatch(.*)*', name: 'NotFound', component: () => import('@/views/error/NotFound.vue'), meta: { title: '404' } }
]

const router = createRouter({
  history: createWebHistory('/conference/'),
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

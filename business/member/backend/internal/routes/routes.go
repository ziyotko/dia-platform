package routes

import (
	"time"

	"member/internal/controllers"
	"member/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine) {
	prefix := "/member/api"

	// 登录/注册 IP 级限速（每分钟最多 10 次）
	authRateLimiter := middleware.NewIPRateLimiter(10, time.Minute)
	// check-exists 枚举接口限速（每分钟最多 30 次）
	checkLimiter := middleware.NewIPRateLimiter(30, time.Minute)

	// Controllers
	authCtrl := controllers.AuthController{}
	appCtrl := controllers.ApplicationController{}
	memberCtrl := controllers.MemberController{}
	feeCtrl := controllers.FeeController{}
	certCtrl := controllers.CertificateController{}
	orgCtrl := controllers.OrganizationController{}
	memberOrgCtrl := controllers.MemberOrgController{}
	levelCtrl := controllers.MemberLevelController{}
	feeStdCtrl := controllers.FeeStandardController{}
	msgCtrl := controllers.MessageController{}
	articleCtrl := controllers.ArticleController{}
	announceCtrl := controllers.AnnouncementController{}
	dashCtrl := controllers.DashboardController{}
	certTplCtrl := controllers.CertificateTemplateController{}
	sysConfigCtrl := controllers.SystemConfigController{}

	// === Public routes (no auth) ===
	public := r.Group(prefix)
	{
		// Auth
		public.GET("/captcha", authCtrl.GetCaptcha)
		public.POST("/auth/register", middleware.RateLimitByIP(authRateLimiter, "请求过于频繁，请稍后再试"), authCtrl.Register)
		public.POST("/auth/check-exists", middleware.RateLimitByIP(checkLimiter, "请求过于频繁，请稍后再试"), authCtrl.CheckExists)
		public.POST("/auth/login", middleware.RateLimitByIP(authRateLimiter, "请求过于频繁，请稍后再试"), authCtrl.Login)
		public.POST("/auth/send-reset-email", authCtrl.RequestPasswordReset)
		public.POST("/auth/reset-password", authCtrl.ResetPassword)

		// Site info
		public.GET("/site-info", authCtrl.GetSiteInfo)

		// Download application template
		public.GET("/application-template", authCtrl.DownloadApplicationTemplate)

		// Download charter document
		public.GET("/charter", authCtrl.DownloadCharter)

		// Announcements (public)
		public.GET("/announcements", announceCtrl.GetPublishedAnnouncements)
		public.GET("/announcements/:id", announceCtrl.GetAnnouncement)

		// Organizations (public tree for registration)
		public.GET("/organizations/tree", orgCtrl.GetTree)

		// Member levels (public list for dropdowns)
		public.GET("/member-levels", levelCtrl.List)

		// Public upload (for registration certificate upload)
		public.POST("/upload", authCtrl.UploadFile)
	}

	// === Member routes (auth required) ===
	member := r.Group(prefix)
	member.Use(middleware.Auth())
	{
		// Profile
		member.GET("/member/profile", authCtrl.GetProfile)
		member.PUT("/member/profile", authCtrl.UpdateProfile)
		member.PUT("/member/change-password", authCtrl.ChangePassword)

		// Dashboard
		member.GET("/member/dashboard", dashCtrl.GetMemberDashboard)

		// Applications
		member.POST("/applications", appCtrl.CreateApplication)
		member.POST("/applications/draft", appCtrl.SaveDraft)
		member.POST("/applications/:id/withdraw", appCtrl.WithdrawApplication)
		member.GET("/applications", appCtrl.GetMyApplications)
		member.GET("/applications/:id", appCtrl.GetApplication)

		// Fees
		member.GET("/fees", feeCtrl.GetMyFees)
		member.POST("/fees/:id/pay", feeCtrl.PayFee)
		member.POST("/fees/:id/invoice", feeCtrl.ApplyInvoice)

		// Certificates
		member.GET("/certificates", certCtrl.GetMyCertificates)
		member.GET("/certificates/:id", certCtrl.GetCertificate)
		member.POST("/certificates/renew", certCtrl.RenewCertificate)

		// Joined organizations
		member.GET("/member/orgs", memberOrgCtrl.GetMyOrgs)
		member.POST("/member/orgs", memberOrgCtrl.JoinOrg)
		member.DELETE("/member/orgs/:id", memberOrgCtrl.LeaveOrg)

		// Messages
		member.POST("/messages", msgCtrl.CreateMessage)
		member.GET("/messages", msgCtrl.GetMyMessages)
		member.GET("/messages/:id", msgCtrl.GetMessage)

		// Articles (member)
		member.GET("/article-categories", articleCtrl.ListCategories)
		member.POST("/articles", articleCtrl.CreateArticle)
		member.PUT("/articles/:id", articleCtrl.UpdateArticle)
		member.DELETE("/articles/:id", articleCtrl.DeleteArticle)
		member.GET("/articles", articleCtrl.GetMyArticles)
		member.GET("/articles/:id", articleCtrl.GetArticle)

		// Articles (public published)
		member.GET("/published-articles", articleCtrl.ListArticles)
		member.GET("/published-articles/:id", articleCtrl.GetPublishedArticle)
	}

	// === Admin routes (auth + admin role) ===
	admin := r.Group(prefix)
	admin.Use(middleware.Auth(), middleware.AdminOnly())
	{
		// Member management
		admin.GET("/admin/members", memberCtrl.ListMembers)
		admin.GET("/admin/members/:id", memberCtrl.GetMember)
		admin.PUT("/admin/members/:id/status", memberCtrl.UpdateMemberStatus)
		admin.PUT("/admin/members/:id/level", memberCtrl.UpdateMemberLevel)
		admin.GET("/admin/members/:id/level-options", memberCtrl.GetMemberLevelOptions)
		admin.DELETE("/admin/members/:id", memberCtrl.DeleteMember)
		admin.GET("/admin/member-stats", memberCtrl.GetMemberStats)

		// Application review
		admin.GET("/admin/applications", appCtrl.ListApplications)
		admin.PUT("/admin/applications/:id/review", appCtrl.ReviewApplication)

		// Fee management
		admin.GET("/admin/members/:id/fee-info", feeCtrl.GetMemberFeeInfo)
		admin.POST("/admin/fees", feeCtrl.CreateFee)
		admin.PUT("/admin/fees/:id", feeCtrl.UpdateFee)
		admin.POST("/admin/fees/:id/confirm", feeCtrl.ConfirmFee)
		admin.DELETE("/admin/fees/:id", feeCtrl.DeleteFee)
		admin.GET("/admin/fees", feeCtrl.ListAllFees)
		admin.POST("/admin/fees/:id/issue-invoice", feeCtrl.IssueInvoice)

		// Certificate management
		admin.POST("/admin/certificates", certCtrl.CreateCertificate)
		admin.PUT("/admin/certificates/:id", certCtrl.UpdateCertificate)
		// Certificate template management
		admin.GET("/admin/certificate-templates", certTplCtrl.List)
		admin.GET("/admin/certificate-templates/:id", certTplCtrl.Get)
		admin.POST("/admin/certificate-templates", certTplCtrl.Create)
		admin.PUT("/admin/certificate-templates/:id", certTplCtrl.Update)
		admin.DELETE("/admin/certificate-templates/:id", certTplCtrl.Delete)

		// Organization management
		admin.POST("/admin/organizations", orgCtrl.CreateOrganization)
		admin.PUT("/admin/organizations/:id", orgCtrl.UpdateOrganization)
		admin.DELETE("/admin/organizations/:id", orgCtrl.DeleteOrganization)
		admin.GET("/admin/organizations/:id", orgCtrl.GetOrganization)
		admin.GET("/admin/organizations/:id/levels", orgCtrl.GetOrgLevels)
		admin.PUT("/admin/organizations/:id/levels", orgCtrl.SetOrgLevels)

		// Member level management
		admin.GET("/admin/member-levels", levelCtrl.List)
		admin.POST("/admin/member-levels", levelCtrl.Create)
		admin.PUT("/admin/member-levels/:id", levelCtrl.Update)
		admin.DELETE("/admin/member-levels/:id", levelCtrl.Delete)
		admin.PUT("/admin/member-levels/:id/move-up", levelCtrl.MoveUp)
		admin.PUT("/admin/member-levels/:id/move-down", levelCtrl.MoveDown)

		// Fee standard management
		admin.GET("/admin/fee-standards", feeStdCtrl.ListAll)
		admin.GET("/admin/fee-standards/levels/:levelId", feeStdCtrl.ListByLevel)
		admin.POST("/admin/fee-standards", feeStdCtrl.Upsert)
		admin.POST("/admin/fee-standards/batch", feeStdCtrl.BatchUpsert)
		admin.DELETE("/admin/fee-standards/:id", feeStdCtrl.Delete)

		// Message management
		admin.GET("/admin/messages", msgCtrl.ListAllMessages)
		admin.PUT("/admin/messages/:id/reply", msgCtrl.ReplyMessage)
		admin.DELETE("/admin/messages/:id", msgCtrl.DeleteMessage)

		// Article management
		admin.POST("/admin/article-categories", articleCtrl.CreateCategory)
		admin.PUT("/admin/article-categories/:id", articleCtrl.UpdateCategory)
		admin.DELETE("/admin/article-categories/:id", articleCtrl.DeleteCategory)
		admin.GET("/admin/articles", articleCtrl.ListAllArticles)
		admin.PUT("/admin/articles/:id/review", articleCtrl.ReviewArticle)
		admin.DELETE("/admin/articles/:id", articleCtrl.AdminDeleteArticle)

		// Announcement management
		admin.POST("/admin/announcements", announceCtrl.CreateAnnouncement)
		admin.PUT("/admin/announcements/:id", announceCtrl.UpdateAnnouncement)
		admin.DELETE("/admin/announcements/:id", announceCtrl.DeleteAnnouncement)
		admin.GET("/admin/announcements", announceCtrl.ListAllAnnouncements)

		// System config management
		admin.GET("/admin/system-configs", sysConfigCtrl.List)
		admin.POST("/admin/system-configs", sysConfigCtrl.Create)
		admin.PUT("/admin/system-configs/:id", sysConfigCtrl.Update)
		admin.DELETE("/admin/system-configs/:id", sysConfigCtrl.Delete)
	}
}

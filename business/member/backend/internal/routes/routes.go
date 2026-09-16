package routes

import (
	"time"

	"member/internal/controllers"
	"member/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine) {
	prefix := "/member/api"

	// 限流统一走 Redis 版 middleware.RateLimitMiddleware（与 portal 一致：多实例共享计数，Redis 不可用时进程内兜底）

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
	opLogCtrl := controllers.OperationLogController{}

	// === Public routes (no auth) ===
	public := r.Group(prefix)
	{
		// Auth
		// 验证码接口独立限流：30 次/分钟（对齐 portal 的 captcha_rate_limit），防刷验证码
		public.GET("/captcha", middleware.RateLimitMiddleware(30, time.Minute), authCtrl.GetCaptcha)
		// 注册限流：10 次/分钟，防刷账号
		public.POST("/auth/register", middleware.RateLimitMiddleware(10, time.Minute), authCtrl.Register)
		// check-exists 枚举接口限流：30 次/分钟，防止探测已注册账号
		public.POST("/auth/check-exists", middleware.RateLimitMiddleware(30, time.Minute), authCtrl.CheckExists)
		// 登录限流：10 次/分钟，防暴力破解
		public.POST("/auth/login", middleware.RateLimitMiddleware(10, time.Minute), authCtrl.Login)

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
	admin.Use(middleware.Auth(), middleware.AdminOnly(), middleware.OperationLog())
	{
		// Member management
		admin.POST("/admin/members", memberCtrl.CreateMember)
		admin.GET("/admin/members", memberCtrl.ListMembers)
		admin.GET("/admin/members/:id", memberCtrl.GetMember)
		admin.PUT("/admin/members/:id/status", memberCtrl.UpdateMemberStatus)
		admin.PUT("/admin/members/:id/level", memberCtrl.UpdateMemberLevel)
		admin.PUT("/admin/members/:id/reset-password", memberCtrl.ResetMemberPassword)
		admin.GET("/admin/members/:id/level-options", memberCtrl.GetMemberLevelOptions)
		admin.GET("/admin/members/:id/level-changes", memberCtrl.GetMemberLevelChanges)
		admin.GET("/admin/members/:id/orgs", memberCtrl.GetMemberJoinedOrgs)
		admin.DELETE("/admin/members/:id", memberCtrl.DeleteMember)
		admin.GET("/admin/member-stats", memberCtrl.GetMemberStats)
		admin.GET("/admin/member-exists", memberCtrl.CheckMemberExists)
		admin.GET("/admin/member-level-changes", memberCtrl.ListLevelChanges)
		admin.GET("/admin/member-level-changes/years", memberCtrl.ListLevelChangeYears)
		admin.GET("/admin/member-level-changes/export", memberCtrl.ExportLevelChanges)
		admin.GET("/admin/member-profile-changes", memberCtrl.ListProfileChanges)

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
		admin.GET("/admin/certificates", certCtrl.ListCertificates)
		admin.POST("/admin/certificates/:id/generate", certCtrl.RegenerateCertificate)
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

		// Operation logs
		admin.GET("/admin/operation-logs", opLogCtrl.List)
	}
}

package routes

import (
	"member/internal/controllers"
	"member/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine) {
	prefix := "/member/api"

	// Controllers
	authCtrl := controllers.AuthController{}
	appCtrl := controllers.ApplicationController{}
	memberCtrl := controllers.MemberController{}
	feeCtrl := controllers.FeeController{}
	certCtrl := controllers.CertificateController{}
	orgCtrl := controllers.OrganizationController{}
	memberOrgCtrl := controllers.MemberOrgController{}
	levelCtrl := controllers.MemberLevelController{}
	msgCtrl := controllers.MessageController{}
	articleCtrl := controllers.ArticleController{}
	announceCtrl := controllers.AnnouncementController{}
	dashCtrl := controllers.DashboardController{}

	// === Public routes (no auth) ===
	public := r.Group(prefix)
	{
		// Auth
		public.GET("/captcha", authCtrl.GetCaptcha)
		public.POST("/auth/register", authCtrl.Register)
		public.POST("/auth/login", authCtrl.Login)
		public.POST("/auth/send-reset-email", authCtrl.RequestPasswordReset)
		public.POST("/auth/reset-password", authCtrl.ResetPassword)

		// Site info
		public.GET("/site-info", authCtrl.GetSiteInfo)

		// Download application template
		public.GET("/application-template", authCtrl.DownloadApplicationTemplate)

		// Announcements (public)
		public.GET("/announcements", announceCtrl.GetPublishedAnnouncements)
		public.GET("/announcements/:id", announceCtrl.GetAnnouncement)

		// Organizations (public tree for registration)
		public.GET("/organizations/tree", orgCtrl.GetTree)

		// Member levels (public list for dropdowns)
		public.GET("/member-levels", levelCtrl.List)
	}

	// === Member routes (auth required) ===
	member := r.Group(prefix)
	member.Use(middleware.Auth())
	{
		// Profile
		member.GET("/member/profile", authCtrl.GetProfile)
		member.PUT("/member/profile", authCtrl.UpdateProfile)
		member.PUT("/member/change-password", authCtrl.ChangePassword)

		// Upload
		member.POST("/upload", authCtrl.UploadFile)

		// Dashboard
		member.GET("/member/dashboard", dashCtrl.GetMemberDashboard)

		// Applications
		member.POST("/applications", appCtrl.CreateApplication)
		member.POST("/applications/draft", appCtrl.SaveDraft)
		member.GET("/applications", appCtrl.GetMyApplications)
		member.GET("/applications/:id", appCtrl.GetApplication)

		// Fees
		member.GET("/fees", feeCtrl.GetMyFees)
		member.POST("/fees/:id/pay", feeCtrl.PayFee)

		// Certificates
		member.GET("/certificates", certCtrl.GetMyCertificates)
		member.GET("/certificates/:id", certCtrl.GetCertificate)

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
		admin.DELETE("/admin/members/:id", memberCtrl.DeleteMember)
		admin.GET("/admin/member-stats", memberCtrl.GetMemberStats)

		// Application review
		admin.GET("/admin/applications", appCtrl.ListApplications)
		admin.PUT("/admin/applications/:id/review", appCtrl.ReviewApplication)

		// Fee management
		admin.POST("/admin/fees", feeCtrl.CreateFee)
		admin.PUT("/admin/fees/:id", feeCtrl.UpdateFee)
		admin.GET("/admin/fees", feeCtrl.ListAllFees)

		// Certificate management
		admin.POST("/admin/certificates", certCtrl.CreateCertificate)
		admin.PUT("/admin/certificates/:id", certCtrl.UpdateCertificate)
		admin.POST("/admin/certificates/generate/:id", certCtrl.GenerateCertificate)

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

		// Message management
		admin.GET("/admin/messages", msgCtrl.ListAllMessages)
		admin.PUT("/admin/messages/:id/reply", msgCtrl.ReplyMessage)

		// Article management
		admin.POST("/admin/article-categories", articleCtrl.CreateCategory)
		admin.PUT("/admin/article-categories/:id", articleCtrl.UpdateCategory)
		admin.DELETE("/admin/article-categories/:id", articleCtrl.DeleteCategory)
		admin.GET("/admin/articles", articleCtrl.ListAllArticles)
		admin.PUT("/admin/articles/:id/review", articleCtrl.ReviewArticle)

		// Announcement management
		admin.POST("/admin/announcements", announceCtrl.CreateAnnouncement)
		admin.PUT("/admin/announcements/:id", announceCtrl.UpdateAnnouncement)
		admin.DELETE("/admin/announcements/:id", announceCtrl.DeleteAnnouncement)
		admin.GET("/admin/announcements", announceCtrl.ListAllAnnouncements)
	}
}

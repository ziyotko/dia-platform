package routes

import (
	"time"

	"application/internal/controllers"
	"application/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine) {
	api := r.Group("/application/api")

	// === Controllers ===
	authCtrl := &controllers.AuthController{}
	uploadCtrl := &controllers.UploadController{}
	categoryCtrl := &controllers.CategoryController{}
	batchCtrl := &controllers.BatchController{}
	appCtrl := &controllers.ApplicationController{}
	reviewCtrl := &controllers.ReviewController{}
	resultCtrl := &controllers.ResultController{}
	notifCtrl := &controllers.NotificationController{}
	userCtrl := &controllers.UserController{}
	expertCtrl := &controllers.ExpertController{}
	dashboardCtrl := &controllers.DashboardController{}
	auditCtrl := &controllers.AuditController{}

	// === Public routes ===
	// 验证码接口独立限流：30 次/分钟（对齐 portal 的 captcha_rate_limit），防刷验证码；按真实客户端 IP
	api.GET("/captcha", middleware.RateLimitMiddleware(30, time.Minute), authCtrl.GetCaptcha)
	api.POST("/member/register", authCtrl.UserRegister)
	api.POST("/member/login", authCtrl.UserLogin)
	api.POST("/admin/login", authCtrl.AdminLogin)

	// === Applicant routes (申报人, JWT required) ===
	member := api.Group("/member")
	member.Use(middleware.UserAuth())
	{
		// Profile
		member.GET("/profile", authCtrl.GetUserProfile)
		member.PUT("/profile", authCtrl.UpdateUserProfile)
		member.PUT("/change-password", authCtrl.ChangeUserPassword)

		// Browse published batches & results
		member.GET("/batches", batchCtrl.ListVisible)
		member.GET("/categories", categoryCtrl.List)
		member.GET("/announcements", resultCtrl.ListPublishedAnnouncements)
		// 逐条结果公示：与文本公示公告并列，申报人在同一页看到两种形式
		member.GET("/results", resultCtrl.ListPublishedResults)

		// Applications (项目申报)
		member.GET("/applications", appCtrl.MyApplications)
		member.POST("/applications", appCtrl.Create)
		member.GET("/applications/:id", appCtrl.GetMine)
		member.PUT("/applications/:id", appCtrl.UpdateDraft)
		member.DELETE("/applications/:id", appCtrl.DeleteDraft)
		member.POST("/applications/:id/submit", appCtrl.Submit)
		member.POST("/applications/:id/withdraw", appCtrl.Withdraw)
		member.PUT("/applications/:id/materials", appCtrl.SaveMaterials)

		// Certificates (证书下载)
		member.GET("/certificates", resultCtrl.MyCertificates)

		// Notifications (进度通知)
		member.GET("/notifications", notifCtrl.MyNotifications)
		member.PUT("/notifications/:id/read", notifCtrl.MarkRead)
		member.PUT("/notifications/read-all", notifCtrl.MarkAllRead)
		member.GET("/notifications/unread-count", notifCtrl.UnreadCount)

		// Dashboard
		member.GET("/dashboard", dashboardCtrl.UserStats)

		// Upload (材料上传)
		member.POST("/upload", uploadCtrl.Upload)
	}

	// === Admin routes (管理人/评审人, JWT + permission required) ===
	admin := api.Group("/admin")
	admin.Use(middleware.AdminAuth())
	{
		// Profile
		admin.GET("/profile", authCtrl.GetAdminProfile)
		admin.PUT("/profile", authCtrl.UpdateAdminProfile)
		admin.PUT("/change-password", authCtrl.ChangeAdminPassword)

		// Dashboard
		admin.GET("/dashboard", dashboardCtrl.AdminStats)

		// Upload
		admin.POST("/upload", uploadCtrl.Upload)

		// Categories (项目类别)
		admin.GET("/categories", middleware.PermissionGuard("category:view"), categoryCtrl.List)
		admin.POST("/categories", middleware.PermissionGuard("category:create"), categoryCtrl.Create)
		admin.PUT("/categories/:id", middleware.PermissionGuard("category:edit"), categoryCtrl.Update)
		admin.DELETE("/categories/:id", middleware.PermissionGuard("category:delete"), categoryCtrl.Delete)

		// Batches (申报批次)
		admin.GET("/batches", middleware.PermissionGuard("batch:view"), batchCtrl.List)
		admin.GET("/batches/:id", middleware.PermissionGuard("batch:view"), batchCtrl.GetByID)
		admin.POST("/batches", middleware.PermissionGuard("batch:create"), batchCtrl.Create)
		admin.PUT("/batches/:id", middleware.PermissionGuard("batch:edit"), batchCtrl.Update)
		admin.DELETE("/batches/:id", middleware.PermissionGuard("batch:delete"), batchCtrl.Delete)
		admin.POST("/batches/:id/publish", middleware.PermissionGuard("batch:publish"), batchCtrl.Publish)
		admin.POST("/batches/:id/close", middleware.PermissionGuard("batch:publish"), batchCtrl.Close)
		admin.POST("/batches/:id/start-review", middleware.PermissionGuard("batch:publish"), batchCtrl.StartReview)

		// Applications (项目申报管理 / 初审)
		admin.GET("/applications", middleware.PermissionGuard("application:view"), appCtrl.List)
		admin.GET("/applications/:id", middleware.PermissionGuard("application:view"), appCtrl.GetByID)
		admin.POST("/applications/:id/preliminary", middleware.PermissionGuard("application:preliminary"), appCtrl.PreliminaryReview)
		admin.POST("/applications/:id/assign", middleware.PermissionGuard("review:assign"), appCtrl.AssignReviewers)
		admin.POST("/applications/:id/finalize", middleware.PermissionGuard("result:manage"), appCtrl.Finalize)
		admin.POST("/applications/:id/publish", middleware.PermissionGuard("result:publish"), appCtrl.PublishResult)
		admin.POST("/applications/:id/revoke", middleware.PermissionGuard("result:manage"), appCtrl.RevokeResult)

		// Reviews (专家评审)
		admin.GET("/reviewers", middleware.PermissionGuard("review:assign"), reviewCtrl.ListReviewers)
		admin.GET("/reviews", reviewCtrl.MyAssignments)
		admin.GET("/reviews/:id", reviewCtrl.GetAssignment)
		admin.POST("/reviews/:id/submit", middleware.PermissionGuard("review:score"), reviewCtrl.SubmitReview)
		admin.GET("/review-assignments", middleware.PermissionGuard("review:assign"), reviewCtrl.ListAssignments)
		// 全部评审任务总览
		admin.GET("/review-tasks", middleware.PermissionGuard("review:assign"), reviewCtrl.ListReviewTasks)

		// Announcements (结果公示)
		admin.GET("/announcements", middleware.PermissionGuard("announcement:manage"), resultCtrl.ListAnnouncements)
		admin.POST("/announcements", middleware.PermissionGuard("announcement:manage"), resultCtrl.CreateAnnouncement)
		admin.PUT("/announcements/:id", middleware.PermissionGuard("announcement:manage"), resultCtrl.UpdateAnnouncement)
		admin.DELETE("/announcements/:id", middleware.PermissionGuard("announcement:manage"), resultCtrl.DeleteAnnouncement)
		admin.POST("/announcements/:id/publish", middleware.PermissionGuard("announcement:manage"), resultCtrl.PublishAnnouncement)
		// 用已公示的逐条结果生成公告正文，避免两套公示口径不一致
		admin.GET("/announcements/preview-content", middleware.PermissionGuard("announcement:manage"), resultCtrl.AnnouncementPreview)

		// Certificates (证书管理)
		admin.GET("/certificates", middleware.PermissionGuard("certificate:manage"), resultCtrl.ListCertificates)
		admin.POST("/certificates", middleware.PermissionGuard("certificate:manage"), resultCtrl.IssueCertificate)
		admin.PUT("/certificates/:id", middleware.PermissionGuard("certificate:manage"), resultCtrl.UpdateCertificate)
		admin.POST("/certificates/:id/void", middleware.PermissionGuard("certificate:manage"), resultCtrl.VoidCertificate)

		// Notifications (通知管理)
		admin.POST("/notifications", middleware.PermissionGuard("notification:send"), notifCtrl.Send)

		// Users (申报人管理)
		admin.GET("/users", middleware.PermissionGuard("user:manage"), userCtrl.ListUsers)
		admin.POST("/users", middleware.PermissionGuard("user:manage"), userCtrl.CreateUser)
		admin.PUT("/users/:id", middleware.PermissionGuard("user:manage"), userCtrl.UpdateUser)
		admin.DELETE("/users/:id", middleware.PermissionGuard("user:manage"), userCtrl.DeleteUser)
		admin.PUT("/users/:id/status", middleware.PermissionGuard("user:manage"), userCtrl.SetUserStatus)

		// Experts (专家库管理)
		admin.GET("/experts", middleware.PermissionGuard("expert:manage"), expertCtrl.List)
		admin.POST("/experts", middleware.PermissionGuard("expert:manage"), expertCtrl.Create)
		admin.PUT("/experts/:id", middleware.PermissionGuard("expert:manage"), expertCtrl.Update)
		admin.DELETE("/experts/:id", middleware.PermissionGuard("expert:manage"), expertCtrl.Delete)
		admin.PUT("/experts/:id/status", middleware.PermissionGuard("expert:manage"), expertCtrl.SetStatus)

		// Admins & Roles (账号/角色管理)
		admin.GET("/admins", middleware.PermissionGuard("admin:manage"), userCtrl.ListAdmins)
		admin.POST("/admins", middleware.PermissionGuard("admin:manage"), userCtrl.CreateAdmin)
		admin.PUT("/admins/:id", middleware.PermissionGuard("admin:manage"), userCtrl.UpdateAdmin)
		admin.DELETE("/admins/:id", middleware.PermissionGuard("admin:manage"), userCtrl.DeleteAdmin)
		admin.GET("/roles", middleware.PermissionGuard("role:manage"), userCtrl.ListRoles)

		// Audit logs
		admin.GET("/audit-logs", middleware.PermissionGuard("audit:view"), auditCtrl.List)
	}
}

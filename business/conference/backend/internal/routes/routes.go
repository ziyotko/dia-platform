package routes

import (
	"conference/internal/controllers"
	"conference/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine) {
	api := r.Group("/conference/api")

	// === Controllers ===
	authCtrl := &controllers.AuthController{}
	meetingCtrl := &controllers.MeetingController{}
	regCtrl := &controllers.RegistrationController{}
	signInCtrl := &controllers.SignInController{}
	voteCtrl := &controllers.VoteController{}
	financeCtrl := &controllers.FinanceController{}
	liveCtrl := &controllers.LiveController{}
	surveyCtrl := &controllers.SurveyController{}
	creditCtrl := &controllers.CreditController{}
	archiveCtrl := &controllers.ArchiveController{}
	notifCtrl := &controllers.NotificationController{}
	dashboardCtrl := &controllers.DashboardController{}
	auditCtrl := &controllers.AuditController{}
	userCtrl := &controllers.UserController{}

	// === Public routes ===
	api.GET("/captcha", authCtrl.GetCaptcha)

	// Member auth (public)
	pubMember := api.Group("/member")
	{
		pubMember.POST("/register", authCtrl.MemberRegister)
		pubMember.POST("/login", authCtrl.MemberLogin)
	}

	// Admin auth (public)
	pubAdmin := api.Group("/admin")
	{
		pubAdmin.POST("/login", authCtrl.AdminLogin)
	}

	// === Member routes (JWT required) ===
	memberAuth := api.Group("/member")
	memberAuth.Use(middleware.MemberAuth())
	{
		// Profile
		memberAuth.GET("/profile", authCtrl.GetMemberProfile)
		memberAuth.PUT("/profile", authCtrl.UpdateMemberProfile)
		memberAuth.PUT("/change-password", authCtrl.ChangeMemberPassword)

		// Meetings
		memberAuth.GET("/meetings", meetingCtrl.ListAvailable)
		memberAuth.GET("/meetings/:id", meetingCtrl.GetDetail)

		// Registrations
		memberAuth.POST("/meetings/:id/register", regCtrl.Register)
		memberAuth.GET("/meetings/:id/registration", regCtrl.MyRegistrationStatus)
		memberAuth.GET("/registrations", regCtrl.MyRegistrations)
		memberAuth.DELETE("/registrations/:id", regCtrl.Cancel)

		// Sign-in
		memberAuth.GET("/meetings/:id/sign-in/status", signInCtrl.GetStatus)
		memberAuth.POST("/meetings/:id/sign-in/qrcode", signInCtrl.GenerateQRCode)
		memberAuth.POST("/meetings/:id/sign-in/online", signInCtrl.SignInOnline)
		memberAuth.POST("/meetings/:id/sign-out", signInCtrl.SignOut)

		// Vote
		memberAuth.GET("/votes", voteCtrl.ListAvailable)
		memberAuth.GET("/votes/:id/has-voted", voteCtrl.HasVoted)
		memberAuth.POST("/votes/:id/cast", voteCtrl.CastVote)

		// Finance
		memberAuth.POST("/meetings/:id/orders", financeCtrl.CreateOrder)
		memberAuth.POST("/orders/:id/pay", financeCtrl.PayOrder)
		memberAuth.POST("/orders/:id/refund", financeCtrl.ApplyRefund)
		memberAuth.GET("/orders", financeCtrl.MyOrders)
		memberAuth.PUT("/orders/:id/invoice", financeCtrl.SaveInvoice)
		memberAuth.GET("/orders/:id/invoice", financeCtrl.GetInvoice)

		// Live
		memberAuth.GET("/meetings/:id/live/url", liveCtrl.GetPlayURL)
		memberAuth.GET("/meetings/:id/live/messages", liveCtrl.GetMessages)
		memberAuth.POST("/meetings/:id/live/messages", liveCtrl.SendMessage)
		memberAuth.POST("/meetings/:id/live/viewing", liveCtrl.StartViewing)

		// Survey
		memberAuth.GET("/surveys", surveyCtrl.ListAvailable)
		memberAuth.GET("/surveys/:id/has-submitted", surveyCtrl.HasSubmitted)
		memberAuth.POST("/surveys/:id/submit", surveyCtrl.SubmitAnswers)

		// Credit
		memberAuth.GET("/credits", creditCtrl.MyCredits)

		// Notification
		memberAuth.GET("/notifications", notifCtrl.MyNotifications)
		memberAuth.PUT("/notifications/:id/read", notifCtrl.MarkRead)
		memberAuth.GET("/notifications/unread-count", notifCtrl.GetUnreadCount)

		// Dashboard
		memberAuth.GET("/dashboard", dashboardCtrl.GetMemberStats)
	}

	// === Admin routes (JWT + role guard) ===
	adminAuth := api.Group("/admin")
	adminAuth.Use(middleware.AdminAuth())
	{
		// Profile
		adminAuth.GET("/profile", authCtrl.GetAdminProfile)
		adminAuth.PUT("/profile", authCtrl.UpdateAdminProfile)
		adminAuth.PUT("/change-password", authCtrl.ChangeAdminPassword)

		// Dashboard
		adminAuth.GET("/dashboard", dashboardCtrl.GetAdminStats)

		// Meetings
		adminAuth.POST("/meetings", meetingCtrl.Create)
		adminAuth.PUT("/meetings/:id", meetingCtrl.Update)
		adminAuth.DELETE("/meetings/:id", meetingCtrl.Delete)
		adminAuth.GET("/meetings", meetingCtrl.List)
		adminAuth.GET("/meetings/:id", meetingCtrl.GetByID)
		adminAuth.POST("/meetings/:id/close", meetingCtrl.Close)
		adminAuth.PUT("/meetings/:id/agendas", meetingCtrl.SaveAgendas)
		adminAuth.PUT("/meetings/:id/guests", meetingCtrl.SaveGuests)

		// Registrations
		adminAuth.GET("/registrations", regCtrl.List)
		adminAuth.POST("/registrations/:id/approve", regCtrl.Approve)
		adminAuth.POST("/registrations/:id/reject", regCtrl.Reject)
		adminAuth.POST("/registrations/:id/promote", regCtrl.PromoteWaitlist)
		adminAuth.GET("/registrations/waitlist", regCtrl.GetWaitlist)

		// Sign-in
		adminAuth.GET("/sign-ins", signInCtrl.List)
		adminAuth.POST("/sign-ins/qrcode", signInCtrl.SignInByQR)
		adminAuth.GET("/meetings/:id/sign-in/stats", signInCtrl.GetStats)

		// Vote
		adminAuth.POST("/votes", voteCtrl.Create)
		adminAuth.PUT("/votes/:id", voteCtrl.Update)
		adminAuth.DELETE("/votes/:id", voteCtrl.Delete)
		adminAuth.GET("/votes", voteCtrl.List)
		adminAuth.GET("/votes/:id", voteCtrl.GetByID)
		adminAuth.GET("/votes/:id/results", voteCtrl.GetResults)

		// Finance
		adminAuth.GET("/orders", financeCtrl.ListOrders)
		adminAuth.GET("/refunds", financeCtrl.ListRefunds)
		adminAuth.POST("/refunds/:id/process", financeCtrl.ProcessRefund)
		adminAuth.GET("/ledger", financeCtrl.GetLedger)

		// Live
		adminAuth.POST("/meetings/:id/live/start", liveCtrl.StartLive)
		adminAuth.POST("/meetings/:id/live/stop", liveCtrl.StopLive)
		adminAuth.GET("/meetings/:id/live/config", liveCtrl.GetLiveConfig)
		adminAuth.POST("/meetings/:id/vod/upload", liveCtrl.UploadVod)
		adminAuth.PUT("/meetings/:id/vod/replay", liveCtrl.SetReplayStatus)
		adminAuth.GET("/viewing-logs", liveCtrl.GetViewingLogs)

		// Survey
		adminAuth.POST("/surveys", surveyCtrl.Create)
		adminAuth.PUT("/surveys/:id", surveyCtrl.Update)
		adminAuth.DELETE("/surveys/:id", surveyCtrl.Delete)
		adminAuth.GET("/surveys", surveyCtrl.List)
		adminAuth.GET("/surveys/:id", surveyCtrl.GetByID)
		adminAuth.GET("/surveys/:id/results", surveyCtrl.GetResults)
		adminAuth.GET("/surveys/:id/answers", surveyCtrl.GetAnswerDetails)

		// Credit
		adminAuth.POST("/meetings/:id/credits", creditCtrl.SetMeetingCredits)
		adminAuth.POST("/credits/manual", creditCtrl.ManualAdjust)
		adminAuth.GET("/credits", creditCtrl.ListRecords)
		adminAuth.GET("/credits/stats", creditCtrl.GetStats)

		// Archive
		adminAuth.POST("/meetings/:id/archive", archiveCtrl.Archive)
		adminAuth.GET("/archives", archiveCtrl.List)
		adminAuth.GET("/archives/:id", archiveCtrl.GetByID)
		adminAuth.POST("/archives/:id/materials", archiveCtrl.AddMaterial)

		// Notification
		adminAuth.POST("/notifications", notifCtrl.Send)
		adminAuth.GET("/notifications", notifCtrl.ListSent)
		adminAuth.GET("/notifications/:id/receipt", notifCtrl.GetReadReceipt)

		// Audit
		adminAuth.GET("/audit-logs", auditCtrl.List)

		// User management
		adminAuth.GET("/users", userCtrl.List)
		adminAuth.POST("/users", userCtrl.Create)
		adminAuth.PUT("/users/:id", userCtrl.Update)
		adminAuth.DELETE("/users/:id", userCtrl.Delete)
		adminAuth.GET("/users/:id", userCtrl.GetByID)
		adminAuth.PUT("/users/:id/validity", userCtrl.SetValidity)

		// Admin management
		adminAuth.GET("/admins", userCtrl.ListAdmins)
		adminAuth.POST("/admins", userCtrl.CreateAdmin)
		adminAuth.PUT("/admins/:id", userCtrl.UpdateAdmin)

		// Roles
		adminAuth.GET("/roles", userCtrl.ListRoles)
	}
}

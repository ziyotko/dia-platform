package models

type Admin struct {
	BaseModel
	Username string `gorm:"size:64;uniqueIndex" json:"username"`
	Password string `gorm:"size:128" json:"-"`
	RealName string `gorm:"size:64" json:"realName"`
	Phone    string `gorm:"size:32" json:"phone"`
	Email    string `gorm:"size:128" json:"email"`
	RoleCode string `gorm:"size:32" json:"roleCode"`
	Status   int    `gorm:"default:1" json:"status"`
}

func (Admin) TableName() string {
	return "conference_admins"
}

type Role struct {
	BaseModel
	Code        string `gorm:"size:32;uniqueIndex" json:"code"`
	Name        string `gorm:"size:64" json:"name"`
	Permissions string `gorm:"type:text" json:"permissions"` // JSON array of permission codes
	Description string `gorm:"size:256" json:"description"`
}

func (Role) TableName() string {
	return "conference_roles"
}

// Permission codes
const (
	PermMeetingView   = "meeting:view"
	PermMeetingCreate = "meeting:create"
	PermMeetingEdit   = "meeting:edit"
	PermMeetingDelete = "meeting:delete"
	PermMeetingClose  = "meeting:close"

	PermRegView    = "registration:view"
	PermRegApprove = "registration:approve"
	PermRegImport  = "registration:import"
	PermRegExport  = "registration:export"

	PermSignInView   = "signin:view"
	PermSignInManage = "signin:manage"
	PermSignInExport = "signin:export"

	PermVoteView   = "vote:view"
	PermVoteCreate = "vote:create"
	PermVoteResult = "vote:result"
	PermVoteExport = "vote:export"

	PermFinanceView   = "finance:view"
	PermFinanceManage = "finance:manage"
	PermFinanceRefund = "finance:refund"
	PermFinanceExport = "finance:export"

	PermLiveView   = "live:view"
	PermLiveManage = "live:manage"

	PermSurveyView   = "survey:view"
	PermSurveyCreate = "survey:create"
	PermSurveyResult = "survey:result"
	PermSurveyExport = "survey:export"

	PermCreditView   = "credit:view"
	PermCreditManage = "credit:manage"

	PermArchiveView   = "archive:view"
	PermArchiveExport = "archive:export"

	PermNotifySend = "notification:send"
	PermNotifyView = "notification:view"

	PermUserManage = "user:manage"

	PermAuditView = "audit:view"

	PermDashboardView = "dashboard:view"

	PermRoleManage = "role:manage"
)

// GetDefaultRolePermissions returns permission codes for each role
func GetDefaultRolePermissions() map[string][]string {
	return map[string][]string{
		RoleSuperAdmin: {
			PermMeetingView, PermMeetingCreate, PermMeetingEdit, PermMeetingDelete, PermMeetingClose,
			PermRegView, PermRegApprove, PermRegImport, PermRegExport,
			PermSignInView, PermSignInManage, PermSignInExport,
			PermVoteView, PermVoteCreate, PermVoteResult, PermVoteExport,
			PermFinanceView, PermFinanceManage, PermFinanceRefund, PermFinanceExport,
			PermLiveView, PermLiveManage,
			PermSurveyView, PermSurveyCreate, PermSurveyResult, PermSurveyExport,
			PermCreditView, PermCreditManage,
			PermArchiveView, PermArchiveExport,
			PermNotifySend, PermNotifyView,
			PermUserManage,
			PermAuditView,
			PermDashboardView,
			PermRoleManage,
		},
		RoleMeetingAdmin: {
			PermMeetingView, PermMeetingCreate, PermMeetingEdit, PermMeetingClose,
			PermRegView, PermRegApprove, PermRegImport, PermRegExport,
			PermSignInView, PermSignInManage, PermSignInExport,
			PermLiveView, PermLiveManage,
			PermSurveyView, PermSurveyCreate, PermSurveyResult, PermSurveyExport,
			PermArchiveView, PermArchiveExport,
			PermDashboardView,
		},
		RoleFinance: {
			PermFinanceView, PermFinanceManage, PermFinanceRefund, PermFinanceExport,
			PermDashboardView,
		},
		RoleBranchAdmin: {
			PermMeetingView,
			PermRegView, PermRegApprove,
			PermSignInView,
			PermVoteView, PermVoteResult,
			PermFinanceView,
			PermLiveView,
			PermSurveyView, PermSurveyResult,
			PermCreditView,
			PermArchiveView,
			PermDashboardView,
		},
		RoleSupervisor: {
			PermMeetingView,
			PermRegView, PermRegExport,
			PermSignInView, PermSignInExport,
			PermVoteView, PermVoteResult,
			PermFinanceView, PermFinanceExport,
			PermLiveView,
			PermSurveyView, PermSurveyResult,
			PermCreditView,
			PermArchiveView,
			PermAuditView,
			PermDashboardView,
		},
	}
}

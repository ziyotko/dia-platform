package service

import (
	"member/internal/models"
	"member/pkg/db"
)

type DashboardService struct{}

// MemberDashboard holds the member center dashboard data
type MemberDashboard struct {
	Member              MemberInfo            `json:"member"`
	Application         *AppSummary           `json:"application"`
	UnreadMessages      int64                 `json:"unread_messages"`
	ArticleCount        int64                 `json:"article_count"`
	LatestAnnouncements []models.Announcement `json:"latest_announcements"`
	FeeSummary          *FeeSummary           `json:"fee_summary"`
}

type AppSummary struct {
	ID     uint64 `json:"id"`
	Status string `json:"status"`
	Label  string `json:"label"`
}

type FeeSummary struct {
	PaidCount   int64   `json:"paid_count"`
	UnpaidCount int64   `json:"unpaid_count"`
	TotalPaid   float64 `json:"total_paid"`
}

// GetMemberDashboard returns the member center dashboard data
func (s *DashboardService) GetMemberDashboard(memberID uint64) (*MemberDashboard, error) {
	var member models.Member
	if err := db.DB.First(&member, memberID).Error; err != nil {
		return nil, err
	}

	dash := &MemberDashboard{
		Member: MemberInfo{
			ID:          member.ID,
			Username:    member.Username,
			Mobile:      member.Mobile,
			Email:       member.Email,
			MemberType:  member.MemberType,
			MemberLevel: member.MemberLevel,
			Status:      member.Status,
			IsAdmin:     member.IsAdmin,
			Avatar:      member.Avatar,
			CompanyName: member.CompanyName,
		},
	}

	// Latest application
	var app models.Application
	if err := db.DB.Where("member_id = ?", memberID).Order("created_at DESC").First(&app).Error; err == nil {
		dash.Application = &AppSummary{
			ID:     app.ID,
			Status: app.Status,
			Label:  statusLabel(app.Status),
		}
	}

	// Unread messages
	db.DB.Model(&models.MemberMessage{}).Where("member_id = ? AND status = ?", memberID, models.MessageStatusUnread).Count(&dash.UnreadMessages)

	// Article count
	db.DB.Model(&models.Article{}).Where("member_id = ?", memberID).Count(&dash.ArticleCount)

	// Latest announcements
	db.DB.Where("published_at IS NOT NULL").Order("is_pinned DESC, published_at DESC").Limit(5).Find(&dash.LatestAnnouncements)

	// Fee summary
	fs := &FeeSummary{}
	db.DB.Model(&models.FeeRecord{}).Where("member_id = ? AND status = ?", memberID, models.FeeStatusPaid).Count(&fs.PaidCount)
	db.DB.Model(&models.FeeRecord{}).Where("member_id = ? AND status = ?", memberID, models.FeeStatusUnpaid).Count(&fs.UnpaidCount)
	db.DB.Model(&models.FeeRecord{}).Where("member_id = ? AND status = ?", memberID, models.FeeStatusPaid).
		Select("COALESCE(SUM(amount), 0)").Scan(&fs.TotalPaid)
	dash.FeeSummary = fs

	return dash, nil
}

func statusLabel(status string) string {
	labels := map[string]string{
		models.AppStatusDraft:         "草稿",
		models.AppStatusPendingReview: "待审核",
		models.AppStatusApproved:      "已通过",
		models.AppStatusRejected:      "已拒绝",
	}
	if l, ok := labels[status]; ok {
		return l
	}
	return status
}

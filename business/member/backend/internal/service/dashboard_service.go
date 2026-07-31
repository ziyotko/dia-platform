package service

import (
	"sort"

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
	LatestFeeLevel      string                `json:"latest_fee_level"`
	Organizations       []DashboardOrgInfo    `json:"organizations"`
}

// DashboardOrgInfo holds org display info for the dashboard
type DashboardOrgInfo struct {
	ID         uint64 `json:"id"`
	OrgID      uint64 `json:"org_id"`
	OrgName    string `json:"org_name"`
	LevelName  string `json:"level_name"`
	JoinedAt   string `json:"joined_at"`
	IsFeeBased bool   `json:"is_fee_based"` // true if from approved application (paid join)
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

	// Latest paid fee level — level name from the newest (most recent year) paid fee record
	var latestFee models.FeeRecord
	if err := db.DB.Where("member_id = ? AND status = ?", memberID, models.FeeStatusPaid).
		Order("year DESC, id DESC").First(&latestFee).Error; err == nil {
		dash.LatestFeeLevel = latestFee.LevelName
	}

	// Organizations — query all from member_user_orgs, mark fee-based if has approved application
	var orgInfos []DashboardOrgInfo

	// Collect approved applications (fee-based joins)
	approvedOrgIDs := make(map[uint64]bool)
	var approvedApps []models.Application
	db.DB.Where("member_id = ? AND status = ?", memberID, models.AppStatusApproved).
		Preload("Org").Find(&approvedApps)
	for _, a := range approvedApps {
		if a.Org.ID == 0 {
			continue
		}
		approvedOrgIDs[a.OrgID] = true
		orgInfos = append(orgInfos, DashboardOrgInfo{
			ID:         a.ID,
			OrgID:      a.OrgID,
			OrgName:    a.Org.Name,
			LevelName:  latestFee.LevelName,
			JoinedAt:   a.CreatedAt.Local().Format("2006-01-02"),
			IsFeeBased: true,
		})
	}

	// Query all org memberships from member_user_orgs (direct joins)
	var memberOrgs []models.MemberOrganization
	db.DB.Where("member_id = ?", memberID).Preload("Org").Find(&memberOrgs)
	for _, mo := range memberOrgs {
		if mo.Org.ID == 0 {
			continue
		}
		// Skip orgs already added via approved application (avoid duplicates)
		if approvedOrgIDs[mo.OrgID] {
			continue
		}
		// Resolve level name from LevelID
		levelName := ""
		if mo.LevelID > 0 {
			var lvl models.MemberLevel
			if err := db.DB.First(&lvl, mo.LevelID).Error; err == nil {
				levelName = lvl.Name
			}
		}
		orgInfos = append(orgInfos, DashboardOrgInfo{
			ID:         mo.ID,
			OrgID:      mo.OrgID,
			OrgName:    mo.Org.Name,
			LevelName:  levelName,
			JoinedAt:   mo.JoinedAt.Local().Format("2006-01-02"),
			IsFeeBased: false,
		})
	}

	// Sort: fee-based first, then direct joins
	sort.SliceStable(orgInfos, func(i, j int) bool {
		if orgInfos[i].IsFeeBased != orgInfos[j].IsFeeBased {
			return orgInfos[i].IsFeeBased && !orgInfos[j].IsFeeBased
		}
		return false
	})

	dash.Organizations = orgInfos

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

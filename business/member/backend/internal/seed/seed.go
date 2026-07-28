package seed

import (
	"member/internal/models"
	"member/pkg/db"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func Run() {
	// Create default admin account
	createAdmin()

	// Create default organizations
	createDefaultOrgs()

	// Create default article categories
	createArticleCategories()

	// Create system config
	createSystemConfig()
}

func createAdmin() {
	var count int64
	db.DB.Model(&models.Member{}).Where("is_admin = ?", true).Count(&count)
	if count > 0 {
		return
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	admin := models.Member{
		Username:    "admin",
		Password:    string(hashed),
		Mobile:      "13800000000",
		Email:       "admin@ceeia.com",
		MemberType:  models.MemberTypeUnit,
		Status:      models.MemberStatusActive,
		IsAdmin:     true,
		CompanyName: "中国电器工业协会",
	}
	db.DB.Create(&admin)
}

func createDefaultOrgs() {
	var count int64
	db.DB.Model(&models.Organization{}).Count(&count)
	if count > 0 {
		return
	}

	orgs := []models.Organization{
		{Name: "中国电器工业协会", Type: "association", Sort: 1},
		{Name: "变压器分会", ParentID: 1, Type: "branch", Sort: 1},
		{Name: "高压开关分会", ParentID: 1, Type: "branch", Sort: 2},
		{Name: "电线电缆分会", ParentID: 1, Type: "branch", Sort: 3},
		{Name: "电机分会", ParentID: 1, Type: "branch", Sort: 4},
		{Name: "新能源电器分会", ParentID: 1, Type: "branch", Sort: 5},
		{Name: "绝缘材料分会", ParentID: 1, Type: "branch", Sort: 6},
		{Name: "电力电子分会", ParentID: 1, Type: "branch", Sort: 7},
	}
	db.DB.Create(&orgs)
}

func createArticleCategories() {
	var count int64
	db.DB.Model(&models.ArticleCategory{}).Count(&count)
	if count > 0 {
		return
	}

	cats := []models.ArticleCategory{
		{Name: "行业动态", Sort: 1},
		{Name: "技术交流", Sort: 2},
		{Name: "政策法规", Sort: 3},
		{Name: "标准化", Sort: 4},
		{Name: "展会信息", Sort: 5},
	}
	db.DB.Create(&cats)
}

func createSystemConfig() {
	var count int64
	db.DB.Model(&models.SystemConfig{}).Count(&count)
	if count > 0 {
		return
	}

	configs := []models.SystemConfig{
		{Key: "site_name", Value: "中国电器工业协会会员系统", Description: "网站名称"},
		{Key: "site_description", Value: "中国电器工业协会官方会员服务系统", Description: "网站描述"},
		{Key: "bank_name", Value: "中国工商银行北京分行", Description: "开户银行"},
		{Key: "bank_account", Value: "02000041092001234567", Description: "银行账号"},
		{Key: "bank_account_name", Value: "中国电器工业协会", Description: "账户名称"},
		{Key: "fee_amount", Value: "2000", Description: "年会费标准（元）"},
		{Key: "contact_phone", Value: "010-68166500", Description: "联系电话"},
		{Key: "contact_email", Value: "member@ceeia.com", Description: "联系邮箱"},
		{Key: "icp_no", Value: "京ICP备09041796号-1", Description: "ICP备案号"},
	}
	db.DB.Create(&configs)

	// Create a sample announcement
	now := time.Now()
	announcement := models.Announcement{
		Title:       "欢迎使用中国电器工业协会会员系统",
		Content:     "尊敬的会员单位：\n\n欢迎使用中国电器工业协会会员系统。本系统为您提供入会申请、会费管理、证书下载、在线留言等一站式会员服务。\n\n如有任何疑问，请致电010-68166500。",
		Type:        "notice",
		IsPinned:    true,
		PublishedAt: &models.LocalTime{Time: now},
		CreatedBy:   "admin",
	}
	db.DB.Create(&announcement)
}

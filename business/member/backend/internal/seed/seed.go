package seed

import (
	"member/internal/models"
	"member/internal/service"
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

	// Create default member levels
	createMemberLevels()
}

func createAdmin() {
	var count int64
	db.DB.Model(&models.Member{}).Where("is_admin = ?", true).Count(&count)
	if count > 0 {
		return
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(service.DefaultPassword), bcrypt.DefaultCost)
	admin := models.Member{
		Username:    "admin",
		Password:    string(hashed),
		Mobile:      "13800000000",
		Email:       "admin@test.com",
		MemberType:  models.MemberTypeUnit,
		Status:      models.MemberStatusActive,
		IsAdmin:     true,
		CompanyName: "XXX协会",
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
		{Name: "机构", Type: "root", Sort: 0},
	}
	db.DB.Create(&orgs)

	// Get the root ID
	var root models.Organization
	db.DB.Where("type = ?", "root").First(&root)

	// Children: branches + representatives under the single root
	children := []models.Organization{
		{Name: "分会", ParentID: root.ID, Type: "branch", Sort: 1},
		{Name: "代表处", ParentID: root.ID, Type: "representative", Sort: 10},
	}
	db.DB.Create(&children)
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
		{Key: "site_name", Value: "XXXXXXXX协会会员系统", Description: "网站名称"},
		{Key: "site_description", Value: "XXXXXXXXXXXXX协会官方会员服务系统", Description: "网站描述"},
		{Key: "bank_name", Value: "中国工商银行北京分行", Description: "开户银行"},
		{Key: "bank_account", Value: "XXXXXXXXXXXXX7", Description: "银行账号"},
		{Key: "bank_account_name", Value: "XXXXXXXXXXXXX", Description: "账户名称"},
		{Key: "contact_phone", Value: "010-ccccccccc0", Description: "联系电话"},
		{Key: "contact_email", Value: "member@ccccccc.com", Description: "联系邮箱"},
		{Key: "copyright_name", Value: "XXXXXXXXXXXXX", Description: "版权所有者"},
		{Key: "icp_no", Value: "京ICP备0cccc796号-1", Description: "ICP备案号"},
		{Key: "beian_no", Value: "京公网安备1101ccccc号", Description: "网安备案号"},
	}
	db.DB.Create(&configs)

	// Create a sample announcement
	now := time.Now()
	announcement := models.Announcement{
		Title:       "欢迎使用协会会员系统",
		Content:     "尊敬的会员单位：\n\n欢迎使用协会会员系统。本系统为您提供入会申请、会费管理、证书下载、在线留言等一站式会员服务。\n\n如有任何疑问，请致电010-XXXXXXX。",
		Type:        "notice",
		IsPinned:    true,
		PublishedAt: &models.LocalTime{Time: now},
		CreatedBy:   "admin",
	}
	db.DB.Create(&announcement)
}

func createMemberLevels() {
	var count int64
	db.DB.Model(&models.MemberLevel{}).Count(&count)
	if count > 0 {
		return
	}

	levels := []models.MemberLevel{
		{Name: "会员单位", Level: 0, Description: "普通会员单位"},
		{Name: "理事单位", Level: 1, Description: "理事会员单位"},
		{Name: "副理事长单位", Level: 2, Description: "副理事长会员单位"},
		{Name: "理事长单位", Level: 3, Description: "理事长会员单位"},
	}
	db.DB.Create(&levels)
}

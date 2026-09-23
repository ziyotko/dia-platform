package seed

import (
	"member/internal/models"
	"member/internal/service"
	"member/pkg/db"
	"member/pkg/utils"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// defaultFeeAmount 种子会费标准的占位金额：全新库若没有会费标准，会员端无法缴费、
// 管理员新增会员也会写入金额为 0 的费用记录，因此先建一条占位值，由管理员在后台改成实际标准。
const defaultFeeAmount = 2000.00

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

	// 机构关联等级 + 当年会费标准（缺失时补默认值）：
	// 「入会审批 / 管理员新增会员 / 会费确认」都要求这两种配置存在，否则主流程走不通。
	associateDefaultOrgLevels()
	createDefaultFeeStandards()
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

// associateDefaultOrgLevels 为「尚未配置任何关联等级」的机构补齐等级关联（机构 × 全部等级）。
// 入会审批、管理员新增会员、会员加入机构都要求机构已配置等级；全新库若不带默认值，
// 管理员必须先手工配置才能走通主流程。这里按机构逐个判断，只处理空配置的机构，
// 不会覆盖/删减管理员已配置的等级范围。
func associateDefaultOrgLevels() {
	var levels []models.MemberLevel
	if err := db.DB.Order("level ASC").Find(&levels).Error; err != nil || len(levels) == 0 {
		return
	}
	var orgs []models.Organization
	if err := db.DB.Order("id ASC").Find(&orgs).Error; err != nil || len(orgs) == 0 {
		return
	}

	filled := 0
	for _, org := range orgs {
		var count int64
		db.DB.Model(&models.MemberOrgLevel{}).Where("org_id = ?", org.ID).Count(&count)
		if count > 0 {
			continue
		}
		for _, l := range levels {
			if err := db.DB.Create(&models.MemberOrgLevel{OrgID: org.ID, LevelID: l.ID}).Error; err != nil {
				utils.LogWarn("种子数据：为机构 %d 关联等级 %d 失败：%v", org.ID, l.ID, err)
			}
		}
		filled++
	}
	if filled > 0 {
		utils.LogWarn("种子数据：已为 %d 个未配置等级的机构关联全部 %d 个会员等级，请在后台「组织机构」按实际情况调整", filled, len(levels))
	}
}

// createDefaultFeeStandards 为每个会员等级补齐「当年」会费标准（金额为占位值 defaultFeeAmount）。
// 只补齐缺失的（等级 + 当年）组合，不会覆盖管理员已配置的金额。
func createDefaultFeeStandards() {
	var levels []models.MemberLevel
	if err := db.DB.Order("level ASC").Find(&levels).Error; err != nil || len(levels) == 0 {
		return
	}

	year := time.Now().Year()
	created := 0
	for _, l := range levels {
		var count int64
		db.DB.Model(&models.MemberFeeStandard{}).Where("level_id = ? AND year = ?", l.ID, year).Count(&count)
		if count > 0 {
			continue
		}
		if err := db.DB.Create(&models.MemberFeeStandard{LevelID: l.ID, Year: year, Amount: defaultFeeAmount}).Error; err != nil {
			utils.LogWarn("种子数据：创建等级 %d 的 %d 年度会费标准失败：%v", l.ID, year, err)
			continue
		}
		created++
	}
	if created > 0 {
		utils.LogWarn("种子数据：已创建 %d 年度 %d 条会费标准（占位金额 %.0f 元），请在后台「会费标准」中改为实际标准", year, created, defaultFeeAmount)
	}
}

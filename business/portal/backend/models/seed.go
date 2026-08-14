package models

import (
	"gorm.io/gorm"

	"server/utils"
)

// 系统内置默认角色。
// ID 与现有硬编码约定保持一致：
//   - 1 = 超级管理员（拥有一切权限）
//   - 2 = 普通管理员（授权机构/组织管理范围）
//   - 3 = 内容审核（仅审阅/通过/驳回）
//   - 4 = 内容作者（仅自有内容，需审核后发布）
var defaultRoles = []Role{
	{
		Model:       gorm.Model{ID: 1},
		Name:        "超级管理员",
		Code:        "super_admin",
		Description: "系统最高权限角色，拥有平台全部功能模块的访问、配置与管理权限，不受任何权限范围限制。通常仅授予系统运维或平台负责人，请谨慎分配。",
		Status:      1,
	},
	{
		Model:       gorm.Model{ID: 2},
		Name:        "普通管理员",
		Code:        "admin",
		Description: "负责所辖机构/组织范围内的日常运营管理，可管理授权机构下的成员、栏目、内容与基础配置，权限限定在授权范围内，无法操作其他机构及平台级全局设置。",
		Status:      1,
	},
	{
		Model:       gorm.Model{ID: 3},
		Name:        "内容审核",
		Code:        "content_reviewer",
		Description: "专职内容审核角色，仅可对提交审核的内容进行审阅、通过或驳回操作，并可在驳回时附上审核意见；不具备内容的直接发布、编辑或删除权限，亦不可修改栏目与审核流程配置。",
		Status:      1,
	},
	{
		Model:       gorm.Model{ID: 4},
		Name:        "内容作者",
		Code:        "content_author",
		Description: "内容创作角色，可创建和编辑本人创作的内容，并可提交送审；内容须经审核通过后方可发布，不可直接发布，亦不可编辑或删除他人内容。",
		Status:      1,
	},
}

// SeedDefaultRoles 启动时检查默认角色是否存在，不存在则自动创建（幂等，按 code 判定）。
func SeedDefaultRoles() {
	for i := range defaultRoles {
		role := defaultRoles[i]
		var existing Role
		err := utils.DB.Where("code = ?", role.Code).First(&existing).Error
		if err == nil {
			// 已存在，跳过
			continue
		}
		if err != gorm.ErrRecordNotFound {
			utils.Logger.Warnf("检查默认角色[%s]失败: %v", role.Code, err)
			continue
		}
		if createErr := utils.DB.Create(&role).Error; createErr != nil {
			utils.Logger.Warnf("创建默认角色[%s]失败: %v", role.Code, createErr)
			continue
		}
		utils.Logger.Infof("已自动创建默认角色: %s (%s)", role.Name, role.Code)
	}
}

// 系统内置默认用户（与默认角色一一对应）。
// 初始密码统一为 1qaz@WSX，写入时由 User.BeforeCreate 自动进行 SM3 加盐加密。
var defaultUsers = []User{
	{
		// 超级管理员固定为 ID=1，与前端 users.vue 的内置用户保护逻辑保持一致
		Model:    gorm.Model{ID: 1},
		Username: "超级管理员",
		Account:  "admin",
		Email:    "admin@example.com",
		Password: "1qaz@WSX",
		Mobile:   "1",
		RoleIds:  "1",
		Status:   1,
	},
	{
		Username: "普通管理员",
		Account:  "operator",
		Email:    "operator@example.com",
		Password: "1qaz@WSX",
		Mobile:   "2",
		RoleIds:  "2",
		Status:   1,
	},
	{
		Username: "内容审核",
		Account:  "reviewer",
		Email:    "reviewer@example.com",
		Password: "1qaz@WSX",
		Mobile:   "3",
		RoleIds:  "3",
		Status:   1,
	},
	{
		Username: "内容作者",
		Account:  "author",
		Email:    "author@example.com",
		Password: "1qaz@WSX",
		Mobile:   "4",
		RoleIds:  "4",
		Status:   1,
	},
}

// SeedDefaultUsers 启动时检查默认用户是否存在，不存在则自动创建（幂等，按 account 判定）。
func SeedDefaultUsers() {
	for i := range defaultUsers {
		user := defaultUsers[i]
		var existing User
		err := utils.DB.Where("account = ?", user.Account).First(&existing).Error
		if err == nil {
			// 已存在，跳过
			continue
		}
		if err != gorm.ErrRecordNotFound {
			utils.Logger.Warnf("检查默认用户[%s]失败: %v", user.Account, err)
			continue
		}
		if createErr := utils.DB.Create(&user).Error; createErr != nil {
			utils.Logger.Warnf("创建默认用户[%s]失败: %v", user.Account, createErr)
			continue
		}
		utils.Logger.Infof("已自动创建默认用户: %s (%s)", user.Username, user.Account)
	}
}

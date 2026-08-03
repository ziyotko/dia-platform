package seed

import (
	"encoding/json"

	"conference/internal/models"
	"conference/pkg/db"

	"golang.org/x/crypto/bcrypt"
)

func Run() {
	seedRoles()
	seedAdmin()
	seedSystemConfigs()
}

func seedRoles() {
	roles := []models.Role{
		{Code: models.RoleSuperAdmin, Name: "超级管理员", Description: "拥有全部权限"},
		{Code: models.RoleMeetingAdmin, Name: "会务管理员", Description: "会议、报名、签到、直播、问卷、档案管理"},
		{Code: models.RoleFinance, Name: "财务", Description: "订单、收费、退费、开票管理"},
		{Code: models.RoleBranchAdmin, Name: "分会管理员", Description: "仅查看本分会数据"},
		{Code: models.RoleSupervisor, Name: "监事", Description: "只读权限"},
	}

	permMap := models.GetDefaultRolePermissions()
	for _, role := range roles {
		var count int64
		db.DB.Model(&models.Role{}).Where("code = ?", role.Code).Count(&count)
		if count == 0 {
			perms, _ := json.Marshal(permMap[role.Code])
			role.Permissions = string(perms)
			db.DB.Create(&role)
		}
	}
}

func seedAdmin() {
	var count int64
	db.DB.Model(&models.Admin{}).Where("username = ?", "admin").Count(&count)
	if count == 0 {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		admin := models.Admin{
			Username: "admin",
			Password: string(hashedPassword),
			RealName: "超级管理员",
			RoleCode: models.RoleSuperAdmin,
			Status:   1,
		}
		db.DB.Create(&admin)
	}
}

type SystemConfig struct {
	models.BaseModel
	Key   string `gorm:"size:128;uniqueIndex" json:"key"`
	Value string `gorm:"type:text" json:"value"`
}

func (SystemConfig) TableName() string {
	return "conference_system_configs"
}

func seedSystemConfigs() {
	configs := []SystemConfig{
		{Key: "site_name", Value: "会议管理系统"},
		{Key: "copyright", Value: "© 2024 会议管理系统"},
		{Key: "icp_no", Value: ""},
		{Key: "beian_no", Value: ""},
	}
	for _, cfg := range configs {
		var count int64
		db.DB.Model(&SystemConfig{}).Where("`key` = ?", cfg.Key).Count(&count)
		if count == 0 {
			db.DB.Create(&cfg)
		}
	}
}

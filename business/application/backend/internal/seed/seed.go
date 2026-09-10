package seed

import (
	"encoding/json"

	"application/internal/models"
	"application/pkg/db"

	"golang.org/x/crypto/bcrypt"
)

func Run() {
	seedRoles()
	seedAdmins()
	seedSystemConfigs()
}

func seedRoles() {
	roles := []models.Role{
		{Code: models.RoleSuperAdmin, Name: "超级管理员", Description: "拥有全部权限"},
		{Code: models.RoleManager, Name: "管理人", Description: "项目批次、初审、评审分配、结果公示、证书管理"},
		{Code: models.RoleReviewer, Name: "评审人", Description: "专家评审打分"},
	}

	permMap := map[string][]string{
		models.RoleSuperAdmin: {
			"dashboard:view",
			"batch:view", "batch:create", "batch:edit", "batch:delete", "batch:publish",
			"category:view", "category:create", "category:edit", "category:delete",
			"application:view", "application:preliminary",
			"review:assign", "review:score",
			"result:publish", "result:manage",
			"certificate:manage",
			"announcement:manage",
			"notification:send",
			"user:manage", "admin:manage", "role:manage",
			"expert:manage",
			"audit:view",
		},
		models.RoleManager: {
			"dashboard:view",
			"batch:view", "batch:create", "batch:edit", "batch:publish",
			"category:view", "category:create", "category:edit",
			"application:view", "application:preliminary",
			"review:assign",
			"result:publish", "result:manage",
			"certificate:manage",
			"announcement:manage",
			"notification:send",
			"user:manage",
			"expert:manage",
			"audit:view",
		},
		models.RoleReviewer: {
			"dashboard:view",
			"batch:view",
			"application:view",
			"review:score",
		},
	}

	for _, role := range roles {
		perms, _ := json.Marshal(permMap[role.Code])
		var count int64
		db.DB.Model(&models.Role{}).Where("code = ?", role.Code).Count(&count)
		if count == 0 {
			role.Permissions = string(perms)
			db.DB.Create(&role)
		} else {
			db.DB.Model(&models.Role{}).Where("code = ?", role.Code).Update("permissions", string(perms))
		}
	}
}

func seedAdmins() {
	admins := []models.Admin{
		{Username: "admin", RealName: "超级管理员", RoleCode: models.RoleSuperAdmin, Status: 1},
		{Username: "manager", RealName: "项目管理员", RoleCode: models.RoleManager, Status: 1},
		{Username: "reviewer", RealName: "评审专家", RoleCode: models.RoleReviewer, Status: 1},
	}
	for _, a := range admins {
		var count int64
		db.DB.Model(&models.Admin{}).Where("username = ?", a.Username).Count(&count)
		if count == 0 {
			hashed, _ := bcrypt.GenerateFromPassword([]byte("1qaz@WSX"), bcrypt.DefaultCost)
			a.Password = string(hashed)
			db.DB.Create(&a)
		}
	}
}

type SystemConfig struct {
	models.BaseModel
	Key   string `gorm:"size:128;uniqueIndex" json:"key"`
	Value string `gorm:"type:text" json:"value"`
}

func (SystemConfig) TableName() string {
	return "application_system_configs"
}

func seedSystemConfigs() {
	configs := []SystemConfig{
		{Key: "site_name", Value: "项目申报评审系统"},
		{Key: "copyright", Value: "© 2026 项目申报评审系统"},
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

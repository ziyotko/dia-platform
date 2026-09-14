package models

import (
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// 系统内置角色 ID（与 seed.go 默认角色、前端 src/utils/permission.ts 保持一致）
const (
	RoleIDSuperAdmin = 1 // 超级管理员
	RoleIDAdmin      = 2 // 普通管理员
)

// IsAdminRoleID 判断单个角色 ID 是否属于管理员角色。
func IsAdminRoleID(id int) bool {
	return id == RoleIDSuperAdmin || id == RoleIDAdmin
}

// HasAdminRoleIDs 判断角色 ID 列表是否包含管理员角色。
func HasAdminRoleIDs(roleIds []int) bool {
	for _, id := range roleIds {
		if IsAdminRoleID(id) {
			return true
		}
	}
	return false
}

// HasAdminRoleStr 判断逗号分隔的角色 ID 串（User.RoleIds 的存储格式）是否包含管理员角色。
func HasAdminRoleStr(roleIdsStr string) bool {
	for _, part := range strings.Split(roleIdsStr, ",") {
		if id, err := strconv.Atoi(strings.TrimSpace(part)); err == nil && IsAdminRoleID(id) {
			return true
		}
	}
	return false
}

type Role struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Name        string         `gorm:"size:50;not null" json:"name"`
	Code        string         `gorm:"unique;size:50;not null" json:"code"`
	Description string         `gorm:"size:255" json:"description"`
	Status      int            `gorm:"default:1" json:"status"`
	Permissions string         `gorm:"size:500" json:"permissions"`
}

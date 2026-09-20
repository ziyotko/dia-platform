package models

import (
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// 系统内置角色 ID（与 seed.go 默认角色、前端 src/utils/permission.ts 保持一致）。
// 注：ID 2 原为「普通管理员（admin）」，该角色已整体下线移除，不再播种；
// 历史库中若仍存在 ID=2 的角色记录，一律按普通自定义角色对待（不再是内置角色，也不再是管理员）。
// ID 3/4 保持原值，不做重排。
const (
	RoleIDSuperAdmin      = 1 // 管理员（原「超级管理员」，2026-09-21 更名，code 保持 super_admin）
	RoleIDContentReviewer = 3 // 内容审核
	RoleIDContentAuthor   = 4 // 内容作者
)

// builtinRoleIDs 系统内置角色集合，由启动种子维护，不允许改名或删除。
var builtinRoleIDs = map[uint]struct{}{
	RoleIDSuperAdmin:      {},
	RoleIDContentReviewer: {},
	RoleIDContentAuthor:   {},
}

// IsBuiltinRoleID 判断是否系统内置角色。
func IsBuiltinRoleID(id uint) bool {
	_, ok := builtinRoleIDs[id]
	return ok
}

// IsAdminRoleID 判断单个角色 ID 是否属于管理员角色（仅角色 1）。
func IsAdminRoleID(id int) bool {
	return id == RoleIDSuperAdmin
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

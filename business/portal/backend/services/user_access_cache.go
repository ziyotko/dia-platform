package services

import (
	"strings"
	"sync"
	"time"

	"server/models"
	"server/utils"
)

// userAccess 用户的访问上下文：菜单 API 前缀 + 管理员标记。
// 每个请求原都要解析一次（角色 → 权限 → 菜单 → 前缀，共 4~6 次查询），而菜单/角色极少变化，
// 因此按用户缓存一个很短的 TTL。
// 取舍：不做写路径主动失效，避免「漏埋失效点导致权限长期不生效」这类更严重的问题；
// 15s 内收敛，对「改完菜单马上点」几乎无感，同时消除绝大多数重复查询。
type userAccess struct {
	Prefixes []string
	IsAdmin  bool
	ExpireAt time.Time
}

const userAccessCacheTTL = 15 * time.Second

var (
	userAccessMu    sync.RWMutex
	userAccessCache = make(map[uint]userAccess)
)

// GetUserAccess 读取（或构建）用户的访问上下文（菜单 API 前缀 + 管理员标记）
func GetUserAccess(userID uint) (userAccess, error) {
	now := time.Now()
	userAccessMu.RLock()
	entry, ok := userAccessCache[userID]
	userAccessMu.RUnlock()
	if ok && now.Before(entry.ExpireAt) {
		return entry, nil
	}

	prefixes, isAdmin, err := buildUserAccess(userID)
	if err != nil {
		return userAccess{}, err
	}
	entry = userAccess{Prefixes: prefixes, IsAdmin: isAdmin, ExpireAt: now.Add(userAccessCacheTTL)}
	userAccessMu.Lock()
	userAccessCache[userID] = entry
	userAccessMu.Unlock()
	return entry, nil
}

// InvalidateUserAccessCache 立即失效缓存：不传 userID 表示清空全部（菜单/角色写路径可按需调用）
func InvalidateUserAccessCache(userID ...uint) {
	userAccessMu.Lock()
	defer userAccessMu.Unlock()
	if len(userID) == 0 {
		userAccessCache = make(map[uint]userAccess)
		return
	}
	for _, id := range userID {
		delete(userAccessCache, id)
	}
}

// buildUserAccess 解析用户的菜单 API 前缀与管理员标记（口径与 middleware.IsAdminUser 原实现一致）
func buildUserAccess(userID uint) ([]string, bool, error) {
	menuService := &MenuService{}
	menus, err := menuService.GetUserMenus(userID)
	if err != nil {
		return nil, false, err
	}
	prefixes := collectMenuAPIPrefixes(menus)

	var user models.User
	if err := utils.DB.First(&user, userID).Error; err != nil {
		return nil, false, err
	}
	isAdmin := false
	if models.HasAdminRoleStr(user.RoleIds) {
		var enabled int64
		if err := utils.DB.Model(&models.Role{}).
			Where("id = ? AND status = ?", models.RoleIDSuperAdmin, 1).
			Count(&enabled).Error; err == nil && enabled > 0 {
			isAdmin = true
		}
	}
	return prefixes, isAdmin, nil
}

// collectMenuAPIPrefixes 收集菜单（含子菜单）中所有非空的 api_prefix。
// 一个菜单可声明多个前缀（逗号分隔），便于「一个页面调用多个模块接口」的场景。
func collectMenuAPIPrefixes(menus []models.Menu) []string {
	var prefixes []string
	var collect func([]models.Menu)
	collect = func(items []models.Menu) {
		for _, m := range items {
			for _, prefix := range strings.Split(m.APIPrefix, ",") {
				if prefix = strings.TrimSpace(prefix); prefix != "" {
					prefixes = append(prefixes, prefix)
				}
			}
			if len(m.Children) > 0 {
				collect(m.Children)
			}
		}
	}
	collect(menus)
	return prefixes
}

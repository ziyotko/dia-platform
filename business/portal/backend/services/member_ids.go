package services

import (
	"fmt"
	"strconv"
	"strings"

	"server/models"
	"server/utils"
)

// 成员列表存储上限（与 models 中 user_ids 字段长度保持一致）：
// department.user_ids = varchar(500)、organization.user_ids = varchar(1000)。
// 超出上限时 MySQL 非严格模式会静默截断（导致成员丢失），故落库前显式校验。
const (
	departmentMemberIDsMaxChars   = 500
	organizationMemberIDsMaxChars = 1000
)

// parseMemberIDList 解析逗号分隔的成员 ID 字符串（department/organization 的 user_ids 列）：
// 跳过空项与非法值，保持原顺序并去重。
func parseMemberIDList(raw string) []uint {
	if strings.TrimSpace(raw) == "" {
		return []uint{}
	}
	parts := strings.Split(raw, ",")
	result := make([]uint, 0, len(parts))
	seen := make(map[uint]bool, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p == "" {
			continue
		}
		val, err := strconv.ParseUint(p, 10, 32)
		if err != nil || val == 0 || seen[uint(val)] {
			continue
		}
		seen[uint(val)] = true
		result = append(result, uint(val))
	}
	return result
}

// normalizeMemberIDs 规范化成员 ID 列表：去重（保持顺序）→ 剔除不存在的用户 → 校验存储长度。
// 返回规范化后的 ID 列表、可直接入库的逗号分隔字符串，以及被剔除的无效 ID 数量。
// 目的：避免脏 ID（已删除用户、错误 ID）入库，并避免 user_ids 超长被静默截断。
func normalizeMemberIDs(ids []uint, maxChars int) ([]uint, string, int, error) {
	cleaned := make([]uint, 0, len(ids))
	seen := make(map[uint]bool, len(ids))
	for _, id := range ids {
		if id == 0 || seen[id] {
			continue
		}
		seen[id] = true
		cleaned = append(cleaned, id)
	}
	if len(cleaned) == 0 {
		return []uint{}, "", 0, nil
	}

	// 只保留真实存在的用户（models.User 带软删除，已删除用户会被自动排除）
	var existing []uint
	if err := utils.DB.Model(&models.User{}).Where("id IN ?", cleaned).Pluck("id", &existing).Error; err != nil {
		return nil, "", 0, err
	}
	valid := make(map[uint]bool, len(existing))
	for _, id := range existing {
		valid[id] = true
	}

	kept := make([]uint, 0, len(cleaned))
	strs := make([]string, 0, len(cleaned))
	for _, id := range cleaned {
		if !valid[id] {
			continue
		}
		kept = append(kept, id)
		strs = append(strs, strconv.FormatUint(uint64(id), 10))
	}

	joined := strings.Join(strs, ",")
	if maxChars > 0 && len(joined) > maxChars {
		return nil, "", 0, fmt.Errorf("成员数量过多，超出存储上限（最多约 %d 个用户），请减少成员数量后重试", maxChars/2)
	}
	return kept, joined, len(cleaned) - len(kept), nil
}

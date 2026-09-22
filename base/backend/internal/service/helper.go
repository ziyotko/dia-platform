package service

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

// ensureRecordExists 在写操作前确认目标记录存在（且属于当前租户）。
//
// 为什么不用 RowsAffected 判断 Update：MySQL 默认返回「实际变更的行数」而不是「命中的行数」，
// 没改动任何字段时会是 0，会把「原样保存」误判为失败，所以更新前先做一次存在性/归属校验。
func ensureRecordExists(query *gorm.DB, message string) error {
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New(message)
	}
	return nil
}

// ensureDeleteAffected 校验删除确实命中了记录。
// 物理删除会真实删除行，RowsAffected 可靠；命中 0 行说明记录不存在或不属于当前租户，
// 此时必须报错，否则接口返回「删除成功」但数据仍在。
func ensureDeleteAffected(res *gorm.DB, message string) error {
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New(message)
	}
	return nil
}

// isDuplicateEntry 判断错误是否来自唯一索引冲突（MySQL 1062）。
//
// 为什么需要：唯一性最终由数据库索引保证，service 层的前置查询只是友好提示，
// 并发下两个请求可能都通过预检查；此时必须把驱动返回的
// "Error 1062: Duplicate entry '1-john' for key 'uk_base_user_tenant_username'"
// 转成可读文案，而不是把 SQL 细节抛给用户。
func isDuplicateEntry(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate entry") || strings.Contains(msg, "error 1062")
}

// 平台内置数据（tenant_id = 0）对租户可见但不可由租户修改/删除，统一提示文案。
const msgNotOwnedOrMissing = "记录不存在，或不属于当前租户（平台内置数据不可修改）"

// msgNotOwnedOrMissingDelete 同上，用于删除场景。
const msgNotOwnedOrMissingDelete = "记录不存在，或不属于当前租户（平台内置数据不可删除）"

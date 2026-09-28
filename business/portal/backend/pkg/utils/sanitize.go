package utils

import (
	"errors"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

// SanitizeError 生成可安全返回给前端的错误信息：
//   - 数据库/驱动错误（SQL 语法、表名/字段名、唯一键冲突等）只写入服务端日志，
//     对外仅返回通用前缀文案，避免泄露 SQL 细节与库表结构；
//   - 业务错误（errors.New 自定义文案、参数校验错误等）保留原文，便于用户理解。
//
// 用法：ctx.JSON(http.StatusOK, utils.Error(1, utils.SanitizeError("创建用户失败", err)))
func SanitizeError(msg string, err error) string {
	if err == nil {
		return msg
	}
	if isDatabaseError(err) {
		if Logger != nil {
			Logger.Errorf("%s: %s", msg, err)
		}
		return msg
	}
	return msg + ": " + err.Error()
}

// SafeErrText 返回可直接展示给前端的错误文本：
//   - 业务错误（errors.New 自定义文案）保留原文；
//   - 数据库/驱动错误只写入服务端日志，对外返回通用兜底文案。
//
// 适用于没有固定前缀、直接透传 err.Error() 的场景。
func SafeErrText(err error) string {
	if err == nil {
		return ""
	}
	if isDatabaseError(err) {
		if Logger != nil {
			Logger.Errorf("数据库操作失败: %s", err)
		}
		return "操作失败，请稍后重试"
	}
	return err.Error()
}

// isDatabaseError 判断错误是否来自数据库/驱动，需对外隐藏细节。
func isDatabaseError(err error) bool {
	if errors.Is(err, gorm.ErrRecordNotFound) ||
		errors.Is(err, gorm.ErrDuplicatedKey) ||
		errors.Is(err, gorm.ErrForeignKeyViolated) ||
		errors.Is(err, gorm.ErrInvalidData) ||
		errors.Is(err, gorm.ErrInvalidValue) {
		return true
	}
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return true
	}
	return false
}

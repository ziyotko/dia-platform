package services

import "errors"

var (
	ErrUserNotFound     = errors.New("用户不存在")
	ErrInvalidPassword  = errors.New("密码错误")
	ErrInvalidCaptcha   = errors.New("验证码错误")
	ErrUserDisabled     = errors.New("用户已禁用")
	ErrTokenExpired     = errors.New("token已过期")
	ErrTokenInvalid     = errors.New("token无效")
	ErrTokenBlacklisted = errors.New("token已被拉黑")
)

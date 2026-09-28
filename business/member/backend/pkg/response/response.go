package response

import (
	"net/http"

	"member/pkg/utils"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// 业务码统一口径（与 portal / application 保持一致）：
//
//	0   成功
//	1   业务失败（参数错误 / 无权限 / 记录不存在 / 服务端异常——用文案区分，不靠数字）
//	401 登录态失效
//	429 限流
//
// HTTP 状态码恒为 200，业务码只在响应体里（前端拦截器只对 0 放行、对 401 登出）。
// 新增接口请用 CodeFail，不要再引入 400/403/404/500 这类业务码。
const (
	CodeSuccess         = 0
	CodeFail            = 1
	CodeUnauthorized    = 401
	CodeTooManyRequests = 429
)

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	})
}

func Error(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
	})
}

func ErrorWithData(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

func BadRequest(c *gin.Context, message string) {
	Error(c, CodeFail, message)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, CodeUnauthorized, message)
}

func Forbidden(c *gin.Context, message string) {
	Error(c, CodeFail, message)
}

func NotFound(c *gin.Context, message string) {
	Error(c, CodeFail, message)
}

func ServerError(c *gin.Context, message string) {
	Error(c, CodeFail, message)
}

// TooManyRequests 限流统一出口（业务码 429）。
func TooManyRequests(c *gin.Context, message string) {
	Error(c, CodeTooManyRequests, message)
}

// ServerErrorFrom 处理服务端错误：业务错误（中文提示）原样返回；
// 数据库/IO 等底层错误统一脱敏为「服务器内部错误」并写入服务端日志。
func ServerErrorFrom(c *gin.Context, err error) {
	ServerError(c, utils.SafeErrMessage(err, "服务器内部错误"))
}

// BadRequestFrom 处理参数/业务校验错误：业务错误原样返回，底层错误统一脱敏。
func BadRequestFrom(c *gin.Context, err error) {
	BadRequest(c, utils.SafeErrMessage(err, "操作失败，请稍后重试"))
}

// NotFoundFrom 处理「记录不存在」类错误：业务错误原样返回，底层错误统一脱敏。
func NotFoundFrom(c *gin.Context, err error) {
	NotFound(c, utils.SafeErrMessage(err, "记录不存在"))
}

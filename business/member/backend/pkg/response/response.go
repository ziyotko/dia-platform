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

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
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
	Error(c, 400, message)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, 401, message)
}

func Forbidden(c *gin.Context, message string) {
	Error(c, 403, message)
}

func NotFound(c *gin.Context, message string) {
	Error(c, 404, message)
}

func ServerError(c *gin.Context, message string) {
	Error(c, 500, message)
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

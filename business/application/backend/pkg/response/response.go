package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"application/pkg/utils"
)

const (
	CodeSuccess      = 0
	CodeError        = 500
	CodeBadRequest   = 400
	CodeUnauthorized = 401
	CodeForbidden    = 403
	CodeNotFound     = 404
)

type Result struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func Ok(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Result{Code: CodeSuccess, Message: "success", Data: data})
}

func OkWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Result{Code: CodeSuccess, Message: message, Data: data})
}

// fail 是所有错误响应的唯一出口：统一做文案脱敏。
// 业务错误（中文提示）原样下发；数据库/网络等底层错误（英文或含 SQL 片段）替换为通用文案并记服务端日志。
// 这样控制器里直接写 response.Fail(c, err.Error()) 也不会把表名/列名/SQL 泄露给前端。
func fail(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Result{
		Code:    code,
		Message: utils.SafeMessage(message, "操作失败，请稍后重试"),
		Data:    nil,
	})
}

func Fail(c *gin.Context, message string) {
	fail(c, CodeError, message)
}

func FailWithCode(c *gin.Context, code int, message string) {
	fail(c, code, message)
}

func BadRequest(c *gin.Context, message string) {
	fail(c, CodeBadRequest, message)
}

func Unauthorized(c *gin.Context) {
	c.JSON(http.StatusOK, Result{Code: CodeUnauthorized, Message: "未登录或登录已过期", Data: nil})
}

func Forbidden(c *gin.Context) {
	c.JSON(http.StatusOK, Result{Code: CodeForbidden, Message: "无操作权限", Data: nil})
}

func NotFound(c *gin.Context) {
	c.JSON(http.StatusOK, Result{Code: CodeNotFound, Message: "资源不存在", Data: nil})
}

type PageData struct {
	List  interface{} `json:"list"`
	Total int64       `json:"total"`
}

func Page(c *gin.Context, list interface{}, total int64) {
	c.JSON(http.StatusOK, Result{Code: CodeSuccess, Message: "success", Data: PageData{List: list, Total: total}})
}

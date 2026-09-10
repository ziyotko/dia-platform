package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
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

func Fail(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Result{Code: CodeError, Message: message, Data: nil})
}

func FailWithCode(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Result{Code: code, Message: message, Data: nil})
}

func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Result{Code: CodeBadRequest, Message: message, Data: nil})
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

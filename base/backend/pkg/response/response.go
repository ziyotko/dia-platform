package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Result struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

const (
	CodeSuccess         = 0
	CodeError           = 500
	CodeUnauthorized    = 401
	CodeForbidden       = 403
	CodeNotFound        = 404
	CodeBadRequest      = 400
	CodeTooManyRequests = 429
)

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

func Page(c *gin.Context, list interface{}, total int64) {
	c.JSON(http.StatusOK, Result{
		Code:    CodeSuccess,
		Message: "success",
		Data: map[string]interface{}{
			"list":  list,
			"total": total,
		},
	})
}

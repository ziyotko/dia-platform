package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"server/config"
	"server/models"
	"server/services"
	"server/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var logService = &services.LogService{}

type OperationLogMiddleware struct{}

func NewOperationLogMiddleware() *OperationLogMiddleware {
	return &OperationLogMiddleware{}
}

func getLogType(method, path string) string {
	if strings.Contains(path, "/login") {
		return "LOGIN"
	}
	switch method {
	case "POST":
		return "CREATE"
	case "PUT", "PATCH":
		return "UPDATE"
	case "DELETE":
		return "DELETE"
	case "GET":
		return "READ"
	default:
		return "READ"
	}
}

func getModule(path string) string {
	path = strings.TrimPrefix(path, config.AppConfig.Server.ApiPrefix)
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		return "系统"
	}

	moduleMap := map[string]string{
		"login":         "认证模块",
		"logout":        "认证模块",
		"profile":       "个人信息",
		"users":         "用户管理",
		"roles":         "角色管理",
		"menus":         "菜单管理",
		"articles":      "文章管理",
		"categories":    "分类管理",
		"tags":          "标签管理",
		"ads":           "广告管理",
		"links":         "友链管理",
		"logs":          "日志管理",
		"captcha":       "认证模块",
		"dashboard":     "统计模块",
		"workflows":     "流程管理",
		"templates":     "模板管理",
		"departments":   "部门管理",
		"settings":      "系统设置",
		"pages":         "页面管理",
		"columns":       "栏目管理",
		"login-logs":    "登录日志",
		"static-logs":   "静态日志",
		"static-pages":  "静态页面",
		"organizations": "机构管理",
		"uploads":       "文件上传",
		"site-info":     "站点信息",
	}

	if module, ok := moduleMap[parts[0]]; ok {
		return module
	}
	return parts[0]
}

func getDescription(method, path, module string) string {
	actionMap := map[string]string{
		"POST":   "新增",
		"PUT":    "修改",
		"PATCH":  "修改",
		"DELETE": "删除",
		"GET":    "查询",
	}
	action := actionMap[method]
	if action == "" {
		action = "操作"
	}
	if strings.Contains(path, "/login") {
		return "用户登录"
	}
	if strings.Contains(path, "/logout") {
		return "用户登出"
	}
	if strings.Contains(path, "/profile") {
		return "查询个人资料"
	}
	if strings.Contains(path, "/captcha") {
		return "获取验证码"
	}
	return action + module + "数据"
}

type customResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *customResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *customResponseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func OperationLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		path := c.Request.URL.Path
		method := c.Request.Method

		var params string
		if method == "POST" || method == "PUT" || method == "PATCH" {
			contentType := c.Request.Header.Get("Content-Type")
			if strings.Contains(contentType, "multipart/form-data") {
				params = "[文件上传]"
			} else {
				bodyBytes, _ := c.GetRawData()
				if len(bodyBytes) > 0 {
					var prettyJSON bytes.Buffer
					if err := json.Indent(&prettyJSON, bodyBytes, "", "  "); err == nil {
						params = prettyJSON.String()
					} else {
						params = string(bodyBytes)
					}
					c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				}
			}
		} else {
			query := c.Request.URL.RawQuery
			if query != "" {
				params = query
			} else {
				params = "{}"
			}
		}

		blw := &customResponseWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		endTime := time.Now()
		duration := endTime.Sub(startTime).Milliseconds()
		statusCode := c.Writer.Status()

		if strings.HasPrefix(path, "/swagger") || path == "/favicon.ico" {
			return
		}

		userID, exists := c.Get("userID")
		if !exists {
			userID = uint(0)
		}

		username := ""
		if uid, ok := userID.(uint); ok && uid > 0 {
			var user models.User
			if err := utils.DB.First(&user, uid).Error; err == nil {
				if user.Username != "" {
					username = user.Username
				} else if user.Account != "" {
					username = user.Account
				} else {
					username = user.Email
				}
			}
		} else {
			if strings.Contains(path, "/login") {
				username = "unknown"
			}
		}

		logType := getLogType(method, path)
		module := getModule(path)
		description := getDescription(method, path, module)

		if len(params) > 2000 {
			params = params[:2000] + "..."
		}

		log := &models.OperationLog{
			UserID:      userID.(uint),
			Username:    username,
			Type:        logType,
			Module:      module,
			Description: description,
			Method:      method,
			Path:        path,
			Params:      params,
			IP:          c.ClientIP(),
			UserAgent:   c.Request.UserAgent(),
			Duration:    duration,
			StatusCode:  statusCode,
		}

		go func(l *models.OperationLog) {
			if err := logService.CreateLog(l); err != nil {
				utils.Logger.Errorf("记录操作日志失败: %v", err)
			}
		}(log)
	}
}

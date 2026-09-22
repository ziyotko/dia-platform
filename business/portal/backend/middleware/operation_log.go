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

// operationLogParamsMaxRunes 操作日志请求体最大保留字符数（按 rune 计）
const operationLogParamsMaxRunes = 2000

func getLogType(method, path string) string {
	// 仅精确匹配登录接口，避免 /login-logs 等路径被误判为登录操作
	if strings.Trim(strings.TrimPrefix(path, config.AppConfig.Server.ApiPrefix), "/") == "login" {
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
		"login":          "认证模块",
		"logout":         "认证模块",
		"profile":        "个人信息",
		"users":          "用户管理",
		"roles":          "角色管理",
		"menus":          "菜单管理",
		"articles":       "文章管理",
		"categories":     "分类管理",
		"tags":           "标签管理",
		"ads":            "广告管理",
		"links":          "友链管理",
		"logs":           "日志管理",
		"captcha":        "认证模块",
		"dashboard":      "统计模块",
		"workflows":      "流程管理",
		"templates":      "模板管理",
		"departments":    "部门管理",
		"settings":       "系统设置",
		"pages":          "页面管理",
		"columns":        "栏目管理",
		"login-logs":     "登录日志",
		"static-logs":    "静态日志",
		"static-pages":   "静态查询",
		"static":         "静态管理",
		"organizations":  "机构管理",
		"uploads":        "文件上传",
		"site-info":      "站点信息",
		"workflow-roles": "流程角色",
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
	// 与 getLogType 一致：先去掉全局 API 前缀再比较。
	// 原先用 strings.Contains 直接匹配带前缀的完整路径，导致 DELETE /login-logs 命中 /login 被记成「用户登录」、
	// PUT /profile/password 命中 /profile 被记成「查询个人资料」，审计语义错误。
	p := "/" + strings.Trim(strings.TrimPrefix(path, config.AppConfig.Server.ApiPrefix), "/")
	switch p {
	case "/login":
		return "用户登录"
	case "/logout":
		return "用户登出"
	case "/captcha":
		return "获取验证码"
	case "/profile/password":
		return "修改个人密码"
	case "/profile":
		return action + "个人资料"
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

// sensitiveParamKeys 请求体中需要脱敏的字段名（小写匹配），避免密码/令牌/联系方式被明文写入操作日志。
// email/phone/mobile 属个人信息（日志保留半年且管理端可查），一并脱敏。
var sensitiveParamKeys = map[string]bool{
	"password":      true,
	"oldpassword":   true,
	"newpassword":   true,
	"emailpassword": true,
	"token":         true,
	"signkey":       true,
	"email":         true,
	"phone":         true,
	"mobile":        true,
}

// formatParamsForLog 将请求体格式化为便于阅读的 JSON 字符串，并对敏感字段脱敏。
func formatParamsForLog(body []byte) string {
	var data any
	if err := json.Unmarshal(body, &data); err == nil {
		maskSensitiveFields(data)
		if b, err := json.MarshalIndent(data, "", "  "); err == nil {
			return string(b)
		}
	}
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
		return prettyJSON.String()
	}
	return string(body)
}

// maskSensitiveFields 递归将 map/数组中敏感字段的值替换为 ******。
func maskSensitiveFields(v any) {
	switch t := v.(type) {
	case map[string]any:
		for k, val := range t {
			if sensitiveParamKeys[strings.ToLower(k)] {
				t[k] = "******"
				continue
			}
			maskSensitiveFields(val)
		}
	case []any:
		for _, item := range t {
			maskSensitiveFields(item)
		}
	}
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
					params = formatParamsForLog(bodyBytes)
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

		if runes := []rune(params); len(runes) > operationLogParamsMaxRunes {
			// 必须按字符（rune）截断：params 是含中文的 JSON，按字节切会把一个汉字切成半个 UTF-8 序列，
			// 在 utf8mb4 严格模式下整条日志写入失败（1366）并被静默丢弃。
			params = string(runes[:operationLogParamsMaxRunes]) + "..."
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

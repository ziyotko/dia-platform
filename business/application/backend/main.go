package main

import (
	"runtime/debug"
	"strconv"
	"time"

	// 内嵌时区数据库：DSN 的 loc=Asia/Shanghai 在没装 tzdata 的精简镜像上也能解析
	_ "time/tzdata"

	"application/config"
	"application/internal/middleware"
	"application/internal/models"
	"application/internal/routes"
	"application/internal/seed"
	"application/internal/service"
	"application/pkg/captcha"
	"application/pkg/db"
	"application/pkg/redis"
	"application/pkg/response"
	"application/pkg/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load config
	config.Load("config.yaml")

	// 2. Init logger
	utils.InitLogger()

	// 3. Init DB
	db.Init(&config.Cfg.MySQL)

	// 4. Init Redis
	redis.Init(&config.Cfg.Redis)

	// 5. Init captcha store
	captcha.Init()

	// 6. Init IP limiter
	middleware.InitIPLimiter(config.Cfg.Server.MaxConcurrentIPs)

	// 7. AutoMigrate
	db.DB.AutoMigrate(
		&models.User{},
		&models.Admin{},
		&models.Role{},
		&models.Expert{},
		&models.ProjectCategory{},
		&models.ProjectBatch{},
		&models.Application{},
		&models.ApplicationMaterial{},
		&models.ReviewAssignment{},
		&models.Certificate{},
		&models.Announcement{},
		&models.Notification{},
		&models.AuditLog{},
		&seed.SystemConfig{},
	)

	// 8. Seed default data
	seed.Run()

	// 9. Setup Gin
	// gin.SetMode 遇到非法值会 panic，因此统一走 normalizeGinMode（空/非法回退 release）
	gin.SetMode(normalizeGinMode(config.Cfg.Server.Mode))
	r := gin.New()
	// 中间件顺序与 portal / member 对齐：CORS → 安全头 → 请求体上限 → 访问日志 → IP 并发限制 → panic 兜底
	r.Use(middleware.CORS())
	r.Use(middleware.SecurityHeaders())
	// 限制非 multipart 请求体大小（JSON 体可能被日志类中间件整体读入内存）
	r.Use(middleware.BodyLimitMiddleware(config.Cfg.Server.MaxJSONBodyMB))
	r.Use(middleware.Logger())
	r.Use(middleware.IPLimit())
	// panic 兜底：裸 gin.Recovery() 会返回 HTTP 500 空 body，破坏「HTTP 200 + 业务码」约定
	r.Use(gin.CustomRecovery(func(c *gin.Context, err any) {
		utils.LogError("服务内部异常: %v\n%s", err, debug.Stack())
		response.Fail(c, "服务内部异常，请稍后重试")
	}))
	r.MaxMultipartMemory = 64 << 20 // 64MB
	// 上传目录不再静态托管（原先 r.Static(upload_dir_prefix+"/uploads", "./uploads") 等于"知道 URL 就能下载"）：
	// 现在只能通过带鉴权的 GET /member/files、GET /admin/files 读取。
	// 若 Nginx 里还配了 /uploads/ 直接指向磁盘，需一并删除（见 DEPLOY.md 四、Nginx 示例）。
	// 限定可信代理，保证 c.ClientIP() 取到的是真实客户端 IP（限流按真实 IP 计数）。
	// SetTrustedProxies 会校验条目（IP 或 CIDR），写错时回退为「不信任任何代理」，
	// 否则 gin 会 panic；此时 ClientIP() 取到的是代理 IP，所有人都共用一个限流桶。
	if err := r.SetTrustedProxies(config.Cfg.Server.TrustedProxies); err != nil {
		utils.Logger.Warn("server.trusted_proxies 配置无效，已回退为不信任任何代理（限流将按代理 IP 计数）：" + err.Error())
		_ = r.SetTrustedProxies(nil)
	} else if gin.Mode() == gin.ReleaseMode && len(config.Cfg.Server.TrustedProxies) == 0 {
		utils.Logger.Warn("server.trusted_proxies 为空：经反向代理部署时限流/审计拿到的是代理 IP，请填写 Nginx 地址")
	}

	// 10. Register routes
	routes.Register(r)

	// 11. Background job: close batches whose application window has expired
	go func() {
		batchSvc := service.BatchService{}
		batchSvc.AutoCloseExpired()
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			batchSvc.AutoCloseExpired()
		}
	}()

	// 12. Start server
	utils.Logger.Info("Application server starting on port " + strconv.Itoa(config.Cfg.Server.Port))
	r.Run("0.0.0.0:" + strconv.Itoa(config.Cfg.Server.Port))
}

// normalizeGinMode 校验 gin 运行模式（仅支持 debug/release/test）：
// 空值或非法值一律回退 release。（gin.SetMode 遇到非法值会 panic，配错即启动失败。）
func normalizeGinMode(mode string) string {
	switch mode {
	case gin.DebugMode, gin.ReleaseMode, gin.TestMode:
		return mode
	}
	if mode != "" {
		utils.Logger.Warnf("server.mode=%q 无效（仅支持 debug/release/test），已回退为 release", mode)
	}
	return gin.ReleaseMode
}

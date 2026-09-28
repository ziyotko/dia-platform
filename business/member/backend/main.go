package main

import (
	"runtime/debug"
	"strconv"

	// 内嵌时区数据库：DSN 的 loc=Asia/Shanghai 在没装 tzdata 的精简镜像上也能解析
	_ "time/tzdata"

	"member/config"
	"member/internal/middleware"
	"member/internal/models"
	"member/internal/routes"
	"member/internal/seed"
	"member/pkg/db"
	"member/pkg/redis"
	"member/pkg/response"
	"member/pkg/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load config
	config.Load("config.yaml")

	// Init logger
	utils.InitLogger()

	// Init database
	db.Init(&config.Cfg.MySQL)

	// Init Redis
	redis.Init(&config.Cfg.Redis)

	// Captcha is initialized on-demand via redis store

	// Init IP limiter
	middleware.InitIPLimiter(config.Cfg.Server.MaxConcurrentIPs)

	// Auto migrate all models
	if err := db.DB.AutoMigrate(
		&models.Member{},
		&models.Application{},
		&models.Certificate{},
		&models.FeeRecord{},
		&models.MemberMessage{},
		&models.Organization{},
		&models.MemberOrganization{},
		&models.Article{},
		&models.ArticleCategory{},
		&models.Announcement{},
		&models.SystemConfig{},
		&models.MemberLevel{},
		&models.MemberOrgLevel{},
		&models.MemberFeeStandard{},
		&models.MemberCertificateTemplate{},
		&models.MemberLevelChange{},
		&models.ProfileChange{},
		&models.OperationLog{},
	); err != nil {
		utils.Logger.Fatalf("AutoMigrate failed: %v", err)
	}

	// 补建 member_fee_records 的 (member_id, year) 唯一索引（幂等：必要时先清理历史重复行，失败不阻断启动）
	db.EnsureUniqueMemberFeeIndex()

	// Seed default data
	seed.Run()

	// Setup Gin
	// gin.SetMode 遇到非法值会 panic，因此统一走 normalizeGinMode（空/非法回退 release）
	gin.SetMode(normalizeGinMode(config.Cfg.Server.Mode))
	r := gin.New()

	// Global middleware（顺序与 portal / application 对齐）
	r.Use(middleware.CORS())
	r.Use(middleware.SecurityHeaders())
	// 限制非 multipart 请求体大小：JSON 体会被操作日志中间件整体读入内存
	r.Use(middleware.BodyLimitMiddleware(config.Cfg.Server.MaxJSONBodyMB))
	r.Use(middleware.Logger())
	r.Use(middleware.IPLimit())
	// panic 兜底：裸 gin.Recovery() 会返回 HTTP 500 空 body，破坏「HTTP 200 + 业务码」约定，
	// 且堆栈只写 stderr、不进 logs/*.log
	r.Use(gin.CustomRecovery(func(c *gin.Context, err any) {
		utils.LogError("服务内部异常: %v\n%s", err, debug.Stack())
		response.Error(c, response.CodeFail, "服务内部异常，请稍后重试")
	}))
	// 可信代理：写错时回退为「不信任任何代理」并告警（否则 gin 取到的是代理 IP，限流会共用一个桶）
	if err := r.SetTrustedProxies(config.Cfg.Server.TrustedProxies); err != nil {
		utils.LogWarn("server.trusted_proxies 配置无效，已回退为不信任任何代理（限流将按代理 IP 计数）：%s", err)
		_ = r.SetTrustedProxies(nil)
	}

	// Serve uploaded files
	r.Static(config.Cfg.Server.UploadDirPrefix+"/uploads", "./uploads")

	// Register routes
	routes.Register(r)

	// Start server
	addr := config.Cfg.Server.Host + ":" + strconv.Itoa(config.Cfg.Server.Port)
	utils.Logger.Infof("Member server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		utils.Logger.Fatalf("Server failed: %v", err)
	}
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

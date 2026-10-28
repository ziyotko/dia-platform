package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	// 内嵌时区数据库：DSN 的 loc=Asia/Shanghai 在没装 tzdata 的精简镜像上也能解析
	_ "time/tzdata"

	"member/config"
	"member/internal/middleware"
	"member/internal/models"
	"member/internal/routes"
	"member/internal/seed"
	"member/internal/service"
	"member/pkg/db"
	"member/pkg/redis"
	"member/pkg/response"
	"member/pkg/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	startedAt := time.Now()

	// Load config
	config.Load("config.yaml")

	// Init logger
	utils.InitLogger()

	// 上线检查（不阻断启动）：CORS 白名单为空或仍是 localhost 时告警
	warnOriginConfig()

	// MySQL 与 Redis 互不依赖：并行初始化，省掉一次串行握手（任一失败仍会 Fatal 退出）
	initDataStores()

	// Captcha is initialized on-demand via redis store

	// Init IP limiter
	middleware.InitIPLimiter(config.Cfg.Server.MaxConcurrentIPs)

	// 结构迁移是启动阶段最重的一步（每张表都要探测 information_schema）。
	// 开关见 config.yaml 的 startup.auto_migrate；环境变量 MEMBER_AUTO_MIGRATE 优先。
	if envBool("MEMBER_AUTO_MIGRATE", config.Cfg.Startup.AutoMigrate) {
		migrateSchema()
	} else {
		utils.Logger.Info("已跳过数据库结构迁移（startup.auto_migrate=false），请确保表结构已由发布流水线维护")
	}

	// Seed default data：幂等的「表为空才写入」，但仍需对每张表做一次 COUNT。
	// 开关见 config.yaml 的 startup.auto_seed；环境变量 MEMBER_AUTO_SEED 优先。
	if envBool("MEMBER_AUTO_SEED", config.Cfg.Startup.AutoSeed) {
		seed.Run()
	} else {
		utils.Logger.Info("已跳过种子数据初始化（startup.auto_seed=false）")
	}

	// Setup Gin
	r := buildRouter()

	// 先绑定端口：监听失败能立刻暴露，且后续后台任务不会推迟端口就绪时间
	addr := config.Cfg.Server.Host + ":" + strconv.Itoa(config.Cfg.Server.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		utils.Logger.Fatalf("监听 %s 失败：%v", addr, err)
	}

	// 会费关闭时幂等回填历史会籍、状态与证书（不改动任何历史费用数据）。
	// 该任务逐条查库并可能生成证书 PDF，放在监听之后异步执行，不阻塞对外服务；
	// 若要求「回填完成」才算启动完成，把 config.yaml 的 startup.sync_migrate 设为 true
	// （或环境变量 MEMBER_SYNC_MIGRATE=true）即可恢复同步执行。
	if envBool("MEMBER_SYNC_MIGRATE", config.Cfg.Startup.SyncMigrate) {
		service.MigrateFeesDisabledData()
	} else {
		go func() {
			defer func() {
				if rec := recover(); rec != nil {
					utils.LogError("会费关闭数据回填 panic：%v\n%s", rec, debug.Stack())
				}
			}()
			service.MigrateFeesDisabledData()
			utils.Logger.Info("会费关闭数据回填完成")
		}()
	}

	srv := &http.Server{
		Handler:           r,
		ReadHeaderTimeout: durationOrDefault(config.Cfg.Server.ReadHeaderTimeout, 15*time.Second),
		IdleTimeout:       durationOrDefault(config.Cfg.Server.IdleTimeout, 120*time.Second),
	}

	// 优雅退出：收到 SIGINT/SIGTERM 后停止接收新连接，给在途请求 10s 收尾
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		utils.Logger.Info("收到退出信号，正在关闭服务…")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			utils.Logger.Warnf("服务关闭未在超时内完成：%v", err)
		}
	}()

	utils.Logger.Infof("Member server ready on %s（启动耗时 %s）", addr, time.Since(startedAt).Round(time.Millisecond))
	if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		utils.Logger.Fatalf("Server failed: %v", err)
	}
}

// initDataStores 并行初始化 MySQL 与 Redis：两者互不依赖，串行执行会白等一次网络往返。
// 两个 Init 内部失败都会直接 Fatal 退出，这里只需等待都完成。
func initDataStores() {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		db.Init(&config.Cfg.MySQL)
	}()
	go func() {
		defer wg.Done()
		redis.Init(&config.Cfg.Redis)
	}()
	wg.Wait()
}

// migrateSchema 建表与补索引。AutoMigrate 会逐表探测 information_schema，是启动耗时的主要来源；
// schema 稳定后可用 MEMBER_AUTO_MIGRATE=false 跳过。
func migrateSchema() {
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
}

// buildRouter 组装 gin 引擎：全局中间件顺序与 portal / application 对齐。
func buildRouter() *gin.Engine {
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
	return r
}

// envBool 读取布尔环境变量；未设置时返回 def，值非法时告警并回退 def。
func envBool(key string, def bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		utils.Logger.Warnf("环境变量 %s=%q 不是合法布尔值（可用 true/false/1/0），已按默认值 %t 处理", key, v, def)
		return def
	}
	return b
}

// durationOrDefault 解析 Go duration 字符串（如 "15s"）；空值/非法值回退 def 并告警。
func durationOrDefault(raw string, def time.Duration) time.Duration {
	v := strings.TrimSpace(raw)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil || d <= 0 {
		utils.Logger.Warnf("时长配置 %q 无效（示例：15s / 2m），已回退默认值 %s", raw, def)
		return def
	}
	return d
}

// warnOriginConfig 上线检查：release 模式下 CORS 白名单为空或仍是本机地址时给出告警。
// 前端跨域失败最常见的两个原因就是这两项没改。
func warnOriginConfig() {
	if config.Cfg.Server.Mode != gin.ReleaseMode {
		return
	}
	origins := config.Cfg.Server.AllowedOrigins
	if len(origins) == 0 {
		utils.LogWarn("server.allowed_origins 为空：浏览器跨域请求会被拒绝，请填写前端实际域名")
		return
	}
	for _, o := range origins {
		if strings.Contains(o, "localhost") || strings.Contains(o, "127.0.0.1") {
			utils.LogWarn("server.allowed_origins 含本机来源 %q：release 模式下请改为前端实际域名", o)
			return
		}
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

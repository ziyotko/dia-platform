package main

import (
	"context"
	"errors"
	"fmt"
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

	"github.com/gin-gonic/gin"

	"portal/config"
	"portal/internal/middleware"
	"portal/internal/models"
	"portal/internal/routes"
	"portal/internal/services"
	"portal/pkg/utils"
)

func main() {
	startedAt := time.Now()

	config.InitConfig()
	utils.InitLogger()

	// MySQL 与 4 个 Redis 库互不依赖：并行建连，省掉多次串行握手（任一失败仍会 Fatal 退出）
	initDataStores()
	logPhase(startedAt, "数据源连接完成")

	// 表结构同步：AutoMigrate 会对每张表探测 information_schema，是启动阶段最重的一步。
	// 开关见 config.yaml 的 startup.auto_migrate；环境变量 PORTAL_AUTO_MIGRATE 优先。
	if envBool("PORTAL_AUTO_MIGRATE", config.AppConfig.Startup.AutoMigrate) {
		migrateStartedAt := time.Now()
		migrateSchema()
		logPhase(migrateStartedAt, "表结构同步完成")
	} else {
		utils.Logger.Info("[startup] 已跳过表结构同步（startup.auto_migrate=false），请确保表结构已由发布流水线维护")
	}

	// 数据初始化/清理任务：均为幂等操作，但含逐条查询与大表全表扫描。
	// 默认放在端口监听之后异步执行，不推迟端口就绪时间；
	// startup.sync_maintenance=true（或环境变量 PORTAL_SYNC_MAINTENANCE=true）时改为同步执行。
	syncMaintenance := envBool("PORTAL_SYNC_MAINTENANCE", config.AppConfig.Startup.SyncMaintenance)
	if syncMaintenance {
		runMaintenance()
	}

	// 生产环境使用 release 模式，避免输出敏感调试信息。
	// gin.SetMode 遇到非法值会 panic，因此统一走 normalizeGinMode（空/非法回退 release）。
	gin.SetMode(normalizeGinMode(config.AppConfig.Server.Mode))

	router := buildRouter()

	addr := fmt.Sprintf("%s:%s", config.AppConfig.Server.Host, config.AppConfig.Server.Port)
	// 先绑定端口：监听失败能立刻暴露，且后台维护任务不会推迟端口就绪时间
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		utils.Logger.Fatalf("监听 %s 失败: %v", addr, err)
	}

	srv := &http.Server{
		Handler:           router,
		ReadHeaderTimeout: durationOrDefault(config.AppConfig.Server.ReadHeaderTimeout, 15*time.Second),
		IdleTimeout:       durationOrDefault(config.AppConfig.Server.IdleTimeout, 120*time.Second),
	}

	// 优雅退出：收到 SIGINT/SIGTERM 后停止接收新连接，给在途请求 10s 收尾
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		utils.Logger.Info("[shutdown] 收到退出信号，正在关闭服务…")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			utils.Logger.Warnf("[shutdown] 服务关闭未在超时内完成: %v", err)
		}
	}()

	utils.Logger.Infof("[startup] 服务就绪 %s，启动耗时 %s", addr, time.Since(startedAt).Round(time.Millisecond))
	println("启动成功，访问地址: " + addr)

	if !syncMaintenance {
		go func() {
			defer func() {
				if rec := recover(); rec != nil {
					utils.Logger.Errorf("[startup] 维护任务异常: %v\n%s", rec, debug.Stack())
				}
			}()
			runMaintenance()
		}()
	}

	if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		utils.Logger.Fatalf("Server failed: %v", err)
	}
}

// initDataStores 并行初始化 MySQL 与各 Redis 库：它们互不依赖，串行执行会白等多次网络往返。
// 各 Init 内部失败会直接退出进程，这里只需等待全部完成。
func initDataStores() {
	tasks := []func(){
		utils.InitDB,
		utils.InitRedisCaptcha,
		utils.InitRedisCache,
		utils.InitRedisAnti,
		// 外部会员（member）令牌黑名单库（只读）：用于「会员登出后 portal 侧立即失效」
		utils.InitRedisMember,
	}
	var wg sync.WaitGroup
	wg.Add(len(tasks))
	for _, task := range tasks {
		go func(fn func()) {
			defer wg.Done()
			fn()
		}(task)
	}
	wg.Wait()
}

// migrateSchema 建表/补列：单表迁移失败只记录日志，不阻断启动，
// 否则后续代码只会以「未知列」的形式报错，排查时找不到根因。
func migrateSchema() {
	for _, m := range models.AllModels() {
		if err := utils.DB.AutoMigrate(m); err != nil {
			utils.Logger.Errorf("[migrate] 表结构同步失败 %T: %s", m, err)
		}
	}
	// FULLTEXT 索引不在 AutoMigrate 能力范围内（文章搜索的 MATCH ... AGAINST 依赖它），启动时幂等补齐
	models.EnsureFulltextIndexes()
}

// runMaintenance 执行幂等的启动期数据任务（历史软删除清理 / 内置数据初始化 / 静态化参数预热）。
// 默认在端口监听后异步调用；它们失败都不影响服务对外提供（各自只记日志）。
func runMaintenance() {
	// 统一硬删：模型已不再内嵌 gorm.DeletedAt（查询不过滤 deleted_at），
	// 先把旧库遗留的软删除数据物理清掉，否则它们会重新"出现"在列表里。
	if envBool("PORTAL_PURGE_SOFT_DELETED", config.AppConfig.Startup.PurgeLegacySoftDeleted) {
		purgeStartedAt := time.Now()
		if purged := models.PurgeLegacySoftDeletedRows(); purged > 0 {
			utils.Logger.Infof("[migrate] 共物理删除 %d 条历史软删除记录", purged)
		}
		logPhase(purgeStartedAt, "历史软删除清理完成")
	}

	// 初始化内置角色/用户/菜单/权限（缺失才创建；顺序有依赖：权限依赖角色与菜单）
	if envBool("PORTAL_AUTO_SEED", config.AppConfig.Startup.AutoSeed) {
		seedStartedAt := time.Now()
		models.SeedDefaultRoles()
		models.SeedDefaultUsers()
		models.SeedDefaultMenus()
		models.SeedDefaultRolePermissions()
		logPhase(seedStartedAt, "内置数据初始化完成")
	} else {
		utils.Logger.Info("[startup] 已跳过内置数据初始化（startup.auto_seed=false）")
	}

	// 启动时加载静态化参数（静态化输出路径/访问地址/访问令牌名/首页整体变灰）并写入缓存；
	// 缓存未命中时读取逻辑会自动回源 DB，因此无需阻塞启动。
	settingsService := &services.SettingsService{}
	if err := settingsService.LoadStaticParamsToCache(); err != nil {
		utils.Logger.Warnf("加载静态化参数到缓存失败: %s", err)
	} else {
		utils.Logger.Infof("静态化参数已加载到缓存 (key: %s)", services.StaticParamsCacheKey)
	}
}

// buildRouter 组装 gin 引擎与全部路由。
func buildRouter() *gin.Engine {
	router := gin.New()
	// panic 兜底：gin.Recovery() 会返回 HTTP 500 空响应体，违反「HTTP 200 + 业务码」约定
	// （前端只能显示「请求失败(HTTP 500)」），且堆栈只写 stderr、不进 logs/*.log。
	// 这里改为记录到日志文件并返回统一业务错误。
	router.Use(middleware.GinLogger(), gin.CustomRecovery(func(c *gin.Context, err any) {
		utils.Logger.Errorf("服务内部异常: %v\n%s", err, debug.Stack())
		c.JSON(http.StatusOK, utils.Error(1, "服务内部异常，请稍后重试"))
	}))

	// 可信反向代理读取自配置；若为空则回退为仅信任本机回环，避免误信任任意代理头
	trustedProxies := config.AppConfig.Server.TrustedProxies
	if len(trustedProxies) == 0 {
		trustedProxies = []string{"127.0.0.1"}
	}
	if err := router.SetTrustedProxies(trustedProxies); err != nil {
		utils.Logger.Warnf("设置可信代理失败，回退为仅信任回环地址: %s", err)
		_ = router.SetTrustedProxies([]string{"127.0.0.1"})
	}

	router.Use(middleware.CorsMiddleware())
	router.Use(middleware.SecurityHeaders())
	// 限制非 multipart 请求体大小：JSON 请求体会被防重放/操作日志中间件整体读入内存，
	// 不设上限时单个超大 body 即可把进程内存打满（文件上传不受影响，另有 50MB/800MB 限制）
	router.Use(middleware.BodyLimitMiddleware(config.AppConfig.Server.MaxJSONBodyMB))

	router.Static(config.AppConfig.Server.UploadDirPrefix+"/uploads", "./uploads")

	routes.SetupRoutes(router)
	return router
}

// logPhase 输出启动阶段耗时（便于定位下一次变慢的环节）。
func logPhase(startedAt time.Time, message string) {
	utils.Logger.Infof("[startup] %s（累计 %s）", message, time.Since(startedAt).Round(time.Millisecond))
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

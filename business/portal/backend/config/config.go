package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config 的段名与 member / application / base 对齐：server / mysql / redis / jwt / log。
// 历史键名（database.*、redis.host|port、redis.db|db1|db2、jwt.expires_hour）仍可用，
// 由 normalizeLegacyConfigKeys 在读取后转成新键名（仅用于老部署过渡）。
type Config struct {
	Server ServerConfig `mapstructure:"server"`
	MySQL  MySQLConfig  `mapstructure:"mysql"`
	Redis  RedisConfig  `mapstructure:"redis"`
	JWT    JWTConfig    `mapstructure:"jwt"`
	Log    LogConfig    `mapstructure:"log"`
}

type ServerConfig struct {
	Port             string `mapstructure:"port"`
	Host             string `mapstructure:"host"`
	Mode             string `mapstructure:"mode"`
	ApiPrefix        string `mapstructure:"api_prefix"`
	UploadDirPrefix  string `mapstructure:"upload_dir_prefix"`
	MaxConcurrentIPs int    `mapstructure:"max_concurrent_ips"`
	// 非 multipart（JSON）请求体上限（MB）：富文本正文会内联 base64 图片，需留足余量；
	// <=0 时回退 64MB。文件上传（multipart）不受此限制。
	MaxJSONBodyMB  int      `mapstructure:"max_json_body_mb"`
	AllowedOrigins []string `mapstructure:"allowed_origins"`

	// 可信反向代理地址，决定 gin 是否信任 X-Forwarded-For / X-Real-IP
	TrustedProxies []string `mapstructure:"trusted_proxies"`

	// 站点分析等公开写接口的限流配置（按真实客户端 IP）
	AnalyticsRateLimit      int `mapstructure:"analytics_rate_limit"`          // 每窗口允许的最大请求数
	AnalyticsRateWindowSecs int `mapstructure:"analytics_rate_window_seconds"` // 限流窗口（秒）

	// 登录接口限流（防暴力破解，按真实客户端 IP）
	LoginRateLimit      int `mapstructure:"login_rate_limit"`          // 每窗口允许的最大请求数
	LoginRateWindowSecs int `mapstructure:"login_rate_window_seconds"` // 限流窗口（秒）

	// 验证码接口限流（防刷验证码，按真实客户端 IP）
	CaptchaRateLimit      int `mapstructure:"captcha_rate_limit"`          // 每窗口允许的最大请求数
	CaptchaRateWindowSecs int `mapstructure:"captcha_rate_window_seconds"` // 限流窗口（秒）

	// 公开只读接口限流（/site-info、/search/articles 等无需认证的查询接口，按真实客户端 IP）
	PublicRateLimit      int `mapstructure:"public_rate_limit"`          // 每窗口允许的最大请求数（<=0 时回退 60）
	PublicRateWindowSecs int `mapstructure:"public_rate_window_seconds"` // 限流窗口（秒，<=0 时回退 60）

	// 文件上传限流（POST /upload，按真实客户端 IP）：防止已认证账号循环上传大文件写满磁盘
	UploadRateLimit      int `mapstructure:"upload_rate_limit"`          // 每窗口允许的最大上传次数（<=0 时回退 60）
	UploadRateWindowSecs int `mapstructure:"upload_rate_window_seconds"` // 限流窗口（秒，<=0 时回退 60）

	// 单账号每日上传量上限（MB，按实际上传字节累计）：频率限流挡不住「慢速大量上传」，
	// 单文件上限也挡不住多次上传，故再加一层配额。<=0 时回退 defaultUploadDailyQuotaMB。
	UploadDailyQuotaMB int `mapstructure:"upload_daily_quota_mb"`

	// 防重放/请求签名校验（主动检测与临时封禁）
	ReplayWindowSecs int `mapstructure:"replay_window_seconds"` // 时间戳新鲜度窗口（秒），默认 120
	ReplayMaxFail    int `mapstructure:"replay_max_fail"`       // 触发临时封禁的失败次数
	ReplayBanMinutes int `mapstructure:"replay_ban_minutes"`    // 临时封禁时长（分钟）
}

// MySQLConfig 数据库连接（键名与 member / application / base 的 mysql 段一致）。
// parse_time / loc / timeout / read_timeout / write_timeout / conn_max_* 是 portal 特有的可调项，
// 其它项目使用代码默认值（本机时区 + 无读写超时），这里保留可配置能力。
type MySQLConfig struct {
	Host            string `mapstructure:"host"`
	Port            string `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"db_name"`
	Charset         string `mapstructure:"charset"`
	ParseTime       bool   `mapstructure:"parse_time"`
	Loc             string `mapstructure:"loc"`
	Timeout         string `mapstructure:"timeout"`
	ReadTimeout     string `mapstructure:"read_timeout"`
	WriteTimeout    string `mapstructure:"write_timeout"`
	MaxIdle         int    `mapstructure:"max_idle"`
	MaxOpen         int    `mapstructure:"max_open"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime int    `mapstructure:"conn_max_idle_time"`
}

// RedisConfig Redis 连接：addr（host:port）+ 三个用途库（与 member / application 的命名一致）。
type RedisConfig struct {
	Addr         string `mapstructure:"addr"`
	Password     string `mapstructure:"password"`
	CaptchaDB    int    `mapstructure:"captcha_db"`
	AntiReplayDB int    `mapstructure:"anti_replay_db"`
	CacheDB      int    `mapstructure:"cache_db"`
}

type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
	Issuer      string `mapstructure:"issuer"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Path       string `mapstructure:"path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

var AppConfig Config

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %s", err)
	}

	// 旧键名 → 新键名的兼容转换（必须在 Unmarshal 之前）
	normalizeLegacyConfigKeys()

	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("Unable to decode config: %s", err)
	}

	applyEnvOverrides()
}

// normalizeLegacyConfigKeys 把历史键名转成统一键名（新键名已存在时不覆盖，即新键优先）：
//
//	database.*                 → mysql.*（username→user、dbname→db_name、max_idle_conns→max_idle、max_open_conns→max_open）
//	redis.host + redis.port   → redis.addr
//	redis.db / db1 / db2      → redis.captcha_db / anti_replay_db / cache_db
//	jwt.expires_hour          → jwt.expire_hours
//
// 目的：老部署的 config.yaml 不改也能启动（迁移期只提示一次告警）。
func normalizeLegacyConfigKeys() {
	if !viper.IsSet("mysql") && viper.IsSet("database") {
		for _, key := range viper.AllKeys() {
			if !strings.HasPrefix(key, "database.") {
				continue
			}
			field := strings.TrimPrefix(key, "database.")
			switch field {
			case "username":
				field = "user"
			case "dbname":
				field = "db_name"
			case "max_idle_conns":
				field = "max_idle"
			case "max_open_conns":
				field = "max_open"
			}
			viper.Set("mysql."+field, viper.Get(key))
		}
		log.Println("[WARN] config.yaml 仍在使用旧键名 database.*，请迁移为 mysql.*（user / db_name / max_idle / max_open）")
	}

	if !viper.IsSet("redis.addr") && viper.IsSet("redis.host") {
		viper.Set("redis.addr", fmt.Sprintf("%s:%s", viper.GetString("redis.host"), viper.GetString("redis.port")))
	}
	if !viper.IsSet("redis.captcha_db") && viper.IsSet("redis.db") {
		viper.Set("redis.captcha_db", viper.GetInt("redis.db"))
	}
	if !viper.IsSet("redis.anti_replay_db") && viper.IsSet("redis.db1") {
		viper.Set("redis.anti_replay_db", viper.GetInt("redis.db1"))
	}
	if !viper.IsSet("redis.cache_db") && viper.IsSet("redis.db2") {
		viper.Set("redis.cache_db", viper.GetInt("redis.db2"))
	}
	if !viper.IsSet("jwt.expire_hours") && viper.IsSet("jwt.expires_hour") {
		viper.Set("jwt.expire_hours", viper.GetInt("jwt.expires_hour"))
	}
}

// applyEnvOverrides 允许通过环境变量覆盖敏感配置，避免把密钥/密码提交到仓库。
// 生产环境务必注入：PORTAL_JWT_SECRET、PORTAL_DB_PASSWORD。
func applyEnvOverrides() {

	if v := os.Getenv("PORTAL_JWT_SECRET"); v != "" {
		AppConfig.JWT.Secret = v
	}

	if v := os.Getenv("PORTAL_DB_PASSWORD"); v != "" {
		AppConfig.MySQL.Password = v
	}

	if v := os.Getenv("PORTAL_MODE"); v != "" {
		AppConfig.Server.Mode = v
	}

	if v := os.Getenv("PORTAL_JWT_SECRET"); v != "" {
		AppConfig.JWT.Secret = v
	}

	if AppConfig.JWT.Secret == "" || AppConfig.JWT.Secret == "PORTAL_JWT_SECRET" {
		log.Printf("[WARN] 未设置 JWT 密钥！必须设置环境变量 PORTAL_JWT_SECRET。")
		log.Fatal("程序退出")
	}

	if AppConfig.MySQL.Password == "" || AppConfig.MySQL.Password == "PORTAL_DB_PASSWORD" {
		log.Printf("[WARN] 未设置数据库密码！必须设置环境变量 PORTAL_DB_PASSWORD。")
		log.Fatal("程序退出")
	}
}

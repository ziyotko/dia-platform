package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server      ServerConfig      `mapstructure:"server"`
	MySQL       MySQLConfig       `mapstructure:"mysql"`
	Redis       RedisConfig       `mapstructure:"redis"`
	JWT         JWTConfig         `mapstructure:"jwt"`
	Log         LogConfig         `mapstructure:"log"`
	Certificate CertificateConfig `mapstructure:"certificate"`
	Features    FeaturesConfig    `mapstructure:"features"`
	Startup     StartupConfig     `mapstructure:"startup"`
}

type FeaturesConfig struct {
	FeesEnabled bool `mapstructure:"fees_enabled"`
}

// StartupConfig 启动阶段的维护动作开关。
// 默认值与历史行为一致：启动时做结构迁移与种子数据、会费关闭回填异步执行。
// 数据已稳定后关掉 auto_migrate / auto_seed 可显著缩短启动时间（见 DEPLOY.md）。
// 对应环境变量 MEMBER_AUTO_MIGRATE / MEMBER_AUTO_SEED / MEMBER_SYNC_MIGRATE 优先级更高。
type StartupConfig struct {
	AutoMigrate bool `mapstructure:"auto_migrate"`
	AutoSeed    bool `mapstructure:"auto_seed"`
	SyncMigrate bool `mapstructure:"sync_migrate"`
}

// CertificateConfig 会员证书 PDF 生成配置。
type CertificateConfig struct {
	// 中文字体文件（TTF/OTF，不支持 TTC）；留空则按系统常见字体自动查找
	FontPath string `mapstructure:"font_path"`
}

type ServerConfig struct {
	Host             string   `mapstructure:"host"`
	Port             int      `mapstructure:"port"`
	Mode             string   `mapstructure:"mode"`
	APIPrefix        string   `mapstructure:"api_prefix"`
	MaxConcurrentIPs int      `mapstructure:"max_concurrent_ips"`
	UploadDirPrefix  string   `mapstructure:"upload_dir_prefix"`
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	TrustedProxies   []string `mapstructure:"trusted_proxies"`

	// HTTP 服务端超时（Go duration 字符串，如 "15s"）；留空或非法值回退代码默认值
	// （ReadHeaderTimeout 15s 防 Slowloris，IdleTimeout 120s 控制 keep-alive 空闲连接）。
	ReadHeaderTimeout string `mapstructure:"read_header_timeout"`
	IdleTimeout       string `mapstructure:"idle_timeout"`

	// 非 multipart（JSON）请求体上限（MB）：<=0 回退 64。文件上传（multipart）不受此限制。
	MaxJSONBodyMB int `mapstructure:"max_json_body_mb"`

	// 固定窗口限流（次数/分钟，按真实客户端 IP；<=0 回退代码默认值）。
	// 与 portal / application 同名同义：登录、验证码、公开只读、上传。
	LoginRateLimit   int `mapstructure:"login_rate_limit"`
	CaptchaRateLimit int `mapstructure:"captcha_rate_limit"`
	PublicRateLimit  int `mapstructure:"public_rate_limit"`
	UploadRateLimit  int `mapstructure:"upload_rate_limit"`
}

type MySQLConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"db_name"`
	Charset  string `mapstructure:"charset"`
	MaxOpen  int    `mapstructure:"max_open"`
	MaxIdle  int    `mapstructure:"max_idle"`

	// 时区与超时（与 portal 的 mysql 段同名同义）：
	// loc 决定时间按哪个时区写入/读出（留空用 Local，此时部署机时区必须是 Asia/Shanghai）；
	// timeout/read_timeout/write_timeout 为 go-sql-driver 的参数，留空则不附加（用驱动默认值）。
	Loc          string `mapstructure:"loc"`
	Timeout      string `mapstructure:"timeout"`
	ReadTimeout  string `mapstructure:"read_timeout"`
	WriteTimeout string `mapstructure:"write_timeout"`
}

type RedisConfig struct {
	Addr         string `mapstructure:"addr"`
	Password     string `mapstructure:"password"`
	CaptchaDB    int    `mapstructure:"captcha_db"`
	AntiReplayDB int    `mapstructure:"anti_replay_db"`

	// 连接 / 读写超时与连接池；留空或 <=0 时用 go-redis 默认值（5s / 3s / 3s / 10×CPU）。
	// Redis 不可达时，DialTimeout 直接决定请求被拖住多久，跨机房部署建议调小。
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	PoolSize     int           `mapstructure:"pool_size"`
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

var Cfg *Config

func Load(path string) {
	// 向后兼容旧部署：配置项缺失时保持原有会费功能开启。
	viper.SetDefault("features.fees_enabled", true)
	// 缺省值与历史行为一致（迁移/种子执行、回填异步），避免旧配置直接跳过建表。
	viper.SetDefault("startup.auto_migrate", true)
	viper.SetDefault("startup.auto_seed", true)
	viper.SetDefault("startup.sync_migrate", false)
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}
	Cfg = &Config{}
	if err := viper.Unmarshal(Cfg); err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}
	applyEnvOverrides()
}

// applyEnvOverrides 允许通过环境变量覆盖敏感配置，避免把密钥/密码提交到仓库。
// 生产环境务必通过环境变量注入：MEMBER_JWT_SECRET、MEMBER_DB_PASSWORD、MEMBER_MODE。
func applyEnvOverrides() {
	if v := os.Getenv("MEMBER_JWT_SECRET"); v != "" {
		Cfg.JWT.Secret = v
	}

	if v := os.Getenv("MEMBER_DB_PASSWORD"); v != "" {
		Cfg.MySQL.Password = v
	}
	// 连接地址同样允许环境变量覆盖，避免同一份配置在不同环境（开发/预发/生产）分叉。
	if v := strings.TrimSpace(os.Getenv("MEMBER_DB_HOST")); v != "" {
		Cfg.MySQL.Host = v
	}
	if v := strings.TrimSpace(os.Getenv("MEMBER_DB_PORT")); v != "" {
		if port, err := strconv.Atoi(v); err == nil && port > 0 && port < 65536 {
			Cfg.MySQL.Port = port
		} else {
			log.Printf("[WARN] MEMBER_DB_PORT=%q 不是合法端口，继续使用 config.yaml 的 mysql.port=%d", v, Cfg.MySQL.Port)
		}
	}
	if v := strings.TrimSpace(os.Getenv("MEMBER_DB_USER")); v != "" {
		Cfg.MySQL.User = v
	}
	if v := strings.TrimSpace(os.Getenv("MEMBER_DB_NAME")); v != "" {
		Cfg.MySQL.DBName = v
	}
	if v := os.Getenv("MEMBER_MODE"); v != "" {
		Cfg.Server.Mode = v
	}
	if v := os.Getenv("MEMBER_CERT_FONT"); v != "" {
		Cfg.Certificate.FontPath = v
	}

	if Cfg.Server.Mode == "" {
		Cfg.Server.Mode = "release"
	}

	if Cfg.JWT.Secret == "" || Cfg.JWT.Secret == "MEMBER_JWT_SECRET" {
		log.Printf("[WARN] 未设置 JWT 密钥！必须设置环境变量 MEMBER_JWT_SECRET。")
		log.Fatal("程序退出")
	}

	if Cfg.MySQL.Password == "" || Cfg.MySQL.Password == "MEMBER_DB_PASSWORD" {
		log.Printf("[WARN] 未设置数据库密码！必须设置环境变量 MEMBER_DB_PASSWORD。")
		log.Fatal("程序退出")
	}
}

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config 应用程序配置
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Redis    RedisConfig    `json:"redis"`
	MongoDB  MongoConfig    `json:"mongodb"`
	JWT      JWTConfig      `json:"jwt"`
	Log      LogConfig      `json:"log"`
}

// ServerConfig HTTP服务器配置
type ServerConfig struct {
	Address      string `json:"address"`
	Port         int    `json:"port"`
	Mode         string `json:"mode"` // debug, release, test
	ReadTimeout  int    `json:"read_timeout"`
	WriteTimeout int    `json:"write_timeout"`
}

// DatabaseConfig MySQL数据库配置
type DatabaseConfig struct {
	Host            string `json:"host"`
	Port            int    `json:"port"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	Database        string `json:"database"`
	Charset         string `json:"charset"`
	MaxIdleConns    int    `json:"max_idle_conns"`
	MaxOpenConns    int    `json:"max_open_conns"`
	ConnMaxLifetime int    `json:"conn_max_lifetime"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	Database int    `json:"database"`
	PoolSize int    `json:"pool_size"`
}

// MongoConfig MongoDB配置
type MongoConfig struct {
	URI      string `json:"uri"`
	Database string `json:"database"`
	Timeout  int    `json:"timeout"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	SecretKey       string `json:"secret_key"`
	ExpirationHours int    `json:"expiration_hours"` // 小时
	Issuer          string `json:"issuer"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level         string `json:"level"`
	Format        string `json:"format"` // json, text
	Output        string `json:"output"` // stdout, file, both
	Filename      string `json:"filename"`
	MaxSize       int    `json:"max_size"`        // 单个日志文件最大大小(MB)
	MaxAge        int    `json:"max_age"`         // 日志文件保留天数
	MaxBackups    int    `json:"max_backups"`     // 保留的旧日志文件数量
	Compress      bool   `json:"compress"`        // 是否压缩旧日志文件
	LocalTime     bool   `json:"local_time"`      // 是否使用本地时间
	RotateDaily   bool   `json:"rotate_daily"`    // 是否按天轮转
	EnableConsole bool   `json:"enable_console"`  // 是否同时输出到控制台
	EnableFile    bool   `json:"enable_file"`     // 是否输出到文件
	LogDir        string `json:"log_dir"`         // 日志目录
	AccessLogFile string `json:"access_log_file"` // 访问日志文件名
	ErrorLogFile  string `json:"error_log_file"`  // 错误日志文件名
	CronLogFile   string `json:"cron_log_file"`   // Cron服务日志文件名
}

// Load 加载配置
func Load() (*Config, error) {
	cfg := &Config{}

	// 设置默认配置
	setDefaults(cfg)

	// 从配置文件加载
	if err := loadFromFile(cfg); err != nil {
		// 配置文件不存在时使用默认配置，但记录警告
		fmt.Printf("Warning: Failed to load config file, using defaults: %v\n", err)
	}

	// 从环境变量覆盖配置
	loadFromEnv(cfg)

	// 验证配置
	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// setDefaults 设置默认配置
func setDefaults(cfg *Config) {
	cfg.Server = ServerConfig{
		Address:      ":8080",
		Port:         8080,
		Mode:         "debug",
		ReadTimeout:  60,
		WriteTimeout: 60,
	}

	cfg.Database = DatabaseConfig{
		Host:            "localhost",
		Port:            3306,
		Username:        "root",
		Password:        "",
		Database:        "go_platform",
		Charset:         "utf8mb4",
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: 3600,
	}

	cfg.Redis = RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		Database: 0,
		PoolSize: 10,
	}

	cfg.MongoDB = MongoConfig{
		URI:      "mongodb://localhost:27017",
		Database: "go_platform",
		Timeout:  10,
	}

	cfg.JWT = JWTConfig{
		SecretKey:       "your-secret-key-change-in-production",
		ExpirationHours: 24,
		Issuer:          "go-api-admin-im-platform",
	}

	cfg.Log = LogConfig{
		Level:         "info",
		Format:        "json",
		Output:        "both",
		Filename:      "app.log",
		MaxSize:       100,
		MaxAge:        30,
		MaxBackups:    10,
		Compress:      true,
		LocalTime:     true,
		RotateDaily:   true,
		EnableConsole: true,
		EnableFile:    true,
		LogDir:        "logs",
		AccessLogFile: "access.log",
		ErrorLogFile:  "error.log",
		CronLogFile:   "cron.log",
	}
}

// loadFromFile 从配置文件加载
func loadFromFile(cfg *Config) error {
	env := os.Getenv("GO_ENV")
	if env == "" {
		env = "development"
	}

	filename := fmt.Sprintf("configs/%s.json", env)

	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, cfg)
}

// loadFromEnv 从环境变量加载配置
func loadFromEnv(cfg *Config) {
	// Server配置
	if addr := os.Getenv("SERVER_ADDRESS"); addr != "" {
		cfg.Server.Address = addr
	}
	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Server.Port = p
			cfg.Server.Address = fmt.Sprintf(":%d", p)
		}
	}
	if mode := os.Getenv("GIN_MODE"); mode != "" {
		cfg.Server.Mode = mode
	}

	// Database配置
	if host := os.Getenv("DB_HOST"); host != "" {
		cfg.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Database.Port = p
		}
	}
	if user := os.Getenv("DB_USERNAME"); user != "" {
		cfg.Database.Username = user
	}
	if pass := os.Getenv("DB_PASSWORD"); pass != "" {
		cfg.Database.Password = pass
	}
	if db := os.Getenv("DB_DATABASE"); db != "" {
		cfg.Database.Database = db
	}

	// Redis配置
	if host := os.Getenv("REDIS_HOST"); host != "" {
		cfg.Redis.Host = host
	}
	if port := os.Getenv("REDIS_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Redis.Port = p
		}
	}
	if pass := os.Getenv("REDIS_PASSWORD"); pass != "" {
		cfg.Redis.Password = pass
	}

	// MongoDB配置
	if uri := os.Getenv("MONGO_URI"); uri != "" {
		cfg.MongoDB.URI = uri
	}
	if db := os.Getenv("MONGO_DATABASE"); db != "" {
		cfg.MongoDB.Database = db
	}

	// JWT配置
	if secret := os.Getenv("JWT_SECRET_KEY"); secret != "" {
		cfg.JWT.SecretKey = secret
	}
	if expire := os.Getenv("JWT_EXPIRATION_HOURS"); expire != "" {
		if e, err := strconv.Atoi(expire); err == nil {
			cfg.JWT.ExpirationHours = e
		}
	}
	if issuer := os.Getenv("JWT_ISSUER"); issuer != "" {
		cfg.JWT.Issuer = issuer
	}
}

// validate 验证配置
func validate(cfg *Config) error {
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", cfg.Server.Port)
	}

	if cfg.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if cfg.Database.Username == "" {
		return fmt.Errorf("database username is required")
	}

	if cfg.Database.Database == "" {
		return fmt.Errorf("database name is required")
	}

	if cfg.JWT.SecretKey == "" || cfg.JWT.SecretKey == "your-secret-key-change-in-production" {
		return fmt.Errorf("JWT secret key must be set and not use default value")
	}

	if cfg.JWT.ExpirationHours <= 0 {
		return fmt.Errorf("JWT expiration hours must be positive")
	}

	if !contains([]string{"debug", "release", "test"}, cfg.Server.Mode) {
		return fmt.Errorf("invalid server mode: %s", cfg.Server.Mode)
	}

	if !contains([]string{"debug", "info", "warn", "error"}, cfg.Log.Level) {
		return fmt.Errorf("invalid log level: %s", cfg.Log.Level)
	}

	return nil
}

// contains 检查字符串是否在切片中
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

// GetDSN 获取MySQL数据源名称
func (cfg *Config) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Database,
		cfg.Database.Charset,
	)
}

// GetRedisAddr 获取Redis地址
func (cfg *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
}

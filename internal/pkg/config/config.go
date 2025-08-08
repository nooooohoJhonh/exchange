package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
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
	Mode         string `json:"mode"`
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
	ExpirationHours int    `json:"expiration_hours"`
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
	setDefaults(cfg)

	// 从配置文件加载
	if err := loadFromFile(cfg); err != nil {
		fmt.Printf("警告: 配置文件加载失败，使用默认配置: %v\n", err)
	}

	// 从环境变量覆盖配置
	loadFromEnv(cfg)

	// 验证配置
	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return cfg, nil
}

// setDefaults 设置默认配置
func setDefaults(cfg *Config) {
	// 服务器默认配置
	cfg.Server.Address = "0.0.0.0"
	cfg.Server.Port = 8080
	cfg.Server.Mode = "debug"
	cfg.Server.ReadTimeout = 30
	cfg.Server.WriteTimeout = 30

	// 数据库默认配置
	cfg.Database.Host = "localhost"
	cfg.Database.Port = 3306
	cfg.Database.Username = "root"
	cfg.Database.Password = ""
	cfg.Database.Database = "exchange"
	cfg.Database.Charset = "utf8mb4"
	cfg.Database.MaxIdleConns = 10
	cfg.Database.MaxOpenConns = 100
	cfg.Database.ConnMaxLifetime = 3600

	// Redis默认配置
	cfg.Redis.Host = "localhost"
	cfg.Redis.Port = 6379
	cfg.Redis.Password = ""
	cfg.Redis.Database = 0
	cfg.Redis.PoolSize = 10

	// MongoDB默认配置
	cfg.MongoDB.URI = "mongodb://localhost:27017"
	cfg.MongoDB.Database = "exchange"
	cfg.MongoDB.Timeout = 10

	// JWT默认配置
	cfg.JWT.SecretKey = "your-secret-key"
	cfg.JWT.ExpirationHours = 24
	cfg.JWT.Issuer = "exchange"

	// 日志默认配置
	cfg.Log.Level = "info"
	cfg.Log.Format = "json"
	cfg.Log.Output = "both"
	cfg.Log.Filename = "app.log"
	cfg.Log.MaxSize = 100
	cfg.Log.MaxAge = 30
	cfg.Log.MaxBackups = 10
	cfg.Log.Compress = true
	cfg.Log.LocalTime = true
	cfg.Log.RotateDaily = true
	cfg.Log.EnableConsole = true
	cfg.Log.EnableFile = true
	cfg.Log.LogDir = "logs"
	cfg.Log.AccessLogFile = "access.log"
	cfg.Log.ErrorLogFile = "error.log"
	cfg.Log.CronLogFile = "cron.log"
}

// loadFromFile 从配置文件加载
func loadFromFile(cfg *Config) error {
	configFile := "configs/development.json"
	if os.Getenv("ENV") == "production" {
		configFile = "configs/production.json"
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, cfg)
}

// loadFromEnv 从环境变量加载
func loadFromEnv(cfg *Config) {
	// 服务器配置
	if val := os.Getenv("SERVER_ADDRESS"); val != "" {
		cfg.Server.Address = val
	}
	if val := os.Getenv("SERVER_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			cfg.Server.Port = port
		}
	}
	if val := os.Getenv("SERVER_MODE"); val != "" {
		cfg.Server.Mode = val
	}

	// 数据库配置
	if val := os.Getenv("DB_HOST"); val != "" {
		cfg.Database.Host = val
	}
	if val := os.Getenv("DB_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			cfg.Database.Port = port
		}
	}
	if val := os.Getenv("DB_USERNAME"); val != "" {
		cfg.Database.Username = val
	}
	if val := os.Getenv("DB_PASSWORD"); val != "" {
		cfg.Database.Password = val
	}
	if val := os.Getenv("DB_DATABASE"); val != "" {
		cfg.Database.Database = val
	}

	// Redis配置
	if val := os.Getenv("REDIS_HOST"); val != "" {
		cfg.Redis.Host = val
	}
	if val := os.Getenv("REDIS_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			cfg.Redis.Port = port
		}
	}
	if val := os.Getenv("REDIS_PASSWORD"); val != "" {
		cfg.Redis.Password = val
	}

	// JWT配置
	if val := os.Getenv("JWT_SECRET_KEY"); val != "" {
		cfg.JWT.SecretKey = val
	}
}

// validate 验证配置
func validate(cfg *Config) error {
	// 验证服务器配置
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("无效的服务器端口: %d", cfg.Server.Port)
	}

	// 验证数据库配置
	if cfg.Database.Host == "" {
		return fmt.Errorf("数据库主机不能为空")
	}
	if cfg.Database.Port <= 0 || cfg.Database.Port > 65535 {
		return fmt.Errorf("无效的数据库端口: %d", cfg.Database.Port)
	}
	if cfg.Database.Username == "" {
		return fmt.Errorf("数据库用户名不能为空")
	}
	if cfg.Database.Database == "" {
		return fmt.Errorf("数据库名不能为空")
	}

	// 验证Redis配置
	if cfg.Redis.Host == "" {
		return fmt.Errorf("Redis主机不能为空")
	}
	if cfg.Redis.Port <= 0 || cfg.Redis.Port > 65535 {
		return fmt.Errorf("无效的Redis端口: %d", cfg.Redis.Port)
	}

	// 验证JWT配置
	if cfg.JWT.SecretKey == "" {
		return fmt.Errorf("JWT密钥不能为空")
	}
	if cfg.JWT.ExpirationHours <= 0 {
		return fmt.Errorf("JWT过期时间必须大于0")
	}

	return nil
}

// GetDSN 获取数据库连接字符串
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

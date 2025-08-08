package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"

	"exchange/internal/pkg/config"
)

// Level 日志级别
type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

// String 返回日志级别字符串
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// LogEntry 日志条目
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Service   string                 `json:"service"`
	Context   map[string]interface{} `json:"context,omitempty"`
	File      string                 `json:"file,omitempty"`
	Line      int                    `json:"line,omitempty"`
}

// Logger 日志记录器
type Logger struct {
	level         Level
	format        string
	outputs       []io.Writer
	service       string
	accessLogger  *lumberjack.Logger
	errorLogger   *lumberjack.Logger
	generalLogger *lumberjack.Logger
	cleanupMgr    *LogCleanupManager
	currentDate   string // 当前日期，用于跟踪日期变化
	mu            sync.RWMutex
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// Init 初始化日志系统
func Init(cfg *config.LogConfig) error {
	var err error
	once.Do(func() {
		err = initLogger(cfg)
	})
	return err
}

// initLogger 初始化日志记录器
func initLogger(cfg *config.LogConfig) error {
	level := parseLevel(cfg.Level)

	// 创建日志目录
	if cfg.EnableFile && cfg.LogDir != "" {
		if err := os.MkdirAll(cfg.LogDir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}
	}

	logger := &Logger{
		level:   level,
		format:  cfg.Format,
		service: "exchange",
		outputs: make([]io.Writer, 0),
	}

	// 添加控制台输出
	if cfg.EnableConsole {
		logger.outputs = append(logger.outputs, os.Stdout)
	}

	// 添加文件输出
	if cfg.EnableFile && cfg.LogDir != "" {
		// 生成按天的日志文件名
		today := time.Now().Format("2006-01-02")
		logger.currentDate = today // 初始化当前日期

		// 通用日志文件
		if cfg.Filename != "" {
			generalLogFile := filepath.Join(cfg.LogDir, fmt.Sprintf("%s_%s.log", strings.TrimSuffix(cfg.Filename, ".log"), today))
			logger.generalLogger = &lumberjack.Logger{
				Filename:   generalLogFile,
				MaxSize:    cfg.MaxSize, // MB
				MaxBackups: cfg.MaxBackups,
				MaxAge:     cfg.MaxAge, // days
				Compress:   cfg.Compress,
			}
			logger.outputs = append(logger.outputs, logger.generalLogger)
		}

		// 访问日志文件
		accessLogFile := filepath.Join(cfg.LogDir, fmt.Sprintf("access_%s.log", today))
		logger.accessLogger = &lumberjack.Logger{
			Filename:   accessLogFile,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}

		// 错误日志文件
		errorLogFile := filepath.Join(cfg.LogDir, fmt.Sprintf("error_%s.log", today))
		logger.errorLogger = &lumberjack.Logger{
			Filename:   errorLogFile,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}
	}

	defaultLogger = logger

	// 清理任务已注册到定时任务系统中，每天凌晨2点自动执行
	// 但是仍然需要初始化清理管理器，以便手动清理功能可用
	if cfg.EnableFile && cfg.LogDir != "" {
		logger.cleanupMgr = NewLogCleanupManager(cfg)
	}

	Info("Logger initialized with daily rotation", map[string]interface{}{
		"log_dir": cfg.LogDir,
		"level":   cfg.Level,
		"format":  cfg.Format,
	})

	return nil
}

// parseLevel 解析日志级别
func parseLevel(levelStr string) Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	default:
		return InfoLevel
	}
}

// getDailyLogFile 获取当天的日志文件名
func (l *Logger) getDailyLogFile(baseName string) string {
	today := time.Now().Format("2006-01-02")
	return fmt.Sprintf("%s_%s.log", strings.TrimSuffix(baseName, ".log"), today)
}

// updateLogFiles 更新日志文件为当天的文件
func (l *Logger) updateLogFiles() {
	today := time.Now().Format("2006-01-02")

	// 检查日期是否变化
	if l.currentDate == today {
		return
	}
	l.currentDate = today

	// 更新通用日志文件
	if l.generalLogger != nil {
		newFilename := filepath.Join(filepath.Dir(l.generalLogger.Filename), l.getDailyLogFile(filepath.Base(l.generalLogger.Filename)))
		if l.generalLogger.Filename != newFilename {
			l.generalLogger.Filename = newFilename
		}
	}

	// 更新访问日志文件
	if l.accessLogger != nil {
		newFilename := filepath.Join(filepath.Dir(l.accessLogger.Filename), fmt.Sprintf("access_%s.log", today))
		if l.accessLogger.Filename != newFilename {
			l.accessLogger.Filename = newFilename
		}
	}

	// 更新错误日志文件
	if l.errorLogger != nil {
		newFilename := filepath.Join(filepath.Dir(l.errorLogger.Filename), fmt.Sprintf("error_%s.log", today))
		if l.errorLogger.Filename != newFilename {
			l.errorLogger.Filename = newFilename
		}
	}
}

// log 记录日志
func (l *Logger) log(level Level, message string, context map[string]interface{}) {
	// 检查日志级别
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// 检查是否需要更新日志文件（每天更新一次）
	l.updateLogFiles()

	entry := LogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05.000"),
		Level:     level.String(),
		Message:   message,
		Service:   l.service,
		Context:   context,
	}

	// 添加调用位置信息
	if _, file, line, ok := runtime.Caller(2); ok {
		entry.File = filepath.Base(file)
		entry.Line = line
	}

	var output string
	if l.format == "json" {
		data, _ := json.Marshal(entry)
		output = string(data)
	} else {
		output = l.formatText(entry)
	}

	// 写入到所有输出
	for _, writer := range l.outputs {
		fmt.Fprintln(writer, output)
	}

	// 根据日志级别写入到特定文件
	if level >= ErrorLevel && l.errorLogger != nil {
		fmt.Fprintln(l.errorLogger, output)
	}
}

// logAccess 记录访问日志
func (l *Logger) logAccess(entry LogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 检查是否需要更新日志文件
	l.updateLogFiles()

	var output string
	if l.format == "json" {
		data, _ := json.Marshal(entry)
		output = string(data)
	} else {
		output = l.formatText(entry)
	}

	// 写入访问日志文件
	if l.accessLogger != nil {
		fmt.Fprintln(l.accessLogger, output)
	}

	// 如果启用控制台输出，也写入控制台
	for _, writer := range l.outputs {
		if writer == os.Stdout || writer == os.Stderr {
			fmt.Fprintln(writer, output)
		}
	}
}

// formatText 格式化文本日志
func (l *Logger) formatText(entry LogEntry) string {
	var contextStr string
	if len(entry.Context) > 0 {
		contextData, _ := json.Marshal(entry.Context)
		contextStr = fmt.Sprintf(" context=%s", string(contextData))
	}

	return fmt.Sprintf("[%s] %s %s%s",
		entry.Timestamp,
		entry.Level,
		entry.Message,
		contextStr,
	)
}

// Debug 记录调试日志
func Debug(message string, context ...map[string]interface{}) {
	if defaultLogger == nil {
		log.Printf("[DEBUG] %s", message)
		return
	}

	var ctx map[string]interface{}
	if len(context) > 0 {
		ctx = context[0]
	}

	defaultLogger.log(DebugLevel, message, ctx)
}

// Info 记录信息日志
func Info(message string, context ...map[string]interface{}) {
	if defaultLogger == nil {
		log.Printf("[INFO] %s", message)
		return
	}

	var ctx map[string]interface{}
	if len(context) > 0 {
		ctx = context[0]
	}

	defaultLogger.log(InfoLevel, message, ctx)
}

// Warn 记录警告日志
func Warn(message string, context ...map[string]interface{}) {
	if defaultLogger == nil {
		log.Printf("[WARN] %s", message)
		return
	}

	var ctx map[string]interface{}
	if len(context) > 0 {
		ctx = context[0]
	}

	defaultLogger.log(WarnLevel, message, ctx)
}

// Error 记录错误日志
func Error(message string, context ...map[string]interface{}) {
	if defaultLogger == nil {
		log.Printf("[ERROR] %s", message)
		return
	}

	var ctx map[string]interface{}
	if len(context) > 0 {
		ctx = context[0]
	}

	defaultLogger.log(ErrorLevel, message, ctx)
}

// Access 记录访问日志
func Access(message string, context map[string]interface{}) {
	if defaultLogger == nil {
		log.Printf("[ACCESS] %s", message)
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     "ACCESS",
		Message:   message,
		Service:   defaultLogger.service,
		Context:   context,
	}

	defaultLogger.logAccess(entry)
}

// Performance 记录性能日志
func Performance(message string, context map[string]interface{}) {
	if defaultLogger == nil {
		log.Printf("[PERF] %s", message)
		return
	}

	if context == nil {
		context = make(map[string]interface{})
	}
	context["type"] = "performance"

	defaultLogger.log(InfoLevel, message, context)
}

// Security 记录安全日志
func Security(message string, context map[string]interface{}) {
	if defaultLogger == nil {
		log.Printf("[SECURITY] %s", message)
		return
	}

	if context == nil {
		context = make(map[string]interface{})
	}
	context["type"] = "security"

	defaultLogger.log(WarnLevel, message, context)
}

// Audit 记录审计日志
func Audit(message string, context map[string]interface{}) {
	if defaultLogger == nil {
		log.Printf("[AUDIT] %s", message)
		return
	}

	if context == nil {
		context = make(map[string]interface{})
	}
	context["type"] = "audit"

	defaultLogger.log(InfoLevel, message, context)
}

// Fatal 记录致命错误日志并退出程序
func Fatal(message string, context ...map[string]interface{}) {
	var ctx map[string]interface{}
	if len(context) > 0 {
		ctx = context[0]
	}

	if defaultLogger != nil {
		defaultLogger.log(ErrorLevel, message, ctx)
	} else {
		log.Printf("[FATAL] %s", message)
	}

	os.Exit(1)
}

// Rotate 手动轮转日志文件
func Rotate() error {
	if defaultLogger == nil {
		return fmt.Errorf("logger not initialized")
	}

	defaultLogger.mu.Lock()
	defer defaultLogger.mu.Unlock()

	var errs []error

	if defaultLogger.generalLogger != nil {
		if err := defaultLogger.generalLogger.Rotate(); err != nil {
			errs = append(errs, fmt.Errorf("failed to rotate general log: %w", err))
		}
	}

	if defaultLogger.accessLogger != nil {
		if err := defaultLogger.accessLogger.Rotate(); err != nil {
			errs = append(errs, fmt.Errorf("failed to rotate access log: %w", err))
		}
	}

	if defaultLogger.errorLogger != nil {
		if err := defaultLogger.errorLogger.Rotate(); err != nil {
			errs = append(errs, fmt.Errorf("failed to rotate error log: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("log rotation errors: %v", errs)
	}

	return nil
}

// Flush 刷新日志缓冲区
func Flush() error {
	if defaultLogger == nil {
		return nil
	}

	defaultLogger.mu.Lock()
	defer defaultLogger.mu.Unlock()

	// 对于lumberjack，我们可以通过写入一个空字符串来强制刷新
	// 但这会在日志中留下空行，所以我们不这样做
	// 相反，我们依赖于lumberjack的内部缓冲机制

	return nil
}

// Close 关闭日志记录器
func Close() error {
	if defaultLogger == nil {
		return nil
	}

	defaultLogger.mu.Lock()
	defer defaultLogger.mu.Unlock()

	var errs []error

	if defaultLogger.generalLogger != nil {
		if err := defaultLogger.generalLogger.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close general log: %w", err))
		}
	}

	if defaultLogger.accessLogger != nil {
		if err := defaultLogger.accessLogger.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close access log: %w", err))
		}
	}

	if defaultLogger.errorLogger != nil {
		if err := defaultLogger.errorLogger.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close error log: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("log close errors: %v", errs)
	}

	return nil
}

// GetLogger 获取默认日志记录器
func GetLogger() *Logger {
	return defaultLogger
}

// SetLevel 设置日志级别
func SetLevel(level Level) {
	if defaultLogger != nil {
		defaultLogger.mu.Lock()
		defaultLogger.level = level
		defaultLogger.mu.Unlock()
	}
}

// WithContext 创建带上下文的日志记录器
func WithContext(ctx context.Context) *ContextLogger {
	return &ContextLogger{
		ctx:    ctx,
		logger: defaultLogger,
	}
}

// ContextLogger 带上下文的日志记录器
type ContextLogger struct {
	ctx    context.Context
	logger *Logger
}

// Debug 记录调试日志
func (cl *ContextLogger) Debug(message string, context ...map[string]interface{}) {
	Debug(message, context...)
}

// Info 记录信息日志
func (cl *ContextLogger) Info(message string, context ...map[string]interface{}) {
	Info(message, context...)
}

// Warn 记录警告日志
func (cl *ContextLogger) Warn(message string, context ...map[string]interface{}) {
	Warn(message, context...)
}

// Error 记录错误日志
func (cl *ContextLogger) Error(message string, context ...map[string]interface{}) {
	Error(message, context...)
}

// ForceCleanup 强制清理日志文件
func ForceCleanup(dryRun bool) error {
	if defaultLogger == nil || defaultLogger.cleanupMgr == nil {
		return fmt.Errorf("logger or cleanup manager not initialized")
	}

	return defaultLogger.cleanupMgr.ForceCleanup(dryRun)
}

// GetLogStats 获取日志统计信息
func GetLogStats() (map[string]interface{}, error) {
	if defaultLogger == nil || defaultLogger.cleanupMgr == nil {
		return nil, fmt.Errorf("logger or cleanup manager not initialized")
	}

	return defaultLogger.cleanupMgr.GetLogStats()
}

// StartCleanupScheduler 启动清理调度器（如果还没有启动）
func StartCleanupScheduler() {
	if defaultLogger == nil || defaultLogger.cleanupMgr == nil {
		Error("Cannot start cleanup scheduler: logger not initialized", nil)
		return
	}

	Info("Cleanup scheduler is already running", map[string]interface{}{
		"log_dir": defaultLogger.cleanupMgr.config.LogDir,
	})
}

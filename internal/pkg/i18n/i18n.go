package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"

	appLogger "exchange/internal/pkg/logger"
)

//go:embed locales/*.json
var localeFS embed.FS

// I18nManager 国际化管理器
type I18nManager struct {
	bundle          *i18n.Bundle
	supportedLangs  []language.Tag
	defaultLanguage language.Tag
	mutex           sync.RWMutex
}

// NewI18nManager 创建国际化管理器
func NewI18nManager(defaultLang string) *I18nManager {
	// 解析默认语言
	defaultTag, err := language.Parse(defaultLang)
	if err != nil {
		defaultTag = language.English
	}

	// 创建Bundle
	bundle := i18n.NewBundle(defaultTag)
	bundle.RegisterUnmarshalFunc("json", func(data []byte, v interface{}) error {
		return json.Unmarshal(data, v)
	})

	manager := &I18nManager{
		bundle:          bundle,
		defaultLanguage: defaultTag,
		supportedLangs:  []language.Tag{defaultTag},
	}

	// 加载内嵌的翻译文件
	manager.loadEmbeddedTranslations()

	return manager
}

// loadEmbeddedTranslations 加载内嵌的翻译文件
func (m *I18nManager) loadEmbeddedTranslations() {
	// 支持的语言列表
	languages := []string{"en", "zh"}

	for _, lang := range languages {
		filename := fmt.Sprintf("locales/%s.json", lang)

		// 从内嵌文件系统读取
		data, err := localeFS.ReadFile(filename)
		if err != nil {
			// 如果文件不存在，使用默认翻译
			m.loadDefaultTranslations(lang)
			continue
		}

		// 解析语言标签
		langTag, err := language.Parse(lang)
		if err != nil {
			appLogger.Error("Failed to parse language tag", map[string]interface{}{
				"language": lang,
				"error":    err.Error(),
			})
			continue
		}

		// 加载翻译文件
		_, err = m.bundle.ParseMessageFileBytes(data, filename)
		if err != nil {
			appLogger.Error("Failed to parse translation file", map[string]interface{}{
				"file":  filename,
				"error": err.Error(),
			})
			// 如果解析失败，使用默认翻译
			m.loadDefaultTranslations(lang)
			continue
		}

		// 添加到支持的语言列表
		if !m.containsLanguage(langTag) {
			m.supportedLangs = append(m.supportedLangs, langTag)
		}

		appLogger.Info("Loaded translation file", map[string]interface{}{
			"language": lang,
			"file":     filename,
		})
	}
}

// loadDefaultTranslations 加载默认翻译（当文件不存在时）
func (m *I18nManager) loadDefaultTranslations(lang string) {
	translations := m.getDefaultTranslations(lang)
	if translations == nil {
		return
	}

	// 解析语言标签
	langTag, err := language.Parse(lang)
	if err != nil {
		return
	}

	// 创建消息文件
	for key, value := range translations {
		message := &i18n.Message{
			ID:    key,
			Other: value,
		}
		m.bundle.AddMessages(langTag, message)
	}

	// 添加到支持的语言列表
	if !m.containsLanguage(langTag) {
		m.supportedLangs = append(m.supportedLangs, langTag)
	}

	appLogger.Info("Loaded default translations", map[string]interface{}{
		"language": lang,
		"keys":     len(translations),
	})
}

// getDefaultTranslations 获取默认翻译
func (m *I18nManager) getDefaultTranslations(lang string) map[string]string {
	switch lang {
	case "en":
		return map[string]string{
			// 通用消息
			"success":               "Success",
			"error":                 "Error",
			"invalid_request":       "Invalid request",
			"internal_server_error": "Internal server error",
			"not_found":             "Not found",
			"method_not_allowed":    "Method not allowed",
			"too_many_requests":     "Too many requests",

			// 认证相关
			"unauthorized":             "Unauthorized",
			"forbidden":                "Forbidden",
			"invalid_token":            "Invalid or expired token",
			"token_required":           "Authorization token is required",
			"invalid_credentials":      "Invalid username or password",
			"account_inactive":         "Account is not active",
			"insufficient_permissions": "Insufficient permissions",

			// 验证相关
			"validation_failed":  "Validation failed",
			"required_field":     "This field is required",
			"invalid_email":      "Invalid email format",
			"invalid_password":   "Password does not meet requirements",
			"password_too_short": "Password must be at least 8 characters",
			"password_too_long":  "Password must be less than 128 characters",
			"username_too_short": "Username must be at least 3 characters",
			"username_too_long":  "Username must be less than 50 characters",

			// 用户相关
			"user_not_found":      "User not found",
			"user_already_exists": "User already exists",
			"user_created":        "User created successfully",
			"user_updated":        "User updated successfully",
			"user_deleted":        "User deleted successfully",
			"login_successful":    "Login successful",
			"logout_successful":   "Logout successful",

			// 数据库相关
			"database_error":         "Database error",
			"record_not_found":       "Record not found",
			"duplicate_entry":        "Duplicate entry",
			"foreign_key_constraint": "Foreign key constraint violation",

			// 缓存相关
			"cache_error": "Cache error",
			"cache_miss":  "Cache miss",

			// 文件相关
			"file_upload_failed": "File upload failed",
			"file_not_found":     "File not found",
			"invalid_file_type":  "Invalid file type",
			"file_too_large":     "File too large",

			// 系统相关
			"not_implemented": "Feature not implemented yet",
			"request_timeout": "Request timeout",
		}
	case "zh":
		return map[string]string{
			// 通用消息
			"success":               "成功",
			"error":                 "错误",
			"invalid_request":       "无效请求",
			"internal_server_error": "内部服务器错误",
			"not_found":             "未找到",
			"method_not_allowed":    "方法不允许",
			"too_many_requests":     "请求过于频繁",

			// 认证相关
			"unauthorized":             "未授权",
			"forbidden":                "禁止访问",
			"invalid_token":            "无效或过期的令牌",
			"token_required":           "需要授权令牌",
			"invalid_credentials":      "用户名或密码错误",
			"account_inactive":         "账户未激活",
			"insufficient_permissions": "权限不足",

			// 验证相关
			"validation_failed":  "验证失败",
			"required_field":     "此字段为必填项",
			"invalid_email":      "邮箱格式无效",
			"invalid_password":   "密码不符合要求",
			"password_too_short": "密码至少需要8个字符",
			"password_too_long":  "密码不能超过128个字符",
			"username_too_short": "用户名至少需要3个字符",
			"username_too_long":  "用户名不能超过50个字符",

			// 用户相关
			"user_not_found":      "用户不存在",
			"user_already_exists": "用户已存在",
			"user_created":        "用户创建成功",
			"user_updated":        "用户更新成功",
			"user_deleted":        "用户删除成功",
			"login_successful":    "登录成功",
			"logout_successful":   "退出成功",

			// 数据库相关
			"database_error":         "数据库错误",
			"record_not_found":       "记录不存在",
			"duplicate_entry":        "重复条目",
			"foreign_key_constraint": "外键约束违反",

			// 缓存相关
			"cache_error": "缓存错误",
			"cache_miss":  "缓存未命中",

			// 文件相关
			"file_upload_failed": "文件上传失败",
			"file_not_found":     "文件不存在",
			"invalid_file_type":  "无效的文件类型",
			"file_too_large":     "文件过大",

			// 系统相关
			"not_implemented": "功能尚未实现",
			"request_timeout": "请求超时",
		}
	default:
		return nil
	}
}

// containsLanguage 检查是否包含指定语言
func (m *I18nManager) containsLanguage(lang language.Tag) bool {
	for _, supported := range m.supportedLangs {
		if supported == lang {
			return true
		}
	}
	return false
}

// GetLocalizer 获取本地化器
func (m *I18nManager) GetLocalizer(lang string) *i18n.Localizer {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 解析语言标签
	langTag, err := language.Parse(lang)
	if err != nil {
		langTag = m.defaultLanguage
	}

	// 创建本地化器，支持语言回退
	return i18n.NewLocalizer(m.bundle, langTag.String(), m.defaultLanguage.String())
}

// Translate 翻译文本
func (m *I18nManager) Translate(lang, key string, templateData map[string]interface{}) string {
	localizer := m.GetLocalizer(lang)

	// 创建本地化配置
	config := &i18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: templateData,
	}

	// 执行翻译
	result, err := localizer.Localize(config)
	if err != nil {
		// 如果翻译失败，返回键名
		appLogger.Warn("Translation failed", map[string]interface{}{
			"language": lang,
			"key":      key,
			"error":    err.Error(),
		})
		return key
	}

	return result
}

// GetSupportedLanguages 获取支持的语言列表
func (m *I18nManager) GetSupportedLanguages() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	languages := make([]string, len(m.supportedLangs))
	for i, lang := range m.supportedLangs {
		languages[i] = lang.String()
	}
	return languages
}

// GetDefaultLanguage 获取默认语言
func (m *I18nManager) GetDefaultLanguage() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.defaultLanguage.String()
}

// GetLanguageFromContext 从Gin上下文获取语言
func GetLanguageFromContext(c *gin.Context) string {
	// 1. 从查询参数获取
	if lang := c.Query("lang"); lang != "" {
		return lang
	}

	// 2. 从自定义头获取
	if lang := c.GetHeader("X-Language"); lang != "" {
		return lang
	}

	// 3. 从Accept-Language头获取
	if acceptLang := c.GetHeader("Accept-Language"); acceptLang != "" {
		// 解析Accept-Language头，取第一个语言
		langs := parseAcceptLanguage(acceptLang)
		if len(langs) > 0 {
			return langs[0]
		}
	}

	// 4. 从上下文获取（可能由中间件设置）
	if lang, exists := c.Get("language"); exists {
		if langStr, ok := lang.(string); ok {
			return langStr
		}
	}

	// 5. 返回默认语言（中文）
	return "zh"
}

// parseAcceptLanguage 解析Accept-Language头
func parseAcceptLanguage(acceptLang string) []string {
	var languages []string

	// 简单解析Accept-Language头
	// 格式: en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7
	parts := strings.Split(acceptLang, ",")
	for _, part := range parts {
		// 移除权重信息
		lang := strings.Split(strings.TrimSpace(part), ";")[0]
		// 只取主语言代码
		if idx := strings.Index(lang, "-"); idx > 0 {
			lang = lang[:idx]
		}
		if lang != "" {
			languages = append(languages, lang)
		}
	}

	return languages
}

// 全局i18n管理器实例
var globalI18n *I18nManager
var once sync.Once

// GetGlobalI18n 获取全局i18n管理器
func GetGlobalI18n() *I18nManager {
	once.Do(func() {
		globalI18n = NewI18nManager("zh")
	})
	return globalI18n
}

// GetTranslatedMessage 从上下文中获取翻译后的消息
func GetTranslatedMessage(c *gin.Context, messageKey string, templateData map[string]interface{}) string {
	// 获取语言
	lang := GetLanguageFromContext(c)

	// 获取全局i18n管理器
	manager := GetGlobalI18n()

	// 翻译消息
	return manager.Translate(lang, messageKey, templateData)
}

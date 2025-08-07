package middleware

import (
	"github.com/gin-gonic/gin"

	"exchange/internal/pkg/i18n"
)

// I18nMiddleware 国际化中间件
func I18nMiddleware(i18nManager *i18n.I18nManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取客户端语言偏好
		lang := i18n.GetLanguageFromContext(c)
		
		// 验证语言是否支持
		supportedLangs := i18nManager.GetSupportedLanguages()
		if !containsLang(supportedLangs, lang) {
			lang = i18nManager.GetDefaultLanguage()
		}
		
		// 将语言设置到上下文中
		c.Set("language", lang)
		c.Set("i18n", i18nManager)
		
		// 设置响应头
		c.Header("Content-Language", lang)
		
		c.Next()
	}
}

// GetI18nFromContext 从上下文获取i18n管理器
func GetI18nFromContext(c *gin.Context) *i18n.I18nManager {
	if i18nManager, exists := c.Get("i18n"); exists {
		if manager, ok := i18nManager.(*i18n.I18nManager); ok {
			return manager
		}
	}
	// 如果上下文中没有，返回全局实例
	return i18n.GetGlobalI18n()
}

// GetLanguageFromContext 从上下文获取当前语言
func GetLanguageFromContext(c *gin.Context) string {
	if lang, exists := c.Get("language"); exists {
		if langStr, ok := lang.(string); ok {
			return langStr
		}
	}
	return "en"
}

// Translate 翻译文本（从上下文获取语言）
func Translate(c *gin.Context, key string, templateData map[string]interface{}) string {
	i18nManager := GetI18nFromContext(c)
	lang := GetLanguageFromContext(c)
	return i18nManager.Translate(lang, key, templateData)
}

// containsLang 检查切片是否包含指定元素
func containsLang(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
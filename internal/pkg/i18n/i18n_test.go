package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestI18nManager_NewI18nManager(t *testing.T) {
	manager := NewI18nManager("en")
	
	assert.NotNil(t, manager)
	assert.Equal(t, "en", manager.GetDefaultLanguage())
	
	supportedLangs := manager.GetSupportedLanguages()
	assert.Contains(t, supportedLangs, "en")
	assert.Contains(t, supportedLangs, "zh")
}

func TestI18nManager_Translate(t *testing.T) {
	manager := NewI18nManager("en")
	
	// 测试英文翻译
	result := manager.Translate("en", "success", nil)
	assert.Equal(t, "Success", result)
	
	// 测试中文翻译
	result = manager.Translate("zh", "success", nil)
	assert.Equal(t, "成功", result)
	
	// 测试不存在的键
	result = manager.Translate("en", "nonexistent_key", nil)
	assert.Equal(t, "nonexistent_key", result)
	
	// 测试不支持的语言（应该回退到默认语言）
	result = manager.Translate("fr", "success", nil)
	assert.Equal(t, "Success", result)
}

func TestI18nManager_TranslateWithTemplate(t *testing.T) {
	manager := NewI18nManager("en")
	
	// 由于我们的翻译文件中没有模板变量，这里只是测试接口
	templateData := map[string]interface{}{
		"name": "John",
		"count": 5,
	}
	
	result := manager.Translate("en", "success", templateData)
	assert.Equal(t, "Success", result)
}

func TestI18nManager_GetLocalizer(t *testing.T) {
	manager := NewI18nManager("en")
	
	// 测试获取英文本地化器
	localizer := manager.GetLocalizer("en")
	assert.NotNil(t, localizer)
	
	// 测试获取中文本地化器
	localizer = manager.GetLocalizer("zh")
	assert.NotNil(t, localizer)
	
	// 测试获取不支持语言的本地化器（应该回退到默认语言）
	localizer = manager.GetLocalizer("fr")
	assert.NotNil(t, localizer)
}

func TestParseAcceptLanguage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple language",
			input:    "en",
			expected: []string{"en"},
		},
		{
			name:     "language with region",
			input:    "en-US",
			expected: []string{"en"},
		},
		{
			name:     "multiple languages with quality",
			input:    "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7",
			expected: []string{"en", "en", "zh", "zh"},
		},
		{
			name:     "complex accept language",
			input:    "zh-CN,zh;q=0.9,en;q=0.8",
			expected: []string{"zh", "zh", "en"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAcceptLanguage(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
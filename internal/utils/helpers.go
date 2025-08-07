package utils

import (
	"strings"
	"time"
)

// Contains 检查字符串是否包含子字符串
func Contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// GetCurrentTimestamp 获取当前时间戳
func GetCurrentTimestamp() int64 {
	return time.Now().Unix()
}

// GetCurrentTimestampNano 获取当前纳秒时间戳
func GetCurrentTimestampNano() int64 {
	return time.Now().UnixNano()
}

// FormatTime 格式化时间
func FormatTime(t time.Time, layout string) string {
	return t.Format(layout)
}

// ParseTime 解析时间字符串
func ParseTime(timeStr, layout string) (time.Time, error) {
	return time.Parse(layout, timeStr)
}

// IsEmptyString 检查字符串是否为空
func IsEmptyString(s string) bool {
	return strings.TrimSpace(s) == ""
}

// TruncateString 截断字符串
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength] + "..."
}

// RemoveSpecialChars 移除特殊字符
func RemoveSpecialChars(s string) string {
	var result strings.Builder
	for _, char := range s {
		if (char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == ' ' || char == '-' || char == '_' {
			result.WriteRune(char)
		}
	}
	return result.String()
}

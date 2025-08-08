package utils

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"exchange/internal/pkg/i18n"
)

// 错误码定义
const (
	CodeSuccess       = 100 // 成功
	CodeFailure       = 101 // 失败
	CodeUnauthorized  = 401 // 未授权（token失效）
	CodeForbidden     = 403 // 禁止访问
	CodeInternalError = 500 // 内部错误
)

// APIResponse 统一API响应格式
type APIResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

// getI18nManager 获取国际化管理器
func getI18nManager(c *gin.Context) *i18n.I18nManager {
	if manager, exists := c.Get("i18n"); exists {
		if mgr, ok := manager.(*i18n.I18nManager); ok {
			return mgr
		}
	}
	return i18n.GetGlobalI18n()
}

// getLanguage 获取语言
func getLanguage(c *gin.Context) string {
	return i18n.GetLanguageFromContext(c)
}

// getRequestID 获取请求ID
func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// buildResponse 构建响应
func buildResponse(c *gin.Context, code int, messageKey string, data interface{}, templateData map[string]interface{}) APIResponse {
	i18nManager := getI18nManager(c)
	lang := getLanguage(c)
	message := i18nManager.Translate(lang, messageKey, templateData)

	// 如果是错误响应，将错误详情包含在Data中
	if code != CodeSuccess && templateData != nil {
		if data == nil {
			data = templateData
		} else {
			// 如果data已经存在，将templateData合并到data中
			if dataMap, ok := data.(map[string]interface{}); ok {
				for k, v := range templateData {
					dataMap[k] = v
				}
			}
		}
	}

	return APIResponse{
		Code:      code,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
		RequestID: getRequestID(c),
	}
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	response := buildResponse(c, CodeSuccess, "success", data, nil)
	c.JSON(http.StatusOK, response)
}

// SuccessWithMessage 带自定义消息的成功响应
func SuccessWithMessage(c *gin.Context, messageKey string, data interface{}, templateData map[string]interface{}) {
	response := buildResponse(c, CodeSuccess, messageKey, data, templateData)
	c.JSON(http.StatusOK, response)
}

// ErrorResponse 错误响应
func ErrorResponse(c *gin.Context, messageKey string, templateData map[string]interface{}) {
	response := buildResponse(c, CodeFailure, messageKey, nil, templateData)
	c.JSON(http.StatusOK, response)
}

// ErrorWithData 带数据的错误响应
func ErrorWithData(c *gin.Context, messageKey string, data interface{}, templateData map[string]interface{}) {
	response := buildResponse(c, CodeFailure, messageKey, data, templateData)
	c.JSON(http.StatusOK, response)
}

// ErrorWithNotFund 获取不到请求
func ErrorWithNotFund(c *gin.Context, messageKey string, templateData map[string]interface{}) {
	response := buildResponse(c, CodeFailure, messageKey, nil, templateData)
	c.JSON(http.StatusBadRequest, response)
}

// ErrorResponseWithAuth 认证错误响应
func ErrorResponseWithAuth(c *gin.Context, messageKey string, templateData map[string]interface{}) {
	response := buildResponse(c, CodeUnauthorized, messageKey, nil, templateData)
	c.JSON(http.StatusOK, response)
}

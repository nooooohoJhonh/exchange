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

// ResponseBuilder 响应构建器
type ResponseBuilder struct {
	context *gin.Context
	i18n    *i18n.I18nManager
	lang    string
}

// NewResponseBuilder 创建响应构建器
func NewResponseBuilder(c *gin.Context) *ResponseBuilder {
	// 获取i18n管理器
	var i18nManager *i18n.I18nManager
	if manager, exists := c.Get("i18n"); exists {
		if mgr, ok := manager.(*i18n.I18nManager); ok {
			i18nManager = mgr
		}
	}
	if i18nManager == nil {
		i18nManager = i18n.GetGlobalI18n()
	}

	// 获取语言，使用i18n包的语言获取函数
	lang := i18n.GetLanguageFromContext(c)

	return &ResponseBuilder{
		context: c,
		i18n:    i18nManager,
		lang:    lang,
	}
}

// translate 翻译消息
func (rb *ResponseBuilder) translate(key string, templateData map[string]interface{}) string {
	return rb.i18n.Translate(rb.lang, key, templateData)
}

// getRequestID 获取请求ID
func (rb *ResponseBuilder) getRequestID() string {
	if requestID, exists := rb.context.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// buildResponse 构建响应
func (rb *ResponseBuilder) buildResponse(code int, messageKey string, data interface{}, templateData map[string]interface{}) APIResponse {
	message := rb.translate(messageKey, templateData)

	response := APIResponse{
		Code:      code,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
		RequestID: rb.getRequestID(),
	}

	return response
}

// Success 成功响应
func (rb *ResponseBuilder) Success(data interface{}) {
	response := rb.buildResponse(CodeSuccess, "success", data, nil)
	rb.context.JSON(http.StatusOK, response)
}

// SuccessWithMessage 带自定义消息的成功响应
func (rb *ResponseBuilder) SuccessWithMessage(messageKey string, data interface{}, templateData map[string]interface{}) {
	response := rb.buildResponse(CodeSuccess, messageKey, data, templateData)
	rb.context.JSON(http.StatusOK, response)
}

// Error 错误响应
func (rb *ResponseBuilder) Error(code int, messageKey string, templateData map[string]interface{}) {
	httpStatus := getHTTPStatus(code)
	response := rb.buildResponse(code, messageKey, templateData, templateData)
	rb.context.JSON(httpStatus, response)
}

// ErrorWithData 带数据的错误响应
func (rb *ResponseBuilder) ErrorWithData(code int, messageKey string, data interface{}, templateData map[string]interface{}) {
	httpStatus := getHTTPStatus(code)
	response := rb.buildResponse(code, messageKey, data, templateData)
	rb.context.JSON(httpStatus, response)
}

// 便捷函数（默认使用中文）

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	NewResponseBuilder(c).Success(data)
}

// SuccessWithMessage 带自定义消息的成功响应
func SuccessWithMessage(c *gin.Context, messageKey string, data interface{}, templateData map[string]interface{}) {
	NewResponseBuilder(c).SuccessWithMessage(messageKey, data, templateData)
}

// ErrorResponse 错误响应
func ErrorResponse(c *gin.Context, messageKey string, templateData map[string]interface{}) {
	NewResponseBuilder(c).Error(CodeFailure, messageKey, templateData)
}

// ErrorWithData 带数据的错误响应
func ErrorWithData(c *gin.Context, messageKey string, data interface{}, templateData map[string]interface{}) {
	NewResponseBuilder(c).ErrorWithData(CodeFailure, messageKey, data, templateData)
}

// ErrorResponseWithAuth 认证错误响应（返回401错误码）
func ErrorResponseWithAuth(c *gin.Context, messageKey string, templateData map[string]interface{}) {
	NewResponseBuilder(c).Error(CodeUnauthorized, messageKey, templateData)
}

// getHTTPStatus 根据错误码获取HTTP状态码
func getHTTPStatus(code int) int {
	// 所有响应都返回200状态码，前端通过code字段判断业务状态
	return http.StatusOK
}

package middleware

import (
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	appLogger "exchange/internal/pkg/logger"
)

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	RequestCount    int64         `json:"request_count"`
	TotalDuration   time.Duration `json:"total_duration"`
	AverageDuration time.Duration `json:"average_duration"`
	MinDuration     time.Duration `json:"min_duration"`
	MaxDuration     time.Duration `json:"max_duration"`
	ErrorCount      int64         `json:"error_count"`
	LastUpdated     time.Time     `json:"last_updated"`
}

// EndpointMetrics 端点指标
type EndpointMetrics struct {
	mu      sync.RWMutex
	metrics map[string]*PerformanceMetrics
}

var (
	globalMetrics = &EndpointMetrics{
		metrics: make(map[string]*PerformanceMetrics),
	}
)

// PerformanceMiddleware 性能监控中间件
func PerformanceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// 获取内存使用情况（请求前）
		var memBefore runtime.MemStats
		runtime.ReadMemStats(&memBefore)
		
		// 处理请求
		c.Next()
		
		// 计算处理时间
		duration := time.Since(start)
		
		// 获取内存使用情况（请求后）
		var memAfter runtime.MemStats
		runtime.ReadMemStats(&memAfter)
		
		// 构建端点标识
		endpoint := c.Request.Method + " " + c.FullPath()
		if endpoint == " " {
			endpoint = c.Request.Method + " " + c.Request.URL.Path
		}
		
		// 更新指标
		updateMetrics(endpoint, duration, c.Writer.Status() >= 400)
		
		// 记录性能日志
		fields := map[string]interface{}{
			"endpoint":        endpoint,
			"method":          c.Request.Method,
			"path":            c.Request.URL.Path,
			"status":          c.Writer.Status(),
			"duration":        duration.String(),
			"duration_ms":     float64(duration.Nanoseconds()) / 1e6,
			"memory_alloc":    memAfter.Alloc - memBefore.Alloc,
			"memory_total":    memAfter.TotalAlloc - memBefore.TotalAlloc,
			"goroutines":      runtime.NumGoroutine(),
			"client_ip":       c.ClientIP(),
			"user_agent":      c.Request.UserAgent(),
			"request_size":    c.Request.ContentLength,
			"response_size":   c.Writer.Size(),
		}
		
		// 添加用户信息
		if userID, exists := GetUserID(c); exists {
			fields["user_id"] = userID
		}
		
		// 记录慢请求
		if duration > 1*time.Second {
			fields["slow_request"] = true
			appLogger.Warn("Slow request detected", fields)
		}
		
		// 记录高内存使用
		memUsed := memAfter.Alloc - memBefore.Alloc
		if memUsed > 10*1024*1024 { // 10MB
			fields["high_memory_usage"] = true
			appLogger.Warn("High memory usage detected", fields)
		}
		
		// 记录性能日志
		appLogger.Performance("Request performance", fields)
	}
}

// updateMetrics 更新指标
func updateMetrics(endpoint string, duration time.Duration, isError bool) {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	
	metrics, exists := globalMetrics.metrics[endpoint]
	if !exists {
		metrics = &PerformanceMetrics{
			MinDuration: duration,
			MaxDuration: duration,
		}
		globalMetrics.metrics[endpoint] = metrics
	}
	
	// 更新计数
	metrics.RequestCount++
	metrics.TotalDuration += duration
	metrics.AverageDuration = time.Duration(int64(metrics.TotalDuration) / metrics.RequestCount)
	
	// 更新最小/最大时间
	if duration < metrics.MinDuration {
		metrics.MinDuration = duration
	}
	if duration > metrics.MaxDuration {
		metrics.MaxDuration = duration
	}
	
	// 更新错误计数
	if isError {
		metrics.ErrorCount++
	}
	
	metrics.LastUpdated = time.Now()
}

// GetMetrics 获取性能指标
func GetMetrics() map[string]*PerformanceMetrics {
	globalMetrics.mu.RLock()
	defer globalMetrics.mu.RUnlock()
	
	// 创建副本以避免并发访问问题
	result := make(map[string]*PerformanceMetrics)
	for endpoint, metrics := range globalMetrics.metrics {
		result[endpoint] = &PerformanceMetrics{
			RequestCount:    metrics.RequestCount,
			TotalDuration:   metrics.TotalDuration,
			AverageDuration: metrics.AverageDuration,
			MinDuration:     metrics.MinDuration,
			MaxDuration:     metrics.MaxDuration,
			ErrorCount:      metrics.ErrorCount,
			LastUpdated:     metrics.LastUpdated,
		}
	}
	
	return result
}

// ResetMetrics 重置指标
func ResetMetrics() {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	
	globalMetrics.metrics = make(map[string]*PerformanceMetrics)
	
	appLogger.Info("Performance metrics reset", nil)
}

// GetSystemMetrics 获取系统指标
func GetSystemMetrics() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return map[string]interface{}{
		"timestamp":     time.Now(),
		"goroutines":    runtime.NumGoroutine(),
		"memory": map[string]interface{}{
			"alloc":       m.Alloc,
			"total_alloc": m.TotalAlloc,
			"sys":         m.Sys,
			"num_gc":      m.NumGC,
			"gc_cpu_fraction": m.GCCPUFraction,
		},
		"gc": map[string]interface{}{
			"num_gc":        m.NumGC,
			"pause_total":   m.PauseTotalNs,
			"pause_avg":     float64(m.PauseTotalNs) / float64(m.NumGC+1),
		},
	}
}

// PerformanceReportMiddleware 性能报告中间件（定期输出性能报告）
func PerformanceReportMiddleware(interval time.Duration) gin.HandlerFunc {
	ticker := time.NewTicker(interval)
	
	go func() {
		for range ticker.C {
			generatePerformanceReport()
		}
	}()
	
	return func(c *gin.Context) {
		c.Next()
	}
}

// generatePerformanceReport 生成性能报告
func generatePerformanceReport() {
	metrics := GetMetrics()
	systemMetrics := GetSystemMetrics()
	
	report := map[string]interface{}{
		"timestamp":      time.Now(),
		"endpoints":      metrics,
		"system_metrics": systemMetrics,
		"summary": map[string]interface{}{
			"total_endpoints": len(metrics),
		},
	}
	
	// 计算总体统计
	var totalRequests int64
	var totalErrors int64
	var slowestEndpoint string
	var slowestDuration time.Duration
	
	for endpoint, metric := range metrics {
		totalRequests += metric.RequestCount
		totalErrors += metric.ErrorCount
		
		if metric.MaxDuration > slowestDuration {
			slowestDuration = metric.MaxDuration
			slowestEndpoint = endpoint
		}
	}
	
	summary := report["summary"].(map[string]interface{})
	summary["total_requests"] = totalRequests
	summary["total_errors"] = totalErrors
	summary["error_rate"] = float64(totalErrors) / float64(totalRequests) * 100
	summary["slowest_endpoint"] = slowestEndpoint
	summary["slowest_duration"] = slowestDuration.String()
	
	appLogger.Info("Performance report", report)
}

// SlowQueryMiddleware 慢查询监控中间件
func SlowQueryMiddleware(threshold time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		c.Next()
		
		duration := time.Since(start)
		
		if duration > threshold {
			fields := map[string]interface{}{
				"endpoint":    c.Request.Method + " " + c.FullPath(),
				"duration":    duration.String(),
				"duration_ms": float64(duration.Nanoseconds()) / 1e6,
				"threshold":   threshold.String(),
				"status":      c.Writer.Status(),
				"client_ip":   c.ClientIP(),
				"user_agent":  c.Request.UserAgent(),
			}
			
			if userID, exists := GetUserID(c); exists {
				fields["user_id"] = userID
			}
			
			appLogger.Warn("Slow query detected", fields)
		}
	}
}

// MemoryMonitorMiddleware 内存监控中间件
func MemoryMonitorMiddleware(threshold uint64) gin.HandlerFunc {
	return func(c *gin.Context) {
		var memBefore runtime.MemStats
		runtime.ReadMemStats(&memBefore)
		
		c.Next()
		
		var memAfter runtime.MemStats
		runtime.ReadMemStats(&memAfter)
		
		memUsed := memAfter.Alloc - memBefore.Alloc
		
		if memUsed > threshold {
			fields := map[string]interface{}{
				"endpoint":     c.Request.Method + " " + c.FullPath(),
				"memory_used":  memUsed,
				"threshold":    threshold,
				"memory_mb":    float64(memUsed) / 1024 / 1024,
				"status":       c.Writer.Status(),
				"client_ip":    c.ClientIP(),
			}
			
			if userID, exists := GetUserID(c); exists {
				fields["user_id"] = userID
			}
			
			appLogger.Warn("High memory usage detected", fields)
		}
	}
}
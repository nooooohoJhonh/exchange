package cron

import (
	"context"
	"net/http"
	"time"

	"exchange/internal/pkg/database"

	"github.com/gin-gonic/gin"
)

// Monitor Web监控界面
type Monitor struct {
	redis *database.RedisService
}

// NewMonitor 创建Web监控界面
func NewMonitor(redis *database.RedisService) *Monitor {
	return &Monitor{
		redis: redis,
	}
}

// RegisterRoutes 注册Web路由
func (m *Monitor) RegisterRoutes(r *gin.Engine) {
	// 静态文件
	r.Static("/static", "./static")

	// 主页
	r.GET("/", m.Index)

	// API接口
	api := r.Group("/api")
	{
		api.GET("/status", m.GetStatus)
		api.GET("/instances", m.GetInstances)
		api.GET("/tasks", m.GetTasks)
	}
}

// Index 主页
func (m *Monitor) Index(c *gin.Context) {
	c.HTML(http.StatusOK, "monitor.html", gin.H{
		"title": "分布式定时任务监控",
		"time":  time.Now().Format("2006-01-02 15:04:05"),
	})
}

// GetStatus 获取服务状态
func (m *Monitor) GetStatus(c *gin.Context) {
	// 获取活跃实例数量
	instances, err := m.getActiveInstances(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	status := map[string]interface{}{
		"status":         "running",
		"instance_count": len(instances),
		"start_time":     time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}

// GetInstances 获取实例列表
func (m *Monitor) GetInstances(c *gin.Context) {
	instances, err := m.getActiveInstances(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    instances,
	})
}

// GetTasks 获取任务列表
func (m *Monitor) GetTasks(c *gin.Context) {
	// 从所有实例中收集任务信息
	instances, err := m.getActiveInstances(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// 收集所有任务
	taskMap := make(map[string]interface{})
	for _, instance := range instances {
		for _, taskName := range instance.Tasks {
			if _, exists := taskMap[taskName]; !exists {
				taskMap[taskName] = map[string]interface{}{
					"name":        taskName,
					"description": "任务描述",
					"instances":   []string{instance.InstanceID},
				}
			} else {
				// 添加实例到现有任务
				if task, ok := taskMap[taskName].(map[string]interface{}); ok {
					if instances, ok := task["instances"].([]string); ok {
						instances = append(instances, instance.InstanceID)
						task["instances"] = instances
					}
				}
			}
		}
	}

	// 转换为数组
	var tasks []map[string]interface{}
	for _, task := range taskMap {
		tasks = append(tasks, task.(map[string]interface{}))
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tasks,
	})
}

// getActiveInstances 获取活跃实例
func (m *Monitor) getActiveInstances(ctx context.Context) ([]*InstanceInfo, error) {
	instanceRegistry := NewInstanceRegistry(m.redis, "1.0.0")
	return instanceRegistry.GetActiveInstances(ctx)
}

package cron

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// APIHandler 分布式定时任务API处理器
type APIHandler struct {
	manager *DistributedTaskManager
}

// NewAPIHandler 创建API处理器
func NewAPIHandler(manager *DistributedTaskManager) *APIHandler {
	return &APIHandler{
		manager: manager,
	}
}

// RegisterRoutes 注册API路由
func (h *APIHandler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/cron")
	{
		api.GET("/status", h.GetStatus)
		api.GET("/instances", h.GetInstances)
		api.GET("/tasks", h.GetTasks)
		api.GET("/tasks/:name/stats", h.GetTaskStats)
		api.GET("/tasks/:name/history", h.GetTaskHistory)
	}
}

// StatusResponse 状态响应
type StatusResponse struct {
	InstanceID    string    `json:"instance_id"`
	Status        string    `json:"status"`
	StartTime     time.Time `json:"start_time"`
	ActiveTasks   int       `json:"active_tasks"`
	InstanceCount int       `json:"instance_count"`
	Uptime        string    `json:"uptime"`
}

// GetStatus 获取服务状态
func (h *APIHandler) GetStatus(c *gin.Context) {
	ctx := context.Background()

	instanceCount, err := h.manager.GetInstanceCount(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取实例数量失败: " + err.Error(),
		})
		return
	}

	// 获取活跃任务数量
	activeTasks := 0
	instances, err := h.manager.GetActiveInstances(ctx)
	if err == nil {
		// 统计所有实例中的任务数量（去重）
		taskSet := make(map[string]bool)
		for _, instance := range instances {
			for _, taskName := range instance.Tasks {
				taskSet[taskName] = true
			}
		}
		activeTasks = len(taskSet)
	}

	response := &StatusResponse{
		InstanceID:    h.manager.GetInstanceID(),
		Status:        "running",
		StartTime:     time.Now(), // 这里可以添加实际的启动时间
		ActiveTasks:   activeTasks,
		InstanceCount: instanceCount,
		Uptime:        "0s", // 这里可以计算实际的运行时间
	}

	c.JSON(http.StatusOK, response)
}

// GetInstances 获取实例列表
func (h *APIHandler) GetInstances(c *gin.Context) {
	ctx := context.Background()

	instances, err := h.manager.GetActiveInstances(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取实例列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"instances": instances,
		"count":     len(instances),
	})
}

// GetTasks 获取任务列表（从Redis中获取）
func (h *APIHandler) GetTasks(c *gin.Context) {
	ctx := context.Background()

	// 从Redis中获取所有活跃实例的任务信息
	instances, err := h.manager.GetActiveInstances(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取实例列表失败: " + err.Error(),
		})
		return
	}

	// 收集所有任务信息
	taskMap := make(map[string]gin.H)
	for _, instance := range instances {
		for _, taskName := range instance.Tasks {
			if _, exists := taskMap[taskName]; !exists {
				taskMap[taskName] = gin.H{
					"name":        taskName,
					"description": "从实例中获取的任务",
					"schedule":    "未知", // 这里可以从Redis中获取更详细的信息
					"instances":   []string{instance.InstanceID},
				}
			} else {
				// 如果任务已存在，添加实例ID
				if instances, ok := taskMap[taskName]["instances"].([]string); ok {
					taskMap[taskName]["instances"] = append(instances, instance.InstanceID)
				}
			}
		}
	}

	// 转换为切片
	var tasks []gin.H
	for _, task := range taskMap {
		tasks = append(tasks, task)
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
		"count": len(tasks),
	})
}

// GetTaskStats 获取任务统计信息
func (h *APIHandler) GetTaskStats(c *gin.Context) {
	taskName := c.Param("name")
	ctx := context.Background()

	stats, err := h.manager.GetTaskStats(ctx, taskName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取任务统计信息失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_name": taskName,
		"stats":     stats,
	})
}

// GetTaskHistory 获取任务执行历史
func (h *APIHandler) GetTaskHistory(c *gin.Context) {
	taskName := c.Param("name")
	ctx := context.Background()

	history, err := h.manager.stateManager.GetTaskExecutionHistory(ctx, taskName, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取任务执行历史失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_name": taskName,
		"history":   history,
		"count":     len(history),
	})
}

// HealthCheck 健康检查
func (h *APIHandler) HealthCheck(c *gin.Context) {
	ctx := context.Background()

	// 检查Redis连接
	_, err := h.manager.GetInstanceCount(ctx)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  "Redis连接失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "healthy",
		"instance_id": h.manager.GetInstanceID(),
		"timestamp":   time.Now().Unix(),
		"mode":        "monitoring",
	})
}

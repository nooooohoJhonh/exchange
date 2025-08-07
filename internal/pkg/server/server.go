package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"exchange/internal/pkg/config"
	"exchange/internal/pkg/logger"
)

// GinServer Gin服务器
type GinServer struct {
	config     *config.Config
	engine     *gin.Engine
	httpServer *http.Server
}

// NewGinServer 创建Gin服务器
func NewGinServer(cfg *config.Config) *GinServer {
	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 创建Gin引擎
	engine := gin.New()

	server := &GinServer{
		config: cfg,
		engine: engine,
	}

	return server
}

// SetupRoutes 设置路由（由外部模块调用）
func (s *GinServer) SetupRoutes(setupFunc func(*gin.Engine)) {
	setupFunc(s.engine)
	logger.Info("Routes setup successfully", nil)
}

// Start 启动服务器
func (s *GinServer) Start() error {
	// 创建HTTP服务器
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.Server.Port),
		Handler:      s.engine,
		ReadTimeout:  time.Duration(s.config.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(s.config.Server.WriteTimeout) * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 启动服务器
	logger.Info("Starting HTTP server", map[string]interface{}{
		"address": fmt.Sprintf(":%d", s.config.Server.Port),
		"mode":    s.config.Server.Mode,
	})

	// 检查端口是否被占用
	listener, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		logger.Error("Port is already in use", map[string]interface{}{
			"port":  s.config.Server.Port,
			"error": err.Error(),
		})
		return fmt.Errorf("port %d is already in use: %w", s.config.Server.Port, err)
	}
	listener.Close()

	// 在goroutine中启动服务器
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed to start", map[string]interface{}{
				"error": err.Error(),
				"port":  s.config.Server.Port,
			})
		}
	}()

	// 等待中断信号
	return s.waitForShutdown()
}

// waitForShutdown 等待关闭信号
func (s *GinServer) waitForShutdown() error {
	// 创建信号通道
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号
	<-quit
	logger.Info("Shutting down server...", nil)

	// 优雅关闭
	return s.Shutdown()
}

// Shutdown 优雅关闭服务器
func (s *GinServer) Shutdown() error {
	// 创建超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 关闭HTTP服务器
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			logger.Error("HTTP server forced to shutdown", map[string]interface{}{
				"error": err.Error(),
			})
			return err
		}
	}

	logger.Info("Server shutdown complete", nil)
	return nil
}

// GetEngine 获取Gin引擎（用于测试）
func (s *GinServer) GetEngine() *gin.Engine {
	return s.engine
}

// GetConfig 获取配置
func (s *GinServer) GetConfig() *config.Config {
	return s.config
}

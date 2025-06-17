package http

import (
	"context"
	"errors"
	"fmt"
	"github.com/blackarbiter/go-sac/internal/storage/service"
	"github.com/blackarbiter/go-sac/pkg/config"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// Server 实现HTTP服务器
type Server struct {
	config  *config.Config
	factory service.StorageProcessorFactory
	engine  *gin.Engine
	server  *http.Server
}

// NewServer 创建HTTP服务器实例
func NewServer(cfg *config.Config, factory service.StorageProcessorFactory) *Server {
	// 创建Gin引擎
	engine := gin.Default()

	// 创建服务器实例
	server := &Server{
		config:  cfg,
		factory: factory,
		engine:  engine,
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.Server.HTTP.Port),
			Handler: engine,
		},
	}

	// 注册路由
	handler := NewHandler(factory)
	handler.RegisterRoutes(engine)

	return server
}

// Start 启动HTTP服务器
func (s *Server) Start(ctx context.Context) error {
	// 启动服务器
	go func() {
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	// 等待上下文取消
	<-ctx.Done()

	// 创建关闭上下文
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 优雅关闭服务器
	return s.server.Shutdown(shutdownCtx)
}

// Stop 停止HTTP服务器
func (s *Server) Stop(ctx context.Context) error {
	// 创建关闭上下文
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 优雅关闭服务器
	return s.server.Shutdown(shutdownCtx)
}

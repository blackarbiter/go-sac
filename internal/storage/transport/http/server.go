package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/blackarbiter/go-sac/internal/storage/service"
	"github.com/blackarbiter/go-sac/pkg/config"
	"github.com/blackarbiter/go-sac/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Server 表示HTTP服务器
type Server struct {
	router *gin.Engine
	server *http.Server
	cfg    *config.Config
}

// NewServer 创建一个新的HTTP服务器
func NewServer(cfg *config.Config, storageService service.StorageService) *Server {
	// 设置存储服务
	SetStorageService(storageService)

	// 创建路由
	router := NewRouter()

	return &Server{
		router: router,
		cfg:    cfg,
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.Server.HTTP.Port),
			Handler: router,
		},
	}
}

// Start 启动HTTP服务器
func (s *Server) Start(ctx context.Context) error {
	logger.Logger.Info("starting HTTP server", zap.Int("port", s.cfg.Server.HTTP.Port))

	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

// Stop 停止HTTP服务器
func (s *Server) Stop(ctx context.Context) {
	logger.Logger.Info("stopping HTTP server")

	if err := s.server.Shutdown(ctx); err != nil {
		logger.Logger.Error("server shutdown error", zap.Error(err))
	}

	logger.Logger.Info("HTTP server stopped")
}

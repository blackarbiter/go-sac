package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/blackarbiter/go-sac/pkg/config"
	"github.com/blackarbiter/go-sac/pkg/logger"
	"go.uber.org/zap"
)

//go:generate wire
func main() {
	// 初始化上下文
	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 配置加载
	cfg, err := config.Load("scan")
	if err != nil {
		panic("config load failed: " + err.Error())
	}

	// 日志初始化
	logger.InitZapWithRotation("dev")

	// 依赖注入构建应用
	app, cleanup, err := InitializeApplication(cfg)
	if err != nil {
		logger.Logger.Fatal("initialize failed", zap.Error(err))
	}
	defer cleanup()

	// 启动扫描服务
	go func() {
		if err := app.ScanService.Start(ctx); err != nil {
			logger.Logger.Error("scan service error", zap.Error(err))
		}
	}()

	logger.Logger.Info("scan service started")

	// 优雅停机处理
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	app.ScanService.Stop(shutdownCtx)
	logger.Logger.Info("service stopped gracefully")
}

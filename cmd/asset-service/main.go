package main

import (
	"context"
	"github.com/blackarbiter/go-sac/pkg/mq/rabbitmq"
	"os/signal"
	"syscall"
	"time"

	"github.com/blackarbiter/go-sac/pkg/config"
	"github.com/blackarbiter/go-sac/pkg/domain"
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
	cfg, err := config.Load("asset")
	if err != nil {
		panic("config load failed: " + err.Error())
	}

	// 日志初始化，使用lumberjack进行日志轮转
	logger.InitZapWithRotation("dev")

	// 依赖注入构建应用
	app, cleanup, err := InitializeApplication(cfg)
	if err != nil {
		logger.Logger.Fatal("initialize failed", zap.Error(err))
	}
	defer cleanup()

	// 注册所有资产处理器
	registerAssetProcessors(app)

	// 启动HTTP服务
	go func() {
		if err := app.HTTPServer.Start(ctx); err != nil {
			logger.Logger.Error("http server error", zap.Error(err))
		}
	}()

	logger.Logger.Info("asset service started")

	// 启动MQ消费者
	go func() {
		logger.Logger.Info("starting RabbitMQ consumer")
		if err := app.MQConsumer.Consume(context.Background(), rabbitmq.AssetTaskQueue, app.AssetHandler); err != nil {
			logger.Logger.Error("MQ consumer failed", zap.Error(err))
		}
	}()

	// 优雅停机处理
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	app.HTTPServer.Stop(shutdownCtx)
	logger.Logger.Info("service stopped gracefully")

	// 停止MQ消费者
	if err := app.MQConsumer.Close(); err != nil {
		logger.Logger.Error("MQ consumer close error", zap.Error(err))
	}

	logger.Logger.Info("rabbitmq stopped gracefully")
}

// registerAssetProcessors 注册所有资产处理器
func registerAssetProcessors(app *Application) {
	app.Factory.RegisterDefaultProcessors(app.Repository)

	logger.Logger.Info("all asset processors registered",
		zap.Strings("processors", []string{
			domain.AssetTypeRequirement.String(),
			domain.AssetTypeDesignDocument.String(),
			domain.AssetTypeRepository.String(),
			domain.AssetTypeUploadedFile.String(),
			domain.AssetTypeImage.String(),
			domain.AssetTypeDomain.String(),
			domain.AssetTypeIP.String(),
		}))
}

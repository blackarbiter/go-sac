// cmd/storage-service/wire.go
//go:build wireinject
// +build wireinject

package main

import (
	"github.com/blackarbiter/go-sac/internal/storage/repository"
	"github.com/blackarbiter/go-sac/internal/storage/service"
	"github.com/blackarbiter/go-sac/internal/storage/transport/http"
	"github.com/blackarbiter/go-sac/internal/storage/transport/mq"
	"github.com/blackarbiter/go-sac/pkg/config"
	"github.com/blackarbiter/go-sac/pkg/mq/rabbitmq"
	"github.com/blackarbiter/go-sac/pkg/storage/minio"
	"github.com/google/wire"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Application 聚合所有核心组件
type Application struct {
	HTTPServer     *http.Server
	DB             *gorm.DB
	Factory        *service.ProcessorFactory
	Repository     repository.Repository
	MQConsumer     *rabbitmq.ResultConsumer
	StorageHandler *mq.StorageMessageHandler
}

var (
	// ApplicationSet 是整个应用的依赖集合
	ApplicationSet = wire.NewSet(
		// 应用结构
		wire.Struct(new(Application), "*"),
		ProvideLogger,
		ProvideProcessorFactory,
		ProvideMinIOStorage,

		// 类型绑定
		wire.Bind(new(service.StorageProcessorFactory), new(*service.ProcessorFactory)),

		// 导入各层的Provider集合
		repository.ProviderSet,
		service.ProviderSet,
		http.ProviderSet,
		ProvideStorageMessageHandler,
	)
)

// ProvideLogger 提供日志实例
func ProvideLogger() *zap.Logger {
	logger, _ := zap.NewProduction()
	return logger
}

// ProvideMinIOStorage 提供MinIO存储实例
func ProvideMinIOStorage(cfg *config.Config) (*minio.Storage, error) {
	return minio.NewStorage(
		cfg.Storage.MinIO.Endpoint,
		cfg.Storage.MinIO.AccessKey,
		cfg.Storage.MinIO.SecretKey,
		cfg.Storage.MinIO.Bucket,
		cfg.Storage.MinIO.UseSSL,
	)
}

// InitializeApplication 通过Wire自动生成
func InitializeApplication(cfg *config.Config) (*Application, func(), error) {
	panic(wire.Build(ApplicationSet))
}

// ProvideProcessorFactory 提供处理器工厂
func ProvideProcessorFactory() *service.ProcessorFactory {
	return service.NewProcessorFactory()
}

// provideAssetMessageHandler 提供资产消息处理器
func ProvideStorageMessageHandler(
	factory service.StorageProcessorFactory,
) *mq.StorageMessageHandler {
	return mq.NewStorageMessageHandler(factory)
}

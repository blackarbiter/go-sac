// cmd/asset-service/wire.go
//go:build wireinject
// +build wireinject

package main

import (
	"github.com/blackarbiter/go-sac/internal/asset/repository"
	"github.com/blackarbiter/go-sac/internal/asset/service"
	"github.com/blackarbiter/go-sac/internal/asset/transport/http"
	"github.com/blackarbiter/go-sac/internal/asset/transport/mq"
	"github.com/blackarbiter/go-sac/pkg/config"
	"github.com/blackarbiter/go-sac/pkg/mq/rabbitmq"
	"github.com/google/wire"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Application 聚合所有核心组件
type Application struct {
	HTTPServer   *http.Server
	DB           *gorm.DB
	Factory      *service.ProcessorFactory
	Repository   repository.Repository
	MQConsumer   *rabbitmq.AssetConsumer // 新增
	AssetBinder  *http.AssetBinder       // 新增
	AssetHandler *mq.AssetMessageHandler // 新增
}

var (
	// ApplicationSet 是整个应用的依赖集合
	ApplicationSet = wire.NewSet(
		// 应用结构
		wire.Struct(new(Application), "*"),

		// 服务组件
		ProvideLogger,
		ProvideProcessorFactory,

		// 类型绑定
		wire.Bind(new(service.AssetProcessorFactory), new(*service.ProcessorFactory)),

		// 导入各层的Provider集合
		repository.ProviderSet,
		service.ProviderSet,
		http.ProviderSet,
		// 提供消息处理器
		provideAssetMessageHandler,
	)
)

// ProvideLogger 提供日志实例
func ProvideLogger() *zap.Logger {
	logger, _ := zap.NewProduction()
	return logger
}

// ProvideProcessorFactory 提供处理器工厂
func ProvideProcessorFactory() *service.ProcessorFactory {
	return service.NewProcessorFactory()
}

// InitializeApplication 通过Wire自动生成
func InitializeApplication(cfg *config.Config) (*Application, func(), error) {
	panic(wire.Build(ApplicationSet))
}

// provideAssetMessageHandler 提供资产消息处理器
func provideAssetMessageHandler(
	binder *http.AssetBinder,
	factory service.AssetProcessorFactory,
) *mq.AssetMessageHandler {
	return mq.NewAssetMessageHandler(binder, factory)
}

//go:build wireinject
// +build wireinject

package main

import (
	"github.com/blackarbiter/go-sac/internal/scan/service"
	"github.com/blackarbiter/go-sac/pkg/config"
	"github.com/blackarbiter/go-sac/pkg/mq/rabbitmq"
	"github.com/blackarbiter/go-sac/pkg/scanner"
	"github.com/blackarbiter/go-sac/pkg/scanner/impl"
	"github.com/google/wire"
)

// Application 聚合所有核心组件
type Application struct {
	ScanService *service.ScanService
}

var (
	// ApplicationSet 是整个应用的依赖集合
	ApplicationSet = wire.NewSet(
		// 应用结构
		wire.Struct(new(Application), "*"),

		// 服务组件
		service.ProviderSet,

		// 基础设施组件
		provideConnectionManager,
		wire.Bind(new(scanner.ScannerFactory), new(*impl.ScannerFactoryImpl)),
		impl.NewScannerFactory,
	)
)

// provideConnectionManager 提供 RabbitMQ 连接管理器
func provideConnectionManager(cfg *config.Config) *rabbitmq.ConnectionManager {
	return rabbitmq.NewConnectionManager(cfg.GetRabbitMQURL(), 3)
}

// InitializeApplication 通过Wire自动生成
func InitializeApplication(cfg *config.Config) (*Application, func(), error) {
	panic(wire.Build(ApplicationSet))
}

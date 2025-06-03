//go:build wireinject
// +build wireinject

package main

import (
	"context"

	"github.com/blackarbiter/go-sac/internal/scan/service"
	"github.com/blackarbiter/go-sac/pkg/cache/redis"
	"github.com/blackarbiter/go-sac/pkg/config"
	"github.com/blackarbiter/go-sac/pkg/domain"
	"github.com/blackarbiter/go-sac/pkg/logger"
	"github.com/blackarbiter/go-sac/pkg/metrics"
	"github.com/blackarbiter/go-sac/pkg/mq/rabbitmq"
	"github.com/blackarbiter/go-sac/pkg/scanner"
	scanner_impl "github.com/blackarbiter/go-sac/pkg/scanner/impl"
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
		provideMetrics,
		provideTimeoutController,
		provideRedisConnector,
		provideCircuitBreaker,
		wire.Bind(new(scanner.ScannerFactory), new(*scanner.ScannerFactoryImpl)),
		provideScannerFactory,
	)
)

// provideConnectionManager 提供 RabbitMQ 连接管理器
func provideConnectionManager(cfg *config.Config) *rabbitmq.ConnectionManager {
	return rabbitmq.NewConnectionManager(cfg.GetRabbitMQURL(), 3)
}

// provideMetrics 提供指标收集器
func provideMetrics() *metrics.ScannerMetrics {
	metrics := metrics.NewScannerMetrics()
	metrics.Register()
	return metrics
}

// provideTimeoutController 提供超时控制器
func provideTimeoutController(metrics *metrics.ScannerMetrics) *scanner.TimeoutController {
	return scanner.NewTimeoutController(metrics)
}

// provideRedisConnector 提供 Redis 连接器
func provideRedisConnector(cfg *config.Config) (*redis.Connector, error) {
	return redis.NewConnector(
		context.Background(),
		cfg.GetRedisAddr(),
		cfg.GetRedisPassword(),
		cfg.GetRedisDB(),
		cfg.GetRedisPoolSize(),
	)
}

func provideCircuitBreaker(cfg *config.Config) *scanner.CircuitBreaker {
	threshold, criticalThreshold, resetTimeout := cfg.GetCircuitBreakerConfig()
	return scanner.NewCircuitBreaker(threshold, criticalThreshold, resetTimeout)
}

// provideScannerFactory provides a scanner factory with default scanners
func provideScannerFactory(
	timeoutCtrl *scanner.TimeoutController,
	metrics *metrics.ScannerMetrics,
	circuitBreaker *scanner.CircuitBreaker,
	cfg *config.Config,
) *scanner.ScannerFactoryImpl {
	return scanner.NewScannerFactory(
		func() map[domain.ScanType]scanner.TaskExecutor {
			return scanner_impl.CreateDefaultScanners(timeoutCtrl, logger.Logger, metrics, nil, cfg)
		}, metrics, circuitBreaker,
	)
}

// InitializeApplication 通过Wire自动生成
func InitializeApplication(cfg *config.Config) (*Application, func(), error) {
	panic(wire.Build(ApplicationSet))
}

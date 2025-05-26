// cmd/task-service/wire.go
//go:build wireinject
// +build wireinject

package main

import (
	"github.com/blackarbiter/go-sac/internal/task/repository"
	"github.com/blackarbiter/go-sac/internal/task/service"
	"github.com/blackarbiter/go-sac/internal/task/transport/http"
	"github.com/blackarbiter/go-sac/pkg/config"
	"github.com/google/wire"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Application 聚合所有核心组件
type Application struct {
	HTTPServer *http.Server
	DB         *gorm.DB
}

var (
	// ApplicationSet 是整个应用的依赖集合
	ApplicationSet = wire.NewSet(
		// 应用结构
		wire.Struct(new(Application), "*"),

		// 服务组件
		ProvideHTTPServer,
		ProvideLogger,

		// 导入各层的Provider集合
		repository.ProviderSet,
		service.ProviderSet,
	)
)

// ProvideHTTPServer 提供HTTP服务实例
func ProvideHTTPServer(cfg *config.Config, taskService service.TaskService) *http.Server {
	return http.NewServer(cfg, taskService)
}

// ProvideLogger 提供日志实例
func ProvideLogger() *zap.Logger {
	logger, _ := zap.NewProduction()
	return logger
}

// InitializeApplication 通过Wire自动生成
func InitializeApplication(cfg *config.Config) (*Application, func(), error) {
	panic(wire.Build(ApplicationSet))
}

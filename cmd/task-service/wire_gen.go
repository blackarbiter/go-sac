// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/blackarbiter/go-sac/internal/task/repository"
	"github.com/blackarbiter/go-sac/internal/task/service"
	"github.com/blackarbiter/go-sac/internal/task/transport/http"
	"github.com/blackarbiter/go-sac/pkg/config"
	"github.com/blackarbiter/go-sac/pkg/storage/mysql"
	"github.com/google/wire"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Injectors from wire.go:

// InitializeApplication 通过Wire自动生成
func InitializeApplication(cfg *config.Config) (*Application, func(), error) {
	logger := ProvideLogger()
	db, err := mysql.ProvideGormDB(cfg, logger)
	if err != nil {
		return nil, nil, err
	}
	taskRepository := repository.ProvideTaskRepository(db, cfg)
	taskPublisher, err := service.ProvideTaskPublisher(cfg)
	if err != nil {
		return nil, nil, err
	}
	taskService := service.ProvideTaskService(taskRepository, taskPublisher)
	server := ProvideHTTPServer(cfg, taskService)
	application := &Application{
		HTTPServer: server,
		DB:         db,
	}
	return application, func() {
	}, nil
}

// wire.go:

// Application 聚合所有核心组件
type Application struct {
	HTTPServer *http.Server
	DB         *gorm.DB
}

var (
	// ApplicationSet 是整个应用的依赖集合
	ApplicationSet = wire.NewSet(wire.Struct(new(Application), "*"), ProvideHTTPServer,
		ProvideLogger, repository.ProviderSet, service.ProviderSet,
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

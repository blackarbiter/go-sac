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
)

// Application 聚合所有核心组件
type Application struct {
	HTTPServer *http.Server
}

var (
	// ApplicationSet 是整个应用的依赖集合
	ApplicationSet = wire.NewSet(
		// 应用结构
		wire.Struct(new(Application), "*"),

		// 服务组件
		ProvideHTTPServer,
		ProvideTaskService,
		ProvideTaskRepository,
	)
)

// ProvideTaskRepository 提供任务仓库实例
func ProvideTaskRepository(cfg *config.Config) repository.TaskRepository {
	return repository.NewTaskRepository(cfg)
}

// ProvideTaskService 提供任务服务实例
func ProvideTaskService(repo repository.TaskRepository) service.TaskService {
	return service.NewTaskService(repo)
}

// ProvideHTTPServer 提供HTTP服务实例
func ProvideHTTPServer(cfg *config.Config, taskService service.TaskService) *http.Server {
	return http.NewServer(cfg, taskService)
}

// InitializeApplication 通过Wire自动生成
func InitializeApplication(cfg *config.Config) (*Application, func(), error) {
	panic(wire.Build(ApplicationSet))
}

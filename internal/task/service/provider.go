package service

import (
	"github.com/blackarbiter/go-sac/internal/task/repository"
	"github.com/blackarbiter/go-sac/pkg/config"
	"github.com/blackarbiter/go-sac/pkg/mq/rabbitmq"
	"github.com/google/wire"
)

// ProviderSet 是任务服务提供者集合
var ProviderSet = wire.NewSet(
	ProvideTaskService,
	ProvideTaskPublisher,
)

// ProvideTaskService 提供任务服务实例
func ProvideTaskService(repo repository.TaskRepository, publisher *rabbitmq.TaskPublisher) TaskService {
	return NewTaskService(repo, publisher)
}

// ProvideTaskPublisher 提供任务发布者实例
func ProvideTaskPublisher(cfg *config.Config) (*rabbitmq.TaskPublisher, error) {
	// 获取RabbitMQ连接URL
	rabbitURL := cfg.GetRabbitMQURL()

	// 创建RabbitMQ连接管理器
	connManager := rabbitmq.NewConnectionManager(rabbitURL, 3)

	// 获取连接
	conn, err := connManager.GetConnection()
	if err != nil {
		return nil, err
	}

	// 初始化RabbitMQ基础设施
	if err := rabbitmq.Setup(conn); err != nil {
		return nil, err
	}

	// 创建任务发布者
	publisher, err := rabbitmq.NewTaskPublisher(conn)
	if err != nil {
		return nil, err
	}

	return publisher, nil
}

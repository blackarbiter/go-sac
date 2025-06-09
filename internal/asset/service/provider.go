package service

import (
	"github.com/blackarbiter/go-sac/pkg/config"
	"github.com/blackarbiter/go-sac/pkg/mq/rabbitmq"
	"github.com/google/wire"
)

// ProviderSet 是 service 层的依赖注入集合
var ProviderSet = wire.NewSet(
	ProvideAssetConsumer,
)

// ProvideAssetConsumer 提供任务发布者实例
func ProvideAssetConsumer(cfg *config.Config) (*rabbitmq.AssetConsumer, error) {
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
	consumer, err := rabbitmq.NewAssetConsumer(conn)
	if err != nil {
		return nil, err
	}

	return consumer, nil
}

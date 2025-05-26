package rabbitmq

import (
	"context"
	"fmt"
	"time"

	"github.com/blackarbiter/go-sac/pkg/mq"
	amqp "github.com/rabbitmq/amqp091-go"
)

// TaskPublisher implements the TaskPublisher interface for RabbitMQ
type TaskPublisher struct {
	conn     *amqp.Connection
	producer *EnhancedProducer
}

// NewTaskPublisher creates a new instance of TaskPublisher
func NewTaskPublisher(conn *amqp.Connection) (*TaskPublisher, error) {
	// 检查exchange是否存在
	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}
	defer channel.Close() // 只用于检查exchange，后续会创建新channel

	// 确认当前exchange存在
	err = channel.ExchangeDeclarePassive(
		TaskDispatchExchange,
		"topic", // exchange类型
		true,    // durable
		false,   // auto-delete
		false,   // internal
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("exchange %s not found, please run setup first: %w", TaskDispatchExchange, err)
	}

	// 使用已有的连接创建EnhancedProducer
	config := ProducerConfig{
		RetryCount:    3,
		RetryInterval: time.Second,
	}

	producer, err := NewEnhancedProducerWithConnection(conn, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	return &TaskPublisher{
		conn:     conn,
		producer: producer,
	}, nil
}

// Publish sends a message to specified exchange with routing key
func (p *TaskPublisher) Publish(ctx context.Context, exchange, routingKey string, message []byte) error {
	return p.producer.Publish(ctx, exchange, routingKey, message)
}

// PublishScanTask publishes a scan task with the given type and priority
func (p *TaskPublisher) PublishScanTask(ctx context.Context, taskType string, priority int, payload []byte) error {
	// Map priority level to routing key suffix
	priorityMap := map[int]string{
		0: "low",
		1: "medium",
		2: "high",
	}

	priorityStr, ok := priorityMap[priority]
	if !ok {
		priorityStr = "low" // Default to low priority
	}

	// Create routing key in format "scan.[type].[priority]"
	routingKey := fmt.Sprintf("scan.%s.%s", taskType, priorityStr)

	// Set message priority based on level
	priorityValue := priority * 5 // Convert to 0/5/10

	// Create message with appropriate headers
	return p.producer.PublishWithHeaders(
		ctx,
		TaskDispatchExchange,
		routingKey,
		amqp.Table{
			"taskType": taskType,
			"priority": priority,
		},
		uint8(priorityValue),
		payload,
	)
}

// PublishAssetTask publishes an asset task with the given operation
func (p *TaskPublisher) PublishAssetTask(ctx context.Context, operation string, payload []byte) error {
	// Create routing key in format "asset.[operation]"
	routingKey := fmt.Sprintf("asset.%s", operation)

	// Create message with appropriate headers
	return p.producer.PublishWithHeaders(
		ctx,
		TaskDispatchExchange,
		routingKey,
		amqp.Table{
			"operation": operation,
		},
		0, // No priority for asset tasks
		payload,
	)
}

// DeleteScanTask 从消息队列中删除扫描任务
func (p *TaskPublisher) DeleteScanTask(ctx context.Context, scanType string, priority int, payload []byte) error {
	// 构建路由键
	routingKey := fmt.Sprintf("scan.%s", scanType)

	// 从队列中删除消息
	return p.producer.PublishWithHeaders(
		ctx,
		"",         // exchange
		routingKey, // routing key
		amqp.Table{
			"priority": priority,
		},
		uint8(priority),
		payload,
	)
}

// DeleteAssetTask 从消息队列中删除资产任务
func (p *TaskPublisher) DeleteAssetTask(ctx context.Context, operation string, payload []byte) error {
	// 构建路由键
	routingKey := fmt.Sprintf("asset.%s", operation)

	// 从队列中删除消息
	return p.producer.PublishWithHeaders(
		ctx,
		"",         // exchange
		routingKey, // routing key
		amqp.Table{},
		0, // 默认优先级
		payload,
	)
}

// Close closes the publisher
func (p *TaskPublisher) Close() error {
	return p.producer.Close()
}

// Ensure TaskPublisher implements the mq.TaskPublisher interface
var _ mq.TaskPublisher = (*TaskPublisher)(nil)

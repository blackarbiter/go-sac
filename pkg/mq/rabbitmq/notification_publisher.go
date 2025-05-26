package rabbitmq

import (
	"context"
	"fmt"
	"time"

	"github.com/blackarbiter/go-sac/pkg/mq"
	amqp "github.com/rabbitmq/amqp091-go"
)

// NotificationPublisher 实现通知发布接口
type NotificationPublisher struct {
	conn     *amqp.Connection
	producer *EnhancedProducer
}

// NewNotificationPublisher 创建通知发布者实例
func NewNotificationPublisher(conn *amqp.Connection) (*NotificationPublisher, error) {
	// 检查exchange是否存在
	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}
	defer channel.Close()

	// 确认当前exchange存在
	err = channel.ExchangeDeclarePassive(
		NotificationExchange,
		"fanout", // exchange类型
		true,     // durable
		false,    // auto-delete
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("exchange %s not found, please run setup first: %w", NotificationExchange, err)
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

	return &NotificationPublisher{
		conn:     conn,
		producer: producer,
	}, nil
}

// PublishNotification 发布通知消息
func (p *NotificationPublisher) PublishNotification(ctx context.Context, payload []byte) error {
	return p.producer.Publish(
		ctx,
		NotificationExchange,
		"", // fanout exchange不需要routing key
		payload,
	)
}

// Publish 实现 Publisher 接口
func (p *NotificationPublisher) Publish(ctx context.Context, exchange, routingKey string, message []byte) error {
	return p.producer.Publish(ctx, exchange, routingKey, message)
}

// Close 关闭发布者
func (p *NotificationPublisher) Close() error {
	return p.producer.Close()
}

// Ensure NotificationPublisher implements the mq.NotificationPublisher interface
var _ mq.NotificationPublisher = (*NotificationPublisher)(nil)

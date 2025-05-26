package rabbitmq

import (
	"context"
	"fmt"
	"time"

	"github.com/blackarbiter/go-sac/pkg/mq"
	amqp "github.com/rabbitmq/amqp091-go"
)

// ResultPublisher 实现结果发布接口
type ResultPublisher struct {
	conn     *amqp.Connection
	producer *EnhancedProducer
}

// NewResultPublisher 创建结果发布者实例
func NewResultPublisher(conn *amqp.Connection) (*ResultPublisher, error) {
	// 检查exchange是否存在
	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}
	defer channel.Close()

	// 确认当前exchange存在
	err = channel.ExchangeDeclarePassive(
		ResultProcessExchange,
		"direct", // exchange类型
		true,     // durable
		false,    // auto-delete
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("exchange %s not found, please run setup first: %w", ResultProcessExchange, err)
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

	return &ResultPublisher{
		conn:     conn,
		producer: producer,
	}, nil
}

// PublishScanResult 发布扫描结果
func (p *ResultPublisher) PublishScanResult(ctx context.Context, result []byte) error {
	return p.producer.Publish(
		ctx,
		ResultProcessExchange,
		ResultStoragePattern,
		result,
	)
}

// Publish 实现 Publisher 接口
func (p *ResultPublisher) Publish(ctx context.Context, exchange, routingKey string, message []byte) error {
	return p.producer.Publish(ctx, exchange, routingKey, message)
}

// Close 关闭发布者
func (p *ResultPublisher) Close() error {
	return p.producer.Close()
}

// Ensure ResultPublisher implements the mq.ResultPublisher interface
var _ mq.ResultPublisher = (*ResultPublisher)(nil)

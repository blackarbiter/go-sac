package rabbitmq

import (
	"context"
	"fmt"
	"log"

	"github.com/blackarbiter/go-sac/pkg/mq"
	amqp "github.com/rabbitmq/amqp091-go"
)

// ScanConsumer implements a specialized consumer for scan tasks
type ScanConsumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	done    chan struct{}
}

// NewScanConsumer creates a new instance of ScanConsumer
func NewScanConsumer(conn *amqp.Connection) (*ScanConsumer, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// 检查队列是否存在
	for _, queueName := range []string{
		ScanHighPriorityQueue,
		ScanMediumPriorityQueue,
		ScanLowPriorityQueue,
	} {
		_, err := channel.QueueInspect(queueName)
		if err != nil {
			channel.Close()
			return nil, fmt.Errorf("queue %s not found, please run setup first: %w", queueName, err)
		}
	}

	return &ScanConsumer{
		conn:    conn,
		channel: channel,
		done:    make(chan struct{}),
	}, nil
}

// ConsumeHighPriority consumes messages from the high priority scan queue
func (c *ScanConsumer) ConsumeHighPriority(ctx context.Context, handler mq.MessageHandler) error {
	return c.consume(ctx, ScanHighPriorityQueue, 10, handler)
}

// ConsumeMediumPriority consumes messages from the medium priority scan queue
func (c *ScanConsumer) ConsumeMediumPriority(ctx context.Context, handler mq.MessageHandler) error {
	return c.consume(ctx, ScanMediumPriorityQueue, 5, handler)
}

// ConsumeLowPriority consumes messages from the low priority scan queue
func (c *ScanConsumer) ConsumeLowPriority(ctx context.Context, handler mq.MessageHandler) error {
	return c.consume(ctx, ScanLowPriorityQueue, 3, handler)
}

// Consume consumes messages from the specified queue
func (c *ScanConsumer) Consume(ctx context.Context, queueName string, handler mq.MessageHandler) error {
	// Default prefetch for regular consumption
	return c.consume(ctx, queueName, 1, handler)
}

// consume is the internal method that handles actual message consumption
func (c *ScanConsumer) consume(ctx context.Context, queueName string, prefetchCount int, handler mq.MessageHandler) error {
	// Set QoS/prefetch count
	if err := c.channel.Qos(prefetchCount, 0, false); err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	deliveries, err := c.channel.Consume(
		queueName,
		"",    // consumer tag - auto generated
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		for {
			select {
			case <-c.done:
				return
			case <-ctx.Done():
				return
			case delivery, ok := <-deliveries:
				if !ok {
					log.Printf("Consumer channel closed")
					return
				}

				// Process the message
				err := handler.HandleMessage(ctx, delivery.Body)
				if err != nil {
					log.Printf("Error processing message: %v", err)
					// 不再重新入队，直接拒绝消息，消息将进入死信队列
					err := delivery.Nack(false, false)
					if err != nil {
						log.Printf("Delivery to dead letter error...")
						return
					} // 第二个参数设为false，表示不重新入队
				} else {
					// Ack message when processed successfully
					err := delivery.Ack(false)
					if err != nil {
						return
					}
				}
			}
		}
	}()

	return nil
}

// Close closes the consumer
func (c *ScanConsumer) Close() error {
	// Signal the consumer goroutine to stop
	close(c.done)

	// Close the channel and connection
	if err := c.channel.Close(); err != nil {
		return fmt.Errorf("failed to close channel: %w", err)
	}

	return nil
}

// Republish 重新发布消息到指定队列
// scanType: 扫描类型（如"vulnerability", "port", "discovery", "retry"等）
// priority: 优先级（0-低, 1-中, 2-高）
// message: 消息内容
func (c *ScanConsumer) Republish(ctx context.Context, scanType string, priority int, message []byte) error {
	// 确定路由键
	routingKey := fmt.Sprintf("scan.%s", scanType)

	// 根据优先级确定路由模式后缀
	var prioritySuffix string
	switch priority {
	case 2:
		prioritySuffix = "high"
	case 1:
		prioritySuffix = "medium"
	case 0:
		prioritySuffix = "low"
	default:
		prioritySuffix = "medium" // 默认使用中优先级
	}

	// 完整的路由键
	routingKey = fmt.Sprintf("scan.%s.%s", scanType, prioritySuffix)

	// 设置消息属性
	headers := amqp.Table{}
	if priority >= 0 && priority <= 9 {
		headers["x-priority"] = int32(priority)
	}

	// 发布消息
	err := c.channel.PublishWithContext(
		ctx,
		TaskDispatchExchange, // 使用正确的交换机名称
		routingKey,           // 路由键
		true,                 // mandatory
		false,                // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // 持久化消息
			Priority:     uint8(priority),
			Headers:      headers,
			Body:         message,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to republish message: %w", err)
	}

	return nil
}

// Ensure ScanConsumer implements the mq.Consumer interface
var _ mq.Consumer = (*ScanConsumer)(nil)

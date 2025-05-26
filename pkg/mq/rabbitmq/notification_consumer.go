package rabbitmq

import (
	"context"
	"fmt"
	"log"

	"github.com/blackarbiter/go-sac/pkg/mq"
	amqp "github.com/rabbitmq/amqp091-go"
)

// NotificationConsumer 实现通知消费者
type NotificationConsumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	done    chan struct{}
}

// NewNotificationConsumer 创建通知消费者实例
func NewNotificationConsumer(conn *amqp.Connection) (*NotificationConsumer, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// 检查队列是否存在
	for _, queueName := range []string{
		NotificationEmailQueue,
		NotificationSMSQueue,
		NotificationSystemQueue,
	} {
		_, err := channel.QueueInspect(queueName)
		if err != nil {
			channel.Close()
			return nil, fmt.Errorf("queue %s not found, please run setup first: %w", queueName, err)
		}
	}

	return &NotificationConsumer{
		conn:    conn,
		channel: channel,
		done:    make(chan struct{}),
	}, nil
}

// ConsumeEmail 消费邮件通知
func (c *NotificationConsumer) ConsumeEmail(ctx context.Context, handler mq.MessageHandler) error {
	return c.consume(ctx, NotificationEmailQueue, handler)
}

// ConsumeSMS 消费短信通知
func (c *NotificationConsumer) ConsumeSMS(ctx context.Context, handler mq.MessageHandler) error {
	return c.consume(ctx, NotificationSMSQueue, handler)
}

// ConsumeSystem 消费系统通知
func (c *NotificationConsumer) ConsumeSystem(ctx context.Context, handler mq.MessageHandler) error {
	return c.consume(ctx, NotificationSystemQueue, handler)
}

// Consume 实现 Consumer 接口
func (c *NotificationConsumer) Consume(ctx context.Context, queueName string, handler mq.MessageHandler) error {
	return c.consume(ctx, queueName, handler)
}

// consume 内部消费方法
func (c *NotificationConsumer) consume(ctx context.Context, queueName string, handler mq.MessageHandler) error {
	// 设置QoS
	if err := c.channel.Qos(1, 0, false); err != nil {
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

				// 处理消息
				err := handler.HandleMessage(ctx, delivery.Body)
				if err != nil {
					log.Printf("Error processing message: %v", err)
					// 不再重新入队，直接拒绝消息，消息将进入死信队列
					err := delivery.Nack(false, false)
					if err != nil {
						log.Printf("Delivery to dead letter error...")
						return
					}
				} else {
					// 成功处理消息后确认
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

// Close 关闭消费者
func (c *NotificationConsumer) Close() error {
	// 通知消费者goroutine停止
	close(c.done)

	// 关闭channel和连接
	if err := c.channel.Close(); err != nil {
		return fmt.Errorf("failed to close channel: %w", err)
	}

	return nil
}

// Ensure NotificationConsumer implements the mq.Consumer interface
var _ mq.Consumer = (*NotificationConsumer)(nil)

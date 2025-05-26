package rabbitmq

import (
	"context"
	"fmt"
	"log"

	"github.com/blackarbiter/go-sac/pkg/mq"
	amqp "github.com/rabbitmq/amqp091-go"
)

// ResultConsumer 实现结果存储消费者
type ResultConsumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	done    chan struct{}
}

// NewResultConsumer 创建结果存储消费者实例
func NewResultConsumer(conn *amqp.Connection) (*ResultConsumer, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// 检查队列是否存在
	_, err = channel.QueueInspect(ResultStorageQueue)
	if err != nil {
		channel.Close()
		return nil, fmt.Errorf("queue %s not found, please run setup first: %w", ResultStorageQueue, err)
	}

	return &ResultConsumer{
		conn:    conn,
		channel: channel,
		done:    make(chan struct{}),
	}, nil
}

// Consume 开始消费结果消息
func (c *ResultConsumer) Consume(ctx context.Context, queueName string, handler mq.MessageHandler) error {
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
func (c *ResultConsumer) Close() error {
	// 通知消费者goroutine停止
	close(c.done)

	// 关闭channel和连接
	if err := c.channel.Close(); err != nil {
		return fmt.Errorf("failed to close channel: %w", err)
	}

	return nil
}

// Ensure ResultConsumer implements the mq.Consumer interface
var _ mq.Consumer = (*ResultConsumer)(nil)

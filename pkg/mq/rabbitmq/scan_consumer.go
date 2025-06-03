package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/blackarbiter/go-sac/pkg/logger"
	"github.com/blackarbiter/go-sac/pkg/service"

	"github.com/blackarbiter/go-sac/pkg/mq"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// ScanConsumer implements a specialized consumer for scan tasks
type ScanConsumer struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	done      chan struct{}
	scheduler *service.PriorityScheduler
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
					logger.Logger.Warn("Consumer channel closed")
					return
				}

				// Process the message
				err := handler.HandleMessage(ctx, delivery.Body)
				if err != nil {
					logger.Logger.Error("Error processing message: ", zap.Error(err))
					// 不再重新入队，直接拒绝消息，消息将进入死信队列
					err := delivery.Nack(false, false)
					if err != nil {
						logger.Logger.Error("Delivery to dead letter error: ", zap.Error(err))
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

// ConsumeToScheduler 新增方法：将消息投递到调度器通道
func (c *ScanConsumer) ConsumeToScheduler(ctx context.Context, scheduler *service.PriorityScheduler) error {
	if c.scheduler == nil {
		return errors.New("scheduler not set")
	}

	const backoffInterval = 5 * time.Second
	backoffTicker := time.NewTicker(backoffInterval)
	defer backoffTicker.Stop()

	// 创建消费者
	createConsumer := func(queue string) (<-chan amqp.Delivery, error) {
		return c.channel.Consume(
			queue,
			"",    // consumer tag - auto generated
			false, // auto-ack
			false, // exclusive
			false, // no-local
			false, // no-wait
			nil,   // args
		)
	}

	// 初始化消费者
	highMsgs, err := createConsumer(ScanHighPriorityQueue)
	if err != nil {
		return fmt.Errorf("failed to consume high queue: %w", err)
	}
	medMsgs, err := createConsumer(ScanMediumPriorityQueue)
	if err != nil {
		return fmt.Errorf("failed to consume medium queue: %w", err)
	}
	lowMsgs, err := createConsumer(ScanLowPriorityQueue)
	if err != nil {
		return fmt.Errorf("failed to consume low queue: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil

		default:
			// 全系统背压检查
			if scheduler.State.ShouldStopConsuming() {
				// 背压状态下：仅当优先级通道未满时才消费
				delivered := false
				// 尝试填充高优先级通道（非阻塞）
				if len(scheduler.HighPriorityChan) < cap(scheduler.HighPriorityChan) {
					select {
					case delivery := <-highMsgs:
						scheduler.HighPriorityChan <- delivery
						delivered = true
					default:
					}
				}

				// 尝试填充中优先级通道（非阻塞）
				if !delivered && len(scheduler.MedPriorityChan) < cap(scheduler.MedPriorityChan) {
					select {
					case delivery := <-medMsgs:
						scheduler.MedPriorityChan <- delivery
						delivered = true
					default:
					}
				}

				// 尝试填充低优先级通道（非阻塞）
				if !delivered && len(scheduler.LowPriorityChan) < cap(scheduler.LowPriorityChan) {
					select {
					case delivery := <-lowMsgs:
						scheduler.LowPriorityChan <- delivery
						delivered = true
					default:
					}
				}

				// 如果所有通道都满或没有消息，则进入背压等待
				if !delivered {
					select {
					case <-backoffTicker.C:
						logger.Logger.Warn("System in backpressure, suspending consumption")
					case <-ctx.Done():
						return nil
					}
				}
			} else {
				// 正常状态：非阻塞消费逻辑
				delivered := false
				select {
				case delivery := <-highMsgs:
					scheduler.HighPriorityChan <- delivery
					delivered = true
				case delivery := <-medMsgs:
					scheduler.MedPriorityChan <- delivery
					delivered = true
				case delivery := <-lowMsgs:
					scheduler.LowPriorityChan <- delivery
					delivered = true
				default:
				}

				// 降低空转CPU消耗
				if !delivered {
					time.Sleep(300 * time.Millisecond)
				}
			}
		}
	}
}

func (c *ScanConsumer) SetScheduler(scheduler *service.PriorityScheduler) {
	c.scheduler = scheduler
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

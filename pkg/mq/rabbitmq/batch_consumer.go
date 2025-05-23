package rabbitmq

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/blackarbiter/go-sac/pkg/metrics"
	"github.com/blackarbiter/go-sac/pkg/mq"
	amqp "github.com/rabbitmq/amqp091-go"
)

// BatchMessageHandler 批量消息处理接口
type BatchMessageHandler interface {
	// HandleBatch 处理一批消息
	HandleBatch(ctx context.Context, messages [][]byte) error
}

// BatchMessageHandlerWithResults 增强的批量消息处理接口，支持返回每个消息的处理结果
type BatchMessageHandlerWithResults interface {
	// HandleBatchWithResults 处理一批消息，返回每个消息的处理结果
	// 返回的错误数组与输入消息数组一一对应，nil表示成功，非nil表示失败
	HandleBatchWithResults(ctx context.Context, messages [][]byte) []error
}

// BatchConsumer 实现批量消费消息的消费者
type BatchConsumer struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	batchSize    int
	batchTimeout time.Duration
	done         chan struct{}
}

// NewBatchConsumer 创建新的批量消费者
func NewBatchConsumer(conn *amqp.Connection, batchSize int, batchTimeout time.Duration) (*BatchConsumer, error) {
	if batchSize <= 0 {
		batchSize = 100 // 默认批量大小
	}

	if batchTimeout <= 0 {
		batchTimeout = 5 * time.Second // 默认批量超时
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &BatchConsumer{
		conn:         conn,
		channel:      channel,
		batchSize:    batchSize,
		batchTimeout: batchTimeout,
		done:         make(chan struct{}),
	}, nil
}

// ConsumeBatch 批量消费消息
func (c *BatchConsumer) ConsumeBatch(ctx context.Context, queueName string, handler BatchMessageHandler) error {
	// 设置预取数量为批量大小的2倍，以确保有足够的消息可用
	if err := c.channel.Qos(c.batchSize*2, 0, false); err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	deliveries, err := c.channel.Consume(
		queueName,
		"",    // consumer tag
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to consume from queue %s: %w", queueName, err)
	}

	go c.processBatches(ctx, deliveries, handler, queueName)
	return nil
}

// ConsumeBatchWithResults 批量消费消息，支持单条消息确认/拒绝
func (c *BatchConsumer) ConsumeBatchWithResults(ctx context.Context, queueName string, handler BatchMessageHandlerWithResults) error {
	// 设置预取数量为批量大小的2倍，以确保有足够的消息可用
	if err := c.channel.Qos(c.batchSize*2, 0, false); err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	deliveries, err := c.channel.Consume(
		queueName,
		"",    // consumer tag
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to consume from queue %s: %w", queueName, err)
	}

	go c.processBatchesWithResults(ctx, deliveries, handler, queueName)
	return nil
}

// 批量处理消息
func (c *BatchConsumer) processBatches(ctx context.Context, deliveries <-chan amqp.Delivery, handler BatchMessageHandler, queueName string) {
	var (
		batch     = make([]amqp.Delivery, 0, c.batchSize)
		batchData = make([][]byte, 0, c.batchSize)
		ticker    = time.NewTicker(c.batchTimeout)
		mu        sync.Mutex
	)
	defer ticker.Stop()

	// 处理当前批次
	processBatch := func() {
		mu.Lock()
		defer mu.Unlock()

		if len(batch) == 0 {
			return
		}

		// 记录批次大小
		batchSize := len(batch)

		// 准备消息数据
		batchData = batchData[:0] // 清空但保留容量
		for _, delivery := range batch {
			batchData = append(batchData, delivery.Body)
		}

		// 记录开始处理时间
		startTime := time.Now()

		// 处理批次
		err := handler.HandleBatch(ctx, batchData)

		// 计算处理时间
		duration := time.Since(startTime).Seconds()

		// 记录处理时间指标
		status := "success"
		if err != nil {
			status = "failure"
		}
		metrics.MessageProcessingDuration.WithLabelValues(queueName, status).Observe(duration)

		// 根据处理结果确认或拒绝消息
		if err != nil {
			log.Printf("[BatchConsumer] Failed to process batch of %d messages from queue %s: %v",
				batchSize, queueName, err)

			// 记录指标
			metrics.BatchProcessingCounter.WithLabelValues(queueName, "failure").Inc()
			metrics.BatchMessageCounter.WithLabelValues(queueName, "failure").Add(float64(batchSize))

			// 拒绝整个批次
			for _, delivery := range batch {
				delivery.Nack(false, true) // 单个拒绝并重新入队
			}
		} else {
			log.Printf("[BatchConsumer] Successfully processed batch of %d messages from queue %s",
				batchSize, queueName)

			// 记录指标
			metrics.BatchProcessingCounter.WithLabelValues(queueName, "success").Inc()
			metrics.BatchMessageCounter.WithLabelValues(queueName, "success").Add(float64(batchSize))

			// 确认整个批次
			for _, delivery := range batch {
				delivery.Ack(false) // 单个确认
			}
		}

		// 清空批次
		batch = batch[:0] // 清空但保留容量
	}

	for {
		select {
		case <-c.done:
			// 处理剩余的批次
			processBatch()
			return

		case <-ctx.Done():
			// 上下文取消，处理剩余的批次
			processBatch()
			return

		case <-ticker.C:
			// 超时，处理当前批次
			processBatch()

		case delivery, ok := <-deliveries:
			if !ok {
				// 通道关闭
				processBatch()
				return
			}

			mu.Lock()
			batch = append(batch, delivery)

			// 如果达到批次大小，立即处理
			if len(batch) >= c.batchSize {
				mu.Unlock()
				processBatch()
				ticker.Reset(c.batchTimeout) // 重置定时器
			} else {
				mu.Unlock()
			}
		}
	}
}

// 批量处理消息，支持单条消息确认/拒绝
func (c *BatchConsumer) processBatchesWithResults(ctx context.Context, deliveries <-chan amqp.Delivery, handler BatchMessageHandlerWithResults, queueName string) {
	var (
		batch     = make([]amqp.Delivery, 0, c.batchSize)
		batchData = make([][]byte, 0, c.batchSize)
		ticker    = time.NewTicker(c.batchTimeout)
		mu        sync.Mutex
	)
	defer ticker.Stop()

	// 处理当前批次
	processBatch := func() {
		mu.Lock()
		defer mu.Unlock()

		if len(batch) == 0 {
			return
		}

		// 记录批次大小
		batchSize := len(batch)

		// 准备消息数据
		batchData = batchData[:0] // 清空但保留容量
		for _, delivery := range batch {
			batchData = append(batchData, delivery.Body)
		}

		// 记录开始处理时间
		startTime := time.Now()

		// 处理批次，获取每条消息的结果
		results := handler.HandleBatchWithResults(ctx, batchData)

		// 计算处理时间
		duration := time.Since(startTime).Seconds()

		// 记录处理时间指标
		metrics.MessageProcessingDuration.WithLabelValues(queueName, "batch").Observe(duration)

		// 统计成功和失败数量
		successCount := 0
		failureCount := 0

		// 根据每条消息的处理结果确认或拒绝
		for i, err := range results {
			if i >= len(batch) {
				// 防止索引越界
				break
			}

			if err != nil {
				// 处理失败，拒绝并重新入队
				log.Printf("[BatchConsumer] Failed to process message %d in batch from queue %s: %v",
					i, queueName, err)
				batch[i].Nack(false, true)
				failureCount++
			} else {
				// 处理成功，确认消息
				batch[i].Ack(false)
				successCount++
			}
		}

		// 记录指标
		metrics.BatchProcessingCounter.WithLabelValues(queueName, "mixed").Inc()
		metrics.BatchMessageCounter.WithLabelValues(queueName, "success").Add(float64(successCount))
		metrics.BatchMessageCounter.WithLabelValues(queueName, "failure").Add(float64(failureCount))

		log.Printf("[BatchConsumer] Processed batch of %d messages from queue %s: %d succeeded, %d failed",
			batchSize, queueName, successCount, failureCount)

		// 清空批次
		batch = batch[:0] // 清空但保留容量
	}

	for {
		select {
		case <-c.done:
			// 处理剩余的批次
			processBatch()
			return

		case <-ctx.Done():
			// 上下文取消，处理剩余的批次
			processBatch()
			return

		case <-ticker.C:
			// 超时，处理当前批次
			processBatch()

		case delivery, ok := <-deliveries:
			if !ok {
				// 通道关闭
				processBatch()
				return
			}

			mu.Lock()
			batch = append(batch, delivery)

			// 如果达到批次大小，立即处理
			if len(batch) >= c.batchSize {
				mu.Unlock()
				processBatch()
				ticker.Reset(c.batchTimeout) // 重置定时器
			} else {
				mu.Unlock()
			}
		}
	}
}

// Close 关闭批量消费者
func (c *BatchConsumer) Close() error {
	close(c.done)
	return c.channel.Close()
}

// MessageHandlerAdapter 将普通MessageHandler适配为BatchMessageHandler
type MessageHandlerAdapter struct {
	handler mq.MessageHandler
}

// NewMessageHandlerAdapter 创建消息处理器适配器
func NewMessageHandlerAdapter(handler mq.MessageHandler) *MessageHandlerAdapter {
	return &MessageHandlerAdapter{handler: handler}
}

// HandleBatch 实现BatchMessageHandler接口，将批处理拆分为单个处理
func (a *MessageHandlerAdapter) HandleBatch(ctx context.Context, messages [][]byte) error {
	for _, msg := range messages {
		if err := a.handler.HandleMessage(ctx, msg); err != nil {
			return err
		}
	}
	return nil
}

// MessageHandlerAdapterWithResults 将普通MessageHandler适配为BatchMessageHandlerWithResults
type MessageHandlerAdapterWithResults struct {
	handler mq.MessageHandler
}

// NewMessageHandlerAdapterWithResults 创建支持单条结果的消息处理器适配器
func NewMessageHandlerAdapterWithResults(handler mq.MessageHandler) *MessageHandlerAdapterWithResults {
	return &MessageHandlerAdapterWithResults{handler: handler}
}

// HandleBatchWithResults 实现BatchMessageHandlerWithResults接口，将批处理拆分为单个处理
func (a *MessageHandlerAdapterWithResults) HandleBatchWithResults(ctx context.Context, messages [][]byte) []error {
	results := make([]error, len(messages))
	for i, msg := range messages {
		results[i] = a.handler.HandleMessage(ctx, msg)
	}
	return results
}

// StandardConsumer 适配BatchConsumer到标准Consumer接口
type StandardConsumer struct {
	batchConsumer *BatchConsumer
	withResults   bool // 是否使用单条消息确认/拒绝模式
}

// NewStandardConsumer 创建标准消费者适配器
func NewStandardConsumer(batchConsumer *BatchConsumer, withResults bool) *StandardConsumer {
	return &StandardConsumer{
		batchConsumer: batchConsumer,
		withResults:   withResults,
	}
}

// Consume 实现Consumer接口
func (c *StandardConsumer) Consume(ctx context.Context, queueName string, handler mq.MessageHandler) error {
	if c.withResults {
		// 使用支持单条消息确认/拒绝的模式
		adapter := NewMessageHandlerAdapterWithResults(handler)
		return c.batchConsumer.ConsumeBatchWithResults(ctx, queueName, adapter)
	} else {
		// 使用传统的批量确认/拒绝模式
		adapter := NewMessageHandlerAdapter(handler)
		return c.batchConsumer.ConsumeBatch(ctx, queueName, adapter)
	}
}

// Close 关闭消费者
func (c *StandardConsumer) Close() error {
	return c.batchConsumer.Close()
}

// 确保StandardConsumer实现mq.Consumer接口
var _ mq.Consumer = (*StandardConsumer)(nil)

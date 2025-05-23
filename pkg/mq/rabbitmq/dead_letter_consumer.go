package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/blackarbiter/go-sac/pkg/metrics"
	"github.com/blackarbiter/go-sac/pkg/mq"
	amqp "github.com/rabbitmq/amqp091-go"
)

// DeadLetterConsumer 专门处理进入死信队列的消息
type DeadLetterConsumer struct {
	conn       *amqp.Connection
	channel    *amqp.Channel
	done       chan struct{}
	retryCount int // 最大重试次数
}

// RetryInfo 用于记录消息重试信息
type RetryInfo struct {
	Count      int       `json:"count"`
	LastRetry  time.Time `json:"last_retry"`
	OriginalEx string    `json:"original_exchange"`
	OriginalRK string    `json:"original_routing_key"`
}

// NewDeadLetterConsumer 创建一个新的死信消费者
func NewDeadLetterConsumer(conn *amqp.Connection, maxRetries int) (*DeadLetterConsumer, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// 检查死信队列是否存
	_, err = channel.QueueInspect(RetryQueue5Min)
	if err != nil {
		channel.Close()
		return nil, fmt.Errorf("retry queue %s not found, please run setup first: %w", RetryQueue5Min, err)
	}

	return &DeadLetterConsumer{
		conn:       conn,
		channel:    channel,
		done:       make(chan struct{}),
		retryCount: maxRetries,
	}, nil
}

// Start 开始消费死信队列
func (c *DeadLetterConsumer) Start(ctx context.Context) error {
	// 设置QoS
	if err := c.channel.Qos(1, 0, false); err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	deliveries, err := c.channel.Consume(
		RetryQueue5Min,
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

	go c.processDeadLetters(ctx, deliveries)
	return nil
}

// 处理死信队列中的消息
func (c *DeadLetterConsumer) processDeadLetters(ctx context.Context, deliveries <-chan amqp.Delivery) {
	for {
		select {
		case <-c.done:
			return
		case <-ctx.Done():
			return
		case delivery, ok := <-deliveries:
			if !ok {
				log.Printf("Dead letter channel closed")
				return
			}

			c.handleDeadLetter(ctx, delivery, RetryQueue5Min)
		}
	}
}

// 处理单个死信消息
func (c *DeadLetterConsumer) handleDeadLetter(ctx context.Context, delivery amqp.Delivery, queueName string) {
	// 提取消息ID
	messageID := delivery.MessageId
	if messageID == "" {
		// 如果没有消息ID，使用内容的哈希作为ID
		messageID = fmt.Sprintf("%x", delivery.Body[:min(16, len(delivery.Body))])
	}

	// 记录结构化日志
	log.Printf("[DeadLetter] Received message - Queue: %s, MessageID: %s, Body: %s",
		queueName, messageID, string(delivery.Body))

	// 1. 提取原始的交换机和路由键
	xDeath, ok := getXDeathHeader(delivery)
	if !ok {
		reason := "missing_x_death_header"
		log.Printf("[DeadLetter] ERROR: Cannot process dead letter without x-death header - MessageID: %s, Queue: %s",
			messageID, queueName)

		// 记录指标
		metrics.DeadLetterCounter.WithLabelValues(queueName, "0", reason).Inc()

		c.sendToManualIntervention(ctx, delivery, reason)
		delivery.Ack(false)
		return
	}

	// 2. 提取或初始化重试信息
	retryInfo := getRetryInfo(delivery)
	retryCountStr := fmt.Sprintf("%d", retryInfo.Count)

	// 检查是否超过重试次数
	if retryInfo.Count >= c.retryCount {
		reason := "max_retry_exceeded"
		log.Printf("[DeadLetter] WARN: Message exceeded retry limit (%d/%d) - MessageID: %s, Queue: %s",
			retryInfo.Count, c.retryCount, messageID, queueName)

		// 记录指标
		metrics.DeadLetterCounter.WithLabelValues(queueName, retryCountStr, reason).Inc()

		c.sendToManualIntervention(ctx, delivery, reason)
		delivery.Ack(false)
		return
	}

	// 3. 更新重试信息
	retryInfo.Count++
	retryInfo.LastRetry = time.Now()
	newRetryCountStr := fmt.Sprintf("%d", retryInfo.Count)

	// 尝试从x-death获取原始交换机和路由键
	if retryInfo.OriginalEx == "" && len(xDeath) > 0 {
		if exchange, ok := xDeath[0]["exchange"].(string); ok {
			retryInfo.OriginalEx = exchange
		}
		if routingKeys, ok := xDeath[0]["routing-keys"].([]interface{}); ok && len(routingKeys) > 0 {
			if rk, ok := routingKeys[0].(string); ok {
				retryInfo.OriginalRK = rk
			}
		}
	}

	// 4. 重新发布消息到原始交换机
	if retryInfo.OriginalEx == "" || retryInfo.OriginalRK == "" {
		reason := "missing_routing_info"
		log.Printf("[DeadLetter] ERROR: Cannot determine original exchange/routing key - MessageID: %s, Queue: %s",
			messageID, queueName)

		// 记录指标
		metrics.DeadLetterCounter.WithLabelValues(queueName, retryCountStr, reason).Inc()

		c.sendToManualIntervention(ctx, delivery, reason)
		delivery.Ack(false)
		return
	}

	// 5. 准备重试发布
	retryInfoBytes, _ := json.Marshal(retryInfo)
	headers := amqp.Table{}
	if delivery.Headers != nil {
		headers = delivery.Headers
	}
	headers["x-retry-info"] = string(retryInfoBytes)

	err := c.channel.PublishWithContext(
		ctx,
		retryInfo.OriginalEx,
		retryInfo.OriginalRK,
		true,  // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:     delivery.ContentType,
			ContentEncoding: delivery.ContentEncoding,
			DeliveryMode:    delivery.DeliveryMode,
			Priority:        delivery.Priority,
			CorrelationId:   delivery.CorrelationId,
			ReplyTo:         delivery.ReplyTo,
			Expiration:      delivery.Expiration,
			MessageId:       delivery.MessageId,
			Timestamp:       delivery.Timestamp,
			Type:            delivery.Type,
			UserId:          delivery.UserId,
			AppId:           delivery.AppId,
			Body:            delivery.Body,
			Headers:         headers,
		},
	)

	if err != nil {
		reason := "republish_failed"
		log.Printf("[DeadLetter] ERROR: Failed to republish message - MessageID: %s, Queue: %s, Error: %v",
			messageID, queueName, err)

		// 记录指标
		metrics.DeadLetterCounter.WithLabelValues(queueName, retryCountStr, reason).Inc()

		// 重新入队等待下次处理
		delivery.Nack(false, true)
		return
	}

	// 记录成功重试
	log.Printf("[DeadLetter] Successfully republished message for retry - MessageID: %s, Queue: %s, Attempt: %d/%d",
		messageID, queueName, retryInfo.Count, c.retryCount)

	// 记录指标
	metrics.DeadLetterCounter.WithLabelValues(queueName, newRetryCountStr, "retry").Inc()

	delivery.Ack(false)
}

// 将消息转发到人工干预队列
func (c *DeadLetterConsumer) sendToManualIntervention(ctx context.Context, delivery amqp.Delivery, reason string) {
	// 提取消息ID
	messageID := delivery.MessageId
	if messageID == "" {
		messageID = fmt.Sprintf("%x", delivery.Body[:min(16, len(delivery.Body))])
	}

	// 添加失败原因到头部
	headers := amqp.Table{}
	if delivery.Headers != nil {
		for k, v := range delivery.Headers {
			headers[k] = v
		}
	}
	headers["x-failure-reason"] = reason

	err := c.channel.PublishWithContext(
		ctx,
		RetryExchange,
		"manual.intervention",
		true,  // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:     delivery.ContentType,
			ContentEncoding: delivery.ContentEncoding,
			DeliveryMode:    delivery.DeliveryMode,
			Priority:        delivery.Priority,
			CorrelationId:   delivery.CorrelationId,
			ReplyTo:         delivery.ReplyTo,
			Expiration:      delivery.Expiration,
			MessageId:       messageID,
			Timestamp:       delivery.Timestamp,
			Type:            delivery.Type,
			UserId:          delivery.UserId,
			AppId:           delivery.AppId,
			Body:            delivery.Body,
			Headers:         headers,
		},
	)
	if err != nil {
		log.Printf("[DeadLetter] ERROR: Failed to send message to manual intervention - MessageID: %s, Error: %v",
			messageID, err)
	} else {
		log.Printf("[DeadLetter] Sent message to manual intervention - MessageID: %s, Reason: %s",
			messageID, reason)
	}
}

// 从消息中提取重试信息
func getRetryInfo(delivery amqp.Delivery) RetryInfo {
	retryInfo := RetryInfo{}

	if delivery.Headers == nil {
		return retryInfo
	}

	if retryInfoStr, ok := delivery.Headers["x-retry-info"].(string); ok {
		if err := json.Unmarshal([]byte(retryInfoStr), &retryInfo); err != nil {
			log.Printf("Failed to unmarshal retry info: %v", err)
		}
	}

	return retryInfo
}

// 从消息中提取x-death头部
func getXDeathHeader(delivery amqp.Delivery) ([]amqp.Table, bool) {
	if delivery.Headers == nil {
		return nil, false
	}

	xDeath, ok := delivery.Headers["x-death"].([]interface{})
	if !ok || len(xDeath) == 0 {
		return nil, false
	}

	result := make([]amqp.Table, 0, len(xDeath))
	for _, item := range xDeath {
		if table, ok := item.(amqp.Table); ok {
			result = append(result, table)
		}
	}

	return result, len(result) > 0
}

// Close 关闭消费者
func (c *DeadLetterConsumer) Close() error {
	close(c.done)
	return c.channel.Close()
}

// Consume 实现mq.Consumer接口
func (c *DeadLetterConsumer) Consume(ctx context.Context, queueName string, handler mq.MessageHandler) error {
	// 对于死信消费者，我们有特殊的处理逻辑，但为了满足接口要求，提供此方法
	// 创建一个适配器，将我们的特殊处理逻辑封装为标准的MessageHandler
	adaptedHandler := func(delivery amqp.Delivery) {
		// 提取消息ID
		messageID := delivery.MessageId
		if messageID == "" {
			messageID = fmt.Sprintf("%x", delivery.Body[:min(16, len(delivery.Body))])
		}

		// 记录开始处理
		startTime := time.Now()

		// 处理消息
		err := handler.HandleMessage(ctx, delivery.Body)

		// 计算处理时间
		duration := time.Since(startTime).Seconds()

		if err != nil {
			// 处理失败
			log.Printf("[DeadLetter] Handler error for message - Queue: %s, MessageID: %s, Error: %v",
				queueName, messageID, err)

			// 记录指标
			metrics.MessageProcessingCounter.WithLabelValues(queueName, "failure").Inc()
			metrics.MessageProcessingDuration.WithLabelValues(queueName, "failure").Observe(duration)

			delivery.Nack(false, true) // 重新入队
		} else {
			// 处理成功
			log.Printf("[DeadLetter] Successfully processed message - Queue: %s, MessageID: %s",
				queueName, messageID)

			// 记录指标
			metrics.MessageProcessingCounter.WithLabelValues(queueName, "success").Inc()
			metrics.MessageProcessingDuration.WithLabelValues(queueName, "success").Observe(duration)

			delivery.Ack(false)
		}
	}

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
					log.Printf("Dead letter channel closed")
					return
				}
				adaptedHandler(delivery)
			}
		}
	}()

	return nil
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Ensure DeadLetterConsumer implements the mq.Consumer interface
var _ mq.Consumer = (*DeadLetterConsumer)(nil)

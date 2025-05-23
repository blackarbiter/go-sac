package rabbitmq

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"sync"
	"time"

	"github.com/blackarbiter/go-sac/pkg/mq"
)

// IdempotentConsumer 提供幂等性保证的消费者
type IdempotentConsumer struct {
	consumer        mq.Consumer // 被装饰的消费者
	processedCache  *processedMessageCache
	processCallback ProcessCallback
}

// ProcessCallback 定义消息处理回调函数类型
type ProcessCallback func(ctx context.Context, message []byte, messageID string) error

// processedMessageCache 用于缓存已处理消息
type processedMessageCache struct {
	cache    map[string]time.Time
	mu       sync.RWMutex
	ttl      time.Duration
	maxItems int
}

// newProcessedMessageCache 创建新的消息缓存
func newProcessedMessageCache(ttl time.Duration, maxItems int) *processedMessageCache {
	cache := &processedMessageCache{
		cache:    make(map[string]time.Time),
		ttl:      ttl,
		maxItems: maxItems,
	}

	// 启动清理过期项的后台任务
	go cache.cleanupLoop()

	return cache
}

// cleanupLoop 定期清理过期缓存项
func (c *processedMessageCache) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

// cleanup 清理过期缓存项
func (c *processedMessageCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for id, timestamp := range c.cache {
		if now.Sub(timestamp) > c.ttl {
			delete(c.cache, id)
		}
	}
}

// add 添加处理过的消息到缓存
func (c *processedMessageCache) add(messageID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 如果缓存已满，删除最早的项
	if len(c.cache) >= c.maxItems {
		var oldestID string
		var oldestTime time.Time
		first := true

		for id, t := range c.cache {
			if first || t.Before(oldestTime) {
				oldestID = id
				oldestTime = t
				first = false
			}
		}

		if oldestID != "" {
			delete(c.cache, oldestID)
		}
	}

	c.cache[messageID] = time.Now()
}

// exists 检查消息是否已处理
func (c *processedMessageCache) exists(messageID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, exists := c.cache[messageID]
	return exists
}

// NewIdempotentConsumer 创建幂等性消费者装饰器
func NewIdempotentConsumer(consumer mq.Consumer, ttl time.Duration, maxCacheItems int) *IdempotentConsumer {
	if ttl <= 0 {
		ttl = 24 * time.Hour // 默认缓存1天
	}

	if maxCacheItems <= 0 {
		maxCacheItems = 10000 // 默认最多缓存10000条消息ID
	}

	return &IdempotentConsumer{
		consumer:       consumer,
		processedCache: newProcessedMessageCache(ttl, maxCacheItems),
	}
}

// 实现MessageHandler接口的适配器
type idempotentHandler struct {
	originalHandler mq.MessageHandler
	cache           *processedMessageCache
}

// HandleMessage 处理消息并保证幂等性
func (h *idempotentHandler) HandleMessage(ctx context.Context, message []byte) error {
	// 计算消息的唯一标识
	messageID := generateMessageID(message)

	// 检查是否已经处理过此消息
	if h.cache.exists(messageID) {
		log.Printf("Message %s already processed, skipping", messageID)
		return nil
	}

	// 处理消息
	err := h.originalHandler.HandleMessage(ctx, message)
	if err != nil {
		// 处理失败，不记录到缓存中
		return err
	}

	// 处理成功，记录到缓存
	h.cache.add(messageID)
	return nil
}

// Consume 消费消息并确保幂等性
func (c *IdempotentConsumer) Consume(ctx context.Context, queueName string, handler mq.MessageHandler) error {
	// 创建包装了幂等性逻辑的处理器
	idempotentHandler := &idempotentHandler{
		originalHandler: handler,
		cache:           c.processedCache,
	}

	// 使用被装饰的消费者进行实际消费
	return c.consumer.Consume(ctx, queueName, idempotentHandler)
}

// Close 关闭消费者
func (c *IdempotentConsumer) Close() error {
	return c.consumer.Close()
}

// 生成消息的唯一标识
func generateMessageID(message []byte) string {
	hash := sha256.Sum256(message)
	return hex.EncodeToString(hash[:])
}

// 确保IdempotentConsumer实现mq.Consumer接口
var _ mq.Consumer = (*IdempotentConsumer)(nil)

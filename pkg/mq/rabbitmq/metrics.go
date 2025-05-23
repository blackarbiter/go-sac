package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// MetricsCollector 用于收集RabbitMQ指标
type MetricsCollector struct {
	baseURL      string
	username     string
	password     string
	client       *http.Client
	metricsCache sync.Map
	interval     time.Duration
	done         chan struct{}
}

// QueueMetrics 表示队列的指标数据
type QueueMetrics struct {
	Name                string  `json:"name"`
	Messages            int     `json:"messages"`
	MessagesReady       int     `json:"messages_ready"`
	MessagesUnacked     int     `json:"messages_unacknowledged"`
	Consumers           int     `json:"consumers"`
	ConsumerUtilization float64 `json:"consumer_utilisation"`
}

// ExchangeMetrics 表示交换机的指标数据
type ExchangeMetrics struct {
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	MessageRate float64 `json:"message_stats.publish_details.rate"`
}

// NewMetricsCollector 创建新的指标收集器
func NewMetricsCollector(baseURL, username, password string, interval time.Duration) *MetricsCollector {
	if interval <= 0 {
		interval = 30 * time.Second // 默认30秒收集一次
	}

	return &MetricsCollector{
		baseURL:  baseURL,
		username: username,
		password: password,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		interval: interval,
		done:     make(chan struct{}),
	}
}

// Start 开始收集指标
func (c *MetricsCollector) Start(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	// 立即收集一次指标
	c.collectMetrics()

	go func() {
		for {
			select {
			case <-c.done:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				c.collectMetrics()
			}
		}
	}()
}

// Stop 停止收集指标
func (c *MetricsCollector) Stop() {
	close(c.done)
}

// GetQueueMetrics 获取队列的指标
func (c *MetricsCollector) GetQueueMetrics(queueName string) (QueueMetrics, bool) {
	value, ok := c.metricsCache.Load("queue:" + queueName)
	if !ok {
		return QueueMetrics{}, false
	}
	metrics, ok := value.(QueueMetrics)
	return metrics, ok
}

// GetExchangeMetrics 获取交换机的指标
func (c *MetricsCollector) GetExchangeMetrics(exchangeName string) (ExchangeMetrics, bool) {
	value, ok := c.metricsCache.Load("exchange:" + exchangeName)
	if !ok {
		return ExchangeMetrics{}, false
	}
	metrics, ok := value.(ExchangeMetrics)
	return metrics, ok
}

// 收集所有指标
func (c *MetricsCollector) collectMetrics() {
	// 收集队列指标
	c.collectQueueMetrics()

	// 收集交换机指标
	c.collectExchangeMetrics()
}

// 收集队列指标
func (c *MetricsCollector) collectQueueMetrics() {
	url := fmt.Sprintf("%s/api/queues", c.baseURL)

	var queues []QueueMetrics
	err := c.makeAPIRequest(url, &queues)
	if err != nil {
		log.Printf("Failed to collect queue metrics: %v", err)
		return
	}

	for _, queue := range queues {
		c.metricsCache.Store("queue:"+queue.Name, queue)
	}
}

// 收集交换机指标
func (c *MetricsCollector) collectExchangeMetrics() {
	url := fmt.Sprintf("%s/api/exchanges", c.baseURL)

	var exchanges []ExchangeMetrics
	err := c.makeAPIRequest(url, &exchanges)
	if err != nil {
		log.Printf("Failed to collect exchange metrics: %v", err)
		return
	}

	for _, exchange := range exchanges {
		c.metricsCache.Store("exchange:"+exchange.Name, exchange)
	}
}

// 执行API请求
func (c *MetricsCollector) makeAPIRequest(url string, result interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned non-OK status: %d, body: %s", resp.StatusCode, body)
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

// PrintQueueStatus 打印所有队列的状态信息
func (c *MetricsCollector) PrintQueueStatus() {
	c.metricsCache.Range(func(key, value interface{}) bool {
		keyStr, ok := key.(string)
		if !ok || len(keyStr) < 6 || keyStr[:6] != "queue:" {
			return true
		}

		metrics, ok := value.(QueueMetrics)
		if !ok {
			return true
		}

		log.Printf("Queue: %s, Messages: %d, Ready: %d, Unacked: %d, Consumers: %d",
			metrics.Name, metrics.Messages, metrics.MessagesReady, metrics.MessagesUnacked, metrics.Consumers)

		return true
	})
}

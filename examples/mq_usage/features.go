package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/blackarbiter/go-sac/pkg/mq/rabbitmq"
)

// 示例消息处理器
type ExampleMessageHandler struct {
	Name string
}

func (h *ExampleMessageHandler) HandleMessage(ctx context.Context, message []byte) error {
	log.Printf("[%s] Processing message: %s", h.Name, string(message))

	// 对包含"fail"的消息模拟处理失败
	if strings.Contains(string(message), "fail") {
		log.Printf("[%s] Failing message intentionally", h.Name)
		return errors.New("simulated failure for testing")
	}

	// 模拟处理时间
	time.Sleep(100 * time.Millisecond)
	return nil
}

// 批量消息处理器示例
type BatchHandler struct {
	Name string
}

func (h *BatchHandler) HandleBatch(ctx context.Context, messages [][]byte) error {
	log.Printf("[%s] Processing batch of %d messages", h.Name, len(messages))

	// 打印批次的前3条消息
	for i, msg := range messages {
		if i >= 3 {
			log.Printf("[%s] ... and %d more messages", h.Name, len(messages)-3)
			break
		}
		log.Printf("[%s] Batch item %d: %s", h.Name, i+1, string(msg))
	}

	// 模拟批量处理时间
	time.Sleep(200 * time.Millisecond)
	return nil
}

// 批量消息处理器示例，支持单条消息确认/拒绝
type BatchHandlerWithResults struct {
	Name string
}

func (h *BatchHandlerWithResults) HandleBatchWithResults(ctx context.Context, messages [][]byte) []error {
	log.Printf("[%s] Processing batch of %d messages with individual results", h.Name, len(messages))

	results := make([]error, len(messages))

	// 处理每条消息
	for i, msg := range messages {
		// 打印前3条消息
		if i < 3 {
			log.Printf("[%s] Batch item %d: %s", h.Name, i+1, string(msg))
		} else if i == 3 {
			log.Printf("[%s] ... and %d more messages", h.Name, len(messages)-3)
		}

		// 模拟处理时间
		time.Sleep(50 * time.Millisecond)

		// 对包含"fail"的消息模拟失败
		if strings.Contains(string(msg), "fail") {
			log.Printf("[%s] Failing message %d intentionally", h.Name, i+1)
			results[i] = errors.New("simulated failure for testing")
		}
	}

	return results
}

// runAdvancedFeatures 演示所有高级功能的使用方式
func runAdvancedFeatures() {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 获取RabbitMQ连接URL
	rabbitURL := getEnvOrDefault("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")

	// 示例1: 连接管理
	log.Println("=== 连接管理示例 ===")
	connManager := rabbitmq.NewConnectionManager(rabbitURL, 3)

	// 设置连接状态回调
	connManager.SetConnectionStateCallback(func(connected bool) {
		if connected {
			log.Println("RabbitMQ连接已恢复")
		} else {
			log.Println("RabbitMQ连接已断开，正在尝试重连...")
		}
	})

	// 获取连接
	conn1, err := connManager.GetConnection()
	if err != nil {
		log.Fatalf("获取连接失败: %v", err)
	}

	// 初始化RabbitMQ基础设施
	if err := rabbitmq.Setup(conn1); err != nil {
		log.Fatalf("初始化RabbitMQ基础设施失败: %v", err)
	}
	log.Println("RabbitMQ基础设施初始化成功")

	// 示例2: 发布消息
	log.Println("\n=== 发布消息示例 ===")
	taskPublisher, err := rabbitmq.NewTaskPublisher(conn1)
	if err != nil {
		log.Fatalf("创建任务发布者失败: %v", err)
	}
	defer taskPublisher.Close()

	// 发布一些测试消息，包括正常消息和会失败的消息
	messages := []struct {
		Type     string
		Priority int
		Payload  string
	}{
		{"vulnerability", 2, `{"id":"msg1","target":"10.0.0.1","type":"normal"}`},
		{"port", 1, `{"id":"msg2","target":"10.0.0.2","type":"fail"}`}, // 这条会失败
		{"discovery", 0, `{"id":"msg3","target":"10.0.0.3","type":"normal"}`},
		{"vulnerability", 2, `{"id":"msg4","target":"10.0.0.4","type":"normal"}`},
	}

	for _, msg := range messages {
		err := taskPublisher.PublishScanTask(ctx, msg.Type, msg.Priority, []byte(msg.Payload))
		if err != nil {
			log.Printf("发布消息失败: %v", err)
		} else {
			log.Printf("发布消息成功: %s (优先级: %d)", msg.Payload, msg.Priority)
		}
	}

	// 示例3: 标准消费
	log.Println("\n=== 标准消费示例 ===")
	scanConsumer, err := rabbitmq.NewScanConsumer(conn1)
	if err != nil {
		log.Fatalf("创建扫描消费者失败: %v", err)
	}

	// 使用普通消息处理器消费高优先级队列
	standardHandler := &ExampleMessageHandler{Name: "标准消费者"}
	if err := scanConsumer.ConsumeHighPriority(ctx, standardHandler); err != nil {
		log.Fatalf("消费高优先级队列失败: %v", err)
	}
	log.Println("开始消费高优先级队列")

	// 示例4: 死信队列处理
	log.Println("\n=== 死信队列处理示例 ===")
	deadLetterConsumer, err := rabbitmq.NewDeadLetterConsumer(conn1, 3)
	if err != nil {
		log.Fatalf("创建死信消费者失败: %v", err)
	}

	// 启动死信处理
	//if err := deadLetterConsumer.Start(ctx); err != nil {
	//	log.Fatalf("启动死信处理失败: %v", err)
	//}
	//log.Println("死信处理器已启动")

	// 使用专门的DeadLetterHandler处理死信消息
	deadLetterHandler := &DeadLetterHandler{}
	if err := deadLetterConsumer.Consume(ctx, rabbitmq.RetryQueue5Min, deadLetterHandler); err != nil {
		log.Fatalf("Failed to start dead letter handler: %v", err)
	}
	log.Println("Dead letter handler started successfully")

	// 示例5: 幂等性消费
	log.Println("\n=== 幂等性消费示例 ===")
	// 创建一个中优先级队列的消费者
	mediumConsumer, err := rabbitmq.NewScanConsumer(conn1)
	if err != nil {
		log.Fatalf("创建中优先级消费者失败: %v", err)
	}

	// 使用IdempotentConsumer包装普通消费者
	idempotentHandler := &ExampleMessageHandler{Name: "幂等性消费者"}
	// 创建幂等性消费者，设置24小时缓存过期时间，最多缓存10000条消息
	idempotentConsumer := rabbitmq.NewIdempotentConsumer(mediumConsumer, 24*time.Hour, 10000)

	// 使用幂等性消费者消费中优先级队列
	if err := idempotentConsumer.Consume(ctx, rabbitmq.ScanMediumPriorityQueue, idempotentHandler); err != nil {
		log.Fatalf("幂等性消费中优先级队列失败: %v", err)
	}

	log.Println("开始幂等性消费中优先级队列")

	// 示例6: 批量处理
	log.Println("\n=== 批量处理示例 ===")
	batchConsumer, err := rabbitmq.NewBatchConsumer(conn1, 5, 3*time.Second)
	if err != nil {
		log.Fatalf("创建批量消费者失败: %v", err)
	}

	// 使用传统的批量处理器消费低优先级队列
	batchHandler := &BatchHandler{Name: "传统批量处理器"}
	if err := batchConsumer.ConsumeBatch(ctx, rabbitmq.ScanLowPriorityQueue, batchHandler); err != nil {
		log.Fatalf("批量消费低优先级队列失败: %v", err)
	}
	log.Println("开始传统批量消费低优先级队列")

	// 创建另一个批量消费者，使用支持单条消息确认/拒绝的模式
	batchConsumerWithResults, err := rabbitmq.NewBatchConsumer(conn1, 5, 3*time.Second)
	if err != nil {
		log.Fatalf("创建支持单条确认的批量消费者失败: %v", err)
	}

	// 使用支持单条确认/拒绝的批量处理器
	batchHandlerWithResults := &BatchHandlerWithResults{Name: "增强型批量处理器"}
	if err := batchConsumerWithResults.ConsumeBatchWithResults(ctx, rabbitmq.ScanMediumPriorityQueue, batchHandlerWithResults); err != nil {
		log.Fatalf("增强型批量消费中优先级队列失败: %v", err)
	}
	log.Println("开始增强型批量消费中优先级队列(支持单条消息确认/拒绝)")

	// 示例7: 指标收集
	log.Println("\n=== 指标收集示例 ===")
	metricsCollector := rabbitmq.NewMetricsCollector(
		"http://localhost:15672", // RabbitMQ管理界面URL
		"guest",                  // 用户名
		"guest",                  // 密码
		30*time.Second,           // 收集间隔
	)

	// 启动指标收集
	metricsCollector.Start(ctx)
	log.Println("指标收集器已启动")

	// 每10秒打印一次队列状态
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				log.Println("\n=== 队列状态 ===")
				metricsCollector.PrintQueueStatus()
			}
		}
	}()

	// 每5秒重新发布一些测试消息，用于演示幂等性和批量处理
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		batchSize := 10
		count := 0

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				count++
				log.Printf("\n=== 第%d次发布测试消息 ===", count)

				// 发布一批消息用于测试批量处理
				for i := 0; i < batchSize; i++ {
					payload := fmt.Sprintf(`{"id":"batch-%d-%d","target":"192.168.1.%d","type":"normal"}`, count, i, i)
					if err := taskPublisher.PublishScanTask(ctx, "discovery", 0, []byte(payload)); err != nil {
						log.Printf("发布批量测试消息失败: %v", err)
					}
				}
				log.Printf("发布了%d条批量测试消息到低优先级队列", batchSize)

				// 重新发布之前的消息，用于测试幂等性
				for _, msg := range messages {
					err := taskPublisher.PublishScanTask(ctx, msg.Type, msg.Priority, []byte(msg.Payload))
					if err != nil {
						log.Printf("重新发布消息失败: %v", err)
					}
				}
				log.Printf("重新发布了%d条测试消息", len(messages))
			}
		}
	}()

	// 示例8: 综合使用场景 - 演示完整的消息处理流程
	log.Println("\n=== 综合使用场景示例 ===")
	// 使用连接池获取另一个连接
	conn2, err := connManager.GetConnection()
	if err != nil {
		log.Fatalf("获取第二个连接失败: %v", err)
	}
	defer connManager.ReleaseConnection(conn2)

	// 创建资产任务发布者
	assetPublisher, err := rabbitmq.NewTaskPublisher(conn2)
	if err != nil {
		log.Fatalf("创建资产发布者失败: %v", err)
	}
	defer assetPublisher.Close()

	// 发布一些资产任务
	assetOperations := []string{"create", "update", "delete"}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		ticker := time.NewTicker(8 * time.Second)
		defer ticker.Stop()

		count := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				count++
				operation := assetOperations[count%len(assetOperations)]
				payload := fmt.Sprintf(`{"id":"asset-%d","operation":"%s","data":{"name":"server-%d"}}`, count, operation, count)

				if err := assetPublisher.PublishAssetTask(ctx, operation, []byte(payload)); err != nil {
					log.Printf("发布资产任务失败: %v", err)
				} else {
					log.Printf("发布资产任务成功: %s", payload)
				}
			}
		}
	}()

	// 等待中断信号
	log.Println("\n服务已启动，按Ctrl+C退出...")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("正在关闭...")
	// 取消上下文
	cancel()

	// 等待所有后台任务完成
	wg.Wait()

	// 停止指标收集
	metricsCollector.Stop()

	// 关闭所有连接
	scanConsumer.Close()
	deadLetterConsumer.Close()
	idempotentConsumer.Close()
	batchConsumer.Close()
	batchConsumerWithResults.Close()
	connManager.CloseAll()

	log.Println("服务已正常关闭")
}

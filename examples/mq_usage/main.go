package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/blackarbiter/go-sac/pkg/mq/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

// ScanTaskHandler 实现MessageHandler接口处理扫描任务
type ScanTaskHandler struct{}

func (h *ScanTaskHandler) HandleMessage(ctx context.Context, message []byte) error {
	log.Printf("Processing scan task: %s", string(message))

	// 模拟处理，对包含"fail"的消息故意失败
	if strings.Contains(string(message), "fail") {
		log.Printf("Task failed intentionally for testing")
		return errors.New("task failed intentionally")
	}

	// 模拟处理时间
	time.Sleep(500 * time.Millisecond)
	return nil
}

// DeadLetterHandler 实现MessageHandler接口处理死信消息
type DeadLetterHandler struct{}

func (h *DeadLetterHandler) HandleMessage(ctx context.Context, message []byte) error {
	log.Printf("Dead letter handler processing message: %s", string(message))

	// 在实际应用中，这里可能会记录失败消息、发送告警等
	// 例如可以将失败消息写入数据库，发送通知等
	log.Printf("处理失败消息: %s", string(message))

	// 模拟处理时间
	time.Sleep(200 * time.Millisecond)
	return nil
}

func main() {
	// 定义命令行参数
	mode := flag.String("mode", "advanced", "运行模式: basic(基本示例) 或 advanced(高级功能示例)")
	flag.Parse()

	// 根据模式选择运行不同的示例
	switch *mode {
	case "basic":
		log.Println("运行基本示例...")
		runBasicExample()
	case "advanced":
		log.Println("运行高级功能示例...")
		runAdvancedFeatures()
	default:
		log.Fatalf("未知模式: %s, 请使用 'basic' 或 'advanced'", *mode)
	}
}

// runBasicExample 运行基本的RabbitMQ示例
func runBasicExample() {
	// Setup RabbitMQ connection
	log.Println("Connecting to RabbitMQ...")

	// 获取RabbitMQ连接URL，允许通过环境变量配置
	rabbitURL := getEnvOrDefault("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")

	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	// Initialize RabbitMQ infrastructure (run once at startup)
	log.Println("Setting up RabbitMQ infrastructure...")
	if err := rabbitmq.Setup(conn); err != nil {
		log.Fatalf("Failed to setup RabbitMQ: %v", err)
	}
	log.Println("RabbitMQ infrastructure initialized successfully")

	// Create context for shutdown coordination
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Example: Task Service publishes scan tasks
	log.Println("Creating task publisher...")
	taskPublisher, err := rabbitmq.NewTaskPublisher(conn)
	if err != nil {
		log.Fatalf("Failed to create task publisher: %v", err)
	}
	defer taskPublisher.Close()
	log.Println("Task publisher created successfully")

	// Example: Scan Service consumes scan tasks
	log.Println("Creating scan consumer...")
	scanConsumer, err := rabbitmq.NewScanConsumer(conn)
	if err != nil {
		log.Fatalf("Failed to create scan consumer: %v", err)
	}
	defer scanConsumer.Close()
	log.Println("Scan consumer created successfully")

	// 创建死信消费者
	log.Println("Creating dead letter consumer...")
	deadLetterConsumer, err := rabbitmq.NewDeadLetterConsumer(conn, 3)
	if err != nil {
		log.Fatalf("Failed to create dead letter consumer: %v", err)
	}
	defer deadLetterConsumer.Close()

	// 启动死信处理
	if err := deadLetterConsumer.Start(ctx); err != nil {
		log.Fatalf("Failed to start dead letter consumer: %v", err)
	}
	log.Println("Dead letter consumer started successfully")

	// 使用专门的DeadLetterHandler处理死信消息
	deadLetterHandler := &DeadLetterHandler{}
	if err := deadLetterConsumer.Consume(ctx, rabbitmq.RetryQueue5Min, deadLetterHandler); err != nil {
		log.Fatalf("Failed to start dead letter handler: %v", err)
	}
	log.Println("Dead letter handler started successfully")

	// 使用基本的任务处理器 - 处理失败的消息会直接进入死信队列
	taskHandler := &ScanTaskHandler{}

	// Start consuming from high priority queue
	log.Println("Starting to consume high priority scan tasks...")
	if err := scanConsumer.ConsumeHighPriority(ctx, taskHandler); err != nil {
		log.Fatalf("Failed to start consuming high priority scan tasks: %v", err)
	}

	// Start consuming from medium priority queue
	log.Println("Starting to consume medium priority scan tasks...")
	if err := scanConsumer.ConsumeMediumPriority(ctx, taskHandler); err != nil {
		log.Fatalf("Failed to start consuming medium priority scan tasks: %v", err)
	}

	// Start consuming from low priority queue
	log.Println("Starting to consume low priority scan tasks...")
	if err := scanConsumer.ConsumeLowPriority(ctx, taskHandler); err != nil {
		log.Fatalf("Failed to start consuming low priority scan tasks: %v", err)
	}

	// Publish some tasks for demonstration
	go func() {
		// Wait a bit for consumers to start
		time.Sleep(1 * time.Second)

		log.Println("Publishing tasks...")

		// 正常任务
		if err := taskPublisher.PublishScanTask(ctx, "vulnerability", 2, []byte(`{"target":"10.0.0.1","options":{"deep":true},"error":"123"}`)); err != nil {
			log.Printf("Failed to publish high priority task: %v", err)
		} else {
			log.Println("Published high priority task successfully")
		}

		// 故意失败的任务，用于测试死信机制
		if err := taskPublisher.PublishScanTask(ctx, "port", 1, []byte(`{"target":"10.0.0.2","options":{"ports":"1-1000","fail":true}}`)); err != nil {
			log.Printf("Failed to publish medium priority task: %v", err)
		} else {
			log.Println("Published medium priority task (intentional fail) successfully")
		}

		// 正常任务
		if err := taskPublisher.PublishScanTask(ctx, "discovery", 0, []byte(`{"target":"10.0.0.0/24"}`)); err != nil {
			log.Printf("Failed to publish low priority task: %v", err)
		} else {
			log.Println("Published low priority task successfully")
		}

		// 资产任务
		if err := taskPublisher.PublishAssetTask(ctx, "update", []byte(`{"assetId":"server001","attributes":{"os":"ubuntu","version":"22.04"}}`)); err != nil {
			log.Printf("Failed to publish asset task: %v", err)
		} else {
			log.Println("Published asset task successfully")
		}
	}()

	// Wait for termination signal
	log.Println("Service started. Press CTRL+C to exit...")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	// Cancel context to notify all consumers to stop
	cancel()
	time.Sleep(500 * time.Millisecond) // Give time for graceful shutdown
}

// getEnvOrDefault returns environment variable or default value
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

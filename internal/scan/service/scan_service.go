package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/blackarbiter/go-sac/pkg/service"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"net/http"
	"sync"
	"time"

	"github.com/blackarbiter/go-sac/pkg/domain"

	"github.com/blackarbiter/go-sac/pkg/logger"
	"github.com/blackarbiter/go-sac/pkg/mq/rabbitmq"
	"github.com/blackarbiter/go-sac/pkg/scanner"
	"github.com/google/wire"
)

// ScanService 扫描服务
type ScanService struct {
	connManager     *rabbitmq.ConnectionManager
	scannerFactory  scanner.ScannerFactory
	scanConsumer    *rabbitmq.ScanConsumer
	resultPublisher *rabbitmq.ResultPublisher
	wg              sync.WaitGroup
}

// NewScanService 创建扫描服务
func NewScanService(
	connManager *rabbitmq.ConnectionManager,
	scannerFactory scanner.ScannerFactory,
) *ScanService {
	return &ScanService{
		connManager:    connManager,
		scannerFactory: scannerFactory,
	}
}

// Start 启动扫描服务
func (s *ScanService) Start(ctx context.Context) error {
	// 获取连接
	conn, err := s.connManager.GetConnection()
	if err != nil {
		return err
	}

	// 创建消费者
	s.scanConsumer, err = rabbitmq.NewScanConsumer(conn)
	if err != nil {
		return err
	}

	// 创建结果发布者
	s.resultPublisher, err = rabbitmq.NewResultPublisher(conn)
	if err != nil {
		return err
	}

	// 创建带缓冲的通道（大小根据吞吐量配置）
	scheduler := &service.PriorityScheduler{
		HighPriorityChan: make(chan amqp.Delivery, 1000),
		MedPriorityChan:  make(chan amqp.Delivery, 500),
		LowPriorityChan:  make(chan amqp.Delivery, 200),
		StopChan:         make(chan struct{}),
		Handler:          s,
	}

	// 启动统一消费者（替换原有的三个独立消费者）
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		scheduler.Start(ctx) // 核心调度逻辑
	}()

	// 启动队列监听协程（向调度器填充消息）
	go func() {
		err := s.scanConsumer.ConsumeToScheduler(ctx, scheduler)
		if err != nil {
			logger.Logger.Error("consume to scheduler error: ", zap.Error(err))
		}
	}()
	return nil
}

// Stop 停止扫描服务
func (s *ScanService) Stop(ctx context.Context) {
	if s.scanConsumer != nil {
		s.scanConsumer.Close()
	}
	if s.resultPublisher != nil {
		s.resultPublisher.Close()
	}
	s.wg.Wait()
}

// HandleMessage 实现消息处理接口
func (s *ScanService) HandleMessage(ctx context.Context, message []byte) error {
	// 解析任务
	var task domain.ScanTaskPayload
	if err := json.Unmarshal(message, &task); err != nil {
		return err
	}
	priority, err := GetTaskPriority(task.TaskID, "123")
	logger.Logger.Info("Consumer task: " + task.TaskID + ", Priority: " + string(rune(priority)))
	// 获取对应的扫描器
	sne, err := s.scannerFactory.GetScanner(task.ScanType)
	if err != nil {
		return err
	}

	// 执行扫描
	result, err := sne.Scan(ctx, &task)
	if err != nil {
		return err
	}

	// 发布扫描结果
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return err
	}
	//todo：task状态更新

	return s.resultPublisher.PublishScanResult(ctx, resultBytes)
}

// ProviderSet 提供依赖注入集合
var ProviderSet = wire.NewSet(
	NewScanService,
)

func GetTaskPriority(taskID, authToken string) (int, error) {
	// 创建HTTP客户端
	client := &http.Client{Timeout: 10 * time.Second}

	// 构建请求对象
	url := fmt.Sprintf("http://localhost:8088/api/v1/tasks/%s", taskID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置授权头[7,8](@ref)
	req.Header.Set("Authorization", "Bearer "+authToken)

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("异常状态码: %d", resp.StatusCode)
	}

	// 解析JSON响应[4,5](@ref)
	var task domain.Task
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return 0, fmt.Errorf("JSON解析失败: %w", err)
	}

	return int(task.Priority), nil
}

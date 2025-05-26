package service

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/blackarbiter/go-sac/pkg/domain"

	"github.com/blackarbiter/go-sac/pkg/logger"
	"github.com/blackarbiter/go-sac/pkg/mq/rabbitmq"
	"github.com/blackarbiter/go-sac/pkg/scanner"
	"github.com/google/wire"
	"go.uber.org/zap"
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

	// 启动高优先级队列消费者
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.scanConsumer.ConsumeHighPriority(ctx, s); err != nil {
			logger.Logger.Error("high priority consumer error", zap.Error(err))
		}
	}()

	// 启动中优先级队列消费者
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.scanConsumer.ConsumeMediumPriority(ctx, s); err != nil {
			logger.Logger.Error("medium priority consumer error", zap.Error(err))
		}
	}()

	// 启动低优先级队列消费者
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.scanConsumer.ConsumeLowPriority(ctx, s); err != nil {
			logger.Logger.Error("low priority consumer error", zap.Error(err))
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
	logger.Logger.Info("Consumer task: " + task.TaskID)
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

	return s.resultPublisher.PublishScanResult(ctx, resultBytes)
}

// ProviderSet 提供依赖注入集合
var ProviderSet = wire.NewSet(
	NewScanService,
)

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"errors"

	"github.com/blackarbiter/go-sac/pkg/cache/redis"
	"github.com/blackarbiter/go-sac/pkg/config"
	"github.com/blackarbiter/go-sac/pkg/domain"
	sac_errors "github.com/blackarbiter/go-sac/pkg/errors"
	"github.com/blackarbiter/go-sac/pkg/logger"
	"github.com/blackarbiter/go-sac/pkg/metrics"
	"github.com/blackarbiter/go-sac/pkg/mq/rabbitmq"
	"github.com/blackarbiter/go-sac/pkg/scanner"
	scanner_impl "github.com/blackarbiter/go-sac/pkg/scanner/impl"
	"github.com/blackarbiter/go-sac/pkg/service"
	"github.com/google/wire"
	"go.uber.org/zap"
)

// ScanService 扫描服务
type ScanService struct {
	connManager       *rabbitmq.ConnectionManager
	scannerFactory    scanner.ScannerFactory
	scanConsumer      *rabbitmq.ScanConsumer
	resultPublisher   *rabbitmq.ResultPublisher
	timeoutCtrl       *scanner.TimeoutController
	metrics           *metrics.ScannerMetrics
	taskStatusUpdater *TaskStatusUpdaterImpl
	redisConnector    *redis.Connector
	wg                sync.WaitGroup
	config            *config.Config
	globalWorkerPool  chan struct{}        // 全局协程池
	globalTaskQueue   chan func()          // 全局任务队列
	maxConcurrency    int                  // 全局最大并发数
	queueSize         int                  // 全局队列大小
	state             *service.SystemState // 系统状态管理器
}

// NewScanService 创建扫描服务
func NewScanService(
	connManager *rabbitmq.ConnectionManager,
	timeoutCtrl *scanner.TimeoutController,
	metrics *metrics.ScannerMetrics,
	cfg *config.Config,
	scannerFactory scanner.ScannerFactory,
	redisConnector *redis.Connector,
) (*ScanService, error) {
	// 创建任务状态更新器
	taskStatusUpdater := NewTaskStatusUpdater(cfg.GetTaskApiBaseURL(), cfg.GetAuthToken())

	// 为每个扫描器设置任务状态更新器和结果发布器
	scanners := scannerFactory.GetAllScanners()
	for _, scanner := range scanners {
		if baseScanner, ok := scanner.(interface {
			SetTaskStatusUpdater(scanner_impl.TaskStatusUpdater)
			SetResultPublisher(scanner_impl.ResultPublisher)
		}); ok {
			baseScanner.SetTaskStatusUpdater(taskStatusUpdater)
			baseScanner.SetResultPublisher(NewResultPublisher(nil)) // 将在Start方法中设置
		}
	}

	maxWorkers, queueSize := cfg.GetConcurrencyConfig()

	ss := &ScanService{
		connManager:       connManager,
		scannerFactory:    scannerFactory,
		timeoutCtrl:       timeoutCtrl,
		metrics:           metrics,
		taskStatusUpdater: taskStatusUpdater,
		redisConnector:    redisConnector,
		config:            cfg,
		maxConcurrency:    maxWorkers,
		queueSize:         queueSize,
		globalWorkerPool:  make(chan struct{}, maxWorkers),
		globalTaskQueue:   make(chan func(), queueSize),
		state:             service.NewSystemState(),
	}
	go ss.startGlobalWorkerPool()

	return ss, nil
}

// 启动全局工作池
func (s *ScanService) startGlobalWorkerPool() {
	for task := range s.globalTaskQueue {
		// 获取worker槽位
		s.globalWorkerPool <- struct{}{}

		go func(t func()) {
			defer func() {
				// 释放worker槽位
				<-s.globalWorkerPool

				// 背压解除检查（保守策略）
				currentLen := len(s.globalTaskQueue)
				logger.Logger.Info("Current task queue, ", zap.String("length", strconv.Itoa(currentLen)))
				if currentLen <= s.queueSize/2 {
					s.state.ReleaseBackpressure()
				}
			}()

			t() // 执行任务
		}(task)
	}
}

// Start 启动扫描服务
func (s *ScanService) Start(ctx context.Context) error {
	// 获取连接
	conn, err := s.connManager.GetConnection()
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}

	// 创建消费者
	s.scanConsumer, err = rabbitmq.NewScanConsumer(conn)
	if err != nil {
		return fmt.Errorf("failed to create scan consumer: %w", err)
	}

	// 创建结果发布者
	s.resultPublisher, err = rabbitmq.NewResultPublisher(conn)
	if err != nil {
		return fmt.Errorf("failed to create result publisher: %w", err)
	}

	// 更新所有扫描器的结果发布器
	scanners := s.scannerFactory.GetAllScanners()
	for _, scanner := range scanners {
		if baseScanner, ok := scanner.(interface {
			SetResultPublisher(scanner_impl.ResultPublisher)
		}); ok {
			baseScanner.SetResultPublisher(NewResultPublisher(s.resultPublisher))
		}
	}

	// 创建带缓冲的通道（大小根据吞吐量配置）
	scheduler := service.NewPriorityScheduler(s, s.state, s.config)
	// 将scheduler传递给消费者
	s.scanConsumer.SetScheduler(scheduler)

	// 启动统一消费者
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		scheduler.Start(ctx)
	}()

	// 启动队列监听协程
	go func() {
		err := s.scanConsumer.ConsumeToScheduler(ctx, scheduler)
		if err != nil {
			logger.Logger.Error("consume to scheduler error", zap.Error(err))
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
	if s.redisConnector != nil {
		s.redisConnector.Close()
	}
	s.wg.Wait()
}

// HandleMessage 实现消息处理接口
func (s *ScanService) HandleMessage(ctx context.Context, message []byte) error {
	// 解析任务
	var task domain.ScanTaskPayload
	if err := json.Unmarshal(message, &task); err != nil {
		return fmt.Errorf("failed to unmarshal task: %w", err)
	}

	// 获取分布式锁
	lockKey := fmt.Sprintf("task_lock:%s", task.TaskID)
	distLock := redis.NewDistributedLock(ctx, s.redisConnector.GetClient(), lockKey, 10*time.Minute)

	// 尝试获取锁
	if err := distLock.Acquire(); err != nil {
		if errors.Is(err, redis.ErrLockNotAcquired) {
			logger.Logger.Info("Task already being processed by another instance",
				zap.String("taskID", task.TaskID))
			return nil // 静默返回，避免消息重试
		}
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer distLock.Release()

	// 背压状态检查
	if s.state.ShouldStopProcessing() {
		logger.Logger.Warn("Rejecting task due to system backpressure",
			zap.String("taskID", task.TaskID))
		return sac_errors.NewBackpressureError(s.queueSize)
	}

	// 提交任务到全局队列
	select {
	case s.globalTaskQueue <- func() {
		// 获取对应的扫描器
		scanner, err := s.scannerFactory.GetScanner(task.ScanType)
		if err != nil {
			logger.Logger.Error("failed to get scanner",
				zap.Error(err),
				zap.String("taskID", task.TaskID))
			return
		}

		// 执行扫描任务
		_, err = scanner.SyncExecute(ctx, &task)
		if err != nil {
			logger.Logger.Error("failed to execute scan task",
				zap.Error(err),
				zap.String("taskID", task.TaskID))
		}
	}:
		logger.Logger.Info("Scan task queued",
			zap.String("taskID", task.TaskID),
			zap.String("scanType", task.ScanType.String()))
		return nil
	default:
		// 队列满时触发全链路背压
		s.state.TriggerBackpressure()
		logger.Logger.Warn("Global task queue full, triggering full backpressure",
			zap.String("taskID", task.TaskID))
		return sac_errors.NewBackpressureError(s.queueSize)
	}
}

// ProviderSet 提供依赖注入集合
var ProviderSet = wire.NewSet(
	NewScanService,
)

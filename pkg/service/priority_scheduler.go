package service

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/blackarbiter/go-sac/pkg/errors"
	"github.com/blackarbiter/go-sac/pkg/logger"
	"github.com/blackarbiter/go-sac/pkg/mq"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// PriorityScheduler 新增优先级调度器结构体
type PriorityScheduler struct {
	HighPriorityChan chan amqp.Delivery
	MedPriorityChan  chan amqp.Delivery
	LowPriorityChan  chan amqp.Delivery
	Wg               sync.WaitGroup
	Handler          mq.MessageHandler
	Mu               sync.Mutex   // 状态锁
	State            *SystemState // 系统状态管理器
}

// NewPriorityScheduler creates a new PriorityScheduler instance
func NewPriorityScheduler(handler mq.MessageHandler, state *SystemState) *PriorityScheduler {
	return &PriorityScheduler{
		HighPriorityChan: make(chan amqp.Delivery, 3),
		MedPriorityChan:  make(chan amqp.Delivery, 3),
		LowPriorityChan:  make(chan amqp.Delivery, 3),
		Handler:          handler,
		State:            state,
	}
}

func (s *PriorityScheduler) Start(ctx context.Context) {
	const (
		normalSleep  = 100 * time.Millisecond
		backoffSleep = 500 * time.Millisecond
	)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// 背压状态检查
			if s.State.ShouldStopProcessing() {
				time.Sleep(backoffSleep)
				continue
			}

			// 精确获取各通道消息数（避免竞态）
			highCount := len(s.HighPriorityChan)
			medCount := len(s.MedPriorityChan)
			lowCount := len(s.LowPriorityChan)

			totalWeight := 0
			if highCount > 0 {
				totalWeight += 60
			}
			if medCount > 0 {
				totalWeight += 30
			}
			if lowCount > 0 {
				totalWeight += 10
			}

			if totalWeight == 0 {
				time.Sleep(normalSleep)
				continue
			}

			// 生成随机数决定本次处理的优先级
			r := rand.Intn(totalWeight)
			logger.Logger.Info("Priority Scheduler",
				zap.Int("High Count", highCount),
				zap.Int("Medium Count", medCount),
				zap.Int("Low Count", lowCount),
				zap.Int("Total Weight", totalWeight),
				zap.Int("Random Number", r))
			var msg amqp.Delivery

			// 根据权重选择通道并尝试非阻塞读取
			switch {
			case highCount > 0 && r < 60:
				select {
				case msg = <-s.HighPriorityChan:
					s.processWithPriority(ctx, msg, "high")
				default:
					time.Sleep(normalSleep)
				}
			case medCount > 0 && r < 90:
				select {
				case msg = <-s.MedPriorityChan:
					s.processWithPriority(ctx, msg, "medium")
				default:
					time.Sleep(normalSleep)
				}
			case lowCount > 0:
				select {
				case msg = <-s.LowPriorityChan:
					s.processWithPriority(ctx, msg, "low")
				default:
					time.Sleep(normalSleep)
				}
			}
		}
	}
}

func (s *PriorityScheduler) processWithPriority(ctx context.Context, msg amqp.Delivery, priority string) {
	logger.Logger.Info("Process message", zap.String("Priority", priority))
	err := s.Handler.HandleMessage(ctx, msg.Body)
	if err != nil {
		// 背压错误特殊处理
		if errors.IsBackpressureError(err) {
			logger.Logger.Warn("Message rejected due to backpressure",
				zap.String("priority", priority))
			// 重新入队等待后续处理
			msg.Nack(false, true)
		} else {
			// 其他错误处理
			logger.Logger.Error("Failed to process message",
				zap.String("priority", priority),
				zap.Error(err))
			msg.Nack(false, false)
		}
		return
	}

	// 成功处理
	msg.Ack(false)
}

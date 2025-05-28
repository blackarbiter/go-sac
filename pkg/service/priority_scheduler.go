package service

import (
	"context"
	"github.com/blackarbiter/go-sac/pkg/logger"
	"github.com/blackarbiter/go-sac/pkg/mq"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"math/rand"
	"sync"
	"time"
)

// PriorityScheduler 新增优先级调度器结构体
type PriorityScheduler struct {
	HighPriorityChan chan amqp.Delivery
	MedPriorityChan  chan amqp.Delivery
	LowPriorityChan  chan amqp.Delivery
	StopChan         chan struct{}
	Wg               sync.WaitGroup
	Handler          mq.MessageHandler
}

// 修改文档2中的PriorityScheduler的Start方法
func (s *PriorityScheduler) Start(ctx context.Context) {
	rand.Seed(time.Now().UnixNano())

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// 检测各通道是否有消息
			highAvailable := len(s.HighPriorityChan) > 0
			medAvailable := len(s.MedPriorityChan) > 0
			lowAvailable := len(s.LowPriorityChan) > 0

			totalWeight := 0
			if highAvailable {
				totalWeight += 60
			}
			if medAvailable {
				totalWeight += 30
			}
			if lowAvailable {
				totalWeight += 10
			}

			if totalWeight == 0 {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// 生成随机数决定本次处理的优先级
			r := rand.Intn(totalWeight)
			var msg amqp.Delivery
			processed := false

			// 根据权重选择通道并尝试非阻塞读取
			switch {
			case highAvailable && r < 60:
				select {
				case msg = <-s.HighPriorityChan:
					s.processWithPriority(ctx, msg, "high")
					processed = true
				default:
					// 高优先级通道在检查后可能已无消息，尝试降级
				}
			case medAvailable && (r < 100 || !highAvailable):
				select {
				case msg = <-s.MedPriorityChan:
					s.processWithPriority(ctx, msg, "medium")
					processed = true
				default:
				}
			case lowAvailable:
				select {
				case msg = <-s.LowPriorityChan:
					s.processWithPriority(ctx, msg, "low")
					processed = true
				default:
				}
			}

			// 如果首选通道未处理成功，尝试降级到其他可用通道
			if !processed {
				// 重新计算可用通道
				hasHigh := len(s.HighPriorityChan) > 0
				hasMed := len(s.MedPriorityChan) > 0
				hasLow := len(s.LowPriorityChan) > 0

				// 按优先级顺序降级处理
				switch {
				case hasHigh:
					select {
					case msg = <-s.HighPriorityChan:
						s.processWithPriority(ctx, msg, "high")
					default:
					}
				case hasMed:
					select {
					case msg = <-s.MedPriorityChan:
						s.processWithPriority(ctx, msg, "medium")
					default:
					}
				case hasLow:
					select {
					case msg = <-s.LowPriorityChan:
						s.processWithPriority(ctx, msg, "low")
					default:
					}
				}
			}
		}
	}
}

func (s *PriorityScheduler) processWithPriority(ctx context.Context, d amqp.Delivery, priority string) {
	// 实现带优先级的处理逻辑
	if err := s.Handler.HandleMessage(ctx, d.Body); err != nil {
		// 出现错误不执行重试
		// todo：task通知
		logger.Logger.Error("Process with priority error: ", zap.Error(err))
	} else {
		d.Ack(false)
	}
}

package service

import (
	"context"

	"github.com/blackarbiter/go-sac/pkg/logger"
	"github.com/blackarbiter/go-sac/pkg/mq/rabbitmq"
	"go.uber.org/zap"
)

// ResultPublisherImpl 实现结果发布接口
type ResultPublisherImpl struct {
	publisher *rabbitmq.ResultPublisher
}

// NewResultPublisher 创建结果发布器
func NewResultPublisher(publisher *rabbitmq.ResultPublisher) *ResultPublisherImpl {
	return &ResultPublisherImpl{
		publisher: publisher,
	}
}

// PublishScanResult 发布扫描结果
func (p *ResultPublisherImpl) PublishScanResult(ctx context.Context, result []byte) error {
	if err := p.publisher.PublishScanResult(ctx, result); err != nil {
		logger.Logger.Error("Failed to publish scan result",
			zap.Error(err))
		return err
	}

	logger.Logger.Info("Scan result published successfully")
	return nil
}

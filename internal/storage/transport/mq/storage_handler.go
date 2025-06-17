package mq

import (
	"context"
	"encoding/json"

	"github.com/blackarbiter/go-sac/internal/storage/service"
	"github.com/blackarbiter/go-sac/pkg/domain"
	"github.com/blackarbiter/go-sac/pkg/mq"
)

// StorageMessageHandler 处理资产相关的 MQ 消息
type StorageMessageHandler struct {
	factory service.StorageProcessorFactory
}

// NewStorageMessageHandler creates a new Consumer instance
func NewStorageMessageHandler(factory service.StorageProcessorFactory) *StorageMessageHandler {
	return &StorageMessageHandler{factory: factory}
}

// HandleMessage processes incoming messages
func (c *StorageMessageHandler) HandleMessage(ctx context.Context, message []byte) error {
	var result domain.ScanResult
	if err := json.Unmarshal(message, &result); err != nil {
		return err
	}

	processor, err := c.factory.GetProcessor(result.ScanType)
	if err != nil {
		return err
	}

	return processor.Process(ctx, &result)
}

// Ensure Consumer implements mq.MessageHandler
var _ mq.MessageHandler = (*StorageMessageHandler)(nil)

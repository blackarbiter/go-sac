package mq

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/blackarbiter/go-sac/internal/asset/dto"
	"github.com/blackarbiter/go-sac/internal/asset/service"
	"github.com/blackarbiter/go-sac/internal/asset/transport/http"
)

// AssetMessageHandler 处理资产相关的 MQ 消息
type AssetMessageHandler struct {
	binder  *http.AssetBinder
	factory service.AssetProcessorFactory
}

// NewAssetMessageHandler 创建资产消息处理器
func NewAssetMessageHandler(binder *http.AssetBinder, factory service.AssetProcessorFactory) *AssetMessageHandler {
	return &AssetMessageHandler{
		binder:  binder,
		factory: factory,
	}
}

// HandleMessage 处理消息
func (h *AssetMessageHandler) HandleMessage(ctx context.Context, body []byte) error {
	// 解析消息头
	var msgHeader struct {
		Type   string `json:"type"`
		Action string `json:"action"`
	}
	if err := json.Unmarshal(body, &msgHeader); err != nil {
		return fmt.Errorf("failed to unmarshal message header: %w", err)
	}

	// 只处理资产操作消息
	if msgHeader.Type != "asset_operation" {
		return nil
	}

	// 解析资产操作消息
	var assetMsg struct {
		AssetType string          `json:"asset_type"`
		Payload   json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(body, &assetMsg); err != nil {
		return fmt.Errorf("failed to unmarshal asset operation: %w", err)
	}

	// 获取处理器
	processor, err := h.factory.GetProcessor(assetMsg.AssetType)
	if err != nil {
		return fmt.Errorf("unsupported asset type: %w", err)
	}

	// 处理不同操作
	switch msgHeader.Action {
	case "create":
		return h.handleCreate(ctx, processor, assetMsg.AssetType, assetMsg.Payload)
	case "update":
		return h.handleUpdate(ctx, processor, assetMsg.AssetType, assetMsg.Payload)
	case "delete":
		return h.handleDelete(ctx, processor, assetMsg.Payload)
	default:
		return fmt.Errorf("unsupported action: %s", msgHeader.Action)
	}
}

// handleCreate 处理创建操作
func (h *AssetMessageHandler) handleCreate(
	ctx context.Context,
	processor service.AssetProcessor,
	assetType string,
	payload json.RawMessage,
) error {
	// 绑定请求
	req, err := h.binder.Bind(assetType, payload)
	if err != nil {
		return fmt.Errorf("failed to bind request: %w", err)
	}

	// 类型断言
	baseReq, ok := req.(dto.BaseRequest)
	if !ok {
		return fmt.Errorf("invalid request format for create")
	}

	// 创建资产
	baseAsset := baseReq.ToBaseAsset(assetType)
	_, err = processor.Create(ctx, baseAsset, req)
	return err
}

// handleUpdate 处理更新操作
func (h *AssetMessageHandler) handleUpdate(
	ctx context.Context,
	processor service.AssetProcessor,
	assetType string,
	payload json.RawMessage,
) error {
	// 解析更新消息
	var updateMsg struct {
		ID      uint            `json:"id"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(payload, &updateMsg); err != nil {
		return fmt.Errorf("failed to unmarshal update message: %w", err)
	}

	// 绑定请求
	req, err := h.binder.Bind(assetType, updateMsg.Payload)
	if err != nil {
		return fmt.Errorf("failed to bind request: %w", err)
	}

	// 类型断言
	baseReq, ok := req.(dto.BaseRequest)
	if !ok {
		return fmt.Errorf("invalid request format for update")
	}

	// 更新资产
	baseAsset := baseReq.ToBaseAsset(assetType)
	return processor.Update(ctx, updateMsg.ID, baseAsset, req)
}

// handleDelete 处理删除操作
func (h *AssetMessageHandler) handleDelete(
	ctx context.Context,
	processor service.AssetProcessor,
	payload json.RawMessage,
) error {
	// 解析删除消息
	var deleteMsg struct {
		ID uint `json:"id"`
	}
	if err := json.Unmarshal(payload, &deleteMsg); err != nil {
		return fmt.Errorf("failed to unmarshal delete message: %w", err)
	}

	// 删除资产
	return processor.Delete(ctx, deleteMsg.ID)
}

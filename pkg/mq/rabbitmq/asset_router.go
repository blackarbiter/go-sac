package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// MessageHandler 消息处理函数类型
type MessageHandler func(context.Context, []byte) error

// MessageRouter 消息路由器
type MessageRouter struct {
	handlers map[string]MessageHandler
}

// NewMessageRouter 创建消息路由器
func NewMessageRouter() *MessageRouter {
	return &MessageRouter{
		handlers: make(map[string]MessageHandler),
	}
}

// RegisterHandler 注册消息处理器
func (r *MessageRouter) RegisterHandler(messageType string, handler MessageHandler) {
	r.handlers[messageType] = handler
}

// RouteMessage 路由消息到对应的处理器
func (r *MessageRouter) RouteMessage(ctx context.Context, msg amqp.Delivery) error {
	// 解析消息类型
	var msgHeader struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(msg.Body, &msgHeader); err != nil {
		return fmt.Errorf("failed to unmarshal message header: %w", err)
	}

	// 查找处理器
	handler, exists := r.handlers[msgHeader.Type]
	if !exists {
		return fmt.Errorf("no handler for message type: %s", msgHeader.Type)
	}

	// 处理消息
	return handler(ctx, msg.Body)
}

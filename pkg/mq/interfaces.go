package mq

import (
	"context"
)

// MessageHandler defines the interface for handling messages
type MessageHandler interface {
	// HandleMessage processes a message and returns an error if processing fails
	HandleMessage(ctx context.Context, message []byte) error
}

// Publisher defines the interface for publishing messages
type Publisher interface {
	// Publish sends a message to a specific exchange with routing key
	Publish(ctx context.Context, exchange, routingKey string, message []byte) error

	// Close closes the publisher connection
	Close() error
}

// Consumer defines the interface for consuming messages
type Consumer interface {
	// Consume starts consuming messages from the specified queue
	Consume(ctx context.Context, queueName string, handler MessageHandler) error

	// Close stops consuming and closes the connection
	Close() error
}

// TaskPublisher defines the interface for publishing task-specific messages
type TaskPublisher interface {
	Publisher

	// PublishScanTask publishes a scan task with the given type and priority
	PublishScanTask(ctx context.Context, taskType string, priority int, payload []byte) error

	// PublishAssetTask publishes an asset task with the given operation
	PublishAssetTask(ctx context.Context, operation string, payload []byte) error
}

// ResultPublisher defines the interface for publishing result-specific messages
type ResultPublisher interface {
	Publisher

	// PublishScanResult publishes a scan result to be stored
	PublishScanResult(ctx context.Context, payload []byte) error
}

// NotificationPublisher defines the interface for publishing notifications
type NotificationPublisher interface {
	Publisher

	// PublishNotification publishes a notification message
	PublishNotification(ctx context.Context, payload []byte) error
}

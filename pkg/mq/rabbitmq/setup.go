package rabbitmq

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go" // Modern RabbitMQ client
)

// ExchangeConfig defines exchange configuration
type ExchangeConfig struct {
	Name       string
	Type       string
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Arguments  amqp.Table
}

// QueueConfig defines queue configuration
type QueueConfig struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Arguments  amqp.Table
}

// BindingConfig defines binding between exchange and queue
type BindingConfig struct {
	QueueName    string
	ExchangeName string
	RoutingKey   string
	NoWait       bool
	Arguments    amqp.Table
}

// Setup handles the initialization of RabbitMQ exchanges, queues and bindings
func Setup(conn *amqp.Connection) error {
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	// Setup exchanges
	if err := setupExchanges(ch); err != nil {
		return fmt.Errorf("failed to setup exchanges: %w", err)
	}

	// Setup queues
	if err := setupQueues(ch); err != nil {
		return fmt.Errorf("failed to setup queues: %w", err)
	}

	// Setup bindings
	if err := setupBindings(ch); err != nil {
		return fmt.Errorf("failed to setup bindings: %w", err)
	}

	log.Println("RabbitMQ setup completed successfully")
	return nil
}

func setupExchanges(ch *amqp.Channel) error {
	exchanges := []ExchangeConfig{
		{
			Name:       TaskDispatchExchange,
			Type:       "topic",
			Durable:    true,
			AutoDelete: false,
			Internal:   false,
			NoWait:     false,
			Arguments:  nil,
		},
		{
			Name:       ResultProcessExchange,
			Type:       "direct",
			Durable:    true,
			AutoDelete: false,
			Internal:   false,
			NoWait:     false,
			Arguments:  nil,
		},
		{
			Name:       NotificationExchange,
			Type:       "fanout",
			Durable:    true,
			AutoDelete: false,
			Internal:   false,
			NoWait:     false,
			Arguments:  nil,
		},
		{
			Name:       RetryExchange,
			Type:       "topic",
			Durable:    true,
			AutoDelete: false,
			Internal:   false,
			NoWait:     false,
			Arguments:  nil,
		},
	}

	for _, exchange := range exchanges {
		err := ch.ExchangeDeclare(
			exchange.Name,
			exchange.Type,
			exchange.Durable,
			exchange.AutoDelete,
			exchange.Internal,
			exchange.NoWait,
			exchange.Arguments,
		)
		if err != nil {
			return fmt.Errorf("failed to declare exchange %s: %w", exchange.Name, err)
		}
		log.Printf("Exchange declared: %s", exchange.Name)
	}

	return nil
}

func setupQueues(ch *amqp.Channel) error {
	queues := []QueueConfig{
		{
			Name:       ScanHighPriorityQueue,
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Arguments: amqp.Table{
				MaxPriority:          10,
				DeadLetterExchange:   RetryExchange,
				DeadLetterRoutingKey: "retry.scan.high",
			},
		},
		{
			Name:       ScanMediumPriorityQueue,
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Arguments: amqp.Table{
				MaxPriority:          5,
				DeadLetterExchange:   RetryExchange,
				DeadLetterRoutingKey: "retry.scan.medium",
			},
		},
		{
			Name:       ScanLowPriorityQueue,
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Arguments: amqp.Table{
				DeadLetterExchange:   RetryExchange,
				DeadLetterRoutingKey: "retry.scan.low",
			},
		},
		{
			Name:       AssetTaskQueue,
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Arguments: amqp.Table{
				DeadLetterExchange:   RetryExchange,
				DeadLetterRoutingKey: "retry.asset",
			},
		},
		{
			Name:       ResultStorageQueue,
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Arguments: amqp.Table{
				DeadLetterExchange:   RetryExchange,
				DeadLetterRoutingKey: "retry.storage",
			},
		},
		{
			Name:       NotificationEmailQueue,
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Arguments:  nil,
		},
		{
			Name:       NotificationSMSQueue,
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Arguments:  nil,
		},
		{
			Name:       NotificationSystemQueue,
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Arguments:  nil,
		},
		{
			Name:       RetryQueue5Min,
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Arguments: amqp.Table{
				MessageTTL:           300000, // 5 minutes in milliseconds
				DeadLetterExchange:   TaskDispatchExchange,
				DeadLetterRoutingKey: "", // Will be set dynamically based on original routing key
			},
		},
		{
			Name:       ManualInterventionQueue,
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Arguments:  nil,
		},
	}

	for _, queue := range queues {
		_, err := ch.QueueDeclare(
			queue.Name,
			queue.Durable,
			queue.AutoDelete,
			queue.Exclusive,
			queue.NoWait,
			queue.Arguments,
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", queue.Name, err)
		}
		log.Printf("Queue declared: %s", queue.Name)
	}

	return nil
}

func setupBindings(ch *amqp.Channel) error {
	bindings := []BindingConfig{
		// Scan task bindings
		{
			QueueName:    ScanHighPriorityQueue,
			ExchangeName: TaskDispatchExchange,
			RoutingKey:   ScanHighPattern,
			NoWait:       false,
			Arguments:    nil,
		},
		{
			QueueName:    ScanMediumPriorityQueue,
			ExchangeName: TaskDispatchExchange,
			RoutingKey:   ScanMediumPattern,
			NoWait:       false,
			Arguments:    nil,
		},
		{
			QueueName:    ScanLowPriorityQueue,
			ExchangeName: TaskDispatchExchange,
			RoutingKey:   ScanLowPattern,
			NoWait:       false,
			Arguments:    nil,
		},
		// Asset task binding
		{
			QueueName:    AssetTaskQueue,
			ExchangeName: TaskDispatchExchange,
			RoutingKey:   AssetPattern,
			NoWait:       false,
			Arguments:    nil,
		},
		// Result storage binding
		{
			QueueName:    ResultStorageQueue,
			ExchangeName: ResultProcessExchange,
			RoutingKey:   ResultStoragePattern,
			NoWait:       false,
			Arguments:    nil,
		},
		// Notification bindings
		{
			QueueName:    NotificationEmailQueue,
			ExchangeName: NotificationExchange,
			RoutingKey:   "", // Fanout exchange ignores routing key
			NoWait:       false,
			Arguments:    nil,
		},
		{
			QueueName:    NotificationSMSQueue,
			ExchangeName: NotificationExchange,
			RoutingKey:   "", // Fanout exchange ignores routing key
			NoWait:       false,
			Arguments:    nil,
		},
		{
			QueueName:    NotificationSystemQueue,
			ExchangeName: NotificationExchange,
			RoutingKey:   "", // Fanout exchange ignores routing key
			NoWait:       false,
			Arguments:    nil,
		},
		// Retry bindings
		{
			QueueName:    RetryQueue5Min,
			ExchangeName: RetryExchange,
			RoutingKey:   RetryPattern,
			NoWait:       false,
			Arguments:    nil,
		},
		{
			QueueName:    ManualInterventionQueue,
			ExchangeName: RetryExchange,
			RoutingKey:   ManualPattern,
			NoWait:       false,
			Arguments:    nil,
		},
	}

	for _, binding := range bindings {
		err := ch.QueueBind(
			binding.QueueName,
			binding.RoutingKey,
			binding.ExchangeName,
			binding.NoWait,
			binding.Arguments,
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue %s to exchange %s with routing key %s: %w",
				binding.QueueName, binding.ExchangeName, binding.RoutingKey, err)
		}
		log.Printf("Queue binding created: %s -> %s with routing key %s",
			binding.QueueName, binding.ExchangeName, binding.RoutingKey)
	}

	return nil
}

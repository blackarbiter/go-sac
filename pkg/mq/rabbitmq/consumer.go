package rabbitmq

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  ConsumerConfig
}

type ConsumerConfig struct {
	URL           string
	QueueName     string
	DLXExchange   string
	DLXRoutingKey string
	PrefetchCount int
	RetryLimit    int
	RetryDelay    time.Duration
}

func NewConsumer(config ConsumerConfig) (*Consumer, error) {
	c := &Consumer{config: config}
	if err := c.connect(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Consumer) connect() error {
	var err error
	if c.conn, err = amqp.Dial(c.config.URL); err != nil {
		return err
	}

	if c.channel, err = c.conn.Channel(); err != nil {
		return err
	}

	// Declare DLX
	if err := c.channel.ExchangeDeclare(
		c.config.DLXExchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	// Declare main queue with DLX
	args := amqp.Table{
		"x-dead-letter-exchange":    c.config.DLXExchange,
		"x-dead-letter-routing-key": c.config.DLXRoutingKey,
	}

	_, err = c.channel.QueueDeclare(
		c.config.QueueName,
		true,
		false,
		false,
		false,
		args,
	)
	if err != nil {
		return err
	}

	if err := c.channel.Qos(c.config.PrefetchCount, 0, false); err != nil {
		return err
	}

	return nil
}

func (c *Consumer) Consume(ctx context.Context, handler func(context.Context, amqp.Delivery) error) error {
	msgs, err := c.channel.Consume(
		c.config.QueueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	for msg := range msgs {
		select {
		case <-ctx.Done():
			return nil
		default:
			if err := c.processMessage(msg, handler); err != nil {
				fmt.Printf("Error processing message: %v\n", err)
			}
		}
	}
	return nil
}

func (c *Consumer) processMessage(msg amqp.Delivery, handler func(context.Context, amqp.Delivery) error) error {
	ctx := context.Background()
	retryCount := getRetryCount(msg.Headers)

	if retryCount >= c.config.RetryLimit {
		msg.Nack(false, false)
		return fmt.Errorf("message exceeded retry limit")
	}

	err := handler(ctx, msg)
	if err != nil {
		newHeaders := incrementRetryCount(msg.Headers)
		delay := c.calculateDelay(retryCount)

		// Re-publish with delay
		err = c.channel.Publish(
			"",
			c.config.QueueName,
			true,
			false,
			amqp.Publishing{
				Headers:      newHeaders,
				Body:         msg.Body,
				DeliveryMode: amqp.Persistent,
				Expiration:   fmt.Sprintf("%d", delay.Milliseconds()),
			},
		)
		if err != nil {
			return fmt.Errorf("failed to republish message: %v", err)
		}
		msg.Ack(false)
		return nil
	}

	msg.Ack(false)
	return nil
}

func getRetryCount(headers amqp.Table) int {
	if val, ok := headers["x-retry-count"]; ok {
		if count, ok := val.(int32); ok {
			return int(count)
		}
	}
	return 0
}

func incrementRetryCount(headers amqp.Table) amqp.Table {
	newHeaders := make(amqp.Table)
	for k, v := range headers {
		newHeaders[k] = v
	}
	newHeaders["x-retry-count"] = getRetryCount(headers) + 1
	return newHeaders
}

func (c *Consumer) calculateDelay(retryCount int) time.Duration {
	return time.Duration(retryCount+1) * c.config.RetryDelay
}

func (c *Consumer) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	return c.conn.Close()
}

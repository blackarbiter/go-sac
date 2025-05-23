package rabbitmq

import (
	"context"
	"errors"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// EnhancedProducer extends Producer with additional capabilities
type EnhancedProducer struct {
	*Producer
}

// NewEnhancedProducer creates a new enhanced producer using URL
func NewEnhancedProducer(config ProducerConfig) (*EnhancedProducer, error) {
	producer, err := NewProducer(config)
	if err != nil {
		return nil, err
	}

	return &EnhancedProducer{
		Producer: producer,
	}, nil
}

// NewEnhancedProducerWithConnection creates a new enhanced producer using an existing connection
func NewEnhancedProducerWithConnection(conn *amqp.Connection, config ProducerConfig) (*EnhancedProducer, error) {
	producer, err := NewProducerWithConnection(conn, config)
	if err != nil {
		return nil, err
	}

	return &EnhancedProducer{
		Producer: producer,
	}, nil
}

// PublishWithHeaders publishes a message with custom headers and priority
func (p *EnhancedProducer) PublishWithHeaders(
	ctx context.Context,
	exchange string,
	routingKey string,
	headers amqp.Table,
	priority uint8,
	body []byte,
) error {
	for i := 0; i < p.config.RetryCount; i++ {
		err := p.publishWithHeaders(ctx, exchange, routingKey, headers, priority, body)
		if err == nil {
			return nil
		}
		time.Sleep(p.config.RetryInterval)
		p.reconnect()
	}
	return errors.New("max retry attempts reached")
}

func (p *EnhancedProducer) publishWithHeaders(
	ctx context.Context,
	exchange string,
	routingKey string,
	headers amqp.Table,
	priority uint8,
	body []byte,
) error {
	err := p.channel.PublishWithContext(
		ctx,
		exchange,
		routingKey,
		true,  // mandatory
		false, // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/octet-stream",
			Body:         body,
			Headers:      headers,
			Priority:     priority,
		},
	)
	if err != nil {
		return err
	}

	select {
	case confirmed := <-p.confirmsChan:
		if !confirmed.Ack {
			return errors.New("message not acknowledged by broker")
		}
	case <-time.After(5 * time.Second):
		return errors.New("confirm timed out")
	}

	return nil
}

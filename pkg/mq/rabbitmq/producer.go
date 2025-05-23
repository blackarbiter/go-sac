package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Producer struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	confirmsChan <-chan amqp.Confirmation
	config       ProducerConfig
}

type ProducerConfig struct {
	URL           string
	RetryCount    int
	RetryInterval time.Duration
}

// NewProducer creates a new producer using URL
func NewProducer(config ProducerConfig) (*Producer, error) {
	p := &Producer{config: config}
	if err := p.connect(); err != nil {
		return nil, err
	}
	return p, nil
}

// NewProducerWithConnection creates a new producer using an existing connection
func NewProducerWithConnection(conn *amqp.Connection, config ProducerConfig) (*Producer, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Enable publisher confirms
	if err := channel.Confirm(false); err != nil {
		channel.Close()
		return nil, fmt.Errorf("failed to enable confirm mode: %w", err)
	}

	p := &Producer{
		conn:         conn,
		channel:      channel,
		confirmsChan: channel.NotifyPublish(make(chan amqp.Confirmation, 1)),
		config:       config,
	}

	return p, nil
}

func (p *Producer) connect() error {
	var err error
	if p.conn, err = amqp.Dial(p.config.URL); err != nil {
		return err
	}

	if p.channel, err = p.conn.Channel(); err != nil {
		return err
	}

	// Enable publisher confirms
	if err := p.channel.Confirm(false); err != nil {
		return err
	}
	p.confirmsChan = p.channel.NotifyPublish(make(chan amqp.Confirmation, 1))

	return nil
}

func (p *Producer) Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	for i := 0; i < p.config.RetryCount; i++ {
		err := p.publish(ctx, exchange, routingKey, body)
		if err == nil {
			return nil
		}
		time.Sleep(p.config.RetryInterval)
		p.reconnect()
	}
	return errors.New("max retry attempts reached")
}

func (p *Producer) publish(ctx context.Context, exchange, routingKey string, body []byte) error {
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
			Headers:      amqp.Table{},
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

func (p *Producer) reconnect() {
	p.conn.Close()
	p.connect()
}

func (p *Producer) Close() error {
	if p.channel != nil {
		p.channel.Close()
	}
	return p.conn.Close()
}

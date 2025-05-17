package rabbitmq

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type DLXHandler struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  DLXConfig
}

type DLXConfig struct {
	URL        string
	QueueName  string
	Exchange   string
	RoutingKey string
}

func NewDLXHandler(config DLXConfig) (*DLXHandler, error) {
	d := &DLXHandler{config: config}
	if err := d.connect(); err != nil {
		return nil, err
	}
	return d, nil
}

func (d *DLXHandler) connect() error {
	var err error
	if d.conn, err = amqp.Dial(d.config.URL); err != nil {
		return err
	}

	if d.channel, err = d.conn.Channel(); err != nil {
		return err
	}

	// Declare DLX queue
	_, err = d.channel.QueueDeclare(
		d.config.QueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// Bind DLX queue to exchange
	err = d.channel.QueueBind(
		d.config.QueueName,
		d.config.RoutingKey,
		d.config.Exchange,
		false,
		nil,
	)
	return err
}

func (d *DLXHandler) ProcessDLX(handler func(amqp.Delivery)) error {
	msgs, err := d.channel.Consume(
		d.config.QueueName,
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

	go func() {
		for msg := range msgs {
			handler(msg)
			msg.Ack(false)
		}
	}()

	return nil
}

func (d *DLXHandler) Close() error {
	if d.channel != nil {
		d.channel.Close()
	}
	return d.conn.Close()
}

package broker

import (
	"fmt"

	"github.com/streadway/amqp"
)

type RabbitMQBroker struct {
	conn *amqp.Connection
}

func New(url string) (*RabbitMQBroker, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	return &RabbitMQBroker{conn: conn}, nil
}

func (b *RabbitMQBroker) DeclareQueue(queueName string) error {
	ch, err := b.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}
	defer func(ch *amqp.Channel) {
		err = ch.Close()
		if err != nil {
			return
		}
	}(ch)

	_, err = ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare a queue: %w", err)
	}

	return nil
}

func (b *RabbitMQBroker) Publish(queueName string, message []byte) error {
	ch, err := b.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}
	defer func(ch *amqp.Channel) {
		err = ch.Close()
		if err != nil {
			return
		}
	}(ch)

	err = ch.Publish(
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        message,
		})
	if err != nil {
		return fmt.Errorf("failed to publish a message: %w", err)
	}

	return nil
}

func (b *RabbitMQBroker) Consume(queueName string) (<-chan amqp.Delivery, error) {
	ch, err := b.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	msgs, err := ch.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register a consumer: %w", err)
	}

	return msgs, nil
}

func (b *RabbitMQBroker) Close() error {
	return b.conn.Close()
}

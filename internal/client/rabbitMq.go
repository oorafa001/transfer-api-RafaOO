package clients

import (
	"fmt"
	"time"
	"transfers-api/internal/config"
	"transfers-api/internal/logging"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQClient struct {
	conn  *amqp.Connection
	queue string
}

func NewRabbitMQClient(cfg config.RabbitMQ) *RabbitMQClient {
	conn, err := amqp.Dial(
		fmt.Sprintf(
			"amqp://%s:%s@%s:%d/",
			cfg.Username,
			cfg.Password,
			cfg.Hostname,
			cfg.Port,
		),
	)
	if err != nil {
		logging.Logger.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	return &RabbitMQClient{
		conn:  conn,
		queue: cfg.QueueName,
	}
}

func NewClientRabbitMQClient(cfg config.RabbitMQ) *RabbitMQClient {
	conn, err := amqp.Dial(
		fmt.Sprintf(
			"amqp://%s:%s@%s:%d/",
			cfg.Username,
			cfg.Password,
			cfg.Hostname,
			cfg.Port,
		),
	)
	if err != nil {
		logging.Logger.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	return &RabbitMQClient{
		conn:  conn,
		queue: cfg.QueueName,
	}
}

func (c *RabbitMQClient) Publish(operation string, transferID string) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(
		c.queue,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	body := fmt.Sprintf("%s:%s", operation, transferID)
	err = ch.Publish(
		"",      // exchange
		c.queue, // routing key (queue name)
		false,   // mandatory
		false,   // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	return nil
}

func (c *RabbitMQClient) Listener() error {

	ch, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(
		c.queue,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	msgs, err := ch.Consume(
		c.queue,
		"",
		false, // manual ack
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logging.Logger.Fatalf("consume error: %v", err)
	}

	fmt.Println("Consumer listening...")

	for msg := range msgs {
		fmt.Printf("Message: %s\n", string(msg.Body))

		// Simula procesamiento
		time.Sleep(500 * time.Millisecond)

		msg.Ack(false)
	}

	return fmt.Errorf("channel closed")
}

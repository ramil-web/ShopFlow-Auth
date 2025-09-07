package services

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type EmailService struct {
	MQConn *amqp.Connection
}

type EmailMessage struct {
	Email    string `json:"email"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

// SendEmailToQueue (можно использовать для других сервисов при необходимости)
func (s *EmailService) SendEmailToQueue(email, login, password string) error {
	ch, err := s.MQConn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open RabbitMQ channel: %w", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"email_queue",
		true, false, false, false, nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	msg := EmailMessage{Email: email, Login: login, Password: password}
	body, _ := json.Marshal(msg)

	if err := ch.Publish("", q.Name, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	}); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

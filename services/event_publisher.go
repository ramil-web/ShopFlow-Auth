package services

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type EventPublisher struct {
	MQConn *amqp.Connection
}

func NewEventPublisher(conn *amqp.Connection) *EventPublisher {
	return &EventPublisher{MQConn: conn}
}

func (p *EventPublisher) PublishUserRegistered(userID uint, email, login, password string) error {
	ch, err := p.MQConn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	msg := map[string]interface{}{
		"user_id":  userID,
		"email":    email,
		"login":    login,
		"password": password,
	}
	body, _ := json.Marshal(msg)

	exchange := "shopflow.events"   // общий exchange для всех событий
	routingKey := "user.registered" // routing key, на который подписан воркер

	if err := ch.Publish(exchange, routingKey, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	}); err != nil {
		return err
	}

	log.Println("[INFO] user.registered event published for user:", userID)
	return nil
}

package worker

import (
	"auth/services"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type UserRegisteredEvent struct {
	UserID   uint   `json:"user_id"`
	Email    string `json:"email,omitempty"`
	Login    string `json:"login,omitempty"`
	Password string `json:"password,omitempty"`
}

func StartUserRegisteredWorker(conn *amqp.Connection, emailService *services.EmailService) {
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("user.registered: open channel:", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare("user.registered.queue", true, false, false, false, nil)
	if err != nil {
		log.Fatal("user.registered: declare queue:", err)
	}

	if err := ch.QueueBind(q.Name, "user.registered", "auth.events", false, nil); err != nil {
		log.Fatal("user.registered: bind:", err)
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Fatal("user.registered: consume:", err)
	}

	go func() {
		for d := range msgs {
			var evt UserRegisteredEvent
			if err := json.Unmarshal(d.Body, &evt); err != nil {
				log.Println("user.registered: invalid payload:", err)
				continue
			}

			email := evt.Email
			if email == "" {
				email = lookupUserEmailViaBus(evt.UserID, conn)
			}

			subject := "Welcome to Shopflow"
			body := fmt.Sprintf("Hello %s,\nYour account was created. Password: %s", evt.Login, evt.Password)

			emailService.SendEmailToQueue(email, subject, body)
		}
	}()

	log.Println("Notification worker (user.registered) started")
}

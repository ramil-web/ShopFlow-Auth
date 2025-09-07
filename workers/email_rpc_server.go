package worker

import (
	"auth/models"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/gorm"
)

type UserEmailRequest struct {
	UserID uint `json:"user_id"`
}

type UserEmailResponse struct {
	Email string `json:"email"`
}

func StartEmailRPCServer(conn *amqp.Connection, db *gorm.DB) {
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("RPC: failed to open channel:", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare("auth.user.email.rpc", false, false, false, false, nil)
	if err != nil {
		log.Fatal("RPC: declare queue:", err)
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Fatal("RPC: consume:", err)
	}

	go func() {
		for d := range msgs {
			var req UserEmailRequest
			if err := json.Unmarshal(d.Body, &req); err != nil {
				log.Println("RPC: invalid request:", err)
				continue
			}

			var user models.User
			if err := db.First(&user, req.UserID).Error; err != nil {
				log.Println("RPC: user not found:", err)
				continue
			}

			resp := UserEmailResponse{Email: user.Email}
			body, _ := json.Marshal(resp)

			err = ch.Publish("", d.ReplyTo, false, false, amqp.Publishing{
				ContentType:   "application/json",
				CorrelationId: d.CorrelationId,
				Body:          body,
			})
			if err != nil {
				log.Println("RPC: failed to reply:", err)
			}
		}
	}()

	log.Println("Auth RPC server (auth.user.email.rpc) started")
}

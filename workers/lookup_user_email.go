package worker

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type UpUserEmailRequest struct {
	UserID uint `json:"user_id"`
}
type UpUserEmailResponse struct {
	Email string `json:"email"`
}

func lookupUserEmailViaBus(userID uint, conn *amqp.Connection) string {
	ch, err := conn.Channel()
	if err != nil {
		log.Println("RPC client: channel error:", err)
		return ""
	}
	defer ch.Close()

	replyQueue, err := ch.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		log.Println("RPC client: reply queue error:", err)
		return ""
	}

	corrID := fmt.Sprintf("rpc-%d", userID)
	req := UserEmailRequest{UserID: userID}
	body, _ := json.Marshal(req)

	err = ch.Publish("", "auth.user.email.rpc", false, false, amqp.Publishing{
		ContentType:   "application/json",
		CorrelationId: corrID,
		ReplyTo:       replyQueue.Name,
		Body:          body,
	})
	if err != nil {
		log.Println("RPC client: publish error:", err)
		return ""
	}

	msgs, _ := ch.Consume(replyQueue.Name, "", true, true, false, false, nil)
	for d := range msgs {
		if d.CorrelationId == corrID {
			var resp UserEmailResponse
			if err := json.Unmarshal(d.Body, &resp); err != nil {
				log.Println("RPC client: invalid response:", err)
				return ""
			}
			return resp.Email
		}
	}
	return ""
}

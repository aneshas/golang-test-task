package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	r := gin.Default()

	conn, cleanup := conn()
	defer cleanup()

	r.POST("/message", newMessageHandler(conn))

	log.Fatal(r.Run(":8081"))
}

type httpError struct {
	Error string `json:"error"`
}

type message struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Message  string `json:"message"`
}

func newMessageHandler(conn *amqp.Connection) gin.HandlerFunc {
	ch, err := conn.Channel()
	check(err)

	q, err := ch.QueueDeclare(
		"messages",
		true,
		false,
		false,
		false,
		nil,
	)

	check(err)

	return func(c *gin.Context) {
		var msg message

		err := c.BindJSON(&msg)
		if err != nil {
			c.JSON(http.StatusBadRequest, httpError{Error: err.Error()})
			return
		}

		data, _ := json.Marshal(msg)

		err = ch.PublishWithContext(
			c.Request.Context(),
			"",
			q.Name,
			false,
			false,
			amqp.Publishing{
				DeliveryMode: amqp.Persistent,
				ContentType:  "application/json",
				Body:         data,
			})
		if err != nil {
			c.JSON(http.StatusBadRequest, httpError{Error: err.Error()})
			return
		}

		c.Status(http.StatusOK)
	}
}

func conn() (*amqp.Connection, func()) {
	conn, err := amqp.Dial("amqp://user:password@rabbitmq:7001/")
	check(err)

	return conn, func() {
		_ = conn.Close()
	}
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"os/signal"
)

func main() {
	conn, cleanup := conn()
	defer cleanup()

	db := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	defer db.Close()

	p := newProcessor(conn, db)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go p.Run(ctx)

	ch := make(chan os.Signal, 1)

	signal.Notify(ch, os.Interrupt)

	<-ch
}

type message struct {
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Message  string `json:"message"`
}

func newProcessor(conn *amqp.Connection, db *redis.Client) *processor {
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

	return &processor{
		ch: ch,
		q:  q,
		db: db,
	}
}

type processor struct {
	ch *amqp.Channel
	q  amqp.Queue
	db *redis.Client
}

func (p *processor) Run(ctx context.Context) {
	messages, err := p.ch.Consume(
		p.q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	check(err)

	log.Println("waiting for messages...")

	for {
		select {
		case msg := <-messages:
			fmt.Printf("<- Got a message: %s\n", msg.Body)

			var m message

			err := json.Unmarshal(msg.Body, &m)
			if err != nil {
				log.Println("error unmarshalling message: ", err)
				break
			}

			cmd := p.db.LPush(ctx, fmt.Sprintf("%s:%s", m.Sender, m.Receiver), msg.Body)

			if cmd.Err() != nil {
				log.Println("error saving message: ", err)
				break
			}

			err = msg.Ack(false)
			if err != nil {
				log.Println("Ack error: ", err)
			}

		case <-ctx.Done():
			log.Println("Consumer exiting...")

			return
		}
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

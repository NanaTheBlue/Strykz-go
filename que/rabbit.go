package que

import (
	"context"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Rabbit interface {
	Publish(ctx context.Context, queue string, body []byte) error
	Close()
}

type rabbit struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewPublisher(uri string) (Rabbit, error) {
	conn, err := amqp.Dial(uri)
	failOnError(err, "Failed to connect to RabbitMQ")
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	fmt.Println("We Connected To Rabbit!!!!")
	return &rabbit{conn: conn, ch: ch}, nil
}

func (r *rabbit) Publish(ctx context.Context, queue string, body []byte) error {
	if _, err := r.ch.QueueDeclare(queue, true, false, false, false, nil); err != nil {
		return err
	}
	return r.ch.PublishWithContext(ctx, "", queue, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        body,
	})
}
func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func (r *rabbit) Close() {
	r.ch.Close()
	r.conn.Close()
}

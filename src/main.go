package main

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

func main() {
	server()
}

func server() {
	conn, ch, q := getQueue()
	defer conn.Close()
	defer ch.Close()

	msg := amqp.Publishing{
		ContentType: "text/plain", // rabbitmq sends data alwass as byte trains
		Body:        []byte("Hello RabbitMQ"),
	} // messages lifecycle

	ch.Publish("", q.Name, false, false, msg) //"" does not have a specific name so the default exchange is used
}

func getQueue() (*amqp.Connection, *amqp.Channel, *amqp.Queue) {
	conn, err := amqp.Dial("amqp://guest@localhost:5672")
	failOnError(err, "Failed to connect to rabbitMQ")
	ch, err := conn.Channel() // this is prone to error as with regular networks
	failOnError(err, "Failed to open a channel")
	q, err := ch.QueueDeclare("hello", // queue name
		false, // durable. determines if the messages should be saved to disk after they are added to the queue. With this options set, the messages will survive a server restart. Messages that are process is decreased dramatically.
		false, // autoDelete, tells what to do if there are no active consumers. If this is true message will be deleted from the queue. False prompts it to hang around till the consumer comes and receives it.
		false, // exclusive flag allows us to set this queue to be only accessible from the connection that requests it. If a request is made to get a queue with the same nameand configuration from another channel, then it will receive an error if this is true.
		// otherwise it will receive a connection to the same queue and the two connections will share it
		false, // nowait flag instructs rabbitmq to only return a pre-existing queue that matches the provided configuration. If a properly configured queue is not found on the server then the channel will receive an error
		// flase because there isn't any on the server
		nil) // args will be used on certain scenarious such as the headers that will bematches by this queue if it is bound to a headers exchange.
	failOnError(err, "Failed to declare a queue")

	return conn, ch, &q

} // this is now autamitacally bound to a default exchange. If it succeeds.
// any messages with "hello" routing key will be sent to this queue.
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

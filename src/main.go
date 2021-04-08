package main

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

func main() {
	go client() // this is now a gorutine
	go server() // the server is also now a gorutine
	// to allow both functions to max temselves out
	var a string
	fmt.Scanln(&a)
}

func client() { // receive messages.
	conn, ch, q := getQueue()
	defer conn.Close() // close connection and channel when done with them
	defer ch.Close()

	msgs, err := ch.Consume(q.Name, // name of the queue to get messages from, No exchange as excahnges are used only to publish messages to rabbitmq
		"", // this idendities the connection to the queue uniquely. To determine who is listerning on the queue. This is importan when multiple clients and receiving messags from the same queue in a dircet exchange.
		// rabbitmq will distribute the messages amongs the clients so that they are roughly load balanced.
		// It can also be important if a client needs to cancel its connection since that lets rabbitmq know it should no longer try to send messages to that specific receiver.
		// passing in an empty string rabbitmq will assign a value by itself since there is no specific need to track it
		true, // autoAck flag indicates if we want to automatically acknowledge the sucessful receipt of a message
		// Set to true normally so the server an reeive messages a quiicly as possible to conserve resources.
		// If the call also updates a record in a database. False due to manually updating it. rabbitmq as a repository o ensure that entire transaciton is done and messaes are not lost before the entire trasaction is done
		false, // exclusive flag is used to make sure the clien is the only consumer for this queue. If this is true an error will occur if other clients are already registered or if another client tries to listen to this queue later.
		false, // noLocal flag prevets Rabbitmq fron sending messages to clients that are on the same connectino as the sender.
		false, // nowait flag instructs rabbitmq to only return a pre-existing queue that matches the provided configuration. If a properly configured queue is not found on the server then the channel will receive an error
		// flase because there isn't any on the server
		nil)
	failOnError(err, "failed to register a consumer")

	for msg := range msgs {
		log.Printf("Received message with message: %s", msg.Body)
	}
}

func server() {
	conn, ch, q := getQueue()
	defer conn.Close()
	defer ch.Close()

	msg := amqp.Publishing{
		ContentType: "text/plain", // rabbitmq sends data alwass as byte trains
		Body:        []byte("Hello RabbitMQ"),
	} // messages lifecycle
	//for {// generates as fast as possible and gives a good benchmark

	ch.Publish("", q.Name, false, false, msg) //"" does not have a specific name so the default exchange is used
	//}use a VM for such problems
	// not a persistence strate to rely on
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

package qutils

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

const SensorListQueue = "SensorList"
const SensorDiscoveryExchange = "SensorDiscovery"

func GetChannel(url string) (*amqp.Connection, *amqp.Channel) { // will return pointers for connection and channel
	conn, err := amqp.Dial(url)
	failOnError(err, "Failed to establish connection o message broker")
	ch, err := conn.Channel()
	failOnError(err, "Failed to get channel for connection")

	return conn, ch
}

func GetQueue(name string, ch *amqp.Channel, autodelete bool) *amqp.Queue { // retursn a queue to get a refrence to a queue object
	q, err := ch.QueueDeclare(
		name,
		false,      // durable not, because it may not be too critical if one of the brokern messages will go down
		autodelete, // autoDelete false as the sensor readings would be deleted if there is noo active listeners
		// this might happen if we are patching the coordinators
		false, // this is the only connection o be able to work with this queue
		// since we need the sensor and coordinators to be able to access the queue we will use false here as well
		false, // the noWait flag is used to send the message without waiting to see if the queue needs to be set up
		// if one is not ready the message fails, since it expects each sensor to set up its own queue. So this is false right now.
		nil) // the default exchange that does not need any additionaly configuration information. So the value nil is enough right now.

	failOnError(err, "Failed to declare queue")

	return &q
}

func failOnError(err error, msg string) { // package level failOnError function
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

// We must remember to delete the queue after each time the queue is tempraryly used by the sensor.
// Although there is a more easier wa to do this.

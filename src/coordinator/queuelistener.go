package coordinators

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"rabbitmq_go/src/dto"
	"rabbitmq_go/src/qutils"

	"github.com/streadway/amqp"
)

// the coordiantors are the second application that I'm going to build on this system.
// They have several responsibilities. The first of which is to interact with the sensors via the
// message broker.

const url = "amqp://guest:guest@localhost:5672"

type QueueListener struct { // this will discover the data queues. Discover the messags and eventually translate them into events in an EventAggregator
	conn    *amqp.Connection
	ch      *amqp.Channel
	sources map[string]<-chan amqp.Delivery // registry of all of the sources the coordinator is listening on.The first is that we are eventually going to be able to pull the sensor and have them publish their queue names. By storing the sources, we'll be able to check if we've already registered each sensor so that we don't register it twice. The other reason that I can think of is to close down listeners if the associated sensor goes off-line. I'm not planning on implementing that in this course, but it should be pretty easy to set this up on your own.
	ea      *EventAggregator
}

func NewQueueListener() *QueueListener { // constructor function
	ql := QueueListener{
		sources: make(map[string]<-chan amqp.Delivery),
		ea:      NewEventAggregator(),
	}

	ql.conn, ql.ch = qutils.GetChannel(url)

	return &ql
}

func (ql *QueueListener) ListenForNewSource() { // ql.ch <- object declaration
	q := qutils.GetQueue("", ql.ch) // RabbitMQ will see that there is no name for the queue and will create the name by itself. So there are no conflicts with queue naming.
	// By default when queues are created. They are bound to the default exchange.
	// Now it will be he fanout exchange. So we will need o rebind it again.
	ql.ch.QueueBind(
		q.Name,        // So it now known to which queue to bind to.
		"",            // routing key now for fanout exchanges ignore this field
		"amqp.fanout", // messages are coming in from here
		false,
		nil)

	msgs, _ := ql.ch.Consume( // receive the messages that are sent to it.
		q.Name, // queue sring,
		"",     // consumer string,
		true,   // autoAck bool,
		false,  // exclusive bool,
		false,  // noLocal bool,
		false,  // noWait bool,
		nil)    // args amqp.table)

	for msg := range msgs { // this now indicates that a new sensor has come online and it is ready to send readings to the system
		sourceChan, _ := ql.ch.Consume( // channel consumer method to get access to that sensors queue
			string(msg.Body), // queue string,
			"",               // consumer string,
			true,             // autoAck bool,
			false,            // exclusive bool,
			false,            // noLocal bool,
			false,            // noWait bool,
			nil)              // args amqp.Table
		// This will wait for messages to come in to the messages channel
		// When a message does come in. It is going to indicate that a new sensor has come online
		// And is ready to send readings into the system.
		// ql.ch.Consume method to get access that sensors queue

		if ql.sources[string(msg.Body)] == nil {
			ql.sources[string(msg.Body)] = sourceChan

			go ql.AddListener(sourceChan) // this will listen the messages from the channel
		}
	}
}

func (ql *QueueListener) AddListener(msgs <-chan amqp.Delivery) { // receive only channel that accepts message
	for msg := range msgs {
		r := bytes.NewReader(msg.Body)
		d := gob.NewDecoder(r)
		sd := new(dto.SensorMessage)
		d.Decode(sd)

		fmt.Printf("Received message: %v\n", sd)

		ed := EventData{ // event data object and populate it
			Name:      sd.Name,
			Timestamp: sd.Timestamp,
			Value:     sd.Value,
		}

		ql.ea.PublishEvent("MessageReceived_"+msg.routingkey, ed) // so that each event is registered uniquely
	}
}

// Currently, our sensors publish the name of their data queue as soon as they come online. They are publishing to a fanout exchange, which is perfect for informing a lot of coordinators at the same time. However, if a coordinator starts up after a sensor, it currently has no way of knowing about it. To address this, I want to create a new exchange in the message broker. This will be a fanout exchange that is going to work in the opposite direction of the other message flows. In this case, a coordinator will make a discovery request to the exchange. That message will fanout to all of the sensors that will respond by publishing their data queue's name to the fanout exchange that we've been using. This setup is going to get some messages to the coordinators that aren't interested in it. But we're already filtering out redundant queue names in the coordinators, so I think that this will be just fine.

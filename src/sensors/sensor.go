package main

import (
	// generating simulated data
	"bytes"
	"encoding/gob"
	"flag"
	"log"
	"math/rand"
	"rabbitmq_go/src/dto"
	"rabbitmq_go/src/qutils"
	"strconv"
	"time"

	"github.com/streadway/amqp"
)

var url = "amqp://guest:guest@localhost:5672"

var name = flag.String("name", "sensor", "name of the sensor")
var freq = flag.Uint("freq", 5, "update frequency in cucles/sec")
var max = flag.Float64("max", 5., "maximum value for generated readings")
var min = flag.Float64("max", 1., "minimum value for generated readings")
var stepSize = flag.Float64("step", 0.1, "maximum allowable change per measurement")

var r = rand.New(rand.NewSource(time.Now().UnixNano())) // UnixNano is great to work with nanosecond scales

var value = r.Float64()*(*max-*min) + *min
var nom = (*max-*min)/2 + *min // nominal sensor value

func main() {
	flag.Parse()

	conn, ch := qutils.GetChannel(url)
	defer conn.Close() // defered calls to close the connection when we are doe with it
	defer ch.Close()   // defered calls to close the connection when we are doe with it

	dataQueue := qutils.GetQueue(*name, ch)
	sensorQueue := qutils.GetQueue(qutils.SensorListQueue, ch)

	msg := amqp.Publishing{
		Body: []byte(*name),
	}
	ch.Publish( // this will publish if new sensor will come online.
		"",
		sensorQueue.Name,
		false,
		false,
		msg)

	// 5 cycles/ sec = 200 milliseconds / cycle
	dur, _ := time.ParseDuration(strconv.Itoa(1000/int(*freq)) + "ms")

	signal := time.Tick(dur)

	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)

	for range signal {
		calcValue()
		reading := dto.SensorMessage{
			Name:      *name,
			Value:     value,
			Timestamp: time.Now(),
		}
		// now the messages is prepared to be encoded for transmission
		buf.Reset() // any initial data is removed and buffer pointer is set to the initial position
		enc.Encode(reading)

		msg := amqp.Publishing{
			Body: buf.Bytes(),
		}

		ch.Publish(
			"",             // name of the exchange, we are using the default direct exchange so no strings are required
			dataQueue.Name, // routing key. reference to the dataQueue.Name is the more correct way. Key is the routing key for the queue
			false,          // this will throw an error if there is no queue defined to receive he publishing
			false,          // the immediate argument determines if the publish argument should trigger a failure of there aren't any consumers currently on this queue
			// This can be set to false for two reasons. First we want rabbitmq to set things up if the coordinators are not running for any reason; and second; I could technically receive messages through the reference that I created earlier, so there is a consumer as far as rabbitmq is concerned
			msg)
		log.Printf("Reading sent. Value: %v\n", value)
	}
}

func calcValue() {
	// every value is gonna change with evenry time step by a maximum of the step size up here

	var maxStep, minStep float64 // These will hold the maximum amount that the value can increase or decrease in this step

	if value < nom {
		maxStep = *stepSize
		minStep = -1 * *stepSize * (value - *min) / (nom - *min) // range between nominal value and the minimum limit
	} else {
		maxStep = *stepSize * (*max - value) / (*max - nom)
		minStep = -1 * *stepSize
	}

	value += r.Float64()*(maxStep-minStep) + minStep
}

// go run sensor.go --help
// go run src\distributed\sensors\sensor\sensor.goz
// remember RabbitMQ receives messages and exchanges and then uses their configuration and information within the message to
// determine which queues to deliver to. That means that technically we only need to provide the queue's name or routing key when publishing
// However, this is not enough to ensure that the queue with that routingkey actually exists
// By declaring the queue here we can be sure that RabbitMQ has set it up properly and it will be ready for us to use
// By sending the message off.
// To kick the sensor off.

// Is to handle the issue of letting the coordinators know when a sensor comes online and starts to send readings.
// Since each sensor will create a new queue it will be impossible for the coordinators to efficiently discover them
// without a little bit of help. The key to having dynamic queue names in this application is that I will have one queue that is
// well known thoughout the entire applicaiton and can be used to send the name of each queue as it is created.

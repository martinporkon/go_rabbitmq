package main

import (
	// generating simulated data
	"bytes"
	"encoding/gob"
	"flag"
	"log"
	"math/rand"
	"rabbitmq_go/src/dto"
	"strconv"
	"time"
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

package dto

import (
	"encoding/gob"
	"time"
)

type SensorMessage struct {
	Name      string
	Value     float64
	Timestamp time.Time // lag between mesaurements, to be sure when the measurement was actually taken
	// json or protocol buffers
}

func init() {
	gob.Register(SensorMessage{}) // That way every consumer. Can rely on the sensormessage object being ready to send over the wire.
}

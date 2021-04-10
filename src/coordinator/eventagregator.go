package main

import "time"

type EventAggregator struct {
	listeners map[string][]func(interface{})
	// The way that is going to work is that I'll expose a method that will allow anything to register as an event listener. The way that it will do that is by providing the event that it's interested in, and the callback function that will handle the event. When an event does occur, the EventAggregator will loop through each of the listeners registered for the event and call their callbacks in turn.
}

type EventRaiser interface {
	AddListener(eventName string, f func(interface{})) // empty interface so we can send than more amount of one type of data to the listeners
}

func NewEventAggregator() *EventAggregator {
	ea := EventAggregator{
		listeners: make(map[string][]func(interface{})),
	}

	return &ea
}

func (ea *EventAggregator) AddListener(name string, f func(interface{})) {
	ea.listeners[name] = append(ea.listeners[name], f)
}

func (ea *EventAggregator) PublishEvent(name string, eventData interface{}) { // more general events to the listeners
	if ea.listeners[name] != nil {
		for _, r := range ea.listeners[name] {
			r(eventData) // not a pointer to it, but the whole message itself. This is going to trigger a copy of the object to be sent to each consumer which will ensure that they don't start fiddling around with the data and confuse one another
		}
	}
}

// TODO Unsubscribe from an event

type EventData struct {
	Name      string // sensor name
	Value     float64
	Timestamp time.Time
}

//

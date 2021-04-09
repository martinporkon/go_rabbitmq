package main

import (
	"fmt"
	coordinators "rabbitmq_go/src/coordinator"
)

func main() {
	ql := coordinators.NewQueueListener()
	go ql.ListenForNewSource()

	var a string
	fmt.Scanln(&a) // this will keep the applicaton alive as the queuelistener does it stuff
}

// go run src\coordinator\exec\main.go

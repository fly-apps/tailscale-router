package main

import (
	"log"
	"os"
	"os/signal"
)

func main() {
	log.Println("swarm: TODO")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
}

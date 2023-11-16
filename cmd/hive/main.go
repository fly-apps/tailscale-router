package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/ananthb/tailscale-router/internal/handlers"
)

func main() {
	log.SetPrefix("hive ")
	h := &handlers.H{
		FlyMachinesAPI: os.Getenv("FLY_MACHINES_API_HOSTNAME"),
		FlyAPIToken:    os.Getenv("FLY_API_TOKEN"),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/bees", h.ListBees)

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	log.Printf("server listening at %s\n", server.Addr)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
}

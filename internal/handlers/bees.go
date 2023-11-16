package handlers

import (
	"net/http"
)

type H struct {
	FlyMachinesAPI string
	FlyAPIToken    string
}

func (h H) ListBees(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("bees!"))
}

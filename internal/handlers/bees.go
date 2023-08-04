package handlers

import (
	"net/http"
)

type H struct {
	APIHostname string
	APIToken    string
}

func (h *H) ListBees(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("bees!"))
}

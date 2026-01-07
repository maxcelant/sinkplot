package routes

import (
	"fmt"
	"net/http"
)

// Reverse proxying is just a handler
type Handler struct {
	Upstreams []string
	Matchers  MatcherList
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle roundtrip of request
	msg := fmt.Sprintf("proxy handler: upstreams: %v", h.Upstreams)
	w.Write([]byte(msg))
}

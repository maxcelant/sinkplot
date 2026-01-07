package routes

import (
	"math/rand"
	"net/http"
)

type Upstreams struct {
	SocketAddrs []string
	Strategy    LoadbalanceStrategy
}

// Reverse proxying is just a handler
type Handler struct {
	Upstreams Upstreams
	Matchers  MatcherList
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Choose an upstream using a strategy
	// Create a new request
	// run RoundTrip
	// Forward response
}

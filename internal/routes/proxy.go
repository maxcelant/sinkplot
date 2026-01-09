package routes

import (
	"io"
	"net/http"
)

type Upstreams struct {
	SocketAddrs []string
	Strategy    LoadbalanceStrategy
}

type Transport struct {
	http.RoundTripper
}

// Reverse proxying is just a handler
type Handler struct {
	Upstreams Upstreams
	Matchers  MatcherList
	Transport Transport
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upstreamHost := h.Upstreams.Strategy.Pick()
	// Modify the original host with the chosen upstream
	r.URL.Host = upstreamHost
	// We need to set http or https on the request, only support http for now
	r.URL.Scheme = "http"
	res, err := h.Transport.RoundTrip(r)
	if err != nil {
		http.Error(w, "error occurred while performing roundtrip", http.StatusBadRequest)
		return
	}
	defer res.Body.Close()
	w.WriteHeader(http.StatusOK)
	io.Copy(w, res.Body)
}

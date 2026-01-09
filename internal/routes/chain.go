package routes

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/maxcelant/sinkplot/internal/schema"
)

func Compile(app schema.App) (http.Handler, error) {
	handlers := make([]Handler, len(app.Routes))
	for _, r := range app.Routes {
		i := slices.IndexFunc(app.Sinks, func(sink schema.Sink) bool {
			return sink.Name == r.Sink
		})
		if i == -1 {
			return nil, fmt.Errorf("failed to find sink with name '%s'", r.Sink)
		}
		addrs := make([]string, len(app.Sinks[i].Upstreams))
		for j, u := range app.Sinks[i].Upstreams {
			addrs[j] = fmt.Sprintf("%s:%d", u.Address, u.Port)
		}
		rh := Handler{
			Transport: Transport{
				RoundTripper: http.DefaultTransport,
			},
			Upstreams: Upstreams{
				Strategy: RandomStrategy{addrs},
			},
			Matchers: []Matcher{PathMatcher{Path: r.Path}},
		}
		handlers[i] = rh
	}
	// Chain the handlers together
	next := emptyHandler
	for _, r := range handlers {
		next = wrapRoutes(r)(next)
	}
	return next, nil
}

func wrapRoutes(rh Handler) Middleware {
	// We need to do it this way because we need to inject some the route matching context in the handler chain
	return func(next http.Handler) http.Handler {
		// This is where we need to actually do the routing matches
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If the request matches this handler, stop the chain and handle the request accordingly.
			if rh.Matchers.Match(*r) {
				rh.ServeHTTP(w, r)
				return
			}
			// Otherwise, just continue down the chain
			next.ServeHTTP(w, r)
		})
	}
}

package routes

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/maxcelant/sinkplot/internal/schema"
)

// Compile will create a handler chain based off of the given config schema
func Compile(app schema.App) (http.Handler, error) {
	handlers := make([]Handler, len(app.Routes))
	for _, r := range app.Routes {
		i := slices.IndexFunc(app.Sinks, func(sink schema.Sink) bool {
			return sink.Name == r.Sink
		})
		if i == -1 {
			return nil, fmt.Errorf("failed to find sink with name '%s'", r.Sink)
		}
		lbStrategy := pickLoadbalanceStrategy(app.Sinks[i].Upstreams)
		rh := Handler{
			Transport: Transport{
				RoundTripper: http.DefaultTransport,
			},
			Upstreams: Upstreams{
				Strategy: lbStrategy,
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
	// Wrap with logging middleware as the outermost layer
	return loggerRoute(next), nil
}

// pickLoadbalanceStrategy picks an appropriate loadbalancing strategy based on the fields in the Sinkfile
func pickLoadbalanceStrategy(upstreams []schema.Upstream) LoadbalanceStrategy {
	// TODO: revisit once defaulter and validator are complete
	addrs := make([]string, len(upstreams))
	for i, u := range upstreams {
		addrs[i] = fmt.Sprintf("%s:%d", u.Address, u.Port)
	}
	return RandomStrategy{addrs}
}

// wrapRoutes creates a handler chain to easily perform route matching
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

package config

import (
	"fmt"
	"net/http"
	"os"
	"slices"

	"gopkg.in/yaml.v3"
)

// Reverse proxying is just a handler
type RouteHandler struct {
	Upstreams []string
}

func (RouteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle roundtrip of request
	w.Write([]byte("proxy handler"))
}

func Load() (http.Handler, error) {
	var cfg V1_ConfigSchema
	buf, err := os.ReadFile("tests/sample_01.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	err = yaml.Unmarshal(buf, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config to yaml: %w", err)
	}
	handlers, err := compileRouteHandlers(cfg.App)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	chain := compileHandlerChain(handlers)
	return chain, nil
}

func compileRouteHandlers(app App) ([]RouteHandler, error) {
	handlers := make([]RouteHandler, len(app.Routes))
	for _, r := range app.Routes {
		i := slices.IndexFunc(app.Sinks, func(sink Sink) bool {
			return sink.Name == r.Sink
		})
		if i == -1 {
			return []RouteHandler{}, fmt.Errorf("failed to find sink with name '%s'", r.Sink)
		}
		upstreams := make([]string, len(app.Sinks[i].Upstreams))
		for _, u := range app.Sinks[i].Upstreams {
			upstreams[i] = u.Address
		}
		handlers[i] = RouteHandler{Upstreams: upstreams}
	}
	return handlers, nil
}

func compileHandlerChain(routes []RouteHandler) http.Handler {
	// Chain the handlers together
	next := emptyHandler
	for _, r := range routes {
		next = wrapRoutes(r)(next)
	}
	return next
}

// func wrapRoute(

func wrapRoutes(req RouteHandler) Middleware {
	// We need to do it this way because we need to inject some the route matching context in the handler chain
	return func(next http.Handler) http.Handler {
		// This is where we need to actually do the routing matches
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If the request matches this handler, stop the chain and handle the request accordingly.
			req.ServeHTTP(w, r)
			// Otherwise, just continue down the chain
			next.ServeHTTP(w, r)
		})
	}
}

package routes

import (
	"fmt"
	"net/http"
	"regexp"
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
		lbStrategy := compileRoutingStrategy(app.Sinks[i].Upstreams)
		matchers, err := compileMatchers(r)
		if err != nil {
			return nil, fmt.Errorf("failed to compile matchers: %w", err)
		}
		rh := Handler{
			Transport: Transport{
				RoundTripper: http.DefaultTransport,
			},
			Upstreams: Upstreams{
				Strategy: lbStrategy,
			},
			Matchers: matchers,
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

func compileMatchers(route schema.Route) (MatcherList, error) {
	var ml []Matcher
	switch route.Path {
	case "exact":
		ml = append(ml, PathMatcher{route.Path})
	case "prefix":
		ml = append(ml, PrefixMatcher{route.Path})
	case "regex":
		re, err := regexp.Compile(route.Path)
		if err != nil {
			return ml, fmt.Errorf("failed to turn route path into regex: %w", err)
		}
		ml = append(ml, RegexMatcher{re})
	}
	if route.Methods != nil && len(*route.Methods) != 0 {
		ml = append(ml, MethodMatcher{*route.Methods})
	}
	return ml, nil
}

// compileRoutingStrategy picks an appropriate loadbalancing strategy based on the fields in the Sinkfile
func compileRoutingStrategy(upstreams []schema.Upstream) LoadbalanceStrategy {
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

package config

import (
	"fmt"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"
)

func Load() {
	var cfg V1_ConfigSchema
	buf, err := os.ReadFile("tests/sample_01.yaml")
	if err != nil {
		fmt.Println(err)
	}
	err = yaml.Unmarshal(buf, &cfg)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(cfg.App.Name)
	compileHandlerChain(cfg.App.Routes)
}

func compileHandlerChain(routes []Route) http.Handler {
	// Turn the routes into http.Handlers
	var handlers []MiddlewareHandler
	for _, route := range routes {
		handlers = append(handlers, MiddlewareHandlerFunc(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
			w.Write([]byte(route.Name))
			next.ServeHTTP(w, r)
		}))
	}

	// Chain them handlers together
	var next http.Handler
	for i, h := range handlers {
		next = wrapRoute(h, next)(routes[i])
	}
	return next
}

func wrapRoute(h MiddlewareHandler, next http.Handler) func(Route) http.Handler {
	// We need to do it this way because we need to inject some the route matching context in the handler chain
	return func(r Route) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r, next)
		})
	}
}

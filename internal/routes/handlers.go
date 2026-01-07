package routes

import "net/http"

type Middleware func(http.Handler) http.Handler

type MiddlewareHandler interface {
	ServeHTTP(http.ResponseWriter, *http.Request, http.Handler)
}

var emptyHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("No matching route found")) })

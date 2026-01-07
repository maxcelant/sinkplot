package config

import "net/http"

type Middleware func(http.Handler) http.Handler

type MiddlewareHandler interface {
	ServeHTTP(http.ResponseWriter, *http.Request, http.Handler)
}

type MiddlewareHandlerFunc func(http.ResponseWriter, *http.Request, http.Handler)

func (m MiddlewareHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request, h http.Handler) {
	m.ServeHTTP(w, r, h)
}

var emptyHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("No matching route found")) })

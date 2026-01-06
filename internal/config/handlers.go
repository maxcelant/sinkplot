package config

import "net/http"

type MiddlewareHandler interface {
	ServeHTTP(http.ResponseWriter, *http.Request, http.Handler)
}

type MiddlewareHandlerFunc func(http.ResponseWriter, *http.Request, http.Handler)

func (m MiddlewareHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request, h http.Handler) {
	m.ServeHTTP(w, r, h)
}

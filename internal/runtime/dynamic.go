package runtime

import (
	"net/http"
	"sync/atomic"
)

type DynamicHandler struct {
	h atomic.Value
}

func NewDynamicHandler(initial http.Handler) *DynamicHandler {
	d := &DynamicHandler{}
	d.h.Store(initial)
	return d
}

func (d *DynamicHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d.h.Load().(http.Handler).ServeHTTP(w, r)
}

func (d *DynamicHandler) Update(h http.Handler) {
	d.h.Store(h)
}

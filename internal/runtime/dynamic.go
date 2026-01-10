package runtime

import (
	"net/http"
	"sync/atomic"
)

type dynamicHandler struct {
	h atomic.Value
}

func (d *dynamicHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d.h.Load().(http.Handler).ServeHTTP(w, r)
}

func (d *dynamicHandler) reload(h http.Handler) {
	d.h.Store(h)
}

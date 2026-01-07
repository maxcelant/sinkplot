package ctrl

import (
	"io"
	"net/http"

	"github.com/maxcelant/sinkplot/internal/runtime"
	"github.com/maxcelant/sinkplot/internal/schema"
	"gopkg.in/yaml.v3"
)

func NewServer(dh *runtime.DynamicHandler) http.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var cfg schema.Config
		if err := yaml.Unmarshal(body, &cfg); err != nil {
			http.Error(w, "invalid yaml", http.StatusBadRequest)
			return
		}

		h, err := runtime.Load(cfg)
		if err != nil {
			http.Error(w, "failed to create new handler chain", http.StatusBadRequest)
			return
		}
		dh.Update(h)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("successfully updated config"))
	})
	return http.Server{
		Addr:    ":8443",
		Handler: handler,
	}
}

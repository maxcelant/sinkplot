package runtime

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/maxcelant/sinkplot/internal/routes"
	"github.com/maxcelant/sinkplot/internal/schema"
	"gopkg.in/yaml.v3"
	"k8s.io/utils/ptr"
)

type ManagerOptions struct {
	masterPort *int
}

type Manager interface {
	Start(*schema.Config) error
	Stop() error
}

type serverManager struct {
	ctx     context.Context
	handler *dynamicHandler
	worker  *http.Server
	master  *http.Server
}

func NewManager(ctx context.Context, opts ManagerOptions) Manager {
	if opts.masterPort == nil || *opts.masterPort == 0 {
		opts.masterPort = ptr.To(8443)
	}
	dh := &dynamicHandler{}
	// NOTE: Since we will want to support 1 or more workers in the future, we will probably need to create a
	// workergroup object here and make sure they all have the same dynamic handler
	worker := &http.Server{Handler: dh, Addr: ":8080"}
	master := func() *http.Server {
		mux := http.NewServeMux()
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		})
		mux.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
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

			h, err := routes.Compile(cfg.App)
			if err != nil {
				http.Error(w, "failed to create new handler chain", http.StatusBadRequest)
				return
			}
			dh.reload(h)
			w.WriteHeader(http.StatusOK)
			log.Println("updated configuration")
			w.Write([]byte("successfully updated config\n"))
		})
		return &http.Server{
			Addr:    fmt.Sprintf(":%d", *opts.masterPort),
			Handler: mux,
		}
	}()

	return &serverManager{
		ctx:     ctx,
		handler: dh,
		worker:  worker,
		master:  master,
	}
}

func (m *serverManager) Start(initCfg *schema.Config) error {
	h, err := routes.Compile(initCfg.App)
	if err != nil {
		return fmt.Errorf("failed to load the initial config: %w", err)
	}
	m.handler.reload(h)
	go func() {
		log.Println("starting worker server on :8080")
		if err := m.worker.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	go func() {
		log.Printf("starting master server on %s\n", m.master.Addr)
		if err := m.master.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// context will block until a signal is triggered / context is cancelled
	<-m.ctx.Done()
	return nil
}

func (m *serverManager) Stop() error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	log.Println("shutting down...")
	if err := m.master.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shut down master server: %w", err)
	}
	if err := m.worker.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shut down worker server: %w", err)
	}
	return nil

}

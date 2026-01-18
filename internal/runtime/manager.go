package runtime

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/maxcelant/sinkplot/internal/routes"
	"github.com/maxcelant/sinkplot/internal/schema"
	"github.com/rs/zerolog/log"
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
	workers runnableGroup
	master  *http.Server
}

// NewManager creates a new cancellable server manager that manages both the worker group and the config server
func NewManager(ctx context.Context, opts ManagerOptions) Manager {
	if opts.masterPort == nil || *opts.masterPort == 0 {
		opts.masterPort = ptr.To(8443)
	}
	dh := &dynamicHandler{}
	manager := &serverManager{
		ctx:     ctx,
		handler: dh,
	}
	manager.workers = NewWorkerGroup(dh)
	manager.master = func() *http.Server {
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
			log.Info().Msg("updated configuration")
			w.Write([]byte("successfully updated config\n"))
		})
		return &http.Server{
			Addr:    fmt.Sprintf(":%d", *opts.masterPort),
			Handler: mux,
		}
	}()

	return manager
}

// Start takes the initial configuration so that it can create the handler chain and start the worker group
func (m *serverManager) Start(initCfg *schema.Config) error {
	h, err := routes.Compile(initCfg.App)
	if err != nil {
		return fmt.Errorf("failed to load the initial config: %w", err)
	}
	m.handler.reload(h)
	go func() {
		if err := m.workers.Start(initCfg.App.Listeners); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("worker server failed")
		}
	}()
	go func() {
		log.Info().Msgf("starting master server on %s", m.master.Addr)
		if err := m.master.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("master server failed")
		}
	}()

	// context will block until a signal is triggered / context is cancelled
	<-m.ctx.Done()
	return nil
}

// Stop gracefully shuts down the config server and the worker group
func (m *serverManager) Stop() error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	log.Info().Msg("shutting down gracefully...")
	if err := m.master.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shut down master server: %w", err)
	}
	if err := m.workers.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shut down worker server: %w", err)
	}
	return nil

}

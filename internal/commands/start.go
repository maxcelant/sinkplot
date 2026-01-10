package start

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/maxcelant/sinkplot/internal/admission"
	"github.com/maxcelant/sinkplot/internal/ctrl"
	"github.com/maxcelant/sinkplot/internal/routes"
	"github.com/maxcelant/sinkplot/internal/runtime"
	"github.com/maxcelant/sinkplot/internal/schema"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start sinkplot proxy in the foreground",
		Long: `Start the sinkplot reverse proxy in the foreground.

This command loads the configuration from a Sinkfile (YAML or JSON),
validates it, and starts both the worker server (port 8080) and the
control server (port 8443). The worker server handles incoming HTTP
requests and routes them to configured upstreams. The control server
accepts live configuration updates via POST requests.

Use --path to specify a custom config file location. The proxy runs
until interrupted (Ctrl+C), then gracefully shuts down.`,
		Run: runStart,
	}
}

func runStart(cmd *cobra.Command, args []string) {
	path, err := cmd.Flags().GetString("path")
	if err != nil {
		log.Fatal(fmt.Errorf("failed to find valid Sinkfile path: %w", err))
	}
	cfg, err := schema.Load(path)
	log.Printf("loading initial config from '%s'", path)
	if err := admission.Default(&cfg.App); err != nil {
		log.Fatal(fmt.Errorf("failed to default config object: %w", err))
	}
	if err := admission.Validate(&cfg.App); err != nil {
		log.Fatal(fmt.Errorf("failed to validate config object: %w", err))
	}
	h, err := routes.Compile(cfg.App)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to load the initial config: %w", err))
	}
	dh := runtime.NewDynamicHandler(h)
	workerSrv := runtime.NewWorkerServer(dh)
	ctrlSrv := ctrl.NewServer(dh)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		log.Println("starting worker server on 8080")
		if err := workerSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	go func() {
		log.Println("starting control server on 8443")
		if err := ctrlSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	workerSrv.Shutdown(shutdownCtx)
	ctrlSrv.Shutdown(shutdownCtx)

}

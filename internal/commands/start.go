package start

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/maxcelant/sinkplot/internal/ctrl"
	"github.com/maxcelant/sinkplot/internal/routes"
	"github.com/maxcelant/sinkplot/internal/runtime"
	"github.com/maxcelant/sinkplot/internal/schema"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start sinkplot proxy in the foreground",
		Long:  "I'll fill this in later :)",
		Run:   runStart,
	}
}

func runStart(cmd *cobra.Command, args []string) {
	var cfg schema.Config
	path, err := cmd.Flags().GetString("path")
	if err != nil {
		log.Fatal(fmt.Errorf("failed to find valid Sinkfile path: %w", err))
	}
	buf, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to read file: %w", err))
	}
	err = yaml.Unmarshal(buf, &cfg)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to unmarshal config to yaml: %w", err))
	}
	log.Printf("loading initial config @ '%s'", path)
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

package start

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/maxcelant/sinkplot/internal/runtime"
	"github.com/spf13/cobra"
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
	chain, err := runtime.Load()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("starting server on 8080")
	workerSrv := http.Server{
		Addr:    ":8080",
		Handler: chain,
	}
	ctrlSrv := http.Server{
		Addr:    ":8443",
		Handler: chain,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		if err := workerSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	go func() {
		if err := ctrlSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	workerSrv.Shutdown(shutdownCtx)

}

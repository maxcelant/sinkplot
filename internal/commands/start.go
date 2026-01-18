package start

import (
	"context"
	"os"
	"os/signal"

	"github.com/maxcelant/jap/internal/config"
	"github.com/maxcelant/jap/internal/runtime"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start sinkplot proxy in the foreground",
		Long: `Start the sinkplot reverse proxy in the foreground.

This command loads the configuration from a config file (YAML or JSON),
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
		log.Fatal().Err(err).Msg("failed to find valid Sinkfile path")
	}

	log.Info().Str("path", path).Msg("starting jap")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	m := runtime.NewManager(ctx, runtime.ManagerOptions{})

	cfg, err := config.Load(path)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}
	log.Info().Int("routes", len(cfg.App.Routes)).Int("sinks", len(cfg.App.Sinks)).Msg("config loaded")

	if err := m.Start(cfg); err != nil {
		log.Fatal().Err(err).Msg("failed to start server manager")
	}

	if err := m.Stop(); err != nil {
		log.Fatal().Err(err).Msg("failed to stop server manager gracefully")
	}
	log.Info().Msg("shutdown complete")
}

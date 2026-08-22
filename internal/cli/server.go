package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/codingguna/aio-panel/internal/api"
	"github.com/codingguna/aio-panel/internal/config"
	"github.com/codingguna/aio-panel/internal/db"
	"github.com/codingguna/aio-panel/internal/logger"
	"github.com/spf13/cobra"
)

var (
	portFlag int
	hostFlag string
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the AIO-PANEL web server and daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		if portFlag > 0 {
			cfg.Server.Port = portFlag
		}
		if hostFlag != "" {
			cfg.Server.Host = hostFlag
		}
		if verbose {
			cfg.Logging.Level = "debug"
		}

		// Ensure directories
		if err := cfg.EnsureDirectories(); err != nil {
			return fmt.Errorf("failed to initialize directories: %w", err)
		}

		// Setup Logger
		logFile, err := logger.Setup(logger.Config{
			Level:    cfg.Logging.Level,
			Format:   cfg.Logging.Format,
			FilePath: cfg.Logging.File,
		})
		if err != nil {
			return fmt.Errorf("failed to initialize logger: %w", err)
		}
		if logFile != nil {
			defer logFile.Close()
		}

		// Initialize Embedded SQLite DB
		store, err := db.Open(cfg.Database.Path)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer store.Close()

		// Start HTTP Server
		srv := api.NewServer(cfg, store)

		// Setup graceful shutdown
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

		go func() {
			if err := srv.Start(); err != nil {
				logger.Log.Error("server stopped with error", "error", err)
			}
		}()

		fmt.Printf("\n🚀 AIO-PANEL is running at http://%s:%d\n\n", cfg.Server.Host, cfg.Server.Port)

		<-stop
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		return srv.Shutdown(ctx)
	},
}

func init() {
	serverCmd.Flags().IntVarP(&portFlag, "port", "p", 0, "Port to listen on (default: 5555)")
	serverCmd.Flags().StringVarP(&hostFlag, "host", "H", "", "Host IP to bind to (default: 0.0.0.0)")
}

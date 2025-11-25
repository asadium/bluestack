package cli

import (
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"github.com/asad/bluestack/internal/config"
	"github.com/asad/bluestack/internal/core"
	"github.com/asad/bluestack/internal/httpx"
	"github.com/asad/bluestack/internal/logging"
	"github.com/asad/bluestack/internal/services/blob"
)

var (
	// Version is set at build time via ldflags.
	// Example: go build -ldflags "-X github.com/asad/bluestack/internal/cli.Version=1.0.0"
	Version = "dev"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "bluestack",
	Short: "Azure LocalStack-style emulator",
	Long: `Bluestack is a local Azure service emulator that provides
minimal but realistic Azure API implementations for local development and testing.

It exposes a single edge HTTP port that routes requests to various service modules
(Blob Storage, Queues, Key Vault, etc.).`,
}

// startCmd represents the start command.
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Bluestack server",
	Long: `Start the Bluestack edge server on the configured port.
The server will listen for HTTP requests and route them to enabled services.`,
	RunE: runStart,
}

// versionCmd represents the version command.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `Print the version number of Bluestack.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("bluestack version %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(versionCmd)
}

// Execute is the entry point for the CLI. It should be called from main.go.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// runStart initializes and starts the HTTP server.
func runStart(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Initialize logger
	logger, err := logging.NewLogger(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	logger.Info("starting bluestack",
		logging.String("version", Version),
		logging.Int("edge_port", cfg.EdgePort),
		logging.String("data_dir", cfg.DataDir),
		logging.String("log_level", cfg.LogLevel),
	)

	// Initialize blob store
	blobStore, err := blob.NewFileBlobStore(cfg.DataDir)
	if err != nil {
		return fmt.Errorf("failed to initialize blob store: %w", err)
	}

	// Create and register services
	blobService := blob.NewBlobService(blobStore, logger)
	core.RegisterService(blobService)

	logger.Info("registered services",
		logging.Int("count", len(core.GetRegisteredServices())),
	)

	// Create edge router
	router := httpx.NewEdgeRouter(cfg, logger)

	// Start HTTP server
	addr := fmt.Sprintf(":%d", cfg.EdgePort)
	logger.Info("listening on edge port",
		logging.String("address", addr),
	)

	if err := http.ListenAndServe(addr, router); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}


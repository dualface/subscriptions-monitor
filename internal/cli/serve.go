package cli

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/user/subscriptions-monitor/internal/api"
)

func init() {
	serveCmd.Flags().IntP("port", "p", 3456, "Port to listen on")
	serveCmd.Flags().StringP("host", "H", "localhost", "Host to listen on")
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the API server",
	RunE: func(cmd *cobra.Command, args []string) error {
		port, _ := cmd.Flags().GetInt("port")
		host, _ := cmd.Flags().GetString("host")

		cfg, registry, err := setup(cmd)
		if err != nil {
			return err
		}

		if port == 3456 && cfg.Settings.APIPort != 0 {
			port = cfg.Settings.APIPort
		}

		addr := fmt.Sprintf("%s:%d", host, port)
		server := api.NewServer(registry, cfg, addr)

		fmt.Printf("Starting API server on http://%s\n", addr)
		fmt.Println("Endpoints:")
		fmt.Println("  GET /api/v1/health    - Health check")
		fmt.Println("  GET /api/v1/usage     - Get usage data (query: provider, name)")
		fmt.Println("  GET /api/v1/providers - List available providers")

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			if err := server.Start(); err != nil && err != http.ErrServerClosed {
				fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			}
		}()

		<-quit
		fmt.Println("\nShutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), cfg.Settings.Timeout)
		defer cancel()

		return server.Shutdown(ctx)
	},
}

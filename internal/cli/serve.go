package cli

import (
	"fmt"

	"github.com/spf13/cobra"
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

		fmt.Printf("Starting API server on %s:%d...\n", host, port)

		return nil
	},
}

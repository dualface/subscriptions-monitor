package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/user/subscriptions-monitor/internal/adapter"
	"github.com/user/subscriptions-monitor/internal/config"
	"github.com/user/subscriptions-monitor/internal/provider"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "sub-mon",
		Short: "AI Subscriptions Monitor",
		Long:  `A tool to monitor usage and costs for various AI service subscriptions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return queryCmd.RunE(cmd, args)
		},
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/sub-mon/config.yaml)")

	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(queryCmd)

	rootCmd.Flags().BoolP("json", "j", false, "Output as JSON")
	rootCmd.Flags().StringP("provider", "p", "", "Filter by provider ID")
	rootCmd.Flags().StringP("name", "n", "", "Filter by subscription name")
}

func setup(cmd *cobra.Command) (*config.Config, *provider.Registry, error) {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	registry := provider.NewRegistry()
	adapter.RegisterAll(registry)

	return cfg, registry, nil
}

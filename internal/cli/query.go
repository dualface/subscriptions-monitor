package cli

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/user/subscriptions-monitor/internal/provider"
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query subscription usage",
	Long:  `Fetches and displays usage data for your configured AI service subscriptions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, registry, err := setup(cmd)
		if err != nil {
			return err
		}

		providerFilter, _ := cmd.Flags().GetString("provider")
		nameFilter, _ := cmd.Flags().GetString("name")

		var filteredSubs []provider.SubscriptionEntry
		for _, sub := range cfg.Subscriptions {
			if providerFilter != "" && sub.Provider != providerFilter {
				continue
			}
			if nameFilter != "" && sub.Name != nameFilter {
				continue
			}
			filteredSubs = append(filteredSubs, sub)
		}

		ctx, cancel := context.WithTimeout(context.Background(), cfg.Settings.Timeout)
		defer cancel()

		snapshots := registry.FetchAll(ctx, filteredSubs)

		jsonOutput, _ := cmd.Flags().GetBool("json")
		if jsonOutput {
			return PrintJSON(snapshots)
		}

		PrintTable(snapshots)
		return nil
	},
}

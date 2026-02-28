package cmd

import (
	"fmt"

	"github.com/Capmus-Team/supost-cli-mark/internal/adapters"
	"github.com/Capmus-Team/supost-cli-mark/internal/config"
	"github.com/Capmus-Team/supost-cli-mark/internal/repository"
	"github.com/Capmus-Team/supost-cli-mark/internal/service"

	"github.com/spf13/cobra"
)

var listingsCmd = &cobra.Command{
	Use:   "listings",
	Short: "List active marketplace listings",
	Long:  `Display all active listings from the marketplace. Uses in-memory seed data by default.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		// Composition root: choose the repository adapter.
		// Swap to repository.NewPostgres(cfg.DatabaseURL) when ready.
		repo := repository.NewInMemory()
		svc := service.NewListingService(repo)

		listings, err := svc.ListActive(cmd.Context())
		if err != nil {
			return fmt.Errorf("fetching listings: %w", err)
		}

		return adapters.Render(cfg.Format, listings)
	},
}

func init() {
	rootCmd.AddCommand(listingsCmd)
	listingsCmd.Flags().StringP("category", "c", "", "filter by category")
}

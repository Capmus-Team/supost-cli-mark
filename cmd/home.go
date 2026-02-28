package cmd

import (
	"fmt"
	"time"

	"github.com/Capmus-Team/supost-cli-mark/internal/adapters"
	"github.com/Capmus-Team/supost-cli-mark/internal/config"
	"github.com/Capmus-Team/supost-cli-mark/internal/domain"
	"github.com/Capmus-Team/supost-cli-mark/internal/repository"
	"github.com/Capmus-Team/supost-cli-mark/internal/service"
	"github.com/spf13/cobra"
)

const featuredJobPostLimit = 3

var homeCmd = &cobra.Command{
	Use:   "home",
	Short: "Render the SUPost home feed",
	Long:  "Show recently posted active posts from the post table.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		limit, err := cmd.Flags().GetInt("limit")
		if err != nil {
			return fmt.Errorf("reading limit flag: %w", err)
		}
		cacheTTL, err := cmd.Flags().GetDuration("cache-ttl")
		if err != nil {
			return fmt.Errorf("reading cache-ttl flag: %w", err)
		}
		refresh, err := cmd.Flags().GetBool("refresh")
		if err != nil {
			return fmt.Errorf("reading refresh flag: %w", err)
		}

		var (
			posts        []domain.Post
			featuredJobs []domain.Post
			usedCache    bool
			cacheLoadErr error
			sections     []domain.HomeCategorySection
			sectionsOK   bool
			sectionsErr  error
		)
		if cfg.DatabaseURL != "" {
			var ok bool
			posts, ok, cacheLoadErr = getCachedHomePosts(cfg.DatabaseURL, refresh, cacheTTL, limit)
			usedCache = cacheLoadErr == nil && ok
			sections, sectionsOK, sectionsErr = getCachedHomeCategorySections(cfg.DatabaseURL, refresh, cacheTTL)
		}

		var (
			repo      service.HomeRepository
			closeRepo func() error
		)
		if cfg.DatabaseURL != "" {
			pgRepo, err := repository.NewPostgres(cfg.DatabaseURL)
			if err != nil {
				if usedCache {
					return renderHomeOutput(cmd, cfg.Format, posts, nil, sections)
				}
				return fmt.Errorf("connecting to postgres: %w", err)
			}
			repo = pgRepo
			closeRepo = pgRepo.Close
		} else {
			repo = repository.NewInMemory()
		}
		if closeRepo != nil {
			defer func() {
				_ = closeRepo()
			}()
		}

		svc := service.NewHomeService(repo)

		if !usedCache {
			posts, err = svc.ListRecentActive(cmd.Context(), limit)
			if err != nil {
				return fmt.Errorf("fetching recent active posts: %w", err)
			}

			if cfg.DatabaseURL != "" && cacheTTL > 0 {
				_ = adapters.SaveHomePostsCache(posts)
			}
		} else if cacheLoadErr != nil && cfg.Verbose {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: loading home cache: %v\n", cacheLoadErr)
		}

		if !sectionsOK {
			sections, err = svc.ListCategorySections(cmd.Context())
			if err != nil {
				if cfg.Verbose {
					fmt.Fprintf(cmd.ErrOrStderr(), "warning: loading category sections: %v\n", err)
				}
				sections = nil
			} else if cfg.DatabaseURL != "" && cacheTTL > 0 && len(sections) > 0 {
				_ = adapters.SaveHomeCategorySectionsCache(sections)
			}
		} else if sectionsErr != nil && cfg.Verbose {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: loading category section cache: %v\n", sectionsErr)
		}

		featuredJobs, err = svc.ListRecentActiveByCategory(cmd.Context(), domain.CategoryJobsOffCampus, featuredJobPostLimit)
		if err != nil {
			if cfg.Verbose {
				fmt.Fprintf(cmd.ErrOrStderr(), "warning: loading featured job posts: %v\n", err)
			}
			featuredJobs = nil
		}

		return renderHomeOutput(cmd, cfg.Format, posts, featuredJobs, sections)
	},
}

func init() {
	rootCmd.AddCommand(homeCmd)
	homeCmd.Flags().Int("limit", 50, "number of recent active posts to show")
	homeCmd.Flags().Duration("cache-ttl", 30*time.Second, "cache TTL for home feed when using database")
	homeCmd.Flags().Bool("refresh", false, "bypass cache and fetch fresh data from database")
}

func getCachedHomePosts(databaseURL string, refresh bool, ttl time.Duration, limit int) ([]domain.Post, bool, error) {
	if databaseURL == "" || refresh || ttl <= 0 {
		return nil, false, nil
	}
	posts, ok, err := adapters.LoadHomePostsCache(ttl, limit)
	if err != nil {
		return nil, false, err
	}
	return posts, ok, nil
}

func getCachedHomeCategorySections(databaseURL string, refresh bool, ttl time.Duration) ([]domain.HomeCategorySection, bool, error) {
	if databaseURL == "" || refresh || ttl <= 0 {
		return nil, false, nil
	}
	sections, ok, err := adapters.LoadHomeCategorySectionsCache(ttl)
	if err != nil {
		return nil, false, err
	}
	return sections, ok, nil
}

func renderHomeOutput(cmd *cobra.Command, format string, posts []domain.Post, featuredJobs []domain.Post, sections []domain.HomeCategorySection) error {
	if !cmd.Flags().Changed("format") && (format == "" || format == "json") {
		return adapters.RenderHomePosts(cmd.OutOrStdout(), posts, featuredJobs, sections)
	}
	if format == "text" || format == "table" {
		return adapters.RenderHomePosts(cmd.OutOrStdout(), posts, featuredJobs, sections)
	}
	return adapters.Render(format, posts)
}

package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/Trungsherlock/jobgocli/internal/scraper"
	"github.com/Trungsherlock/jobgocli/internal/worker"
	"github.com/Trungsherlock/jobgocli/internal/database"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:	"search",
	Short:	"Triggers a full scrape of all enabled companies, stores new jobs, deduplicates against existing jobs",
	RunE: func(cmd *cobra.Command, args []string) error {
		timeout, _ := cmd.Flags().GetDuration("timeout")
		platformFilter, _ := cmd.Flags().GetString("platform")
		companyFilter, _ := cmd.Flags().GetString("company")

		// Get companies to scrape
		companies, err := db.ListCompanies()
		if err != nil {
			return fmt.Errorf("listing companies: %w", err)
		}

		if len(companies) == 0 {
			fmt.Println("No companies to scrape. Add some with: jobgo company add")
			return nil
		}

		// Apply filters
		var filtered []database.Company
		for _, c := range companies {
			if !c.Enabled {
				continue
			}
			if platformFilter != "" && c.Platform != platformFilter {
				continue
			}
			if companyFilter != "" && c.Name != companyFilter {
				continue
			}
			filtered = append(filtered, c)
		}

		if len(filtered) == 0 {
			fmt.Println("No companies match the specified filters.")
			return nil
		}

		fmt.Printf("Scraping %d companies... \n", len(filtered))

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// Run worker pool
		registry := scraper.NewRegistry()
		pool := worker.NewPool(registry, db, 5)
		results := pool.Run(ctx, filtered)

		var totalNew, failures int
		for _, r := range results {
			if r.Err != nil {
				failures++
				fmt.Printf("  FAIL  %s: %v\n", r.Company.Name, r.Err)
			} else {
				fmt.Printf("  OK    %s: %d new jobs\n", r.Company.Name, r.JobCount)
				totalNew += r.JobCount
			}
		}

		fmt.Printf("\nDone. Found %d new jobs from %d companies. %d failed.\n",
			totalNew, len(filtered)-failures, failures)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	
	searchCmd.Flags().String("company", "", "Company name")
	searchCmd.Flags().String("platform", "", "ATS platform (lever, greenhouse)")
	searchCmd.Flags().Duration("timeout", 30*time.Second, "Per-company scrape timeout")
}
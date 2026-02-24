package cli

import (
    "fmt"

    "github.com/Trungsherlock/jobgo/internal/h1b"
    "github.com/spf13/cobra"
)

var h1bCmd = &cobra.Command{
    Use:   "h1b",
    Short: "H1B visa sponsorship data commands",
}

var h1bImportCmd = &cobra.Command{
    Use:   "import <csv-file>",
    Short: "Import USCIS H1B employer data from CSV",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        filePath := args[0]
        fmt.Printf("Importing H1B data from %s...\n", filePath)

        count, err := h1b.ImportSponsors(db, filePath)
        if err != nil {
            return fmt.Errorf("importing sponsors: %w", err)
        }

        fmt.Printf("Imported %d H1B sponsor records.\n", count)

        // Auto-link tracked companies
        link, _ := cmd.Flags().GetBool("link")
        if link {
            linked, err := h1b.LinkCompanies(db)
            if err != nil {
                return fmt.Errorf("linking companies: %w", err)
            }
            fmt.Printf("Linked %d tracked companies to H1B sponsor data.\n", linked)
        }

        return nil
    },
}

var h1bStatusCmd = &cobra.Command{
    Use:   "status",
    Short: "Show H1B sponsorship status for tracked companies",
    RunE: func(cmd *cobra.Command, args []string) error {
        companies, err := db.ListCompanies()
        if err != nil {
            return err
        }
        for _, c := range companies {
            if c.SponsorsH1b {
                rate := 0.0
                if c.H1bApprovalRate != nil {
                    rate = *c.H1bApprovalRate
                }
                filed := 0
                if c.H1bTotalFiled != nil {
                    filed = *c.H1bTotalFiled
                }
                fmt.Printf("  ✓ %s — %.0f%% approval, %d petitions\n", c.Name, rate, filed)
            } else {
                fmt.Printf("  ? %s — no H1B data\n", c.Name)
            }
        }
        return nil
    },
}

func init() {
    rootCmd.AddCommand(h1bCmd)
    h1bCmd.AddCommand(h1bImportCmd)
    h1bCmd.AddCommand(h1bStatusCmd)
    h1bImportCmd.Flags().Bool("link", true, "Auto-link tracked companies to sponsor data")
}

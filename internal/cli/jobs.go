package cli

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"text/tabwriter"
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"
	"github.com/Trungsherlock/jobgo/internal/filter"
)

var jobsCmd = &cobra.Command{
	Use:	"jobs",
	Short:	"Manage jobs",
}

var jobsListCmd = &cobra.Command{
	Use:	"list",
	Short:	"List jobs sorted by match score",
	RunE: func(cmd *cobra.Command, args []string) error {
		minScore, _ := cmd.Flags().GetFloat64("min-score")
		company, _ := cmd.Flags().GetString("company")
		onlyNew, _ := cmd.Flags().GetBool("new")
		titleFlag, _ := cmd.Flags().GetString("title")
		locationFlag, _ := cmd.Flags().GetString("location")
		h1bOnly, _ := cmd.Flags().GetBool("h1b")
		newGradOnly, _ := cmd.Flags().GetBool("new-grad")


		jobs, err := db.ListJobs(minScore, company, onlyNew, false, false, false, false)
		if err != nil {
			return fmt.Errorf("listing jobs: %w", err)
		}	

		params := filter.Params{}
		if titleFlag != "" {
			params.Titles = strings.Split(titleFlag, ",")
		}
		if locationFlag != "" {
			params.Locations = strings.Split(locationFlag, ",")
		}
		params.NewGrad = newGradOnly
		params.H1BOnly = h1bOnly

		var sponsorIDs map[string]bool
		if h1bOnly {
			companies, _ := db.ListCompanies()
			sponsorIDs = make(map[string]bool)
			for _, c := range companies {
				if c.SponsorsH1b {
					sponsorIDs[c.ID] = true
				}
			}
		}

		jobs = filter.Apply(jobs, filter.Build(params, sponsorIDs))

		if len(jobs) == 0 {
			fmt.Println("No jobs found matching the criteria.")
			return nil
		}

		output, _ := cmd.Flags().GetString("output")
		if output == "json" {
			data, _ := json.MarshalIndent(jobs, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "ID\tSCORE\tTITLE\tCOMPANY\tLOCATION\tSTATUS")
		for _, j := range jobs {
			id := j.ID
			score := "-"
			if j.SkillScore != nil {
				score = fmt.Sprintf("%.0f", *j.SkillScore)
			}
			location := ""
			if j.Location != nil {
				location = *j.Location
			}
			title := j.Title
			if len(title) > 45 {
				title = title[:42] + "..."
			}
			companyName := j.CompanyID[:8]
			c, err := db.GetCompany(j.CompanyID)
			if err == nil {
				companyName = c.Name
			}

			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", id, score, title, companyName, location, j.Status)
		}
		_ = w.Flush()
		fmt.Printf("\n%d jobs total\n", len(jobs))
		return nil
	},
}

var jobsShowCmd = &cobra.Command{
	Use:	"show",
	Short:	"Show full job description + match score + match reason",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		job, err := db.GetJob(args[0])
		if err != nil {
			return fmt.Errorf("getting job: %w", err)
		}

		// Get company name
		companyName := job.CompanyID
		c, err := db.GetCompany(job.CompanyID)
		if err == nil {
			companyName = c.Name
		}

		fmt.Printf("Title:       %s\n", job.Title)
		fmt.Printf("Company:     %s\n", companyName)
		if job.Location != nil {
			fmt.Printf("Location:    %s\n", *job.Location)
		}
		fmt.Printf("Remote:      %v\n", job.Remote)
		if job.Department != nil {
			fmt.Printf("Department:  %s\n", *job.Department)
		}
		fmt.Printf("URL:         %s\n", job.URL)
		fmt.Printf("Status:      %s\n", job.Status)

		if job.SkillScore != nil {
			fmt.Printf("Skill Score: %.0f\n", *job.SkillScore)
		}
		if job.SkillReason != nil {
			fmt.Printf("Skill Reason:   %s\n", *job.SkillReason)
		}
		if job.SkillMatched != nil {
			fmt.Printf("Matched: 	%s\n", *job.SkillMatched)
		}
		if job.SkillMissing != nil {
			fmt.Printf("Missing:	%s\n", *job.SkillMissing)
		}

		if job.Description != nil {
			fmt.Printf("\n--- Description ---\n%s\n", *job.Description)
		}

		if job.ExperienceLevel != nil {
			fmt.Printf("Experience Level: %s\n", *job.ExperienceLevel)
		}
		fmt.Printf("New Grad:         %v\n", job.IsNewGrad)
		fmt.Printf("Visa Mentioned:   %v\n", job.VisaMentioned)
		if job.VisaSentiment != nil {
			fmt.Printf("Visa Sentiment:   %s\n", *job.VisaSentiment)
		}

		return nil
	},
}

var jobsOpenCmd = &cobra.Command{
	Use:	"open",
	Short:	"Opens the job URL in the default browser",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		job, err := db.GetJob(args[0])
		if err != nil {
			return fmt.Errorf("getting job: %w", err)
		}
		fmt.Printf("Opening %s ...\n", job.URL)
		return openBrowser(job.URL)
	},
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}

var jobsUpdateCmd = &cobra.Command{
	Use:   "update [id]",
	Short: "Update job application status",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		status, _ := cmd.Flags().GetString("status")
		notes, _ := cmd.Flags().GetString("notes")

		if status == "" {
			return fmt.Errorf("--status is required")
		}

		job, err := db.GetJob(args[0])
		if err != nil {
			return fmt.Errorf("finding job: %w", err)
		}

		if err := db.UpdateApplication(job.ID, status, notes); err != nil {
			return fmt.Errorf("updating: %w", err)
		}

		fmt.Printf("Updated %s to status: %s\n", job.Title, status)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(jobsCmd)
	jobsCmd.AddCommand(jobsListCmd)
	jobsCmd.AddCommand(jobsShowCmd)
	jobsCmd.AddCommand(jobsOpenCmd)
	jobsCmd.AddCommand(jobsUpdateCmd)

	jobsListCmd.Flags().Float64("min-score", 0, "Minimum skill score (0-100)")
	jobsListCmd.Flags().String("company", "", "Filter by company ID")
	jobsListCmd.Flags().Bool("new", false, "Only unseen jobs")
	jobsListCmd.Flags().String("title", "", "Filter by title (e.g. 'software engineer,backend engineer')")
	jobsListCmd.Flags().String("location", "", "Filter by location (e.g. 'US,remote')")
	jobsListCmd.Flags().Bool("new-grad", false, "Only new-grad friendly jobs")
	jobsListCmd.Flags().Bool("h1b", false, "Only H1B-sponsoring companies")
	jobsListCmd.Flags().String("output", "", "Output format: json")
}
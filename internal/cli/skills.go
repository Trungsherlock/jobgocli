package cli

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/Trungsherlock/jobgo/internal/skills"
	"github.com/spf13/cobra"
)

var skillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "Skill taxonomy and gap analysis",
}

var skillsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all skills in the taxonomy",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Taxonomy contains %d canonical skills:\n\n", len(skills.Skills))
		for _, s := range skills.Skills {
			fmt.Printf("  %s\n", s)
		}
		fmt.Printf("\n%d aliases defined.\n", len(skills.Aliases))
		return nil
	},
}

var skillsGapCmd = &cobra.Command{
	Use:   "gap",
	Short: "Show top missing skills across your highest-matched jobs",
	RunE: func(cmd *cobra.Command, args []string) error {
		minScore, _ := cmd.Flags().GetFloat64("min-score")
		topN, _ := cmd.Flags().GetInt("top")

		jobs, err := db.ListJobs(minScore, "", false, false, false, false, false)
		if err != nil {
			return fmt.Errorf("listing jobs: %w", err)
		}

		freq := map[string]int{}
		total := 0
		for _, j := range jobs {
			if j.SkillMissing == nil {
				continue
			}
			var missing []string
			if err := json.Unmarshal([]byte(*j.SkillMissing), &missing); err != nil {
				continue
			}
			for _, s := range missing {
				freq[s]++
			}
			total++
		}

		if total == 0 {
			fmt.Println("No scored jobs found. Run: jobgo scan")
			return nil
		}

		type gap struct {
			Skill string
			Count int
		}
		gaps := make([]gap, 0, len(freq))
		for skill, count := range freq {
			gaps = append(gaps, gap{skill, count})
		}
		sort.Slice(gaps, func(i, j int) bool { return gaps[i].Count > gaps[j].Count })
		if len(gaps) > topN {
			gaps = gaps[:topN]
		}

		fmt.Printf("Top %d missing skills across %d scored jobs (score >= %.0f):\n\n", len(gaps), total, minScore)
		for i, g := range gaps {
			pct := float64(g.Count) / float64(total) * 100
			fmt.Printf("  %2d. %-25s  missing in %d jobs (%.0f%%)\n", i+1, g.Skill, g.Count, pct)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(skillsCmd)
	skillsCmd.AddCommand(skillsListCmd)
	skillsCmd.AddCommand(skillsGapCmd)

	skillsGapCmd.Flags().Float64("min-score", 50, "Only analyze jobs above this score")
	skillsGapCmd.Flags().Int("top", 10, "Number of top skills to show")
}

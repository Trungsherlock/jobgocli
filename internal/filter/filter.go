package filter

import (
	"strings"

	"github.com/Trungsherlock/jobgocli/internal/database"
)

type Filter interface {
	Name() string
	Apply(job database.Job) bool
}

func Apply(jobs []database.Job, filters []Filter) []database.Job {
	if len(filters) == 0 {
		return jobs
	}
	result := make([]database.Job, 0, len(jobs))
	for _, job := range jobs {
		pass := true
		for _, f := range filters {
			if !f.Apply(job) {
				pass = false
				break
			}
		}
		if pass {
			result = append(result, job)
		}
	}
	return result
}

var titleAliases = map[string]string{
	"swe":		"software engineer",
	"sde":		"software engineer",
	"mle":		"machine learning engineer",
	"ml":		"machine learning",
	"sre":		"site reliability engineer",
	"devops":	"developer operations",
	"fe": 		"frontend",
	"be": 		"backend",
}

type TitleFilter struct {
	Titles []string
}

func (f *TitleFilter) Name() string {return "title"}

func (f *TitleFilter) Apply(job database.Job) bool {
	if len(f.Titles) == 0 {
		return true
	}
	titleLower := strings.ToLower(job.Title)
	for _, t := range f.Titles {
		tLower := strings.ToLower(strings.TrimSpace(t))
		if canonical, ok := titleAliases[tLower]; ok {
			tLower = canonical
		}
		if strings.Contains(titleLower, tLower) {
			return true
		}
	}
	return false
}

type LocationFilter struct {
	Locations []string
}

func (f *LocationFilter) Name() string {return "location"}

func (f *LocationFilter) Apply(job database.Job) bool {
	if len(f.Locations) == 0 {
		return true
	}
	jobLocation := ""
	if job.Location != nil {
		jobLocation = strings.ToLower(*job.Location)
	}
	for _, loc := range f.Locations {
		locLower := strings.ToLower(strings.TrimSpace(loc))
		if locLower == "remote" {
			if job.Remote || strings.Contains(jobLocation, "remote") {
				return true
			}
			continue
		}
		if strings.Contains(jobLocation, locLower) {
			return true
		}
	}
	return false
}

type NewGradFilter struct {}

func (f *NewGradFilter) Name() string {return "new_grad"}

func (f *NewGradFilter) Apply(job database.Job) bool {
	if job.IsNewGrad {
		return true
	}
	if job.ExperienceLevel != nil && *job.ExperienceLevel == "new_grad" {
		return false
	}
	return false
}

type H1BFilter struct {
	SponsorIDs map[string]bool
}

func (f *H1BFilter) Name() string {return "h1b"}

func (f *H1BFilter) Apply(job database.Job) bool {
	return f.SponsorIDs[job.CompanyID]
}

type Params struct {
    Titles    []string
    Locations []string
    NewGrad   bool
    H1BOnly   bool
}

func Build(p Params, h1bSponsorIDs map[string]bool) []Filter {
    var filters []Filter
    if len(p.Titles) > 0 {
        filters = append(filters, &TitleFilter{Titles: p.Titles})
    }
    if len(p.Locations) > 0 {
        filters = append(filters, &LocationFilter{Locations: p.Locations})
    }
    if p.NewGrad {
        filters = append(filters, &NewGradFilter{})
    }
    if p.H1BOnly {
        filters = append(filters, &H1BFilter{SponsorIDs: h1bSponsorIDs})
    }
    return filters
}
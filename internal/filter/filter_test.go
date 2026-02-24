package filter

import (
    "testing"

    "github.com/Trungsherlock/jobgo/internal/database"
)

func strPtr(s string) *string { return &s }

func TestTitleFilter(t *testing.T) {
    f := &TitleFilter{Titles: []string{"software engineer", "backend engineer"}}

    pass := []string{"Software Engineer II", "Senior Backend Engineer", "New Grad Software Engineer"}
    for _, title := range pass {
        job := database.Job{Title: title}
        if !f.Apply(job) {
            t.Errorf("TitleFilter should pass %q", title)
        }
    }

    fail := []string{"Product Manager", "Data Scientist", "Marketing Lead"}
    for _, title := range fail {
        job := database.Job{Title: title}
        if f.Apply(job) {
            t.Errorf("TitleFilter should reject %q", title)
        }
    }
}

func TestTitleFilter_Alias(t *testing.T) {
    f := &TitleFilter{Titles: []string{"swe"}}
    job := database.Job{Title: "Software Engineer"}
    if !f.Apply(job) {
        t.Error("TitleFilter should resolve 'swe' alias to 'software engineer'")
    }
}

func TestLocationFilter_Remote(t *testing.T) {
    f := &LocationFilter{Locations: []string{"remote"}}

    remoteJob := database.Job{Remote: true}
    if !f.Apply(remoteJob) {
        t.Error("should pass remote job")
    }

    onsite := database.Job{Remote: false, Location: strPtr("New York")}
    if f.Apply(onsite) {
        t.Error("should reject non-remote job")
    }
}

func TestLocationFilter_Country(t *testing.T) {
    f := &LocationFilter{Locations: []string{"US"}}

    pass := []string{"San Francisco, US", "New York, US", "United States"}
    for _, loc := range pass {
        job := database.Job{Location: strPtr(loc)}
        if !f.Apply(job) {
            t.Errorf("should pass location %q", loc)
        }
    }

    fail := []string{"London, UK", "Berlin, Germany"}
    for _, loc := range fail {
        job := database.Job{Location: strPtr(loc)}
        if f.Apply(job) {
            t.Errorf("should reject location %q", loc)
        }
    }
}

func TestNewGradFilter(t *testing.T) {
    f := &NewGradFilter{}

    if !f.Apply(database.Job{IsNewGrad: true}) {
        t.Error("should pass new grad job")
    }
    lvl := "new_grad"
    if !f.Apply(database.Job{ExperienceLevel: &lvl}) {
        t.Error("should pass job with new_grad experience level")
    }
    if f.Apply(database.Job{IsNewGrad: false}) {
        t.Error("should reject non-new-grad job")
    }
}

func TestH1BFilter(t *testing.T) {
    sponsors := map[string]bool{"company-1": true, "company-2": true}
    f := &H1BFilter{SponsorIDs: sponsors}

    if !f.Apply(database.Job{CompanyID: "company-1"}) {
        t.Error("should pass H1B sponsor")
    }
    if f.Apply(database.Job{CompanyID: "company-99"}) {
        t.Error("should reject non-sponsor")
    }
}

func TestApply_MultipleFilters(t *testing.T) {
    jobs := []database.Job{
        {Title: "Software Engineer", Remote: true, IsNewGrad: true, CompanyID: "c1"},
        {Title: "Product Manager", Remote: true, IsNewGrad: true, CompanyID: "c1"},
        {Title: "Software Engineer", Remote: false, IsNewGrad: true, CompanyID: "c1"},
        {Title: "Software Engineer", Remote: true, IsNewGrad: false, CompanyID: "c1"},
    }

    filters := []Filter{
        &TitleFilter{Titles: []string{"software engineer"}},
        &LocationFilter{Locations: []string{"remote"}},
        &NewGradFilter{},
    }

    result := Apply(jobs, filters)
    if len(result) != 1 {
        t.Errorf("expected 1 job, got %d: %v", len(result), result)
    }
}

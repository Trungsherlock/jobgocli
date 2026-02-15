package scraper

import (
	"context"
	"testing"
	"time"
)

func TestGreenhouseFetchJobs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	scraper := NewGreenhouseScraper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Use a known Greenhouse company
	jobs, err := scraper.FetchJobs(ctx, "airbnb")
	if err != nil {
		t.Fatalf("FetchJobs: %v", err)
	}

	t.Logf("Found %d jobs from Greenhouse/airbnb", len(jobs))

	if len(jobs) == 0 {
		t.Log("Warning: no jobs returned")
		return
	}

	j := jobs[0]
	if j.ExternalID == "" {
		t.Error("ExternalID is empty")
	}
	if j.Title == "" {
		t.Error("Title is empty")
	}
	if j.URL == "" {
		t.Error("URL is empty")
	}

	t.Logf("Sample job: %s — %s (%s)", j.Title, j.Location, j.URL)
}

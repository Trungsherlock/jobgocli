package h1b

import (
	"testing"

	"github.com/Trungsherlock/jobgo/internal/database"
)

func TestClassifyJob(t *testing.T) {
	tests := []struct {
		name          string
		title         string
		description   string
		wantLevel     string
		wantNewGrad   bool
		wantVisa      bool
		wantSentiment string
	}{
		{
			name:        "new grad role",
			title:       "Software Engineer, New Grad",
			description: "Join our team as a new graduate",
			wantLevel:   "entry",
			wantNewGrad: true,
		},
		{
			name:        "senior role by title",
			title:       "Senior Backend Engineer",
			description: "We are looking for an experienced engineer",
			wantLevel:   "senior",
			wantNewGrad: false,
		},
		{
			name:        "senior role by years",
			title:       "Backend Engineer",
			description: "5+ years of experience required",
			wantLevel:   "senior",
			wantNewGrad: false,
		},
		{
			name:          "no sponsorship",
			title:         "Software Engineer",
			description:   "Must be authorized to work in the US. No visa sponsorship available.",
			wantLevel:     "mid",
			wantNewGrad:   false,
			wantVisa:      true,
			wantSentiment: "negative",
		},
		{
			name:          "sponsors visa",
			title:         "Backend Engineer",
			description:   "We will sponsor H1B visa for the right candidate",
			wantLevel:     "mid",
			wantNewGrad:   false,
			wantVisa:      true,
			wantSentiment: "positive",
		},
		{
			name:        "intern role",
			title:       "Software Engineering Intern",
			description: "Summer internship program for students",
			wantLevel:   "intern",
			wantNewGrad: false,
		},
		{
			name:        "staff engineer",
			title:       "Staff Software Engineer",
			description: "Lead technical direction across teams",
			wantLevel:   "staff",
			wantNewGrad: false,
		},
		{
			name:        "entry level by description",
			title:       "Software Engineer",
			description: "Entry level position, 1-2 years of experience",
			wantLevel:   "entry",
			wantNewGrad: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc := tt.description
			job := database.Job{
				Title:       tt.title,
				Description: &desc,
			}

			level, newGrad, visa, sentiment := ClassifyJob(job)

			if level != tt.wantLevel {
				t.Errorf("level = %q, want %q", level, tt.wantLevel)
			}
			if newGrad != tt.wantNewGrad {
				t.Errorf("newGrad = %v, want %v", newGrad, tt.wantNewGrad)
			}
			if visa != tt.wantVisa {
				t.Errorf("visa = %v, want %v", visa, tt.wantVisa)
			}
			if sentiment != tt.wantSentiment {
				t.Errorf("sentiment = %q, want %q", sentiment, tt.wantSentiment)
			}
		})
	}
}

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Google LLC", "google"},
		{"Apple Inc.", "apple"},
		{"Meta, Inc.", "meta"},
		{"Amazon.com, Inc", "amazoncom"},
		{"NVIDIA Corporation", "nvidia corporation"},
		{"  Stripe  ", "stripe"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := NormalizeName(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestScoreH1B(t *testing.T) {
	visaSentimentPos := "positive"
	visaSentimentNeg := "negative"
	approvalRate := 95.0
	totalFiled := 100

	tests := []struct {
		name       string
		job        database.Job
		company    database.Company
		profile    database.Profile
		wantDelta  float64
		wantReason bool // just check reason is non-empty
	}{
		{
			name:      "visa not required - no adjustment",
			job:       database.Job{},
			company:   database.Company{SponsorsH1b: true, H1bApprovalRate: &approvalRate},
			profile:   database.Profile{VisaRequired: false},
			wantDelta: 0,
		},
		{
			name:       "sponsor with high approval",
			job:        database.Job{},
			company:    database.Company{SponsorsH1b: true, H1bApprovalRate: &approvalRate, H1bTotalFiled: &totalFiled},
			profile:    database.Profile{VisaRequired: true},
			wantDelta:  15,
			wantReason: true,
		},
		{
			name:       "negative visa sentiment",
			job:        database.Job{VisaSentiment: &visaSentimentNeg},
			company:    database.Company{},
			profile:    database.Profile{VisaRequired: true},
			wantDelta:  -20,
			wantReason: true,
		},
		{
			name:       "sponsor + positive visa",
			job:        database.Job{VisaSentiment: &visaSentimentPos},
			company:    database.Company{SponsorsH1b: true, H1bApprovalRate: &approvalRate},
			profile:    database.Profile{VisaRequired: true},
			wantDelta:  25, // 15 + 10
			wantReason: true,
		},
		{
			name:       "new grad with junior profile",
			job:        database.Job{IsNewGrad: true},
			company:    database.Company{},
			profile:    database.Profile{VisaRequired: true, ExperienceYears: 1},
			wantDelta:  5,
			wantReason: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adj := ScoreH1B(tt.job, tt.company, tt.profile)
			if adj.Delta != tt.wantDelta {
				t.Errorf("Delta = %.0f, want %.0f", adj.Delta, tt.wantDelta)
			}
			if tt.wantReason && adj.Reason == "" {
				t.Error("expected non-empty reason")
			}
			if !tt.wantReason && adj.Reason != "" {
				t.Errorf("expected empty reason, got %q", adj.Reason)
			}
		})
	}
}

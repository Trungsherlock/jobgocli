package h1b

import (
	"strings"

	"github.com/Trungsherlock/jobgocli/internal/database"
)

func ClassifyJob(job database.Job) (experienceLevel string, isNewGrad bool, visaMentioned bool, visaSentiment string) {
	title := strings.ToLower(job.Title)
	desc := ""
	if job.Description != nil {
		desc = strings.ToLower(*job.Description)
	}

	combined := title + " " + desc
	experienceLevel, isNewGrad = detectExperienceLevel(title, combined)
	visaMentioned, visaSentiment = detectVisaStance(combined)

	return
}

func detectExperienceLevel(title, combined string) (string, bool) {
	newGradPatterns := []string{
		"new grad", "new graduate", "university grad",
		"entry level", "entry-level", "junior",
		"associate", "early career", "early-career",
		"campus", "recent graduate", "fresh graduate",
		"0-2 years", "0-1 years", "1-2 years",
	}

	for _, p := range newGradPatterns {
		if strings.Contains(combined, p) {
			return "entry", true
		}
	}

	if strings.Contains(title, "intern") {
		return "intern", false
	}
	if strings.Contains(title, "senior") || strings.Contains(title, "sr.") || strings.Contains(title, "sr ") {
		return "senior", false
	}
	if strings.Contains(title, "staff") || strings.Contains(title, "principal") {
		return "staff", false
	}
	if strings.Contains(title, "lead") || strings.Contains(title, "manager") || strings.Contains(title, "director") {
		return "lead", false
	}

	// Check description for years of experience
	if strings.Contains(combined, "5+ years") || strings.Contains(combined, "7+ years") ||
		strings.Contains(combined, "10+ years") || strings.Contains(combined, "8+ years") {
		return "senior", false
	}
	if strings.Contains(combined, "3+ years") || strings.Contains(combined, "4+ years") {
		return "mid", false
	}
	if strings.Contains(combined, "1+ years") || strings.Contains(combined, "2+ years") {
		return "entry", false
	}

	return "mid", false
}

func detectVisaStance(combined string) (bool, string) {
	positivePatterns := []string{
		"visa sponsorship available",
		"will sponsor",
		"sponsorship provided",
		"we sponsor",
		"open to sponsorship",
		"h1b sponsorship",
		"visa support",
	}

	negativePatterns := []string{
		"no visa sponsorship",
		"not sponsor",
		"cannot sponsor",
		"will not sponsor",
		"without sponsorship",
		"no sponsorship",
		"must be authorized",
		"must be legally authorized",
		"authorized to work",
		"without visa sponsorship",
		"u.s. citizen",
		"us citizen",
		"permanent resident",
		"green card",
		"security clearance required",
	}

	for _, p := range negativePatterns {
		if strings.Contains(combined, p) {
			return true, "negative"
		}
	}

	for _, p := range positivePatterns {
		if strings.Contains(combined, p) {
			return true, "positive"
		}
	}

	visaKeywords := []string{"visa", "sponsorship", "h1b", "h-1b", "work authorization"}
	for _, k := range visaKeywords {
		if strings.Contains(combined, k) {
			return true, "neutral"
		}
	}

	return false, ""
}

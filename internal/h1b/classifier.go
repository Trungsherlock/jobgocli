package h1b

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/Trungsherlock/jobgo/internal/database"
)

// matches "2-8 YOE", "3–5 years", "0-2 yrs" etc.
var yoeRangeRE = regexp.MustCompile(`(?i)(\d+)\s*[-–]\s*(\d+)\s*(?:years?|yrs?|yoe)\b`)

// matches "8+ years", "3+ yoe", "5 yrs", "10 years" etc.
var yoeRE = regexp.MustCompile(`(?i)(\d+)\+?\s*(?:years?|yrs?|yoe)\b`)

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

	// Check title first for explicit YOE range (e.g. "(2-8 YOE)", "(0-2 years)")
	if m := yoeRangeRE.FindStringSubmatch(title); m != nil {
		upper, _ := strconv.Atoi(m[2])
		return levelByYears(upper)
	}

	// Check combined text for any YOE mention (e.g. "8+ yoe", "3+ years of experience")
	if m := yoeRE.FindStringSubmatch(combined); m != nil {
		n, _ := strconv.Atoi(m[1])
		return levelByYears(n)
	}

	return "mid", false
}

func levelByYears(n int) (string, bool) {
	switch {
	case n >= 5:
		return "senior", false
	case n >= 3:
		return "mid", false
	default:
		return "entry", false
	}
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

package skills

import (
    "slices"
    "testing"
)

func TestExtractFromJob(t *testing.T) {
    desc := `
We are hiring a backend engineer.

Requirements:
You must have experience with Go, PostgreSQL, and Docker.
Kubernetes knowledge is required.

Nice to have:
Terraform and Datadog experience preferred.
gRPC is a bonus.

About us:
We use Git, Linux, and REST APIs daily.
`
    result := ExtractFromJob(desc)

    // Required skills
    for _, skill := range []string{"Go", "PostgreSQL", "Docker", "Kubernetes"} {
        if !slices.Contains(result.Required, skill) {
            t.Errorf("expected %q in required skills, got %v", skill, result.Required)
        }
    }

    // Preferred skills
    for _, skill := range []string{"Terraform", "Datadog", "gRPC"} {
        if !slices.Contains(result.Preferred, skill) {
            t.Errorf("expected %q in preferred skills, got %v", skill, result.Preferred)
        }
    }

    // No duplicates across sections
    all := append(append(result.Required, result.Preferred...), result.Mentioned...)
    seen := map[string]bool{}
    for _, s := range all {
        if seen[s] {
            t.Errorf("duplicate skill %q across sections", s)
        }
        seen[s] = true
    }
}

func TestExtractFromResume(t *testing.T) {
    resume := `
John Doe — Software Engineer
Skills: Go, Python, PostgreSQL, Docker, Kubernetes, AWS, Terraform
Experience with gRPC and REST APIs.
Built CI/CD pipelines using GitHub Actions.
`
    skills := ExtractFromResume(resume)
    for _, want := range []string{"Go", "Python", "PostgreSQL", "Docker", "Kubernetes", "AWS", "Terraform", "gRPC", "REST", "GitHub Actions"} {
        if !slices.Contains(skills, want) {
            t.Errorf("expected %q in resume skills, got %v", want, skills)
        }
    }
}

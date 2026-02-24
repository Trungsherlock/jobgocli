package matcher

import (
    "testing"

    "github.com/Trungsherlock/jobgo/internal/database"
)

func TestSkillScorer(t *testing.T) {
    scorer := NewSkillScorer()

    profile := database.Profile{
        Skills: `["Go","PostgreSQL","Docker","Kubernetes","AWS"]`,
    }

    tests := []struct {
        name     string
        desc     string
        minScore float64
        maxScore float64
    }{
        {
            name: "strong required match",
            desc: "Requirements:\nGo, PostgreSQL, Docker, Kubernetes\nNice to have:\nTerraform",
            minScore: 65,
            maxScore: 100,
        },
        {
            name: "partial required match",
            desc: "Requirements:\nGo, Java, Kafka, PostgreSQL\nNice to have:\nDocker",
            minScore: 30,
            maxScore: 65,
        },
        {
            name: "no skill overlap",
            desc: "Requirements:\nJava, Spring Boot, Oracle, Jenkins\nNice to have:\nMaven",
            minScore: 0,
            maxScore: 35,
        },
        {
            name: "no description",
            desc: "",
            minScore: 0,
            maxScore: 0,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            job := database.Job{Description: strPtr(tt.desc)}
            result := scorer.Score(job, profile)
            if result.Score < tt.minScore || result.Score > tt.maxScore {
                t.Errorf("score %.1f not in [%.0f, %.0f]. matched=%v missing=%v reason=%s",
                    result.Score, tt.minScore, tt.maxScore,
                    result.MatchedSkills, result.MissingSkills, result.Reason)
            }
            t.Logf("%.1f — %s", result.Score, result.Reason)
        })
    }
}

func TestSkillScorer_NoProfile(t *testing.T) {
    scorer := NewSkillScorer()
    job := database.Job{Description: strPtr("Requirements: Go, Docker")}
    result := scorer.Score(job, database.Profile{Skills: ""})
    if result.Score != 0 {
        t.Errorf("expected score 0 for empty profile, got %.1f", result.Score)
    }
}

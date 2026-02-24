package skills

import "testing"

func TestNormalize(t *testing.T) {
    tests := []struct {
        input    string
        expected string
    }{
        {"golang", "Go"},
        {"Golang", "Go"},
        {"k8s", "Kubernetes"},
        {"postgres", "PostgreSQL"},
        {"Postgres", "PostgreSQL"},
        {"nodejs", "Node.js"},
        {"react.js", "React"},
        {"gh actions", "GitHub Actions"},
        {"Go", "Go"},
        {"Docker", "Docker"},
        {"UNKNOWN_SKILL_XYZ", "UNKNOWN_SKILL_XYZ"}, // passthrough
    }
    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            got := Normalize(tt.input)
            if got != tt.expected {
                t.Errorf("Normalize(%q) = %q, want %q", tt.input, got, tt.expected)
            }
        })
    }
}

func TestIsKnown(t *testing.T) {
    known := []string{"Go", "Python", "Docker", "Kubernetes", "PostgreSQL", "React", "AWS"}
    for _, s := range known {
        if !IsKnown(s) {
            t.Errorf("IsKnown(%q) = false, want true", s)
        }
    }
    unknown := []string{"FooBarBaz", "MyCustomTool", ""}
    for _, s := range unknown {
        if IsKnown(s) {
            t.Errorf("IsKnown(%q) = true, want false", s)
        }
    }
}

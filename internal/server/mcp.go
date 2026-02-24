package server

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/Trungsherlock/jobgo/internal/database"
	"github.com/Trungsherlock/jobgo/internal/filter"
	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

type MCPServer struct {
	db     *database.DB
	server *mcpserver.MCPServer
}

func NewMCP(db *database.DB) *MCPServer {
	s := mcpserver.NewMCPServer(
		"JobGo",
		"1.0.0",
		mcpserver.WithToolCapabilities(true),
	)

	m := &MCPServer{db: db, server: s}
	m.registerTools()
	return m
}

func (m *MCPServer) registerTools() {
	// search_jobs tool
	m.server.AddTool(
		mcp.NewTool("search_jobs",
			mcp.WithDescription("Search for jobs matching criteria. Returns a list of job postings with match scores."),
			mcp.WithNumber("min_score", mcp.Description("Minimum match score (0-100)"), mcp.DefaultNumber(0)),
			mcp.WithString("title", mcp.Description("Filter by job title (e.g. 'software engineer')")),
			mcp.WithString("location", mcp.Description("Filter by location (e.g. 'US,remote')")),
			mcp.WithBoolean("new_only", mcp.Description("Only return unseen jobs"), mcp.DefaultBool(false)),
			mcp.WithBoolean("new_grad", mcp.Description("Only return new-grad friendly jobs"), mcp.DefaultBool(false)),
			mcp.WithBoolean("h1b_only", mcp.Description("Only return jobs from H1B sponsors"), mcp.DefaultBool(false)),
		),
		m.searchJobs,
	)

	// get_job_details tool
	m.server.AddTool(
		mcp.NewTool("get_job_details",
			mcp.WithDescription("Get full details of a specific job posting including description and match reason."),
			mcp.WithString("job_id", mcp.Required(), mcp.Description("The job ID")),
		),
		m.getJobDetails,
	)

	// list_companies tool
	m.server.AddTool(
		mcp.NewTool("list_companies",
			mcp.WithDescription("List all tracked companies and their scraping status."),
		),
		m.listCompanies,
	)

	// get_profile tool
	m.server.AddTool(
		mcp.NewTool("get_profile",
			mcp.WithDescription("Get the user's profile including skills, preferred roles, and locations."),
		),
		m.getProfile,
	)

	// get_stats tool
	m.server.AddTool(
		mcp.NewTool("get_stats",
			mcp.WithDescription("Get application pipeline statistics showing how many applications are in each stage."),
		),
		m.getStats,
	)

	// analyze_skill_gap tool
	m.server.AddTool(
		mcp.NewTool("analyze_skill_gap",
            mcp.WithDescription("Analyze which skills appear most often in your top-matched jobs but are missing from your profile. Useful for identifying what to learn next."),
            mcp.WithNumber("min_score", mcp.Description("Only analyze jobs above this score (default 60)"), mcp.DefaultNumber(60)),
        ),
		m.analyzeSkillGap,
	)
}

func (m *MCPServer) searchJobs(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, _ := req.Params.Arguments.(map[string]interface{})
	minScore, _ := args["min_score"].(float64)
	titleParam, _ := args["title"].(string)
	locationParam, _ := args["location"].(string)
	newOnly, _ := args["new_only"].(bool)
	newGrad, _ := args["new_grad"].(bool)
	h1bOnly, _ := args["h1b_only"].(bool)

	jobs, err := m.db.ListJobs(minScore, "", newOnly, false, false, false, false)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	params := filter.Params{NewGrad: newGrad, H1BOnly: h1bOnly}
	if titleParam != "" {
		params.Titles = strings.Split(titleParam, ",")
	}
	if locationParam != "" {
		params.Locations = strings.Split(locationParam, ",")
	}
	var sponsorIDs map[string]bool
	if h1bOnly {
		companies, _ := m.db.ListCompanies()
		sponsorIDs = make(map[string]bool)
		for _, c := range companies {
			if c.SponsorsH1b {
				sponsorIDs[c.ID] = true
			}
		}
	}
	jobs = filter.Apply(jobs, filter.Build(params, sponsorIDs))

	// Build a concise summary for the AI
	type jobSummary struct {
		ID       		string   	`json:"id"`
		Title    		string   	`json:"title"`
		Company  		string   	`json:"company"`
		Location 		string   	`json:"location"`
		Remote   		bool     	`json:"remote"`
		SkillScore    	*float64 	`json:"skill_score"`
		MatchedSkills   *string 	`json:"matched_skills,omitempty"`
		MissingSkills   *string 	`json:"missing_skills,omitempty"`
		Status   		string   	`json:"status"`
		URL      		string   	`json:"url"`
		IsNewGrad		bool		`json:"is_new_grad"`
	}

	summaries := make([]jobSummary, 0, len(jobs))
	for _, j := range jobs {
		location := ""
		if j.Location != nil {
			location = *j.Location
		}
		summaries = append(summaries, jobSummary{
			ID:       		j.ID,
			Title:    		j.Title,
			Location: 		location,
			Remote:   		j.Remote,
			SkillScore:    	j.SkillScore,
			MatchedSkills:  j.SkillMatched,
			MissingSkills:  j.SkillMissing,
			Status:   		j.Status,
			URL:      		j.URL,
			IsNewGrad: 		j.IsNewGrad,
		})
	}

	data, _ := json.MarshalIndent(summaries, "", "  ")
	return mcp.NewToolResultText(fmt.Sprintf("Found %d jobs:\n%s", len(summaries), string(data))), nil
}

func (m *MCPServer) getJobDetails(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, _ := req.Params.Arguments.(map[string]interface{})
	jobID, _ := args["job_id"].(string)

	job, err := m.db.GetJob(jobID)
	if err != nil {
		return mcp.NewToolResultError("Job not found: " + err.Error()), nil
	}

	companyName := job.CompanyID[:8]
	c, err := m.db.GetCompany(job.CompanyID)
	if err == nil {
		companyName = c.Name
	}

	details := map[string]interface{}{
		"id":      job.ID,
		"title":   job.Title,
		"company": companyName,
		"url":     job.URL,
		"remote":  job.Remote,
		"status":  job.Status,
	}

	if job.Location != nil {
		details["location"] = *job.Location
	}
	if job.Department != nil {
		details["department"] = *job.Department
	}
	if job.Description != nil {
		details["description"] = *job.Description
	}
	if job.SkillScore != nil {
		details["skill_score"] = *job.SkillScore
	}
	if job.SkillReason != nil {
		details["skill_reason"] = *job.SkillReason
	}
	if job.SkillMatched != nil {
		details["matched_skills"] = *job.SkillMatched
	}
	if job.SkillMissing != nil {
		details["missing_skills"] = *job.SkillMissing
	}
	details["is_new_grad"] = job.IsNewGrad
	details["visa_mentioned"] = job.VisaMentioned
	if job.VisaSentiment != nil {
		details["visa_sentiment"] = *job.VisaSentiment
	}
	if job.ExperienceLevel != nil {
		details["experience_level"] = *job.ExperienceLevel
	}
	if c, err := m.db.GetCompany(job.CompanyID); err == nil {
		details["sponsors_h1b"] = c.SponsorsH1b
		if c.H1bApprovalRate != nil {
			details["h1b_approval_rate"] = *c.H1bApprovalRate
		}
	}

	data, _ := json.MarshalIndent(details, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (m *MCPServer) listCompanies(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	companies, err := m.db.ListCompanies()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	data, _ := json.MarshalIndent(companies, "", "  ")
	return mcp.NewToolResultText(fmt.Sprintf("%d companies:\n%s", len(companies), string(data))), nil
}

func (m *MCPServer) getProfile(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	profile, err := m.db.GetProfile()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if profile == nil {
		return mcp.NewToolResultText("No profile configured. Use 'jobgo profile set' to create one."), nil
	}
	data, _ := json.MarshalIndent(profile, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (m *MCPServer) getStats(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	summaries, err := m.db.GetApplicationSummary()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if len(summaries) == 0 {
		return mcp.NewToolResultText("No applications yet."), nil
	}
	data, _ := json.MarshalIndent(summaries, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

func (m *MCPServer) analyzeSkillGap(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    args, _ := req.Params.Arguments.(map[string]interface{})
    minScore, _ := args["min_score"].(float64)
    if minScore == 0 {
        minScore = 60
    }

    jobs, err := m.db.ListJobs(minScore, "", false, false, false, false, false)
    if err != nil {
        return mcp.NewToolResultError(err.Error()), nil
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
        return mcp.NewToolResultText("No scored jobs found. Run a scan first."), nil
    }

    type skillGap struct {
        Skill     string  `json:"skill"`
        Count     int     `json:"missing_in_jobs"`
        Frequency float64 `json:"frequency_pct"`
    }

    gaps := make([]skillGap, 0, len(freq))
    for skill, count := range freq {
        gaps = append(gaps, skillGap{
            Skill:     skill,
            Count:     count,
            Frequency: float64(count) / float64(total) * 100,
        })
    }

    // Sort by frequency descending
    sort.Slice(gaps, func(i, j int) bool {
        return gaps[i].Count > gaps[j].Count
    })

    // Top 10
    if len(gaps) > 10 {
        gaps = gaps[:10]
    }

    data, _ := json.MarshalIndent(gaps, "", "  ")
    return mcp.NewToolResultText(fmt.Sprintf(
        "Top missing skills across %d jobs (score >= %.0f):\n%s",
        total, minScore, string(data),
    )), nil
}


func (m *MCPServer) ServeStdio() error {
	return mcpserver.ServeStdio(m.server)
}

func (m *MCPServer) ServeSSE(port int) error {
	sseServer := mcpserver.NewSSEServer(m.server)
	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("MCP SSE server listening on http://localhost%s\n", addr)
	return sseServer.Start(addr)
}

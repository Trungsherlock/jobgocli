# JobGo

A Go CLI that crawls jobs from your target companies, scores them against your skill profile, and notifies you in real-time — with H1B visa sponsorship tracking for international students and new grads.

## Features

- **Skill-based scoring** — Extracts required/preferred/mentioned skills from job descriptions and scores 0–100 based on weighted overlap with your profile
- **LLM matching** — Optional Claude-powered semantic skill scoring (keyword, llm, or hybrid mode)
- **Composable filters** — Title, location, new-grad, and H1B filters work independently from scoring
- **Skill taxonomy** — 80+ canonical skills with alias resolution (`k8s`→`Kubernetes`, `golang`→`Go`, etc.)
- **Skill gap analysis** — Finds which skills appear most in your top jobs but are missing from your profile
- **Multi-platform scraping** — Lever and Greenhouse career pages, concurrent worker pool
- **H1B sponsorship tracking** — Import USCIS employer data, auto-link companies, filter by sponsor status
- **Watch mode** — Background polling with profile-aware filtering, desktop/terminal/webhook notifications
- **Application tracking** — Pipeline from `new` → `applied` → `interview` → `offer`
- **REST API + Chrome extension** — Browse and filter jobs from a side panel
- **MCP server** — Expose tools for AI agents via Model Context Protocol

## Architecture

```
cmd/jobgo/              CLI entrypoint
internal/
  cli/                  Cobra command definitions
  database/             SQLite + migration runner + repositories
  scraper/              Scraper interface + Lever/Greenhouse adapters
  skills/               Skill taxonomy, alias resolution, job/resume extractor
  matcher/              Keyword scorer, LLM scorer, hybrid pipeline
  filter/               Composable filters: title, location, new-grad, H1B
  worker/               Goroutine worker pool
  notifier/             Terminal, desktop, and webhook notifiers
  server/               REST API (chi) + MCP server (stdio/SSE)
  h1b/                  H1B importer, classifier, and scorer
migrations/             Versioned SQL migrations (001–005)
data/                   companies.csv, h1b_employers.csv
extension/              Chrome MV3 side panel
```

---

## Quick Start

### 1. Install

```bash
git clone https://github.com/Trungsherlock/jobgo.git
cd jobgocli
make install   # installs 'jobgo' to $GOPATH/bin
```

`make install` runs `go install ./cmd/jobgo` from the repo root — after that, `jobgo` is available anywhere in your terminal.

To just build locally without installing:

```bash
go build -o bin/jobgo ./cmd/jobgo
./bin/jobgo --help
```

### 2. Set up your profile

```bash
jobgo profile set \
  --name "Jane Smith" \
  --skills "Go,PostgreSQL,Docker,Kubernetes,AWS" \
  --roles "backend engineer,SRE" \
  --locations "remote,San Francisco" \
  --experience 1 \
  --visa          # include if you need H1B sponsorship

jobgo profile show
```

Skills are normalized automatically — `k8s`, `golang`, `postgres` are resolved to their canonical names.

### 3. Add companies to track

```bash
# One at a time
jobgo company add --name "Stripe" --platform lever --slug stripe
jobgo company add --name "Airbnb" --platform greenhouse --slug airbnb

# Or bulk import
jobgo company import data/companies.csv
jobgo company list
```

The CSV format is `name,platform,slug`. Edit [data/companies.csv](data/companies.csv) to add your targets.

### 4. Scrape and score jobs

```bash
jobgo search
```

This scrapes all enabled companies, stores new jobs, and scores each one against your profile. A score of 80+ means you match most required skills.

---

## Daily Workflow

### Browse jobs

```bash
# All jobs, sorted by skill score
jobgo jobs list

# Filter by minimum score
jobgo jobs list --min-score 60

# Only new (unseen) jobs above a threshold
jobgo jobs list --new --min-score 50

# Filter by title keyword
jobgo jobs list --title "backend engineer,SRE"

# Filter by location (supports aliases: US, UK, remote)
jobgo jobs list --location "US,remote"

# New grad roles only
jobgo jobs list --new-grad

# H1B sponsors only
jobgo jobs list --h1b

# Combine any filters
jobgo jobs list --min-score 60 --title "software engineer" --location "remote" --h1b

# JSON output
jobgo jobs list --output json | jq '.[].title'

# View full job details (description + skill match breakdown)
jobgo jobs show <job-id>

# Open in browser
jobgo jobs open <job-id>
```

### Understand your skill gaps

```bash
# List all 80+ skills in the taxonomy
jobgo skills list

# See which skills appear most in your top jobs but you're missing
jobgo skills gap
jobgo skills gap --min-score 60 --top 15
```

Example output:
```
Top 10 missing skills across 47 scored jobs (score >= 50):

   1. Kafka                    missing in 23 jobs (49%)
   2. Terraform                missing in 19 jobs (40%)
   3. AWS                      missing in 17 jobs (36%)
   ...
```

### Track applications

```bash
# Mark as applied
jobgo apply <job-id> --notes "Applied via website"

# Update status
jobgo jobs update <job-id> --status interview --notes "Phone screen Friday"

# View pipeline summary
jobgo status
```

Valid statuses: `new`, `applied`, `interview`, `offer`, `rejected`, `withdrawn`

### Watch mode

```bash
# Poll every 30 minutes, notify on high-match new jobs
jobgo watch --interval 30m --min-score 50
```

Watch mode scrapes, scores new jobs, then applies your profile filters (preferred roles, locations, visa requirement) before sending notifications. Press `Ctrl+C` to stop.

---

## H1B Workflow

For international students and new grads needing sponsorship:

```bash
# 1. Mark visa required in profile
jobgo profile set --visa

# 2. Download USCIS H1B Employer Data Hub CSV from uscis.gov and import
jobgo h1b import data/h1b_employers.csv
jobgo h1b status

# 3. Scrape and score
jobgo search

# 4. Filter to H1B sponsors only
jobgo jobs list --h1b --min-score 50

# 5. Combine with new-grad filter
jobgo jobs list --h1b --new-grad --location "remote,US"
```

---

## Scoring System

JobGo scores jobs on **skill match only** (0–100). Filters like title, location, new-grad, and H1B are applied separately after scoring — so you always see the true skill fit regardless of where you want to work.

### How scores are calculated

The job description is parsed into three sections:

| Section | Triggered by | Weight |
|---------|-------------|--------|
| Required | "Requirements:", "Qualifications:" | 70% |
| Preferred | "Nice to have:", "Preferred qualifications:" | 20% |
| Mentioned | Everything else | 10% |

```
score = (matched_required / total_required) × 70
      + (matched_preferred / total_preferred) × 20
      + (matched_mentioned / total_mentioned) × 10
```

### Matcher modes

Configure in `~/.jobgo/config.yaml`:

```yaml
matcher:
  type: hybrid      # keyword (default), llm, or hybrid
  llm_threshold: 30 # only call LLM if keyword score >= this

anthropic_api_key: sk-ant-...
```

| Mode | Description |
|------|-------------|
| `keyword` | Fast, deterministic, no API key needed |
| `llm` | Claude scores each job (slow, costs tokens) |
| `hybrid` | Keyword first; LLM only if score ≥ threshold (best balance) |

---

## Configuration

Config file: `~/.jobgo/config.yaml`

```yaml
matcher:
  type: hybrid
  llm_threshold: 30

anthropic_api_key: sk-ant-...

notify:
  - desktop
  # - webhook

# webhook_url: https://hooks.slack.com/services/...
```

Or set via environment variable:

```bash
export ANTHROPIC_API_KEY=sk-ant-...
```

---

## API Server

```bash
# REST API (for Chrome extension or custom integrations)
jobgo serve --port 8080

# MCP server (stdio, for Claude Code / Claude Desktop)
jobgo serve --mcp

# MCP server (SSE, for remote clients)
jobgo serve --mcp-sse --port 9090
```

### REST Endpoints

| Method | Path | Query params |
|--------|------|--------------|
| GET | `/api/jobs` | `min_score`, `company_id`, `new`, `title`, `location`, `h1b`, `new_grad`, `in_cart` |
| GET | `/api/jobs/:id` | — |
| GET | `/api/companies` | — |
| POST | `/api/companies` | body: `{name, platform, slug}` |
| DELETE | `/api/companies/:id` | — |
| GET | `/api/profile` | — |
| GET | `/api/stats` | — |
| GET | `/api/h1b/sponsors` | — |
| GET | `/api/h1b/status` | — |
| GET | `/api/jobcart` | — |
| POST | `/api/jobcart/:id` | — |
| DELETE | `/api/jobcart/:id` | — |
| POST | `/api/jobcart/scan` | — |

### MCP Tools (for Claude Code / Claude Desktop)

Add to your Claude config:

```json
{
  "mcpServers": {
    "jobgo": {
      "command": "jobgo",
      "args": ["serve", "--mcp"]
    }
  }
}
```

| Tool | Description |
|------|-------------|
| `search_jobs` | Search with `min_score`, `title`, `location`, `new_only`, `new_grad`, `h1b_only` |
| `get_job_details` | Full description + skill match breakdown |
| `list_companies` | Tracked companies + H1B status |
| `get_profile` | User profile |
| `get_stats` | Application pipeline counts |
| `analyze_skill_gap` | Top missing skills across scored jobs |

---

## Chrome Extension

Load `extension/` as an unpacked extension in Chrome (`chrome://extensions` → Developer mode → Load unpacked).

In the Settings tab, set:
- **Backend URL**: `http://localhost:8080`
- **Min Score**: minimum skill score to display

The Jobs tab supports live filtering by title, location, new-grad, and H1B toggle.

---

## Supported Platforms

| Platform | API |
|----------|-----|
| Lever | `api.lever.co/v0/postings/{slug}` |
| Greenhouse | `boards.greenhouse.io/v1/boards/{slug}/jobs` |

---

## Development

```bash
# Build
go build -o bin/jobgo ./cmd/jobgo

# Run all tests
go test ./...

# Test specific packages
go test ./internal/skills/... ./internal/matcher/... ./internal/filter/...

# On Windows: kill old process before rebuilding
Stop-Process -Name jobgo -Force   # PowerShell
```

## Tech Stack

- **Go** — CLI, concurrency, HTTP server
- **SQLite** (`modernc.org/sqlite`) — Zero-config embedded database
- **Cobra/Viper** — CLI framework + config management
- **Chi** — Lightweight HTTP router
- **mcp-go** — Model Context Protocol server SDK
- **Claude API** — LLM-powered job matching
- **USCIS H1B Employer Data Hub** — Visa sponsorship history

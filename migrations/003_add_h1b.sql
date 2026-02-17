CREATE TABLE IF NOT EXISTS h1b_sponsors (
    id TEXT PRIMARY KEY,
    company_name TEXT NOT NULL,
    normalized_name TEXT NOT NULL,
    city TEXT,
    state TEXT,
    naics_code TEXT,
    fiscal_year INTEGER,
    initial_approvals INTEGER,
    initial_denials INTEGER,
    continuing_approvals INTEGER,
    continuing_denials INTEGER,
    approval_rate REAL,
    total_petitions INTEGER
);

CREATE TABLE IF NOT EXISTS h1b_lcas (
    id TEXT PRIMARY KEY,
    employer_name TEXT,
    job_title TEXT,
    soc_code TEXT,
    wage_from REAL,
    wage_to REAL,
    wage_unit TEXT,
    worksite_city TEXT,
    worksite_state TEXT,
    submit_date DATE,
    decision_date DATE,
    status TEXT
);

CREATE INDEX idx_h1b_sponsors_normalized ON h1b_sponsors(normalized_name);

CREATE INDEX idx_h1b_lcas_employer ON h1b_lcas(employer_name);

ALTER TABLE companies ADD COLUMN h1b_sponsor_id TEXT;

ALTER TABLE companies ADD COLUMN sponsors_h1b BOOLEAN DEFAULT 0;

ALTER TABLE companies ADD COLUMN h1b_approval_rate REAL;

ALTER TABLE companies ADD COLUMN h1b_total_filed INTEGER;

ALTER TABLE jobs ADD COLUMN experience_level TEXT;

ALTER TABLE jobs ADD COLUMN visa_mentioned BOOLEAN DEFAULT 0;

ALTER TABLE jobs ADD COLUMN visa_sentiment TEXT;

ALTER TABLE jobs ADD COLUMN is_new_grad BOOLEAN DEFAULT 0;

ALTER TABLE profile ADD COLUMN visa_required BOOLEAN DEFAULT 0;

ALTER TABLE profile ADD COLUMN experience_level TEXT;
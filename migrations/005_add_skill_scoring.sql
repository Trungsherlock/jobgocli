ALTER TABLE jobs ADD COLUMN skill_score REAL;
ALTER TABLE jobs ADD COLUMN skill_matched TEXT;
ALTER TABLE jobs ADD COLUMN skill_missing TEXT;
ALTER TABLE jobs ADD COLUMN skill_reason TEXT;
ALTER TABLE jobs ADD COLUMN skill_scored_at DATETIME;
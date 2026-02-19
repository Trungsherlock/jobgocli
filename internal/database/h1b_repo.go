package database

func (d *DB) UpsertH1bSponsor(s H1bSponsors) error {
	_, err := d.Exec(`
	INSERT INTO h1b_sponsors (id, company_name, normalized_name, city, state, naics_code,
	fiscal_year, initial_approvals, initial_denials, continuing_approvals, continuing_denials,
	approval_rate, total_petitions)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		approval_rate = excluded.approval_rate,
		total_petitions = excluded.total_petitions`,
		s.ID, s.CompanyName, s.NormalizedName, s.City, s.State, s.NaicsCode,
		s.FiscalYear, s.InitialApprovals, s.InitialDenials, s.ContinuingApprovals, s.ContinuingDenials,
		s.ApprovalRate, s.TotalPetitions,
	)
	return err
}

func (d *DB) FindSponsorByName(normalizedName string) (*H1bSponsors, error) {
	s := &H1bSponsors{}
	err := d.QueryRow(
		`SELECT id, company_name, normalized_name, city, state, naics_code,
		fiscal_year, initial_approvals, initial_denials, continuing_approvals, continuing_denials,
		approval_rate, total_petitions
		FROM h1b_sponsors WHERE normalized_name = ?`, normalizedName,
	).Scan(&s.ID, &s.CompanyName, &s.NormalizedName, &s.City, &s.State, &s.NaicsCode,
	&s.FiscalYear, &s.InitialApprovals, &s.InitialDenials, &s.ContinuingApprovals, &s.ContinuingDenials,
	&s.ApprovalRate, &s.TotalPetitions)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (d *DB) LinkCompanyToSponsor(companyID, sponsorID string, approvalRate float64, totalFiled int) error {
	_, err := d.Exec(
		`UPDATE companies SET h1b_sponsor_id = ?, sponsors_h1b = 1, h1b_approval_rate = ?, h1b_total_filed = ? WHERE id = ?`,
		sponsorID, approvalRate, totalFiled, companyID,
	)
	return err
}

func (d *DB) CountSponsors() (int, error) {
	var count int
	err := d.QueryRow("SELECT COUNT(*) FROM h1b_sponsors").Scan(&count)
	return count, err
}
package drift

func (s *GeneratedTestStore) persistGeneratedTestDB(gt *GeneratedTest) error {
	ctx, cancel := driftDBContext()
	defer cancel()
	_, err := s.dbPool.Exec(ctx, `
		INSERT INTO drift_generated_tests (id, drift_id, test_name, test_code, confidence_score, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			drift_id = EXCLUDED.drift_id,
			test_name = EXCLUDED.test_name,
			test_code = EXCLUDED.test_code,
			confidence_score = EXCLUDED.confidence_score,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at`,
		gt.ID, gt.DriftID, gt.TestName, gt.TestCode, gt.ConfidenceScore, gt.Status, gt.CreatedAt, gt.UpdatedAt)
	return err
}

func (s *GeneratedTestStore) generatedTestsByDriftDB(driftID string) ([]GeneratedTest, error) {
	ctx, cancel := driftDBContext()
	defer cancel()
	rows, err := s.dbPool.Query(ctx, `
		SELECT id, drift_id, test_name, test_code, COALESCE(confidence_score, 0), status, created_at, updated_at
		FROM drift_generated_tests WHERE drift_id = $1 ORDER BY created_at DESC`, driftID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tests []GeneratedTest
	for rows.Next() {
		gt, err := scanGeneratedTest(rows)
		if err != nil {
			return tests, err
		}
		tests = append(tests, *gt)
	}
	return tests, rows.Err()
}

func (s *GeneratedTestStore) getGeneratedTestDB(id string) (*GeneratedTest, error) {
	ctx, cancel := driftDBContext()
	defer cancel()
	row := s.dbPool.QueryRow(ctx, `
		SELECT id, drift_id, test_name, test_code, COALESCE(confidence_score, 0), status, created_at, updated_at
		FROM drift_generated_tests WHERE id = $1`, id)
	return scanGeneratedTest(row)
}

func scanGeneratedTest(row interface{ Scan(dest ...any) error }) (*GeneratedTest, error) {
	var gt GeneratedTest
	if err := row.Scan(&gt.ID, &gt.DriftID, &gt.TestName, &gt.TestCode, &gt.ConfidenceScore, &gt.Status, &gt.CreatedAt, &gt.UpdatedAt); err != nil {
		return nil, err
	}
	return &gt, nil
}

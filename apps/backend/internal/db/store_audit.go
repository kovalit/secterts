package db

import "context"

// InsertAuditLog appends an audit entry. metadata must be valid JSON bytes.
func (s *Store) InsertAuditLog(ctx context.Context, l *AuditLog) error {
	metadata := l.Metadata
	if len(metadata) == 0 {
		metadata = []byte("{}")
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO audit_logs (user_id, company_id, action, entity_type, entity_id, ip, user_agent, metadata)
		VALUES ($1,$2,$3,$4,$5,$6::inet,$7,$8)`,
		l.UserID, l.CompanyID, l.Action, l.EntityType, l.EntityID, l.IP, l.UserAgent, metadata)
	return err
}

// ListAuditLogs returns recent audit entries for a user (most recent first).
func (s *Store) ListAuditLogs(ctx context.Context, userID string, limit int) ([]AuditLog, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, company_id, action, entity_type, entity_id, ip::text, user_agent, metadata, created_at
		FROM audit_logs WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []AuditLog
	for rows.Next() {
		var l AuditLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.CompanyID, &l.Action, &l.EntityType, &l.EntityID, &l.IP, &l.UserAgent, &l.Metadata, &l.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

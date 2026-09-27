package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

// CreateBackupExport records the start of a backup run.
func (s *Store) CreateBackupExport(ctx context.Context, createdBy *string) (*BackupExport, error) {
	var b BackupExport
	err := s.pool.QueryRow(ctx, `
		INSERT INTO backup_exports (created_by, status) VALUES ($1, 'started')
		RETURNING id, created_by, status, file_name, storage_target, sha256, error_message, created_at, finished_at`, createdBy,
	).Scan(&b.ID, &b.CreatedBy, &b.Status, &b.FileName, &b.StorageTarget, &b.SHA256, &b.ErrorMessage, &b.CreatedAt, &b.FinishedAt)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// FinishBackupExport marks a backup run success/failed with metadata.
func (s *Store) FinishBackupExport(ctx context.Context, id, status string, fileName, storageTarget, sha256, errMsg *string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE backup_exports
		SET status=$2, file_name=$3, storage_target=$4, sha256=$5, error_message=$6, finished_at=now()
		WHERE id=$1`, id, status, fileName, storageTarget, sha256, errMsg)
	return err
}

// ListBackupExports returns recent backup runs.
func (s *Store) ListBackupExports(ctx context.Context, limit int) ([]BackupExport, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, created_by, status, file_name, storage_target, sha256, error_message, created_at, finished_at
		FROM backup_exports ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []BackupExport
	for rows.Next() {
		var b BackupExport
		if err := rows.Scan(&b.ID, &b.CreatedBy, &b.Status, &b.FileName, &b.StorageTarget, &b.SHA256, &b.ErrorMessage, &b.CreatedAt, &b.FinishedAt); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// GetBackupExport returns one backup run by id.
func (s *Store) GetBackupExport(ctx context.Context, id string) (*BackupExport, error) {
	var b BackupExport
	err := s.pool.QueryRow(ctx, `
		SELECT id, created_by, status, file_name, storage_target, sha256, error_message, created_at, finished_at
		FROM backup_exports WHERE id = $1`, id,
	).Scan(&b.ID, &b.CreatedBy, &b.Status, &b.FileName, &b.StorageTarget, &b.SHA256, &b.ErrorMessage, &b.CreatedAt, &b.FinishedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

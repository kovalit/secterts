package db

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

const secretNoteColumns = `
	id, owner_user_id, title, type, encrypted_text, text_nonce, text_key_version,
	created_at, updated_at, deleted_at`

func scanSecretNote(row pgx.Row) (*SecretNote, error) {
	var n SecretNote
	err := row.Scan(&n.ID, &n.OwnerUserID, &n.Title, &n.Type,
		&n.EncryptedText, &n.TextNonce, &n.TextKeyVersion,
		&n.CreatedAt, &n.UpdatedAt, &n.DeletedAt)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

// SecretNoteFilter holds optional list filters.
type SecretNoteFilter struct {
	Type  string
	Query string
}

// ListSecretNotes returns non-deleted notes for a user matching filters.
func (s *Store) ListSecretNotes(ctx context.Context, ownerUserID string, f SecretNoteFilter) ([]SecretNote, error) {
	conds := []string{"owner_user_id = $1", "deleted_at IS NULL"}
	args := []any{ownerUserID}

	if f.Type != "" {
		args = append(args, f.Type)
		conds = append(conds, fmt.Sprintf("type = $%d", len(args)))
	}
	if f.Query != "" {
		args = append(args, "%"+strings.ToLower(f.Query)+"%")
		conds = append(conds, fmt.Sprintf("lower(title) LIKE $%d", len(args)))
	}

	query := "SELECT " + secretNoteColumns + " FROM secret_notes WHERE " +
		strings.Join(conds, " AND ") + " ORDER BY updated_at DESC"

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []SecretNote
	for rows.Next() {
		n, err := scanSecretNote(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *n)
	}
	return out, rows.Err()
}

// GetSecretNote returns a single non-deleted note owned by the user.
func (s *Store) GetSecretNote(ctx context.Context, ownerUserID, id string) (*SecretNote, error) {
	row := s.pool.QueryRow(ctx, "SELECT "+secretNoteColumns+
		" FROM secret_notes WHERE id = $1 AND owner_user_id = $2 AND deleted_at IS NULL", id, ownerUserID)
	n, err := scanSecretNote(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return n, err
}

// CreateSecretNote inserts a new encrypted note.
func (s *Store) CreateSecretNote(ctx context.Context, n *SecretNote) (*SecretNote, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO secret_notes (owner_user_id, title, type, encrypted_text, text_nonce, text_key_version)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING `+secretNoteColumns,
		n.OwnerUserID, n.Title, n.Type, n.EncryptedText, n.TextNonce, n.TextKeyVersion)
	return scanSecretNote(row)
}

// UpdateSecretNote rewrites an existing note.
func (s *Store) UpdateSecretNote(ctx context.Context, n *SecretNote) (*SecretNote, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE secret_notes SET
			title=$3, type=$4, encrypted_text=$5, text_nonce=$6, text_key_version=$7, updated_at=now()
		WHERE id=$1 AND owner_user_id=$2 AND deleted_at IS NULL
		RETURNING `+secretNoteColumns,
		n.ID, n.OwnerUserID, n.Title, n.Type, n.EncryptedText, n.TextNonce, n.TextKeyVersion)
	res, err := scanSecretNote(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return res, err
}

// SoftDeleteSecretNote marks a note as deleted.
func (s *Store) SoftDeleteSecretNote(ctx context.Context, ownerUserID, id string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE secret_notes SET deleted_at = now()
		WHERE id = $1 AND owner_user_id = $2 AND deleted_at IS NULL`, id, ownerUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

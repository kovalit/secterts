package db

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// CreateExtensionToken stores a hashed extension token.
func (s *Store) CreateExtensionToken(ctx context.Context, userID, tokenHash, name string, scopes []string, expiresAt time.Time) (*ExtensionToken, error) {
	var t ExtensionToken
	err := s.pool.QueryRow(ctx, `
		INSERT INTO extension_tokens (user_id, token_hash, name, scopes, expires_at)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id, user_id, token_hash, name, scopes, expires_at, revoked_at, created_at, last_used_at`,
		userID, tokenHash, name, scopes, expiresAt,
	).Scan(&t.ID, &t.UserID, &t.TokenHash, &t.Name, &t.Scopes, &t.ExpiresAt, &t.RevokedAt, &t.CreatedAt, &t.LastUsedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// ListExtensionTokens returns a user's tokens, newest first.
func (s *Store) ListExtensionTokens(ctx context.Context, userID string) ([]ExtensionToken, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, token_hash, name, scopes, expires_at, revoked_at, created_at, last_used_at
		FROM extension_tokens WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ExtensionToken
	for rows.Next() {
		var t ExtensionToken
		if err := rows.Scan(&t.ID, &t.UserID, &t.TokenHash, &t.Name, &t.Scopes, &t.ExpiresAt, &t.RevokedAt, &t.CreatedAt, &t.LastUsedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// GetActiveTokenByHash returns a valid (non-revoked, non-expired) token.
func (s *Store) GetActiveTokenByHash(ctx context.Context, tokenHash string) (*ExtensionToken, error) {
	var t ExtensionToken
	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id, token_hash, name, scopes, expires_at, revoked_at, created_at, last_used_at
		FROM extension_tokens
		WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > now()`, tokenHash,
	).Scan(&t.ID, &t.UserID, &t.TokenHash, &t.Name, &t.Scopes, &t.ExpiresAt, &t.RevokedAt, &t.CreatedAt, &t.LastUsedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// TouchTokenLastUsed updates last_used_at for a token.
func (s *Store) TouchTokenLastUsed(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx, `UPDATE extension_tokens SET last_used_at = now() WHERE id = $1`, id)
	return err
}

// RevokeExtensionToken revokes a token owned by the user.
func (s *Store) RevokeExtensionToken(ctx context.Context, userID, id string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE extension_tokens SET revoked_at = now()
		WHERE id = $1 AND user_id = $2 AND revoked_at IS NULL`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

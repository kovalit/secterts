package db

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// CreateUser inserts a new user and returns the created row.
func (s *Store) CreateUser(ctx context.Context, email, passwordHash string) (*User, error) {
	var u User
	err := s.pool.QueryRow(ctx, `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, email, password_hash, email_verified, email_2fa_enabled, is_active, created_at, updated_at`,
		email, passwordHash,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.EmailVerified, &u.Email2FAEnabled, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetUserByEmail looks up a user by (case-insensitive) email.
func (s *Store) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return s.scanUser(ctx, `
		SELECT id, email, password_hash, email_verified, email_2fa_enabled, is_active, created_at, updated_at
		FROM users WHERE email = $1`, email)
}

// GetUserByID looks up a user by id.
func (s *Store) GetUserByID(ctx context.Context, id string) (*User, error) {
	return s.scanUser(ctx, `
		SELECT id, email, password_hash, email_verified, email_2fa_enabled, is_active, created_at, updated_at
		FROM users WHERE id = $1`, id)
}

func (s *Store) scanUser(ctx context.Context, query string, args ...any) (*User, error) {
	var u User
	err := s.pool.QueryRow(ctx, query, args...).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.EmailVerified, &u.Email2FAEnabled, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// SetEmailVerified marks a user's email as verified.
func (s *Store) SetEmailVerified(ctx context.Context, userID string) error {
	_, err := s.pool.Exec(ctx, `UPDATE users SET email_verified = TRUE, updated_at = now() WHERE id = $1`, userID)
	return err
}

// UpdateUserPassword replaces the stored password hash.
func (s *Store) UpdateUserPassword(ctx context.Context, userID, passwordHash string) error {
	_, err := s.pool.Exec(ctx, `UPDATE users SET password_hash = $2, updated_at = now() WHERE id = $1`, userID, passwordHash)
	return err
}

// --- email auth codes ---

// CreateEmailAuthCode stores a hashed 2FA/verification code.
func (s *Store) CreateEmailAuthCode(ctx context.Context, userID, codeHash, purpose string, expiresAt time.Time) (*EmailAuthCode, error) {
	var c EmailAuthCode
	err := s.pool.QueryRow(ctx, `
		INSERT INTO email_auth_codes (user_id, code_hash, purpose, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, code_hash, purpose, expires_at, consumed_at, attempts, created_at`,
		userID, codeHash, purpose, expiresAt,
	).Scan(&c.ID, &c.UserID, &c.CodeHash, &c.Purpose, &c.ExpiresAt, &c.ConsumedAt, &c.Attempts, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// GetEmailAuthCodeByID returns a code row by its id and expected purpose.
func (s *Store) GetEmailAuthCodeByID(ctx context.Context, id, purpose string) (*EmailAuthCode, error) {
	var c EmailAuthCode
	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id, code_hash, purpose, expires_at, consumed_at, attempts, created_at
		FROM email_auth_codes
		WHERE id = $1 AND purpose = $2`,
		id, purpose,
	).Scan(&c.ID, &c.UserID, &c.CodeHash, &c.Purpose, &c.ExpiresAt, &c.ConsumedAt, &c.Attempts, &c.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// IncrementCodeAttempts bumps the attempts counter for a code.
func (s *Store) IncrementCodeAttempts(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx, `UPDATE email_auth_codes SET attempts = attempts + 1 WHERE id = $1`, id)
	return err
}

// ConsumeCode marks a code as consumed so it cannot be reused.
func (s *Store) ConsumeCode(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx, `UPDATE email_auth_codes SET consumed_at = now() WHERE id = $1`, id)
	return err
}

// --- sessions ---

// CreateSession stores a new session with a hashed refresh token.
func (s *Store) CreateSession(ctx context.Context, userID, refreshTokenHash string, userAgent, ip *string, expiresAt time.Time) (*Session, error) {
	var sess Session
	err := s.pool.QueryRow(ctx, `
		INSERT INTO sessions (user_id, refresh_token_hash, user_agent, ip, expires_at)
		VALUES ($1, $2, $3, $4::inet, $5)
		RETURNING id, user_id, refresh_token_hash, user_agent, ip::text, expires_at, revoked_at, created_at`,
		userID, refreshTokenHash, userAgent, ip, expiresAt,
	).Scan(&sess.ID, &sess.UserID, &sess.RefreshTokenHash, &sess.UserAgent, &sess.IP, &sess.ExpiresAt, &sess.RevokedAt, &sess.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &sess, nil
}

// GetSessionByTokenHash returns an active (non-revoked, non-expired) session.
func (s *Store) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*Session, error) {
	var sess Session
	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id, refresh_token_hash, user_agent, ip::text, expires_at, revoked_at, created_at
		FROM sessions
		WHERE refresh_token_hash = $1 AND revoked_at IS NULL AND expires_at > now()`,
		tokenHash,
	).Scan(&sess.ID, &sess.UserID, &sess.RefreshTokenHash, &sess.UserAgent, &sess.IP, &sess.ExpiresAt, &sess.RevokedAt, &sess.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &sess, nil
}

// ListSessions returns all sessions for a user, newest first.
func (s *Store) ListSessions(ctx context.Context, userID string) ([]Session, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, refresh_token_hash, user_agent, ip::text, expires_at, revoked_at, created_at
		FROM sessions WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Session
	for rows.Next() {
		var sess Session
		if err := rows.Scan(&sess.ID, &sess.UserID, &sess.RefreshTokenHash, &sess.UserAgent, &sess.IP, &sess.ExpiresAt, &sess.RevokedAt, &sess.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, sess)
	}
	return out, rows.Err()
}

// RevokeSession revokes a single session owned by the user.
func (s *Store) RevokeSession(ctx context.Context, userID, sessionID string) error {
	_, err := s.pool.Exec(ctx, `UPDATE sessions SET revoked_at = now() WHERE id = $1 AND user_id = $2 AND revoked_at IS NULL`, sessionID, userID)
	return err
}

// RevokeSessionByTokenHash revokes the session matching a refresh token (logout).
func (s *Store) RevokeSessionByTokenHash(ctx context.Context, tokenHash string) error {
	_, err := s.pool.Exec(ctx, `UPDATE sessions SET revoked_at = now() WHERE refresh_token_hash = $1 AND revoked_at IS NULL`, tokenHash)
	return err
}

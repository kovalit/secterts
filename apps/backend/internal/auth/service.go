package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/kovalit/secrets-center/backend/internal/db"
	"github.com/kovalit/secrets-center/backend/internal/mailer"
)

const (
	codeLength      = 6
	codeTTL         = 10 * time.Minute
	maxCodeAttempts = 5
	sessionTTL      = 30 * 24 * time.Hour
	sessionTokenLen = 32

	purposeLogin2FA    = "login_2fa"
	purposePasswordRst = "password_reset"
)

// Public sentinel errors mapped to HTTP codes by the handler.
var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidCode        = errors.New("invalid or expired code")
	ErrTooManyAttempts    = errors.New("too many attempts")
)

// Service holds dependencies for auth operations.
type Service struct {
	store  *db.Store
	mailer mailer.Mailer
}

// NewService builds the auth service.
func NewService(store *db.Store, m mailer.Mailer) *Service {
	return &Service{store: store, mailer: m}
}

// Register creates a new user account.
func (s *Service) Register(ctx context.Context, email, password string) (*db.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if _, err := s.store.GetUserByEmail(ctx, email); err == nil {
		return nil, ErrEmailTaken
	} else if !errors.Is(err, db.ErrNotFound) {
		return nil, err
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}
	return s.store.CreateUser(ctx, email, hash)
}

// Login verifies credentials (step 1) and issues an email 2FA code. It returns
// the challenge id (the code row id) the client passes to VerifyEmailCode.
func (s *Service) Login(ctx context.Context, email, password string) (challengeID string, user *db.User, err error) {
	email = strings.TrimSpace(strings.ToLower(email))
	user, err = s.store.GetUserByEmail(ctx, email)
	if errors.Is(err, db.ErrNotFound) {
		// Run a dummy verify to reduce timing side-channels, then fail.
		_, _ = VerifyPassword(password, "$argon2id$v=19$m=65536,t=1,p=4$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
		return "", nil, ErrInvalidCredentials
	}
	if err != nil {
		return "", nil, err
	}
	if !user.IsActive {
		return "", nil, ErrInvalidCredentials
	}

	ok, err := VerifyPassword(password, user.PasswordHash)
	if err != nil || !ok {
		return "", nil, ErrInvalidCredentials
	}

	challengeID, err = s.issueCode(ctx, user, purposeLogin2FA, "Secrets Center — код входа")
	if err != nil {
		return "", nil, err
	}
	return challengeID, user, nil
}

// VerifyEmailCode validates the 2FA code (step 2) and creates a session,
// returning the raw session token to set as a cookie.
func (s *Service) VerifyEmailCode(ctx context.Context, challengeID, code string, userAgent, ip *string) (token string, expiresAt time.Time, user *db.User, err error) {
	rec, err := s.consumeCode(ctx, challengeID, code, purposeLogin2FA)
	if err != nil {
		return "", time.Time{}, nil, err
	}

	user, err = s.store.GetUserByID(ctx, rec.UserID)
	if err != nil {
		return "", time.Time{}, nil, err
	}
	if !user.EmailVerified {
		_ = s.store.SetEmailVerified(ctx, user.ID)
	}

	token, err = generateToken(sessionTokenLen)
	if err != nil {
		return "", time.Time{}, nil, err
	}
	expiresAt = time.Now().Add(sessionTTL)
	if _, err = s.store.CreateSession(ctx, user.ID, hashToken(token), userAgent, ip, expiresAt); err != nil {
		return "", time.Time{}, nil, err
	}
	return token, expiresAt, user, nil
}

// Logout revokes the session identified by the raw token.
func (s *Service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.store.RevokeSessionByTokenHash(ctx, hashToken(token))
}

// PasswordResetRequest issues a reset code. To avoid leaking which emails
// exist, it returns an empty challenge id (and nil error) for unknown emails.
func (s *Service) PasswordResetRequest(ctx context.Context, email string) (challengeID string, err error) {
	email = strings.TrimSpace(strings.ToLower(email))
	user, err := s.store.GetUserByEmail(ctx, email)
	if errors.Is(err, db.ErrNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return s.issueCode(ctx, user, purposePasswordRst, "Secrets Center — сброс пароля")
}

// PasswordResetConfirm validates a reset code and sets a new password.
func (s *Service) PasswordResetConfirm(ctx context.Context, challengeID, code, newPassword string) error {
	rec, err := s.consumeCode(ctx, challengeID, code, purposePasswordRst)
	if err != nil {
		return err
	}
	hash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	return s.store.UpdateUserPassword(ctx, rec.UserID, hash)
}

// issueCode generates a numeric code, stores its hash and emails the plaintext.
func (s *Service) issueCode(ctx context.Context, user *db.User, purpose, subject string) (string, error) {
	code, err := generateNumericCode(codeLength)
	if err != nil {
		return "", err
	}
	rec, err := s.store.CreateEmailAuthCode(ctx, user.ID, hashToken(code), purpose, time.Now().Add(codeTTL))
	if err != nil {
		return "", err
	}

	body := fmt.Sprintf("Ваш код: %s\n\nКод действует %d минут. Если вы не запрашивали его, проигнорируйте это письмо.",
		code, int(codeTTL.Minutes()))
	if err := s.mailer.Send(user.Email, subject, body); err != nil {
		return "", err
	}
	return rec.ID, nil
}

// consumeCode validates a code by challenge id, enforcing expiry and attempts,
// then marks it consumed.
func (s *Service) consumeCode(ctx context.Context, challengeID, code, purpose string) (*db.EmailAuthCode, error) {
	if challengeID == "" || code == "" {
		return nil, ErrInvalidCode
	}
	// We do not know the user id yet; look up the code by id + purpose.
	rec, err := s.getCodeByID(ctx, challengeID, purpose)
	if err != nil {
		return nil, ErrInvalidCode
	}
	if rec.ConsumedAt != nil || time.Now().After(rec.ExpiresAt) {
		return nil, ErrInvalidCode
	}
	if rec.Attempts >= maxCodeAttempts {
		return nil, ErrTooManyAttempts
	}

	if hashToken(code) != rec.CodeHash {
		_ = s.store.IncrementCodeAttempts(ctx, rec.ID)
		return nil, ErrInvalidCode
	}
	if err := s.store.ConsumeCode(ctx, rec.ID); err != nil {
		return nil, err
	}
	return rec, nil
}

// getCodeByID fetches a code row by id and purpose using the user id it stores.
func (s *Service) getCodeByID(ctx context.Context, id, purpose string) (*db.EmailAuthCode, error) {
	return s.store.GetEmailAuthCodeByID(ctx, id, purpose)
}

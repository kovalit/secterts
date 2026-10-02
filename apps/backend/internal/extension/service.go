package extension

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/kovalit/secrets-center/backend/internal/crypto"
	"github.com/kovalit/secrets-center/backend/internal/db"
)

// tokenTTL is how long a new extension token stays valid.
const tokenTTL = 180 * 24 * time.Hour

// Errors surfaced to the handler.
var (
	ErrDomainMismatch = errors.New("domain does not match the entry")
	ErrNoPassword     = errors.New("entry has no password")
)

// Service implements extension token + lookup logic.
type Service struct {
	store *db.Store
	enc   crypto.Encryptor
}

// New builds the service.
func New(store *db.Store, enc crypto.Encryptor) *Service {
	return &Service{store: store, enc: enc}
}

// TokenView describes a stored token (without the secret value).
type TokenView struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Scopes     []string   `json:"scopes"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  time.Time  `json:"expires_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
	RevokedAt  *time.Time `json:"revoked_at"`
}

func toTokenView(t *db.ExtensionToken) TokenView {
	return TokenView{
		ID: t.ID, Name: t.Name, Scopes: t.Scopes,
		CreatedAt: t.CreatedAt, ExpiresAt: t.ExpiresAt, LastUsedAt: t.LastUsedAt, RevokedAt: t.RevokedAt,
	}
}

// CreateToken issues a new token; the raw value is returned only once.
func (s *Service) CreateToken(ctx context.Context, userID, name string) (rawToken string, view TokenView, err error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "Chrome Extension"
	}
	rawToken, err = generateToken()
	if err != nil {
		return "", TokenView{}, err
	}
	t, err := s.store.CreateExtensionToken(ctx, userID, hashToken(rawToken), name, []string{"passwords:read"}, time.Now().Add(tokenTTL))
	if err != nil {
		return "", TokenView{}, err
	}
	return rawToken, toTokenView(t), nil
}

// ListTokens returns a user's tokens.
func (s *Service) ListTokens(ctx context.Context, userID string) ([]TokenView, error) {
	tokens, err := s.store.ListExtensionTokens(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]TokenView, 0, len(tokens))
	for i := range tokens {
		out = append(out, toTokenView(&tokens[i]))
	}
	return out, nil
}

// RevokeToken revokes a token owned by the user.
func (s *Service) RevokeToken(ctx context.Context, userID, id string) error {
	return s.store.RevokeExtensionToken(ctx, userID, id)
}

// Authenticate resolves a raw bearer token to its active DB record.
func (s *Service) Authenticate(ctx context.Context, rawToken string) (*db.ExtensionToken, error) {
	tok, err := s.store.GetActiveTokenByHash(ctx, hashToken(rawToken))
	if err != nil {
		return nil, err
	}
	_ = s.store.TouchTokenLastUsed(ctx, tok.ID)
	return tok, nil
}

// LookupItem is a single lookup result (no password).
type LookupItem struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	Login      *string `json:"login"`
	FaviconURL *string `json:"favicon_url"`
	Scope      string  `json:"scope"`
}

// Lookup returns the user's entries for an exact domain.
func (s *Service) Lookup(ctx context.Context, userID, domain string) ([]LookupItem, error) {
	domain = strings.ToLower(strings.TrimSpace(domain))
	domain = strings.TrimPrefix(domain, "www.")
	entries, err := s.store.LookupByDomain(ctx, userID, domain)
	if err != nil {
		return nil, err
	}
	out := make([]LookupItem, 0, len(entries))
	for i := range entries {
		e := &entries[i]
		// Prefer a user-uploaded custom icon over the resolved favicon.
		icon := e.FaviconURL
		if e.CustomIcon != nil && *e.CustomIcon != "" {
			icon = e.CustomIcon
		}
		out = append(out, LookupItem{ID: e.ID, Title: e.Title, Login: e.Login, FaviconURL: icon, Scope: e.Scope})
	}
	return out, nil
}

// Reveal returns the plaintext password for an entry, checking domain match and
// ownership. Used by the extension only after an explicit user click.
func (s *Service) Reveal(ctx context.Context, userID, entryID, domain string) (string, error) {
	e, err := s.store.GetPasswordEntry(ctx, userID, entryID)
	if err != nil {
		return "", err
	}
	domain = strings.ToLower(strings.TrimSpace(domain))
	domain = strings.TrimPrefix(domain, "www.")
	if e.Domain == nil || *e.Domain != domain {
		return "", ErrDomainMismatch
	}
	if len(e.EncryptedPassword) == 0 {
		return "", ErrNoPassword
	}
	plaintext, err := s.enc.DecryptString(e.EncryptedPassword, e.PasswordNonce, e.PasswordKeyVersion)
	if err != nil {
		return "", err
	}
	// Record usage so the health check can surface stale, unused records.
	_ = s.store.TouchPasswordEntryUsed(ctx, userID, entryID)
	return plaintext, nil
}

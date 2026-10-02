// Package secretnotes implements free-form encrypted secret notes: a title, a
// type (api_key / token / credentials / ...) and an encrypted text body.
package secretnotes

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/kovalit/secrets-center/backend/internal/crypto"
	"github.com/kovalit/secrets-center/backend/internal/db"
)

// validTypes is the fixed set of secret-note types (mirrors the DB CHECK and
// the frontend type catalogue).
var validTypes = map[string]bool{
	"api_key":     true,
	"token":       true,
	"credentials": true,
	"server":      true,
	"database":    true,
	"ssh_key":     true,
	"certificate": true,
	"environment": true,
	"generic":     true,
}

// Errors surfaced to the handler.
var (
	ErrTitleReq    = errors.New("title is required")
	ErrTextReq     = errors.New("text is required")
	ErrInvalidType = errors.New("unknown secret note type")
)

// Service implements secret-note business logic.
type Service struct {
	store *db.Store
	enc   crypto.Encryptor
}

// New builds the service.
func New(store *db.Store, enc crypto.Encryptor) *Service {
	return &Service{store: store, enc: enc}
}

// View is the safe representation of a note (never includes the text).
type View struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Type      string    `json:"type"`
	HasText   bool      `json:"has_text"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toView(n *db.SecretNote) View {
	return View{
		ID:        n.ID,
		Title:     n.Title,
		Type:      n.Type,
		HasText:   len(n.EncryptedText) > 0,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
	}
}

// WriteInput is the payload for creating/updating a note.
type WriteInput struct {
	Title string
	Type  string
	Text  string // required on create; empty on update = keep existing
}

// List returns the user's notes as safe views.
func (s *Service) List(ctx context.Context, ownerID string, f db.SecretNoteFilter) ([]View, error) {
	notes, err := s.store.ListSecretNotes(ctx, ownerID, f)
	if err != nil {
		return nil, err
	}
	out := make([]View, 0, len(notes))
	for i := range notes {
		out = append(out, toView(&notes[i]))
	}
	return out, nil
}

// Get returns a single note view.
func (s *Service) Get(ctx context.Context, ownerID, id string) (*View, error) {
	n, err := s.store.GetSecretNote(ctx, ownerID, id)
	if err != nil {
		return nil, err
	}
	v := toView(n)
	return &v, nil
}

// Create validates, encrypts and stores a new note.
func (s *Service) Create(ctx context.Context, ownerID string, in WriteInput) (*View, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, ErrTitleReq
	}
	noteType, err := normalizeType(in.Type)
	if err != nil {
		return nil, err
	}
	if in.Text == "" {
		return nil, ErrTextReq
	}

	note := &db.SecretNote{OwnerUserID: ownerID, Title: title, Type: noteType}
	encText, nonce, ver, err := s.enc.EncryptString(in.Text)
	if err != nil {
		return nil, err
	}
	note.EncryptedText, note.TextNonce, note.TextKeyVersion = encText, nonce, ver

	created, err := s.store.CreateSecretNote(ctx, note)
	if err != nil {
		return nil, err
	}
	v := toView(created)
	return &v, nil
}

// Update validates and rewrites a note. An empty text keeps the old one.
func (s *Service) Update(ctx context.Context, ownerID, id string, in WriteInput) (*View, error) {
	existing, err := s.store.GetSecretNote(ctx, ownerID, id)
	if err != nil {
		return nil, err
	}
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, ErrTitleReq
	}
	noteType, err := normalizeType(in.Type)
	if err != nil {
		return nil, err
	}

	existing.Title = title
	existing.Type = noteType
	if in.Text != "" {
		encText, nonce, ver, err := s.enc.EncryptString(in.Text)
		if err != nil {
			return nil, err
		}
		existing.EncryptedText, existing.TextNonce, existing.TextKeyVersion = encText, nonce, ver
	}

	updated, err := s.store.UpdateSecretNote(ctx, existing)
	if err != nil {
		return nil, err
	}
	v := toView(updated)
	return &v, nil
}

// Delete soft-deletes a note.
func (s *Service) Delete(ctx context.Context, ownerID, id string) error {
	return s.store.SoftDeleteSecretNote(ctx, ownerID, id)
}

// Reveal decrypts and returns the plaintext text. The caller writes the audit
// entry.
func (s *Service) Reveal(ctx context.Context, ownerID, id string) (string, error) {
	n, err := s.store.GetSecretNote(ctx, ownerID, id)
	if err != nil {
		return "", err
	}
	return s.enc.DecryptString(n.EncryptedText, n.TextNonce, n.TextKeyVersion)
}

func normalizeType(t string) (string, error) {
	t = strings.TrimSpace(t)
	if t == "" {
		return "generic", nil
	}
	if !validTypes[t] {
		return "", ErrInvalidType
	}
	return t, nil
}

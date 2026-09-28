package passwords

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/kovalit/secrets-center/backend/internal/crypto"
	"github.com/kovalit/secrets-center/backend/internal/db"
)

// Validation errors surfaced to the handler.
var (
	ErrInvalidScope   = errors.New("scope must be 'personal' or 'commercial'")
	ErrCompanyReq     = errors.New("commercial entries require a company_id")
	ErrCompanyOnPers  = errors.New("personal entries must not have a company_id")
	ErrGroupRequired  = errors.New("group_id is required")
	ErrPasswordReq    = errors.New("password is required")
	ErrCompanyUnknown = errors.New("company not found")
	ErrGroupUnknown   = errors.New("group not found")
)

// Service implements password entry business logic.
type Service struct {
	store *db.Store
	enc   crypto.Encryptor
}

// New builds the passwords service.
func New(store *db.Store, enc crypto.Encryptor) *Service {
	return &Service{store: store, enc: enc}
}

// WriteInput is the payload for creating/updating an entry.
type WriteInput struct {
	Scope      string
	CompanyID  *string
	GroupID    string
	Title      string
	SiteURL    *string
	Login      *string
	Password   string // required on create; empty on update = keep existing
	Comment    *string
	IconSource *string // "favicon" | "group" | "custom"; nil = auto-detect
	CustomIcon *string // data URL, used when IconSource == "custom"
}

// View is the safe representation of an entry (never includes the password).
type View struct {
	ID          string    `json:"id"`
	Scope       string    `json:"scope"`
	CompanyID   *string   `json:"company_id"`
	GroupID     string    `json:"group_id"`
	Title       string    `json:"title"`
	SiteURL     *string   `json:"site_url"`
	Domain      *string   `json:"domain"`
	FaviconURL  *string   `json:"favicon_url"`
	IconSource  string    `json:"icon_source"`
	CustomIcon  *string   `json:"custom_icon"`
	Login       *string   `json:"login"`
	HasPassword bool      `json:"has_password"`
	HasComment  bool      `json:"has_comment"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toView(e *db.PasswordEntry) View {
	return View{
		ID:          e.ID,
		Scope:       e.Scope,
		CompanyID:   e.CompanyID,
		GroupID:     e.GroupID,
		Title:       e.Title,
		SiteURL:     e.SiteURL,
		Domain:      e.Domain,
		FaviconURL:  e.FaviconURL,
		IconSource:  e.IconSource,
		CustomIcon:  e.CustomIcon,
		Login:       e.Login,
		HasPassword: len(e.EncryptedPassword) > 0,
		HasComment:  len(e.EncryptedComment) > 0,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

// ListGroups returns all password groups.
func (s *Service) ListGroups(ctx context.Context) ([]db.PasswordGroup, error) {
	return s.store.ListPasswordGroups(ctx)
}

// List returns the user's entries as safe views.
func (s *Service) List(ctx context.Context, ownerID string, f db.PasswordFilter) ([]View, error) {
	entries, err := s.store.ListPasswordEntries(ctx, ownerID, f)
	if err != nil {
		return nil, err
	}
	out := make([]View, 0, len(entries))
	for i := range entries {
		out = append(out, toView(&entries[i]))
	}
	return out, nil
}

// Get returns a single entry view.
func (s *Service) Get(ctx context.Context, ownerID, id string) (*View, error) {
	e, err := s.store.GetPasswordEntry(ctx, ownerID, id)
	if err != nil {
		return nil, err
	}
	v := toView(e)
	return &v, nil
}

// Create validates, encrypts and stores a new entry.
func (s *Service) Create(ctx context.Context, ownerID string, in WriteInput) (*View, error) {
	if err := s.validate(ctx, ownerID, in, true); err != nil {
		return nil, err
	}

	entry := &db.PasswordEntry{
		OwnerUserID: ownerID,
		CompanyID:   in.CompanyID,
		GroupID:     in.GroupID,
		Scope:       in.Scope,
		Title:       strings.TrimSpace(in.Title),
		SiteURL:     trimPtr(in.SiteURL),
		Login:       in.Login,
		IconSource:  "group",
	}

	s.applyIcon(ctx, entry, in)

	encPw, nonce, ver, err := s.enc.EncryptString(in.Password)
	if err != nil {
		return nil, err
	}
	entry.EncryptedPassword, entry.PasswordNonce, entry.PasswordKeyVersion = encPw, nonce, ver

	if in.Comment != nil && *in.Comment != "" {
		if err := s.encryptComment(entry, *in.Comment); err != nil {
			return nil, err
		}
	}

	created, err := s.store.CreatePasswordEntry(ctx, entry)
	if err != nil {
		return nil, err
	}
	v := toView(created)
	return &v, nil
}

// Update validates and rewrites an entry. An empty password keeps the old one.
func (s *Service) Update(ctx context.Context, ownerID, id string, in WriteInput) (*View, error) {
	existing, err := s.store.GetPasswordEntry(ctx, ownerID, id)
	if err != nil {
		return nil, err
	}
	if err := s.validate(ctx, ownerID, in, false); err != nil {
		return nil, err
	}

	existing.CompanyID = in.CompanyID
	existing.GroupID = in.GroupID
	existing.Scope = in.Scope
	existing.Title = strings.TrimSpace(in.Title)
	existing.SiteURL = trimPtr(in.SiteURL)
	existing.Login = in.Login

	s.applyIcon(ctx, existing, in)

	if in.Password != "" {
		encPw, nonce, ver, err := s.enc.EncryptString(in.Password)
		if err != nil {
			return nil, err
		}
		existing.EncryptedPassword, existing.PasswordNonce, existing.PasswordKeyVersion = encPw, nonce, ver
	}

	if in.Comment != nil {
		if *in.Comment == "" {
			existing.EncryptedComment, existing.CommentNonce, existing.CommentKeyVersion = nil, nil, nil
		} else if err := s.encryptComment(existing, *in.Comment); err != nil {
			return nil, err
		}
	}

	updated, err := s.store.UpdatePasswordEntry(ctx, existing)
	if err != nil {
		return nil, err
	}
	v := toView(updated)
	return &v, nil
}

// Delete soft-deletes an entry.
func (s *Service) Delete(ctx context.Context, ownerID, id string) error {
	return s.store.SoftDeletePasswordEntry(ctx, ownerID, id)
}

// Reveal decrypts and returns the plaintext password. The caller is responsible
// for writing the audit entry.
func (s *Service) Reveal(ctx context.Context, ownerID, id string) (string, error) {
	e, err := s.store.GetPasswordEntry(ctx, ownerID, id)
	if err != nil {
		return "", err
	}
	return s.enc.DecryptString(e.EncryptedPassword, e.PasswordNonce, e.PasswordKeyVersion)
}

// RevealComment decrypts the optional comment.
func (s *Service) RevealComment(ctx context.Context, ownerID, id string) (string, error) {
	e, err := s.store.GetPasswordEntry(ctx, ownerID, id)
	if err != nil {
		return "", err
	}
	if len(e.EncryptedComment) == 0 || e.CommentKeyVersion == nil {
		return "", nil
	}
	return s.enc.DecryptString(e.EncryptedComment, e.CommentNonce, *e.CommentKeyVersion)
}

func (s *Service) encryptComment(e *db.PasswordEntry, comment string) error {
	enc, nonce, ver, err := s.enc.EncryptString(comment)
	if err != nil {
		return err
	}
	e.EncryptedComment, e.CommentNonce = enc, nonce
	e.CommentKeyVersion = &ver
	return nil
}

// applyIcon derives the domain from the site URL and resolves the entry icon
// according to the requested source:
//
//   - "custom": keep the supplied data URL (falls back to auto if it is invalid).
//   - "group":  use the group icon, clearing any favicon/custom icon.
//   - "favicon" or nil (auto): probe the site favicon, falling back to group.
func (s *Service) applyIcon(ctx context.Context, e *db.PasswordEntry, in WriteInput) {
	// Derive the domain first; it is needed for lookups and favicon probing.
	e.Domain = nil
	if e.SiteURL != nil && *e.SiteURL != "" {
		if domain := ExtractDomain(*e.SiteURL); domain != "" {
			e.Domain = &domain
		}
	}

	source := ""
	if in.IconSource != nil {
		source = *in.IconSource
	}

	switch source {
	case "custom":
		if in.CustomIcon != nil && validCustomIcon(*in.CustomIcon) {
			e.IconSource = "custom"
			e.CustomIcon = in.CustomIcon
			e.FaviconURL = nil
			return
		}
		// Invalid/empty custom icon: fall through to auto-detection.
	case "group":
		e.IconSource = "group"
		e.CustomIcon = nil
		e.FaviconURL = nil
		return
	}

	// Auto / "favicon": probe the site favicon, otherwise use the group icon.
	e.CustomIcon = nil
	e.FaviconURL = nil
	e.IconSource = "group"
	if e.Domain != nil {
		if faviconURL, ok := ResolveFavicon(ctx, *e.Domain); ok {
			e.FaviconURL = &faviconURL
			e.IconSource = "favicon"
		}
	}
}

func (s *Service) validate(ctx context.Context, ownerID string, in WriteInput, isCreate bool) error {
	if in.Scope != "personal" && in.Scope != "commercial" {
		return ErrInvalidScope
	}
	if in.GroupID == "" {
		return ErrGroupRequired
	}
	if _, err := s.store.GetPasswordGroup(ctx, in.GroupID); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return ErrGroupUnknown
		}
		return err
	}
	if in.Scope == "commercial" {
		if in.CompanyID == nil || *in.CompanyID == "" {
			return ErrCompanyReq
		}
		if _, err := s.store.GetCompany(ctx, ownerID, *in.CompanyID); err != nil {
			if errors.Is(err, db.ErrNotFound) {
				return ErrCompanyUnknown
			}
			return err
		}
	} else if in.CompanyID != nil && *in.CompanyID != "" {
		return ErrCompanyOnPers
	}
	if isCreate && in.Password == "" {
		return ErrPasswordReq
	}
	if strings.TrimSpace(in.Title) == "" {
		return errors.New("title is required")
	}
	return nil
}

func trimPtr(p *string) *string {
	if p == nil {
		return nil
	}
	v := strings.TrimSpace(*p)
	if v == "" {
		return nil
	}
	return &v
}

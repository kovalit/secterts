package db

import "time"

// All UUID columns are handled as strings for simplicity; pgx encodes/decodes
// them against PostgreSQL's uuid type transparently.

// User is a row of the users table (password_hash never leaves the backend).
type User struct {
	ID              string
	Email           string
	PasswordHash    string
	EmailVerified   bool
	Email2FAEnabled bool
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// EmailAuthCode is a hashed one-time code used for email 2FA / verification.
type EmailAuthCode struct {
	ID         string
	UserID     string
	CodeHash   string
	Purpose    string
	ExpiresAt  time.Time
	ConsumedAt *time.Time
	Attempts   int
	CreatedAt  time.Time
}

// Session is a web session backed by a hashed refresh token cookie.
type Session struct {
	ID               string
	UserID           string
	RefreshTokenHash string
	UserAgent        *string
	IP               *string
	ExpiresAt        time.Time
	RevokedAt        *time.Time
	CreatedAt        time.Time
}

// Company is owned by a single user in the MVP.
type Company struct {
	ID          string
	OwnerUserID string
	Name        string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// PasswordGroup is a system or user group for password entries.
type PasswordGroup struct {
	ID        string
	Slug      string
	Label     string
	Icon      string
	IsSystem  bool
	SortOrder int
	CreatedAt time.Time
}

// PasswordEntry stores an encrypted password (+ optional encrypted comment).
type PasswordEntry struct {
	ID          string
	OwnerUserID string
	CompanyID   *string
	GroupID     string

	Scope      string
	Title      string
	SiteURL    *string
	Domain     *string
	FaviconURL *string
	IconSource string
	CustomIcon *string

	// EntryType classifies the record (password, api_key, ssh_key, ...).
	EntryType string
	// ExpiresAt is an optional expiration date (API keys, certs, tokens, ...).
	ExpiresAt *time.Time
	// Owner is the free-text responsible owner; empty = "secret without owner".
	Owner *string
	// PasswordStrength is a 0..4 score computed on write (nil = unknown).
	PasswordStrength *int
	// LastUsedAt is the last time the secret was revealed (nil = never).
	LastUsedAt *time.Time

	Login              *string
	EncryptedPassword  []byte
	PasswordNonce      []byte
	PasswordKeyVersion int

	EncryptedComment  []byte
	CommentNonce      []byte
	CommentKeyVersion *int

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// AppProject groups application secrets.
type AppProject struct {
	ID          string
	OwnerUserID string
	CompanyID   *string
	Name        string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// AppEnvironment is dev/staging/prod (or custom) inside a project.
type AppEnvironment struct {
	ID        string
	ProjectID string
	Name      string
	SortOrder int
	CreatedAt time.Time
}

// AppSecret stores an encrypted key/value secret for an environment.
type AppSecret struct {
	ID            string
	ProjectID     string
	EnvironmentID string

	Key             string
	EncryptedValue  []byte
	ValueNonce      []byte
	ValueKeyVersion int

	EncryptedComment  []byte
	CommentNonce      []byte
	CommentKeyVersion *int

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// SecretNote stores an encrypted free-form secret (title + type + text).
type SecretNote struct {
	ID          string
	OwnerUserID string

	Title string
	Type  string

	EncryptedText  []byte
	TextNonce      []byte
	TextKeyVersion int

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// ExtensionToken is a hashed bearer token used by the Chrome extension.
type ExtensionToken struct {
	ID         string
	UserID     string
	TokenHash  string
	Name       string
	Scopes     []string
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	CreatedAt  time.Time
	LastUsedAt *time.Time
}

// AuditLog records security-relevant actions.
type AuditLog struct {
	ID         string
	UserID     *string
	CompanyID  *string
	Action     string
	EntityType string
	EntityID   *string
	IP         *string
	UserAgent  *string
	Metadata   []byte
	CreatedAt  time.Time
}

// BackupExport tracks backup runs.
type BackupExport struct {
	ID            string
	CreatedBy     *string
	Status        string
	FileName      *string
	StorageTarget *string
	SHA256        *string
	ErrorMessage  *string
	CreatedAt     time.Time
	FinishedAt    *time.Time
}

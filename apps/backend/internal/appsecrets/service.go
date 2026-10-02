// Package appsecrets implements projects, environments and encrypted key/value
// application secrets.
package appsecrets

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/kovalit/secrets-center/backend/internal/crypto"
	"github.com/kovalit/secrets-center/backend/internal/db"
)

// defaultEnvironments are seeded for every new project.
var defaultEnvironments = []struct {
	Name  string
	Order int
}{
	{"dev", 10},
	{"staging", 20},
	{"prod", 30},
}

// Errors surfaced to the handler.
var (
	ErrForbidden  = errors.New("forbidden")
	ErrKeyReq     = errors.New("key is required")
	ErrValueReq   = errors.New("value is required")
	ErrNameReq    = errors.New("name is required")
	ErrEnvUnknown = errors.New("environment not found")
)

// Service implements app-secrets logic.
type Service struct {
	store *db.Store
	enc   crypto.Encryptor
}

// New builds the service.
func New(store *db.Store, enc crypto.Encryptor) *Service {
	return &Service{store: store, enc: enc}
}

// --- responses ---

// ProjectView includes environment/secret counts.
type ProjectView struct {
	ID          string    `json:"id"`
	CompanyID   *string   `json:"company_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	EnvCount    int       `json:"env_count"`
	SecretCount int       `json:"secret_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// EnvironmentView is the environment representation.
type EnvironmentView struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

// SecretView never includes the value.
type SecretView struct {
	ID            string    `json:"id"`
	ProjectID     string    `json:"project_id"`
	EnvironmentID string    `json:"environment_id"`
	Key           string    `json:"key"`
	HasValue      bool      `json:"has_value"`
	HasComment    bool      `json:"has_comment"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// --- projects ---

// ListProjects returns projects with counts.
func (s *Service) ListProjects(ctx context.Context, ownerID string) ([]ProjectView, error) {
	projects, err := s.store.ListAppProjects(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	out := make([]ProjectView, 0, len(projects))
	for i := range projects {
		p := &projects[i]
		envCount, secretCount, _ := s.store.ProjectStats(ctx, p.ID)
		out = append(out, ProjectView{
			ID: p.ID, CompanyID: p.CompanyID, Name: p.Name, Description: p.Description,
			EnvCount: envCount, SecretCount: secretCount, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
		})
	}
	return out, nil
}

// GetProject returns one project with counts.
func (s *Service) GetProject(ctx context.Context, ownerID, id string) (*ProjectView, error) {
	p, err := s.store.GetAppProject(ctx, ownerID, id)
	if err != nil {
		return nil, err
	}
	envCount, secretCount, _ := s.store.ProjectStats(ctx, p.ID)
	return &ProjectView{
		ID: p.ID, CompanyID: p.CompanyID, Name: p.Name, Description: p.Description,
		EnvCount: envCount, SecretCount: secretCount, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}, nil
}

// CreateProject inserts a project and seeds dev/staging/prod environments.
func (s *Service) CreateProject(ctx context.Context, ownerID string, companyID *string, name string, description *string) (*ProjectView, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrNameReq
	}
	if companyID != nil && *companyID != "" {
		if _, err := s.store.GetCompany(ctx, ownerID, *companyID); err != nil {
			return nil, ErrForbidden
		}
	} else {
		companyID = nil
	}

	p, err := s.store.CreateAppProject(ctx, ownerID, companyID, name, trimPtr(description))
	if err != nil {
		return nil, err
	}
	for _, e := range defaultEnvironments {
		if _, err := s.store.CreateEnvironment(ctx, p.ID, e.Name, e.Order); err != nil {
			return nil, err
		}
	}
	return s.GetProject(ctx, ownerID, p.ID)
}

// UpdateProject updates a project.
func (s *Service) UpdateProject(ctx context.Context, ownerID, id string, companyID *string, name string, description *string) (*ProjectView, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrNameReq
	}
	if companyID != nil && *companyID != "" {
		if _, err := s.store.GetCompany(ctx, ownerID, *companyID); err != nil {
			return nil, ErrForbidden
		}
	} else {
		companyID = nil
	}
	if _, err := s.store.UpdateAppProject(ctx, ownerID, id, companyID, name, trimPtr(description)); err != nil {
		return nil, err
	}
	return s.GetProject(ctx, ownerID, id)
}

// DeleteProject removes a project.
func (s *Service) DeleteProject(ctx context.Context, ownerID, id string) error {
	return s.store.DeleteAppProject(ctx, ownerID, id)
}

// --- environments ---

// ListEnvironments lists a project's environments after verifying ownership.
func (s *Service) ListEnvironments(ctx context.Context, ownerID, projectID string) ([]EnvironmentView, error) {
	if _, err := s.store.GetAppProject(ctx, ownerID, projectID); err != nil {
		return nil, err
	}
	envs, err := s.store.ListEnvironments(ctx, projectID)
	if err != nil {
		return nil, err
	}
	out := make([]EnvironmentView, 0, len(envs))
	for _, e := range envs {
		out = append(out, EnvironmentView{ID: e.ID, ProjectID: e.ProjectID, Name: e.Name, SortOrder: e.SortOrder})
	}
	return out, nil
}

// CreateEnvironment adds an environment to a project.
func (s *Service) CreateEnvironment(ctx context.Context, ownerID, projectID, name string, sortOrder int) (*EnvironmentView, error) {
	if _, err := s.store.GetAppProject(ctx, ownerID, projectID); err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrNameReq
	}
	e, err := s.store.CreateEnvironment(ctx, projectID, name, sortOrder)
	if err != nil {
		return nil, err
	}
	return &EnvironmentView{ID: e.ID, ProjectID: e.ProjectID, Name: e.Name, SortOrder: e.SortOrder}, nil
}

// UpdateEnvironment renames an environment after verifying ownership.
func (s *Service) UpdateEnvironment(ctx context.Context, ownerID, envID, name string, sortOrder int) (*EnvironmentView, error) {
	if err := s.assertEnvOwner(ctx, ownerID, envID); err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrNameReq
	}
	e, err := s.store.UpdateEnvironment(ctx, envID, name, sortOrder)
	if err != nil {
		return nil, err
	}
	return &EnvironmentView{ID: e.ID, ProjectID: e.ProjectID, Name: e.Name, SortOrder: e.SortOrder}, nil
}

// DeleteEnvironment removes an environment after verifying ownership.
func (s *Service) DeleteEnvironment(ctx context.Context, ownerID, envID string) error {
	if err := s.assertEnvOwner(ctx, ownerID, envID); err != nil {
		return err
	}
	return s.store.DeleteEnvironment(ctx, envID)
}

// --- secrets ---

// ListSecrets returns secrets for a project + environment (by id or name).
func (s *Service) ListSecrets(ctx context.Context, ownerID, projectID, envParam string) ([]SecretView, error) {
	if _, err := s.store.GetAppProject(ctx, ownerID, projectID); err != nil {
		return nil, err
	}
	env, err := s.resolveEnv(ctx, projectID, envParam)
	if err != nil {
		return nil, err
	}
	secrets, err := s.store.ListSecrets(ctx, projectID, env.ID)
	if err != nil {
		return nil, err
	}
	out := make([]SecretView, 0, len(secrets))
	for i := range secrets {
		out = append(out, toSecretView(&secrets[i]))
	}
	return out, nil
}

// SecretWriteInput is the payload for creating/updating a secret.
type SecretWriteInput struct {
	EnvironmentID string
	Key           string
	Value         string // required on create; empty on update = keep
	Comment       *string
}

// CreateSecret encrypts and stores a new secret.
func (s *Service) CreateSecret(ctx context.Context, ownerID, projectID string, in SecretWriteInput) (*SecretView, error) {
	if _, err := s.store.GetAppProject(ctx, ownerID, projectID); err != nil {
		return nil, err
	}
	env, err := s.store.GetEnvironment(ctx, projectID, in.EnvironmentID)
	if err != nil {
		return nil, ErrEnvUnknown
	}
	if strings.TrimSpace(in.Key) == "" {
		return nil, ErrKeyReq
	}
	if in.Value == "" {
		return nil, ErrValueReq
	}

	secret := &db.AppSecret{ProjectID: projectID, EnvironmentID: env.ID, Key: strings.TrimSpace(in.Key)}
	encVal, nonce, ver, err := s.enc.EncryptString(in.Value)
	if err != nil {
		return nil, err
	}
	secret.EncryptedValue, secret.ValueNonce, secret.ValueKeyVersion = encVal, nonce, ver
	if in.Comment != nil && *in.Comment != "" {
		if err := s.encryptComment(secret, *in.Comment); err != nil {
			return nil, err
		}
	}
	created, err := s.store.CreateSecret(ctx, secret)
	if err != nil {
		return nil, err
	}
	v := toSecretView(created)
	return &v, nil
}

// BulkSecretItem is one key/value pair in a bulk import.
type BulkSecretItem struct {
	Key     string
	Value   string
	Comment *string
}

// BulkUpsertSecrets creates or overwrites many secrets in one environment at
// once. It resolves the environment by id or name, skips entries with a blank
// key, and returns the resulting views. Existing keys (even previously deleted
// ones) are overwritten with the new value.
func (s *Service) BulkUpsertSecrets(ctx context.Context, ownerID, projectID, envParam string, items []BulkSecretItem) ([]SecretView, error) {
	if _, err := s.store.GetAppProject(ctx, ownerID, projectID); err != nil {
		return nil, err
	}
	env, err := s.resolveEnv(ctx, projectID, envParam)
	if err != nil {
		return nil, err
	}

	// Validate up front so a bad entry does not leave a partial import.
	type prepared struct {
		key     string
		value   string
		comment *string
	}
	var ready []prepared
	for _, it := range items {
		key := strings.TrimSpace(it.Key)
		if key == "" {
			continue
		}
		if it.Value == "" {
			return nil, ErrValueReq
		}
		ready = append(ready, prepared{key: key, value: it.Value, comment: it.Comment})
	}
	if len(ready) == 0 {
		return nil, ErrKeyReq
	}

	out := make([]SecretView, 0, len(ready))
	for _, p := range ready {
		secret := &db.AppSecret{ProjectID: projectID, EnvironmentID: env.ID, Key: p.key}
		encVal, nonce, ver, err := s.enc.EncryptString(p.value)
		if err != nil {
			return nil, err
		}
		secret.EncryptedValue, secret.ValueNonce, secret.ValueKeyVersion = encVal, nonce, ver

		setComment := p.comment != nil
		if setComment && *p.comment != "" {
			if err := s.encryptComment(secret, *p.comment); err != nil {
				return nil, err
			}
		}
		created, err := s.store.UpsertSecret(ctx, secret, setComment)
		if err != nil {
			return nil, err
		}
		out = append(out, toSecretView(created))
	}
	return out, nil
}

// UpdateSecret rewrites a secret after verifying ownership.
func (s *Service) UpdateSecret(ctx context.Context, ownerID, secretID string, in SecretWriteInput) (*SecretView, error) {
	if err := s.assertSecretOwner(ctx, ownerID, secretID); err != nil {
		return nil, err
	}
	existing, err := s.store.GetSecret(ctx, secretID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Key) != "" {
		existing.Key = strings.TrimSpace(in.Key)
	}
	if in.Value != "" {
		encVal, nonce, ver, err := s.enc.EncryptString(in.Value)
		if err != nil {
			return nil, err
		}
		existing.EncryptedValue, existing.ValueNonce, existing.ValueKeyVersion = encVal, nonce, ver
	}
	if in.Comment != nil {
		if *in.Comment == "" {
			existing.EncryptedComment, existing.CommentNonce, existing.CommentKeyVersion = nil, nil, nil
		} else if err := s.encryptComment(existing, *in.Comment); err != nil {
			return nil, err
		}
	}
	updated, err := s.store.UpdateSecret(ctx, existing)
	if err != nil {
		return nil, err
	}
	v := toSecretView(updated)
	return &v, nil
}

// DeleteSecret soft-deletes a secret after verifying ownership.
func (s *Service) DeleteSecret(ctx context.Context, ownerID, secretID string) error {
	if err := s.assertSecretOwner(ctx, ownerID, secretID); err != nil {
		return err
	}
	return s.store.SoftDeleteSecret(ctx, secretID)
}

// RevealSecret decrypts and returns the plaintext value.
func (s *Service) RevealSecret(ctx context.Context, ownerID, secretID string) (string, error) {
	if err := s.assertSecretOwner(ctx, ownerID, secretID); err != nil {
		return "", err
	}
	secret, err := s.store.GetSecret(ctx, secretID)
	if err != nil {
		return "", err
	}
	return s.enc.DecryptString(secret.EncryptedValue, secret.ValueNonce, secret.ValueKeyVersion)
}

// --- helpers ---

func (s *Service) resolveEnv(ctx context.Context, projectID, envParam string) (*db.AppEnvironment, error) {
	if envParam == "" {
		// default to the first environment (dev) if none specified
		envs, err := s.store.ListEnvironments(ctx, projectID)
		if err != nil {
			return nil, err
		}
		if len(envs) == 0 {
			return nil, ErrEnvUnknown
		}
		return &envs[0], nil
	}
	// Try by name first (dev/staging/prod), then by id.
	if env, err := s.store.GetEnvironmentByName(ctx, projectID, envParam); err == nil {
		return env, nil
	}
	env, err := s.store.GetEnvironment(ctx, projectID, envParam)
	if err != nil {
		return nil, ErrEnvUnknown
	}
	return env, nil
}

func (s *Service) assertEnvOwner(ctx context.Context, ownerID, envID string) error {
	owner, _, err := s.store.EnvironmentOwner(ctx, envID)
	if err != nil {
		return err
	}
	if owner != ownerID {
		return ErrForbidden
	}
	return nil
}

func (s *Service) assertSecretOwner(ctx context.Context, ownerID, secretID string) error {
	owner, _, err := s.store.SecretOwner(ctx, secretID)
	if err != nil {
		return err
	}
	if owner != ownerID {
		return ErrForbidden
	}
	return nil
}

func (s *Service) encryptComment(a *db.AppSecret, comment string) error {
	enc, nonce, ver, err := s.enc.EncryptString(comment)
	if err != nil {
		return err
	}
	a.EncryptedComment, a.CommentNonce = enc, nonce
	a.CommentKeyVersion = &ver
	return nil
}

func toSecretView(a *db.AppSecret) SecretView {
	return SecretView{
		ID: a.ID, ProjectID: a.ProjectID, EnvironmentID: a.EnvironmentID, Key: a.Key,
		HasValue: len(a.EncryptedValue) > 0, HasComment: len(a.EncryptedComment) > 0,
		CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt,
	}
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

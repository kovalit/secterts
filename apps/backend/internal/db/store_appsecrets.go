package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

// --- projects ---

// ListAppProjects returns all projects owned by a user.
func (s *Store) ListAppProjects(ctx context.Context, ownerUserID string) ([]AppProject, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, owner_user_id, company_id, name, description, created_at, updated_at
		FROM app_projects WHERE owner_user_id = $1 ORDER BY name`, ownerUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []AppProject
	for rows.Next() {
		var p AppProject
		if err := rows.Scan(&p.ID, &p.OwnerUserID, &p.CompanyID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// GetAppProject returns a project owned by the user.
func (s *Store) GetAppProject(ctx context.Context, ownerUserID, id string) (*AppProject, error) {
	var p AppProject
	err := s.pool.QueryRow(ctx, `
		SELECT id, owner_user_id, company_id, name, description, created_at, updated_at
		FROM app_projects WHERE id = $1 AND owner_user_id = $2`, id, ownerUserID,
	).Scan(&p.ID, &p.OwnerUserID, &p.CompanyID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// CreateAppProject inserts a project.
func (s *Store) CreateAppProject(ctx context.Context, ownerUserID string, companyID *string, name string, description *string) (*AppProject, error) {
	var p AppProject
	err := s.pool.QueryRow(ctx, `
		INSERT INTO app_projects (owner_user_id, company_id, name, description)
		VALUES ($1,$2,$3,$4)
		RETURNING id, owner_user_id, company_id, name, description, created_at, updated_at`,
		ownerUserID, companyID, name, description,
	).Scan(&p.ID, &p.OwnerUserID, &p.CompanyID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// UpdateAppProject updates a project's name/description/company.
func (s *Store) UpdateAppProject(ctx context.Context, ownerUserID, id string, companyID *string, name string, description *string) (*AppProject, error) {
	var p AppProject
	err := s.pool.QueryRow(ctx, `
		UPDATE app_projects SET company_id=$3, name=$4, description=$5, updated_at=now()
		WHERE id=$1 AND owner_user_id=$2
		RETURNING id, owner_user_id, company_id, name, description, created_at, updated_at`,
		id, ownerUserID, companyID, name, description,
	).Scan(&p.ID, &p.OwnerUserID, &p.CompanyID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// DeleteAppProject removes a project (cascades to environments and secrets).
func (s *Store) DeleteAppProject(ctx context.Context, ownerUserID, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM app_projects WHERE id = $1 AND owner_user_id = $2`, id, ownerUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ProjectStats returns environment and secret counts for a project.
func (s *Store) ProjectStats(ctx context.Context, projectID string) (envCount, secretCount int, err error) {
	err = s.pool.QueryRow(ctx, `SELECT
		(SELECT count(*) FROM app_environments WHERE project_id = $1),
		(SELECT count(*) FROM app_secrets WHERE project_id = $1 AND deleted_at IS NULL)`, projectID,
	).Scan(&envCount, &secretCount)
	return
}

// EnvironmentOwner returns the owner user id and project id for an environment.
func (s *Store) EnvironmentOwner(ctx context.Context, envID string) (ownerUserID, projectID string, err error) {
	err = s.pool.QueryRow(ctx, `
		SELECT p.owner_user_id, p.id
		FROM app_environments e JOIN app_projects p ON p.id = e.project_id
		WHERE e.id = $1`, envID).Scan(&ownerUserID, &projectID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", ErrNotFound
	}
	return ownerUserID, projectID, err
}

// SecretOwner returns the owner user id and project id for a secret.
func (s *Store) SecretOwner(ctx context.Context, secretID string) (ownerUserID, projectID string, err error) {
	err = s.pool.QueryRow(ctx, `
		SELECT p.owner_user_id, p.id
		FROM app_secrets sec JOIN app_projects p ON p.id = sec.project_id
		WHERE sec.id = $1 AND sec.deleted_at IS NULL`, secretID).Scan(&ownerUserID, &projectID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", ErrNotFound
	}
	return ownerUserID, projectID, err
}

// --- environments ---

// ListEnvironments returns all environments of a project.
func (s *Store) ListEnvironments(ctx context.Context, projectID string) ([]AppEnvironment, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, project_id, name, sort_order, created_at
		FROM app_environments WHERE project_id = $1 ORDER BY sort_order, name`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []AppEnvironment
	for rows.Next() {
		var e AppEnvironment
		if err := rows.Scan(&e.ID, &e.ProjectID, &e.Name, &e.SortOrder, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// GetEnvironment returns a single environment.
func (s *Store) GetEnvironment(ctx context.Context, projectID, id string) (*AppEnvironment, error) {
	var e AppEnvironment
	err := s.pool.QueryRow(ctx, `
		SELECT id, project_id, name, sort_order, created_at
		FROM app_environments WHERE id = $1 AND project_id = $2`, id, projectID,
	).Scan(&e.ID, &e.ProjectID, &e.Name, &e.SortOrder, &e.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// GetEnvironmentByName resolves an environment by its name within a project.
func (s *Store) GetEnvironmentByName(ctx context.Context, projectID, name string) (*AppEnvironment, error) {
	var e AppEnvironment
	err := s.pool.QueryRow(ctx, `
		SELECT id, project_id, name, sort_order, created_at
		FROM app_environments WHERE project_id = $1 AND name = $2`, projectID, name,
	).Scan(&e.ID, &e.ProjectID, &e.Name, &e.SortOrder, &e.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// CreateEnvironment inserts an environment.
func (s *Store) CreateEnvironment(ctx context.Context, projectID, name string, sortOrder int) (*AppEnvironment, error) {
	var e AppEnvironment
	err := s.pool.QueryRow(ctx, `
		INSERT INTO app_environments (project_id, name, sort_order)
		VALUES ($1,$2,$3)
		RETURNING id, project_id, name, sort_order, created_at`, projectID, name, sortOrder,
	).Scan(&e.ID, &e.ProjectID, &e.Name, &e.SortOrder, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// UpdateEnvironment renames an environment.
func (s *Store) UpdateEnvironment(ctx context.Context, id, name string, sortOrder int) (*AppEnvironment, error) {
	var e AppEnvironment
	err := s.pool.QueryRow(ctx, `
		UPDATE app_environments SET name=$2, sort_order=$3 WHERE id=$1
		RETURNING id, project_id, name, sort_order, created_at`, id, name, sortOrder,
	).Scan(&e.ID, &e.ProjectID, &e.Name, &e.SortOrder, &e.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// DeleteEnvironment removes an environment (cascades to its secrets).
func (s *Store) DeleteEnvironment(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM app_environments WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// --- secrets ---

const appSecretColumns = `
	id, project_id, environment_id, key, encrypted_value, value_nonce, value_key_version,
	encrypted_comment, comment_nonce, comment_key_version, created_at, updated_at, deleted_at`

func scanAppSecret(row pgx.Row) (*AppSecret, error) {
	var a AppSecret
	err := row.Scan(&a.ID, &a.ProjectID, &a.EnvironmentID, &a.Key, &a.EncryptedValue, &a.ValueNonce, &a.ValueKeyVersion,
		&a.EncryptedComment, &a.CommentNonce, &a.CommentKeyVersion, &a.CreatedAt, &a.UpdatedAt, &a.DeletedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// ListSecrets returns non-deleted secrets for a project/environment.
func (s *Store) ListSecrets(ctx context.Context, projectID, environmentID string) ([]AppSecret, error) {
	rows, err := s.pool.Query(ctx, "SELECT "+appSecretColumns+
		" FROM app_secrets WHERE project_id = $1 AND environment_id = $2 AND deleted_at IS NULL ORDER BY key", projectID, environmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []AppSecret
	for rows.Next() {
		a, err := scanAppSecret(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *a)
	}
	return out, rows.Err()
}

// GetSecret returns a single non-deleted secret by id.
func (s *Store) GetSecret(ctx context.Context, id string) (*AppSecret, error) {
	row := s.pool.QueryRow(ctx, "SELECT "+appSecretColumns+" FROM app_secrets WHERE id = $1 AND deleted_at IS NULL", id)
	a, err := scanAppSecret(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return a, err
}

// CreateSecret inserts an encrypted secret.
func (s *Store) CreateSecret(ctx context.Context, a *AppSecret) (*AppSecret, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO app_secrets (project_id, environment_id, key, encrypted_value, value_nonce, value_key_version,
			encrypted_comment, comment_nonce, comment_key_version)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING `+appSecretColumns,
		a.ProjectID, a.EnvironmentID, a.Key, a.EncryptedValue, a.ValueNonce, a.ValueKeyVersion,
		a.EncryptedComment, a.CommentNonce, a.CommentKeyVersion)
	return scanAppSecret(row)
}

// UpdateSecret rewrites an existing secret.
func (s *Store) UpdateSecret(ctx context.Context, a *AppSecret) (*AppSecret, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE app_secrets SET key=$2, encrypted_value=$3, value_nonce=$4, value_key_version=$5,
			encrypted_comment=$6, comment_nonce=$7, comment_key_version=$8, updated_at=now()
		WHERE id=$1 AND deleted_at IS NULL
		RETURNING `+appSecretColumns,
		a.ID, a.Key, a.EncryptedValue, a.ValueNonce, a.ValueKeyVersion,
		a.EncryptedComment, a.CommentNonce, a.CommentKeyVersion)
	res, err := scanAppSecret(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return res, err
}

// SoftDeleteSecret marks a secret deleted.
func (s *Store) SoftDeleteSecret(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `UPDATE app_secrets SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

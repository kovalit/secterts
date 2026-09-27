package db

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

// ListPasswordGroups returns all groups ordered by sort_order.
func (s *Store) ListPasswordGroups(ctx context.Context) ([]PasswordGroup, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, slug, label, icon, is_system, sort_order, created_at
		FROM password_groups ORDER BY sort_order, label`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []PasswordGroup
	for rows.Next() {
		var g PasswordGroup
		if err := rows.Scan(&g.ID, &g.Slug, &g.Label, &g.Icon, &g.IsSystem, &g.SortOrder, &g.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

// GetPasswordGroup returns a single group by id.
func (s *Store) GetPasswordGroup(ctx context.Context, id string) (*PasswordGroup, error) {
	var g PasswordGroup
	err := s.pool.QueryRow(ctx, `
		SELECT id, slug, label, icon, is_system, sort_order, created_at
		FROM password_groups WHERE id = $1`, id,
	).Scan(&g.ID, &g.Slug, &g.Label, &g.Icon, &g.IsSystem, &g.SortOrder, &g.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &g, nil
}

// PasswordFilter holds optional list filters.
type PasswordFilter struct {
	Scope     string
	CompanyID string
	GroupID   string
	Query     string
	Domain    string
}

const passwordColumns = `
	id, owner_user_id, company_id, group_id, scope, title, site_url, domain, favicon_url, icon_source,
	login, encrypted_password, password_nonce, password_key_version,
	encrypted_comment, comment_nonce, comment_key_version,
	created_at, updated_at, deleted_at`

func scanPasswordEntry(row pgx.Row) (*PasswordEntry, error) {
	var e PasswordEntry
	err := row.Scan(
		&e.ID, &e.OwnerUserID, &e.CompanyID, &e.GroupID, &e.Scope, &e.Title, &e.SiteURL, &e.Domain, &e.FaviconURL, &e.IconSource,
		&e.Login, &e.EncryptedPassword, &e.PasswordNonce, &e.PasswordKeyVersion,
		&e.EncryptedComment, &e.CommentNonce, &e.CommentKeyVersion,
		&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// ListPasswordEntries returns non-deleted entries for a user matching filters.
func (s *Store) ListPasswordEntries(ctx context.Context, ownerUserID string, f PasswordFilter) ([]PasswordEntry, error) {
	var conds []string
	var args []any
	args = append(args, ownerUserID)
	conds = append(conds, "owner_user_id = $1", "deleted_at IS NULL")

	add := func(cond string, val any) {
		args = append(args, val)
		conds = append(conds, fmt.Sprintf(cond, len(args)))
	}

	if f.Scope != "" {
		add("scope = $%d", f.Scope)
	}
	if f.CompanyID != "" {
		add("company_id = $%d", f.CompanyID)
	}
	if f.GroupID != "" {
		add("group_id = $%d", f.GroupID)
	}
	if f.Domain != "" {
		add("domain = $%d", f.Domain)
	}
	if f.Query != "" {
		args = append(args, "%"+strings.ToLower(f.Query)+"%")
		idx := len(args)
		conds = append(conds, fmt.Sprintf("(lower(title) LIKE $%d OR lower(coalesce(login,'')) LIKE $%d OR lower(coalesce(domain,'')) LIKE $%d)", idx, idx, idx))
	}

	query := "SELECT " + passwordColumns + " FROM password_entries WHERE " +
		strings.Join(conds, " AND ") + " ORDER BY updated_at DESC"

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []PasswordEntry
	for rows.Next() {
		e, err := scanPasswordEntry(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *e)
	}
	return out, rows.Err()
}

// GetPasswordEntry returns a single non-deleted entry owned by the user.
func (s *Store) GetPasswordEntry(ctx context.Context, ownerUserID, id string) (*PasswordEntry, error) {
	row := s.pool.QueryRow(ctx, "SELECT "+passwordColumns+
		" FROM password_entries WHERE id = $1 AND owner_user_id = $2 AND deleted_at IS NULL", id, ownerUserID)
	e, err := scanPasswordEntry(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return e, err
}

// LookupByDomain returns entries for a user matching an exact domain (extension).
func (s *Store) LookupByDomain(ctx context.Context, ownerUserID, domain string) ([]PasswordEntry, error) {
	rows, err := s.pool.Query(ctx, "SELECT "+passwordColumns+
		" FROM password_entries WHERE owner_user_id = $1 AND domain = $2 AND deleted_at IS NULL ORDER BY title", ownerUserID, domain)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []PasswordEntry
	for rows.Next() {
		e, err := scanPasswordEntry(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *e)
	}
	return out, rows.Err()
}

// CreatePasswordEntry inserts a new encrypted entry.
func (s *Store) CreatePasswordEntry(ctx context.Context, e *PasswordEntry) (*PasswordEntry, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO password_entries (
			owner_user_id, company_id, group_id, scope, title, site_url, domain, favicon_url, icon_source,
			login, encrypted_password, password_nonce, password_key_version,
			encrypted_comment, comment_nonce, comment_key_version)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		RETURNING `+passwordColumns,
		e.OwnerUserID, e.CompanyID, e.GroupID, e.Scope, e.Title, e.SiteURL, e.Domain, e.FaviconURL, e.IconSource,
		e.Login, e.EncryptedPassword, e.PasswordNonce, e.PasswordKeyVersion,
		e.EncryptedComment, e.CommentNonce, e.CommentKeyVersion)
	return scanPasswordEntry(row)
}

// UpdatePasswordEntry rewrites an existing entry (all encrypted fields included).
func (s *Store) UpdatePasswordEntry(ctx context.Context, e *PasswordEntry) (*PasswordEntry, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE password_entries SET
			company_id=$3, group_id=$4, scope=$5, title=$6, site_url=$7, domain=$8, favicon_url=$9, icon_source=$10,
			login=$11, encrypted_password=$12, password_nonce=$13, password_key_version=$14,
			encrypted_comment=$15, comment_nonce=$16, comment_key_version=$17, updated_at=now()
		WHERE id=$1 AND owner_user_id=$2 AND deleted_at IS NULL
		RETURNING `+passwordColumns,
		e.ID, e.OwnerUserID, e.CompanyID, e.GroupID, e.Scope, e.Title, e.SiteURL, e.Domain, e.FaviconURL, e.IconSource,
		e.Login, e.EncryptedPassword, e.PasswordNonce, e.PasswordKeyVersion,
		e.EncryptedComment, e.CommentNonce, e.CommentKeyVersion)
	res, err := scanPasswordEntry(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return res, err
}

// SoftDeletePasswordEntry marks an entry as deleted.
func (s *Store) SoftDeletePasswordEntry(ctx context.Context, ownerUserID, id string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE password_entries SET deleted_at = now()
		WHERE id = $1 AND owner_user_id = $2 AND deleted_at IS NULL`, id, ownerUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

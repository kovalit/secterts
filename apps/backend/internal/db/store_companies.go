package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

// ListCompanies returns all companies owned by a user.
func (s *Store) ListCompanies(ctx context.Context, ownerUserID string) ([]Company, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, owner_user_id, name, created_at, updated_at
		FROM companies WHERE owner_user_id = $1 ORDER BY name`, ownerUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Company
	for rows.Next() {
		var c Company
		if err := rows.Scan(&c.ID, &c.OwnerUserID, &c.Name, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// GetCompany returns a company if it belongs to the user.
func (s *Store) GetCompany(ctx context.Context, ownerUserID, id string) (*Company, error) {
	var c Company
	err := s.pool.QueryRow(ctx, `
		SELECT id, owner_user_id, name, created_at, updated_at
		FROM companies WHERE id = $1 AND owner_user_id = $2`, id, ownerUserID,
	).Scan(&c.ID, &c.OwnerUserID, &c.Name, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// CreateCompany inserts a new company for the owner.
func (s *Store) CreateCompany(ctx context.Context, ownerUserID, name string) (*Company, error) {
	var c Company
	err := s.pool.QueryRow(ctx, `
		INSERT INTO companies (owner_user_id, name)
		VALUES ($1, $2)
		RETURNING id, owner_user_id, name, created_at, updated_at`, ownerUserID, name,
	).Scan(&c.ID, &c.OwnerUserID, &c.Name, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// UpdateCompany renames a company owned by the user.
func (s *Store) UpdateCompany(ctx context.Context, ownerUserID, id, name string) (*Company, error) {
	var c Company
	err := s.pool.QueryRow(ctx, `
		UPDATE companies SET name = $3, updated_at = now()
		WHERE id = $1 AND owner_user_id = $2
		RETURNING id, owner_user_id, name, created_at, updated_at`, id, ownerUserID, name,
	).Scan(&c.ID, &c.OwnerUserID, &c.Name, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// DeleteCompany removes a company owned by the user.
func (s *Store) DeleteCompany(ctx context.Context, ownerUserID, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM companies WHERE id = $1 AND owner_user_id = $2`, id, ownerUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

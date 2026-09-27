package db

import (
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound is returned by store methods when a row does not exist.
var ErrNotFound = errors.New("not found")

// Store is the data-access layer over the pgx pool. Query files are grouped by
// domain (users, companies, passwords, ...) but all share this one type.
type Store struct {
	pool *pgxpool.Pool
}

// NewStore wraps a pool.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// Pool exposes the underlying pool for backup/export style bulk reads.
func (s *Store) Pool() *pgxpool.Pool { return s.pool }

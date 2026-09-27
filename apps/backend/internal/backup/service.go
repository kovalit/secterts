package backup

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kovalit/secrets-center/backend/internal/db"
)

// Service builds and records logical backup exports.
type Service struct {
	store *db.Store
}

// New builds the service.
func New(store *db.Store) *Service {
	return &Service{store: store}
}

// Result carries the produced export artifact.
type Result struct {
	ExportID string
	FileName string
	SHA256   string
	Data     []byte
}

// Export builds a JSON export, records the run and returns the bytes.
func (s *Service) Export(ctx context.Context, createdBy *string) (*Result, error) {
	run, err := s.store.CreateBackupExport(ctx, createdBy)
	if err != nil {
		return nil, err
	}

	doc, err := build(ctx, s.store.Pool())
	if err != nil {
		msg := err.Error()
		_ = s.store.FinishBackupExport(ctx, run.ID, "failed", nil, nil, nil, &msg)
		return nil, err
	}

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		msg := err.Error()
		_ = s.store.FinishBackupExport(ctx, run.ID, "failed", nil, nil, nil, &msg)
		return nil, err
	}

	sum := sha256.Sum256(data)
	shaHex := hex.EncodeToString(sum[:])
	fileName := fmt.Sprintf("secrets-center-export-%s.json", time.Now().UTC().Format("20060102T150405Z"))
	target := "download"

	if err := s.store.FinishBackupExport(ctx, run.ID, "success", &fileName, &target, &shaHex, nil); err != nil {
		return nil, err
	}

	return &Result{ExportID: run.ID, FileName: fileName, SHA256: shaHex, Data: data}, nil
}

// List returns recent backup runs.
func (s *Service) List(ctx context.Context, limit int) ([]db.BackupExport, error) {
	return s.store.ListBackupExports(ctx, limit)
}

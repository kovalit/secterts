// Package audit records security-relevant actions (reveal/copy/update/delete,
// login events) into the audit_logs table.
package audit

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/kovalit/secrets-center/backend/internal/db"
	"github.com/kovalit/secrets-center/backend/internal/httpx"
)

// Service writes audit entries.
type Service struct {
	store *db.Store
}

// New builds an audit service.
func New(store *db.Store) *Service {
	return &Service{store: store}
}

// Entry describes a single auditable action.
type Entry struct {
	UserID     *string
	CompanyID  *string
	Action     string
	EntityType string
	EntityID   *string
	IP         *string
	UserAgent  *string
	Metadata   map[string]any
}

// Record writes an audit entry. Failures are logged but never block the request.
func (s *Service) Record(ctx context.Context, e Entry) {
	var meta []byte
	if e.Metadata != nil {
		if b, err := json.Marshal(e.Metadata); err == nil {
			meta = b
		}
	}
	err := s.store.InsertAuditLog(ctx, &db.AuditLog{
		UserID:     e.UserID,
		CompanyID:  e.CompanyID,
		Action:     e.Action,
		EntityType: e.EntityType,
		EntityID:   e.EntityID,
		IP:         e.IP,
		UserAgent:  e.UserAgent,
		Metadata:   meta,
	})
	if err != nil {
		log.Printf("audit: failed to record %q: %v", e.Action, err)
	}
}

// FromRequest extracts client IP and user-agent for an audit entry.
func FromRequest(r *http.Request) (ip *string, ua *string) {
	if v := httpx.ClientIP(r); v != nil {
		s := v.String()
		ip = &s
	}
	if v := r.UserAgent(); v != "" {
		ua = &v
	}
	return ip, ua
}

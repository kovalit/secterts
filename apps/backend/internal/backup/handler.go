package backup

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/kovalit/secrets-center/backend/internal/audit"
	"github.com/kovalit/secrets-center/backend/internal/auth"
	"github.com/kovalit/secrets-center/backend/internal/httpx"
)

// Handler exposes the backup HTTP API (behind a web session).
type Handler struct {
	svc   *Service
	audit *audit.Service
}

// NewHandler builds the handler.
func NewHandler(svc *Service, auditSvc *audit.Service) *Handler {
	return &Handler{svc: svc, audit: auditSvc}
}

// Routes returns the /backups subtree.
func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Post("/export", h.export)
	r.Get("/", h.list)
	return r
}

func (h *Handler) export(w http.ResponseWriter, r *http.Request) {
	userID := auth.CurrentUserID(r.Context())
	res, err := h.svc.Export(r.Context(), &userID)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	ip, ua := audit.FromRequest(r)
	h.audit.Record(r.Context(), audit.Entry{UserID: &userID, Action: "backup_export_created", EntityType: "backup_export", EntityID: &res.ExportID, IP: ip, UserAgent: ua})

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", res.FileName))
	w.Header().Set("X-Export-SHA256", res.SHA256)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(res.Data)
}

type backupRunResponse struct {
	ID         string     `json:"id"`
	Status     string     `json:"status"`
	FileName   *string    `json:"file_name"`
	SHA256     *string    `json:"sha256"`
	CreatedAt  time.Time  `json:"created_at"`
	FinishedAt *time.Time `json:"finished_at"`
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	runs, err := h.svc.List(r.Context(), 50)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	out := make([]backupRunResponse, 0, len(runs))
	for i := range runs {
		b := &runs[i]
		out = append(out, backupRunResponse{ID: b.ID, Status: b.Status, FileName: b.FileName, SHA256: b.SHA256, CreatedAt: b.CreatedAt, FinishedAt: b.FinishedAt})
	}
	httpx.JSON(w, http.StatusOK, out)
}

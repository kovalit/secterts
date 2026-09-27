package appsecrets

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/kovalit/secrets-center/backend/internal/audit"
	"github.com/kovalit/secrets-center/backend/internal/auth"
	"github.com/kovalit/secrets-center/backend/internal/db"
	"github.com/kovalit/secrets-center/backend/internal/httpx"
)

// Handler exposes the app-secrets HTTP API.
type Handler struct {
	svc   *Service
	audit *audit.Service
}

// NewHandler builds the handler.
func NewHandler(svc *Service, auditSvc *audit.Service) *Handler {
	return &Handler{svc: svc, audit: auditSvc}
}

// ProjectRoutes handles /app-projects (+ nested environments and secrets).
func (h *Handler) ProjectRoutes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.listProjects)
	r.Post("/", h.createProject)
	r.Get("/{id}", h.getProject)
	r.Put("/{id}", h.updateProject)
	r.Delete("/{id}", h.deleteProject)

	r.Get("/{project_id}/environments", h.listEnvironments)
	r.Post("/{project_id}/environments", h.createEnvironment)

	r.Get("/{project_id}/secrets", h.listSecrets)
	r.Post("/{project_id}/secrets", h.createSecret)
	return r
}

// EnvironmentRoutes handles /app-environments/{id}.
func (h *Handler) EnvironmentRoutes() http.Handler {
	r := chi.NewRouter()
	r.Put("/{id}", h.updateEnvironment)
	r.Delete("/{id}", h.deleteEnvironment)
	return r
}

// SecretRoutes handles /app-secrets/{id}.
func (h *Handler) SecretRoutes() http.Handler {
	r := chi.NewRouter()
	r.Put("/{id}", h.updateSecret)
	r.Delete("/{id}", h.deleteSecret)
	r.Post("/{id}/reveal", h.revealSecret)
	return r
}

// --- projects ---

type projectWriteRequest struct {
	CompanyID   *string `json:"company_id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

func (h *Handler) listProjects(w http.ResponseWriter, r *http.Request) {
	ownerID := auth.CurrentUserID(r.Context())
	items, err := h.svc.ListProjects(r.Context(), ownerID)
	if err != nil {
		h.writeErr(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, items)
}

func (h *Handler) getProject(w http.ResponseWriter, r *http.Request) {
	ownerID := auth.CurrentUserID(r.Context())
	p, err := h.svc.GetProject(r.Context(), ownerID, chi.URLParam(r, "id"))
	if err != nil {
		h.writeErr(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, p)
}

func (h *Handler) createProject(w http.ResponseWriter, r *http.Request) {
	ownerID := auth.CurrentUserID(r.Context())
	var in projectWriteRequest
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Error(w, err)
		return
	}
	p, err := h.svc.CreateProject(r.Context(), ownerID, in.CompanyID, in.Name, in.Description)
	if err != nil {
		h.writeErr(w, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, p)
}

func (h *Handler) updateProject(w http.ResponseWriter, r *http.Request) {
	ownerID := auth.CurrentUserID(r.Context())
	var in projectWriteRequest
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Error(w, err)
		return
	}
	p, err := h.svc.UpdateProject(r.Context(), ownerID, chi.URLParam(r, "id"), in.CompanyID, in.Name, in.Description)
	if err != nil {
		h.writeErr(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, p)
}

func (h *Handler) deleteProject(w http.ResponseWriter, r *http.Request) {
	ownerID := auth.CurrentUserID(r.Context())
	if err := h.svc.DeleteProject(r.Context(), ownerID, chi.URLParam(r, "id")); err != nil {
		h.writeErr(w, err)
		return
	}
	httpx.NoContent(w)
}

// --- environments ---

type envWriteRequest struct {
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

func (h *Handler) listEnvironments(w http.ResponseWriter, r *http.Request) {
	ownerID := auth.CurrentUserID(r.Context())
	items, err := h.svc.ListEnvironments(r.Context(), ownerID, chi.URLParam(r, "project_id"))
	if err != nil {
		h.writeErr(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, items)
}

func (h *Handler) createEnvironment(w http.ResponseWriter, r *http.Request) {
	ownerID := auth.CurrentUserID(r.Context())
	var in envWriteRequest
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Error(w, err)
		return
	}
	if in.SortOrder == 0 {
		in.SortOrder = 100
	}
	e, err := h.svc.CreateEnvironment(r.Context(), ownerID, chi.URLParam(r, "project_id"), in.Name, in.SortOrder)
	if err != nil {
		h.writeErr(w, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, e)
}

func (h *Handler) updateEnvironment(w http.ResponseWriter, r *http.Request) {
	ownerID := auth.CurrentUserID(r.Context())
	var in envWriteRequest
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Error(w, err)
		return
	}
	if in.SortOrder == 0 {
		in.SortOrder = 100
	}
	e, err := h.svc.UpdateEnvironment(r.Context(), ownerID, chi.URLParam(r, "id"), in.Name, in.SortOrder)
	if err != nil {
		h.writeErr(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, e)
}

func (h *Handler) deleteEnvironment(w http.ResponseWriter, r *http.Request) {
	ownerID := auth.CurrentUserID(r.Context())
	if err := h.svc.DeleteEnvironment(r.Context(), ownerID, chi.URLParam(r, "id")); err != nil {
		h.writeErr(w, err)
		return
	}
	httpx.NoContent(w)
}

// --- secrets ---

type secretWriteRequest struct {
	EnvironmentID string  `json:"environment_id"`
	Key           string  `json:"key"`
	Value         string  `json:"value"`
	Comment       *string `json:"comment"`
}

func (h *Handler) listSecrets(w http.ResponseWriter, r *http.Request) {
	ownerID := auth.CurrentUserID(r.Context())
	items, err := h.svc.ListSecrets(r.Context(), ownerID, chi.URLParam(r, "project_id"), r.URL.Query().Get("env"))
	if err != nil {
		h.writeErr(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, items)
}

func (h *Handler) createSecret(w http.ResponseWriter, r *http.Request) {
	ownerID := auth.CurrentUserID(r.Context())
	var in secretWriteRequest
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Error(w, err)
		return
	}
	v, err := h.svc.CreateSecret(r.Context(), ownerID, chi.URLParam(r, "project_id"), SecretWriteInput{
		EnvironmentID: in.EnvironmentID, Key: in.Key, Value: in.Value, Comment: in.Comment,
	})
	if err != nil {
		h.writeErr(w, err)
		return
	}
	ip, ua := audit.FromRequest(r)
	h.audit.Record(r.Context(), audit.Entry{UserID: &ownerID, Action: "app_secret_created", EntityType: "app_secret", EntityID: &v.ID, IP: ip, UserAgent: ua})
	httpx.JSON(w, http.StatusCreated, v)
}

func (h *Handler) updateSecret(w http.ResponseWriter, r *http.Request) {
	ownerID := auth.CurrentUserID(r.Context())
	id := chi.URLParam(r, "id")
	var in secretWriteRequest
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Error(w, err)
		return
	}
	v, err := h.svc.UpdateSecret(r.Context(), ownerID, id, SecretWriteInput{
		Key: in.Key, Value: in.Value, Comment: in.Comment,
	})
	if err != nil {
		h.writeErr(w, err)
		return
	}
	ip, ua := audit.FromRequest(r)
	h.audit.Record(r.Context(), audit.Entry{UserID: &ownerID, Action: "app_secret_updated", EntityType: "app_secret", EntityID: &v.ID, IP: ip, UserAgent: ua})
	httpx.JSON(w, http.StatusOK, v)
}

func (h *Handler) deleteSecret(w http.ResponseWriter, r *http.Request) {
	ownerID := auth.CurrentUserID(r.Context())
	id := chi.URLParam(r, "id")
	if err := h.svc.DeleteSecret(r.Context(), ownerID, id); err != nil {
		h.writeErr(w, err)
		return
	}
	ip, ua := audit.FromRequest(r)
	h.audit.Record(r.Context(), audit.Entry{UserID: &ownerID, Action: "app_secret_deleted", EntityType: "app_secret", EntityID: &id, IP: ip, UserAgent: ua})
	httpx.NoContent(w)
}

func (h *Handler) revealSecret(w http.ResponseWriter, r *http.Request) {
	ownerID := auth.CurrentUserID(r.Context())
	id := chi.URLParam(r, "id")
	value, err := h.svc.RevealSecret(r.Context(), ownerID, id)
	if err != nil {
		h.writeErr(w, err)
		return
	}
	ip, ua := audit.FromRequest(r)
	h.audit.Record(r.Context(), audit.Entry{UserID: &ownerID, Action: "app_secret_revealed", EntityType: "app_secret", EntityID: &id, IP: ip, UserAgent: ua})
	httpx.JSON(w, http.StatusOK, map[string]string{"value": value})
}

func (h *Handler) writeErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, db.ErrNotFound):
		httpx.Error(w, httpx.ErrNotFound("not found"))
	case errors.Is(err, ErrForbidden):
		httpx.Error(w, httpx.ErrForbidden("you do not have access to this resource"))
	case errors.Is(err, ErrKeyReq), errors.Is(err, ErrValueReq), errors.Is(err, ErrNameReq), errors.Is(err, ErrEnvUnknown):
		httpx.Error(w, httpx.ErrBadRequest(err.Error()))
	default:
		httpx.Error(w, err)
	}
}

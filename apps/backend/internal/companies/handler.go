// Package companies implements CRUD for the current user's companies.
package companies

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/kovalit/secrets-center/backend/internal/auth"
	"github.com/kovalit/secrets-center/backend/internal/db"
	"github.com/kovalit/secrets-center/backend/internal/httpx"
)

// Handler exposes the companies HTTP API.
type Handler struct {
	store *db.Store
}

// NewHandler builds the handler.
func NewHandler(store *db.Store) *Handler {
	return &Handler{store: store}
}

// Routes returns the /companies subtree (mounted behind auth).
func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Get("/{id}", h.get)
	r.Put("/{id}", h.update)
	r.Delete("/{id}", h.delete)
	return r
}

type companyResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toResponse(c *db.Company) companyResponse {
	return companyResponse{ID: c.ID, Name: c.Name, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt}
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	ownerID := auth.CurrentUserID(r.Context())
	companies, err := h.store.ListCompanies(r.Context(), ownerID)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	out := make([]companyResponse, 0, len(companies))
	for i := range companies {
		out = append(out, toResponse(&companies[i]))
	}
	httpx.JSON(w, http.StatusOK, out)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	ownerID := auth.CurrentUserID(r.Context())
	c, err := h.store.GetCompany(r.Context(), ownerID, chi.URLParam(r, "id"))
	if err != nil {
		h.writeErr(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toResponse(c))
}

type writeRequest struct {
	Name string `json:"name"`
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	ownerID := auth.CurrentUserID(r.Context())
	var in writeRequest
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Error(w, err)
		return
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		httpx.Error(w, httpx.ErrBadRequest("name is required"))
		return
	}
	c, err := h.store.CreateCompany(r.Context(), ownerID, name)
	if err != nil {
		h.writeErr(w, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, toResponse(c))
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	ownerID := auth.CurrentUserID(r.Context())
	id := chi.URLParam(r, "id")
	var in writeRequest
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Error(w, err)
		return
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		httpx.Error(w, httpx.ErrBadRequest("name is required"))
		return
	}
	c, err := h.store.UpdateCompany(r.Context(), ownerID, id, name)
	if err != nil {
		h.writeErr(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, toResponse(c))
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	ownerID := auth.CurrentUserID(r.Context())
	if err := h.store.DeleteCompany(r.Context(), ownerID, chi.URLParam(r, "id")); err != nil {
		h.writeErr(w, err)
		return
	}
	httpx.NoContent(w)
}

func (h *Handler) writeErr(w http.ResponseWriter, err error) {
	if errors.Is(err, db.ErrNotFound) {
		httpx.Error(w, httpx.ErrNotFound("company not found"))
		return
	}
	// Unique violation on (owner_user_id, name).
	if strings.Contains(err.Error(), "companies_owner_user_id_name_key") || strings.Contains(err.Error(), "duplicate key") {
		httpx.Error(w, httpx.ErrConflict("company with this name already exists"))
		return
	}
	httpx.Error(w, err)
}

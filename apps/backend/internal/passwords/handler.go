package passwords

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/kovalit/secrets-center/backend/internal/audit"
	"github.com/kovalit/secrets-center/backend/internal/auth"
	"github.com/kovalit/secrets-center/backend/internal/db"
	"github.com/kovalit/secrets-center/backend/internal/httpx"
)

// Handler exposes the passwords + groups HTTP API.
type Handler struct {
	svc   *Service
	audit *audit.Service
}

// NewHandler builds the handler.
func NewHandler(svc *Service, auditSvc *audit.Service) *Handler {
	return &Handler{svc: svc, audit: auditSvc}
}

// Routes returns the /passwords subtree (mounted behind auth).
func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Get("/inventory", h.inventory)
	r.Get("/health", h.health)
	r.Post("/", h.create)
	r.Get("/{id}", h.get)
	r.Put("/{id}", h.update)
	r.Delete("/{id}", h.delete)
	r.Post("/{id}/reveal", h.reveal)
	return r
}

// ListGroups handles GET /api/password-groups.
func (h *Handler) ListGroups(w http.ResponseWriter, r *http.Request) {
	groups, err := h.svc.ListGroups(r.Context())
	if err != nil {
		httpx.Error(w, err)
		return
	}
	type groupResp struct {
		ID        string `json:"id"`
		Slug      string `json:"slug"`
		Label     string `json:"label"`
		Icon      string `json:"icon"`
		SortOrder int    `json:"sort_order"`
	}
	out := make([]groupResp, 0, len(groups))
	for _, g := range groups {
		out = append(out, groupResp{ID: g.ID, Slug: g.Slug, Label: g.Label, Icon: g.Icon, SortOrder: g.SortOrder})
	}
	httpx.JSON(w, http.StatusOK, out)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	ownerID := auth.CurrentUserID(r.Context())
	q := r.URL.Query()
	f := db.PasswordFilter{
		Scope:     q.Get("scope"),
		CompanyID: q.Get("company_id"),
		GroupID:   q.Get("group_id"),
		EntryType: q.Get("entry_type"),
		Query:     q.Get("q"),
		Domain:    q.Get("domain"),
	}
	items, err := h.svc.List(r.Context(), ownerID, f)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, items)
}

func (h *Handler) inventory(w http.ResponseWriter, r *http.Request) {
	ownerID := auth.CurrentUserID(r.Context())
	inv, err := h.svc.Inventory(r.Context(), ownerID)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, inv)
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	ownerID := auth.CurrentUserID(r.Context())
	report, err := h.svc.Health(r.Context(), ownerID)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, report)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	ownerID := auth.CurrentUserID(r.Context())
	v, err := h.svc.Get(r.Context(), ownerID, chi.URLParam(r, "id"))
	if err != nil {
		h.writeErr(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, v)
}

type writeRequest struct {
	Scope      string     `json:"scope"`
	CompanyID  *string    `json:"company_id"`
	GroupID    string     `json:"group_id"`
	Title      string     `json:"title"`
	SiteURL    *string    `json:"site_url"`
	Login      *string    `json:"login"`
	Password   string     `json:"password"`
	Comment    *string    `json:"comment"`
	IconSource *string    `json:"icon_source"`
	CustomIcon *string    `json:"custom_icon"`
	EntryType  string     `json:"entry_type"`
	ExpiresAt  *time.Time `json:"expires_at"`
	Owner      *string    `json:"owner"`
}

func (in writeRequest) toInput() WriteInput {
	return WriteInput{
		Scope:      in.Scope,
		CompanyID:  in.CompanyID,
		GroupID:    in.GroupID,
		Title:      in.Title,
		SiteURL:    in.SiteURL,
		Login:      in.Login,
		Password:   in.Password,
		Comment:    in.Comment,
		IconSource: in.IconSource,
		CustomIcon: in.CustomIcon,
		EntryType:  in.EntryType,
		ExpiresAt:  in.ExpiresAt,
		Owner:      in.Owner,
	}
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	ownerID := auth.CurrentUserID(r.Context())
	var in writeRequest
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Error(w, err)
		return
	}
	v, err := h.svc.Create(r.Context(), ownerID, in.toInput())
	if err != nil {
		h.writeErr(w, err)
		return
	}
	ip, ua := audit.FromRequest(r)
	h.audit.Record(r.Context(), audit.Entry{UserID: &ownerID, Action: "password_entry_created", EntityType: "password_entry", EntityID: &v.ID, IP: ip, UserAgent: ua})
	httpx.JSON(w, http.StatusCreated, v)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	ownerID := auth.CurrentUserID(r.Context())
	id := chi.URLParam(r, "id")
	var in writeRequest
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Error(w, err)
		return
	}
	v, err := h.svc.Update(r.Context(), ownerID, id, in.toInput())
	if err != nil {
		h.writeErr(w, err)
		return
	}
	ip, ua := audit.FromRequest(r)
	h.audit.Record(r.Context(), audit.Entry{UserID: &ownerID, Action: "password_entry_updated", EntityType: "password_entry", EntityID: &v.ID, IP: ip, UserAgent: ua})
	httpx.JSON(w, http.StatusOK, v)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	ownerID := auth.CurrentUserID(r.Context())
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), ownerID, id); err != nil {
		h.writeErr(w, err)
		return
	}
	ip, ua := audit.FromRequest(r)
	h.audit.Record(r.Context(), audit.Entry{UserID: &ownerID, Action: "password_entry_deleted", EntityType: "password_entry", EntityID: &id, IP: ip, UserAgent: ua})
	httpx.NoContent(w)
}

func (h *Handler) reveal(w http.ResponseWriter, r *http.Request) {
	ownerID := auth.CurrentUserID(r.Context())
	id := chi.URLParam(r, "id")
	password, err := h.svc.Reveal(r.Context(), ownerID, id)
	if err != nil {
		h.writeErr(w, err)
		return
	}
	ip, ua := audit.FromRequest(r)
	h.audit.Record(r.Context(), audit.Entry{UserID: &ownerID, Action: "password_entry_revealed", EntityType: "password_entry", EntityID: &id, IP: ip, UserAgent: ua})
	httpx.JSON(w, http.StatusOK, map[string]string{"password": password})
}

// writeErr maps service/store errors to HTTP responses.
func (h *Handler) writeErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, db.ErrNotFound):
		httpx.Error(w, httpx.ErrNotFound("password entry not found"))
	case errors.Is(err, ErrInvalidScope), errors.Is(err, ErrCompanyReq), errors.Is(err, ErrCompanyOnPers),
		errors.Is(err, ErrGroupRequired), errors.Is(err, ErrPasswordReq),
		errors.Is(err, ErrCompanyUnknown), errors.Is(err, ErrGroupUnknown), errors.Is(err, ErrInvalidType):
		httpx.Error(w, httpx.ErrBadRequest(err.Error()))
	default:
		httpx.Error(w, err)
	}
}

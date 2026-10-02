package secretnotes

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/kovalit/secrets-center/backend/internal/audit"
	"github.com/kovalit/secrets-center/backend/internal/auth"
	"github.com/kovalit/secrets-center/backend/internal/db"
	"github.com/kovalit/secrets-center/backend/internal/httpx"
)

// Handler exposes the secret-notes HTTP API.
type Handler struct {
	svc   *Service
	audit *audit.Service
}

// NewHandler builds the handler.
func NewHandler(svc *Service, auditSvc *audit.Service) *Handler {
	return &Handler{svc: svc, audit: auditSvc}
}

// Routes returns the /secret-notes subtree (mounted behind auth).
func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Get("/{id}", h.get)
	r.Put("/{id}", h.update)
	r.Delete("/{id}", h.delete)
	r.Post("/{id}/reveal", h.reveal)
	return r
}

type writeRequest struct {
	Title string `json:"title"`
	Type  string `json:"type"`
	Text  string `json:"text"`
}

func (in writeRequest) toInput() WriteInput {
	return WriteInput{Title: in.Title, Type: in.Type, Text: in.Text}
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	ownerID := auth.CurrentUserID(r.Context())
	q := r.URL.Query()
	f := db.SecretNoteFilter{Type: q.Get("type"), Query: q.Get("q")}
	items, err := h.svc.List(r.Context(), ownerID, f)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, items)
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
	h.audit.Record(r.Context(), audit.Entry{UserID: &ownerID, Action: "secret_note_created", EntityType: "secret_note", EntityID: &v.ID, IP: ip, UserAgent: ua})
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
	h.audit.Record(r.Context(), audit.Entry{UserID: &ownerID, Action: "secret_note_updated", EntityType: "secret_note", EntityID: &v.ID, IP: ip, UserAgent: ua})
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
	h.audit.Record(r.Context(), audit.Entry{UserID: &ownerID, Action: "secret_note_deleted", EntityType: "secret_note", EntityID: &id, IP: ip, UserAgent: ua})
	httpx.NoContent(w)
}

func (h *Handler) reveal(w http.ResponseWriter, r *http.Request) {
	ownerID := auth.CurrentUserID(r.Context())
	id := chi.URLParam(r, "id")
	text, err := h.svc.Reveal(r.Context(), ownerID, id)
	if err != nil {
		h.writeErr(w, err)
		return
	}
	ip, ua := audit.FromRequest(r)
	h.audit.Record(r.Context(), audit.Entry{UserID: &ownerID, Action: "secret_note_revealed", EntityType: "secret_note", EntityID: &id, IP: ip, UserAgent: ua})
	httpx.JSON(w, http.StatusOK, map[string]string{"text": text})
}

// writeErr maps service/store errors to HTTP responses.
func (h *Handler) writeErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, db.ErrNotFound):
		httpx.Error(w, httpx.ErrNotFound("secret note not found"))
	case errors.Is(err, ErrTitleReq), errors.Is(err, ErrTextReq), errors.Is(err, ErrInvalidType):
		httpx.Error(w, httpx.ErrBadRequest(err.Error()))
	default:
		httpx.Error(w, err)
	}
}

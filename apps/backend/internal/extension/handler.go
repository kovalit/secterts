package extension

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/kovalit/secrets-center/backend/internal/audit"
	"github.com/kovalit/secrets-center/backend/internal/auth"
	"github.com/kovalit/secrets-center/backend/internal/db"
	"github.com/kovalit/secrets-center/backend/internal/httpx"
)

// Handler exposes the extension HTTP API.
type Handler struct {
	svc   *Service
	audit *audit.Service
}

// NewHandler builds the handler.
func NewHandler(svc *Service, auditSvc *audit.Service) *Handler {
	return &Handler{svc: svc, audit: auditSvc}
}

// Routes returns the full /extension subtree. Token management sits behind the
// web session (requireWebAuth); lookup/reveal sit behind the bearer token.
func (h *Handler) Routes(requireWebAuth func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()

	// Token management (web session with email 2FA).
	r.Group(func(mgmt chi.Router) {
		mgmt.Use(requireWebAuth)
		mgmt.Get("/tokens", h.listTokens)
		mgmt.Post("/tokens", h.createToken)
		mgmt.Delete("/tokens/{id}", h.revokeToken)
	})

	// Lookup / reveal (bearer extension token).
	r.Group(func(bearer chi.Router) {
		bearer.Use(h.bearerAuth)
		bearer.Get("/lookup", h.lookup)
		bearer.Post("/passwords/{id}/reveal", h.reveal)
	})

	return r
}

// bearerAuth authenticates using the Authorization: Bearer <token> header and
// injects the owning user into the request context.
func (h *Handler) bearerAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			httpx.Error(w, httpx.ErrUnauthorized("missing bearer token"))
			return
		}
		raw := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		tok, err := h.svc.Authenticate(r.Context(), raw)
		if err != nil {
			httpx.Error(w, httpx.ErrUnauthorized("invalid or expired extension token"))
			return
		}
		user, err := h.svc.store.GetUserByID(r.Context(), tok.UserID)
		if err != nil || !user.IsActive {
			httpx.Error(w, httpx.ErrUnauthorized("account unavailable"))
			return
		}
		next.ServeHTTP(w, r.WithContext(auth.WithUser(r.Context(), user)))
	})
}

// --- token management (web session) ---

type createTokenRequest struct {
	Name string `json:"name"`
}

func (h *Handler) createToken(w http.ResponseWriter, r *http.Request) {
	userID := auth.CurrentUserID(r.Context())
	var in createTokenRequest
	if err := httpx.Decode(r, &in); err != nil {
		// name is optional; ignore decode errors on empty bodies
		in = createTokenRequest{}
	}
	raw, view, err := h.svc.CreateToken(r.Context(), userID, in.Name)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	ip, ua := audit.FromRequest(r)
	h.audit.Record(r.Context(), audit.Entry{UserID: &userID, Action: "extension_token_created", EntityType: "extension_token", EntityID: &view.ID, IP: ip, UserAgent: ua})
	// The raw token is returned exactly once.
	httpx.JSON(w, http.StatusCreated, map[string]any{"token": raw, "token_info": view})
}

func (h *Handler) listTokens(w http.ResponseWriter, r *http.Request) {
	userID := auth.CurrentUserID(r.Context())
	items, err := h.svc.ListTokens(r.Context(), userID)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, items)
}

func (h *Handler) revokeToken(w http.ResponseWriter, r *http.Request) {
	userID := auth.CurrentUserID(r.Context())
	id := chi.URLParam(r, "id")
	if err := h.svc.RevokeToken(r.Context(), userID, id); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			httpx.Error(w, httpx.ErrNotFound("token not found"))
			return
		}
		httpx.Error(w, err)
		return
	}
	ip, ua := audit.FromRequest(r)
	h.audit.Record(r.Context(), audit.Entry{UserID: &userID, Action: "extension_token_revoked", EntityType: "extension_token", EntityID: &id, IP: ip, UserAgent: ua})
	httpx.NoContent(w)
}

// --- lookup / reveal (bearer auth) ---

func (h *Handler) lookup(w http.ResponseWriter, r *http.Request) {
	userID := auth.CurrentUserID(r.Context())
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		httpx.Error(w, httpx.ErrBadRequest("domain query parameter is required"))
		return
	}
	items, err := h.svc.Lookup(r.Context(), userID, domain)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"domain": strings.TrimPrefix(strings.ToLower(domain), "www."), "items": items})
}

type revealRequest struct {
	Domain string `json:"domain"`
}

func (h *Handler) reveal(w http.ResponseWriter, r *http.Request) {
	userID := auth.CurrentUserID(r.Context())
	id := chi.URLParam(r, "id")
	var in revealRequest
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Error(w, err)
		return
	}
	password, err := h.svc.Reveal(r.Context(), userID, id, in.Domain)
	switch {
	case errors.Is(err, db.ErrNotFound):
		httpx.Error(w, httpx.ErrNotFound("entry not found"))
		return
	case errors.Is(err, ErrDomainMismatch):
		httpx.Error(w, httpx.ErrForbidden("domain does not match the entry"))
		return
	case errors.Is(err, ErrNoPassword):
		httpx.Error(w, httpx.ErrBadRequest("entry has no password"))
		return
	case err != nil:
		httpx.Error(w, err)
		return
	}

	ip, ua := audit.FromRequest(r)
	h.audit.Record(r.Context(), audit.Entry{UserID: &userID, Action: "password_entry_copied_by_extension", EntityType: "password_entry", EntityID: &id, IP: ip, UserAgent: ua})
	httpx.JSON(w, http.StatusOK, map[string]string{"password": password})
}

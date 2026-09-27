package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/kovalit/secrets-center/backend/internal/audit"
	"github.com/kovalit/secrets-center/backend/internal/db"
	"github.com/kovalit/secrets-center/backend/internal/httpx"
)

// Handler exposes the auth HTTP API.
type Handler struct {
	svc          *Service
	audit        *audit.Service
	cookieDomain string
	cookieSecure bool
}

// NewHandler builds the auth handler.
func NewHandler(svc *Service, auditSvc *audit.Service, cookieDomain string, cookieSecure bool) *Handler {
	return &Handler{svc: svc, audit: auditSvc, cookieDomain: cookieDomain, cookieSecure: cookieSecure}
}

// Middleware is an http middleware function.
type Middleware func(http.Handler) http.Handler

// Routes returns the full /auth subtree. Public endpoints get the rate limiter;
// account endpoints require an authenticated session.
func (h *Handler) Routes(requireAuth, rateLimit Middleware) http.Handler {
	r := chi.NewRouter()

	// Public, rate-limited endpoints.
	r.Group(func(pub chi.Router) {
		pub.Use(rateLimit)
		pub.Post("/register", h.register)
		pub.Post("/login", h.login)
		pub.Post("/verify-email-code", h.verify)
		pub.Post("/logout", h.logout)
		pub.Post("/password-reset/request", h.resetRequest)
		pub.Post("/password-reset/confirm", h.resetConfirm)
	})

	// Authenticated account endpoints.
	r.Group(func(priv chi.Router) {
		priv.Use(requireAuth)
		priv.Get("/me", h.me)
		priv.Post("/change-password", h.changePassword)
		priv.Get("/sessions", h.listSessions)
		priv.Delete("/sessions/{id}", h.revokeSession)
	})

	return r
}

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID              string    `json:"id"`
	Email           string    `json:"email"`
	EmailVerified   bool      `json:"email_verified"`
	Email2FAEnabled bool      `json:"email_2fa_enabled"`
	CreatedAt       time.Time `json:"created_at"`
}

func toUserResponse(u *db.User) userResponse {
	return userResponse{
		ID:              u.ID,
		Email:           u.Email,
		EmailVerified:   u.EmailVerified,
		Email2FAEnabled: u.Email2FAEnabled,
		CreatedAt:       u.CreatedAt,
	}
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var in credentials
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Error(w, err)
		return
	}
	if len(in.Password) < 8 {
		httpx.Error(w, httpx.ErrBadRequest("password must be at least 8 characters"))
		return
	}
	if in.Email == "" {
		httpx.Error(w, httpx.ErrBadRequest("email is required"))
		return
	}

	user, err := h.svc.Register(r.Context(), in.Email, in.Password)
	if errors.Is(err, ErrEmailTaken) {
		httpx.Error(w, httpx.ErrConflict("email already registered"))
		return
	}
	if err != nil {
		httpx.Error(w, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, toUserResponse(user))
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var in credentials
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Error(w, err)
		return
	}

	ip, ua := audit.FromRequest(r)
	challengeID, user, err := h.svc.Login(r.Context(), in.Email, in.Password)
	if errors.Is(err, ErrInvalidCredentials) {
		h.audit.Record(r.Context(), audit.Entry{Action: "login_failed", EntityType: "user", IP: ip, UserAgent: ua, Metadata: map[string]any{"email": in.Email}})
		httpx.Error(w, httpx.ErrUnauthorized("invalid email or password"))
		return
	}
	if err != nil {
		httpx.Error(w, err)
		return
	}

	h.audit.Record(r.Context(), audit.Entry{UserID: &user.ID, Action: "email_2fa_sent", EntityType: "user", EntityID: &user.ID, IP: ip, UserAgent: ua})
	httpx.JSON(w, http.StatusOK, map[string]any{
		"challenge_id":    challengeID,
		"requires_2fa":    true,
		"email_2fa":       true,
		"message":         "Код отправлен на email",
	})
}

type verifyRequest struct {
	ChallengeID string `json:"challenge_id"`
	Code        string `json:"code"`
}

func (h *Handler) verify(w http.ResponseWriter, r *http.Request) {
	var in verifyRequest
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Error(w, err)
		return
	}

	ip, ua := audit.FromRequest(r)
	token, expiresAt, user, err := h.svc.VerifyEmailCode(r.Context(), in.ChallengeID, in.Code, ua, ip)
	if errors.Is(err, ErrInvalidCode) {
		httpx.Error(w, httpx.ErrUnauthorized("invalid or expired code"))
		return
	}
	if errors.Is(err, ErrTooManyAttempts) {
		httpx.Error(w, httpx.ErrTooMany("too many attempts, request a new code"))
		return
	}
	if err != nil {
		httpx.Error(w, err)
		return
	}

	h.setSessionCookie(w, token, expiresAt)
	h.audit.Record(r.Context(), audit.Entry{UserID: &user.ID, Action: "login_success", EntityType: "user", EntityID: &user.ID, IP: ip, UserAgent: ua})
	httpx.JSON(w, http.StatusOK, toUserResponse(user))
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(SessionCookieName); err == nil {
		_ = h.svc.Logout(r.Context(), cookie.Value)
	}
	h.clearSessionCookie(w)
	httpx.NoContent(w)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	user := CurrentUser(r.Context())
	if user == nil {
		httpx.Error(w, httpx.ErrUnauthorized("authentication required"))
		return
	}
	httpx.JSON(w, http.StatusOK, toUserResponse(user))
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (h *Handler) changePassword(w http.ResponseWriter, r *http.Request) {
	user := CurrentUser(r.Context())
	var in changePasswordRequest
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Error(w, err)
		return
	}
	if len(in.NewPassword) < 8 {
		httpx.Error(w, httpx.ErrBadRequest("new password must be at least 8 characters"))
		return
	}
	ok, err := VerifyPassword(in.CurrentPassword, user.PasswordHash)
	if err != nil || !ok {
		httpx.Error(w, httpx.ErrBadRequest("current password is incorrect"))
		return
	}
	hash, err := HashPassword(in.NewPassword)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	if err := h.svc.store.UpdateUserPassword(r.Context(), user.ID, hash); err != nil {
		httpx.Error(w, err)
		return
	}
	ip, ua := audit.FromRequest(r)
	h.audit.Record(r.Context(), audit.Entry{UserID: &user.ID, Action: "password_changed", EntityType: "user", EntityID: &user.ID, IP: ip, UserAgent: ua})
	httpx.NoContent(w)
}

type sessionResponse struct {
	ID        string     `json:"id"`
	UserAgent *string    `json:"user_agent"`
	IP        *string    `json:"ip"`
	Current   bool       `json:"current"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at"`
}

func (h *Handler) listSessions(w http.ResponseWriter, r *http.Request) {
	user := CurrentUser(r.Context())
	sessions, err := h.svc.store.ListSessions(r.Context(), user.ID)
	if err != nil {
		httpx.Error(w, err)
		return
	}

	var currentHash string
	if cookie, err := r.Cookie(SessionCookieName); err == nil {
		currentHash = hashToken(cookie.Value)
	}

	out := make([]sessionResponse, 0, len(sessions))
	for _, s := range sessions {
		out = append(out, sessionResponse{
			ID:        s.ID,
			UserAgent: s.UserAgent,
			IP:        s.IP,
			Current:   s.RefreshTokenHash == currentHash,
			CreatedAt: s.CreatedAt,
			ExpiresAt: s.ExpiresAt,
			RevokedAt: s.RevokedAt,
		})
	}
	httpx.JSON(w, http.StatusOK, out)
}

func (h *Handler) revokeSession(w http.ResponseWriter, r *http.Request) {
	user := CurrentUser(r.Context())
	id := chi.URLParam(r, "id")
	if err := h.svc.store.RevokeSession(r.Context(), user.ID, id); err != nil {
		httpx.Error(w, err)
		return
	}
	httpx.NoContent(w)
}

func (h *Handler) resetRequest(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email string `json:"email"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Error(w, err)
		return
	}
	challengeID, err := h.svc.PasswordResetRequest(r.Context(), in.Email)
	if err != nil {
		httpx.Error(w, err)
		return
	}
	// Always 200 so callers cannot probe which emails exist.
	httpx.JSON(w, http.StatusOK, map[string]any{"challenge_id": challengeID})
}

func (h *Handler) resetConfirm(w http.ResponseWriter, r *http.Request) {
	var in struct {
		ChallengeID string `json:"challenge_id"`
		Code        string `json:"code"`
		NewPassword string `json:"new_password"`
	}
	if err := httpx.Decode(r, &in); err != nil {
		httpx.Error(w, err)
		return
	}
	if len(in.NewPassword) < 8 {
		httpx.Error(w, httpx.ErrBadRequest("new password must be at least 8 characters"))
		return
	}
	err := h.svc.PasswordResetConfirm(r.Context(), in.ChallengeID, in.Code, in.NewPassword)
	if errors.Is(err, ErrInvalidCode) {
		httpx.Error(w, httpx.ErrUnauthorized("invalid or expired code"))
		return
	}
	if err != nil {
		httpx.Error(w, err)
		return
	}
	httpx.NoContent(w)
}

// --- cookie helpers ---

func (h *Handler) setSessionCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	cookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
	}
	if h.cookieDomain != "" && h.cookieDomain != "localhost" {
		cookie.Domain = h.cookieDomain
	}
	http.SetCookie(w, cookie)
}

func (h *Handler) clearSessionCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	}
	if h.cookieDomain != "" && h.cookieDomain != "localhost" {
		cookie.Domain = h.cookieDomain
	}
	http.SetCookie(w, cookie)
}

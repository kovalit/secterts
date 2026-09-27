package auth

import (
	"context"
	"net/http"

	"github.com/kovalit/secrets-center/backend/internal/db"
	"github.com/kovalit/secrets-center/backend/internal/httpx"
)

// SessionCookieName is the HttpOnly cookie holding the session refresh token.
const SessionCookieName = "sc_session"

type ctxKey string

const userKey ctxKey = "current_user"

// WithUser stores the current user in the context.
func WithUser(ctx context.Context, u *db.User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

// CurrentUser returns the authenticated user, or nil.
func CurrentUser(ctx context.Context) *db.User {
	if u, ok := ctx.Value(userKey).(*db.User); ok {
		return u
	}
	return nil
}

// CurrentUserID returns the authenticated user's id, or "".
func CurrentUserID(ctx context.Context) string {
	if u := CurrentUser(ctx); u != nil {
		return u.ID
	}
	return ""
}

// RequireAuth authenticates a request from the session cookie and injects the
// user into the context. Unauthenticated requests get 401.
func (s *Service) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(SessionCookieName)
		if err != nil || cookie.Value == "" {
			httpx.Error(w, httpx.ErrUnauthorized("authentication required"))
			return
		}

		sess, err := s.store.GetSessionByTokenHash(r.Context(), hashToken(cookie.Value))
		if err != nil {
			httpx.Error(w, httpx.ErrUnauthorized("invalid or expired session"))
			return
		}

		user, err := s.store.GetUserByID(r.Context(), sess.UserID)
		if err != nil || !user.IsActive {
			httpx.Error(w, httpx.ErrUnauthorized("account unavailable"))
			return
		}

		ctx := WithUser(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

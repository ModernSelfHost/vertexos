package httpapi

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/ModernSelfHost/vertexos/internal/auth"
	"github.com/ModernSelfHost/vertexos/internal/sessions"
	"github.com/ModernSelfHost/vertexos/internal/users"
)

type contextKey string

const userContextKey contextKey = "user"

// AuthMiddleware validates session cookie and attaches user to context.
func AuthMiddleware(userStore *users.Store, sessionStore *sessions.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie(sessions.CookieName)
			if err != nil || c.Value == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			userID, csrf, err := sessionStore.Get(c.Value)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			// CSRF check for unsafe methods.
			if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodOptions {
				if r.Header.Get(sessions.CSRFHeader) != csrf {
					http.Error(w, "csrf failure", http.StatusForbidden)
					return
				}
			}
			u, err := userStore.GetByID(userID)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), userContextKey, u)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserFromContext returns the authenticated user, if any.
func UserFromContext(ctx context.Context) (auth.User, bool) {
	u, ok := ctx.Value(userContextKey).(auth.User)
	return u, ok
}

// RequireRole enforces a minimum role.
func RequireRole(role auth.Role) func(http.Handler) http.Handler {
	order := map[auth.Role]int{auth.RoleGuest: 0, auth.RoleUser: 1, auth.RoleAdmin: 2}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, ok := UserFromContext(r.Context())
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if order[u.Role] < order[role] {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimiter is a simple per-key token bucket in memory.
type RateLimiter struct {
	// not implemented fully; placeholder for login rate limiting
}

// ClientIP extracts client IP.
func ClientIP(r *http.Request) string {
	if x := r.Header.Get("X-Forwarded-For"); x != "" {
		parts := strings.Split(x, ",")
		return strings.TrimSpace(parts[0])
	}
	return r.RemoteAddr
}

// securityHeaders wraps a handler adding standard security headers.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-XSS-Protection", "0")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		next.ServeHTTP(w, r)
	})
}

// SessionTTL default.
var SessionTTL = 24 * time.Hour

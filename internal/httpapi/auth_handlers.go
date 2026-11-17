package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ModernSelfHost/vertexos/internal/auth"
	"github.com/ModernSelfHost/vertexos/internal/ratelimit"
	"github.com/ModernSelfHost/vertexos/internal/sessions"
	"github.com/ModernSelfHost/vertexos/internal/users"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func registerAuthRoutes(mux *http.ServeMux, deps RouterDeps) {
	loginLimiter := ratelimit.New(10, 15*time.Minute)
	mux.HandleFunc("POST /api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if !loginLimiter.Allow(ClientIP(r)) {
			http.Error(w, "too many attempts", http.StatusTooManyRequests)
			return
		}
		if deps.UserStore == nil || deps.SessionStore == nil {
			http.Error(w, "not configured", http.StatusServiceUnavailable)
			return
		}
		var req loginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		u, err := deps.UserStore.GetByUsername(req.Username)
		if err != nil {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		if u.IsLocked() {
			http.Error(w, "account locked", http.StatusLocked)
			return
		}
		if !auth.VerifyPassword(req.Password, u.PasswordHash) {
			_ = deps.UserStore.RecordFailedAttempt(u)
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		_ = deps.UserStore.ResetFailedAttempts(u.ID)
		token, csrf, err := deps.SessionStore.Create(u.ID, SessionTTL)
		if err != nil {
			http.Error(w, "session error", http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     sessions.CookieName,
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   int(SessionTTL.Seconds()),
		})
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"user": map[string]any{
				"id":       u.ID,
				"username": u.Username,
				"role":     u.Role,
			},
			"csrf": csrf,
		})
	})

	mux.HandleFunc("POST /api/v1/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		if c, err := r.Cookie(sessions.CookieName); err == nil {
			_ = deps.SessionStore.Delete(c.Value)
		}
		http.SetCookie(w, &http.Cookie{
			Name:     sessions.CookieName,
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   -1,
		})
		w.WriteHeader(http.StatusNoContent)
	})
}

// ensure users import used
var _ = users.Store{}

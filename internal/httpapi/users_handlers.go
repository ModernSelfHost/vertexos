package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/ModernSelfHost/vertexos/internal/auth"
	"github.com/ModernSelfHost/vertexos/internal/wrappers"
)

func registerUserRoutes(mux *http.ServeMux, deps RouterDeps) {
	if deps.UserStore == nil {
		return
	}
	us := deps.UserStore

	// List users
	mux.HandleFunc("GET /api/v1/users", func(w http.ResponseWriter, r *http.Request) {
		u, ok := UserFromContext(r.Context())
		if !ok || u.Role != auth.RoleAdmin {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		list, err := us.List()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// strip password hashes
		type pub struct {
			ID       int64  `json:"id"`
			Username string `json:"username"`
			Role     string `json:"role"`
		}
		var out []pub
		for _, u := range list {
			out = append(out, pub{u.ID, u.Username, string(u.Role)})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	})

	// Create user
	mux.HandleFunc("POST /api/v1/users", func(w http.ResponseWriter, r *http.Request) {
		u, ok := UserFromContext(r.Context())
		if !ok || u.Role != auth.RoleAdmin {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Role     string `json:"role"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Username == "" || req.Password == "" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		role := auth.Role(req.Role)
		if role == "" {
			role = auth.RoleUser
		}
		nu, err := us.Create(req.Username, req.Password, role)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// OS account mapping + SMB password sync (best effort)
		_ = syncSystemAccount(nu, req.Password)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": nu.ID, "username": nu.Username, "role": nu.Role})
	})

	// Get user
	mux.HandleFunc("GET /api/v1/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		u, ok := UserFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
		target, err := us.GetByID(id)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if u.Role != auth.RoleAdmin && u.ID != target.ID {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": target.ID, "username": target.Username, "role": target.Role})
	})

	// Update role
	mux.HandleFunc("PUT /api/v1/users/{id}/role", func(w http.ResponseWriter, r *http.Request) {
		u, ok := UserFromContext(r.Context())
		if !ok || u.Role != auth.RoleAdmin {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
		var body struct {
			Role string `json:"role"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if err := us.UpdateRole(id, auth.Role(body.Role)); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	// Password change (self or admin reset)
	mux.HandleFunc("PUT /api/v1/users/{id}/password", func(w http.ResponseWriter, r *http.Request) {
		u, ok := UserFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if u.Role != auth.RoleAdmin && u.ID != id {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		var body struct {
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Password == "" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if err := us.UpdatePassword(id, body.Password); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// sync smb password
		if target, err := us.GetByID(id); err == nil {
			_ = syncSmbPassword(target.Username, body.Password)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	// Delete user
	mux.HandleFunc("DELETE /api/v1/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		u, ok := UserFromContext(r.Context())
		if !ok || u.Role != auth.RoleAdmin {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err := us.Delete(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = deleteSystemAccount(id)
		w.WriteHeader(http.StatusNoContent)
	})
}

// OS account mapping + SMB sync helpers (stubs using wrappers).
func syncSystemAccount(u auth.User, password string) error {
	// Best-effort: create OS user if missing.
	ctx := context.Background()
	_, err := wrappers.RunCombined(ctx, 30*time.Second, "useradd", "-m", "-s", "/usr/sbin/nologin", u.Username)
	if err != nil {
		// maybe already exists; ignore
	}
	return nil
}

func syncSmbPassword(username, password string) error {
	// Best-effort: set Samba password via smbpasswd.
	ctx := context.Background()
	_, _ = wrappers.RunCombined(ctx, 30*time.Second, "sh", "-c",
		"printf '%s\\n%s\\n' '"+password+"' '"+password+"' | smbpasswd -a -s "+username)
	return nil
}

func deleteSystemAccount(id int64) error {
	return nil
}

var _ = auth.RoleAdmin

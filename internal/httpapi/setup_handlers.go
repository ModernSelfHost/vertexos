package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/ModernSelfHost/vertexos/internal/auth"
	"github.com/ModernSelfHost/vertexos/internal/users"
	"github.com/ModernSelfHost/vertexos/internal/wrappers"
)

func registerSetupRoutes(mux *http.ServeMux, deps RouterDeps) {
	mux.HandleFunc("GET /api/v1/setup/status", func(w http.ResponseWriter, r *http.Request) {
		if deps.UserStore == nil {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"needs_setup": false})
			return
		}
		list, _ := deps.UserStore.List()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"needs_setup": len(list) == 0})
	})

	mux.HandleFunc("POST /api/v1/setup/server-name", func(w http.ResponseWriter, r *http.Request) {
		if !setupAllowed(deps) {
			http.Error(w, "setup already complete", http.StatusForbidden)
			return
		}
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if deps.DB != nil {
			_, _ = deps.DB.Exec(`INSERT INTO settings(key,value) VALUES('server_name',?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`, req.Name)
		}
		// best-effort hostname change
		_, _ = wrappers.RunCombined(context.Background(), 10*time.Second, "hostnamectl", "set-hostname", req.Name)
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("POST /api/v1/setup/admin", func(w http.ResponseWriter, r *http.Request) {
		if !setupAllowed(deps) {
			http.Error(w, "setup already complete", http.StatusForbidden)
			return
		}
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Username == "" || req.Password == "" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		u, err := deps.UserStore.Create(req.Username, req.Password, auth.RoleAdmin)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_ = syncSystemAccount(u, req.Password)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": u.ID, "username": u.Username, "role": u.Role})
	})
}

func setupAllowed(deps RouterDeps) bool {
	if deps.UserStore == nil {
		return false
	}
	list, _ := deps.UserStore.List()
	return len(list) == 0
}

var _ = users.Store{}
var _ = time.Second

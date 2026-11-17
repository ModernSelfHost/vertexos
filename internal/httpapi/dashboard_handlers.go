package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/ModernSelfHost/vertexos/internal/wrappers"
)

func registerDashboardRoutes(mux *http.ServeMux, deps RouterDeps) {
	mux.HandleFunc("GET /api/v1/dashboard/storage", func(w http.ResponseWriter, r *http.Request) {
		pools, err := wrappers.ListPools(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		type poolSummary struct {
			Name   string `json:"name"`
			Health string `json:"health"`
			Size   string `json:"size"`
			Alloc  string `json:"alloc"`
			Free   string `json:"free"`
		}
		var out []poolSummary
		for _, p := range pools {
			out = append(out, poolSummary{p.Name, p.Health, p.Size, p.Alloc, p.Free})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"pools": out})
	})
}

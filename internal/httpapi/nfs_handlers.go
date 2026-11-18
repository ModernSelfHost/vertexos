package httpapi

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/ModernSelfHost/vertexos/internal/wrappers"
)

func registerNFSRoutes(mux *http.ServeMux, deps RouterDeps) {
	mux.HandleFunc("GET /api/v1/nfs/exports", func(w http.ResponseWriter, r *http.Request) {
		if deps.DB == nil {
			http.Error(w, "not configured", http.StatusServiceUnavailable)
			return
		}
		var val string
		deps.DB.QueryRow(`SELECT value FROM settings WHERE key='nfs_exports'`).Scan(&val)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"exports": val})
	})
	mux.HandleFunc("PUT /api/v1/nfs/exports", func(w http.ResponseWriter, r *http.Request) {
		if deps.DB == nil {
			http.Error(w, "not configured", http.StatusServiceUnavailable)
			return
		}
		var body struct {
			Exports string `json:"exports"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		deps.DB.Exec(`INSERT INTO settings(key,value) VALUES('nfs_exports',?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`, body.Exports)
		_ = os.MkdirAll("/etc/exports.d", 0755)
		_ = os.WriteFile("/etc/exports.d/vertex.exports", []byte(body.Exports+"\n"), 0644)
		_, _ = wrappers.RunCombined(r.Context(), 30*time.Second, "exportfs", "-ra")
		w.WriteHeader(http.StatusNoContent)
	})
}

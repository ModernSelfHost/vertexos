package httpapi

import (
	"encoding/json"
	"net/http"
)

func registerAlertRoutes(mux *http.ServeMux, deps RouterDeps) {
	if deps.DB == nil {
		return
	}
	mux.HandleFunc("GET /api/v1/alerts", func(w http.ResponseWriter, r *http.Request) {
		rows, err := deps.DB.Query(`SELECT id,kind,message,severity,created_at,dismissed FROM alerts ORDER BY id DESC`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		type alert struct {
			ID        int64  `json:"id"`
			Kind      string `json:"kind"`
			Message   string `json:"message"`
			Severity  string `json:"severity"`
			CreatedAt string `json:"created_at"`
			Dismissed bool   `json:"dismissed"`
		}
		var out []alert
		for rows.Next() {
			var a alert
			var dismissed int
			var created string
			rows.Scan(&a.ID, &a.Kind, &a.Message, &a.Severity, &created, &dismissed)
			a.CreatedAt = created
			a.Dismissed = dismissed == 1
			out = append(out, a)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	})
	mux.HandleFunc("POST /api/v1/alerts/{id}/dismiss", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		_, err := deps.DB.Exec(`UPDATE alerts SET dismissed=1 WHERE id=?`, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

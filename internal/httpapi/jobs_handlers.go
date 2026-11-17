package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/ModernSelfHost/vertexos/internal/jobqueue"
)

func registerJobsRoutes(mux *http.ServeMux, deps RouterDeps) {
	if deps.JobManager == nil {
		return
	}
	mux.HandleFunc("GET /api/v1/jobs", func(w http.ResponseWriter, r *http.Request) {
		jobs := deps.JobManager.List()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jobs)
	})
	mux.HandleFunc("GET /api/v1/jobs/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		j, ok := deps.JobManager.Get(id)
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(j)
	})
	mux.HandleFunc("POST /api/v1/jobs/{id}/cancel", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if !deps.JobManager.Cancel(id) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("POST /api/v1/jobs/{id}/retry", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if !deps.JobManager.Retry(id) {
			http.Error(w, "not retryable", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})
}

var _ = jobqueue.Manager{}

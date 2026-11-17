package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/ModernSelfHost/vertexos/internal/apps"
	"github.com/ModernSelfHost/vertexos/internal/jobqueue"
)

func registerAppRoutes(mux *http.ServeMux, deps RouterDeps) {
	mux.HandleFunc("GET /api/v1/apps", func(w http.ResponseWriter, r *http.Request) {
		if deps.AppCatalog == nil {
			http.Error(w, "catalog not configured", http.StatusServiceUnavailable)
			return
		}
		list, err := deps.AppCatalog.List()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(list)
	})

	mux.HandleFunc("POST /api/v1/apps/{id}/install", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if deps.JobManager == nil {
			http.Error(w, "job manager not configured", http.StatusServiceUnavailable)
			return
		}
		rt := deps.AppRuntime
		if rt == nil {
			rt = apps.NoopRuntime{}
		}
		job := deps.JobManager.Enqueue("install-app", func(j *jobqueue.Job, log func(string)) error {
			log("resolving manifest for " + id)
			log("runtime: " + rt.Name())
			return rt.Install(id, map[string]string{})
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(job)
	})

	mux.HandleFunc("DELETE /api/v1/apps/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		rt := deps.AppRuntime
		if rt == nil {
			rt = apps.NoopRuntime{}
		}
		if err := rt.Uninstall(id); err != nil {
			http.Error(w, err.Error(), http.StatusNotImplemented)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

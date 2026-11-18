package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/ModernSelfHost/vertexos/internal/auth"
	"github.com/ModernSelfHost/vertexos/internal/jobqueue"
	"github.com/ModernSelfHost/vertexos/internal/wrappers"
)

func registerUpdateRoutes(mux *http.ServeMux, deps RouterDeps) {
	jm := deps.JobManager

	mux.HandleFunc("GET /api/v1/updates", func(w http.ResponseWriter, r *http.Request) {
		u, ok := UserFromContext(r.Context())
		if !ok || u.Role != auth.RoleAdmin {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		out, err := wrappers.RunCombined(r.Context(), 60*time.Second, "apt-get", "-s", "upgrade")
		if err != nil {
			// fallback to apt list
			out2, err2 := wrappers.RunCombined(r.Context(), 60*time.Second, "apt", "list", "--upgradable")
			if err2 != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			out = out2
		}
		var lines []string
		for _, l := range strings.Split(out, "\n") {
			if strings.TrimSpace(l) != "" {
				lines = append(lines, l)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"updates": lines})
	})

	mux.HandleFunc("POST /api/v1/updates/apply", func(w http.ResponseWriter, r *http.Request) {
		u, ok := UserFromContext(r.Context())
		if !ok || u.Role != auth.RoleAdmin {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if jm == nil {
			http.Error(w, "job manager not configured", http.StatusServiceUnavailable)
			return
		}
		job := jm.Enqueue("os-update", func(j *jobqueue.Job, log func(string)) error {
			// ZFS boot-environment snapshot (best effort)
			_, _ = wrappers.RunCombined(r.Context(), 60*time.Second, "zfs", "snapshot", "rpool/ROOT@pre-update")
			out, err := wrappers.RunCombined(r.Context(), 0, "sh", "-c", "apt-get update && apt-get upgrade -y")
			log(out)
			return err
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(job)
	})
}

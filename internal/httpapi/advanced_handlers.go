package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/ModernSelfHost/vertexos/internal/jobqueue"
	"github.com/ModernSelfHost/vertexos/internal/wrappers"
)

func registerAdvancedRoutes(mux *http.ServeMux, deps RouterDeps) {
	jm := deps.JobManager

	// Datasets
	mux.HandleFunc("GET /api/v1/datasets", func(w http.ResponseWriter, r *http.Request) {
		ds, err := wrappers.ListDatasets(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ds)
	})

	// Pool status (vdevs) raw
	mux.HandleFunc("GET /api/v1/pools/{name}/status", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		out, err := wrappers.RunCombined(r.Context(), 30*time.Second, "zpool", "status", "-j", name)
		if err != nil {
			// fallback without -j
			out2, err2 := wrappers.RunCombined(r.Context(), 30*time.Second, "zpool", "status", name)
			if err2 != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(out2))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(out))
	})

	// Dataset properties
	mux.HandleFunc("GET /api/v1/datasets/{name}/properties", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		prop := r.URL.Query().Get("prop")
		args := []string{"get", "-H", "-p"}
		if prop != "" {
			args = append(args, prop)
		} else {
			args = append(args, "all")
		}
		args = append(args, name)
		out, err := wrappers.RunCombined(r.Context(), 30*time.Second, "zfs", args...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(out))
	})
	mux.HandleFunc("PUT /api/v1/datasets/{name}/properties", func(w http.ResponseWriter, r *http.Request) {
		if jm == nil {
			http.Error(w, "job manager not configured", http.StatusServiceUnavailable)
			return
		}
		name := r.PathValue("name")
		var body struct {
			Property string `json:"property"`
			Value    string `json:"value"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Property == "" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		job := jm.Enqueue("set-dataset-property", func(j *jobqueue.Job, log func(string)) error {
			out, err := wrappers.RunCombined(r.Context(), 0, "zfs", "set", body.Property+"="+body.Value, name)
			log(out)
			return err
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(job)
	})

	// Snapshots list
	mux.HandleFunc("GET /api/v1/snapshots", func(w http.ResponseWriter, r *http.Request) {
		ds := r.URL.Query().Get("dataset")
		args := []string{"list", "-t", "snapshot", "-H", "-o", "name,used,refer"}
		if ds != "" {
			args = append(args, "-r", ds)
		}
		out, err := wrappers.RunCombined(r.Context(), 30*time.Second, "zfs", args...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// parse lines
		type snap struct {
			Name  string `json:"name"`
			Used  string `json:"used"`
			Refer string `json:"refer"`
		}
		var snaps []snap
		for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
			if line == "" {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) < 3 {
				continue
			}
			snaps = append(snaps, snap{Name: fields[0], Used: fields[1], Refer: fields[2]})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(snaps)
	})

	// Create snapshot
	mux.HandleFunc("POST /api/v1/snapshots", func(w http.ResponseWriter, r *http.Request) {
		if jm == nil {
			http.Error(w, "job manager not configured", http.StatusServiceUnavailable)
			return
		}
		var body struct {
			Dataset string `json:"dataset"`
			Name    string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Dataset == "" || body.Name == "" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		snap := body.Dataset + "@" + body.Name
		job := jm.Enqueue("create-snapshot", func(j *jobqueue.Job, log func(string)) error {
			out, err := wrappers.RunCombined(r.Context(), 0, "zfs", "snapshot", snap)
			log(out)
			return err
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(job)
	})

	// Destroy snapshot
	mux.HandleFunc("DELETE /api/v1/snapshots/{name}", func(w http.ResponseWriter, r *http.Request) {
		if jm == nil {
			http.Error(w, "job manager not configured", http.StatusServiceUnavailable)
			return
		}
		name := r.PathValue("name")
		job := jm.Enqueue("destroy-snapshot", func(j *jobqueue.Job, log func(string)) error {
			out, err := wrappers.RunCombined(r.Context(), 0, "zfs", "destroy", name)
			log(out)
			return err
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(job)
	})
}

var _ = strings.TrimSpace

package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ModernSelfHost/vertexos/internal/jobqueue"
	"github.com/ModernSelfHost/vertexos/internal/wrappers"
)

type createPoolRequest struct {
	Name   string   `json:"name"`
	Vdevs  []string `json:"vdevs"`
	Force  bool     `json:"force"`
}

type addVdevRequest struct {
	Vdevs []string `json:"vdevs"`
}

type scrubScheduleRequest struct {
	Pool    string `json:"pool"`
	Cron    string `json:"cron"`
	Enabled bool   `json:"enabled"`
}

func registerPoolRoutes(mux *http.ServeMux, deps RouterDeps) {
	if deps.JobManager == nil {
		return
	}
	jm := deps.JobManager

	mux.HandleFunc("GET /api/v1/pools", func(w http.ResponseWriter, r *http.Request) {
		pools, err := wrappers.ListPools(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(pools)
	})

	mux.HandleFunc("GET /api/v1/pools/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		pools, err := wrappers.ListPools(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, p := range pools {
			if p.Name == name {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(p)
				return
			}
		}
		http.Error(w, "not found", http.StatusNotFound)
	})

	mux.HandleFunc("POST /api/v1/pools", func(w http.ResponseWriter, r *http.Request) {
		var req createPoolRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		job := jm.Enqueue("create-pool", func(j *jobqueue.Job, log func(string)) error {
			args := []string{"create"}
			if req.Force {
				args = append(args, "-f")
			}
			args = append(args, req.Name)
			args = append(args, req.Vdevs...)
			log(fmt.Sprintf("zpool %s", strings.Join(args, " ")))
			out, err := wrappers.RunCombined(r.Context(), 0, "zpool", args...)
			log(out)
			return err
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(job)
	})

	mux.HandleFunc("POST /api/v1/pools/{name}/vdevs", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		var req addVdevRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Vdevs) == 0 {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		job := jm.Enqueue("add-vdev", func(j *jobqueue.Job, log func(string)) error {
			args := append([]string{"add", name}, req.Vdevs...)
			log(fmt.Sprintf("zpool %s", strings.Join(args, " ")))
			out, err := wrappers.RunCombined(r.Context(), 0, "zpool", args...)
			log(out)
			return err
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(job)
	})

	mux.HandleFunc("POST /api/v1/pools/{name}/scrub", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		job := jm.Enqueue("scrub", func(j *jobqueue.Job, log func(string)) error {
			out, err := wrappers.RunCombined(r.Context(), 0, "zpool", "scrub", name)
			log(out)
			return err
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(job)
	})

	mux.HandleFunc("DELETE /api/v1/pools/{name}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		var body struct {
			Confirm string `json:"confirm"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.Confirm != name {
			http.Error(w, "confirmation mismatch", http.StatusBadRequest)
			return
		}
		job := jm.Enqueue("destroy-pool", func(j *jobqueue.Job, log func(string)) error {
			out, err := wrappers.RunCombined(r.Context(), 0, "zpool", "destroy", name)
			log(out)
			return err
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(job)
	})

	// properties
	mux.HandleFunc("GET /api/v1/pools/{name}/properties", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		prop := r.URL.Query().Get("prop")
		args := []string{"get", "-H", "-p"}
		if prop != "" {
			args = append(args, prop)
		} else {
			args = append(args, "all")
		}
		args = append(args, name)
		out, err := wrappers.RunCombined(r.Context(), 30, "zpool", args...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(out))
	})

	mux.HandleFunc("PUT /api/v1/pools/{name}/properties", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("name")
		var body struct {
			Property string `json:"property"`
			Value    string `json:"value"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Property == "" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		job := jm.Enqueue("set-pool-property", func(j *jobqueue.Job, log func(string)) error {
			out, err := wrappers.RunCombined(r.Context(), 0, "zpool", "set", body.Property+"="+body.Value, name)
			log(out)
			return err
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(job)
	})

	// scrub schedules CRUD (DB)
	if deps.DB != nil {
		mux.HandleFunc("GET /api/v1/scrub-schedules", func(w http.ResponseWriter, r *http.Request) {
			rows, err := deps.DB.Query(`SELECT id,pool,cron,enabled,created_at FROM scrub_schedules`)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer rows.Close()
			type sched struct {
				ID        int64  `json:"id"`
				Pool      string `json:"pool"`
				Cron      string `json:"cron"`
				Enabled   bool   `json:"enabled"`
				CreatedAt string `json:"created_at"`
			}
			var out []sched
			for rows.Next() {
				var s sched
				var en int
				var created string
				rows.Scan(&s.ID, &s.Pool, &s.Cron, &en, &created)
				s.Enabled = en == 1
				s.CreatedAt = created
				out = append(out, s)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(out)
		})
		mux.HandleFunc("POST /api/v1/scrub-schedules", func(w http.ResponseWriter, r *http.Request) {
			var req scrubScheduleRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Pool == "" || req.Cron == "" {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			en := 0
			if req.Enabled {
				en = 1
			}
			res, err := deps.DB.Exec(`INSERT INTO scrub_schedules (pool,cron,enabled,created_at) VALUES (?,?,?,?)`,
				req.Pool, req.Cron, en, time.Now().UTC().Format(time.RFC3339))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			id, _ := res.LastInsertId()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": id})
		})
		mux.HandleFunc("DELETE /api/v1/scrub-schedules/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := r.PathValue("id")
			_, err := deps.DB.Exec(`DELETE FROM scrub_schedules WHERE id=?`, id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})
	}
}



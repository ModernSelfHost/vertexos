package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ModernSelfHost/vertexos/internal/jobqueue"
	"github.com/ModernSelfHost/vertexos/internal/storage"
	"github.com/ModernSelfHost/vertexos/internal/wrappers"
)

func registerStorageRoutes(mux *http.ServeMux, deps RouterDeps) {
	mux.HandleFunc("GET /api/v1/drives", func(w http.ResponseWriter, r *http.Request) {
		drives, err := storage.ListDrives(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(drives)
	})

	mux.HandleFunc("POST /api/v1/pools/propose", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Drives []string `json:"drives"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)

		all, err := storage.ListDrives(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// filter to requested
		var sizes []uint64
		if len(req.Drives) == 0 {
			for _, d := range all {
				if d.Type == "disk" && !strings.HasPrefix(d.Name, "zram") && !strings.HasPrefix(d.Name, "loop") {
					sizes = append(sizes, d.SizeBytes)
				}
			}
		} else {
			set := map[string]bool{}
			for _, d := range req.Drives {
				set[d] = true
			}
			for _, d := range all {
				if set[d.Device] || set[d.Name] || set[d.StableID] {
					sizes = append(sizes, d.SizeBytes)
				}
			}
		}
		prop := storage.ProposeLayout(sizes)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(prop)
	})

	// Create pool from a proposal: builds vdev args from layout.
	mux.HandleFunc("POST /api/v1/pools/from-proposal", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name   string   `json:"name"`
			Drives []string `json:"drives"`
			Layout string   `json:"layout"` // optional override
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" || len(req.Drives) == 0 {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		job := deps.JobManager.Enqueue("create-pool", func(j *jobqueue.Job, log func(string)) error {
			args := []string{"create", req.Name}
			layout := req.Layout
			if layout == "" {
				sizes := make([]uint64, 0, len(req.Drives))
				all, _ := storage.ListDrives(r.Context())
				set := map[string]bool{}
				for _, d := range req.Drives {
					set[d] = true
				}
				for _, d := range all {
					if set[d.Device] || set[d.Name] || set[d.StableID] {
						sizes = append(sizes, d.SizeBytes)
					}
				}
				p := storage.ProposeLayout(sizes)
				layout = p.Layout
			}
			if layout != "single" {
				args = append(args, layout)
			}
			args = append(args, req.Drives...)
			log(fmt.Sprintf("zpool %s", strings.Join(args, " ")))
			out, err := wrappers.RunCombined(r.Context(), 0, "zpool", args...)
			log(out)
			return err
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(job)
	})
}

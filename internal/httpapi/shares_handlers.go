package httpapi

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ModernSelfHost/vertexos/internal/jobqueue"
	"github.com/ModernSelfHost/vertexos/internal/wrappers"
)

type shareRecord struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	Dataset         string `json:"dataset"`
	QuotaBytes      int64  `json:"quota_bytes"`
	GuestAccessible bool   `json:"guest_accessible"`
	CreatedAt       string `json:"created_at"`
}

func registerShareRoutes(mux *http.ServeMux, deps RouterDeps) {
	if deps.DB == nil {
		return
	}
	db := deps.DB

	mux.HandleFunc("GET /api/v1/shares", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`SELECT id,name,dataset,quota_bytes,guest_accessible,created_at FROM shares ORDER BY id`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var out []shareRecord
		for rows.Next() {
			var s shareRecord
			var guest int
			var created string
			rows.Scan(&s.ID, &s.Name, &s.Dataset, &s.QuotaBytes, &guest, &created)
			s.GuestAccessible = guest == 1
			s.CreatedAt = created
			out = append(out, s)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	})

	mux.HandleFunc("POST /api/v1/shares", func(w http.ResponseWriter, r *http.Request) {
		if deps.JobManager == nil {
			http.Error(w, "job manager not configured", http.StatusServiceUnavailable)
			return
		}
		var req struct {
			Name            string `json:"name"`
			Dataset         string `json:"dataset"`
			QuotaBytes      int64  `json:"quota_bytes"`
			GuestAccessible bool   `json:"guest_accessible"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		dataset := req.Dataset
		if dataset == "" {
			// default to first pool + /name
			pools, _ := wrappers.ListPools(r.Context())
			if len(pools) == 0 {
				http.Error(w, "no pools available", http.StatusBadRequest)
				return
			}
			dataset = pools[0].Name + "/" + req.Name
		}

		// create dataset
		job := deps.JobManager.Enqueue("create-share-dataset", func(j *jobqueue.Job, log func(string)) error {
			out, err := wrappers.RunCombined(r.Context(), 0, "zfs", "create", dataset)
			log(out)
			if err != nil {
				// dataset may already exist; continue
				log(err.Error())
			}
			if req.QuotaBytes > 0 {
				q := fmt.Sprintf("quota=%d", req.QuotaBytes)
				out2, err2 := wrappers.RunCombined(r.Context(), 0, "zfs", "set", q, dataset)
				log(out2)
				if err2 != nil {
					log(err2.Error())
				}
			}
			// store share record
			guest := 0
			if req.GuestAccessible {
				guest = 1
			}
			_, err = db.Exec(`INSERT INTO shares (name,dataset,quota_bytes,guest_accessible,created_at) VALUES (?,?,?,?,?)`,
				req.Name, dataset, req.QuotaBytes, guest, timeNowRFC3339())
			if err != nil {
				return err
			}
			// regenerate samba config
			regenerateSambaConfig(db)
			return nil
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(job)
	})

	mux.HandleFunc("GET /api/v1/shares/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
		var s shareRecord
		var guest int
		var created string
		err := db.QueryRow(`SELECT id,name,dataset,quota_bytes,guest_accessible,created_at FROM shares WHERE id=?`, id).
			Scan(&s.ID, &s.Name, &s.Dataset, &s.QuotaBytes, &guest, &created)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		s.GuestAccessible = guest == 1
		s.CreatedAt = created
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(s)
	})

	mux.HandleFunc("PUT /api/v1/shares/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
		var req struct {
			QuotaBytes      *int64 `json:"quota_bytes"`
			GuestAccessible *bool  `json:"guest_accessible"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if req.QuotaBytes != nil {
			db.Exec(`UPDATE shares SET quota_bytes=? WHERE id=?`, *req.QuotaBytes, id)
			// apply quota
			var dataset string
			db.QueryRow(`SELECT dataset FROM shares WHERE id=?`, id).Scan(&dataset)
			if dataset != "" {
				wrappers.RunCombined(r.Context(), 30*time.Second, "zfs", "set", fmt.Sprintf("quota=%d", *req.QuotaBytes), dataset)
			}
		}
		if req.GuestAccessible != nil {
			g := 0
			if *req.GuestAccessible {
				g = 1
			}
			db.Exec(`UPDATE shares SET guest_accessible=? WHERE id=?`, g, id)
		}
		regenerateSambaConfig(db)
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("DELETE /api/v1/shares/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
		var dataset string
		db.QueryRow(`SELECT dataset FROM shares WHERE id=?`, id).Scan(&dataset)
		// destroy dataset (best effort)
		if dataset != "" {
			wrappers.RunCombined(r.Context(), 30*time.Second, "zfs", "destroy", dataset)
		}
		db.Exec(`DELETE FROM shares WHERE id=?`, id)
		db.Exec(`DELETE FROM share_access WHERE share_id=?`, id)
		regenerateSambaConfig(db)
		w.WriteHeader(http.StatusNoContent)
	})

	// permissions
	mux.HandleFunc("GET /api/v1/shares/{id}/permissions", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
		rows, err := db.Query(`SELECT user_id,role,access FROM share_access WHERE share_id=?`, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		type perm struct {
			UserID *int64 `json:"user_id,omitempty"`
			Role   string `json:"role,omitempty"`
			Access string `json:"access"`
		}
		var out []perm
		for rows.Next() {
			var p perm
			var uid sql.NullInt64
			var role sql.NullString
			var access string
			rows.Scan(&uid, &role, &access)
			if uid.Valid {
				v := uid.Int64
				p.UserID = &v
			}
			if role.Valid {
				p.Role = role.String
			}
			p.Access = access
			out = append(out, p)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	})

	mux.HandleFunc("PUT /api/v1/shares/{id}/permissions", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
		var req struct {
			Entries []struct {
				UserID *int64 `json:"user_id"`
				Role   string `json:"role"`
				Access string `json:"access"`
			} `json:"entries"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		db.Exec(`DELETE FROM share_access WHERE share_id=?`, id)
		for _, e := range req.Entries {
			access := e.Access
			if access == "" {
				access = "read"
			}
			if e.UserID != nil {
				db.Exec(`INSERT INTO share_access (share_id,user_id,role,access) VALUES (?,?,?,?)`, id, *e.UserID, nil, access)
			} else {
				db.Exec(`INSERT INTO share_access (share_id,user_id,role,access) VALUES (?,?,?,?)`, id, nil, e.Role, access)
			}
		}
		regenerateSambaConfig(db)
		w.WriteHeader(http.StatusNoContent)
	})
}

func timeNowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// regenerateSambaConfig writes a managed Samba include file with one section per share.
func regenerateSambaConfig(db *sql.DB) {
	rows, err := db.Query(`SELECT id,name,dataset,guest_accessible FROM shares`)
	if err != nil {
		return
	}
	defer rows.Close()
	var b strings.Builder
	b.WriteString("# Managed by vertexd - do not edit\n")
	for rows.Next() {
		var id int64
		var name, dataset string
		var guest int
		rows.Scan(&id, &name, &dataset, &guest)
		b.WriteString(fmt.Sprintf("\n[%s]\n", name))
		b.WriteString(fmt.Sprintf("   path = /%s\n", dataset)) // placeholder mount path
		b.WriteString("   browsable = yes\n")
		b.WriteString("   read only = no\n")
		if guest == 1 {
			b.WriteString("   guest ok = yes\n")
		} else {
			b.WriteString("   guest ok = no\n")
		}
		// valid users from permissions
		prows, _ := db.Query(`SELECT user_id,role,access FROM share_access WHERE share_id=? AND access IN ('read','write')`, id)
		var users []string
		for prows != nil && prows.Next() {
			var uid sql.NullInt64
			var role sql.NullString
			var access string
			prows.Scan(&uid, &role, &access)
			if uid.Valid {
				// resolve username
				var username string
				db.QueryRow(`SELECT username FROM users WHERE id=?`, uid.Int64).Scan(&username)
				if username != "" {
					users = append(users, username)
				}
			}
		}
		if prows != nil {
			prows.Close()
		}
		if len(users) > 0 {
			b.WriteString("   valid users = " + strings.Join(users, " ") + "\n")
		}
	}
	// write to managed path
	path := "/etc/samba/smb.conf.d/vertex-shares.conf"
	_ = os.MkdirAll("/etc/samba/smb.conf.d", 0755)
	_ = os.WriteFile(path, []byte(b.String()), 0644)
}

var _ = sql.Open
var _ = time.Now

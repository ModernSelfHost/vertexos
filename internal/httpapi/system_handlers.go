package httpapi

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type exportData struct {
	Settings map[string]string      `json:"settings"`
	Users    []map[string]any       `json:"users"`
	Shares   []map[string]any       `json:"shares"`
}

func registerSystemRoutes(mux *http.ServeMux, deps RouterDeps) {
	mux.HandleFunc("GET /api/v1/system/export", func(w http.ResponseWriter, r *http.Request) {
		if deps.DB == nil {
			http.Error(w, "not configured", http.StatusServiceUnavailable)
			return
		}
		data, err := exportState(deps.DB)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=vertex-export.json")
		_ = json.NewEncoder(w).Encode(data)
	})

	mux.HandleFunc("POST /api/v1/system/import", func(w http.ResponseWriter, r *http.Request) {
		if deps.DB == nil {
			http.Error(w, "not configured", http.StatusServiceUnavailable)
			return
		}
		var data exportData
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if err := importState(deps.DB, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

func exportState(db *sql.DB) (exportData, error) {
	data := exportData{Settings: map[string]string{}, Users: []map[string]any{}, Shares: []map[string]any{}}

	rows, err := db.Query(`SELECT key,value FROM settings`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var k, v string
			rows.Scan(&k, &v)
			data.Settings[k] = v
		}
	}

	urows, err := db.Query(`SELECT id,username,password_hash,role,created_at FROM users`)
	if err == nil {
		defer urows.Close()
		for urows.Next() {
			var id int64
			var username, hash, role, created string
			urows.Scan(&id, &username, &hash, &role, &created)
			data.Users = append(data.Users, map[string]any{
				"id": id, "username": username, "password_hash": hash, "role": role, "created_at": created,
			})
		}
	}

	srows, err := db.Query(`SELECT id,name,dataset,quota_bytes,guest_accessible,created_at FROM shares`)
	if err == nil {
		defer srows.Close()
		for srows.Next() {
			var id int64
			var name, dataset, created string
			var quota int64
			var guest int
			srows.Scan(&id, &name, &dataset, &quota, &guest, &created)
			data.Shares = append(data.Shares, map[string]any{
				"id": id, "name": name, "dataset": dataset, "quota_bytes": quota, "guest_accessible": guest, "created_at": created,
			})
		}
	}
	return data, nil
}

func importState(db *sql.DB, data exportData) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for k, v := range data.Settings {
		if _, err := tx.Exec(`INSERT INTO settings(key,value) VALUES(?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`, k, v); err != nil {
			return err
		}
	}
	for _, u := range data.Users {
		_, err := tx.Exec(
			`INSERT INTO users (id,username,password_hash,role,created_at) VALUES (?,?,?,?,?)
			 ON CONFLICT(id) DO UPDATE SET username=excluded.username,password_hash=excluded.password_hash,role=excluded.role`,
			u["id"], u["username"], u["password_hash"], u["role"], u["created_at"],
		)
		if err != nil {
			return err
		}
	}
	for _, s := range data.Shares {
		_, err := tx.Exec(
			`INSERT INTO shares (id,name,dataset,quota_bytes,guest_accessible,created_at) VALUES (?,?,?,?,?,?)
			 ON CONFLICT(id) DO UPDATE SET name=excluded.name,dataset=excluded.dataset,quota_bytes=excluded.quota_bytes,guest_accessible=excluded.guest_accessible`,
			s["id"], s["name"], s["dataset"], s["quota_bytes"], s["guest_accessible"], s["created_at"],
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

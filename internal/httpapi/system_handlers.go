package httpapi

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ModernSelfHost/vertexos/internal/auth"
	"github.com/ModernSelfHost/vertexos/internal/wrappers"
	_ "modernc.org/sqlite"
)

type exportData struct {
	Settings map[string]string `json:"settings"`
	Users    []map[string]any  `json:"users"`
	Shares   []map[string]any  `json:"shares"`
}

func registerSystemRoutes(mux *http.ServeMux, deps RouterDeps) {
	mux.HandleFunc("GET /api/v1/system/export", func(w http.ResponseWriter, r *http.Request) {
		if deps.DB == nil {
			http.Error(w, "not configured", http.StatusServiceUnavailable)
			return
		}
		if r.URL.Query().Get("format") == "json" {
			data, err := exportState(deps.DB)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Content-Disposition", "attachment; filename=vertex-export.json")
			_ = json.NewEncoder(w).Encode(data)
			return
		}
		// zip export
		buf, err := exportZip(deps.DB)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", "attachment; filename=vertex-export.zip")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", buf.Len()))
		_, _ = w.Write(buf.Bytes())
	})

	mux.HandleFunc("POST /api/v1/system/import", func(w http.ResponseWriter, r *http.Request) {
		if deps.DB == nil {
			http.Error(w, "not configured", http.StatusServiceUnavailable)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		// detect zip
		if len(body) >= 4 && bytes.Equal(body[:4], []byte("PK\x03\x04")) {
			if err := importZip(deps.DB, body); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// multipart
		if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			// re-parse from body not possible; read again impossible. Handle via r.Body already consumed -> fallback error
		}
		// json fallback
		var data exportData
		if err := json.Unmarshal(body, &data); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if err := importState(deps.DB, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	// Settings (admin)
	mux.HandleFunc("GET /api/v1/settings", func(w http.ResponseWriter, r *http.Request) {
		u, ok := UserFromContext(r.Context())
		if !ok || u.Role != auth.RoleAdmin {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		rows, err := deps.DB.Query(`SELECT key,value FROM settings`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		out := map[string]string{}
		for rows.Next() {
			var k, v string
			rows.Scan(&k, &v)
			out[k] = v
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	})
	mux.HandleFunc("PUT /api/v1/settings", func(w http.ResponseWriter, r *http.Request) {
		u, ok := UserFromContext(r.Context())
		if !ok || u.Role != auth.RoleAdmin {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		var body struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Key == "" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		_, err := deps.DB.Exec(`INSERT INTO settings(key,value) VALUES(?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`, body.Key, body.Value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	// Logs view (admin)
	mux.HandleFunc("GET /api/v1/logs", func(w http.ResponseWriter, r *http.Request) {
		u, ok := UserFromContext(r.Context())
		if !ok || u.Role != auth.RoleAdmin {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		out, err := wrappers.RunCombined(r.Context(), 30*time.Second, "journalctl", "-n", "200", "--no-pager", "-o", "short")
		if err != nil {
			out, err = wrappers.RunCombined(r.Context(), 30*time.Second, "tail", "-n", "200", "/var/log/syslog")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(out))
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

// exportZip builds a zip containing a SQLite snapshot, managed configs and manifest.
func exportZip(db *sql.DB) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	// SQLite snapshot via VACUUM INTO
	tmp, err := os.CreateTemp("", "vertex-export-*.db")
	if err != nil {
		return nil, err
	}
	tmp.Close()
	os.Remove(tmp.Name())
	if _, err := db.Exec(fmt.Sprintf("VACUUM INTO '%s'", tmp.Name())); err != nil {
		os.Remove(tmp.Name())
		return nil, fmt.Errorf("vacuum: %w", err)
	}
	dbBytes, err := os.ReadFile(tmp.Name())
	os.Remove(tmp.Name())
	if err != nil {
		return nil, err
	}
	fw, err := zw.Create("vertex.db")
	if err != nil {
		return nil, err
	}
	if _, err := fw.Write(dbBytes); err != nil {
		return nil, err
	}

	// managed configs (best effort)
	configFiles := map[string]string{
		"configs/vertex-shares.conf": "/etc/samba/smb.conf.d/vertex-shares.conf",
	}
	for name, path := range configFiles {
		if b, err := os.ReadFile(path); err == nil {
			fw, err := zw.Create(name)
			if err != nil {
				continue
			}
			_, _ = fw.Write(b)
		}
	}

	// manifest
	manifest := map[string]any{
		"created_at": time.Now().UTC().Format(time.RFC3339),
		"format":     "vertex-export-v1",
	}
	mb, _ := json.MarshalIndent(manifest, "", "  ")
	fw, _ = zw.Create("manifest.json")
	if fw != nil {
		_, _ = fw.Write(mb)
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf, nil
}

// importZip validates and restores from a zip produced by exportZip.
func importZip(db *sql.DB, body []byte) error {
	zr, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return fmt.Errorf("invalid zip: %w", err)
	}

	// find vertex.db
	var dbFile *zip.File
	var configFiles = map[string][]byte{}
	var hasManifest bool
	for _, f := range zr.File {
		switch f.Name {
		case "vertex.db":
			dbFile = f
		case "manifest.json":
			hasManifest = true
		default:
			if strings.HasPrefix(f.Name, "configs/") {
				rc, err := f.Open()
				if err == nil {
					b, _ := io.ReadAll(rc)
					rc.Close()
					configFiles[f.Name] = b
				}
			}
		}
	}
	if dbFile == nil {
		return fmt.Errorf("missing vertex.db in archive")
	}
	if !hasManifest {
		return fmt.Errorf("missing manifest.json in archive")
	}

	// extract snapshot to temp and open
	rc, err := dbFile.Open()
	if err != nil {
		return err
	}
	snapBytes, err := io.ReadAll(rc)
	rc.Close()
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp("", "vertex-import-*.db")
	if err != nil {
		return err
	}
	tmp.Close()
	if err := os.WriteFile(tmp.Name(), snapBytes, 0600); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	defer os.Remove(tmp.Name())

	impDB, err := sql.Open("sqlite", tmp.Name())
	if err != nil {
		return err
	}
	defer impDB.Close()

	// validate required tables
	for _, t := range []string{"users", "shares", "settings"} {
		var name string
		err := impDB.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, t).Scan(&name)
		if err != nil {
			return fmt.Errorf("invalid snapshot: missing table %s", t)
		}
	}

	// read data
	data, err := readExportFromDB(impDB)
	if err != nil {
		return err
	}

	if err := importState(db, data); err != nil {
		return err
	}

	// restore configs (best effort)
	for name, b := range configFiles {
		switch name {
		case "configs/vertex-shares.conf":
			_ = os.MkdirAll("/etc/samba/smb.conf.d", 0755)
			_ = os.WriteFile("/etc/samba/smb.conf.d/vertex-shares.conf", b, 0644)
		}
	}
	return nil
}

func readExportFromDB(db *sql.DB) (exportData, error) {
	data := exportData{Settings: map[string]string{}, Users: []map[string]any{}, Shares: []map[string]any{}}
	rows, _ := db.Query(`SELECT key,value FROM settings`)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var k, v string
			rows.Scan(&k, &v)
			data.Settings[k] = v
		}
	}
	urows, _ := db.Query(`SELECT id,username,password_hash,role,created_at FROM users`)
	if urows != nil {
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
	srows, _ := db.Query(`SELECT id,name,dataset,quota_bytes,guest_accessible,created_at FROM shares`)
	if srows != nil {
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

var _ = time.Second

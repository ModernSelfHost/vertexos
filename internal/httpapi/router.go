package httpapi

import (
	"database/sql"
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"

	"github.com/ModernSelfHost/vertexos/internal/jobqueue"
	"github.com/ModernSelfHost/vertexos/internal/sessions"
	"github.com/ModernSelfHost/vertexos/internal/users"
)

// RouterDeps holds dependencies for the HTTP router.
type RouterDeps struct {
	// StaticDir serves the SPA bundle; empty disables static serving.
	StaticDir string
	// StaticFS, if non-nil, is used instead of StaticDir (for embedded assets).
	StaticFS embed.FS
	// UseEmbedded enables serving from StaticFS.
	UseEmbedded bool

	UserStore    *users.Store
	SessionStore *sessions.Store
	JobManager   *jobqueue.Manager
	DB           *sql.DB
}

// NewRouter builds the top-level HTTP mux.
func NewRouter(deps RouterDeps) http.Handler {
	mux := http.NewServeMux()

	registerAlertRoutes(mux, deps)
	registerAuthRoutes(mux, deps)
	registerDashboardRoutes(mux, deps)
	registerJobsRoutes(mux, deps)
	registerPoolRoutes(mux, deps)
	registerStorageRoutes(mux, deps)
	registerSystemRoutes(mux, deps)
	registerWS(mux, deps)

	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
			"service": "vertexd",
		})
	})

	// Static SPA serving.
	if deps.UseEmbedded {
		if sub, err := fs.Sub(deps.StaticFS, "web/dist"); err == nil {
			mux.Handle("/", spaHandler(http.FS(sub)))
		}
	} else if deps.StaticDir != "" {
		mux.Handle("/", spaHandler(http.Dir(deps.StaticDir)))
	}

	return mux
}

// spaHandler serves index.html for unknown paths (client-side routing).
func spaHandler(fsys http.FileSystem) http.Handler {
	fileServer := http.FileServer(fsys)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try the file first.
		if r.URL.Path != "/" {
			f, err := fsys.Open(strings.TrimPrefix(r.URL.Path, "/"))
			if err == nil {
				f.Close()
				fileServer.ServeHTTP(w, r)
				return
			}
		}
		// Fallback to index.html.
		r2 := *r
		r2.URL.Path = "/"
		fileServer.ServeHTTP(w, &r2)
	})
}

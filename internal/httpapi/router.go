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
	registerUserRoutes(mux, deps)
	registerDashboardRoutes(mux, deps)
	registerJobsRoutes(mux, deps)
	registerNFSRoutes(mux, deps)
	registerPoolRoutes(mux, deps)
	registerShareRoutes(mux, deps)
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

	// Wrap with auth middleware for protected paths.
	if deps.UserStore != nil && deps.SessionStore != nil {
		authMw := AuthMiddleware(deps.UserStore, deps.SessionStore)
		orig := mux
		wrapped := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isPublicPath(r.URL.Path) {
				orig.ServeHTTP(w, r)
				return
			}
			authMw(orig).ServeHTTP(w, r)
		})
		return wrapped
	}
	return mux
}

func isPublicPath(p string) bool {
	switch {
	case p == "/api/v1/health":
		return true
	case p == "/api/v1/auth/login":
		return true
	case p == "/ws":
		return true
	case !strings.HasPrefix(p, "/api/"):
		return true // static
	}
	return false
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

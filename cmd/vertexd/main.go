package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/ModernSelfHost/vertexos/internal/cert"
	"github.com/ModernSelfHost/vertexos/internal/config"
	"github.com/ModernSelfHost/vertexos/internal/auth"
	"github.com/ModernSelfHost/vertexos/internal/httpapi"
	"github.com/ModernSelfHost/vertexos/internal/jobqueue"
	"github.com/grandcat/zeroconf"
	"github.com/ModernSelfHost/vertexos/internal/sessions"
	"github.com/ModernSelfHost/vertexos/internal/state"
	"github.com/ModernSelfHost/vertexos/internal/users"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "config error:", err)
		os.Exit(1)
	}

	logger := newLogger(cfg.LogLevel)
	slog.SetDefault(logger)

	logger.Info("vertexd starting", "listen", cfg.ListenAddr, "data_dir", cfg.DataDir)

	if err := os.MkdirAll(cfg.DataDir, 0o700); err != nil {
		logger.Error("create data dir", "err", err)
		os.Exit(1)
	}

	tlsCert, err := cert.Ensure(cfg.TLSCert, cfg.TLSKey, cfg.DataDir)
	if err != nil {
		logger.Error("tls cert", "err", err)
		os.Exit(1)
	}

	db, err := state.Open(filepath.Join(cfg.DataDir, "vertex.db"))
	if err != nil {
		logger.Error("state db", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	userStore := users.NewStore(db)
	sessionStore := sessions.NewStore(db)
	jobManager := jobqueue.NewManager(db)
	_ = jobManager.LoadFromDB()

	// Bootstrap admin if none exist and password provided.
	if pw := os.Getenv("VERTEX_BOOTSTRAP_ADMIN_PASSWORD"); pw != "" {
		list, _ := userStore.List()
		if len(list) == 0 {
			u, err := userStore.Create("admin", pw, auth.RoleAdmin)
			if err != nil {
				logger.Error("bootstrap admin", "err", err)
			} else {
				logger.Info("bootstrapped admin user", "username", u.Username)
			}
		}
	}

	handler := httpapi.NewRouter(httpapi.RouterDeps{
		StaticDir:     staticDir(cfg),
		UserStore:     userStore,
		SessionStore:  sessionStore,
		JobManager:    jobManager,
		DB:            db,
	})

	srv := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		TLSConfig:         tlsConfig(tlsCert),
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			logger.Error("listen", "err", err)
		}
	}()

	// mDNS advertisement + console URL.
	port := 8443
	if _, p, err := net.SplitHostPort(cfg.ListenAddr); err == nil {
		fmt.Sscan(p, &port)
	}
	logger.Info("vertex available", "url", fmt.Sprintf("https://vertex.local:%d", port))
	_, _ = zeroconf.Register("Vertex", "_https._tcp", "local.", port, []string{"path=/"}, nil)

	<-ctx.Done()
	logger.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown", "err", err)
	}
}

func tlsConfig(c tls.Certificate) *tls.Config {
	return &tls.Config{
		Certificates: []tls.Certificate{c},
		MinVersion:   tls.VersionTLS12,
	}
}

func staticDir(cfg config.Config) string {
	if cfg.DevMode {
		return "./web/dist"
	}
	return "/usr/share/vertex/web"
}

func newLogger(level string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	return slog.New(h)
}

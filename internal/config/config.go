package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds vertexd runtime configuration.
type Config struct {
	ListenAddr string `json:"listen_addr"`
	DataDir    string `json:"data_dir"`
	ConfigFile string `json:"-"`
	TLSCert    string `json:"tls_cert"`
	TLSKey     string `json:"tls_key"`
	LogLevel   string `json:"log_level"`
	DevMode    bool   `json:"dev_mode"`
}

// Defaults returns a Config populated with sane defaults.
func Defaults() Config {
	return Config{
		ListenAddr: ":8443",
		DataDir:    "/var/lib/vertex",
		LogLevel:   "info",
	}
}

// Load parses configuration from flags, environment variables and an optional
// JSON config file. Precedence (lowest to highest): defaults, file, env, flags.
func Load() (Config, error) {
	cfg := Defaults()

	fs := flag.NewFlagSet("vertexd", flag.ContinueOnError)
	configPath := fs.String("config", "", "path to JSON config file")
	listen := fs.String("listen", "", "listen address (host:port)")
	dataDir := fs.String("data-dir", "", "data directory")
	tlsCert := fs.String("tls-cert", "", "TLS certificate path")
	tlsKey := fs.String("tls-key", "", "TLS key path")
	logLevel := fs.String("log-level", "", "log level (debug,info,warn,error)")
	devMode := fs.Bool("dev", false, "development mode")

	if err := fs.Parse(os.Args[1:]); err != nil {
		return cfg, err
	}

	// Config file
	cp := *configPath
	if cp == "" {
		cp = os.Getenv("VERTEX_CONFIG")
	}
	if cp == "" {
		cp = filepath.Join("/etc/vertex", "vertexd.json")
	}
	if f, err := os.Open(cp); err == nil {
		defer f.Close()
		dec := json.NewDecoder(f)
		if err := dec.Decode(&cfg); err != nil {
			return cfg, fmt.Errorf("decode config %s: %w", cp, err)
		}
	}

	// Environment overrides
	if v := os.Getenv("VERTEX_LISTEN"); v != "" {
		cfg.ListenAddr = v
	}
	if v := os.Getenv("VERTEX_DATA_DIR"); v != "" {
		cfg.DataDir = v
	}
	if v := os.Getenv("VERTEX_TLS_CERT"); v != "" {
		cfg.TLSCert = v
	}
	if v := os.Getenv("VERTEX_TLS_KEY"); v != "" {
		cfg.TLSKey = v
	}
	if v := os.Getenv("VERTEX_LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}

	// Flag overrides
	if *listen != "" {
		cfg.ListenAddr = *listen
	}
	if *dataDir != "" {
		cfg.DataDir = *dataDir
	}
	if *tlsCert != "" {
		cfg.TLSCert = *tlsCert
	}
	if *tlsKey != "" {
		cfg.TLSKey = *tlsKey
	}
	if *logLevel != "" {
		cfg.LogLevel = *logLevel
	}
	if *devMode {
		cfg.DevMode = true
	}
	cfg.ConfigFile = cp

	if cfg.ListenAddr == "" {
		cfg.ListenAddr = Defaults().ListenAddr
	}
	if cfg.DataDir == "" {
		cfg.DataDir = Defaults().DataDir
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = Defaults().LogLevel
	}

	return cfg, nil
}

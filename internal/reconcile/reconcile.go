package reconcile

import (
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"os"
)

// Reconciler is a component that can reconcile desired vs actual state.
type Reconciler interface {
	Name() string
	Reconcile() error
}

// Runner runs a set of reconcilers on boot and on demand.
type Runner struct {
	reconcilers []Reconciler
}

func NewRunner(rs ...Reconciler) *Runner {
	return &Runner{reconcilers: rs}
}

func (r *Runner) RunAll() {
	for _, rec := range r.reconcilers {
		if err := rec.Reconcile(); err != nil {
			slog.Error("reconcile failed", "name", rec.Name(), "err", err)
		} else {
			slog.Info("reconciled", "name", rec.Name())
		}
	}
}

// ManagedFile tracks a config file vertexd owns, for direct-edit detection.
type ManagedFile struct {
	Path         string
	LastChecksum string
}

// DetectDirectEdits returns paths whose checksum changed since last recorded.
func DetectDirectEdits(files []ManagedFile) []string {
	var out []string
	for _, f := range files {
		b, err := os.ReadFile(f.Path)
		if err != nil {
			continue
		}
		sum := sha256.Sum256(b)
		cur := hex.EncodeToString(sum[:])
		if f.LastChecksum != "" && cur != f.LastChecksum {
			out = append(out, f.Path)
		}
	}
	return out
}

package apps

// Runtime is the seam where a container runtime (Docker Compose, Podman, ...)
// is plugged in. v1 ships a NoopRuntime; post-v1 a real implementation lands here.
type Runtime interface {
	Name() string
	// Install runs an app given its manifest URL and user inputs.
	Install(appID string, inputs map[string]string) error
	// Uninstall removes an app.
	Uninstall(appID string) error
	// Status returns running state.
	Status(appID string) (string, error)
}

// NoopRuntime is the v1 default: app management disabled.
type NoopRuntime struct{}

func (NoopRuntime) Name() string { return "none" }
func (NoopRuntime) Install(string, map[string]string) error {
	return ErrRuntimeNotConfigured
}
func (NoopRuntime) Uninstall(string) error { return ErrRuntimeNotConfigured }
func (NoopRuntime) Status(string) (string, error) { return "not-installed", nil }

var ErrRuntimeNotConfigured = &runtimeError{"container runtime not configured (post-v1 feature)"}

type runtimeError struct{ msg string }

func (e *runtimeError) Error() string { return e.msg }

package wrappers

import (
	"context"
	"time"
)

// Networkd and systemd helpers are stubs for M0/M1.

// ReloadSystemd runs systemctl daemon-reload.
func ReloadSystemd(ctx context.Context) error {
	_, err := RunCombined(ctx, 30*time.Second, "systemctl", "daemon-reload")
	return err
}

// RestartUnit restarts a systemd unit.
func RestartUnit(ctx context.Context, unit string) error {
	_, err := RunCombined(ctx, 60*time.Second, "systemctl", "restart", unit)
	return err
}

// EnableUnit enables a systemd unit.
func EnableUnit(ctx context.Context, unit string) error {
	_, err := RunCombined(ctx, 30*time.Second, "systemctl", "enable", unit)
	return err
}

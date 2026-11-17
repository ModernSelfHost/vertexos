package wrappers

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

// Result holds command output.
type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// Run executes a command with a timeout and captures stdout/stderr separately.
func Run(ctx context.Context, timeout time.Duration, name string, args ...string) (Result, error) {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(cctx, name, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	res := Result{Stdout: outBuf.String(), Stderr: errBuf.String()}
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			res.ExitCode = ee.ExitCode()
		} else {
			res.ExitCode = -1
		}
		return res, err
	}
	return res, nil
}

// RunCombined runs a command capturing combined stdout+stderr.
func RunCombined(ctx context.Context, timeout time.Duration, name string, args ...string) (string, error) {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(cctx, name, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

var _ = fmt.Sprintf

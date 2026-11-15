//go:build unix

package cmd

import (
	"os/exec"
	"syscall"
)

// exitStatus maps a child's exit to the status a shell would report. A child
// killed by a signal has no exit code (ExitCode returns -1, which would surface
// as 255), so follow the shell convention of 128+signal instead.
func exitStatus(err *exec.ExitError) int {
	if ws, ok := err.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
		return 128 + int(ws.Signal())
	}
	return err.ExitCode()
}

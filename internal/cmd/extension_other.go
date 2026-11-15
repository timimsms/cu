//go:build !unix

package cmd

import "os/exec"

// exitStatus returns the child's exit code. Outside Unix there are no signal
// exits to translate.
func exitStatus(err *exec.ExitError) int {
	return err.ExitCode()
}

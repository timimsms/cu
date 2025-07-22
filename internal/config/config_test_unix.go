//go:build !windows
// +build !windows

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitProjectConfig_Unix(t *testing.T) {
	t.Run("write permission error", func(t *testing.T) {
		if os.Getuid() == 0 {
			t.Skip("Running as root, skipping permission test")
		}

		tmpDir := t.TempDir()
		oldWd, _ := os.Getwd()
		require.NoError(t, os.Chdir(tmpDir))
		defer func() { _ = os.Chdir(oldWd) }()

		// Make directory read-only
		require.NoError(t, os.Chmod(tmpDir, 0500))
		defer func() { _ = os.Chmod(tmpDir, 0750) }()

		err := InitProjectConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write project config")
	})
}
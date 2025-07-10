package version

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	t.Run("version constants exist", func(t *testing.T) {
		// These should be non-empty in a real build
		assert.NotNil(t, Version)
		assert.NotNil(t, Commit)
		assert.NotNil(t, Date)
		assert.NotNil(t, BuiltBy)
	})
	
	t.Run("FullVersion returns formatted version", func(t *testing.T) {
		// Save original values
		origVersion := Version
		origCommit := Commit
		origDate := Date
		
		// Set test values
		Version = "1.2.3"
		Commit = "abc123"
		Date = "2024-01-01"
		
		result := FullVersion()
		assert.Contains(t, result, "1.2.3")
		assert.Contains(t, result, "abc123")
		assert.Contains(t, result, "2024-01-01")
		assert.Contains(t, result, runtime.GOOS)
		assert.Contains(t, result, runtime.GOARCH)
		
		// Restore original values
		Version = origVersion
		Commit = origCommit
		Date = origDate
	})
	
	t.Run("FullVersion handles dev version", func(t *testing.T) {
		// Save original values
		origVersion := Version
		
		// Set dev version
		Version = "dev"
		
		result := FullVersion()
		assert.Contains(t, result, "dev")
		
		// Restore original values
		Version = origVersion
	})
}
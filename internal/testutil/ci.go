package testutil

import (
	"os"
	"testing"
)

// IsCI returns true if running in a CI environment
func IsCI() bool {
	// Check common CI environment variables
	return os.Getenv("CI") == "true" ||
		os.Getenv("GITHUB_ACTIONS") == "true" ||
		os.Getenv("JENKINS_HOME") != "" ||
		os.Getenv("TRAVIS") == "true" ||
		os.Getenv("CIRCLECI") == "true"
}

// SkipIfCI skips the test if running in CI environment
func SkipIfCI(t *testing.T, reason string) {
	if IsCI() {
		t.Skipf("CI: %s", reason)
	}
}

// SkipIfNoKeyring skips the test if keyring is not available (common in CI)
func SkipIfNoKeyring(t *testing.T) {
	if IsCI() {
		t.Skip("CI: keyring not available")
	}
}
package cmd

import (
	"testing"
)

func TestMeCommand(t *testing.T) {
	// Test that the command is properly registered
	if meCmd == nil {
		t.Error("meCmd should not be nil")
	}

	if meCmd.Use != "me" {
		t.Errorf("Expected Use to be 'me', got '%s'", meCmd.Use)
	}

	if meCmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

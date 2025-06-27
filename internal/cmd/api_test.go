package cmd

import (
	"testing"
)

func TestAPICommand(t *testing.T) {
	// Just verify the command exists and can be created
	cmd := apiCmd
	if cmd == nil {
		t.Fatal("apiCmd should not be nil")
	}

	// Check flags exist
	methodFlag := cmd.Flag("method")
	if methodFlag == nil {
		t.Error("method flag should exist")
	}

	dataFlag := cmd.Flag("data")
	if dataFlag == nil {
		t.Error("data flag should exist")
	}

	headerFlag := cmd.Flag("header")
	if headerFlag == nil {
		t.Error("header flag should exist")
	}
}
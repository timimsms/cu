package cmd

import (
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBulkCommand_Structure(t *testing.T) {
	// Test main bulk command
	t.Run("bulk command exists", func(t *testing.T) {
		cmd := bulkCmd
		assert.NotNil(t, cmd)
		assert.Equal(t, "bulk", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)

		// Should have subcommands
		assert.NotEmpty(t, cmd.Commands())
	})

	// Test bulk subcommands
	t.Run("bulk subcommands exist", func(t *testing.T) {
		subcommandNames := make(map[string]bool)
		for _, subcmd := range bulkCmd.Commands() {
			// Extract the base command name (before space)
			baseName := strings.Split(subcmd.Use, " ")[0]
			subcommandNames[baseName] = true
		}

		// Check for expected subcommands
		expectedSubcommands := []string{"update", "close", "delete"}
		for _, expected := range expectedSubcommands {
			assert.True(t, subcommandNames[expected], "Expected subcommand '%s' to exist", expected)
		}
	})

	// Test bulk update command uses
	t.Run("bulk update command uses", func(t *testing.T) {
		// Find the update subcommand
		var updateCmd *cobra.Command
		for _, subcmd := range bulkCmd.Commands() {
			if strings.HasPrefix(subcmd.Use, "update") {
				updateCmd = subcmd
				break
			}
		}

		if assert.NotNil(t, updateCmd, "update subcommand should exist") {
			assert.NotEmpty(t, updateCmd.Short)
			assert.NotNil(t, updateCmd.Run)
		}
	})

	// Test bulk close command
	t.Run("bulk close command", func(t *testing.T) {
		// Find the close subcommand
		var closeCmd *cobra.Command
		for _, subcmd := range bulkCmd.Commands() {
			if strings.HasPrefix(subcmd.Use, "close") {
				closeCmd = subcmd
				break
			}
		}

		if assert.NotNil(t, closeCmd, "close subcommand should exist") {
			assert.NotEmpty(t, closeCmd.Short)
			assert.NotNil(t, closeCmd.Run)
		}
	})

	// Test bulk delete command
	t.Run("bulk delete command", func(t *testing.T) {
		// Find the delete subcommand
		var deleteCmd *cobra.Command
		for _, subcmd := range bulkCmd.Commands() {
			if strings.HasPrefix(subcmd.Use, "delete") {
				deleteCmd = subcmd
				break
			}
		}

		if assert.NotNil(t, deleteCmd, "delete subcommand should exist") {
			assert.NotEmpty(t, deleteCmd.Short)
			assert.NotNil(t, deleteCmd.Run)
		}
	})
}

func TestBulkSummaryRecord(t *testing.T) {
	// Progress lines are suppressed outside table output; with outputFormat
	// unset (the zero value) `human` stays quiet, so this exercises the
	// accounting without writing to stdout.
	var s bulkSummary

	s.record("t1", nil)
	s.record("t2", errors.New("boom"))
	s.record("t3", nil)

	assert.Equal(t, 2, s.Succeeded)
	assert.Equal(t, 1, s.Failed)
	require.Len(t, s.Results, 3)

	assert.Equal(t, bulkOutcome{TaskID: "t1", OK: true}, s.Results[0])
	assert.Equal(t, "t2", s.Results[1].TaskID)
	assert.False(t, s.Results[1].OK)
	assert.Equal(t, "boom", s.Results[1].Error, "the failure reason must survive into structured output")
	assert.True(t, s.Results[2].OK)
}

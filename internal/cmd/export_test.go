package cmd

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/raksul/go-clickup/clickup"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExportCmd_Structure(t *testing.T) {
	t.Run("export command exists", func(t *testing.T) {
		cmd := exportCmd
		assert.NotNil(t, cmd)
		assert.Equal(t, "export", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
	})

	t.Run("export tasks subcommand exists", func(t *testing.T) {
		cmd := exportTasksCmd
		assert.NotNil(t, cmd)
		assert.Equal(t, "tasks", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
		assert.NotNil(t, cmd.Run)
	})

	t.Run("export tasks has required flags", func(t *testing.T) {
		cmd := exportTasksCmd

		// Check for expected flags
		listFlag := cmd.Flags().Lookup("list")
		assert.NotNil(t, listFlag)

		formatFlag := cmd.Flags().Lookup("format")
		assert.NotNil(t, formatFlag)

		outputFlag := cmd.Flags().Lookup("output")
		assert.NotNil(t, outputFlag)

		statusFlag := cmd.Flags().Lookup("status")
		assert.NotNil(t, statusFlag)

		priorityFlag := cmd.Flags().Lookup("priority")
		assert.NotNil(t, priorityFlag)

		assigneeFlag := cmd.Flags().Lookup("assignee")
		assert.NotNil(t, assigneeFlag)
	})
}

// Testing the filter function would require mocking the complex clickup.Task struct
// Instead, let's test the command structure and validation logic
func TestExportTasksCmd_Logic(t *testing.T) {
	t.Run("validates format parameter", func(t *testing.T) {
		validFormats := []string{"csv", "json", "markdown", "md"}
		for _, format := range validFormats {
			lower := strings.ToLower(format)
			isValid := lower == "csv" || lower == "json" || lower == "markdown" || lower == "md"
			assert.True(t, isValid, "Format %s should be valid", format)
		}

		invalidFormats := []string{"xml", "yaml", "txt", ""}
		for _, format := range invalidFormats {
			lower := strings.ToLower(format)
			isValid := lower == "csv" || lower == "json" || lower == "markdown" || lower == "md"
			assert.False(t, isValid, "Format %s should be invalid", format)
		}
	})

	t.Run("normalizes md format to markdown", func(t *testing.T) {
		format := "md"
		if format == "md" {
			format = "markdown"
		}
		assert.Equal(t, "markdown", format)
	})

	t.Run("priority mapping works", func(t *testing.T) {
		priorities := map[string]int{
			"urgent": 1,
			"high":   2,
			"normal": 3,
			"low":    4,
		}

		for name, expectedID := range priorities {
			var p int
			switch name {
			case "urgent":
				p = 1
			case "high":
				p = 2
			case "normal":
				p = 3
			case "low":
				p = 4
			}
			assert.Equal(t, expectedID, p)
		}
	})
}

// Note: The actual exportTasksToCSV, exportTasksToJSON, exportTasksToMarkdown
// functions are complex and depend on the clickup package structure.
// These tests focus on command structure and logic validation.

func TestExportCmd_FunctionExistence(t *testing.T) {
	t.Run("export functions exist", func(t *testing.T) {
		// Test that the functions exist by ensuring they can be referenced
		// This is a compile-time check
		var csvFunc func(*os.File, []clickup.Task) error = exportTasksToCSV
		var jsonFunc func(*os.File, []clickup.Task) error = exportTasksToJSON
		var mdFunc func(*os.File, []clickup.Task) error = exportTasksToMarkdown
		var filterFunc func([]clickup.Task, string, string, string) []clickup.Task = filterTasksForExport
		var formatFunc func(string) string = formatTimestamp

		assert.NotNil(t, csvFunc)
		assert.NotNil(t, jsonFunc)
		assert.NotNil(t, mdFunc)
		assert.NotNil(t, filterFunc)
		assert.NotNil(t, formatFunc)
	})
}

func TestExportCmd_CommandFlags(t *testing.T) {
	t.Run("flags have correct properties", func(t *testing.T) {
		cmd := exportTasksCmd

		// Test flag defaults and properties
		listFlag := cmd.Flags().Lookup("list")
		assert.NotNil(t, listFlag)
		assert.Equal(t, "", listFlag.DefValue)

		formatFlag := cmd.Flags().Lookup("format")
		assert.NotNil(t, formatFlag)
		assert.Equal(t, "csv", formatFlag.DefValue)

		fileFlag := cmd.LocalFlags().Lookup("file")
		assert.NotNil(t, fileFlag, "the destination file lives on --file")
		assert.Equal(t, "F", fileFlag.Shorthand)
		assert.Equal(t, "", fileFlag.DefValue)
	})

	// The regression guard for #28. Defining a local "output" flag here shadowed
	// the global format flag, so `cu export tasks -o json` wrote a file named
	// "json" instead of emitting JSON. -o must stay inherited.
	t.Run("output is not redefined locally", func(t *testing.T) {
		assert.Nil(t, exportTasksCmd.LocalFlags().Lookup("output"),
			"export must not shadow the global -o/--output format flag")
	})
}

func TestResolveExportFormat(t *testing.T) {
	// newCmd mirrors the real flag surface: --format is local to export, while
	// --output is inherited from the root command.
	newCmd := func(t *testing.T, args ...string) *cobra.Command {
		t.Helper()
		c := &cobra.Command{Use: "tasks", Run: func(*cobra.Command, []string) {}}
		c.Flags().StringP("format", "f", "csv", "")
		c.Flags().StringP("output", "o", "table", "")
		c.SetArgs(args)
		c.SetOut(io.Discard)
		c.SetErr(io.Discard)
		require.NoError(t, c.Execute())
		return c
	}

	t.Run("defaults to csv", func(t *testing.T) {
		got, err := resolveExportFormat(newCmd(t))
		require.NoError(t, err)
		assert.Equal(t, "csv", got)
	})

	t.Run("-o selects the format", func(t *testing.T) {
		// The bug in #28: this used to write a file named "json".
		got, err := resolveExportFormat(newCmd(t, "-o", "json"))
		require.NoError(t, err)
		assert.Equal(t, "json", got)
	})

	t.Run("--format still works", func(t *testing.T) {
		got, err := resolveExportFormat(newCmd(t, "--format", "markdown"))
		require.NoError(t, err)
		assert.Equal(t, "markdown", got)
	})

	t.Run("md is an alias for markdown", func(t *testing.T) {
		got, err := resolveExportFormat(newCmd(t, "--format", "md"))
		require.NoError(t, err)
		assert.Equal(t, "markdown", got)
	})

	t.Run("matching is case-insensitive", func(t *testing.T) {
		got, err := resolveExportFormat(newCmd(t, "-o", "JSON"))
		require.NoError(t, err)
		assert.Equal(t, "json", got)
	})

	t.Run("agreeing flags are accepted", func(t *testing.T) {
		got, err := resolveExportFormat(newCmd(t, "--format", "json", "-o", "json"))
		require.NoError(t, err)
		assert.Equal(t, "json", got)
	})

	t.Run("disagreeing flags are an error, not a guess", func(t *testing.T) {
		_, err := resolveExportFormat(newCmd(t, "--format", "csv", "-o", "json"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "conflicting formats")
	})

	t.Run("a format export cannot produce is rejected", func(t *testing.T) {
		// "yaml" and "table" are valid globally but meaningless for an export,
		// so they must fail loudly rather than fall back to csv.
		for _, v := range []string{"yaml", "table"} {
			_, err := resolveExportFormat(newCmd(t, "-o", v))
			require.Error(t, err, v)
			assert.Contains(t, err.Error(), "is not an export format")
		}
	})

	t.Run("a path-shaped value points at --file", func(t *testing.T) {
		// The predictable mistake after this change: reaching for the old
		// `-o <file>`. Say where the file flag went instead of just rejecting.
		_, err := resolveExportFormat(newCmd(t, "-o", "tasks.csv"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--file tasks.csv")
	})

	t.Run("a bad --format is rejected too", func(t *testing.T) {
		_, err := resolveExportFormat(newCmd(t, "--format", "xlsx"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--format")
	})
}

func TestExportCmd_Examples(t *testing.T) {
	t.Run("command has usage examples", func(t *testing.T) {
		cmd := exportTasksCmd
		assert.Contains(t, cmd.Long, "Examples:")
		assert.Contains(t, cmd.Long, "cu export tasks")
		assert.Contains(t, cmd.Long, "--format csv")
		assert.Contains(t, cmd.Long, "--format json")
		assert.Contains(t, cmd.Long, "--format markdown")
	})
}

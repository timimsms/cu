package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/raksul/go-clickup/clickup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/timimsms/cu/internal/output"
)

func TestListCommands_Structure(t *testing.T) {
	// Test main list command
	t.Run("list command exists", func(t *testing.T) {
		cmd := listCmd
		assert.NotNil(t, cmd)
		assert.Equal(t, "list", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)

		// Should have subcommands
		assert.NotEmpty(t, cmd.Commands())
	})

	// Test list lists command
	t.Run("list lists command", func(t *testing.T) {
		cmd := listListCmd
		assert.NotNil(t, cmd)
		assert.Equal(t, "list", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotNil(t, cmd.Run)
	})

	// Test list default command
	t.Run("list default command", func(t *testing.T) {
		cmd := listDefaultCmd
		assert.NotNil(t, cmd)
		assert.Equal(t, "default <list-id>", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotNil(t, cmd.Run)
	})
}

// Regression test for #24: table output must contain a row per list.
func TestListTableRows(t *testing.T) {
	lists := []clickup.List{
		{ID: "901", Name: "Backlog", TaskCount: json.Number("12")},
		{ID: "902", Name: "Sprint", Archived: true},
	}

	rows := listTableRows(lists, "902")
	require.Len(t, rows, 2)
	assert.Equal(t, listRow{ID: "901", Name: "Backlog", Tasks: 12}, rows[0])
	assert.Equal(t, listRow{ID: "902", Name: "Sprint", Default: "*", Archived: true}, rows[1])

	var buf bytes.Buffer
	require.NoError(t, (&output.TableFormatter{Writer: &buf}).Format(rows))

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	require.Len(t, lines, 4, "header, separator and one line per list")
	assert.Equal(t, []string{"id", "name", "default", "tasks", "archived"}, strings.Fields(lines[0]))
	assert.Equal(t, []string{"901", "Backlog", "12", "false"}, strings.Fields(lines[2]))
	assert.Equal(t, []string{"902", "Sprint", "*", "0", "true"}, strings.Fields(lines[3]))
}

func TestListTableRows_Empty(t *testing.T) {
	assert.Empty(t, listTableRows(nil, ""))
}

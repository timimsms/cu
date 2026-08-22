package api

import (
	"encoding/json"
	"testing"

	"github.com/raksul/go-clickup/clickup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// listWithStatuses builds a clickup.List from JSON because the SDK models
// Statuses as an anonymous struct slice, which cannot be written as a literal.
func listWithStatuses(t *testing.T, name, statusesJSON string) *clickup.List {
	t.Helper()
	var list clickup.List
	require.NoError(t, json.Unmarshal([]byte(`{"name":"`+name+`","statuses":`+statusesJSON+`}`), &list))
	return &list
}

func TestClosedStatus(t *testing.T) {
	t.Run("custom status set resolves by type, not name", func(t *testing.T) {
		// The DireLabs status set: no status is called "complete".
		list := listWithStatuses(t, "R&D Projects", `[
			{"status":"backlog","orderindex":0,"type":"open"},
			{"status":"in development","orderindex":3,"type":"custom"},
			{"status":"shipped","orderindex":7,"type":"done"},
			{"status":"cancelled","orderindex":8,"type":"closed"}
		]`)

		got, err := ClosedStatus(list)
		require.NoError(t, err)
		assert.Equal(t, "shipped", got, "done must win over closed")
	})

	t.Run("default status set", func(t *testing.T) {
		list := listWithStatuses(t, "Default", `[
			{"status":"to do","orderindex":0,"type":"open"},
			{"status":"complete","orderindex":1,"type":"closed"}
		]`)

		got, err := ClosedStatus(list)
		require.NoError(t, err)
		assert.Equal(t, "complete", got)
	})

	t.Run("lowest orderindex wins among several done statuses", func(t *testing.T) {
		list := listWithStatuses(t, "Multi", `[
			{"status":"open","orderindex":0,"type":"open"},
			{"status":"released","orderindex":9,"type":"done"},
			{"status":"shipped","orderindex":5,"type":"done"}
		]`)

		got, err := ClosedStatus(list)
		require.NoError(t, err)
		assert.Equal(t, "shipped", got)
	})

	t.Run("no done status is an error naming the available set", func(t *testing.T) {
		list := listWithStatuses(t, "Odd", `[
			{"status":"to do","orderindex":0,"type":"open"},
			{"status":"doing","orderindex":1,"type":"custom"}
		]`)

		_, err := ClosedStatus(list)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "to do")
		assert.Contains(t, err.Error(), "doing")
	})
}

func TestOpenStatus(t *testing.T) {
	t.Run("first open status in board order", func(t *testing.T) {
		list := listWithStatuses(t, "R&D Projects", `[
			{"status":"triage","orderindex":2,"type":"open"},
			{"status":"backlog","orderindex":0,"type":"open"},
			{"status":"shipped","orderindex":7,"type":"done"}
		]`)

		got, err := OpenStatus(list)
		require.NoError(t, err)
		assert.Equal(t, "backlog", got)
	})

	t.Run("no open status is an error", func(t *testing.T) {
		list := listWithStatuses(t, "Closed only", `[
			{"status":"done","orderindex":0,"type":"done"}
		]`)

		_, err := OpenStatus(list)
		assert.Error(t, err)
	})
}

package api

import (
	"errors"
	"testing"

	"github.com/raksul/go-clickup/clickup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pages builds a fetch func returning the given page sizes, recording which
// pages were actually requested.
func pages(sizes ...int) (func(int) ([]clickup.Task, error), *[]int) {
	var requested []int
	return func(page int) ([]clickup.Task, error) {
		requested = append(requested, page)
		if page >= len(sizes) {
			return nil, nil
		}
		return make([]clickup.Task, sizes[page]), nil
	}, &requested
}

func TestPaginate(t *testing.T) {
	t.Run("stops on the first short page", func(t *testing.T) {
		fetch, requested := pages(tasksPerPage, tasksPerPage, 7)

		res, err := paginate(0, fetch)
		require.NoError(t, err)

		assert.Len(t, res.Tasks, tasksPerPage*2+7)
		assert.Equal(t, 3, res.Pages)
		assert.False(t, res.Truncated)
		assert.Equal(t, []int{0, 1, 2}, *requested, "pages are requested in order, and no further")
	})

	t.Run("a single short page is one request", func(t *testing.T) {
		fetch, requested := pages(5)

		res, err := paginate(0, fetch)
		require.NoError(t, err)

		assert.Len(t, res.Tasks, 5)
		assert.Equal(t, []int{0}, *requested)
	})

	t.Run("an empty first page terminates", func(t *testing.T) {
		fetch, requested := pages(0)

		res, err := paginate(0, fetch)
		require.NoError(t, err)

		assert.Empty(t, res.Tasks)
		assert.Equal(t, 1, res.Pages)
		assert.False(t, res.Truncated)
		assert.Equal(t, []int{0}, *requested)
	})

	t.Run("an exactly-full last page costs one extra request", func(t *testing.T) {
		// A full page is indistinguishable from "more to come", so the next
		// page must be requested to learn the set is exhausted.
		fetch, requested := pages(tasksPerPage, 0)

		res, err := paginate(0, fetch)
		require.NoError(t, err)

		assert.Len(t, res.Tasks, tasksPerPage)
		assert.False(t, res.Truncated)
		assert.Equal(t, []int{0, 1}, *requested)
	})

	t.Run("hitting the cap reports truncation instead of pretending completeness", func(t *testing.T) {
		fetch, requested := pages(tasksPerPage, tasksPerPage, tasksPerPage, tasksPerPage)

		res, err := paginate(2, fetch)
		require.NoError(t, err)

		assert.Len(t, res.Tasks, tasksPerPage*2)
		assert.Equal(t, 2, res.Pages)
		assert.True(t, res.Truncated, "callers must be able to tell a capped result from a complete one")
		assert.Equal(t, []int{0, 1}, *requested, "the cap is respected exactly")
	})

	t.Run("errors abort rather than returning a partial page set", func(t *testing.T) {
		boom := errors.New("boom")
		fetch := func(page int) ([]clickup.Task, error) {
			if page == 1 {
				return nil, boom
			}
			return make([]clickup.Task, tasksPerPage), nil
		}

		res, err := paginate(0, fetch)
		require.ErrorIs(t, err, boom)
		assert.Nil(t, res, "a partial result must not be mistaken for a complete one")
	})
}

func TestTeamTaskQueryToSDK(t *testing.T) {
	t.Run("empty filters are omitted", func(t *testing.T) {
		opts := (&TeamTaskQuery{}).toSDK(3)

		assert.Equal(t, 3, opts.Page)
		assert.Nil(t, opts.Statuses)
		assert.Nil(t, opts.Assignees)
		assert.Nil(t, opts.Tags)
		assert.Nil(t, opts.CustomFields)
		assert.False(t, opts.IncludeClosed)
		assert.False(t, opts.Subtasks)
	})

	t.Run("filters are passed through", func(t *testing.T) {
		q := &TeamTaskQuery{
			Statuses:      []string{"in review"},
			Assignees:     []string{"123"},
			Tags:          []string{"unreleased"},
			IncludeClosed: true,
			Subtasks:      true,
			OrderBy:       "updated",
			Reverse:       true,
			CustomFields: clickup.CustomFieldsInGetTasksRequest{
				{FieldId: "f1", Operator: clickup.Equals, Value: []string{"v"}},
			},
		}

		opts := q.toSDK(0)

		assert.Equal(t, []string{"in review"}, opts.Statuses)
		assert.Equal(t, []string{"123"}, opts.Assignees)
		assert.Equal(t, []string{"unreleased"}, opts.Tags)
		assert.True(t, opts.IncludeClosed)
		assert.True(t, opts.Subtasks)
		assert.Equal(t, "updated", opts.OrderBy)
		assert.True(t, opts.Reverse)
		require.Len(t, opts.CustomFields, 1)
		assert.Equal(t, "f1", opts.CustomFields[0].FieldId)
	})
}

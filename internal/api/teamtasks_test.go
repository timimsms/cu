package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
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

func TestTeamTaskQueryValues(t *testing.T) {
	t.Run("empty filters are omitted", func(t *testing.T) {
		v, err := (&TeamTaskQuery{}).values(3)
		require.NoError(t, err)

		assert.Equal(t, url.Values{"page": {"3"}}, v)
	})

	t.Run("filters are passed through", func(t *testing.T) {
		q := &TeamTaskQuery{
			Statuses:      []string{"in review", "open"},
			Assignees:     []string{"123"},
			Tags:          []string{"unreleased"},
			SpaceIDs:      []string{"s1"},
			ProjectIDs:    []string{"f1"},
			ListIDs:       []string{"l1", "l2"},
			IncludeClosed: true,
			Subtasks:      true,
			OrderBy:       "updated",
			Reverse:       true,
		}

		v, err := q.values(0)
		require.NoError(t, err)

		assert.Equal(t, "0", v.Get("page"), "page 0 is sent explicitly")
		assert.Equal(t, []string{"in review", "open"}, v["statuses[]"])
		assert.Equal(t, []string{"123"}, v["assignees[]"])
		assert.Equal(t, []string{"unreleased"}, v["tags[]"])
		assert.Equal(t, []string{"s1"}, v["space_ids[]"])
		assert.Equal(t, []string{"f1"}, v["project_ids[]"])
		assert.Equal(t, []string{"l1", "l2"}, v["list_ids[]"])
		assert.Equal(t, "true", v.Get("include_closed"))
		assert.Equal(t, "true", v.Get("subtasks"))
		assert.Equal(t, "updated", v.Get("order_by"))
		assert.Equal(t, "true", v.Get("reverse"))
		assert.Empty(t, v.Get("archived"))
	})

	t.Run("custom field values are JSON-escaped", func(t *testing.T) {
		q := &TeamTaskQuery{CustomFields: clickup.CustomFieldsInGetTasksRequest{
			{FieldId: "f1", Operator: clickup.Equals, Value: []string{`say "hi" \ bye`}},
			{FieldId: "f2", Operator: clickup.Any, Value: []string{"a", `b"c`}},
			{FieldId: "f3", Operator: clickup.IsNull},
		}}

		v, err := q.values(0)
		require.NoError(t, err)

		raw := v.Get("custom_fields")
		var decoded []map[string]any
		require.NoError(t, json.Unmarshal([]byte(raw), &decoded), "must be valid JSON: %s", raw)
		require.Len(t, decoded, 3)
		assert.Equal(t, `say "hi" \ bye`, decoded[0]["value"])
		assert.Equal(t, "=", decoded[0]["operator"])
		assert.Equal(t, []any{"a", `b"c`}, decoded[1]["value"])
		assert.NotContains(t, decoded[2], "value")
	})
}

// pagedServer serves page sizes in order for path, recording every raw query.
func pagedServer(t *testing.T, path string, sizes ...int) (*httptest.Server, *[]string) {
	t.Helper()
	var queries []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path {
			http.NotFound(w, r)
			return
		}
		queries = append(queries, r.URL.RawQuery)

		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		n := 0
		if page < len(sizes) {
			n = sizes[page]
		}
		tasks := make([]clickup.Task, n)
		for i := range tasks {
			tasks[i].ID = fmt.Sprintf("p%d-%d", page, i)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(clickup.GetTasksResponse{Tasks: tasks})
	}))
	t.Cleanup(srv.Close)
	return srv, &queries
}

func TestSearchTeamTasksOverHTTP(t *testing.T) {
	srv, queries := pagedServer(t, "/team/T1/task", 100, 100, 3)
	c := newTestClient(t, srv)

	res, err := c.SearchTeamTasks(context.Background(), "T1", &TeamTaskQuery{
		SpaceIDs: []string{"S9"},
		Statuses: []string{"in progress"},
		CustomFields: clickup.CustomFieldsInGetTasksRequest{
			{FieldId: "f1", Operator: clickup.Equals, Value: []string{`a"b\c`}},
		},
	})
	require.NoError(t, err)

	assert.Len(t, res.Tasks, 203)
	assert.Equal(t, 3, res.Pages)
	assert.False(t, res.Truncated)
	require.Len(t, *queries, 3)

	wantCF := url.QueryEscape(`[{"field_id":"f1","operator":"=","value":"a\"b\\c"}]`)
	for i, raw := range *queries {
		// url.Values.Encode sorts keys, so the raw string is deterministic.
		want := "custom_fields=" + wantCF +
			"&page=" + strconv.Itoa(i) +
			"&space_ids%5B%5D=S9" +
			"&statuses%5B%5D=in+progress"
		assert.Equal(t, want, raw, "page %d", i)
	}
}

func TestSearchTeamTasksReportsTruncation(t *testing.T) {
	srv, queries := pagedServer(t, "/team/T1/task", 100, 100, 100)
	c := newTestClient(t, srv)

	res, err := c.SearchTeamTasks(context.Background(), "T1", &TeamTaskQuery{MaxPages: 2})
	require.NoError(t, err)

	assert.Len(t, res.Tasks, 200)
	assert.True(t, res.Truncated)
	assert.Len(t, *queries, 2)
}

func TestListTasksAllPagesOverHTTP(t *testing.T) {
	srv, queries := pagedServer(t, "/list/L1/task", 100, 100, 3)
	c := newTestClient(t, srv)

	res, err := c.ListTasksAllPages(context.Background(), "L1", &TaskQueryOptions{
		Statuses: []string{"open"},
	}, 0)
	require.NoError(t, err)

	assert.Len(t, res.Tasks, 203)
	assert.False(t, res.Truncated)
	// The SDK omits page=0, so the first request carries only the filter.
	assert.Equal(t, []string{
		"statuses%5B%5D=open",
		"page=1&statuses%5B%5D=open",
		"page=2&statuses%5B%5D=open",
	}, *queries)
}

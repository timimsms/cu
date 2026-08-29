package api

import (
	"context"
	"fmt"

	"github.com/raksul/go-clickup/clickup"
)

// tasksPerPage is the page size ClickUp uses for paged task endpoints. It is
// fixed server-side, so a short page means the last page.
const tasksPerPage = 100

// DefaultMaxTaskPages bounds pagination so a mistaken filter cannot walk an
// entire workspace forever under the rate limiter. Callers that genuinely want
// everything can raise it.
const DefaultMaxTaskPages = 50

// TeamTaskQuery filters a workspace-wide task search. Zero values mean "no
// filter", matching the API.
type TeamTaskQuery struct {
	Statuses      []string
	Assignees     []string
	Tags          []string
	IncludeClosed bool
	Subtasks      bool
	Archived      bool
	OrderBy       string
	Reverse       bool

	// CustomFields filters server-side on custom field values — the only way
	// to answer "which task has Repo = X" without walking every list.
	CustomFields clickup.CustomFieldsInGetTasksRequest

	// MaxPages bounds pagination; zero means DefaultMaxTaskPages.
	MaxPages int
}

func (q *TeamTaskQuery) toSDK(page int) *clickup.GetTasksOptions {
	opts := &clickup.GetTasksOptions{
		Page:          page,
		IncludeClosed: q.IncludeClosed,
		Subtasks:      q.Subtasks,
		Archived:      q.Archived,
		OrderBy:       q.OrderBy,
		Reverse:       q.Reverse,
	}
	if len(q.Statuses) > 0 {
		opts.Statuses = q.Statuses
	}
	if len(q.Assignees) > 0 {
		opts.Assignees = q.Assignees
	}
	if len(q.Tags) > 0 {
		opts.Tags = q.Tags
	}
	if len(q.CustomFields) > 0 {
		opts.CustomFields = q.CustomFields
	}
	return opts
}

// TaskPageResult reports what a paged fetch actually covered, so callers can
// tell "no more results" from "we stopped early".
type TaskPageResult struct {
	Tasks     []clickup.Task
	Pages     int
	Truncated bool // hit MaxPages with a full page still coming back
}

// paginate walks pages until a short page (the last one) or maxPages, whichever
// comes first. Kept separate from the API calls so the termination and
// truncation rules are testable without a network round trip.
func paginate(maxPages int, fetch func(page int) ([]clickup.Task, error)) (*TaskPageResult, error) {
	if maxPages <= 0 {
		maxPages = DefaultMaxTaskPages
	}

	result := &TaskPageResult{}
	for page := 0; page < maxPages; page++ {
		tasks, err := fetch(page)
		if err != nil {
			return nil, err
		}

		result.Tasks = append(result.Tasks, tasks...)
		result.Pages = page + 1

		// A short page is the last page.
		if len(tasks) < tasksPerPage {
			return result, nil
		}
	}

	// Stopped at the cap with a full page still arriving — the caller must be
	// able to say so rather than present a partial set as complete.
	result.Truncated = true
	return result, nil
}

// SearchTeamTasks queries the workspace-wide Get Filtered Team Tasks endpoint,
// paging until exhaustion.
//
// This replaces walking workspace → spaces → folders → lists and calling
// GetTasks per list: that fan-out costs one request per list under a 100
// req/min limit, and it only ever fetched page 0, so any list past 100 tasks
// was silently cut off.
func (c *Client) SearchTeamTasks(ctx context.Context, teamID string, query *TeamTaskQuery) (*TaskPageResult, error) {
	if query == nil {
		query = &TeamTaskQuery{}
	}
	return paginate(query.MaxPages, func(page int) ([]clickup.Task, error) {
		if err := c.rateLimiter.Wait(ctx); err != nil {
			return nil, err
		}
		tasks, _, err := c.client.Tasks.GetFilteredTeamTasks(ctx, teamID, query.toSDK(page))
		if err != nil {
			return nil, c.handleError(err)
		}
		return tasks, nil
	})
}

// ListTasksAllPages pages a single list to exhaustion. GetTasks fetches only
// the page it is asked for, so callers that want a whole list need this.
func (c *Client) ListTasksAllPages(ctx context.Context, listID string, options *TaskQueryOptions, maxPages int) (*TaskPageResult, error) {
	if options == nil {
		options = &TaskQueryOptions{}
	}
	return paginate(maxPages, func(page int) ([]clickup.Task, error) {
		opts := *options
		opts.Page = page
		return c.GetTasks(ctx, listID, &opts)
	})
}

// FindTasksByCustomField returns the tasks in a workspace whose custom field
// equals value — the reverse lookup from an external identifier back to a task.
func (c *Client) FindTasksByCustomField(ctx context.Context, teamID, fieldID, value string) ([]clickup.Task, error) {
	if fieldID == "" {
		return nil, fmt.Errorf("a custom field id is required")
	}

	query := &TeamTaskQuery{
		IncludeClosed: true, // a match must not depend on the task being open
		CustomFields: clickup.CustomFieldsInGetTasksRequest{
			{
				FieldId:  fieldID,
				Operator: clickup.Equals,
				Value:    []string{value},
			},
		},
	}

	result, err := c.SearchTeamTasks(ctx, teamID, query)
	if err != nil {
		return nil, err
	}
	return result.Tasks, nil
}

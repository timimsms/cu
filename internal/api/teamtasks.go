package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/raksul/go-clickup/clickup"
)

// tasksPerPage is the page size ClickUp uses for paged task endpoints. It is
// fixed server-side, so a short page means the last page.
const tasksPerPage = 100

// DefaultMaxTaskPages bounds pagination so a mistaken filter cannot walk an
// entire workspace forever under the rate limiter. Callers that genuinely want
// everything can set TeamTaskQuery.MaxPages.
const DefaultMaxTaskPages = 50

// TeamTaskQuery filters a workspace-wide task search. Zero values mean "no
// filter", matching the API.
type TeamTaskQuery struct {
	Statuses  []string
	Assignees []string // user ids; the endpoint does not accept usernames
	Tags      []string

	// Location filters. Applying these server-side matters: filtering after
	// the fact runs into MaxPages first, so a narrow space in a large
	// workspace could come back empty and merely "truncated".
	SpaceIDs   []string
	ProjectIDs []string // folder ids — the API still calls folders projects
	ListIDs    []string

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

// values encodes the query for GET /team/{team_id}/task.
//
// This is built by hand rather than through clickup.GetTasksOptions for two
// reasons: the SDK has no space_ids[]/project_ids[]/list_ids[], and its
// custom_fields encoder concatenates JSON strings without escaping, so a value
// containing a quote or backslash produced a malformed filter.
func (q *TeamTaskQuery) values(page int) (url.Values, error) {
	v := url.Values{}
	v.Set("page", strconv.Itoa(page))

	setBool := func(key string, b bool) {
		if b {
			v.Set(key, "true")
		}
	}
	setBool("include_closed", q.IncludeClosed)
	setBool("subtasks", q.Subtasks)
	setBool("archived", q.Archived)
	setBool("reverse", q.Reverse)
	if q.OrderBy != "" {
		v.Set("order_by", q.OrderBy)
	}

	for key, vals := range map[string][]string{
		"statuses[]":    q.Statuses,
		"assignees[]":   q.Assignees,
		"tags[]":        q.Tags,
		"space_ids[]":   q.SpaceIDs,
		"project_ids[]": q.ProjectIDs,
		"list_ids[]":    q.ListIDs,
	} {
		for _, val := range vals {
			v.Add(key, val)
		}
	}

	if len(q.CustomFields) > 0 {
		cf, err := encodeCustomFieldFilter(q.CustomFields)
		if err != nil {
			return nil, err
		}
		v.Set("custom_fields", cf)
	}
	return v, nil
}

// customFieldFilter is the JSON shape the endpoint expects for each
// custom_fields entry. Value is a bare string for one value and an array for
// several, matching what the API documents.
type customFieldFilter struct {
	FieldID  string `json:"field_id"`
	Operator string `json:"operator"`
	Value    any    `json:"value,omitempty"`
}

func encodeCustomFieldFilter(fields clickup.CustomFieldsInGetTasksRequest) (string, error) {
	out := make([]customFieldFilter, 0, len(fields))
	for _, f := range fields {
		entry := customFieldFilter{FieldID: f.FieldId, Operator: f.Operator.String()}
		switch len(f.Value) {
		case 0:
		case 1:
			entry.Value = f.Value[0]
		default:
			entry.Value = f.Value
		}
		out = append(out, entry)
	}
	b, err := json.Marshal(out)
	if err != nil {
		return "", err
	}
	return string(b), nil
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
		v, err := query.values(page)
		if err != nil {
			return nil, err
		}
		req, err := c.client.NewRequest(http.MethodGet, fmt.Sprintf("team/%s/task?%s", url.PathEscape(teamID), v.Encode()), nil)
		if err != nil {
			return nil, err
		}
		var resp clickup.GetTasksResponse
		if _, err := c.client.Do(ctx, req, &resp); err != nil {
			return nil, c.handleError(err)
		}
		return resp.Tasks, nil
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
// The page result is returned whole so callers can report truncation.
func (c *Client) FindTasksByCustomField(ctx context.Context, teamID, fieldID, value string) (*TaskPageResult, error) {
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

	return c.SearchTeamTasks(ctx, teamID, query)
}

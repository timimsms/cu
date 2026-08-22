package api

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/raksul/go-clickup/clickup"
)

// GetList returns a single list, including its status set.
func (c *Client) GetList(ctx context.Context, listID string) (*clickup.List, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	list, _, err := c.client.Lists.GetList(ctx, listID)
	if err != nil {
		return nil, c.handleError(err)
	}
	return &list, nil
}

// statusName is one entry of a list's status set, flattened out of the SDK's
// anonymous struct so it can be passed around and sorted.
type statusName struct {
	Name       string
	Type       string
	OrderIndex float64
}

func listStatuses(list *clickup.List) []statusName {
	out := make([]statusName, 0, len(list.Statuses))
	for _, s := range list.Statuses {
		idx, _ := s.Orderindex.Float64()
		out = append(out, statusName{Name: s.Status, Type: s.Type, OrderIndex: idx})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].OrderIndex < out[j].OrderIndex })
	return out
}

func statusesOfType(list *clickup.List, types ...string) []statusName {
	var out []statusName
	for _, s := range listStatuses(list) {
		for _, t := range types {
			if strings.EqualFold(s.Type, t) {
				out = append(out, s)
				break
			}
		}
	}
	return out
}

func statusNames(list *clickup.List) string {
	all := listStatuses(list)
	names := make([]string, 0, len(all))
	for _, s := range all {
		names = append(names, fmt.Sprintf("%s (%s)", s.Name, s.Type))
	}
	return strings.Join(names, ", ")
}

// ClosedStatus returns the status a list uses to mean "done".
//
// ClickUp lists define their own status sets — "complete" only exists on lists
// that happen to use the default set, so it cannot be assumed. A list's done
// status is identified by its type ("done", or "closed"), not its name.
func ClosedStatus(list *clickup.List) (string, error) {
	// Prefer "done"; "closed" covers sets that only mark a terminal state.
	if s := statusesOfType(list, "done"); len(s) > 0 {
		return s[0].Name, nil
	}
	if s := statusesOfType(list, "closed"); len(s) > 0 {
		return s[0].Name, nil
	}
	return "", fmt.Errorf("list %q has no status of type done or closed (statuses: %s)", list.Name, statusNames(list))
}

// OpenStatus returns the status a list uses to mean "not started" — the first
// status of type "open" in board order.
func OpenStatus(list *clickup.List) (string, error) {
	if s := statusesOfType(list, "open"); len(s) > 0 {
		return s[0].Name, nil
	}
	return "", fmt.Errorf("list %q has no status of type open (statuses: %s)", list.Name, statusNames(list))
}

// StatusResolver resolves the closed/open status for a task's own list,
// caching per list so bulk operations across one list make a single call.
type StatusResolver struct {
	client *Client
	lists  map[string]*clickup.List
}

func NewStatusResolver(client *Client) *StatusResolver {
	return &StatusResolver{client: client, lists: map[string]*clickup.List{}}
}

func (r *StatusResolver) listForTask(ctx context.Context, taskID string) (*clickup.List, error) {
	task, err := r.client.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task.List.ID == "" {
		return nil, fmt.Errorf("could not determine the list for task %s", taskID)
	}
	if l, ok := r.lists[task.List.ID]; ok {
		return l, nil
	}

	list, err := r.client.GetList(ctx, task.List.ID)
	if err != nil {
		return nil, err
	}
	r.lists[task.List.ID] = list
	return list, nil
}

// ClosedStatusForTask returns the done status of the list the task lives in.
func (r *StatusResolver) ClosedStatusForTask(ctx context.Context, taskID string) (string, error) {
	list, err := r.listForTask(ctx, taskID)
	if err != nil {
		return "", err
	}
	return ClosedStatus(list)
}

// OpenStatusForTask returns the open status of the list the task lives in.
func (r *StatusResolver) OpenStatusForTask(ctx context.Context, taskID string) (string, error) {
	list, err := r.listForTask(ctx, taskID)
	if err != nil {
		return "", err
	}
	return OpenStatus(list)
}

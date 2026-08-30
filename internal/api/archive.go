package api

import (
	"context"
	"fmt"

	"github.com/raksul/go-clickup/clickup"
)

// archiveRequest is sent directly rather than through clickup.TaskUpdateRequest,
// which has no archived field — the SDK cannot express this update.
type archiveRequest struct {
	Archived bool `json:"archived"`
}

// SetTaskArchived archives or unarchives a task.
//
// Archiving keeps the task, its history and its URL while removing it from
// active views. That makes it the right primitive for retiring a task whose
// work has moved elsewhere: closing it would assert a status the task never
// reached (and "closed" is not even a status name on every list), and deleting
// it destroys the record.
func (c *Client) SetTaskArchived(ctx context.Context, taskID string, archived bool) (*clickup.Task, error) {
	if taskID == "" {
		return nil, fmt.Errorf("a task id is required")
	}
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	req, err := c.client.NewRequest("PUT", fmt.Sprintf("task/%s", taskID), &archiveRequest{Archived: archived})
	if err != nil {
		return nil, err
	}

	task := new(clickup.Task)
	if _, err := c.client.Do(ctx, req, task); err != nil {
		return nil, c.handleError(err)
	}
	return task, nil
}

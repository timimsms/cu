package api

import (
	"context"
	"net/http"
	"time"

	"github.com/raksul/go-clickup/clickup"
	"github.com/tim/cu/internal/auth"
	"github.com/tim/cu/internal/errors"
)

// Client wraps the ClickUp API client
type Client struct {
	client      *clickup.Client
	rateLimiter *RateLimiter
}

// NewClient creates a new API client
func NewClient() (*Client, error) {
	authMgr := auth.NewManager()
	token, err := authMgr.GetCurrentToken()
	if err != nil {
		return nil, errors.ErrNotAuthenticated
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &retryTransport{
			base: http.DefaultTransport,
		},
	}

	client := clickup.NewClient(httpClient, token.Value)
	
	return &Client{
		client:      client,
		rateLimiter: NewRateLimiter(100, time.Minute), // 100 requests per minute for free tier
	}, nil
}

// GetWorkspaces returns all workspaces the user has access to
func (c *Client) GetWorkspaces(ctx context.Context) ([]clickup.Team, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	teams, _, err := c.client.Teams.GetTeams(ctx)
	if err != nil {
		return nil, c.handleError(err)
	}

	return teams, nil
}

// GetSpaces returns all spaces in a workspace
func (c *Client) GetSpaces(ctx context.Context, workspaceID string) ([]clickup.Space, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	spaces, _, err := c.client.Spaces.GetSpaces(ctx, workspaceID, false)
	if err != nil {
		return nil, c.handleError(err)
	}

	return spaces, nil
}

// GetFolders returns all folders in a space
func (c *Client) GetFolders(ctx context.Context, spaceID string) ([]clickup.Folder, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	folders, _, err := c.client.Folders.GetFolders(ctx, spaceID, false)
	if err != nil {
		return nil, c.handleError(err)
	}

	return folders, nil
}

// GetLists returns all lists in a folder or space
func (c *Client) GetLists(ctx context.Context, folderID string) ([]clickup.List, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	lists, _, err := c.client.Lists.GetLists(ctx, folderID, false)
	if err != nil {
		return nil, c.handleError(err)
	}

	return lists, nil
}

// GetFolderlessLists returns lists directly in a space (not in folders)
func (c *Client) GetFolderlessLists(ctx context.Context, spaceID string) ([]clickup.List, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	lists, _, err := c.client.Lists.GetFolderlessLists(ctx, spaceID, false)
	if err != nil {
		return nil, c.handleError(err)
	}

	return lists, nil
}

// GetTask returns a single task
func (c *Client) GetTask(ctx context.Context, taskID string) (*clickup.Task, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	task, _, err := c.client.Tasks.GetTask(ctx, taskID, &clickup.GetTaskOptions{})
	if err != nil {
		return nil, c.handleError(err)
	}

	return task, nil
}

// GetTasks returns tasks based on query options
func (c *Client) GetTasks(ctx context.Context, listID string, options *TaskQueryOptions) ([]clickup.Task, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	opts := &clickup.GetTasksOptions{
		Page: options.Page,
	}

	if options.Assignees != nil {
		opts.Assignees = options.Assignees
	}
	if options.Statuses != nil {
		opts.Statuses = options.Statuses
	}
	if options.Tags != nil {
		opts.Tags = options.Tags
	}

	tasks, _, err := c.client.Tasks.GetTasks(ctx, listID, opts)
	if err != nil {
		return nil, c.handleError(err)
	}

	return tasks, nil
}

// CreateTask creates a new task
func (c *Client) CreateTask(ctx context.Context, listID string, request *clickup.TaskRequest) (*clickup.Task, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	task, _, err := c.client.Tasks.CreateTask(ctx, listID, request)
	if err != nil {
		return nil, c.handleError(err)
	}

	return task, nil
}

// UpdateTask updates an existing task
func (c *Client) UpdateTask(ctx context.Context, taskID string, opts *clickup.GetTaskOptions, request *clickup.TaskUpdateRequest) (*clickup.Task, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	task, _, err := c.client.Tasks.UpdateTask(ctx, taskID, opts, request)
	if err != nil {
		return nil, c.handleError(err)
	}

	return task, nil
}

// DeleteTask deletes a task
func (c *Client) DeleteTask(ctx context.Context, taskID string) error {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return err
	}

	_, err := c.client.Tasks.DeleteTask(ctx, taskID, &clickup.GetTaskOptions{})
	if err != nil {
		return c.handleError(err)
	}

	return nil
}

// GetCurrentUser returns the authenticated user
func (c *Client) GetCurrentUser(ctx context.Context) (*clickup.User, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	user, _, err := c.client.Authorization.GetAuthorizedUser(ctx)
	if err != nil {
		return nil, c.handleError(err)
	}

	return user, nil
}

// handleError converts API errors to user-friendly errors
func (c *Client) handleError(err error) error {
	if err == nil {
		return nil
	}

	// TODO: Parse HTTP response codes and convert to appropriate errors
	// For now, return the error as-is
	return err
}

// TaskQueryOptions represents options for querying tasks
type TaskQueryOptions struct {
	Page      int
	Assignees []string
	Statuses  []string
	Tags      []string
	Priority  *int
	DueDate   *time.Time
}
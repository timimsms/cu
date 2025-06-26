package api

import (
	"context"
	"fmt"
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

// TaskCreateOptions represents options for creating a task
type TaskCreateOptions struct {
	Name        string
	Description string
	Assignees   []string
	Status      string
	Priority    string
	Tags        []string
	DueDate     string
}

// TaskUpdateOptions represents options for updating a task
type TaskUpdateOptions struct {
	Name            string
	Description     string
	Status          string
	Priority        string
	Tags            []string
	DueDate         string
	AddAssignees    []string
	RemoveAssignees []string
}

// HasUpdates checks if any updates are specified
func (o *TaskUpdateOptions) HasUpdates() bool {
	return o.Name != "" || o.Description != "" || o.Status != "" || o.Priority != "" ||
		len(o.Tags) > 0 || o.DueDate != "" || len(o.AddAssignees) > 0 || len(o.RemoveAssignees) > 0
}

// CreateTask creates a new task with simplified options
func (c *Client) CreateTask(ctx context.Context, listID string, options *TaskCreateOptions) (*clickup.Task, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	// Build the task request
	request := &clickup.TaskRequest{
		Name:        options.Name,
		Description: options.Description,
		Tags:        options.Tags,
	}

	// TODO: Implement user lookup by username to set assignees
	// if len(options.Assignees) > 0 {
	//     request.Assignees = convertUsernamesToIDs(options.Assignees)
	// }

	// Set status if provided
	if options.Status != "" {
		request.Status = options.Status
	}

	// Set priority if provided
	if options.Priority != "" {
		// Convert priority string to int based on ClickUp's scale
		var priorityInt int
		switch options.Priority {
		case "urgent":
			priorityInt = 1
		case "high":
			priorityInt = 2
		case "normal":
			priorityInt = 3
		case "low":
			priorityInt = 4
		default:
			priorityInt = 3 // Default to normal
		}
		request.Priority = priorityInt
	}

	// Set due date if provided
	if options.DueDate != "" {
		// Parse due date
		t, err := parseDueDate(options.DueDate)
		if err == nil {
			request.DueDate = clickup.NewDate(t)
		}
		// If parsing fails, just skip setting the due date
	}

	task, _, err := c.client.Tasks.CreateTask(ctx, listID, request)
	if err != nil {
		return nil, c.handleError(err)
	}

	return task, nil
}

// parseDueDate parses various date formats including relative dates
func parseDueDate(input string) (time.Time, error) {
	now := time.Now()

	// Handle relative dates
	switch input {
	case "today":
		return time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location()), nil
	case "tomorrow":
		return now.AddDate(0, 0, 1), nil
	case "week":
		return now.AddDate(0, 0, 7), nil
	}

	// Try parsing as RFC3339
	if t, err := time.Parse(time.RFC3339, input); err == nil {
		return t, nil
	}

	// Try parsing as date only
	if t, err := time.Parse("2006-01-02", input); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", input)
}

// UpdateTask updates an existing task with simplified options
func (c *Client) UpdateTask(ctx context.Context, taskID string, options *TaskUpdateOptions) (*clickup.Task, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	// Build the update request
	request := &clickup.TaskUpdateRequest{}

	// Set simple string fields
	if options.Name != "" {
		request.Name = options.Name
	}
	if options.Description != "" {
		request.Description = options.Description
	}
	if options.Status != "" {
		request.Status = options.Status
	}

	// Set priority if provided
	if options.Priority != "" {
		// Convert priority string to int based on ClickUp's scale
		var priorityInt int
		switch options.Priority {
		case "urgent":
			priorityInt = 1
		case "high":
			priorityInt = 2
		case "normal":
			priorityInt = 3
		case "low":
			priorityInt = 4
		default:
			priorityInt = 3 // Default to normal
		}
		request.Priority = priorityInt
	}

	// Set tags - this replaces all tags
	if len(options.Tags) > 0 {
		request.Tags = options.Tags
	}

	// Set due date if provided
	if options.DueDate != "" {
		t, err := parseDueDate(options.DueDate)
		if err == nil {
			request.DueDate = clickup.NewDate(t)
		}
	}

	// TODO: Implement user lookup by username to handle assignees
	// if len(options.AddAssignees) > 0 || len(options.RemoveAssignees) > 0 {
	//     request.Assignees = clickup.TaskAssigneeUpdateRequest{
	//         Add: convertUsernamesToIDs(options.AddAssignees),
	//         Rem: convertUsernamesToIDs(options.RemoveAssignees),
	//     }
	// }

	task, _, err := c.client.Tasks.UpdateTask(ctx, taskID, &clickup.GetTaskOptions{}, request)
	if err != nil {
		return nil, c.handleError(err)
	}

	return task, nil
}

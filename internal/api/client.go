package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/raksul/go-clickup/clickup"
	"github.com/tim/cu/internal/auth"
	"github.com/tim/cu/internal/errors"
	"github.com/tim/cu/internal/interfaces"
)

// Client wraps the ClickUp API client
type Client struct {
	client      *clickup.Client
	rateLimiter *RateLimiter
	userLookup  *UserLookup
	authManager interface{ GetCurrentToken() (*auth.Token, error) }
}

// NewClient creates a new API client with the provided auth manager
func NewClient(authManager interface{ GetCurrentToken() (*auth.Token, error) }) *Client {
	return &Client{
		authManager: authManager,
		rateLimiter: NewRateLimiter(100, time.Minute), // 100 requests per minute for free tier
	}
}

// Connect initializes the API connection using the current token
func (c *Client) Connect() error {
	token, err := c.authManager.GetCurrentToken()
	if err != nil {
		return errors.ErrNotAuthenticated
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &retryTransport{
			base: http.DefaultTransport,
		},
	}

	c.client = clickup.NewClient(httpClient, token.Value)
	c.userLookup = NewUserLookup(c)

	return nil
}

// UserLookup returns the user lookup service
func (c *Client) UserLookup() *UserLookup {
	return c.userLookup
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

// GetSpace returns a single space
func (c *Client) GetSpace(ctx context.Context, spaceID string) (*clickup.Space, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	space, _, err := c.client.Spaces.GetSpace(ctx, spaceID)
	if err != nil {
		return nil, c.handleError(err)
	}

	return space, nil
}

// CreateSpace creates a new space in a workspace
func (c *Client) CreateSpace(ctx context.Context, teamID string, request *clickup.SpaceRequest) (*clickup.Space, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	// Convert teamID from string to int (required by the ClickUp API)
	teamIDInt, err := strconv.Atoi(teamID)
	if err != nil {
		return nil, fmt.Errorf("invalid team ID: %w", err)
	}

	space, _, err := c.client.Spaces.CreateSpace(ctx, teamIDInt, request)
	if err != nil {
		return nil, c.handleError(err)
	}

	return space, nil
}

// UpdateSpace updates a space
func (c *Client) UpdateSpace(ctx context.Context, spaceID string, request *clickup.SpaceRequest) (*clickup.Space, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	// Convert spaceID from string to int (required by the ClickUp API)
	spaceIDInt, err := strconv.Atoi(spaceID)
	if err != nil {
		return nil, fmt.Errorf("invalid space ID: %w", err)
	}

	space, _, err := c.client.Spaces.UpdateSpace(ctx, spaceIDInt, request)
	if err != nil {
		return nil, c.handleError(err)
	}

	return space, nil
}

// DeleteSpace deletes a space
func (c *Client) DeleteSpace(ctx context.Context, spaceID string) error {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return err
	}

	// Convert spaceID from string to int (required by the ClickUp API)
	spaceIDInt, err := strconv.Atoi(spaceID)
	if err != nil {
		return fmt.Errorf("invalid space ID: %w", err)
	}

	_, err = c.client.Spaces.DeleteSpace(ctx, spaceIDInt)
	if err != nil {
		return c.handleError(err)
	}

	return nil
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

// CreateFolder creates a new folder in a space
func (c *Client) CreateFolder(ctx context.Context, spaceID string, request *clickup.FolderRequest) (*clickup.Folder, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	// Convert spaceID from string to int (required by the ClickUp API)
	spaceIDInt, err := strconv.Atoi(spaceID)
	if err != nil {
		return nil, fmt.Errorf("invalid space ID: %w", err)
	}

	folder, _, err := c.client.Folders.CreateFolder(ctx, spaceIDInt, request)
	if err != nil {
		return nil, c.handleError(err)
	}

	return folder, nil
}

// GetFolder returns a single folder
func (c *Client) GetFolder(ctx context.Context, folderID string) (*clickup.Folder, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	folder, _, err := c.client.Folders.GetFolder(ctx, folderID)
	if err != nil {
		return nil, c.handleError(err)
	}

	return folder, nil
}

// UpdateFolder updates a folder
func (c *Client) UpdateFolder(ctx context.Context, folderID string, request *clickup.FolderRequest) (*clickup.Folder, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	// Convert folderID from string to int (required by the ClickUp API)
	folderIDInt, err := strconv.Atoi(folderID)
	if err != nil {
		return nil, fmt.Errorf("invalid folder ID: %w", err)
	}

	folder, _, err := c.client.Folders.UpdateFolder(ctx, folderIDInt, request)
	if err != nil {
		return nil, c.handleError(err)
	}

	return folder, nil
}

// DeleteFolder deletes a folder
func (c *Client) DeleteFolder(ctx context.Context, folderID string) error {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return err
	}

	// Convert folderID from string to int (required by the ClickUp API)
	folderIDInt, err := strconv.Atoi(folderID)
	if err != nil {
		return fmt.Errorf("invalid folder ID: %w", err)
	}

	_, err = c.client.Folders.DeleteFolder(ctx, folderIDInt)
	if err != nil {
		return c.handleError(err)
	}

	return nil
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

// CreateList creates a new list in a folder
func (c *Client) CreateList(ctx context.Context, folderID string, request *clickup.ListRequest) (*clickup.List, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	list, _, err := c.client.Lists.CreateList(ctx, folderID, request)
	if err != nil {
		return nil, c.handleError(err)
	}

	return &list, nil
}

// GetList returns a single list
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

// CreateFolderlessList creates a list directly in a space (not in a folder)
func (c *Client) CreateFolderlessList(ctx context.Context, spaceID string, request *clickup.ListRequest) (*clickup.List, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	// Convert spaceID from string to int (required by the ClickUp API)
	spaceIDInt, err := strconv.Atoi(spaceID)
	if err != nil {
		return nil, fmt.Errorf("invalid space ID: %w", err)
	}

	list, _, err := c.client.Lists.CreateFolderlessList(ctx, spaceIDInt, request)
	if err != nil {
		return nil, c.handleError(err)
	}

	return &list, nil
}

// UpdateList updates a list
func (c *Client) UpdateList(ctx context.Context, listID string, request *clickup.ListRequest) (*clickup.List, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	list, _, err := c.client.Lists.UpdateList(ctx, listID, request)
	if err != nil {
		return nil, c.handleError(err)
	}

	return &list, nil
}

// DeleteList deletes a list
func (c *Client) DeleteList(ctx context.Context, listID string) error {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return err
	}

	_, err := c.client.Lists.DeleteList(ctx, listID)
	if err != nil {
		return c.handleError(err)
	}

	return nil
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
func (c *Client) GetTasks(ctx context.Context, listID string, options *interfaces.TaskQueryOptions) ([]clickup.Task, error) {
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

// GetAuthorizedUser returns the authenticated user (alias for GetCurrentUser)
func (c *Client) GetAuthorizedUser(ctx context.Context) (*clickup.User, error) {
	return c.GetCurrentUser(ctx)
}

// GetAuthorizedTeams returns the teams the user has access to
func (c *Client) GetAuthorizedTeams(ctx context.Context) ([]clickup.Team, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	teams, _, err := c.client.Teams.GetTeams(ctx)
	if err != nil {
		return nil, c.handleError(err)
	}

	return teams, nil
}

// GetWorkspaceMembers returns all members of a workspace
func (c *Client) GetWorkspaceMembers(ctx context.Context, workspaceID string) ([]clickup.TeamUser, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	// The ClickUp API doesn't have a direct endpoint for workspace members
	// We need to get teams first, then get members from teams
	teams, _, err := c.client.Teams.GetTeams(ctx)
	if err != nil {
		return nil, c.handleError(err)
	}

	// Find the team with matching ID
	var targetTeam *clickup.Team
	for _, team := range teams {
		if team.ID == workspaceID {
			targetTeam = &team
			break
		}
	}

	if targetTeam == nil {
		return nil, fmt.Errorf("workspace not found: %s", workspaceID)
	}

	// Extract users from team members
	users := make([]clickup.TeamUser, 0, len(targetTeam.Members))
	for _, member := range targetTeam.Members {
		users = append(users, member.User)
	}

	return users, nil
}

// GetMembers returns all members of a list
func (c *Client) GetMembers(ctx context.Context, listID string) ([]clickup.Member, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	members, _, err := c.client.Members.GetListMembers(ctx, listID)
	if err != nil {
		return nil, c.handleError(err)
	}

	return members, nil
}

// View operations

// GetViews returns all views for a list
func (c *Client) GetViews(ctx context.Context, listID string) ([]clickup.View, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	views, _, err := c.client.Views.GetViewsOf(ctx, clickup.ListView, listID)
	if err != nil {
		return nil, c.handleError(err)
	}

	return views, nil
}

// GetView returns a single view
func (c *Client) GetView(ctx context.Context, viewID string) (*clickup.View, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	view, _, err := c.client.Views.GetView(ctx, viewID)
	if err != nil {
		return nil, c.handleError(err)
	}

	return view, nil
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
func (c *Client) CreateTask(ctx context.Context, listID string, options *interfaces.TaskCreateOptions) (*clickup.Task, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	// Build the task request
	request := &clickup.TaskRequest{
		Name:        options.Name,
		Description: options.Description,
		Tags:        options.Tags,
	}

	// Handle assignees
	if len(options.Assignees) > 0 {
		// First, ensure we have loaded users
		workspaces, err := c.GetWorkspaces(ctx)
		if err == nil && len(workspaces) > 0 {
			_ = c.userLookup.LoadWorkspaceUsers(ctx, workspaces[0].ID)
		}

		// Convert usernames to IDs
		ids, err := c.userLookup.ConvertUsernamesToIDs(options.Assignees)
		if err != nil {
			// Log error but continue - don't fail task creation due to assignee issues
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		} else {
			request.Assignees = ids
		}
	}

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
func (c *Client) UpdateTask(ctx context.Context, taskID string, options *interfaces.TaskUpdateOptions) (*clickup.Task, error) {
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

	// Handle assignees
	if len(options.AddAssignees) > 0 || len(options.RemoveAssignees) > 0 {
		// First, ensure we have loaded users
		workspaces, err := c.GetWorkspaces(ctx)
		if err == nil && len(workspaces) > 0 {
			_ = c.userLookup.LoadWorkspaceUsers(ctx, workspaces[0].ID)
		}

		assigneeUpdate := clickup.TaskAssigneeUpdateRequest{}

		// Convert add assignees
		if len(options.AddAssignees) > 0 {
			addIDs, err := c.userLookup.ConvertUsernamesToIDs(options.AddAssignees)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
			} else {
				assigneeUpdate.Add = addIDs
			}
		}

		// Convert remove assignees
		if len(options.RemoveAssignees) > 0 {
			remIDs, err := c.userLookup.ConvertUsernamesToIDs(options.RemoveAssignees)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
			} else {
				assigneeUpdate.Rem = remIDs
			}
		}

		// Only set if we have something to update
		if len(assigneeUpdate.Add) > 0 || len(assigneeUpdate.Rem) > 0 {
			request.Assignees = assigneeUpdate
		}
	}

	task, _, err := c.client.Tasks.UpdateTask(ctx, taskID, &clickup.GetTaskOptions{}, request)
	if err != nil {
		return nil, c.handleError(err)
	}

	return task, nil
}

// Comment-related methods

// GetTaskComments retrieves all comments for a task
func (c *Client) GetTaskComments(ctx context.Context, taskID string) ([]clickup.Comment, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	comments, _, err := c.client.Comments.GetTaskComments(ctx, taskID, nil)
	if err != nil {
		return nil, c.handleError(err)
	}

	return comments, nil
}

// CreateTaskComment creates a new comment on a task
func (c *Client) CreateTaskComment(ctx context.Context, taskID string, text string, assignee string, notifyAll bool) (*clickup.CreateCommentResponse, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	request := &clickup.CommentRequest{
		CommentText: text,
		NotifyAll:   notifyAll,
	}

	// Handle assignee if specified
	if assignee != "" {
		// First, ensure we have loaded users
		workspaces, err := c.GetWorkspaces(ctx)
		if err == nil && len(workspaces) > 0 {
			_ = c.userLookup.LoadWorkspaceUsers(ctx, workspaces[0].ID)
		}

		// Convert username to ID
		userIDs, err := c.userLookup.ConvertUsernamesToIDs([]string{assignee})
		if err != nil || len(userIDs) == 0 {
			return nil, fmt.Errorf("invalid assignee: %w", err)
		}
		request.Assignee = userIDs[0]
	}

	response, _, err := c.client.Comments.CreateTaskComment(ctx, taskID, nil, request)
	if err != nil {
		return nil, c.handleError(err)
	}

	return response, nil
}

// UpdateTaskComment updates an existing comment
func (c *Client) UpdateTaskComment(ctx context.Context, commentID string, text string, resolved bool) error {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return err
	}

	// Convert comment ID to int
	var commentIDInt int
	if _, err := fmt.Sscanf(commentID, "%d", &commentIDInt); err != nil {
		return fmt.Errorf("invalid comment ID format: %w", err)
	}

	request := &clickup.UpdateCommentRequest{
		CommentText: text,
		Resolved:    resolved,
	}

	_, err := c.client.Comments.UpdateComment(ctx, commentIDInt, request)
	if err != nil {
		return c.handleError(err)
	}

	return nil
}

// DeleteTaskComment deletes a comment
func (c *Client) DeleteTaskComment(ctx context.Context, commentID string) error {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return err
	}

	// Convert comment ID to int
	var commentIDInt int
	if _, err := fmt.Sscanf(commentID, "%d", &commentIDInt); err != nil {
		return fmt.Errorf("invalid comment ID format: %w", err)
	}

	_, err := c.client.Comments.DeleteComment(ctx, commentIDInt)
	if err != nil {
		return c.handleError(err)
	}

	return nil
}

// Custom field operations

// GetCustomFields returns custom fields for a list
func (c *Client) GetCustomFields(ctx context.Context, listID string) ([]clickup.CustomField, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	fields, _, err := c.client.CustomFields.GetAccessibleCustomFields(ctx, listID)
	if err != nil {
		return nil, c.handleError(err)
	}

	return fields, nil
}

// SetCustomFieldValue sets a custom field value for a task
func (c *Client) SetCustomFieldValue(ctx context.Context, taskID string, fieldID string, value map[string]interface{}) error {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return err
	}

	_, err := c.client.CustomFields.SetCustomFieldValue(ctx, taskID, fieldID, value, nil)
	if err != nil {
		return c.handleError(err)
	}

	return nil
}

// Goal-related methods

// GetGoals returns all goals for a team
func (c *Client) GetGoals(ctx context.Context, teamID string, includeCompleted bool) ([]clickup.Goal, []clickup.GoalFolder, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, nil, err
	}

	goals, folders, _, err := c.client.Goals.GetGoals(ctx, teamID, includeCompleted)
	if err != nil {
		return nil, nil, c.handleError(err)
	}

	return goals, folders, nil
}

// CreateGoal creates a new goal
func (c *Client) CreateGoal(ctx context.Context, teamID string, request *clickup.CreateGoalRequest) (*clickup.Goal, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	// Convert teamID from string to int (required by the ClickUp API)
	teamIDInt, err := strconv.Atoi(teamID)
	if err != nil {
		return nil, fmt.Errorf("invalid team ID: %w", err)
	}

	goal, _, err := c.client.Goals.CreateGoal(ctx, teamIDInt, request)
	if err != nil {
		return nil, c.handleError(err)
	}

	return goal, nil
}

// GetGoal returns a single goal
func (c *Client) GetGoal(ctx context.Context, goalID string) (*clickup.Goal, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	goal, _, err := c.client.Goals.GetGoal(ctx, goalID)
	if err != nil {
		return nil, c.handleError(err)
	}

	return goal, nil
}

// UpdateGoal updates a goal
func (c *Client) UpdateGoal(ctx context.Context, goalID string, request *clickup.UpdateGoalRequest) (*clickup.Goal, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	goal, _, err := c.client.Goals.UpdateGoal(ctx, goalID, request)
	if err != nil {
		return nil, c.handleError(err)
	}

	return goal, nil
}

// DeleteGoal deletes a goal
func (c *Client) DeleteGoal(ctx context.Context, goalID string) error {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return err
	}

	_, err := c.client.Goals.DeleteGoal(ctx, goalID)
	if err != nil {
		return c.handleError(err)
	}

	return nil
}

// Webhook-related methods

// GetWebhooks returns all webhooks for a team
func (c *Client) GetWebhooks(ctx context.Context, teamID string) ([]clickup.Webhook, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	// Convert teamID from string to int (required by the ClickUp API)
	teamIDInt, err := strconv.Atoi(teamID)
	if err != nil {
		return nil, fmt.Errorf("invalid team ID: %w", err)
	}

	webhooks, _, err := c.client.Webhooks.GetWebhook(ctx, teamIDInt)
	if err != nil {
		return nil, c.handleError(err)
	}

	return webhooks, nil
}

// CreateWebhook creates a new webhook
func (c *Client) CreateWebhook(ctx context.Context, teamID string, request *clickup.WebhookRequest) (*clickup.Webhook, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	// Convert teamID from string to int (required by the ClickUp API)
	teamIDInt, err := strconv.Atoi(teamID)
	if err != nil {
		return nil, fmt.Errorf("invalid team ID: %w", err)
	}

	webhookResp, _, err := c.client.Webhooks.CreateWebhook(ctx, teamIDInt, request)
	if err != nil {
		return nil, c.handleError(err)
	}

	// Return the webhook from the response
	return &webhookResp.Webhook, nil
}

// UpdateWebhook updates a webhook
func (c *Client) UpdateWebhook(ctx context.Context, webhookID string, request *clickup.WebhookRequest) (*clickup.Webhook, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	webhookResp, _, err := c.client.Webhooks.UpdateWebhook(ctx, webhookID, request)
	if err != nil {
		return nil, c.handleError(err)
	}

	// Return the webhook from the response
	return &webhookResp.Webhook, nil
}

// DeleteWebhook deletes a webhook
func (c *Client) DeleteWebhook(ctx context.Context, webhookID string) error {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return err
	}

	_, err := c.client.Webhooks.DeleteWebhook(ctx, webhookID)
	if err != nil {
		return c.handleError(err)
	}

	return nil
}

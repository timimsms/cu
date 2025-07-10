// Package mock provides mock implementations for API testing
package mock

import (
	"context"
	"fmt"
	"sync"

	"github.com/raksul/go-clickup/clickup"
	"github.com/tim/cu/internal/api"
)

// Client is a mock implementation of the API client for testing
type Client struct {
	mu sync.RWMutex

	// User methods
	currentUser      *clickup.User
	currentUserErr   error
	workspaces       []clickup.Team
	workspacesErr    error
	workspaceMembers map[string][]clickup.TeamUser
	membersErr       error

	// Hierarchy methods
	spaces           map[string][]clickup.Space
	spacesErr        error
	folders          map[string][]clickup.Folder
	foldersErr       error
	lists            map[string][]clickup.List
	listsErr         error
	folderlessLists  map[string][]clickup.List
	folderlessErr    error

	// Task methods
	tasks            map[string]*clickup.Task
	taskLists        map[string][]clickup.Task
	tasksErr         error
	createTaskResult *clickup.Task
	createTaskErr    error
	updateTaskResult *clickup.Task
	updateTaskErr    error
	deleteTaskErr    error

	// Comment methods
	comments         map[string][]clickup.Comment
	commentsErr      error
	createCommentRes *clickup.CreateCommentResponse
	createCommentErr error
	updateCommentErr error
	deleteCommentErr error

	// UserLookup
	userLookup *UserLookup

	// Call tracking
	calls []string
}

// NewClient creates a new mock API client
func NewClient() *Client {
	return &Client{
		workspaceMembers: make(map[string][]clickup.TeamUser),
		spaces:           make(map[string][]clickup.Space),
		folders:          make(map[string][]clickup.Folder),
		lists:            make(map[string][]clickup.List),
		folderlessLists:  make(map[string][]clickup.List),
		tasks:            make(map[string]*clickup.Task),
		taskLists:        make(map[string][]clickup.Task),
		comments:         make(map[string][]clickup.Comment),
		userLookup:       NewUserLookup(),
		calls:            []string{},
	}
}

// GetCurrentUser returns the mocked current user
func (c *Client) GetCurrentUser(ctx context.Context) (*clickup.User, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = append(c.calls, "GetCurrentUser")
	return c.currentUser, c.currentUserErr
}

// GetWorkspaces returns the mocked workspaces
func (c *Client) GetWorkspaces(ctx context.Context) ([]clickup.Team, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = append(c.calls, "GetWorkspaces")
	return c.workspaces, c.workspacesErr
}

// GetWorkspaceMembers returns the mocked workspace members
func (c *Client) GetWorkspaceMembers(ctx context.Context, workspaceID string) ([]clickup.TeamUser, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = append(c.calls, fmt.Sprintf("GetWorkspaceMembers(%s)", workspaceID))
	
	if c.membersErr != nil {
		return nil, c.membersErr
	}
	
	members, ok := c.workspaceMembers[workspaceID]
	if !ok {
		return []clickup.TeamUser{}, nil
	}
	return members, nil
}

// GetSpaces returns the mocked spaces
func (c *Client) GetSpaces(ctx context.Context, workspaceID string) ([]clickup.Space, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = append(c.calls, fmt.Sprintf("GetSpaces(%s)", workspaceID))
	
	if c.spacesErr != nil {
		return nil, c.spacesErr
	}
	
	spaces, ok := c.spaces[workspaceID]
	if !ok {
		return []clickup.Space{}, nil
	}
	return spaces, nil
}

// GetFolders returns the mocked folders
func (c *Client) GetFolders(ctx context.Context, spaceID string) ([]clickup.Folder, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = append(c.calls, fmt.Sprintf("GetFolders(%s)", spaceID))
	
	if c.foldersErr != nil {
		return nil, c.foldersErr
	}
	
	folders, ok := c.folders[spaceID]
	if !ok {
		return []clickup.Folder{}, nil
	}
	return folders, nil
}

// GetLists returns the mocked lists
func (c *Client) GetLists(ctx context.Context, folderID string) ([]clickup.List, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = append(c.calls, fmt.Sprintf("GetLists(%s)", folderID))
	
	if c.listsErr != nil {
		return nil, c.listsErr
	}
	
	lists, ok := c.lists[folderID]
	if !ok {
		return []clickup.List{}, nil
	}
	return lists, nil
}

// GetFolderlessLists returns the mocked folderless lists
func (c *Client) GetFolderlessLists(ctx context.Context, spaceID string) ([]clickup.List, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = append(c.calls, fmt.Sprintf("GetFolderlessLists(%s)", spaceID))
	
	if c.folderlessErr != nil {
		return nil, c.folderlessErr
	}
	
	lists, ok := c.folderlessLists[spaceID]
	if !ok {
		return []clickup.List{}, nil
	}
	return lists, nil
}

// GetTask returns the mocked task
func (c *Client) GetTask(ctx context.Context, taskID string) (*clickup.Task, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = append(c.calls, fmt.Sprintf("GetTask(%s)", taskID))
	
	if c.tasksErr != nil {
		return nil, c.tasksErr
	}
	
	task, ok := c.tasks[taskID]
	if !ok {
		return nil, fmt.Errorf("task not found")
	}
	return task, nil
}

// GetTasks returns the mocked tasks for a list
func (c *Client) GetTasks(ctx context.Context, listID string, options *api.TaskQueryOptions) ([]clickup.Task, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = append(c.calls, fmt.Sprintf("GetTasks(%s)", listID))
	
	if c.tasksErr != nil {
		return nil, c.tasksErr
	}
	
	tasks, ok := c.taskLists[listID]
	if !ok {
		return []clickup.Task{}, nil
	}
	
	// Apply basic filtering if options provided
	if options != nil {
		var filtered []clickup.Task
		for _, task := range tasks {
			// Simple status filter
			if len(options.Statuses) > 0 {
				match := false
				for _, status := range options.Statuses {
					if task.Status.Status == status {
						match = true
						break
					}
				}
				if !match {
					continue
				}
			}
			filtered = append(filtered, task)
		}
		return filtered, nil
	}
	
	return tasks, nil
}

// CreateTask returns the mocked created task
func (c *Client) CreateTask(ctx context.Context, listID string, options *api.TaskCreateOptions) (*clickup.Task, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = append(c.calls, fmt.Sprintf("CreateTask(%s)", listID))
	return c.createTaskResult, c.createTaskErr
}

// UpdateTask returns the mocked updated task
func (c *Client) UpdateTask(ctx context.Context, taskID string, options *api.TaskUpdateOptions) (*clickup.Task, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = append(c.calls, fmt.Sprintf("UpdateTask(%s)", taskID))
	return c.updateTaskResult, c.updateTaskErr
}

// DeleteTask returns the mocked delete error
func (c *Client) DeleteTask(ctx context.Context, taskID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = append(c.calls, fmt.Sprintf("DeleteTask(%s)", taskID))
	return c.deleteTaskErr
}

// GetTaskComments returns the mocked comments
func (c *Client) GetTaskComments(ctx context.Context, taskID string) ([]clickup.Comment, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = append(c.calls, fmt.Sprintf("GetTaskComments(%s)", taskID))
	
	if c.commentsErr != nil {
		return nil, c.commentsErr
	}
	
	comments, ok := c.comments[taskID]
	if !ok {
		return []clickup.Comment{}, nil
	}
	return comments, nil
}

// CreateTaskComment returns the mocked comment response
func (c *Client) CreateTaskComment(ctx context.Context, taskID string, text string, assignee string, notifyAll bool) (*clickup.CreateCommentResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = append(c.calls, fmt.Sprintf("CreateTaskComment(%s)", taskID))
	return c.createCommentRes, c.createCommentErr
}

// UpdateTaskComment returns the mocked update error
func (c *Client) UpdateTaskComment(ctx context.Context, commentID string, text string, resolved bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = append(c.calls, fmt.Sprintf("UpdateTaskComment(%s)", commentID))
	return c.updateCommentErr
}

// DeleteTaskComment returns the mocked delete error
func (c *Client) DeleteTaskComment(ctx context.Context, commentID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = append(c.calls, fmt.Sprintf("DeleteTaskComment(%s)", commentID))
	return c.deleteCommentErr
}

// UserLookup returns the mock user lookup service
func (c *Client) UserLookup() *api.UserLookup {
	c.mu.RLock()
	defer c.mu.RUnlock()
	// Return a real UserLookup that wraps our mock
	// This is a bit of a hack but necessary due to the current design
	return nil // We'll need to refactor this
}

// Helper methods for test setup

// SetCurrentUser sets the current user response
func (c *Client) SetCurrentUser(user *clickup.User, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.currentUser = user
	c.currentUserErr = err
}

// SetWorkspaces sets the workspaces response
func (c *Client) SetWorkspaces(workspaces []clickup.Team, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.workspaces = workspaces
	c.workspacesErr = err
}

// SetWorkspaceMembers sets members for a workspace
func (c *Client) SetWorkspaceMembers(workspaceID string, members []clickup.TeamUser) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.workspaceMembers[workspaceID] = members
}

// SetSpaces sets spaces for a workspace
func (c *Client) SetSpaces(workspaceID string, spaces []clickup.Space) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.spaces[workspaceID] = spaces
}

// SetTask sets a task by ID
func (c *Client) SetTask(task *clickup.Task) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tasks[task.ID] = task
}

// SetTaskList sets tasks for a list
func (c *Client) SetTaskList(listID string, tasks []clickup.Task) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.taskLists[listID] = tasks
}

// SetCreateTaskResponse sets the response for CreateTask
func (c *Client) SetCreateTaskResponse(task *clickup.Task, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.createTaskResult = task
	c.createTaskErr = err
}

// SetUpdateTaskResponse sets the response for UpdateTask
func (c *Client) SetUpdateTaskResponse(task *clickup.Task, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.updateTaskResult = task
	c.updateTaskErr = err
}

// SetDeleteTaskError sets the error for DeleteTask
func (c *Client) SetDeleteTaskError(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.deleteTaskErr = err
}

// GetCalls returns the list of method calls made
func (c *Client) GetCalls() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	calls := make([]string, len(c.calls))
	copy(calls, c.calls)
	return calls
}

// Reset clears all mock data and call history
func (c *Client) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.currentUser = nil
	c.currentUserErr = nil
	c.workspaces = nil
	c.workspacesErr = nil
	c.workspaceMembers = make(map[string][]clickup.TeamUser)
	c.membersErr = nil
	
	c.spaces = make(map[string][]clickup.Space)
	c.spacesErr = nil
	c.folders = make(map[string][]clickup.Folder)
	c.foldersErr = nil
	c.lists = make(map[string][]clickup.List)
	c.listsErr = nil
	c.folderlessLists = make(map[string][]clickup.List)
	c.folderlessErr = nil
	
	c.tasks = make(map[string]*clickup.Task)
	c.taskLists = make(map[string][]clickup.Task)
	c.tasksErr = nil
	c.createTaskResult = nil
	c.createTaskErr = nil
	c.updateTaskResult = nil
	c.updateTaskErr = nil
	c.deleteTaskErr = nil
	
	c.comments = make(map[string][]clickup.Comment)
	c.commentsErr = nil
	c.createCommentRes = nil
	c.createCommentErr = nil
	c.updateCommentErr = nil
	c.deleteCommentErr = nil
	
	c.calls = []string{}
}
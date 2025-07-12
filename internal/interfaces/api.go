package interfaces

import (
	"context"
	"time"

	"github.com/raksul/go-clickup/clickup"
)

// APIClient defines the interface for ClickUp API operations
type APIClient interface {
	// Authentication
	GetAuthorizedUser(ctx context.Context) (*clickup.User, error)
	GetAuthorizedTeams(ctx context.Context) ([]clickup.Team, error)

	// Workspace operations
	GetWorkspaces(ctx context.Context) ([]clickup.Team, error)
	
	// Space operations
	GetSpaces(ctx context.Context, teamID string) ([]clickup.Space, error)
	GetSpace(ctx context.Context, spaceID string) (*clickup.Space, error)
	CreateSpace(ctx context.Context, teamID string, request *clickup.SpaceRequest) (*clickup.Space, error)
	UpdateSpace(ctx context.Context, spaceID string, request *clickup.SpaceRequest) (*clickup.Space, error)
	DeleteSpace(ctx context.Context, spaceID string) error

	// Folder operations
	GetFolders(ctx context.Context, spaceID string) ([]clickup.Folder, error)
	GetFolder(ctx context.Context, folderID string) (*clickup.Folder, error)
	CreateFolder(ctx context.Context, spaceID string, request *clickup.FolderRequest) (*clickup.Folder, error)
	UpdateFolder(ctx context.Context, folderID string, request *clickup.FolderRequest) (*clickup.Folder, error)
	DeleteFolder(ctx context.Context, folderID string) error

	// List operations
	GetLists(ctx context.Context, folderID string) ([]clickup.List, error)
	GetFolderlessLists(ctx context.Context, spaceID string) ([]clickup.List, error)
	GetList(ctx context.Context, listID string) (*clickup.List, error)
	CreateList(ctx context.Context, folderID string, request *clickup.ListRequest) (*clickup.List, error)
	CreateFolderlessList(ctx context.Context, spaceID string, request *clickup.ListRequest) (*clickup.List, error)
	UpdateList(ctx context.Context, listID string, request *clickup.ListRequest) (*clickup.List, error)
	DeleteList(ctx context.Context, listID string) error

	// Task operations
	GetTasks(ctx context.Context, listID string, options *TaskQueryOptions) ([]clickup.Task, error)
	GetTask(ctx context.Context, taskID string) (*clickup.Task, error)
	CreateTask(ctx context.Context, listID string, options *TaskCreateOptions) (*clickup.Task, error)
	UpdateTask(ctx context.Context, taskID string, options *TaskUpdateOptions) (*clickup.Task, error)
	DeleteTask(ctx context.Context, taskID string) error

	// User operations
	GetCurrentUser(ctx context.Context) (*clickup.User, error)
	GetWorkspaceMembers(ctx context.Context, workspaceID string) ([]clickup.TeamUser, error)

	// Member operations
	GetMembers(ctx context.Context, listID string) ([]clickup.Member, error)

	// Comment operations
	GetTaskComments(ctx context.Context, taskID string) ([]clickup.Comment, error)
	CreateTaskComment(ctx context.Context, taskID string, text string, assignee string, notifyAll bool) (*clickup.CreateCommentResponse, error)
	UpdateTaskComment(ctx context.Context, commentID string, text string, resolved bool) error
	DeleteTaskComment(ctx context.Context, commentID string) error

	// Custom field operations
	GetCustomFields(ctx context.Context, listID string) ([]clickup.CustomField, error)
	SetCustomFieldValue(ctx context.Context, taskID string, fieldID string, value map[string]interface{}) error

	// View operations
	GetViews(ctx context.Context, listID string) ([]clickup.View, error)
	GetView(ctx context.Context, viewID string) (*clickup.View, error)

	// Goal operations
	GetGoals(ctx context.Context, teamID string, includeCompleted bool) ([]clickup.Goal, []clickup.GoalFolder, error)
	GetGoal(ctx context.Context, goalID string) (*clickup.Goal, error)
	CreateGoal(ctx context.Context, teamID string, request *clickup.CreateGoalRequest) (*clickup.Goal, error)
	UpdateGoal(ctx context.Context, goalID string, request *clickup.UpdateGoalRequest) (*clickup.Goal, error)
	DeleteGoal(ctx context.Context, goalID string) error

	// Webhook operations
	GetWebhooks(ctx context.Context, teamID string) ([]clickup.Webhook, error)
	CreateWebhook(ctx context.Context, teamID string, request *clickup.WebhookRequest) (*clickup.Webhook, error)
	UpdateWebhook(ctx context.Context, webhookID string, request *clickup.WebhookRequest) (*clickup.Webhook, error)
	DeleteWebhook(ctx context.Context, webhookID string) error
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
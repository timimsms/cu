package factory

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/raksul/go-clickup/clickup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tim/cu/internal/interfaces"
	"github.com/tim/cu/internal/mocks"
)

func TestTaskCommand(t *testing.T) {
	t.Run("no subcommand shows error", func(t *testing.T) {
		// Setup
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		
		factory := New(
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Create command
		cmd, err := factory.CreateCommand("task")
		require.NoError(t, err)
		require.NotNil(t, cmd)
		
		// Execute without subcommand
		err = cmd.Execute(context.Background(), []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no subcommand specified")
	})

	t.Run("unknown subcommand", func(t *testing.T) {
		// Setup
		factory := New()
		cmd, err := factory.CreateCommand("task")
		require.NoError(t, err)
		
		// Execute with unknown subcommand
		err = cmd.Execute(context.Background(), []string{"unknown"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown subcommand: unknown")
	})
}

func TestTaskCommand_List(t *testing.T) {
	t.Run("list tasks successfully", func(t *testing.T) {
		// Setup
		mockAPI := &MockAPIClient{}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.Set("default_list", "list123")
		mockConfig.Set("output", "table")
		
		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Mock API response
		mockTasks := []clickup.Task{
			{
				ID:   "task1",
				Name: "Test Task 1",
				Status: clickup.TaskStatus{
					Status: "open",
				},
				Priority: clickup.TaskPriority{
					Priority: "2",
				},
			},
			{
				ID:   "task2",
				Name: "Test Task 2",
				Status: clickup.TaskStatus{
					Status: "in progress",
				},
			},
		}
		mockAPI.GetTasksFunc = func(ctx context.Context, listID string, options *interfaces.TaskQueryOptions) ([]clickup.Task, error) {
			assert.Equal(t, "list123", listID)
			return mockTasks, nil
		}
		
		// Create command
		cmd, err := factory.CreateCommand("task")
		require.NoError(t, err)
		
		// Execute list subcommand
		err = cmd.Execute(context.Background(), []string{"list"})
		assert.NoError(t, err)
		
		// Verify output was called
		assert.Len(t, mockOutput.Printed, 1)
		// Output should have task rows
		assert.NotNil(t, mockOutput.Printed[0])
	})

	t.Run("list with no default list", func(t *testing.T) {
		// Setup
		mockAPI := &MockAPIClient{}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		// No default list set
		
		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Create command
		cmd, err := factory.CreateCommand("task")
		require.NoError(t, err)
		
		// Execute list without list ID
		err = cmd.Execute(context.Background(), []string{"list"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no list specified")
	})

	t.Run("list with API error", func(t *testing.T) {
		// Setup
		mockAPI := &MockAPIClient{}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.Set("default_list", "list123")
		
		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Mock API error
		mockAPI.GetTasksFunc = func(ctx context.Context, listID string, options *interfaces.TaskQueryOptions) ([]clickup.Task, error) {
			return nil, fmt.Errorf("API error")
		}
		
		// Create command
		cmd, err := factory.CreateCommand("task")
		require.NoError(t, err)
		
		// Execute
		err = cmd.Execute(context.Background(), []string{"list"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get tasks")
	})
}

func TestTaskCommand_Create(t *testing.T) {
	t.Run("create task successfully", func(t *testing.T) {
		// Setup
		mockAPI := &MockAPIClient{}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.Set("default_list", "list123")
		
		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Mock API response
		createdTask := &clickup.Task{
			ID:   "task123",
			Name: "New Task",
		}
		mockAPI.CreateTaskFunc = func(ctx context.Context, listID string, options *interfaces.TaskCreateOptions) (*clickup.Task, error) {
			assert.Equal(t, "list123", listID)
			assert.Equal(t, "New Task", options.Name)
			return createdTask, nil
		}
		
		// Create command
		cmd, err := factory.CreateCommand("task")
		require.NoError(t, err)
		
		// Execute create subcommand
		err = cmd.Execute(context.Background(), []string{"create", "New Task"})
		assert.NoError(t, err)
		
		// Verify success message
		assert.Len(t, mockOutput.SuccessMsg, 1)
		assert.Contains(t, mockOutput.SuccessMsg[0], "Created task: New Task (task123)")
	})

	t.Run("create task with no name", func(t *testing.T) {
		// Setup
		mockAPI := &MockAPIClient{}
		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.Set("default_list", "list123")
		
		factory := New(
			WithAPIClient(mockAPI),
			WithConfigProvider(mockConfig),
		)
		cmd, err := factory.CreateCommand("task")
		require.NoError(t, err)
		
		// Execute create without name
		err = cmd.Execute(context.Background(), []string{"create"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task name is required")
	})

	t.Run("create task with options", func(t *testing.T) {
		// Setup
		mockAPI := &MockAPIClient{}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.Set("default_list", "list123")
		
		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Mock API response
		mockAPI.CreateTaskFunc = func(ctx context.Context, listID string, options *interfaces.TaskCreateOptions) (*clickup.Task, error) {
			// Verify options
			assert.Equal(t, "Task with options", options.Name)
			assert.Equal(t, "Task description", options.Description)
			assert.Equal(t, "high", options.Priority)
			assert.Contains(t, options.Tags, "important")
			return &clickup.Task{ID: "task456", Name: options.Name}, nil
		}
		
		// Create command
		cmd, err := factory.CreateCommand("task")
		require.NoError(t, err)
		
		// Get cobra command to set flags
		cobraCmd := cmd.GetCobraCommand()
		createCmd, _, err := cobraCmd.Find([]string{"create"})
		require.NoError(t, err)
		
		// Set flags
		createCmd.Flags().Set("description", "Task description")
		createCmd.Flags().Set("priority", "high")
		createCmd.Flags().Set("tag", "important")
		
		// Execute
		err = createCmd.RunE(createCmd, []string{"Task with options"})
		assert.NoError(t, err)
	})
}

func TestTaskCommand_View(t *testing.T) {
	t.Run("view task successfully", func(t *testing.T) {
		// Setup
		mockAPI := &MockAPIClient{}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.Set("output", "table")
		
		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Mock API response
		mockTask := &clickup.Task{
			ID:   "task123",
			Name: "Test Task",
			Description: "Task description",
			Status: clickup.TaskStatus{
				Status: "open",
			},
			Priority: clickup.TaskPriority{
				Priority: "2",
			},
		}
		mockAPI.GetTaskFunc = func(ctx context.Context, taskID string) (*clickup.Task, error) {
			assert.Equal(t, "task123", taskID)
			return mockTask, nil
		}
		
		// Create command
		cmd, err := factory.CreateCommand("task")
		require.NoError(t, err)
		
		// Execute view subcommand
		err = cmd.Execute(context.Background(), []string{"view", "task123"})
		assert.NoError(t, err)
		
		// Verify output
		assert.Contains(t, mockOutput.InfoMsg, "Task: Test Task")
		assert.Contains(t, mockOutput.InfoMsg, "ID: task123")
		assert.Contains(t, mockOutput.InfoMsg, "Status: open")
	})

	t.Run("view task with no ID", func(t *testing.T) {
		// Setup
		factory := New()
		cmd, err := factory.CreateCommand("task")
		require.NoError(t, err)
		
		// Execute view without ID
		err = cmd.Execute(context.Background(), []string{"view"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task ID is required")
	})
}

func TestTaskCommand_Update(t *testing.T) {
	t.Run("update task successfully", func(t *testing.T) {
		// Setup
		mockAPI := &MockAPIClient{}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		
		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Mock API response
		updatedTask := &clickup.Task{
			ID:   "task123",
			Name: "Updated Task",
		}
		mockAPI.UpdateTaskFunc = func(ctx context.Context, taskID string, options *interfaces.TaskUpdateOptions) (*clickup.Task, error) {
			assert.Equal(t, "task123", taskID)
			assert.Equal(t, "Updated Task", options.Name)
			return updatedTask, nil
		}
		
		// Create command
		cmd, err := factory.CreateCommand("task")
		require.NoError(t, err)
		
		// Get cobra command to set flags
		cobraCmd := cmd.GetCobraCommand()
		updateCmd, _, err := cobraCmd.Find([]string{"update"})
		require.NoError(t, err)
		
		// Set flags
		updateCmd.Flags().Set("name", "Updated Task")
		
		// Execute
		err = updateCmd.RunE(updateCmd, []string{"task123"})
		assert.NoError(t, err)
		
		// Verify success message
		assert.Len(t, mockOutput.SuccessMsg, 1)
		assert.Contains(t, mockOutput.SuccessMsg[0], "Updated task: Updated Task")
	})

	t.Run("update with no changes", func(t *testing.T) {
		// Setup
		mockAPI := &MockAPIClient{}
		factory := New(WithAPIClient(mockAPI))
		cmd, err := factory.CreateCommand("task")
		require.NoError(t, err)
		
		// Get cobra command
		cobraCmd := cmd.GetCobraCommand()
		updateCmd, _, err := cobraCmd.Find([]string{"update"})
		require.NoError(t, err)
		
		// Execute without any update flags
		err = updateCmd.RunE(updateCmd, []string{"task123"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no updates specified")
	})
}

func TestTaskCommand_Close(t *testing.T) {
	t.Run("close task successfully", func(t *testing.T) {
		// Setup
		mockAPI := &MockAPIClient{}
		mockOutput := mocks.NewMockOutputFormatter()
		
		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
		)
		
		// Mock API response
		mockAPI.UpdateTaskFunc = func(ctx context.Context, taskID string, options *interfaces.TaskUpdateOptions) (*clickup.Task, error) {
			assert.Equal(t, "task123", taskID)
			assert.Equal(t, "closed", options.Status)
			return &clickup.Task{ID: taskID, Name: "Closed Task"}, nil
		}
		
		// Create command
		cmd, err := factory.CreateCommand("task")
		require.NoError(t, err)
		
		// Execute close subcommand
		err = cmd.Execute(context.Background(), []string{"close", "task123"})
		assert.NoError(t, err)
		
		// Verify success message
		assert.Len(t, mockOutput.SuccessMsg, 1)
		assert.Contains(t, mockOutput.SuccessMsg[0], "Closed task: Closed Task")
	})
}

func TestTaskCommand_Reopen(t *testing.T) {
	t.Run("reopen task successfully", func(t *testing.T) {
		// Setup
		mockAPI := &MockAPIClient{}
		mockOutput := mocks.NewMockOutputFormatter()
		
		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
		)
		
		// Mock API response
		mockAPI.UpdateTaskFunc = func(ctx context.Context, taskID string, options *interfaces.TaskUpdateOptions) (*clickup.Task, error) {
			assert.Equal(t, "task123", taskID)
			assert.Equal(t, "open", options.Status)
			return &clickup.Task{ID: taskID, Name: "Reopened Task"}, nil
		}
		
		// Create command
		cmd, err := factory.CreateCommand("task")
		require.NoError(t, err)
		
		// Execute reopen subcommand
		err = cmd.Execute(context.Background(), []string{"reopen", "task123"})
		assert.NoError(t, err)
		
		// Verify success message
		assert.Len(t, mockOutput.SuccessMsg, 1)
		assert.Contains(t, mockOutput.SuccessMsg[0], "Reopened task: Reopened Task")
	})
}

func TestTaskCommand_Search(t *testing.T) {
	t.Run("search not implemented", func(t *testing.T) {
		// Setup
		mockAPI := &MockAPIClient{}
		factory := New(WithAPIClient(mockAPI))
		cmd, err := factory.CreateCommand("task")
		require.NoError(t, err)
		
		// Execute search
		err = cmd.Execute(context.Background(), []string{"search", "query"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "search functionality not yet implemented")
	})
}

func TestTaskCommand_Helpers(t *testing.T) {
	t.Run("truncate string", func(t *testing.T) {
		assert.Equal(t, "short", truncate("short", 10))
		assert.Equal(t, "1234567...", truncate("1234567890123", 10))
	})

	t.Run("parse due dates", func(t *testing.T) {
		now := time.Now()
		
		// Test relative dates
		today, err := parseDueDate("today")
		assert.NoError(t, err)
		assert.Equal(t, now.Day(), today.Day())
		
		tomorrow, err := parseDueDate("tomorrow")
		assert.NoError(t, err)
		assert.Equal(t, now.AddDate(0, 0, 1).Day(), tomorrow.Day())
		
		// Test absolute date
		specific, err := parseDueDate("2024-12-25")
		assert.NoError(t, err)
		assert.Equal(t, 25, specific.Day())
		assert.Equal(t, time.December, specific.Month())
		assert.Equal(t, 2024, specific.Year())
		
		// Test invalid date
		_, err = parseDueDate("invalid")
		assert.Error(t, err)
	})
}

// MockAPIClient is a mock implementation of the APIClient interface
type MockAPIClient struct {
	GetTasksFunc      func(ctx context.Context, listID string, options *interfaces.TaskQueryOptions) ([]clickup.Task, error)
	GetTaskFunc       func(ctx context.Context, taskID string) (*clickup.Task, error)
	CreateTaskFunc    func(ctx context.Context, listID string, options *interfaces.TaskCreateOptions) (*clickup.Task, error)
	UpdateTaskFunc    func(ctx context.Context, taskID string, options *interfaces.TaskUpdateOptions) (*clickup.Task, error)
	GetWorkspacesFunc func(ctx context.Context) ([]clickup.Team, error)
	GetSpacesFunc     func(ctx context.Context, teamID string) ([]clickup.Space, error)
}

func (m *MockAPIClient) GetTasks(ctx context.Context, listID string, options *interfaces.TaskQueryOptions) ([]clickup.Task, error) {
	if m.GetTasksFunc != nil {
		return m.GetTasksFunc(ctx, listID, options)
	}
	return nil, fmt.Errorf("GetTasks not implemented")
}

func (m *MockAPIClient) GetTask(ctx context.Context, taskID string) (*clickup.Task, error) {
	if m.GetTaskFunc != nil {
		return m.GetTaskFunc(ctx, taskID)
	}
	return nil, fmt.Errorf("GetTask not implemented")
}

func (m *MockAPIClient) CreateTask(ctx context.Context, listID string, options *interfaces.TaskCreateOptions) (*clickup.Task, error) {
	if m.CreateTaskFunc != nil {
		return m.CreateTaskFunc(ctx, listID, options)
	}
	return nil, fmt.Errorf("CreateTask not implemented")
}

func (m *MockAPIClient) UpdateTask(ctx context.Context, taskID string, options *interfaces.TaskUpdateOptions) (*clickup.Task, error) {
	if m.UpdateTaskFunc != nil {
		return m.UpdateTaskFunc(ctx, taskID, options)
	}
	return nil, fmt.Errorf("UpdateTask not implemented")
}

// Implement other required methods with default behavior
func (m *MockAPIClient) GetAuthorizedUser(ctx context.Context) (*clickup.User, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAPIClient) GetAuthorizedTeams(ctx context.Context) ([]clickup.Team, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAPIClient) GetWorkspaces(ctx context.Context) ([]clickup.Team, error) {
	if m.GetWorkspacesFunc != nil {
		return m.GetWorkspacesFunc(ctx)
	}
	return nil, fmt.Errorf("GetWorkspaces not implemented")
}

func (m *MockAPIClient) GetSpaces(ctx context.Context, teamID string) ([]clickup.Space, error) {
	if m.GetSpacesFunc != nil {
		return m.GetSpacesFunc(ctx, teamID)
	}
	return nil, fmt.Errorf("GetSpaces not implemented")
}

func (m *MockAPIClient) GetSpace(ctx context.Context, spaceID string) (*clickup.Space, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAPIClient) CreateSpace(ctx context.Context, teamID string, request *clickup.SpaceRequest) (*clickup.Space, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAPIClient) UpdateSpace(ctx context.Context, spaceID string, request *clickup.SpaceRequest) (*clickup.Space, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAPIClient) DeleteSpace(ctx context.Context, spaceID string) error {
	return fmt.Errorf("not implemented")
}

func (m *MockAPIClient) GetFolders(ctx context.Context, spaceID string) ([]clickup.Folder, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAPIClient) GetFolder(ctx context.Context, folderID string) (*clickup.Folder, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAPIClient) CreateFolder(ctx context.Context, spaceID string, request *clickup.FolderRequest) (*clickup.Folder, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAPIClient) UpdateFolder(ctx context.Context, folderID string, request *clickup.FolderRequest) (*clickup.Folder, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAPIClient) DeleteFolder(ctx context.Context, folderID string) error {
	return fmt.Errorf("not implemented")
}

func (m *MockAPIClient) GetLists(ctx context.Context, folderID string) ([]clickup.List, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAPIClient) GetFolderlessLists(ctx context.Context, spaceID string) ([]clickup.List, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAPIClient) GetList(ctx context.Context, listID string) (*clickup.List, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAPIClient) CreateList(ctx context.Context, folderID string, request *clickup.ListRequest) (*clickup.List, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAPIClient) CreateFolderlessList(ctx context.Context, spaceID string, request *clickup.ListRequest) (*clickup.List, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAPIClient) UpdateList(ctx context.Context, listID string, request *clickup.ListRequest) (*clickup.List, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAPIClient) DeleteList(ctx context.Context, listID string) error {
	return fmt.Errorf("not implemented")
}

func (m *MockAPIClient) DeleteTask(ctx context.Context, taskID string) error {
	return fmt.Errorf("not implemented")
}

func (m *MockAPIClient) GetCurrentUser(ctx context.Context) (*clickup.User, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAPIClient) GetWorkspaceMembers(ctx context.Context, workspaceID string) ([]clickup.TeamUser, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAPIClient) GetMembers(ctx context.Context, listID string) ([]clickup.Member, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAPIClient) GetTaskComments(ctx context.Context, taskID string) ([]clickup.Comment, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAPIClient) CreateTaskComment(ctx context.Context, taskID string, text string, assignee string, notifyAll bool) (*clickup.CreateCommentResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAPIClient) UpdateTaskComment(ctx context.Context, commentID string, text string, resolved bool) error {
	return fmt.Errorf("not implemented")
}

func (m *MockAPIClient) DeleteTaskComment(ctx context.Context, commentID string) error {
	return fmt.Errorf("not implemented")
}

func (m *MockAPIClient) GetCustomFields(ctx context.Context, listID string) ([]clickup.CustomField, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAPIClient) SetCustomFieldValue(ctx context.Context, taskID string, fieldID string, value map[string]interface{}) error {
	return fmt.Errorf("not implemented")
}

func (m *MockAPIClient) GetViews(ctx context.Context, listID string) ([]clickup.View, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAPIClient) GetView(ctx context.Context, viewID string) (*clickup.View, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAPIClient) GetGoals(ctx context.Context, teamID string, includeCompleted bool) ([]clickup.Goal, []clickup.GoalFolder, error) {
	return nil, nil, fmt.Errorf("not implemented")
}

func (m *MockAPIClient) GetGoal(ctx context.Context, goalID string) (*clickup.Goal, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAPIClient) CreateGoal(ctx context.Context, teamID string, request *clickup.CreateGoalRequest) (*clickup.Goal, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAPIClient) UpdateGoal(ctx context.Context, goalID string, request *clickup.UpdateGoalRequest) (*clickup.Goal, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAPIClient) DeleteGoal(ctx context.Context, goalID string) error {
	return fmt.Errorf("not implemented")
}

func (m *MockAPIClient) GetWebhooks(ctx context.Context, teamID string) ([]clickup.Webhook, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAPIClient) CreateWebhook(ctx context.Context, teamID string, request *clickup.WebhookRequest) (*clickup.Webhook, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAPIClient) UpdateWebhook(ctx context.Context, webhookID string, request *clickup.WebhookRequest) (*clickup.Webhook, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAPIClient) DeleteWebhook(ctx context.Context, webhookID string) error {
	return fmt.Errorf("not implemented")
}

// Ensure MockAPIClient implements APIClient interface
var _ interfaces.APIClient = (*MockAPIClient)(nil)
package factory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/raksul/go-clickup/clickup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tim/cu/internal/interfaces"
	"github.com/tim/cu/internal/mocks"
)

func TestExportCommand(t *testing.T) {
	t.Run("no subcommand shows error", func(t *testing.T) {
		// Setup
		factory := New()
		cmd, err := factory.CreateCommand("export")
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
		cmd, err := factory.CreateCommand("export")
		require.NoError(t, err)

		// Execute with unknown subcommand
		err = cmd.Execute(context.Background(), []string{"unknown"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown subcommand: unknown")
	})
}

func TestExportCommand_Tasks(t *testing.T) {
	t.Run("export tasks from list to CSV", func(t *testing.T) {
		// Setup
		mockAPI := &ExportMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()

		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)

		// Mock tasks
		mockTasks := []clickup.Task{
			{
				ID:          "task1",
				Name:        "Test Task 1",
				Status:      clickup.TaskStatus{Status: "open"},
				Priority:    clickup.TaskPriority{Priority: "2"},
				Assignees:   []clickup.User{{Username: "john"}},
				URL:         "https://app.clickup.com/task1",
				DateCreated: "1234567890",
				DateUpdated: "1234567899",
			},
			{
				ID:          "task2",
				Name:        "Test Task 2",
				Status:      clickup.TaskStatus{Status: "done"},
				Priority:    clickup.TaskPriority{Priority: "1"},
				Assignees:   []clickup.User{{Username: "jane"}},
				URL:         "https://app.clickup.com/task2",
				DateCreated: "1234567891",
				DateUpdated: "1234567898",
			},
		}

		mockAPI.GetTasksFunc = func(ctx context.Context, listID string, opts *interfaces.TaskQueryOptions) ([]clickup.Task, error) {
			assert.Equal(t, "list123", listID)
			return mockTasks, nil
		}

		// Create command and cast to ExportCommand
		cmd, err := factory.CreateCommand("export")
		require.NoError(t, err)
		exportCmd := cmd.(*ExportCommand)

		// Set output to buffer
		outputBuffer := &bytes.Buffer{}
		exportCmd.SetOutputWriter(outputBuffer)

		// Set flags
		exportCmd.listID = "list123"
		exportCmd.format = "csv"

		// Execute
		err = exportCmd.Execute(context.Background(), []string{"tasks"})
		assert.NoError(t, err)

		// Verify CSV output
		csvOutput := outputBuffer.String()
		assert.Contains(t, csvOutput, "ID,Name,Status,Priority,Assignees,Due Date,Created,Updated,URL")
		assert.Contains(t, csvOutput, "task1,Test Task 1,open,high,john")
		assert.Contains(t, csvOutput, "task2,Test Task 2,done,urgent,jane")
	})

	t.Run("export tasks to JSON", func(t *testing.T) {
		// Setup
		mockAPI := &ExportMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()

		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)

		// Mock tasks
		mockTasks := []clickup.Task{
			{
				ID:     "task1",
				Name:   "Test Task 1",
				Status: clickup.TaskStatus{Status: "open"},
			},
		}

		mockAPI.GetTasksFunc = func(ctx context.Context, listID string, opts *interfaces.TaskQueryOptions) ([]clickup.Task, error) {
			return mockTasks, nil
		}

		// Create command and cast to ExportCommand
		cmd, err := factory.CreateCommand("export")
		require.NoError(t, err)
		exportCmd := cmd.(*ExportCommand)

		// Set output to buffer
		outputBuffer := &bytes.Buffer{}
		exportCmd.SetOutputWriter(outputBuffer)

		// Set flags
		exportCmd.listID = "list123"
		exportCmd.format = "json"

		// Execute
		err = exportCmd.Execute(context.Background(), []string{"tasks"})
		assert.NoError(t, err)

		// Verify JSON output
		var tasks []clickup.Task
		err = json.Unmarshal(outputBuffer.Bytes(), &tasks)
		assert.NoError(t, err)
		assert.Len(t, tasks, 1)
		assert.Equal(t, "task1", tasks[0].ID)
	})

	t.Run("export tasks to Markdown", func(t *testing.T) {
		// Setup
		mockAPI := &ExportMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()

		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)

		// Mock tasks with different statuses
		mockTasks := []clickup.Task{
			{
				ID:          "task1",
				Name:        "Open Task",
				Status:      clickup.TaskStatus{Status: "open"},
				Priority:    clickup.TaskPriority{Priority: "2"},
				Description: "This is a test task",
				URL:         "https://app.clickup.com/task1",
			},
			{
				ID:     "task2",
				Name:   "Done Task",
				Status: clickup.TaskStatus{Status: "done"},
			},
		}

		mockAPI.GetTasksFunc = func(ctx context.Context, listID string, opts *interfaces.TaskQueryOptions) ([]clickup.Task, error) {
			return mockTasks, nil
		}

		// Create command and cast to ExportCommand
		cmd, err := factory.CreateCommand("export")
		require.NoError(t, err)
		exportCmd := cmd.(*ExportCommand)

		// Set output to buffer
		outputBuffer := &bytes.Buffer{}
		exportCmd.SetOutputWriter(outputBuffer)

		// Set flags
		exportCmd.listID = "list123"
		exportCmd.format = "markdown"

		// Execute
		err = exportCmd.Execute(context.Background(), []string{"tasks"})
		assert.NoError(t, err)

		// Verify Markdown output
		mdOutput := outputBuffer.String()
		assert.Contains(t, mdOutput, "# Task Report")
		assert.Contains(t, mdOutput, "Total tasks: 2")
		assert.Contains(t, mdOutput, "## Summary by Status")
		assert.Contains(t, mdOutput, "### Open Task")
		assert.Contains(t, mdOutput, "### Done Task")
		assert.Contains(t, mdOutput, "This is a test task")
		assert.Contains(t, mdOutput, "[View in ClickUp]")
	})

	t.Run("export with status filter", func(t *testing.T) {
		// Setup
		mockAPI := &ExportMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()

		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)

		// Mock tasks
		mockTasks := []clickup.Task{
			{ID: "task1", Status: clickup.TaskStatus{Status: "open"}},
			{ID: "task2", Status: clickup.TaskStatus{Status: "done"}},
		}

		// Track query options
		var capturedOpts *interfaces.TaskQueryOptions
		mockAPI.GetTasksFunc = func(ctx context.Context, listID string, opts *interfaces.TaskQueryOptions) ([]clickup.Task, error) {
			capturedOpts = opts
			// Return only open tasks when status filter is applied
			if opts != nil && len(opts.Statuses) > 0 && opts.Statuses[0] == "open" {
				return []clickup.Task{mockTasks[0]}, nil
			}
			return mockTasks, nil
		}

		// Create command
		cmd, err := factory.CreateCommand("export")
		require.NoError(t, err)

		// Get cobra command to set flags
		cobraCmd := cmd.GetCobraCommand()
		tasksCmd, _, err := cobraCmd.Find([]string{"tasks"})
		require.NoError(t, err)

		// Set flags
		tasksCmd.Flags().Set("list", "list123")
		tasksCmd.Flags().Set("status", "open")
		tasksCmd.Flags().Set("format", "json")

		// Set output to buffer
		exportCmd := cmd.(*ExportCommand)
		outputBuffer := &bytes.Buffer{}
		exportCmd.SetOutputWriter(outputBuffer)

		// Execute
		err = tasksCmd.RunE(tasksCmd, []string{})
		assert.NoError(t, err)

		// Verify status filter was applied
		assert.NotNil(t, capturedOpts)
		assert.Equal(t, []string{"open"}, capturedOpts.Statuses)
	})

	t.Run("export with priority filter", func(t *testing.T) {
		// Setup
		mockAPI := &ExportMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()

		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)

		// Track query options
		var capturedOpts *interfaces.TaskQueryOptions
		mockAPI.GetTasksFunc = func(ctx context.Context, listID string, opts *interfaces.TaskQueryOptions) ([]clickup.Task, error) {
			capturedOpts = opts
			return []clickup.Task{}, nil
		}

		// Create command
		cmd, err := factory.CreateCommand("export")
		require.NoError(t, err)

		// Get cobra command to set flags
		cobraCmd := cmd.GetCobraCommand()
		tasksCmd, _, err := cobraCmd.Find([]string{"tasks"})
		require.NoError(t, err)

		// Set flags
		tasksCmd.Flags().Set("list", "list123")
		tasksCmd.Flags().Set("priority", "high")
		tasksCmd.Flags().Set("format", "csv")

		// Set output to buffer
		exportCmd := cmd.(*ExportCommand)
		outputBuffer := &bytes.Buffer{}
		exportCmd.SetOutputWriter(outputBuffer)

		// Execute
		err = tasksCmd.RunE(tasksCmd, []string{})
		assert.NoError(t, err)

		// Verify priority filter was applied
		assert.NotNil(t, capturedOpts)
		assert.NotNil(t, capturedOpts.Priority)
		assert.Equal(t, 2, *capturedOpts.Priority) // high = 2
	})

	t.Run("export all tasks from space", func(t *testing.T) {
		// Setup
		mockAPI := &ExportMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()

		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)

		// Mock workspace and space structure
		mockAPI.GetWorkspacesFunc = func(ctx context.Context) ([]clickup.Team, error) {
			return []clickup.Team{{ID: "workspace1"}}, nil
		}

		mockAPI.GetSpacesFunc = func(ctx context.Context, workspaceID string) ([]clickup.Space, error) {
			return []clickup.Space{
				{ID: "space1", Name: "Test Space"},
				{ID: "space2", Name: "Other Space"},
			}, nil
		}

		mockAPI.GetFoldersFunc = func(ctx context.Context, spaceID string) ([]clickup.Folder, error) {
			if spaceID == "space1" {
				return []clickup.Folder{{ID: "folder1"}}, nil
			}
			return []clickup.Folder{}, nil
		}

		mockAPI.GetListsFunc = func(ctx context.Context, folderID string) ([]clickup.List, error) {
			if folderID == "folder1" {
				return []clickup.List{{ID: "list1"}}, nil
			}
			return []clickup.List{}, nil
		}

		mockAPI.GetFolderlessListsFunc = func(ctx context.Context, spaceID string) ([]clickup.List, error) {
			if spaceID == "space1" {
				return []clickup.List{{ID: "list2"}}, nil
			}
			return []clickup.List{}, nil
		}

		// Track which lists were queried
		var queriedLists []string
		mockAPI.GetTasksFunc = func(ctx context.Context, listID string, opts *interfaces.TaskQueryOptions) ([]clickup.Task, error) {
			queriedLists = append(queriedLists, listID)
			return []clickup.Task{{ID: "task-from-" + listID}}, nil
		}

		// Create command and set space filter
		cmd, err := factory.CreateCommand("export")
		require.NoError(t, err)
		exportCmd := cmd.(*ExportCommand)
		exportCmd.spaceID = "space1"
		exportCmd.format = "json"

		// Set output to buffer
		outputBuffer := &bytes.Buffer{}
		exportCmd.SetOutputWriter(outputBuffer)

		// Execute
		err = exportCmd.Execute(context.Background(), []string{"tasks"})
		assert.NoError(t, err)

		// Verify lists from space1 were queried
		assert.Contains(t, queriedLists, "list1")
		assert.Contains(t, queriedLists, "list2")

		// Verify JSON output contains tasks from both lists
		var tasks []clickup.Task
		err = json.Unmarshal(outputBuffer.Bytes(), &tasks)
		assert.NoError(t, err)
		assert.Len(t, tasks, 2)
	})

	t.Run("export to file", func(t *testing.T) {
		// Setup
		mockAPI := &ExportMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()

		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)

		// Mock tasks
		mockAPI.GetTasksFunc = func(ctx context.Context, listID string, opts *interfaces.TaskQueryOptions) ([]clickup.Task, error) {
			return []clickup.Task{{ID: "task1", Name: "Test"}}, nil
		}

		// Create command
		cmd, err := factory.CreateCommand("export")
		require.NoError(t, err)

		// Get cobra command to set flags
		cobraCmd := cmd.GetCobraCommand()
		tasksCmd, _, err := cobraCmd.Find([]string{"tasks"})
		require.NoError(t, err)

		// Set flags with output file
		tasksCmd.Flags().Set("list", "list123")
		tasksCmd.Flags().Set("format", "csv")
		tasksCmd.Flags().Set("output", "test-export.csv")

		// Execute
		err = tasksCmd.RunE(tasksCmd, []string{})
		assert.NoError(t, err)

		// Verify success message
		assert.Contains(t, mockOutput.SuccessMsg, "Exported 1 task(s) to test-export.csv")

		// Clean up test file
		// Note: In a real test, we would create a temp directory
	})

	t.Run("export with invalid format", func(t *testing.T) {
		// Setup
		mockAPI := &ExportMockAPIClient{MockAPIClient: &MockAPIClient{}}
		factory := New(WithAPIClient(mockAPI))

		// Create command
		cmd, err := factory.CreateCommand("export")
		require.NoError(t, err)
		exportCmd := cmd.(*ExportCommand)

		// Set invalid format
		exportCmd.format = "invalid"
		exportCmd.listID = "list123"

		// Execute
		err = exportCmd.Execute(context.Background(), []string{"tasks"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid format: invalid")
	})

	t.Run("export with invalid output path", func(t *testing.T) {
		// Setup
		mockAPI := &ExportMockAPIClient{MockAPIClient: &MockAPIClient{}}
		factory := New(WithAPIClient(mockAPI))

		// Mock tasks
		mockAPI.GetTasksFunc = func(ctx context.Context, listID string, opts *interfaces.TaskQueryOptions) ([]clickup.Task, error) {
			return []clickup.Task{}, nil
		}

		// Create command
		cmd, err := factory.CreateCommand("export")
		require.NoError(t, err)
		exportCmd := cmd.(*ExportCommand)

		// Set invalid output path
		exportCmd.outputFile = "../../../etc/passwd"
		exportCmd.format = "csv"
		exportCmd.listID = "list123"

		// Execute
		err = exportCmd.Execute(context.Background(), []string{"tasks"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid output file path")
	})

	t.Run("export with no API client", func(t *testing.T) {
		// Setup
		factory := New() // No API client
		cmd, err := factory.CreateCommand("export")
		require.NoError(t, err)

		// Execute
		err = cmd.Execute(context.Background(), []string{"tasks"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API client not initialized")
	})

	t.Run("export with client-side filtering", func(t *testing.T) {
		// Setup
		mockAPI := &ExportMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()

		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)

		// Mock workspace structure without specific list
		mockAPI.GetWorkspacesFunc = func(ctx context.Context) ([]clickup.Team, error) {
			return []clickup.Team{{ID: "workspace1"}}, nil
		}

		mockAPI.GetSpacesFunc = func(ctx context.Context, workspaceID string) ([]clickup.Space, error) {
			return []clickup.Space{{ID: "space1"}}, nil
		}

		mockAPI.GetFoldersFunc = func(ctx context.Context, spaceID string) ([]clickup.Folder, error) {
			return []clickup.Folder{}, nil
		}

		mockAPI.GetFolderlessListsFunc = func(ctx context.Context, spaceID string) ([]clickup.List, error) {
			return []clickup.List{{ID: "list1"}}, nil
		}

		// Return mixed tasks for client-side filtering
		mockAPI.GetTasksFunc = func(ctx context.Context, listID string, opts *interfaces.TaskQueryOptions) ([]clickup.Task, error) {
			return []clickup.Task{
				{
					ID:        "task1",
					Name:      "High Priority Task",
					Status:    clickup.TaskStatus{Status: "open"},
					Priority:  clickup.TaskPriority{Priority: "2"},
					Assignees: []clickup.User{{ID: 123, Username: "john"}},
				},
				{
					ID:        "task2",
					Name:      "Low Priority Task",
					Status:    clickup.TaskStatus{Status: "done"},
					Priority:  clickup.TaskPriority{Priority: "4"},
					Assignees: []clickup.User{{ID: 456, Username: "jane"}},
				},
				{
					ID:       "task3",
					Name:     "No Priority Task",
					Status:   clickup.TaskStatus{Status: "open"},
					Priority: clickup.TaskPriority{},
				},
			}, nil
		}

		// Create command with filters
		cmd, err := factory.CreateCommand("export")
		require.NoError(t, err)
		exportCmd := cmd.(*ExportCommand)
		exportCmd.status = "open"
		exportCmd.priority = "high"
		exportCmd.assignee = "john"
		exportCmd.format = "json"

		// Set output to buffer
		outputBuffer := &bytes.Buffer{}
		exportCmd.SetOutputWriter(outputBuffer)

		// Execute
		err = exportCmd.Execute(context.Background(), []string{"tasks"})
		assert.NoError(t, err)

		// Verify only matching task was exported
		var tasks []clickup.Task
		err = json.Unmarshal(outputBuffer.Bytes(), &tasks)
		assert.NoError(t, err)
		assert.Len(t, tasks, 1)
		assert.Equal(t, "task1", tasks[0].ID)
	})

	t.Run("export markdown format alias", func(t *testing.T) {
		// Setup
		mockAPI := &ExportMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()

		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)

		// Mock tasks
		mockAPI.GetTasksFunc = func(ctx context.Context, listID string, opts *interfaces.TaskQueryOptions) ([]clickup.Task, error) {
			return []clickup.Task{{ID: "task1", Name: "Test"}}, nil
		}

		// Create command
		cmd, err := factory.CreateCommand("export")
		require.NoError(t, err)
		exportCmd := cmd.(*ExportCommand)

		// Set output to buffer
		outputBuffer := &bytes.Buffer{}
		exportCmd.SetOutputWriter(outputBuffer)

		// Set flags with "md" format
		exportCmd.listID = "list123"
		exportCmd.format = "md"

		// Execute
		err = exportCmd.Execute(context.Background(), []string{"tasks"})
		assert.NoError(t, err)

		// Verify markdown output was generated
		assert.Contains(t, outputBuffer.String(), "# Task Report")
	})
}

func TestExportCommand_GetCobraCommand(t *testing.T) {
	t.Run("has correct subcommands", func(t *testing.T) {
		// Setup
		factory := New()
		cmd, err := factory.CreateCommand("export")
		require.NoError(t, err)

		// Get cobra command
		cobraCmd := cmd.GetCobraCommand()

		// Verify subcommands exist
		assert.True(t, cobraCmd.HasSubCommands())

		// Check tasks subcommand
		tasksCmd, _, err := cobraCmd.Find([]string{"tasks"})
		require.NoError(t, err)
		assert.Equal(t, "tasks", tasksCmd.Use)
		assert.NotNil(t, tasksCmd.Flags().Lookup("list"))
		assert.NotNil(t, tasksCmd.Flags().Lookup("space"))
		assert.NotNil(t, tasksCmd.Flags().Lookup("format"))
		assert.NotNil(t, tasksCmd.Flags().Lookup("output"))
		assert.NotNil(t, tasksCmd.Flags().Lookup("status"))
		assert.NotNil(t, tasksCmd.Flags().Lookup("priority"))
		assert.NotNil(t, tasksCmd.Flags().Lookup("assignee"))
	})
}

// ExportMockAPIClient extends MockAPIClient with export-specific functions
type ExportMockAPIClient struct {
	*MockAPIClient
	GetWorkspacesFunc      func(ctx context.Context) ([]clickup.Team, error)
	GetSpacesFunc          func(ctx context.Context, workspaceID string) ([]clickup.Space, error)
	GetTasksFunc           func(ctx context.Context, listID string, opts *interfaces.TaskQueryOptions) ([]clickup.Task, error)
	GetFoldersFunc         func(ctx context.Context, spaceID string) ([]clickup.Folder, error)
	GetListsFunc           func(ctx context.Context, folderID string) ([]clickup.List, error)
	GetFolderlessListsFunc func(ctx context.Context, spaceID string) ([]clickup.List, error)
}

func (m *ExportMockAPIClient) GetWorkspaces(ctx context.Context) ([]clickup.Team, error) {
	if m.GetWorkspacesFunc != nil {
		return m.GetWorkspacesFunc(ctx)
	}
	return nil, fmt.Errorf("GetWorkspaces not implemented")
}

func (m *ExportMockAPIClient) GetSpaces(ctx context.Context, workspaceID string) ([]clickup.Space, error) {
	if m.GetSpacesFunc != nil {
		return m.GetSpacesFunc(ctx, workspaceID)
	}
	return nil, fmt.Errorf("GetSpaces not implemented")
}

func (m *ExportMockAPIClient) GetTasks(ctx context.Context, listID string, opts *interfaces.TaskQueryOptions) ([]clickup.Task, error) {
	if m.GetTasksFunc != nil {
		return m.GetTasksFunc(ctx, listID, opts)
	}
	return nil, fmt.Errorf("GetTasks not implemented")
}

func (m *ExportMockAPIClient) GetFolders(ctx context.Context, spaceID string) ([]clickup.Folder, error) {
	if m.GetFoldersFunc != nil {
		return m.GetFoldersFunc(ctx, spaceID)
	}
	return nil, fmt.Errorf("GetFolders not implemented")
}

func (m *ExportMockAPIClient) GetLists(ctx context.Context, folderID string) ([]clickup.List, error) {
	if m.GetListsFunc != nil {
		return m.GetListsFunc(ctx, folderID)
	}
	return nil, fmt.Errorf("GetLists not implemented")
}

func (m *ExportMockAPIClient) GetFolderlessLists(ctx context.Context, spaceID string) ([]clickup.List, error) {
	if m.GetFolderlessListsFunc != nil {
		return m.GetFolderlessListsFunc(ctx, spaceID)
	}
	return nil, fmt.Errorf("GetFolderlessLists not implemented")
}

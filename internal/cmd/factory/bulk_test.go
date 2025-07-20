package factory

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/raksul/go-clickup/clickup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tim/cu/internal/interfaces"
	"github.com/tim/cu/internal/mocks"
)

func TestBulkCommand(t *testing.T) {
	t.Run("no subcommand shows error", func(t *testing.T) {
		// Setup
		factory := New()
		cmd, err := factory.CreateCommand("bulk")
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
		cmd, err := factory.CreateCommand("bulk")
		require.NoError(t, err)

		// Execute with unknown subcommand
		err = cmd.Execute(context.Background(), []string{"unknown"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown subcommand: unknown")
	})
}

func TestBulkCommand_Update(t *testing.T) {
	t.Run("update multiple tasks with status", func(t *testing.T) {
		// Setup
		mockAPI := &BulkMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()

		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)

		// Track updated tasks
		updatedTasks := make(map[string]*interfaces.TaskUpdateOptions)
		mockAPI.UpdateTaskFunc = func(ctx context.Context, taskID string, opts *interfaces.TaskUpdateOptions) (*clickup.Task, error) {
			updatedTasks[taskID] = opts
			return nil, nil
		}

		// Create command
		cmd, err := factory.CreateCommand("bulk")
		require.NoError(t, err)

		// Get cobra command to set flags
		cobraCmd := cmd.GetCobraCommand()
		updateCmd, _, err := cobraCmd.Find([]string{"update"})
		require.NoError(t, err)

		// Set flags
		_ = updateCmd.Flags().Set("status", "done")
		_ = updateCmd.Flags().Set("yes", "true")

		// Execute
		err = updateCmd.RunE(updateCmd, []string{"task1", "task2", "task3"})
		assert.NoError(t, err)

		// Verify tasks were updated
		assert.Len(t, updatedTasks, 3)
		assert.Equal(t, "done", updatedTasks["task1"].Status)
		assert.Equal(t, "done", updatedTasks["task2"].Status)
		assert.Equal(t, "done", updatedTasks["task3"].Status)

		// Verify success messages
		assert.Contains(t, mockOutput.SuccessMsg, "task1")
		assert.Contains(t, mockOutput.SuccessMsg, "task2")
		assert.Contains(t, mockOutput.SuccessMsg, "task3")
	})

	t.Run("update with priority and tags", func(t *testing.T) {
		// Setup
		mockAPI := &BulkMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()

		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)

		// Track updated tasks
		var capturedOpts *interfaces.TaskUpdateOptions
		mockAPI.UpdateTaskFunc = func(ctx context.Context, taskID string, opts *interfaces.TaskUpdateOptions) (*clickup.Task, error) {
			capturedOpts = opts
			return nil, nil
		}

		// Create command
		cmd, err := factory.CreateCommand("bulk")
		require.NoError(t, err)

		// Get cobra command to set flags
		cobraCmd := cmd.GetCobraCommand()
		updateCmd, _, err := cobraCmd.Find([]string{"update"})
		require.NoError(t, err)

		// Set flags
		_ = updateCmd.Flags().Set("priority", "high")
		_ = updateCmd.Flags().Set("tag", "important,urgent")
		_ = updateCmd.Flags().Set("yes", "true")

		// Execute
		err = updateCmd.RunE(updateCmd, []string{"task1"})
		assert.NoError(t, err)

		// Verify update options
		assert.Equal(t, "high", capturedOpts.Priority)
		assert.Equal(t, []string{"important", "urgent"}, capturedOpts.Tags)
	})

	t.Run("update with assignee changes", func(t *testing.T) {
		// Setup
		mockAPI := &BulkMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()

		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)

		// Track updated tasks
		var capturedOpts *interfaces.TaskUpdateOptions
		mockAPI.UpdateTaskFunc = func(ctx context.Context, taskID string, opts *interfaces.TaskUpdateOptions) (*clickup.Task, error) {
			capturedOpts = opts
			return nil, nil
		}

		// Create command
		cmd, err := factory.CreateCommand("bulk")
		require.NoError(t, err)

		// Get cobra command to set flags
		cobraCmd := cmd.GetCobraCommand()
		updateCmd, _, err := cobraCmd.Find([]string{"update"})
		require.NoError(t, err)

		// Set flags
		_ = updateCmd.Flags().Set("add-assignee", "@john,@jane")
		_ = updateCmd.Flags().Set("remove-assignee", "@bob")
		_ = updateCmd.Flags().Set("yes", "true")

		// Execute
		err = updateCmd.RunE(updateCmd, []string{"task1"})
		assert.NoError(t, err)

		// Verify update options
		assert.Equal(t, []string{"@john", "@jane"}, capturedOpts.AddAssignees)
		assert.Equal(t, []string{"@bob"}, capturedOpts.RemoveAssignees)
	})

	t.Run("update with dry run", func(t *testing.T) {
		// Setup
		mockAPI := &BulkMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()

		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)

		// Track if update was called
		updateCalled := false
		mockAPI.UpdateTaskFunc = func(ctx context.Context, taskID string, opts *interfaces.TaskUpdateOptions) (*clickup.Task, error) {
			updateCalled = true
			return nil, nil
		}

		// Create command
		cmd, err := factory.CreateCommand("bulk")
		require.NoError(t, err)

		// Get cobra command to set flags
		cobraCmd := cmd.GetCobraCommand()
		updateCmd, _, err := cobraCmd.Find([]string{"update"})
		require.NoError(t, err)

		// Set flags
		_ = updateCmd.Flags().Set("status", "done")
		_ = updateCmd.Flags().Set("dry-run", "true")

		// Execute
		err = updateCmd.RunE(updateCmd, []string{"task1", "task2"})
		assert.NoError(t, err)

		// Verify update was NOT called
		assert.False(t, updateCalled)

		// Verify dry run message
		assert.Contains(t, mockOutput.InfoMsg, "Dry run - no changes will be made")
		assert.Contains(t, mockOutput.InfoMsg, "Would update tasks: task1, task2")
	})

	t.Run("update with interactive confirmation", func(t *testing.T) {
		// Setup
		mockAPI := &BulkMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()

		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)

		// Mock successful update
		mockAPI.UpdateTaskFunc = func(ctx context.Context, taskID string, opts *interfaces.TaskUpdateOptions) (*clickup.Task, error) {
			return nil, nil
		}

		// Create command and cast to BulkCommand
		cmd, err := factory.CreateCommand("bulk")
		require.NoError(t, err)
		bulkCmd := cmd.(*BulkCommand)

		// Set up test input (user confirms)
		stdin := strings.NewReader("y\n")
		stdout := &bytes.Buffer{}
		bulkCmd.SetStdin(stdin)
		bulkCmd.SetStdout(stdout)

		// Set status flag
		bulkCmd.status = "done"

		// Execute
		err = bulkCmd.Execute(context.Background(), []string{"update", "task1"})
		assert.NoError(t, err)

		// Verify confirmation prompt was shown
		assert.Contains(t, stdout.String(), "Are you sure you want to update 1 task(s)?")
	})

	t.Run("update cancelled by user", func(t *testing.T) {
		// Setup
		mockAPI := &BulkMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()

		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)

		// Create command and cast to BulkCommand
		cmd, err := factory.CreateCommand("bulk")
		require.NoError(t, err)
		bulkCmd := cmd.(*BulkCommand)

		// Set up test input (user cancels)
		stdin := strings.NewReader("n\n")
		stdout := &bytes.Buffer{}
		bulkCmd.SetStdin(stdin)
		bulkCmd.SetStdout(stdout)

		// Set status flag
		bulkCmd.status = "done"

		// Track if update was called
		updateCalled := false
		mockAPI.UpdateTaskFunc = func(ctx context.Context, taskID string, opts *interfaces.TaskUpdateOptions) (*clickup.Task, error) {
			updateCalled = true
			return nil, nil
		}

		// Execute
		err = bulkCmd.Execute(context.Background(), []string{"update", "task1"})
		assert.NoError(t, err)

		// Verify update was NOT called
		assert.False(t, updateCalled)

		// Verify cancelled message
		assert.Contains(t, mockOutput.InfoMsg, "Cancelled")
	})

	t.Run("update from stdin", func(t *testing.T) {
		// Setup
		mockAPI := &BulkMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()

		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)

		// Create command and cast to BulkCommand
		cmd, err := factory.CreateCommand("bulk")
		require.NoError(t, err)
		bulkCmd := cmd.(*BulkCommand)

		// Set up test input with task IDs from stdin
		stdin := strings.NewReader("task1\ntask2\ntask3\n")
		bulkCmd.SetStdin(stdin)
		bulkCmd.SetStdout(&bytes.Buffer{})

		// Set flags
		bulkCmd.status = "done"
		bulkCmd.yes = true

		// Track updated tasks
		var updatedTasks []string
		mockAPI.UpdateTaskFunc = func(ctx context.Context, taskID string, opts *interfaces.TaskUpdateOptions) (*clickup.Task, error) {
			updatedTasks = append(updatedTasks, taskID)
			return nil, nil
		}

		// Execute with no args (read from stdin)
		err = bulkCmd.Execute(context.Background(), []string{"update"})
		assert.NoError(t, err)

		// Verify tasks were updated
		assert.Equal(t, []string{"task1", "task2", "task3"}, updatedTasks)
	})

	t.Run("update with no updates specified", func(t *testing.T) {
		// Setup
		mockAPI := &BulkMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockConfig := mocks.NewMockConfigProvider()

		factory := New(
			WithAPIClient(mockAPI),
			WithConfigProvider(mockConfig),
		)

		// Create command
		cmd, err := factory.CreateCommand("bulk")
		require.NoError(t, err)

		// Execute without any update flags
		err = cmd.Execute(context.Background(), []string{"update", "task1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no updates specified")
	})

	t.Run("update with no task IDs", func(t *testing.T) {
		// Setup
		mockAPI := &BulkMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockConfig := mocks.NewMockConfigProvider()

		factory := New(
			WithAPIClient(mockAPI),
			WithConfigProvider(mockConfig),
		)

		// Create command and cast to BulkCommand
		cmd, err := factory.CreateCommand("bulk")
		require.NoError(t, err)
		bulkCmd := cmd.(*BulkCommand)

		// Set up empty stdin
		stdin := strings.NewReader("")
		bulkCmd.SetStdin(stdin)

		// Set status flag
		bulkCmd.status = "done"

		// Execute with no args
		err = bulkCmd.Execute(context.Background(), []string{"update"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no task IDs provided")
	})

	t.Run("update with API errors", func(t *testing.T) {
		// Setup
		mockAPI := &BulkMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()

		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)

		// Mock API errors for some tasks
		mockAPI.UpdateTaskFunc = func(ctx context.Context, taskID string, opts *interfaces.TaskUpdateOptions) (*clickup.Task, error) {
			if taskID == "task2" {
				return nil, fmt.Errorf("API error")
			}
			return nil, nil
		}

		// Create command and set flags
		cmd, err := factory.CreateCommand("bulk")
		require.NoError(t, err)
		bulkCmd := cmd.(*BulkCommand)
		bulkCmd.status = "done"
		bulkCmd.yes = true

		// Execute
		err = bulkCmd.Execute(context.Background(), []string{"update", "task1", "task2", "task3"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update 1 task(s)")

		// Verify summary shows correct counts
		// InfoMsg is a slice, need to check all messages
		allInfo := strings.Join(mockOutput.InfoMsg, " ")
		assert.Contains(t, allInfo, "Success: 2")
		assert.Contains(t, allInfo, "Failed:  1")
	})

	t.Run("update with no API client", func(t *testing.T) {
		// Setup
		factory := New() // No API client
		cmd, err := factory.CreateCommand("bulk")
		require.NoError(t, err)

		// Execute
		err = cmd.Execute(context.Background(), []string{"update", "task1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API client not initialized")
	})
}

func TestBulkCommand_Close(t *testing.T) {
	t.Run("close multiple tasks", func(t *testing.T) {
		// Setup
		mockAPI := &BulkMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()

		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)

		// Track updated tasks
		closedTasks := make(map[string]string)
		mockAPI.UpdateTaskFunc = func(ctx context.Context, taskID string, opts *interfaces.TaskUpdateOptions) (*clickup.Task, error) {
			closedTasks[taskID] = opts.Status
			return nil, nil
		}

		// Create command
		cmd, err := factory.CreateCommand("bulk")
		require.NoError(t, err)

		// Get cobra command to set flags
		cobraCmd := cmd.GetCobraCommand()
		closeCmd, _, err := cobraCmd.Find([]string{"close"})
		require.NoError(t, err)

		// Set flags
		_ = closeCmd.Flags().Set("yes", "true")

		// Execute
		err = closeCmd.RunE(closeCmd, []string{"task1", "task2"})
		assert.NoError(t, err)

		// Verify tasks were closed
		assert.Len(t, closedTasks, 2)
		assert.Equal(t, "complete", closedTasks["task1"])
		assert.Equal(t, "complete", closedTasks["task2"])
	})

	t.Run("close with confirmation", func(t *testing.T) {
		// Setup
		mockAPI := &BulkMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()

		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)

		// Create command and cast to BulkCommand
		cmd, err := factory.CreateCommand("bulk")
		require.NoError(t, err)
		bulkCmd := cmd.(*BulkCommand)

		// Set up test input (user confirms)
		stdin := strings.NewReader("y\n")
		stdout := &bytes.Buffer{}
		bulkCmd.SetStdin(stdin)
		bulkCmd.SetStdout(stdout)

		// Track if close was called
		closeCalled := false
		mockAPI.UpdateTaskFunc = func(ctx context.Context, taskID string, opts *interfaces.TaskUpdateOptions) (*clickup.Task, error) {
			closeCalled = true
			return nil, nil
		}

		// Execute
		err = bulkCmd.Execute(context.Background(), []string{"close", "task1"})
		assert.NoError(t, err)

		// Verify close was called
		assert.True(t, closeCalled)

		// Verify confirmation prompt
		assert.Contains(t, stdout.String(), "Are you sure you want to close 1 task(s)?")
	})

	t.Run("close from stdin", func(t *testing.T) {
		// Setup
		mockAPI := &BulkMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()

		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)

		// Create command and cast to BulkCommand
		cmd, err := factory.CreateCommand("bulk")
		require.NoError(t, err)
		bulkCmd := cmd.(*BulkCommand)

		// Set up test input with task IDs from stdin
		stdin := strings.NewReader("task1\ntask2\n")
		bulkCmd.SetStdin(stdin)
		bulkCmd.yes = true

		// Track closed tasks
		var closedTasks []string
		mockAPI.UpdateTaskFunc = func(ctx context.Context, taskID string, opts *interfaces.TaskUpdateOptions) (*clickup.Task, error) {
			if opts.Status == "complete" {
				closedTasks = append(closedTasks, taskID)
			}
			return nil, nil
		}

		// Execute with no args
		err = bulkCmd.Execute(context.Background(), []string{"close"})
		assert.NoError(t, err)

		// Verify tasks were closed
		assert.Equal(t, []string{"task1", "task2"}, closedTasks)
	})
}

func TestBulkCommand_Delete(t *testing.T) {
	t.Run("delete multiple tasks with confirmation", func(t *testing.T) {
		// Setup
		mockAPI := &BulkMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.Set("output", "table")

		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)

		// Track deleted tasks
		var deletedTasks []string
		mockAPI.DeleteTaskFunc = func(ctx context.Context, taskID string) error {
			deletedTasks = append(deletedTasks, taskID)
			return nil
		}

		// Create command and cast to BulkCommand
		cmd, err := factory.CreateCommand("bulk")
		require.NoError(t, err)
		bulkCmd := cmd.(*BulkCommand)

		// Set up test input (user types "delete")
		stdin := strings.NewReader("delete\n")
		stdout := &bytes.Buffer{}
		bulkCmd.SetStdin(stdin)
		bulkCmd.SetStdout(stdout)

		// Execute
		err = bulkCmd.Execute(context.Background(), []string{"delete", "task1", "task2"})
		assert.NoError(t, err)

		// Verify tasks were deleted
		assert.Equal(t, []string{"task1", "task2"}, deletedTasks)

		// Verify warning and confirmation prompt
		assert.Contains(t, mockOutput.WarningMsg[0], "WARNING: This will permanently delete 2 task(s)")
		assert.Contains(t, stdout.String(), "Type 'delete' to confirm:")
	})

	t.Run("delete with --yes flag", func(t *testing.T) {
		// Setup
		mockAPI := &BulkMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()

		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)

		// Track deleted tasks
		var deletedTasks []string
		mockAPI.DeleteTaskFunc = func(ctx context.Context, taskID string) error {
			deletedTasks = append(deletedTasks, taskID)
			return nil
		}

		// Create command
		cmd, err := factory.CreateCommand("bulk")
		require.NoError(t, err)

		// Get cobra command to set flags
		cobraCmd := cmd.GetCobraCommand()
		deleteCmd, _, err := cobraCmd.Find([]string{"delete"})
		require.NoError(t, err)

		// Set flags
		_ = deleteCmd.Flags().Set("yes", "true")

		// Execute
		err = deleteCmd.RunE(deleteCmd, []string{"task1", "task2"})
		assert.NoError(t, err)

		// Verify tasks were deleted without confirmation
		assert.Equal(t, []string{"task1", "task2"}, deletedTasks)
	})

	t.Run("delete cancelled by user", func(t *testing.T) {
		// Setup
		mockAPI := &BulkMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()

		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)

		// Create command and cast to BulkCommand
		cmd, err := factory.CreateCommand("bulk")
		require.NoError(t, err)
		bulkCmd := cmd.(*BulkCommand)

		// Set up test input (user types something other than "delete")
		stdin := strings.NewReader("cancel\n")
		stdout := &bytes.Buffer{}
		bulkCmd.SetStdin(stdin)
		bulkCmd.SetStdout(stdout)

		// Track if delete was called
		deleteCalled := false
		mockAPI.DeleteTaskFunc = func(ctx context.Context, taskID string) error {
			deleteCalled = true
			return nil
		}

		// Execute
		err = bulkCmd.Execute(context.Background(), []string{"delete", "task1"})
		assert.NoError(t, err)

		// Verify delete was NOT called
		assert.False(t, deleteCalled)

		// Verify cancelled message
		assert.Contains(t, mockOutput.InfoMsg, "Cancelled")
	})

	t.Run("delete with output format", func(t *testing.T) {
		// Setup
		mockAPI := &BulkMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.Set("output", "json")

		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)

		// Mock successful deletes
		mockAPI.DeleteTaskFunc = func(ctx context.Context, taskID string) error {
			return nil
		}

		// Create command and set --yes flag
		cmd, err := factory.CreateCommand("bulk")
		require.NoError(t, err)
		bulkCmd := cmd.(*BulkCommand)
		bulkCmd.yes = true

		// Execute
		err = bulkCmd.Execute(context.Background(), []string{"delete", "task1", "task2"})
		assert.NoError(t, err)

		// Verify deleted tasks were output
		assert.Len(t, mockOutput.Printed, 1)
		if deletedTasks, ok := mockOutput.Printed[0].([]string); ok {
			assert.Equal(t, []string{"task1", "task2"}, deletedTasks)
		}
	})

	t.Run("delete with API errors", func(t *testing.T) {
		// Setup
		mockAPI := &BulkMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()

		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)

		// Mock API error for one task
		mockAPI.DeleteTaskFunc = func(ctx context.Context, taskID string) error {
			if taskID == "task2" {
				return fmt.Errorf("API error")
			}
			return nil
		}

		// Create command and set --yes flag
		cmd, err := factory.CreateCommand("bulk")
		require.NoError(t, err)
		bulkCmd := cmd.(*BulkCommand)
		bulkCmd.yes = true

		// Execute
		err = bulkCmd.Execute(context.Background(), []string{"delete", "task1", "task2", "task3"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete 1 task(s)")

		// Verify summary
		// InfoMsg is a slice, need to check all messages
		allInfo := strings.Join(mockOutput.InfoMsg, " ")
		assert.Contains(t, allInfo, "Deleted: 2")
		assert.Contains(t, allInfo, "Failed:  1")
	})
}

func TestBulkCommand_GetCobraCommand(t *testing.T) {
	t.Run("has correct subcommands", func(t *testing.T) {
		// Setup
		factory := New()
		cmd, err := factory.CreateCommand("bulk")
		require.NoError(t, err)

		// Get cobra command
		cobraCmd := cmd.GetCobraCommand()

		// Verify subcommands exist
		assert.True(t, cobraCmd.HasSubCommands())

		// Check update subcommand
		updateCmd, _, err := cobraCmd.Find([]string{"update"})
		require.NoError(t, err)
		assert.Equal(t, "update [task-ids...]", updateCmd.Use)
		assert.NotNil(t, updateCmd.Flags().Lookup("status"))
		assert.NotNil(t, updateCmd.Flags().Lookup("priority"))
		assert.NotNil(t, updateCmd.Flags().Lookup("tag"))
		assert.NotNil(t, updateCmd.Flags().Lookup("add-assignee"))
		assert.NotNil(t, updateCmd.Flags().Lookup("remove-assignee"))
		assert.NotNil(t, updateCmd.Flags().Lookup("yes"))
		assert.NotNil(t, updateCmd.Flags().Lookup("dry-run"))

		// Check close subcommand
		closeCmd, _, err := cobraCmd.Find([]string{"close"})
		require.NoError(t, err)
		assert.Equal(t, "close [task-ids...]", closeCmd.Use)
		assert.NotNil(t, closeCmd.Flags().Lookup("yes"))

		// Check delete subcommand
		deleteCmd, _, err := cobraCmd.Find([]string{"delete"})
		require.NoError(t, err)
		assert.Equal(t, "delete [task-ids...]", deleteCmd.Use)
		assert.NotNil(t, deleteCmd.Flags().Lookup("yes"))
	})
}

// BulkMockAPIClient extends MockAPIClient with bulk-specific functions
type BulkMockAPIClient struct {
	*MockAPIClient
	UpdateTaskFunc func(ctx context.Context, taskID string, opts *interfaces.TaskUpdateOptions) (*clickup.Task, error)
	DeleteTaskFunc func(ctx context.Context, taskID string) error
}

func (m *BulkMockAPIClient) UpdateTask(ctx context.Context, taskID string, opts *interfaces.TaskUpdateOptions) (*clickup.Task, error) {
	if m.UpdateTaskFunc != nil {
		return m.UpdateTaskFunc(ctx, taskID, opts)
	}
	return nil, fmt.Errorf("UpdateTask not implemented")
}

func (m *BulkMockAPIClient) DeleteTask(ctx context.Context, taskID string) error {
	if m.DeleteTaskFunc != nil {
		return m.DeleteTaskFunc(ctx, taskID)
	}
	return fmt.Errorf("DeleteTask not implemented")
}

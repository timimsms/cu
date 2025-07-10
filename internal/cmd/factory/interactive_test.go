package factory

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/raksul/go-clickup/clickup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tim/cu/internal/mocks"
)

func TestInteractiveCommand_Simple(t *testing.T) {
	t.Run("exit from main menu", func(t *testing.T) {
		// Setup
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		
		factory := New(
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Create command
		cmd, err := factory.CreateCommand("interactive")
		require.NoError(t, err)
		require.NotNil(t, cmd)
		
		// Cast to InteractiveCommand and override prompts
		interactiveCmd := cmd.(*InteractiveCommand)
		interactiveCmd.selectPrompt = func(label string, items []string) (int, string, error) {
			if label == "What would you like to do?" {
				return 3, "Exit", nil // Select "Exit"
			}
			return 0, "", fmt.Errorf("unexpected prompt")
		}
		
		// Execute
		err = cmd.Execute(context.Background(), []string{})
		assert.NoError(t, err)
	})

	t.Run("workspace switching message", func(t *testing.T) {
		// Setup
		mockOutput := mocks.NewMockOutputFormatter()
		factory := New(WithOutputFormatter(mockOutput))
		
		// Create command
		cmd, err := factory.CreateCommand("interactive")
		require.NoError(t, err)
		
		// Cast and override prompts
		interactiveCmd := cmd.(*InteractiveCommand)
		callCount := 0
		interactiveCmd.selectPrompt = func(label string, items []string) (int, string, error) {
			callCount++
			if callCount == 1 {
				return 2, "Switch Workspace", nil
			}
			return 3, "Exit", nil
		}
		
		// Execute
		err = cmd.Execute(context.Background(), []string{})
		assert.NoError(t, err)
		assert.Contains(t, mockOutput.WarningMsg[0], "Workspace switching not yet implemented")
	})

	t.Run("prompt error handling", func(t *testing.T) {
		// Setup
		mockOutput := mocks.NewMockOutputFormatter()
		factory := New(WithOutputFormatter(mockOutput))
		
		// Create command
		cmd, err := factory.CreateCommand("interactive")
		require.NoError(t, err)
		
		// Cast and override prompts to return error
		interactiveCmd := cmd.(*InteractiveCommand)
		interactiveCmd.selectPrompt = func(label string, items []string) (int, string, error) {
			return 0, "", errors.New("user cancelled")
		}
		
		// Execute
		err = cmd.Execute(context.Background(), []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "prompt failed")
	})

	// Skip interrupt handling test as it requires API mock
	t.Run("interrupt handling", func(t *testing.T) {
		t.Skip("Requires API mock implementation")
	})

	t.Run("display task details", func(t *testing.T) {
		// Setup
		mockOutput := mocks.NewMockOutputFormatter()
		factory := New(WithOutputFormatter(mockOutput))
		
		// Create command
		cmd, err := factory.CreateCommand("interactive")
		require.NoError(t, err)
		
		interactiveCmd := cmd.(*InteractiveCommand)
		
		// Create test task
		task := clickup.Task{
			ID:          "task123",
			Name:        "Test Task",
			Description: "Test Description",
			Status:      clickup.TaskStatus{Status: "open"},
			Priority: clickup.TaskPriority{
				Priority: "high",
			},
			Assignees: []clickup.User{
				{Username: "user1"},
				{Username: "user2"},
			},
		}
		
		// Override input prompt to just return
		interactiveCmd.inputPrompt = func(label string) (string, error) {
			return "", nil
		}
		
		// Display task details
		interactiveCmd.displayTaskDetails(task)
		
		// Verify output
		assert.Len(t, mockOutput.InfoMsg, 1)
		output := mockOutput.InfoMsg[0]
		assert.Contains(t, output, "task123")
		assert.Contains(t, output, "Test Task")
		assert.Contains(t, output, "Test Description")
		assert.Contains(t, output, "user1, user2")
		assert.Contains(t, output, "High") // Priority should be capitalized
	})

	t.Run("get task priority formatting", func(t *testing.T) {
		factory := New()
		cmd, err := factory.CreateCommand("interactive")
		require.NoError(t, err)
		
		interactiveCmd := cmd.(*InteractiveCommand)
		
		// Test with nil priority
		task := clickup.Task{}
		assert.Equal(t, "Normal", interactiveCmd.getTaskPriority(task))
		
		// Test with various priorities
		testCases := []struct {
			priority string
			expected string
		}{
			{"urgent", "Urgent"},
			{"high", "High"},
			{"normal", "Normal"},
			{"low", "Low"},
			{"unknown", "Normal"},
		}
		
		for _, tc := range testCases {
			task.Priority = clickup.TaskPriority{Priority: tc.priority}
			assert.Equal(t, tc.expected, interactiveCmd.getTaskPriority(task))
		}
	})
}

func TestInteractiveCommand_CobraIntegration(t *testing.T) {
	t.Run("cobra command properties", func(t *testing.T) {
		factory := New()
		cmd, err := factory.CreateCommand("interactive")
		require.NoError(t, err)
		
		cobraCmd := cmd.GetCobraCommand()
		require.NotNil(t, cobraCmd)
		
		assert.Equal(t, "interactive", cobraCmd.Use)
		assert.Equal(t, "Interactive mode for task management", cobraCmd.Short)
		assert.Contains(t, cobraCmd.Long, "Enter interactive mode")
	})
}
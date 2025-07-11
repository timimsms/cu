package factory

import (
	"context"
	"fmt"
	"testing"

	"github.com/raksul/go-clickup/clickup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tim/cu/internal/mocks"
)

func TestSpaceCommand(t *testing.T) {
	t.Run("no subcommand defaults to list", func(t *testing.T) {
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
		
		// Mock API responses
		mockWorkspaces := []clickup.Team{
			{
				ID:   "workspace1",
				Name: "Test Workspace",
			},
		}
		mockSpaces := []clickup.Space{
			{
				ID:   "space1",
				Name: "Test Space",
				Private: false,
				Archived: false,
			},
		}
		
		mockAPI.GetWorkspacesFunc = func(ctx context.Context) ([]clickup.Team, error) {
			return mockWorkspaces, nil
		}
		mockAPI.GetSpacesFunc = func(ctx context.Context, teamID string) ([]clickup.Space, error) {
			assert.Equal(t, "workspace1", teamID)
			return mockSpaces, nil
		}
		
		// Create command
		cmd, err := factory.CreateCommand("space")
		require.NoError(t, err)
		require.NotNil(t, cmd)
		
		// Execute without subcommand (should default to list)
		err = cmd.Execute(context.Background(), []string{})
		assert.NoError(t, err)
		
		// Verify output was called
		assert.Len(t, mockOutput.Printed, 1)
	})

	t.Run("unknown subcommand", func(t *testing.T) {
		// Setup
		factory := New()
		cmd, err := factory.CreateCommand("space")
		require.NoError(t, err)
		
		// Execute with unknown subcommand
		err = cmd.Execute(context.Background(), []string{"unknown"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown subcommand: unknown")
	})
}

func TestSpaceCommand_List(t *testing.T) {
	t.Run("list spaces successfully", func(t *testing.T) {
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
		
		// Mock API responses
		mockWorkspaces := []clickup.Team{
			{
				ID:   "workspace1",
				Name: "Test Workspace",
			},
		}
		mockSpaces := []clickup.Space{
			{
				ID:       "space1",
				Name:     "Public Space",
				Private:  false,
				Archived: false,
			},
			{
				ID:       "space2",
				Name:     "Private Space",
				Private:  true,
				Archived: false,
			},
			{
				ID:       "space3",
				Name:     "Archived Space",
				Private:  false,
				Archived: true,
			},
		}
		
		mockAPI.GetWorkspacesFunc = func(ctx context.Context) ([]clickup.Team, error) {
			return mockWorkspaces, nil
		}
		mockAPI.GetSpacesFunc = func(ctx context.Context, teamID string) ([]clickup.Space, error) {
			assert.Equal(t, "workspace1", teamID)
			return mockSpaces, nil
		}
		
		// Create command
		cmd, err := factory.CreateCommand("space")
		require.NoError(t, err)
		
		// Execute list subcommand
		err = cmd.Execute(context.Background(), []string{"list"})
		assert.NoError(t, err)
		
		// Verify output was called
		assert.Len(t, mockOutput.Printed, 1)
	})

	t.Run("list spaces with json output", func(t *testing.T) {
		// Setup
		mockAPI := &MockAPIClient{}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.Set("output", "json")
		
		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Mock API responses
		mockWorkspaces := []clickup.Team{
			{
				ID:   "workspace1",
				Name: "Test Workspace",
			},
		}
		mockSpaces := []clickup.Space{
			{
				ID:   "space1",
				Name: "Test Space",
			},
		}
		
		mockAPI.GetWorkspacesFunc = func(ctx context.Context) ([]clickup.Team, error) {
			return mockWorkspaces, nil
		}
		mockAPI.GetSpacesFunc = func(ctx context.Context, teamID string) ([]clickup.Space, error) {
			return mockSpaces, nil
		}
		
		// Create command
		cmd, err := factory.CreateCommand("space")
		require.NoError(t, err)
		
		// Execute list subcommand
		err = cmd.Execute(context.Background(), []string{"list"})
		assert.NoError(t, err)
		
		// Verify raw space data was output (json format)
		assert.Len(t, mockOutput.Printed, 1)
	})

	t.Run("list with no workspaces", func(t *testing.T) {
		// Setup
		mockAPI := &MockAPIClient{}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		
		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Mock API response with no workspaces
		mockAPI.GetWorkspacesFunc = func(ctx context.Context) ([]clickup.Team, error) {
			return []clickup.Team{}, nil
		}
		
		// Create command
		cmd, err := factory.CreateCommand("space")
		require.NoError(t, err)
		
		// Execute list
		err = cmd.Execute(context.Background(), []string{"list"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no workspaces found")
	})

	t.Run("list with workspace API error", func(t *testing.T) {
		// Setup
		mockAPI := &MockAPIClient{}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		
		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Mock API error
		mockAPI.GetWorkspacesFunc = func(ctx context.Context) ([]clickup.Team, error) {
			return nil, fmt.Errorf("workspace API error")
		}
		
		// Create command
		cmd, err := factory.CreateCommand("space")
		require.NoError(t, err)
		
		// Execute list
		err = cmd.Execute(context.Background(), []string{"list"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get workspaces")
	})

	t.Run("list with spaces API error", func(t *testing.T) {
		// Setup
		mockAPI := &MockAPIClient{}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		
		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Mock workspace success, spaces error
		mockWorkspaces := []clickup.Team{
			{
				ID:   "workspace1",
				Name: "Test Workspace",
			},
		}
		mockAPI.GetWorkspacesFunc = func(ctx context.Context) ([]clickup.Team, error) {
			return mockWorkspaces, nil
		}
		mockAPI.GetSpacesFunc = func(ctx context.Context, teamID string) ([]clickup.Space, error) {
			return nil, fmt.Errorf("spaces API error")
		}
		
		// Create command
		cmd, err := factory.CreateCommand("space")
		require.NoError(t, err)
		
		// Execute list
		err = cmd.Execute(context.Background(), []string{"list"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get spaces")
	})

	t.Run("list with no API client", func(t *testing.T) {
		// Setup
		factory := New()
		cmd, err := factory.CreateCommand("space")
		require.NoError(t, err)
		
		// Execute list without API client
		err = cmd.Execute(context.Background(), []string{"list"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API client not initialized")
	})
}

func TestSpaceCommand_CobraIntegration(t *testing.T) {
	t.Run("cobra command with subcommands", func(t *testing.T) {
		factory := New()
		cmd, err := factory.CreateCommand("space")
		require.NoError(t, err)
		
		cobraCmd := cmd.GetCobraCommand()
		require.NotNil(t, cobraCmd)
		
		assert.Equal(t, "space", cobraCmd.Use)
		assert.Equal(t, "Manage spaces", cobraCmd.Short)
		
		// Check list subcommand
		listCmd, _, err := cobraCmd.Find([]string{"list"})
		assert.NoError(t, err)
		assert.NotNil(t, listCmd)
		assert.Equal(t, "list", listCmd.Name())
	})
}


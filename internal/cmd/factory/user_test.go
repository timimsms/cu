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

func TestUserCommand(t *testing.T) {
	t.Run("no subcommand defaults to list", func(t *testing.T) {
		// Setup
		mockAPI := &UserMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.Set("output", "table")
		
		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Mock API response
		mockWorkspaces := []clickup.Team{
			{
				ID:   "workspace1",
				Name: "Test Workspace",
			},
		}
		mockUsers := []clickup.TeamUser{
			{
				User: clickup.User{
					ID:       123,
					Username: "testuser",
					Email:    "test@example.com",
				},
				Role: &[]int{1}[0], // Admin role
			},
		}
		
		mockAPI.GetWorkspacesFunc = func(ctx context.Context) ([]clickup.Team, error) {
			return mockWorkspaces, nil
		}
		mockAPI.GetWorkspaceMembersFunc = func(ctx context.Context, workspaceID string) ([]clickup.TeamUser, error) {
			assert.Equal(t, "workspace1", workspaceID)
			return mockUsers, nil
		}
		
		// Create command
		cmd, err := factory.CreateCommand("user")
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
		cmd, err := factory.CreateCommand("user")
		require.NoError(t, err)
		
		// Execute with unknown subcommand
		err = cmd.Execute(context.Background(), []string{"unknown"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown subcommand: unknown")
	})
}

func TestUserCommand_List(t *testing.T) {
	t.Run("list workspace users successfully", func(t *testing.T) {
		// Setup
		mockAPI := &UserMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.Set("output", "table")
		
		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Mock API response
		mockWorkspaces := []clickup.Team{
			{
				ID:   "workspace123",
				Name: "Test Workspace",
			},
		}
		mockUsers := []clickup.TeamUser{
			{
				User: clickup.User{
					ID:       123,
					Username: "alice",
					Email:    "alice@example.com",
				},
				Role: &[]int{1}[0], // Admin role
			},
			{
				User: clickup.User{
					ID:       456,
					Username: "bob",
					Email:    "bob@example.com",
				},
				Role: &[]int{2}[0], // Member role
			},
		}
		
		mockAPI.GetWorkspacesFunc = func(ctx context.Context) ([]clickup.Team, error) {
			return mockWorkspaces, nil
		}
		mockAPI.GetWorkspaceMembersFunc = func(ctx context.Context, workspaceID string) ([]clickup.TeamUser, error) {
			assert.Equal(t, "workspace123", workspaceID)
			return mockUsers, nil
		}
		
		// Create command
		cmd, err := factory.CreateCommand("user")
		require.NoError(t, err)
		
		// Execute list subcommand
		err = cmd.Execute(context.Background(), []string{"list"})
		assert.NoError(t, err)
		
		// Verify output was called
		assert.Len(t, mockOutput.Printed, 1)
		
		// Verify table data structure
		if rows, ok := mockOutput.Printed[0].([]interface{}); ok {
			assert.Len(t, rows, 2) // Two users
		}
	})

	t.Run("list with no workspaces", func(t *testing.T) {
		// Setup
		mockAPI := &UserMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockConfig := mocks.NewMockConfigProvider()
		
		factory := New(
			WithAPIClient(mockAPI),
			WithConfigProvider(mockConfig),
		)
		
		// Mock empty workspaces
		mockAPI.GetWorkspacesFunc = func(ctx context.Context) ([]clickup.Team, error) {
			return []clickup.Team{}, nil
		}
		
		// Create command
		cmd, err := factory.CreateCommand("user")
		require.NoError(t, err)
		
		// Execute list subcommand
		err = cmd.Execute(context.Background(), []string{"list"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no workspaces found")
	})

	t.Run("list with workspace API error", func(t *testing.T) {
		// Setup
		mockAPI := &UserMockAPIClient{MockAPIClient: &MockAPIClient{}}
		factory := New(WithAPIClient(mockAPI))
		
		// Mock API error
		mockAPI.GetWorkspacesFunc = func(ctx context.Context) ([]clickup.Team, error) {
			return nil, fmt.Errorf("API error")
		}
		
		// Create command
		cmd, err := factory.CreateCommand("user")
		require.NoError(t, err)
		
		// Execute list subcommand
		err = cmd.Execute(context.Background(), []string{"list"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get workspaces")
	})

	t.Run("list with members API error", func(t *testing.T) {
		// Setup
		mockAPI := &UserMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockConfig := mocks.NewMockConfigProvider()
		
		factory := New(
			WithAPIClient(mockAPI),
			WithConfigProvider(mockConfig),
		)
		
		// Mock workspaces success but members error
		mockAPI.GetWorkspacesFunc = func(ctx context.Context) ([]clickup.Team, error) {
			return []clickup.Team{{ID: "workspace1", Name: "Test"}}, nil
		}
		mockAPI.GetWorkspaceMembersFunc = func(ctx context.Context, workspaceID string) ([]clickup.TeamUser, error) {
			return nil, fmt.Errorf("members API error")
		}
		
		// Create command
		cmd, err := factory.CreateCommand("user")
		require.NoError(t, err)
		
		// Execute list subcommand
		err = cmd.Execute(context.Background(), []string{"list"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get workspace members")
	})

	t.Run("list with JSON output", func(t *testing.T) {
		// Setup
		mockAPI := &UserMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.Set("output", "json")
		
		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Mock API response
		mockWorkspaces := []clickup.Team{
			{
				ID:   "workspace1",
				Name: "Test Workspace",
			},
		}
		mockUsers := []clickup.TeamUser{
			{
				User: clickup.User{
					ID:       123,
					Username: "testuser",
					Email:    "test@example.com",
				},
				Role: &[]int{1}[0],
			},
		}
		
		mockAPI.GetWorkspacesFunc = func(ctx context.Context) ([]clickup.Team, error) {
			return mockWorkspaces, nil
		}
		mockAPI.GetWorkspaceMembersFunc = func(ctx context.Context, workspaceID string) ([]clickup.TeamUser, error) {
			return mockUsers, nil
		}
		
		// Create command
		cmd, err := factory.CreateCommand("user")
		require.NoError(t, err)
		
		// Execute list subcommand
		err = cmd.Execute(context.Background(), []string{"list"})
		assert.NoError(t, err)
		
		// Verify raw user data was output (not table rows)
		assert.Len(t, mockOutput.Printed, 1)
		if users, ok := mockOutput.Printed[0].([]clickup.TeamUser); ok {
			assert.Len(t, users, 1)
			assert.Equal(t, 123, users[0].User.ID)
		}
	})

	t.Run("list with users without roles", func(t *testing.T) {
		// Setup
		mockAPI := &UserMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.Set("output", "table")
		
		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Mock API response with nil roles
		mockWorkspaces := []clickup.Team{
			{
				ID:   "workspace1",
				Name: "Test Workspace",
			},
		}
		mockUsers := []clickup.TeamUser{
			{
				User: clickup.User{
					ID:       123,
					Username: "testuser",
					Email:    "test@example.com",
				},
				Role: nil, // No role assigned
			},
		}
		
		mockAPI.GetWorkspacesFunc = func(ctx context.Context) ([]clickup.Team, error) {
			return mockWorkspaces, nil
		}
		mockAPI.GetWorkspaceMembersFunc = func(ctx context.Context, workspaceID string) ([]clickup.TeamUser, error) {
			return mockUsers, nil
		}
		
		// Create command
		cmd, err := factory.CreateCommand("user")
		require.NoError(t, err)
		
		// Execute list subcommand
		err = cmd.Execute(context.Background(), []string{"list"})
		assert.NoError(t, err)
		
		// Verify output was called without error
		assert.Len(t, mockOutput.Printed, 1)
	})

	t.Run("list with API client not initialized", func(t *testing.T) {
		// Setup
		factory := New() // No API client
		cmd, err := factory.CreateCommand("user")
		require.NoError(t, err)
		
		// Execute list subcommand
		err = cmd.Execute(context.Background(), []string{"list"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API client not initialized")
	})
}

func TestUserCommand_GetCobraCommand(t *testing.T) {
	t.Run("has correct subcommands", func(t *testing.T) {
		// Setup
		factory := New()
		cmd, err := factory.CreateCommand("user")
		require.NoError(t, err)
		
		// Get cobra command
		cobraCmd := cmd.GetCobraCommand()
		
		// Verify subcommands exist
		assert.True(t, cobraCmd.HasSubCommands())
		
		// Check list subcommand
		listCmd, _, err := cobraCmd.Find([]string{"list"})
		require.NoError(t, err)
		assert.Equal(t, "list", listCmd.Use)
		assert.Equal(t, "List workspace users", listCmd.Short)
	})
}

// UserMockAPIClient extends MockAPIClient with user-specific functions
type UserMockAPIClient struct {
	*MockAPIClient
	GetWorkspacesFunc        func(ctx context.Context) ([]clickup.Team, error)
	GetWorkspaceMembersFunc  func(ctx context.Context, workspaceID string) ([]clickup.TeamUser, error)
}

func (m *UserMockAPIClient) GetWorkspaces(ctx context.Context) ([]clickup.Team, error) {
	if m.GetWorkspacesFunc != nil {
		return m.GetWorkspacesFunc(ctx)
	}
	return nil, fmt.Errorf("GetWorkspaces not implemented")
}

func (m *UserMockAPIClient) GetWorkspaceMembers(ctx context.Context, workspaceID string) ([]clickup.TeamUser, error) {
	if m.GetWorkspaceMembersFunc != nil {
		return m.GetWorkspaceMembersFunc(ctx, workspaceID)
	}
	return nil, fmt.Errorf("GetWorkspaceMembers not implemented")
}
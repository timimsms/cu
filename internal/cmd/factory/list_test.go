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

func TestListCommand(t *testing.T) {
	t.Run("no subcommand defaults to list", func(t *testing.T) {
		// Setup
		mockAPI := &ListMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.Set("output", "table")
		
		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Mock API response
		mockLists := []clickup.List{
			{
				ID:        "list1",
				Name:      "Test List 1",
				Archived:  false,
				TaskCount: 5,
			},
		}
		mockAPI.GetFolderlessListsFunc = func(ctx context.Context, spaceID string) ([]clickup.List, error) {
			return mockLists, nil
		}
		mockAPI.GetFoldersFunc = func(ctx context.Context, spaceID string) ([]clickup.Folder, error) {
			return []clickup.Folder{}, nil
		}
		
		// Create command
		cmd, err := factory.CreateCommand("list")
		require.NoError(t, err)
		require.NotNil(t, cmd)
		
		// Get cobra command to set flags
		cobraCmd := cmd.GetCobraCommand()
		listCmd, _, err := cobraCmd.Find([]string{"list"})
		require.NoError(t, err)
		
		// Set flags
		listCmd.Flags().Set("space", "space123")
		
		// Execute without subcommand (should default to list)
		err = listCmd.RunE(listCmd, []string{})
		assert.NoError(t, err)
		
		// Verify output was called
		assert.Len(t, mockOutput.Printed, 1)
	})

	t.Run("unknown subcommand", func(t *testing.T) {
		// Setup
		factory := New()
		cmd, err := factory.CreateCommand("list")
		require.NoError(t, err)
		
		// Execute with unknown subcommand
		err = cmd.Execute(context.Background(), []string{"unknown"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown subcommand: unknown")
	})
}

func TestListCommand_List(t *testing.T) {
	t.Run("list folderless lists successfully", func(t *testing.T) {
		// Setup
		mockAPI := &ListMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.Set("default_list", "list1")
		mockConfig.Set("output", "table")
		
		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Mock API response
		mockLists := []clickup.List{
			{
				ID:        "list1",
				Name:      "Test List 1",
				Archived:  false,
				TaskCount: 5,
			},
			{
				ID:        "list2",
				Name:      "Test List 2",
				Archived:  true,
				TaskCount: 2,
			},
		}
		mockAPI.GetFolderlessListsFunc = func(ctx context.Context, spaceID string) ([]clickup.List, error) {
			assert.Equal(t, "space123", spaceID)
			return mockLists, nil
		}
		mockAPI.GetFoldersFunc = func(ctx context.Context, spaceID string) ([]clickup.Folder, error) {
			return []clickup.Folder{}, nil
		}
		
		// Create command
		cmd, err := factory.CreateCommand("list")
		require.NoError(t, err)
		
		// Get cobra command to set flags
		cobraCmd := cmd.GetCobraCommand()
		listCmd, _, err := cobraCmd.Find([]string{"list"})
		require.NoError(t, err)
		
		// Set flags
		listCmd.Flags().Set("space", "space123")
		
		// Execute list subcommand
		err = listCmd.RunE(listCmd, []string{})
		assert.NoError(t, err)
		
		// Verify output was called
		assert.Len(t, mockOutput.Printed, 1)
	})

	t.Run("list from folder successfully", func(t *testing.T) {
		// Setup
		mockAPI := &ListMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.Set("output", "table")
		
		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Mock API response
		mockLists := []clickup.List{
			{
				ID:        "list1",
				Name:      "Folder List 1",
				Archived:  false,
				TaskCount: 3,
			},
		}
		mockAPI.GetListsFunc = func(ctx context.Context, folderID string) ([]clickup.List, error) {
			assert.Equal(t, "folder123", folderID)
			return mockLists, nil
		}
		
		// Create command
		cmd, err := factory.CreateCommand("list")
		require.NoError(t, err)
		
		// Get cobra command to set flags
		cobraCmd := cmd.GetCobraCommand()
		listCmd, _, err := cobraCmd.Find([]string{"list"})
		require.NoError(t, err)
		
		// Set flags
		listCmd.Flags().Set("folder", "folder123")
		
		// Execute list subcommand
		err = listCmd.RunE(listCmd, []string{})
		assert.NoError(t, err)
		
		// Verify output was called
		assert.Len(t, mockOutput.Printed, 1)
	})

	t.Run("list with archived filter", func(t *testing.T) {
		// Setup
		mockAPI := &ListMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.Set("output", "table")
		
		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Mock API response with mixed archived/active lists
		mockLists := []clickup.List{
			{
				ID:        "list1",
				Name:      "Active List",
				Archived:  false,
				TaskCount: 3,
			},
			{
				ID:        "list2",
				Name:      "Archived List",
				Archived:  true,
				TaskCount: 1,
			},
		}
		mockAPI.GetFolderlessListsFunc = func(ctx context.Context, spaceID string) ([]clickup.List, error) {
			return mockLists, nil
		}
		mockAPI.GetFoldersFunc = func(ctx context.Context, spaceID string) ([]clickup.Folder, error) {
			return []clickup.Folder{}, nil
		}
		
		// Create command
		cmd, err := factory.CreateCommand("list")
		require.NoError(t, err)
		
		// Get cobra command to set flags
		cobraCmd := cmd.GetCobraCommand()
		listCmd, _, err := cobraCmd.Find([]string{"list"})
		require.NoError(t, err)
		
		// Set flags to include archived
		listCmd.Flags().Set("space", "space123")
		listCmd.Flags().Set("archived", "true")
		
		// Execute list subcommand
		err = listCmd.RunE(listCmd, []string{})
		assert.NoError(t, err)
		
		// Verify output was called
		assert.Len(t, mockOutput.Printed, 1)
	})

	t.Run("list with space and folders", func(t *testing.T) {
		// Setup
		mockAPI := &ListMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.Set("output", "table")
		
		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Mock folderless lists
		mockFolderlessLists := []clickup.List{
			{
				ID:        "list1",
				Name:      "Folderless List",
				Archived:  false,
				TaskCount: 2,
			},
		}
		
		// Mock folders
		mockFolders := []clickup.Folder{
			{
				ID:   "folder1",
				Name: "Test Folder",
			},
		}
		
		// Mock folder lists
		mockFolderLists := []clickup.List{
			{
				ID:        "list2",
				Name:      "Folder List",
				Archived:  false,
				TaskCount: 4,
			},
		}
		
		mockAPI.GetFolderlessListsFunc = func(ctx context.Context, spaceID string) ([]clickup.List, error) {
			return mockFolderlessLists, nil
		}
		mockAPI.GetFoldersFunc = func(ctx context.Context, spaceID string) ([]clickup.Folder, error) {
			return mockFolders, nil
		}
		mockAPI.GetListsFunc = func(ctx context.Context, folderID string) ([]clickup.List, error) {
			return mockFolderLists, nil
		}
		
		// Create command
		cmd, err := factory.CreateCommand("list")
		require.NoError(t, err)
		
		// Get cobra command to set flags
		cobraCmd := cmd.GetCobraCommand()
		listCmd, _, err := cobraCmd.Find([]string{"list"})
		require.NoError(t, err)
		
		// Set flags
		listCmd.Flags().Set("space", "space123")
		
		// Execute list subcommand
		err = listCmd.RunE(listCmd, []string{})
		assert.NoError(t, err)
		
		// Verify output was called
		assert.Len(t, mockOutput.Printed, 1)
	})

	t.Run("list with no space or folder", func(t *testing.T) {
		// Setup
		mockAPI := &ListMockAPIClient{MockAPIClient: &MockAPIClient{}}
		factory := New(WithAPIClient(mockAPI))
		
		// Create command
		cmd, err := factory.CreateCommand("list")
		require.NoError(t, err)
		
		// Get cobra command
		cobraCmd := cmd.GetCobraCommand()
		listCmd, _, err := cobraCmd.Find([]string{"list"})
		require.NoError(t, err)
		
		// Execute without space or folder flags
		err = listCmd.RunE(listCmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "please specify either --space or --folder")
	})

	t.Run("list with API error", func(t *testing.T) {
		// Setup
		mockAPI := &ListMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockConfig := mocks.NewMockConfigProvider()
		
		factory := New(
			WithAPIClient(mockAPI),
			WithConfigProvider(mockConfig),
		)
		
		// Mock API error
		mockAPI.GetFolderlessListsFunc = func(ctx context.Context, spaceID string) ([]clickup.List, error) {
			return nil, fmt.Errorf("API error")
		}
		
		// Create command
		cmd, err := factory.CreateCommand("list")
		require.NoError(t, err)
		
		// Get cobra command to set flags
		cobraCmd := cmd.GetCobraCommand()
		listCmd, _, err := cobraCmd.Find([]string{"list"})
		require.NoError(t, err)
		
		// Set flags
		listCmd.Flags().Set("space", "space123")
		
		// Execute
		err = listCmd.RunE(listCmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get folderless lists")
	})

	t.Run("list with folder API error continues", func(t *testing.T) {
		// Setup
		mockAPI := &ListMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.Set("output", "table")
		
		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Mock folderless lists success
		mockAPI.GetFolderlessListsFunc = func(ctx context.Context, spaceID string) ([]clickup.List, error) {
			return []clickup.List{{ID: "list1", Name: "Folderless List"}}, nil
		}
		
		// Mock folders success
		mockAPI.GetFoldersFunc = func(ctx context.Context, spaceID string) ([]clickup.Folder, error) {
			return []clickup.Folder{{ID: "folder1", Name: "Test Folder"}}, nil
		}
		
		// Mock folder lists error
		mockAPI.GetListsFunc = func(ctx context.Context, folderID string) ([]clickup.List, error) {
			return nil, fmt.Errorf("folder API error")
		}
		
		// Create command
		cmd, err := factory.CreateCommand("list")
		require.NoError(t, err)
		
		// Get cobra command to set flags
		cobraCmd := cmd.GetCobraCommand()
		listCmd, _, err := cobraCmd.Find([]string{"list"})
		require.NoError(t, err)
		
		// Set flags
		listCmd.Flags().Set("space", "space123")
		
		// Execute - should succeed despite folder error
		err = listCmd.RunE(listCmd, []string{})
		assert.NoError(t, err)
		
		// Verify warning was printed
		assert.Len(t, mockOutput.WarningMsg, 1)
		assert.Contains(t, mockOutput.WarningMsg[0], "Failed to get lists from folder Test Folder")
	})

	t.Run("list with JSON output", func(t *testing.T) {
		// Setup
		mockAPI := &ListMockAPIClient{MockAPIClient: &MockAPIClient{}}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.Set("output", "json")
		
		factory := New(
			WithAPIClient(mockAPI),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Mock API response
		mockLists := []clickup.List{
			{
				ID:        "list1",
				Name:      "Test List",
				Archived:  false,
				TaskCount: 5,
			},
		}
		mockAPI.GetFolderlessListsFunc = func(ctx context.Context, spaceID string) ([]clickup.List, error) {
			return mockLists, nil
		}
		mockAPI.GetFoldersFunc = func(ctx context.Context, spaceID string) ([]clickup.Folder, error) {
			return []clickup.Folder{}, nil
		}
		
		// Create command
		cmd, err := factory.CreateCommand("list")
		require.NoError(t, err)
		
		// Get cobra command to set flags
		cobraCmd := cmd.GetCobraCommand()
		listCmd, _, err := cobraCmd.Find([]string{"list"})
		require.NoError(t, err)
		
		// Set flags
		listCmd.Flags().Set("space", "space123")
		
		// Execute list subcommand
		err = listCmd.RunE(listCmd, []string{})
		assert.NoError(t, err)
		
		// Verify raw list data was output (not table rows)
		assert.Len(t, mockOutput.Printed, 1)
		// Should be the raw lists, not processed table rows
		if lists, ok := mockOutput.Printed[0].([]clickup.List); ok {
			assert.Len(t, lists, 1)
			assert.Equal(t, "list1", lists[0].ID)
		}
	})
}

func TestListCommand_Default(t *testing.T) {
	t.Run("set default list globally", func(t *testing.T) {
		// Setup
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		
		factory := New(
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Create command
		cmd, err := factory.CreateCommand("list")
		require.NoError(t, err)
		
		// Execute default subcommand
		err = cmd.Execute(context.Background(), []string{"default", "list123"})
		assert.NoError(t, err)
		
		// Verify config was set
		assert.Equal(t, "list123", mockConfig.GetString("default_list"))
		
		// Verify success message
		assert.Len(t, mockOutput.SuccessMsg, 1)
		assert.Contains(t, mockOutput.SuccessMsg[0], "Default list set to: list123 (global)")
		
		// Verify info message
		assert.Len(t, mockOutput.InfoMsg, 1)
		assert.Contains(t, mockOutput.InfoMsg[0], "Use --project flag")
	})

	t.Run("set default list with project flag", func(t *testing.T) {
		// Setup
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		
		factory := New(
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Create command
		cmd, err := factory.CreateCommand("list")
		require.NoError(t, err)
		
		// Get cobra command to set flags
		cobraCmd := cmd.GetCobraCommand()
		defaultCmd, _, err := cobraCmd.Find([]string{"default"})
		require.NoError(t, err)
		
		// Set project flag
		defaultCmd.Flags().Set("project", "true")
		
		// Execute
		err = defaultCmd.RunE(defaultCmd, []string{"list456"})
		assert.NoError(t, err)
		
		// Verify success message for project config
		assert.Len(t, mockOutput.SuccessMsg, 1)
		assert.Contains(t, mockOutput.SuccessMsg[0], "Default list set to: list456")
		
		// Verify info message about config path
		assert.Len(t, mockOutput.InfoMsg, 1)
		assert.Contains(t, mockOutput.InfoMsg[0], "Saved to project config")
	})

	t.Run("set default list with existing project config", func(t *testing.T) {
		// Setup
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := &mocks.MockConfigWithProject{
			MockConfigProvider: mocks.NewMockConfigProvider(),
		}
		mockConfig.HasProjectConfigVal = true
		
		factory := New(
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Create command
		cmd, err := factory.CreateCommand("list")
		require.NoError(t, err)
		
		// Execute default subcommand
		err = cmd.Execute(context.Background(), []string{"default", "list789"})
		assert.NoError(t, err)
		
		// Verify project config was used
		assert.True(t, mockConfig.ProjectConfigSaved)
		assert.Equal(t, "list789", mockConfig.ProjectSettings["default_list"])
		
		// Verify success message for project config
		assert.Len(t, mockOutput.SuccessMsg, 1)
		assert.Contains(t, mockOutput.SuccessMsg[0], "Default list set to: list789")
	})

	t.Run("set default list with no ID", func(t *testing.T) {
		// Setup
		factory := New()
		cmd, err := factory.CreateCommand("list")
		require.NoError(t, err)
		
		// Execute default without list ID
		err = cmd.Execute(context.Background(), []string{"default"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "list ID is required")
	})

	t.Run("set default list with project config error", func(t *testing.T) {
		// Setup
		mockConfig := &mocks.MockConfigWithProject{
			MockConfigProvider: mocks.NewMockConfigProvider(),
		}
		mockConfig.HasProjectConfigVal = true
		mockConfig.SaveProjectConfigErr = fmt.Errorf("save error")
		
		factory := New(WithConfigProvider(mockConfig))
		
		// Create command
		cmd, err := factory.CreateCommand("list")
		require.NoError(t, err)
		
		// Execute default subcommand
		err = cmd.Execute(context.Background(), []string{"default", "list999"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save project configuration")
	})

	t.Run("set default list with global config save error", func(t *testing.T) {
		// Setup
		mockConfig := &mocks.MockConfigWithSaveError{
			MockConfigProvider: mocks.NewMockConfigProvider(),
		}
		mockConfig.SaveErr = fmt.Errorf("save error")
		
		factory := New(WithConfigProvider(mockConfig))
		
		// Create command
		cmd, err := factory.CreateCommand("list")
		require.NoError(t, err)
		
		// Execute default subcommand
		err = cmd.Execute(context.Background(), []string{"default", "list999"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save configuration")
	})
}

func TestListCommand_GetCobraCommand(t *testing.T) {
	t.Run("has correct subcommands", func(t *testing.T) {
		// Setup
		factory := New()
		cmd, err := factory.CreateCommand("list")
		require.NoError(t, err)
		
		// Get cobra command
		cobraCmd := cmd.GetCobraCommand()
		
		// Verify subcommands exist
		assert.True(t, cobraCmd.HasSubCommands())
		
		// Check list subcommand
		listCmd, _, err := cobraCmd.Find([]string{"list"})
		require.NoError(t, err)
		assert.Equal(t, "list", listCmd.Use)
		
		// Check default subcommand
		defaultCmd, _, err := cobraCmd.Find([]string{"default"})
		require.NoError(t, err)
		assert.Equal(t, "default <list-id>", defaultCmd.Use)
		
		// Verify flags
		assert.True(t, listCmd.Flags().HasFlag("space"))
		assert.True(t, listCmd.Flags().HasFlag("folder"))
		assert.True(t, listCmd.Flags().HasFlag("archived"))
		assert.True(t, defaultCmd.Flags().HasFlag("project"))
	})
}

// Extend MockAPIClient with list-specific functions
type ListMockAPIClient struct {
	*MockAPIClient
	GetFoldersFunc        func(ctx context.Context, spaceID string) ([]clickup.Folder, error)
	GetListsFunc          func(ctx context.Context, folderID string) ([]clickup.List, error)
	GetFolderlessListsFunc func(ctx context.Context, spaceID string) ([]clickup.List, error)
}

func (m *ListMockAPIClient) GetFolders(ctx context.Context, spaceID string) ([]clickup.Folder, error) {
	if m.GetFoldersFunc != nil {
		return m.GetFoldersFunc(ctx, spaceID)
	}
	return nil, fmt.Errorf("GetFolders not implemented")
}

func (m *ListMockAPIClient) GetLists(ctx context.Context, folderID string) ([]clickup.List, error) {
	if m.GetListsFunc != nil {
		return m.GetListsFunc(ctx, folderID)
	}
	return nil, fmt.Errorf("GetLists not implemented")
}

func (m *ListMockAPIClient) GetFolderlessLists(ctx context.Context, spaceID string) ([]clickup.List, error) {
	if m.GetFolderlessListsFunc != nil {
		return m.GetFolderlessListsFunc(ctx, spaceID)
	}
	return nil, fmt.Errorf("GetFolderlessLists not implemented")
}
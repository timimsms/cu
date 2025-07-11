package factory

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tim/cu/internal/mocks"
)

func TestConfigCommand(t *testing.T) {
	t.Run("no subcommand shows error", func(t *testing.T) {
		// Setup
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		
		factory := New(
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Create command
		cmd, err := factory.CreateCommand("config")
		require.NoError(t, err)
		require.NotNil(t, cmd)
		
		// Execute without subcommand
		err = cmd.Execute(context.Background(), []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no subcommand specified")
	})

	t.Run("unknown subcommand", func(t *testing.T) {
		// Setup
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		
		factory := New(
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Create command
		cmd, err := factory.CreateCommand("config")
		require.NoError(t, err)
		
		// Execute with unknown subcommand
		err = cmd.Execute(context.Background(), []string{"unknown"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown subcommand: unknown")
	})
}

func TestConfigCommand_List(t *testing.T) {
	t.Run("list all settings", func(t *testing.T) {
		// Setup
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		
		// Mock config values
		mockConfig.Set("output", "table")
		mockConfig.Set("default_list", "list123")
		mockConfig.Set("debug", true)
		mockConfig.Set("test_number", 42)
		
		factory := New(
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Create command
		cmd, err := factory.CreateCommand("config")
		require.NoError(t, err)
		
		// Execute list subcommand
		err = cmd.Execute(context.Background(), []string{"list"})
		assert.NoError(t, err)
		
		// Verify output
		assert.Len(t, mockOutput.InfoMsg, 1)
		output := mockOutput.InfoMsg[0]
		
		// Check all settings are present
		assert.Contains(t, output, "output=table")
		assert.Contains(t, output, "default_list=list123")
		assert.Contains(t, output, "debug=true")
		assert.Contains(t, output, "test_number=42")
	})

	t.Run("empty settings", func(t *testing.T) {
		// Setup
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		// Empty config
		
		factory := New(
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Create command
		cmd, err := factory.CreateCommand("config")
		require.NoError(t, err)
		
		// Execute
		err = cmd.Execute(context.Background(), []string{"list"})
		assert.NoError(t, err)
		
		// Should output empty string
		assert.Len(t, mockOutput.InfoMsg, 1)
		assert.Equal(t, "", mockOutput.InfoMsg[0])
	})
}

func TestConfigCommand_Get(t *testing.T) {
	t.Run("get existing key", func(t *testing.T) {
		// Setup
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.Set("test_key", "test_value")
		
		factory := New(
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Create command
		cmd, err := factory.CreateCommand("config")
		require.NoError(t, err)
		
		// Execute
		err = cmd.Execute(context.Background(), []string{"get", "test_key"})
		assert.NoError(t, err)
		
		// Verify output
		assert.Len(t, mockOutput.InfoMsg, 1)
		assert.Equal(t, "test_value", mockOutput.InfoMsg[0])
	})

	t.Run("get non-existent key", func(t *testing.T) {
		// Setup
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		
		factory := New(
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Create command
		cmd, err := factory.CreateCommand("config")
		require.NoError(t, err)
		
		// Execute
		err = cmd.Execute(context.Background(), []string{"get", "nonexistent"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "configuration key 'nonexistent' not found")
	})

	t.Run("get with no args", func(t *testing.T) {
		// Setup
		factory := New()
		cmd, err := factory.CreateCommand("config")
		require.NoError(t, err)
		
		// Execute
		err = cmd.Execute(context.Background(), []string{"get"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exactly one argument required")
	})
}

func TestConfigCommand_Set(t *testing.T) {
	t.Run("set string value", func(t *testing.T) {
		// Setup
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		
		factory := New(
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Create command
		cmd, err := factory.CreateCommand("config")
		require.NoError(t, err)
		
		// Execute
		err = cmd.Execute(context.Background(), []string{"set", "key1", "value1"})
		assert.NoError(t, err)
		
		// Verify
		assert.Equal(t, "value1", mockConfig.Get("key1"))
		assert.Contains(t, mockOutput.SuccessMsg[0], "Set key1 to value1")
	})

	t.Run("set boolean true", func(t *testing.T) {
		// Setup
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		
		factory := New(
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Create command
		cmd, err := factory.CreateCommand("config")
		require.NoError(t, err)
		
		// Execute
		err = cmd.Execute(context.Background(), []string{"set", "debug", "true"})
		assert.NoError(t, err)
		
		// Verify
		assert.Equal(t, true, mockConfig.Get("debug"))
		assert.Contains(t, mockOutput.SuccessMsg[0], "Set debug to true")
	})

	t.Run("set boolean false", func(t *testing.T) {
		// Setup
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		
		factory := New(
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Create command
		cmd, err := factory.CreateCommand("config")
		require.NoError(t, err)
		
		// Execute
		err = cmd.Execute(context.Background(), []string{"set", "debug", "FALSE"})
		assert.NoError(t, err)
		
		// Verify - should be lowercase
		assert.Equal(t, false, mockConfig.Get("debug"))
		assert.Contains(t, mockOutput.SuccessMsg[0], "Set debug to FALSE")
	})

	t.Run("set with save support", func(t *testing.T) {
		// Setup
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := &MockConfigWithSave{
			MockConfigProvider: mocks.NewMockConfigProvider(),
		}
		
		factory := New(
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Create command
		cmd, err := factory.CreateCommand("config")
		require.NoError(t, err)
		
		// Execute
		err = cmd.Execute(context.Background(), []string{"set", "key", "value"})
		assert.NoError(t, err)
		
		// Verify save was called
		assert.True(t, mockConfig.SaveCalled)
	})

	t.Run("set with save error", func(t *testing.T) {
		// Setup
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := &MockConfigWithSave{
			MockConfigProvider: mocks.NewMockConfigProvider(),
			SaveError:          fmt.Errorf("save failed"),
		}
		
		factory := New(
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Create command
		cmd, err := factory.CreateCommand("config")
		require.NoError(t, err)
		
		// Execute
		err = cmd.Execute(context.Background(), []string{"set", "key", "value"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save configuration")
	})

	t.Run("set with wrong args", func(t *testing.T) {
		// Setup
		factory := New()
		cmd, err := factory.CreateCommand("config")
		require.NoError(t, err)
		
		// Execute with one arg
		err = cmd.Execute(context.Background(), []string{"set", "key"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exactly two arguments required")
	})
}

func TestConfigCommand_Init(t *testing.T) {
	t.Run("init new project config", func(t *testing.T) {
		// Setup
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := &MockConfigWithProject{
			MockConfigProvider:  mocks.NewMockConfigProvider(),
			HasProjectConfigVal: false,
		}
		
		factory := New(
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Create command
		cmd, err := factory.CreateCommand("config")
		require.NoError(t, err)
		
		// Execute
		err = cmd.Execute(context.Background(), []string{"init"})
		assert.NoError(t, err)
		
		// Verify
		assert.True(t, mockConfig.InitProjectConfigCalled)
		assert.Contains(t, mockOutput.SuccessMsg[0], "Initialized project configuration")
		assert.Contains(t, mockOutput.InfoMsg[0], "project-specific settings")
	})

	t.Run("init with existing config", func(t *testing.T) {
		// Setup
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := &MockConfigWithProject{
			MockConfigProvider:     mocks.NewMockConfigProvider(),
			HasProjectConfigVal:    true,
			ProjectConfigPath:      "/path/to/.cu.yml",
		}
		
		factory := New(
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Create command
		cmd, err := factory.CreateCommand("config")
		require.NoError(t, err)
		
		// Execute
		err = cmd.Execute(context.Background(), []string{"init"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "project config already exists at: /path/to/.cu.yml")
	})

	t.Run("init without support", func(t *testing.T) {
		// Setup
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		
		factory := New(
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Create command
		cmd, err := factory.CreateCommand("config")
		require.NoError(t, err)
		
		// Execute
		err = cmd.Execute(context.Background(), []string{"init"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "project config initialization not supported")
	})
}

func TestConfigCommand_Show(t *testing.T) {
	t.Run("show global config only", func(t *testing.T) {
		// Setup
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.Set("default_space", "space123")
		mockConfig.Set("default_list", "list456")
		mockConfig.Set("output", "table")
		mockConfig.Set("debug", true)
		
		factory := New(
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Create command
		cmd, err := factory.CreateCommand("config")
		require.NoError(t, err)
		
		// Execute
		err = cmd.Execute(context.Background(), []string{"show"})
		assert.NoError(t, err)
		
		// Verify output
		assert.Len(t, mockOutput.InfoMsg, 1)
		output := mockOutput.InfoMsg[0]
		assert.Contains(t, output, "Global Configuration")
		assert.Contains(t, output, "default_space: space123")
		assert.Contains(t, output, "default_list: list456")
		assert.Contains(t, output, "debug: true")
	})

	t.Run("show with project config", func(t *testing.T) {
		// Setup
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := &MockConfigWithProject{
			MockConfigProvider:  mocks.NewMockConfigProvider(),
			HasProjectConfigVal: true,
			ProjectConfigPath:   "/project/.cu.yml",
		}
		mockConfig.Set("default_space", "space123")
		mockConfig.Set("output", "table")
		
		factory := New(
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Create command
		cmd, err := factory.CreateCommand("config")
		require.NoError(t, err)
		
		// Execute
		err = cmd.Execute(context.Background(), []string{"show"})
		assert.NoError(t, err)
		
		// Verify output
		assert.Len(t, mockOutput.InfoMsg, 1)
		output := mockOutput.InfoMsg[0]
		assert.Contains(t, output, "Global Configuration")
		assert.Contains(t, output, "Project Configuration")
		assert.Contains(t, output, "config_path: /project/.cu.yml")
	})

	t.Run("show with json format", func(t *testing.T) {
		// Setup
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.Set("output", "json")
		mockConfig.Set("default_space", "space123")
		
		factory := New(
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Create command
		cmd, err := factory.CreateCommand("config")
		require.NoError(t, err)
		
		// Execute
		err = cmd.Execute(context.Background(), []string{"show"})
		assert.NoError(t, err)
		
		// Verify structured output was called
		assert.Len(t, mockOutput.Printed, 1)
		data := mockOutput.Printed[0].(map[string]interface{})
		assert.Contains(t, data, "global")
	})
}

func TestConfigCommand_CobraIntegration(t *testing.T) {
	t.Run("cobra command with subcommands", func(t *testing.T) {
		factory := New()
		cmd, err := factory.CreateCommand("config")
		require.NoError(t, err)
		
		cobraCmd := cmd.GetCobraCommand()
		require.NotNil(t, cobraCmd)
		
		assert.Equal(t, "config", cobraCmd.Use)
		assert.Equal(t, "Manage cu configuration", cobraCmd.Short)
		
		// Check subcommands
		subcommands := []string{"list", "get", "set", "init", "show"}
		for _, sub := range subcommands {
			subCmd, _, err := cobraCmd.Find([]string{sub})
			assert.NoError(t, err)
			assert.NotNil(t, subCmd)
			assert.Equal(t, sub, subCmd.Name())
		}
	})
}

// Mock implementations for testing

type MockConfigWithSave struct {
	*mocks.MockConfigProvider
	SaveCalled bool
	SaveError  error
}

func (m *MockConfigWithSave) Save() error {
	m.SaveCalled = true
	return m.SaveError
}

type MockConfigWithProject struct {
	*mocks.MockConfigProvider
	HasProjectConfigVal     bool
	ProjectConfigPath       string
	InitProjectConfigCalled bool
	InitProjectConfigError  error
}

func (m *MockConfigWithProject) HasProjectConfig() bool {
	return m.HasProjectConfigVal
}

func (m *MockConfigWithProject) GetProjectConfigPath() string {
	return m.ProjectConfigPath
}

func (m *MockConfigWithProject) InitProjectConfig() error {
	m.InitProjectConfigCalled = true
	return m.InitProjectConfigError
}
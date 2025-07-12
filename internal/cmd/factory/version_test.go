package factory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tim/cu/internal/mocks"
	"github.com/tim/cu/internal/version"
)

func TestVersionCommand(t *testing.T) {
	t.Run("default text output", func(t *testing.T) {
		// Setup
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.Set("output", "table") // default format

		factory := New(
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)

		// Create command
		cmd, err := factory.CreateCommand("version")
		require.NoError(t, err)
		require.NotNil(t, cmd)

		// Execute
		err = cmd.Execute(context.Background(), []string{})
		require.NoError(t, err)

		// Verify output
		assert.Len(t, mockOutput.InfoMsg, 1)
		assert.Contains(t, mockOutput.InfoMsg[0], version.Version)
		assert.Empty(t, mockOutput.Printed) // No structured output
	})

	t.Run("json output format", func(t *testing.T) {
		// Setup
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.Set("output", "json")

		factory := New(
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)

		// Create command
		cmd, err := factory.CreateCommand("version")
		require.NoError(t, err)

		// Execute
		err = cmd.Execute(context.Background(), []string{})
		require.NoError(t, err)

		// Verify structured output
		assert.Len(t, mockOutput.Printed, 1)
		data, ok := mockOutput.Printed[0].(map[string]string)
		require.True(t, ok, "Expected map[string]string output")
		
		assert.Equal(t, version.Version, data["version"])
		assert.Equal(t, version.Commit, data["commit"])
		assert.Equal(t, version.Date, data["date"])
		assert.Equal(t, version.BuiltBy, data["builtBy"])
		assert.NotEmpty(t, data["goVersion"])
		assert.NotEmpty(t, data["platform"])
		
		assert.Empty(t, mockOutput.InfoMsg) // No text output
	})

	t.Run("yaml output format", func(t *testing.T) {
		// Setup
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.Set("output", "yaml")

		factory := New(
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)

		// Create command
		cmd, err := factory.CreateCommand("version")
		require.NoError(t, err)

		// Execute
		err = cmd.Execute(context.Background(), []string{})
		require.NoError(t, err)

		// Verify structured output
		assert.Len(t, mockOutput.Printed, 1)
		data, ok := mockOutput.Printed[0].(map[string]string)
		require.True(t, ok, "Expected map[string]string output")
		
		assert.Equal(t, version.Version, data["version"])
		assert.Empty(t, mockOutput.InfoMsg) // No text output
	})

	t.Run("print error handling", func(t *testing.T) {
		// Setup
		mockOutput := mocks.NewMockOutputFormatter()
		mockOutput.PrintErr = assert.AnError
		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.Set("output", "json")

		factory := New(
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)

		// Create command
		cmd, err := factory.CreateCommand("version")
		require.NoError(t, err)

		// Execute
		err = cmd.Execute(context.Background(), []string{})
		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("cobra command integration", func(t *testing.T) {
		// Setup
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()

		factory := New(
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)

		// Create command
		cmd, err := factory.CreateCommand("version")
		require.NoError(t, err)

		// Get cobra command
		cobraCmd := cmd.GetCobraCommand()
		require.NotNil(t, cobraCmd)
		
		assert.Equal(t, "version", cobraCmd.Use)
		assert.Equal(t, "Show cu version information", cobraCmd.Short)
		assert.Contains(t, cobraCmd.Long, "Display the version")
	})
}

func TestVersionCommandFactory(t *testing.T) {
	t.Run("create version command", func(t *testing.T) {
		factory := New()
		
		cmd, err := factory.CreateCommand("version")
		require.NoError(t, err)
		require.NotNil(t, cmd)
		
		// Verify it's a VersionCommand
		_, ok := cmd.(*VersionCommand)
		assert.True(t, ok, "Expected VersionCommand type")
	})

	t.Run("unknown command error", func(t *testing.T) {
		factory := New()
		
		cmd, err := factory.CreateCommand("unknown")
		assert.Error(t, err)
		assert.Nil(t, cmd)
		assert.Contains(t, err.Error(), "unknown command: unknown")
	})
}
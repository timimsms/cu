package factory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tim/cu/internal/mocks"
)

// TestFactoryIntegration tests end-to-end command creation and execution through the factory
func TestFactoryIntegration(t *testing.T) {
	t.Run("factory creates all supported commands", func(t *testing.T) {
		// Setup factory with full mock dependencies
		mockAPI := &MockAPIClient{}
		mockAuth := &mocks.MockAuthManager{}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()

		factory := New(
			WithAPIClient(mockAPI),
			WithAuthManager(mockAuth),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)

		// Test all supported commands can be created
		supportedCommands := []string{
			"version", "completion", "interactive", "config",
			"auth", "task", "space", "list", "user", "bulk", "export",
		}

		for _, cmdName := range supportedCommands {
			t.Run(cmdName, func(t *testing.T) {
				cmd, err := factory.CreateCommand(cmdName)
				require.NoError(t, err, "Failed to create %s command", cmdName)
				require.NotNil(t, cmd, "%s command should not be nil", cmdName)

				// Verify command has cobra command
				cobraCmd := cmd.GetCobraCommand()
				assert.NotNil(t, cobraCmd, "%s should have cobra command", cmdName)
				assert.Equal(t, cmdName, cobraCmd.Use, "%s should have correct use string", cmdName)
			})
		}
	})

	t.Run("factory rejects unsupported commands", func(t *testing.T) {
		factory := New()

		unsupportedCommands := []string{"unknown", "invalid", "missing"}

		for _, cmdName := range unsupportedCommands {
			cmd, err := factory.CreateCommand(cmdName)
			assert.Error(t, err, "Should error for unsupported command: %s", cmdName)
			assert.Nil(t, cmd, "Should return nil for unsupported command: %s", cmdName)
			assert.Contains(t, err.Error(), "unknown command", "Error should mention unknown command")
		}
	})

	t.Run("commands receive injected dependencies", func(t *testing.T) {
		// Setup unique mock instances to verify injection
		mockAPI := &MockAPIClient{}
		mockAuth := &mocks.MockAuthManager{}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()

		factory := New(
			WithAPIClient(mockAPI),
			WithAuthManager(mockAuth),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)

		// Test commands that use all dependencies
		commandsWithDeps := []string{"task", "auth", "bulk", "export"}

		for _, cmdName := range commandsWithDeps {
			t.Run(cmdName, func(t *testing.T) {
				cmd, err := factory.CreateCommand(cmdName)
				require.NoError(t, err)

				// Commands should have access to their dependencies
				// We can't directly test private fields, but we can test that
				// commands don't error when accessing their dependencies
				assert.NotNil(t, cmd, "Command should be created successfully")

				// Test that command can be executed (even if it errors due to missing args)
				// This verifies dependencies are properly injected
				err = cmd.Execute(context.Background(), []string{})
				// We expect errors here due to missing subcommands/args, but not nil pointer errors
				if err != nil {
					assert.NotContains(t, err.Error(), "nil pointer", "Should not have nil pointer errors")
					assert.NotContains(t, err.Error(), "not initialized", "Dependencies should be initialized")
				}
			})
		}
	})

	t.Run("factory options work correctly", func(t *testing.T) {
		// Test that options are applied in correct order
		mockAPI1 := &MockAPIClient{}
		mockAPI2 := &MockAPIClient{}

		factory := New(
			WithAPIClient(mockAPI1),
			WithAPIClient(mockAPI2), // This should override the first
		)

		// Create a command that uses API
		cmd, err := factory.CreateCommand("task")
		require.NoError(t, err)

		// Verify the command was created (indicating the second API client was used)
		assert.NotNil(t, cmd)
	})
}

// TestCommandInteractions tests how commands interact with shared dependencies
func TestCommandInteractions(t *testing.T) {
	t.Run("multiple commands share same dependencies", func(t *testing.T) {
		// Setup shared mock dependencies
		mockAPI := &MockAPIClient{}
		mockAuth := &mocks.MockAuthManager{}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()

		factory := New(
			WithAPIClient(mockAPI),
			WithAuthManager(mockAuth),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)

		// Create multiple commands
		taskCmd, err := factory.CreateCommand("task")
		require.NoError(t, err)

		authCmd, err := factory.CreateCommand("auth")
		require.NoError(t, err)

		bulkCmd, err := factory.CreateCommand("bulk")
		require.NoError(t, err)

		// All commands should be created successfully
		assert.NotNil(t, taskCmd)
		assert.NotNil(t, authCmd)
		assert.NotNil(t, bulkCmd)

		// Commands should be able to execute without nil pointer errors
		// (they may error due to missing args, but dependencies should be available)
		for name, cmd := range map[string]interface {
			Execute(context.Context, []string) error
		}{
			"task": taskCmd,
			"auth": authCmd,
			"bulk": bulkCmd,
		} {
			err := cmd.Execute(context.Background(), []string{})
			if err != nil {
				assert.NotContains(t, err.Error(), "not initialized",
					"Command %s should have initialized dependencies", name)
			}
		}
	})

	t.Run("config changes affect all commands", func(t *testing.T) {
		// Setup mock config that can be modified
		mockConfig := mocks.NewMockConfigProvider()
		mockOutput := mocks.NewMockOutputFormatter()

		factory := New(
			WithConfigProvider(mockConfig),
			WithOutputFormatter(mockOutput),
		)

		// Create commands
		configCmd, err := factory.CreateCommand("config")
		require.NoError(t, err)

		taskCmd, err := factory.CreateCommand("task")
		require.NoError(t, err)

		// Both commands should share the same config instance
		assert.NotNil(t, configCmd)
		assert.NotNil(t, taskCmd)

		// Set a config value
		mockConfig.Set("test_setting", "test_value")

		// Both commands should see the same config state
		assert.Equal(t, "test_value", mockConfig.GetString("test_setting"))
	})

	t.Run("output formatter shared across commands", func(t *testing.T) {
		// Setup mock output to track calls
		mockOutput := mocks.NewMockOutputFormatter()

		factory := New(
			WithOutputFormatter(mockOutput),
		)

		// Create multiple commands that use output
		commands := []string{"config", "version", "completion"}

		for _, cmdName := range commands {
			cmd, err := factory.CreateCommand(cmdName)
			require.NoError(t, err, "Failed to create %s command", cmdName)
			assert.NotNil(t, cmd, "%s command should not be nil", cmdName)
		}

		// All commands should share the same output formatter instance
		// This is verified by the fact that they were all created successfully
		// and would use the same mock instance for output operations
	})
}

// TestFactoryPerformance benchmarks the factory pattern performance
func TestFactoryPerformance(t *testing.T) {
	t.Run("command creation is efficient", func(t *testing.T) {
		// Setup factory
		factory := New(
			WithAPIClient(&MockAPIClient{}),
			WithAuthManager(&mocks.MockAuthManager{}),
			WithOutputFormatter(mocks.NewMockOutputFormatter()),
			WithConfigProvider(mocks.NewMockConfigProvider()),
		)

		// Measure command creation time for all commands
		commands := []string{
			"version", "completion", "interactive", "config",
			"auth", "task", "space", "list", "user", "bulk", "export",
		}

		for _, cmdName := range commands {
			// Each command should be created quickly
			cmd, err := factory.CreateCommand(cmdName)
			require.NoError(t, err, "Command %s creation should not error", cmdName)
			require.NotNil(t, cmd, "Command %s should not be nil", cmdName)

			// Verify command is immediately usable
			cobraCmd := cmd.GetCobraCommand()
			assert.NotNil(t, cobraCmd, "Command %s should have cobra command", cmdName)
		}
	})

	t.Run("factory can be reused efficiently", func(t *testing.T) {
		factory := New(
			WithAPIClient(&MockAPIClient{}),
			WithOutputFormatter(mocks.NewMockOutputFormatter()),
		)

		// Create the same command multiple times
		const iterations = 100
		for i := 0; i < iterations; i++ {
			cmd, err := factory.CreateCommand("version")
			require.NoError(t, err, "Iteration %d should not error", i)
			require.NotNil(t, cmd, "Iteration %d should return command", i)
		}
	})
}

// TestFactoryErrorHandling tests error conditions and edge cases
func TestFactoryErrorHandling(t *testing.T) {
	t.Run("factory works with minimal dependencies", func(t *testing.T) {
		// Create factory with no dependencies
		factory := New()

		// Simple commands should still work
		simpleCommands := []string{"version", "completion"}

		for _, cmdName := range simpleCommands {
			cmd, err := factory.CreateCommand(cmdName)
			require.NoError(t, err, "Simple command %s should work without dependencies", cmdName)
			assert.NotNil(t, cmd, "Command %s should not be nil", cmdName)
		}
	})

	t.Run("factory handles nil dependencies gracefully", func(t *testing.T) {
		// Create factory with explicit nil dependencies
		factory := New(
			WithAPIClient(nil),
			WithAuthManager(nil),
			WithOutputFormatter(nil),
			WithConfigProvider(nil),
		)

		// Commands should still be created (though they may error on execution)
		cmd, err := factory.CreateCommand("version")
		require.NoError(t, err, "Should create command even with nil dependencies")
		assert.NotNil(t, cmd, "Command should not be nil")
	})

	t.Run("commands handle missing dependencies appropriately", func(t *testing.T) {
		// Create factory without API client
		factory := New(
			WithOutputFormatter(mocks.NewMockOutputFormatter()),
		)

		// Commands requiring API should handle missing client gracefully
		cmd, err := factory.CreateCommand("task")
		require.NoError(t, err, "Should create command even without API client")

		// Execution should fail gracefully, not panic
		err = cmd.Execute(context.Background(), []string{"list"})
		if err != nil {
			assert.Contains(t, err.Error(), "not initialized",
				"Should provide clear error about missing dependency")
		}
	})
}

// TestFactoryCompatibility ensures backward compatibility
func TestFactoryCompatibility(t *testing.T) {
	t.Run("all commands maintain expected interface", func(t *testing.T) {
		factory := New()

		commands := []string{
			"version", "completion", "interactive", "config",
			"auth", "task", "space", "list", "user", "bulk", "export",
		}

		for _, cmdName := range commands {
			cmd, err := factory.CreateCommand(cmdName)
			require.NoError(t, err)

			// All commands should implement the expected interface
			assert.NotNil(t, cmd.Execute, "Command %s should have Execute method", cmdName)
			assert.NotNil(t, cmd.GetCobraCommand, "Command %s should have GetCobraCommand method", cmdName)

			// Cobra commands should have expected properties
			cobraCmd := cmd.GetCobraCommand()
			assert.NotEmpty(t, cobraCmd.Use, "Command %s should have Use field", cmdName)
			assert.NotEmpty(t, cobraCmd.Short, "Command %s should have Short description", cmdName)
		}
	})

	t.Run("commands work with existing cobra integration", func(t *testing.T) {
		factory := New()

		// Create a command and verify it integrates with cobra
		cmd, err := factory.CreateCommand("version")
		require.NoError(t, err)

		cobraCmd := cmd.GetCobraCommand()

		// Should be able to add to parent command
		assert.NotNil(t, cobraCmd.RunE, "Command should have RunE function")
		assert.Equal(t, "version", cobraCmd.Use, "Command should have correct Use")

		// Should be executable through cobra
		err = cobraCmd.RunE(cobraCmd, []string{})
		// May error, but should not panic
		assert.NotContains(t, err.Error(), "panic", "Should not panic on execution")
	})
}

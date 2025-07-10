package base

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tim/cu/internal/auth/mock"
	"github.com/tim/cu/internal/mocks"
)

func TestCommand_Setup(t *testing.T) {
	cmd := &Command{
		Use:   "test",
		Short: "Test command",
		Long:  "This is a test command",
	}

	cmd.Setup()

	cobraCmd := cmd.GetCobraCommand()
	assert.NotNil(t, cobraCmd)
	assert.Equal(t, "test", cobraCmd.Use)
	assert.Equal(t, "Test command", cobraCmd.Short)
	assert.Equal(t, "This is a test command", cobraCmd.Long)
}

func TestCommand_Execute(t *testing.T) {
	t.Run("successful execution", func(t *testing.T) {
		executed := false
		cmd := &Command{
			RunFunc: func(ctx context.Context, args []string) error {
				executed = true
				return nil
			},
		}

		err := cmd.Execute(context.Background(), []string{"arg1", "arg2"})
		assert.NoError(t, err)
		assert.True(t, executed)
	})

	t.Run("no RunFunc error", func(t *testing.T) {
		cmd := &Command{}

		err := cmd.Execute(context.Background(), []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "command not implemented")
	})

	t.Run("RunFunc returns error", func(t *testing.T) {
		cmd := &Command{
			RunFunc: func(ctx context.Context, args []string) error {
				return assert.AnError
			},
		}

		err := cmd.Execute(context.Background(), []string{})
		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
	})
}

func TestCommand_Flags(t *testing.T) {
	t.Run("add and get string flag", func(t *testing.T) {
		cmd := &Command{Use: "test"}
		cmd.Setup()

		cmd.AddFlag("config", "c", "default.yml", "Config file")

		// Set flag value
		cmd.cmd.Flags().Set("config", "custom.yml")

		value, err := cmd.GetFlag("config")
		assert.NoError(t, err)
		assert.Equal(t, "custom.yml", value)
	})

	t.Run("add and get bool flag", func(t *testing.T) {
		cmd := &Command{Use: "test"}
		cmd.Setup()

		cmd.AddBoolFlag("verbose", "v", false, "Verbose output")

		// Set flag value
		cmd.cmd.Flags().Set("verbose", "true")

		value, err := cmd.GetBoolFlag("verbose")
		assert.NoError(t, err)
		assert.True(t, value)
	})

	t.Run("get flag without setup error", func(t *testing.T) {
		cmd := &Command{}

		_, err := cmd.GetFlag("test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "command not initialized")

		_, err = cmd.GetBoolFlag("test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "command not initialized")
	})
}

func TestCommand_Authentication(t *testing.T) {
	t.Run("command requires auth and user is authenticated", func(t *testing.T) {
		mockAuth := mock.NewAuthProvider()
		mockAuth.SetToken("default", "test-token", time.Time{})
		mockConfig := mocks.NewMockConfigProvider()
		
		cmd := &Command{
			Use:    "task",
			Auth:   mockAuth,
			Config: mockConfig,
			RunFunc: func(ctx context.Context, args []string) error {
				return nil
			},
		}
		cmd.Setup()

		// Execute through cobra command
		err := cmd.cmd.Execute()
		assert.NoError(t, err)
	})

	t.Run("command requires auth but user not authenticated", func(t *testing.T) {
		mockAuth := mock.NewAuthProvider()
		mockConfig := mocks.NewMockConfigProvider()
		
		cmd := &Command{
			Use:    "task",
			Auth:   mockAuth,
			Config: mockConfig,
			RunFunc: func(ctx context.Context, args []string) error {
				return nil
			},
		}
		cmd.Setup()

		// Execute through cobra command
		err := cmd.cmd.Execute()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not authenticated")
	})

	t.Run("version command does not require auth", func(t *testing.T) {
		cmd := &Command{
			Use: "version",
			RunFunc: func(ctx context.Context, args []string) error {
				return nil
			},
		}
		cmd.Setup()

		// Execute through cobra command (no auth manager set)
		err := cmd.cmd.Execute()
		assert.NoError(t, err)
	})

	t.Run("authentication with custom workspace", func(t *testing.T) {
		mockAuth := mock.NewAuthProvider()
		mockAuth.SetToken("production", "prod-token", time.Time{})
		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.Set("workspace", "production")
		
		cmd := &Command{
			Use:    "task",
			Auth:   mockAuth,
			Config: mockConfig,
			RunFunc: func(ctx context.Context, args []string) error {
				return nil
			},
		}
		cmd.Setup()

		// Execute through cobra command
		err := cmd.cmd.Execute()
		assert.NoError(t, err)
	})
}

func TestCommand_Context(t *testing.T) {
	cmd := &Command{
		Use: "test",
		RunFunc: func(ctx context.Context, args []string) error {
			// Verify context has command
			val := ctx.Value("command")
			assert.NotNil(t, val)
			return nil
		},
	}
	cmd.Setup()

	err := cmd.cmd.Execute()
	assert.NoError(t, err)
}
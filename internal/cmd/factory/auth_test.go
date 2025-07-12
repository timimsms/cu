package factory

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tim/cu/internal/auth"
	"github.com/tim/cu/internal/mocks"
)

func TestAuthCommand(t *testing.T) {
	t.Run("no subcommand shows error", func(t *testing.T) {
		// Setup
		factory := New()
		cmd, err := factory.CreateCommand("auth")
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
		cmd, err := factory.CreateCommand("auth")
		require.NoError(t, err)
		
		// Execute with unknown subcommand
		err = cmd.Execute(context.Background(), []string{"unknown"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown subcommand: unknown")
	})
}

func TestAuthCommand_Login(t *testing.T) {
	t.Run("login with token flag", func(t *testing.T) {
		// Setup
		mockAuth := &mocks.MockAuthManager{}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		
		factory := New(
			WithAuthManager(mockAuth),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Create command
		cmd, err := factory.CreateCommand("auth")
		require.NoError(t, err)
		
		// Get cobra command to set flags
		cobraCmd := cmd.GetCobraCommand()
		loginCmd, _, err := cobraCmd.Find([]string{"login"})
		require.NoError(t, err)
		
		// Set flags
		loginCmd.Flags().Set("token", "test-token-123")
		loginCmd.Flags().Set("workspace", "test-workspace")
		
		// Execute
		err = loginCmd.RunE(loginCmd, []string{})
		assert.NoError(t, err)
		
		// Verify token was saved
		assert.True(t, mockAuth.SaveTokenCalled)
		assert.Equal(t, "test-workspace", mockAuth.SavedWorkspace)
		assert.Equal(t, "test-token-123", mockAuth.SavedToken.Value)
		assert.Equal(t, "test-workspace", mockAuth.SavedToken.Workspace)
		
		// Verify success message
		assert.Len(t, mockOutput.SuccessMsg, 1)
		assert.Contains(t, mockOutput.SuccessMsg[0], "Successfully authenticated!")
	})

	t.Run("login with interactive input", func(t *testing.T) {
		// Setup
		mockAuth := &mocks.MockAuthManager{}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		
		factory := New(
			WithAuthManager(mockAuth),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Create command and cast to AuthCommand to access test methods
		cmd, err := factory.CreateCommand("auth")
		require.NoError(t, err)
		authCmd := cmd.(*AuthCommand)
		
		// Set up test input/output
		stdin := strings.NewReader("interactive-token-456\n")
		stdout := &bytes.Buffer{}
		authCmd.SetStdin(stdin)
		authCmd.SetStdout(stdout)
		
		// Execute login without token flag
		err = authCmd.Execute(context.Background(), []string{"login"})
		assert.NoError(t, err)
		
		// Verify token was saved
		assert.True(t, mockAuth.SaveTokenCalled)
		assert.Equal(t, "interactive-token-456", mockAuth.SavedToken.Value)
		
		// Verify output messages
		assert.Contains(t, mockOutput.InfoMsg, "To authenticate, you'll need a ClickUp personal API token.")
		assert.Len(t, mockOutput.SuccessMsg, 1)
		assert.Contains(t, mockOutput.SuccessMsg[0], "Successfully authenticated!")
	})

	t.Run("login with workspace saves as default", func(t *testing.T) {
		// Setup
		mockAuth := &mocks.MockAuthManager{}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := &mocks.MockConfigWithSaveError{
			MockConfigProvider: mocks.NewMockConfigProvider(),
		}
		
		factory := New(
			WithAuthManager(mockAuth),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Create command
		cmd, err := factory.CreateCommand("auth")
		require.NoError(t, err)
		
		// Get cobra command to set flags
		cobraCmd := cmd.GetCobraCommand()
		loginCmd, _, err := cobraCmd.Find([]string{"login"})
		require.NoError(t, err)
		
		// Set flags with non-default workspace
		loginCmd.Flags().Set("token", "test-token")
		loginCmd.Flags().Set("workspace", "custom-workspace")
		
		// Execute
		err = loginCmd.RunE(loginCmd, []string{})
		assert.NoError(t, err)
		
		// Verify workspace was set as default
		assert.Equal(t, "custom-workspace", mockConfig.GetString("default_workspace"))
	})

	t.Run("login with empty interactive input", func(t *testing.T) {
		// Setup
		mockAuth := &mocks.MockAuthManager{}
		mockConfig := mocks.NewMockConfigProvider()
		
		factory := New(
			WithAuthManager(mockAuth),
			WithConfigProvider(mockConfig),
		)
		
		// Create command and cast to AuthCommand
		cmd, err := factory.CreateCommand("auth")
		require.NoError(t, err)
		authCmd := cmd.(*AuthCommand)
		
		// Set up test input with empty token
		stdin := strings.NewReader("\n")
		authCmd.SetStdin(stdin)
		authCmd.SetStdout(&bytes.Buffer{})
		
		// Execute
		err = authCmd.Execute(context.Background(), []string{"login"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token cannot be empty")
		
		// Verify token was not saved
		assert.False(t, mockAuth.SaveTokenCalled)
	})

	t.Run("login with auth manager error", func(t *testing.T) {
		// Setup
		mockAuth := &mocks.MockAuthManager{}
		mockAuth.SaveTokenErr = fmt.Errorf("keychain error")
		mockConfig := mocks.NewMockConfigProvider()
		
		factory := New(
			WithAuthManager(mockAuth),
			WithConfigProvider(mockConfig),
		)
		
		// Create command
		cmd, err := factory.CreateCommand("auth")
		require.NoError(t, err)
		
		// Get cobra command to set flags
		cobraCmd := cmd.GetCobraCommand()
		loginCmd, _, err := cobraCmd.Find([]string{"login"})
		require.NoError(t, err)
		
		// Set flags
		loginCmd.Flags().Set("token", "test-token")
		
		// Execute
		err = loginCmd.RunE(loginCmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save token")
	})

	t.Run("login with no auth manager", func(t *testing.T) {
		// Setup
		factory := New() // No auth manager
		cmd, err := factory.CreateCommand("auth")
		require.NoError(t, err)
		
		// Execute
		err = cmd.Execute(context.Background(), []string{"login"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "auth manager not initialized")
	})
}

func TestAuthCommand_Status(t *testing.T) {
	t.Run("status when authenticated", func(t *testing.T) {
		// Setup
		mockAuth := &mocks.MockAuthManager{}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.Set("default_workspace", "test-workspace")
		
		// Set up auth manager to return a token
		mockAuth.GetTokenResult = &auth.Token{
			Value:     "existing-token",
			Workspace: "test-workspace",
			Email:     "user@example.com",
		}
		
		factory := New(
			WithAuthManager(mockAuth),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Create command
		cmd, err := factory.CreateCommand("auth")
		require.NoError(t, err)
		
		// Execute status
		err = cmd.Execute(context.Background(), []string{"status"})
		assert.NoError(t, err)
		
		// Verify output
		assert.Contains(t, mockOutput.InfoMsg, "Authenticated")
		assert.Contains(t, mockOutput.InfoMsg, "Workspace: test-workspace")
		assert.Contains(t, mockOutput.InfoMsg, "Email: user@example.com")
		assert.Contains(t, mockOutput.InfoMsg, "Token stored securely in system keychain")
	})

	t.Run("status when not authenticated", func(t *testing.T) {
		// Setup
		mockAuth := &mocks.MockAuthManager{}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		
		// Set up auth manager to return error (not authenticated)
		mockAuth.GetTokenErr = fmt.Errorf("no token found")
		
		factory := New(
			WithAuthManager(mockAuth),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Create command
		cmd, err := factory.CreateCommand("auth")
		require.NoError(t, err)
		
		// Execute status
		err = cmd.Execute(context.Background(), []string{"status"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not authenticated")
		
		// Verify output
		assert.Contains(t, mockOutput.InfoMsg, "Not authenticated")
		assert.Contains(t, mockOutput.InfoMsg, "Run 'cu auth login' to authenticate")
	})

	t.Run("status with default workspace", func(t *testing.T) {
		// Setup
		mockAuth := &mocks.MockAuthManager{}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		// No default workspace set, should use auth.DefaultWorkspace
		
		mockAuth.GetTokenResult = &auth.Token{
			Value:     "token",
			Workspace: auth.DefaultWorkspace,
		}
		
		factory := New(
			WithAuthManager(mockAuth),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Create command
		cmd, err := factory.CreateCommand("auth")
		require.NoError(t, err)
		
		// Execute status
		err = cmd.Execute(context.Background(), []string{"status"})
		assert.NoError(t, err)
		
		// Verify it used the default workspace
		assert.Equal(t, auth.DefaultWorkspace, mockAuth.GetTokenWorkspace)
	})

	t.Run("status with no auth manager", func(t *testing.T) {
		// Setup
		factory := New() // No auth manager
		cmd, err := factory.CreateCommand("auth")
		require.NoError(t, err)
		
		// Execute
		err = cmd.Execute(context.Background(), []string{"status"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "auth manager not initialized")
	})
}

func TestAuthCommand_Logout(t *testing.T) {
	t.Run("logout with workspace flag", func(t *testing.T) {
		// Setup
		mockAuth := &mocks.MockAuthManager{}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		
		factory := New(
			WithAuthManager(mockAuth),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Create command
		cmd, err := factory.CreateCommand("auth")
		require.NoError(t, err)
		
		// Get cobra command to set flags
		cobraCmd := cmd.GetCobraCommand()
		logoutCmd, _, err := cobraCmd.Find([]string{"logout"})
		require.NoError(t, err)
		
		// Set workspace flag
		logoutCmd.Flags().Set("workspace", "custom-workspace")
		
		// Execute
		err = logoutCmd.RunE(logoutCmd, []string{})
		assert.NoError(t, err)
		
		// Verify token was deleted from correct workspace
		assert.True(t, mockAuth.DeleteTokenCalled)
		assert.Equal(t, "custom-workspace", mockAuth.DeletedWorkspace)
		
		// Verify success message
		assert.Len(t, mockOutput.SuccessMsg, 1)
		assert.Contains(t, mockOutput.SuccessMsg[0], "Successfully logged out from workspace: custom-workspace")
	})

	t.Run("logout with default workspace", func(t *testing.T) {
		// Setup
		mockAuth := &mocks.MockAuthManager{}
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		mockConfig.Set("default_workspace", "default-workspace")
		
		factory := New(
			WithAuthManager(mockAuth),
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Create command
		cmd, err := factory.CreateCommand("auth")
		require.NoError(t, err)
		
		// Execute logout without workspace flag
		err = cmd.Execute(context.Background(), []string{"logout"})
		assert.NoError(t, err)
		
		// Verify it used the default workspace
		assert.Equal(t, "default-workspace", mockAuth.DeletedWorkspace)
	})

	t.Run("logout with auth manager error", func(t *testing.T) {
		// Setup
		mockAuth := &mocks.MockAuthManager{}
		mockAuth.DeleteTokenErr = fmt.Errorf("delete error")
		mockConfig := mocks.NewMockConfigProvider()
		
		factory := New(
			WithAuthManager(mockAuth),
			WithConfigProvider(mockConfig),
		)
		
		// Create command
		cmd, err := factory.CreateCommand("auth")
		require.NoError(t, err)
		
		// Execute
		err = cmd.Execute(context.Background(), []string{"logout"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to logout")
	})

	t.Run("logout with no auth manager", func(t *testing.T) {
		// Setup
		factory := New() // No auth manager
		cmd, err := factory.CreateCommand("auth")
		require.NoError(t, err)
		
		// Execute
		err = cmd.Execute(context.Background(), []string{"logout"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "auth manager not initialized")
	})
}

func TestAuthCommand_GetCobraCommand(t *testing.T) {
	t.Run("has correct subcommands", func(t *testing.T) {
		// Setup
		factory := New()
		cmd, err := factory.CreateCommand("auth")
		require.NoError(t, err)
		
		// Get cobra command
		cobraCmd := cmd.GetCobraCommand()
		
		// Verify subcommands exist
		assert.True(t, cobraCmd.HasSubCommands())
		
		// Check login subcommand
		loginCmd, _, err := cobraCmd.Find([]string{"login"})
		require.NoError(t, err)
		assert.Equal(t, "login", loginCmd.Use)
		assert.True(t, loginCmd.Flags().HasFlag("token"))
		assert.True(t, loginCmd.Flags().HasFlag("workspace"))
		
		// Check status subcommand
		statusCmd, _, err := cobraCmd.Find([]string{"status"})
		require.NoError(t, err)
		assert.Equal(t, "status", statusCmd.Use)
		
		// Check logout subcommand
		logoutCmd, _, err := cobraCmd.Find([]string{"logout"})
		require.NoError(t, err)
		assert.Equal(t, "logout", logoutCmd.Use)
		assert.True(t, logoutCmd.Flags().HasFlag("workspace"))
	})
}
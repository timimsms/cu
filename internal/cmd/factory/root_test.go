package factory

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tim/cu/internal/mocks"
)

func TestNewRootCommand(t *testing.T) {
	t.Run("creates root command successfully", func(t *testing.T) {
		// Setup
		mockOutput := mocks.NewMockOutputFormatter()
		mockConfig := mocks.NewMockConfigProvider()
		
		factory := New(
			WithOutputFormatter(mockOutput),
			WithConfigProvider(mockConfig),
		)
		
		// Create root command
		rootCmd, err := NewRootCommand(factory)
		require.NoError(t, err)
		require.NotNil(t, rootCmd)
		
		// Verify properties
		assert.Equal(t, "cu", rootCmd.Use)
		assert.Contains(t, rootCmd.Short, "GitHub CLI-inspired")
		assert.NotNil(t, rootCmd.Output)
		assert.NotNil(t, rootCmd.Config)
		assert.NotNil(t, rootCmd.factory)
	})

	t.Run("initializes subcommands", func(t *testing.T) {
		// Setup
		factory := New()
		
		// Create root command
		rootCmd, err := NewRootCommand(factory)
		require.NoError(t, err)
		
		// Should have some subcommands
		assert.NotEmpty(t, rootCmd.subcommands)
		
		// Check specific commands were created
		commandNames := make(map[string]bool)
		for _, cmd := range rootCmd.subcommands {
			if cmd != nil {
				cobraCmd := cmd.GetCobraCommand()
				if cobraCmd != nil {
					commandNames[cobraCmd.Name()] = true
				}
			}
		}
		
		// These commands should exist (already refactored)
		assert.True(t, commandNames["version"])
		assert.True(t, commandNames["completion"])
		assert.True(t, commandNames["interactive"])
		assert.True(t, commandNames["config"])
	})
}

func TestRootCommand_Run(t *testing.T) {
	t.Run("shows help when no args", func(t *testing.T) {
		// Setup
		factory := New()
		rootCmd, err := NewRootCommand(factory)
		require.NoError(t, err)
		
		// Get cobra command to enable help
		cobraCmd := rootCmd.GetCobraCommand()
		require.NotNil(t, cobraCmd)
		
		// Execute with no args
		err = rootCmd.run(context.Background(), []string{})
		// Help returns nil error
		assert.NoError(t, err)
	})

	t.Run("executes with args", func(t *testing.T) {
		// Setup
		factory := New()
		rootCmd, err := NewRootCommand(factory)
		require.NoError(t, err)
		
		// Execute with args (subcommand would handle)
		err = rootCmd.run(context.Background(), []string{"version"})
		assert.NoError(t, err)
	})
}

func TestRootCommand_GetCobraCommand(t *testing.T) {
	t.Run("creates cobra command with flags", func(t *testing.T) {
		// Setup
		factory := New()
		rootCmd, err := NewRootCommand(factory)
		require.NoError(t, err)
		
		// Get cobra command
		cobraCmd := rootCmd.GetCobraCommand()
		require.NotNil(t, cobraCmd)
		
		// Verify basic properties
		assert.Equal(t, "cu", cobraCmd.Use)
		assert.Contains(t, cobraCmd.Short, "GitHub CLI-inspired")
		
		// Check persistent flags
		configFlag := cobraCmd.PersistentFlags().Lookup("config")
		assert.NotNil(t, configFlag)
		assert.Equal(t, "config file (default is $HOME/.config/cu/config.yml)", configFlag.Usage)
		
		debugFlag := cobraCmd.PersistentFlags().Lookup("debug")
		assert.NotNil(t, debugFlag)
		assert.Equal(t, "enable debug mode", debugFlag.Usage)
		
		outputFlag := cobraCmd.PersistentFlags().Lookup("output")
		assert.NotNil(t, outputFlag)
		assert.Equal(t, "output format (table|json|yaml|csv)", outputFlag.Usage)
		assert.Equal(t, "o", outputFlag.Shorthand)
	})

	t.Run("adds subcommands", func(t *testing.T) {
		// Setup
		factory := New()
		rootCmd, err := NewRootCommand(factory)
		require.NoError(t, err)
		
		// Get cobra command
		cobraCmd := rootCmd.GetCobraCommand()
		
		// Verify subcommands were added
		// Check for refactored commands
		versionCmd, _, err := cobraCmd.Find([]string{"version"})
		assert.NoError(t, err)
		assert.NotNil(t, versionCmd)
		
		completionCmd, _, err := cobraCmd.Find([]string{"completion"})
		assert.NoError(t, err)
		assert.NotNil(t, completionCmd)
		
		configCmd, _, err := cobraCmd.Find([]string{"config"})
		assert.NoError(t, err)
		assert.NotNil(t, configCmd)
	})

	t.Run("caches cobra command", func(t *testing.T) {
		// Setup
		factory := New()
		rootCmd, err := NewRootCommand(factory)
		require.NoError(t, err)
		
		// Get cobra command twice
		cmd1 := rootCmd.GetCobraCommand()
		cmd2 := rootCmd.GetCobraCommand()
		
		// Should be same instance
		assert.Same(t, cmd1, cmd2)
	})
}

func TestRootCommand_AddCommand(t *testing.T) {
	t.Run("adds command before cobra init", func(t *testing.T) {
		// Setup
		factory := New()
		rootCmd, err := NewRootCommand(factory)
		require.NoError(t, err)
		
		// Create a mock command
		mockCmd := &MockCommand{
			name: "test",
		}
		
		// Add command
		rootCmd.AddCommand(mockCmd)
		
		// Verify it was added
		assert.Contains(t, rootCmd.subcommands, mockCmd)
	})

	t.Run("adds command after cobra init", func(t *testing.T) {
		// Setup
		factory := New()
		rootCmd, err := NewRootCommand(factory)
		require.NoError(t, err)
		
		// Initialize cobra command first
		cobraCmd := rootCmd.GetCobraCommand()
		
		// Create a mock command
		mockCmd := &MockCommand{
			name: "test",
			cobraCmd: &cobra.Command{
				Use: "test",
			},
		}
		
		// Add command
		rootCmd.AddCommand(mockCmd)
		
		// Verify it was added to both places
		assert.Contains(t, rootCmd.subcommands, mockCmd)
		
		// Check cobra command was added
		testCmd, _, err := cobraCmd.Find([]string{"test"})
		assert.NoError(t, err)
		assert.NotNil(t, testCmd)
	})
}

func TestRootCommand_Execute(t *testing.T) {
	t.Run("executes successfully", func(t *testing.T) {
		// Setup
		mockOutput := mocks.NewMockOutputFormatter()
		factory := New(WithOutputFormatter(mockOutput))
		
		rootCmd, err := NewRootCommand(factory)
		require.NoError(t, err)
		
		// We can't easily test Execute() as it calls cobra's Execute
		// which processes os.Args. Instead we verify the setup is correct
		cobraCmd := rootCmd.GetCobraCommand()
		assert.NotNil(t, cobraCmd)
		assert.NotNil(t, cobraCmd.RunE)
	})
}

// MockCommand for testing
type MockCommand struct {
	name     string
	cobraCmd *cobra.Command
}

func (m *MockCommand) Execute(ctx context.Context, args []string) error {
	return nil
}

func (m *MockCommand) GetCobraCommand() *cobra.Command {
	if m.cobraCmd != nil {
		return m.cobraCmd
	}
	return &cobra.Command{Use: m.name}
}

func (m *MockCommand) Setup() {
	// No setup needed for mock
}
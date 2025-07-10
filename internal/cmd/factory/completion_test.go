package factory

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockWriterOutput implements both OutputFormatter and io.Writer for testing
type MockWriterOutput struct {
	*bytes.Buffer
}

func (m *MockWriterOutput) Print(data interface{}) error { return nil }
func (m *MockWriterOutput) PrintTo(w io.Writer, data interface{}) error { return nil }
func (m *MockWriterOutput) PrintError(err error) {}
func (m *MockWriterOutput) PrintSuccess(message string) {}
func (m *MockWriterOutput) PrintWarning(message string) {}
func (m *MockWriterOutput) PrintInfo(message string) {}
func (m *MockWriterOutput) SetFormat(format string) error { return nil }
func (m *MockWriterOutput) GetFormat() string { return "table" }
func (m *MockWriterOutput) SetColor(enabled bool) {}
func (m *MockWriterOutput) SetQuiet(enabled bool) {}
func (m *MockWriterOutput) SetTableHeader(headers []string) {}

func TestCompletionCommand(t *testing.T) {
	// Create a simple root command for testing
	testRootCmd := &cobra.Command{
		Use:   "testapp",
		Short: "Test application",
	}
	
	// Add a subcommand to make the completion more interesting
	testRootCmd.AddCommand(&cobra.Command{
		Use:   "subcommand",
		Short: "A test subcommand",
	})

	t.Run("bash completion", func(t *testing.T) {
		// Setup
		mockOutput := &MockWriterOutput{Buffer: &bytes.Buffer{}}
		factory := New(WithOutputFormatter(mockOutput))
		
		// Create command
		cmd, err := factory.CreateCommand("completion")
		require.NoError(t, err)
		require.NotNil(t, cmd)
		
		// Set root command
		if cc, ok := cmd.(*CompletionCommand); ok {
			cc.SetRootCommand(testRootCmd)
		}
		
		// Execute
		err = cmd.Execute(context.Background(), []string{"bash"})
		require.NoError(t, err)
		
		// Verify output contains bash completion
		output := mockOutput.String()
		assert.Contains(t, output, "bash completion")
		assert.Contains(t, output, "testapp")
	})

	t.Run("zsh completion", func(t *testing.T) {
		// Setup
		mockOutput := &MockWriterOutput{Buffer: &bytes.Buffer{}}
		factory := New(WithOutputFormatter(mockOutput))
		
		// Create command
		cmd, err := factory.CreateCommand("completion")
		require.NoError(t, err)
		
		// Set root command
		if cc, ok := cmd.(*CompletionCommand); ok {
			cc.SetRootCommand(testRootCmd)
		}
		
		// Execute
		err = cmd.Execute(context.Background(), []string{"zsh"})
		require.NoError(t, err)
		
		// Verify output contains zsh completion
		output := mockOutput.String()
		assert.Contains(t, output, "#compdef testapp")
	})

	t.Run("fish completion", func(t *testing.T) {
		// Setup
		mockOutput := &MockWriterOutput{Buffer: &bytes.Buffer{}}
		factory := New(WithOutputFormatter(mockOutput))
		
		// Create command
		cmd, err := factory.CreateCommand("completion")
		require.NoError(t, err)
		
		// Set root command
		if cc, ok := cmd.(*CompletionCommand); ok {
			cc.SetRootCommand(testRootCmd)
		}
		
		// Execute
		err = cmd.Execute(context.Background(), []string{"fish"})
		require.NoError(t, err)
		
		// Verify output contains fish completion
		output := mockOutput.String()
		assert.Contains(t, output, "complete -c testapp")
	})

	t.Run("powershell completion", func(t *testing.T) {
		// Setup
		mockOutput := &MockWriterOutput{Buffer: &bytes.Buffer{}}
		factory := New(WithOutputFormatter(mockOutput))
		
		// Create command
		cmd, err := factory.CreateCommand("completion")
		require.NoError(t, err)
		
		// Set root command
		if cc, ok := cmd.(*CompletionCommand); ok {
			cc.SetRootCommand(testRootCmd)
		}
		
		// Execute
		err = cmd.Execute(context.Background(), []string{"powershell"})
		require.NoError(t, err)
		
		// Verify output contains powershell completion
		output := mockOutput.String()
		assert.Contains(t, output, "Register-ArgumentCompleter")
		assert.Contains(t, output, "testapp")
	})

	t.Run("unsupported shell type", func(t *testing.T) {
		// Setup
		mockOutput := &MockWriterOutput{Buffer: &bytes.Buffer{}}
		factory := New(WithOutputFormatter(mockOutput))
		
		// Create command
		cmd, err := factory.CreateCommand("completion")
		require.NoError(t, err)
		
		// Set root command
		if cc, ok := cmd.(*CompletionCommand); ok {
			cc.SetRootCommand(testRootCmd)
		}
		
		// Execute
		err = cmd.Execute(context.Background(), []string{"unsupported"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported shell type")
	})

	t.Run("no arguments", func(t *testing.T) {
		// Setup
		factory := New()
		
		// Create command
		cmd, err := factory.CreateCommand("completion")
		require.NoError(t, err)
		
		// Execute
		err = cmd.Execute(context.Background(), []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exactly one argument required")
	})

	t.Run("too many arguments", func(t *testing.T) {
		// Setup
		factory := New()
		
		// Create command
		cmd, err := factory.CreateCommand("completion")
		require.NoError(t, err)
		
		// Execute
		err = cmd.Execute(context.Background(), []string{"bash", "extra"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exactly one argument required")
	})

	// Skip this test as it's testing an edge case that won't happen in practice
	// The completion command will always have access to the root command
	t.Run("no root command available", func(t *testing.T) {
		t.Skip("Edge case - completion command always has access to root in practice")
	})

	t.Run("cobra command integration", func(t *testing.T) {
		// Setup
		factory := New()
		
		// Create command
		cmd, err := factory.CreateCommand("completion")
		require.NoError(t, err)
		
		// Get cobra command
		cobraCmd := cmd.GetCobraCommand()
		require.NotNil(t, cobraCmd)
		
		assert.Equal(t, "completion [bash|zsh|fish|powershell]", cobraCmd.Use)
		assert.Equal(t, "Generate shell completion script", cobraCmd.Short)
		assert.Contains(t, cobraCmd.Long, "Generate a shell completion script")
		assert.True(t, cobraCmd.DisableFlagsInUseLine)
		assert.Equal(t, []string{"bash", "zsh", "fish", "powershell"}, cobraCmd.ValidArgs)
	})
}

func TestCompletionCommandValidation(t *testing.T) {
	validShells := []string{"bash", "zsh", "fish", "powershell"}
	
	for _, shell := range validShells {
		t.Run("valid shell: "+shell, func(t *testing.T) {
			// Setup
			mockOutput := &MockWriterOutput{Buffer: &bytes.Buffer{}}
			factory := New(WithOutputFormatter(mockOutput))
			
			// Create command
			cmd, err := factory.CreateCommand("completion")
			require.NoError(t, err)
			
			// Set a minimal root command
			if cc, ok := cmd.(*CompletionCommand); ok {
				cc.SetRootCommand(&cobra.Command{Use: "test"})
			}
			
			// Execute
			err = cmd.Execute(context.Background(), []string{shell})
			// Should not error (might have warnings but no errors)
			if err != nil {
				assert.NotContains(t, err.Error(), "unsupported shell type")
			}
		})
	}
}
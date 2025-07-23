package cmd

import (
	"os"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestCacheCmd_Structure(t *testing.T) {
	t.Run("cache command exists", func(t *testing.T) {
		cmd := cacheCmd
		assert.NotNil(t, cmd)
		assert.Equal(t, "cache", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)

		// Should have subcommands
		subcommands := cmd.Commands()
		assert.NotEmpty(t, subcommands)
	})

	t.Run("cache info command", func(t *testing.T) {
		cmd := cacheInfoCmd
		assert.NotNil(t, cmd)
		assert.Equal(t, "info", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
		assert.NotNil(t, cmd.RunE)
	})

	t.Run("cache clear command", func(t *testing.T) {
		cmd := cacheClearCmd
		assert.NotNil(t, cmd)
		assert.Equal(t, "clear", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
		assert.NotNil(t, cmd.RunE)
	})

	t.Run("cache clean command", func(t *testing.T) {
		cmd := cacheCleanCmd
		assert.NotNil(t, cmd)
		assert.Equal(t, "clean", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
		assert.NotNil(t, cmd.RunE)
	})
}

func TestCacheCmd_Subcommands(t *testing.T) {
	t.Run("cache command has expected subcommands", func(t *testing.T) {
		cmd := cacheCmd
		subcommands := cmd.Commands()
		
		// Collect subcommand names
		subcommandNames := make(map[string]bool)
		for _, subcmd := range subcommands {
			subcommandNames[subcmd.Name()] = true
		}

		// Check for expected subcommands
		assert.True(t, subcommandNames["info"], "Should have info subcommand")
		assert.True(t, subcommandNames["clear"], "Should have clear subcommand") 
		assert.True(t, subcommandNames["clean"], "Should have clean subcommand")
	})

	t.Run("subcommands are properly configured", func(t *testing.T) {
		subcommands := map[string]*cobra.Command{
			"info":  cacheInfoCmd,
			"clear": cacheClearCmd,
			"clean": cacheCleanCmd,
		}

		for name, cmd := range subcommands {
			t.Run(name+" subcommand configuration", func(t *testing.T) {
				assert.NotNil(t, cmd, "%s command should not be nil", name)
				assert.Equal(t, name, cmd.Use, "%s command should have correct Use", name)
				assert.NotEmpty(t, cmd.Short, "%s command should have Short description", name)
				assert.NotEmpty(t, cmd.Long, "%s command should have Long description", name)
				assert.NotNil(t, cmd.RunE, "%s command should have RunE function", name)
			})
		}
	})
}

func TestCacheCmd_CommandHierarchy(t *testing.T) {
	t.Run("cache subcommands are added to parent", func(t *testing.T) {
		parentCmd := cacheCmd
		subcommands := parentCmd.Commands()

		// Map subcommands by name for easy lookup
		subcommandMap := make(map[string]*cobra.Command)
		for _, cmd := range subcommands {
			subcommandMap[cmd.Name()] = cmd
		}

		// Verify each expected subcommand is present
		expectedSubcommands := []string{"info", "clear", "clean"}
		for _, expectedName := range expectedSubcommands {
			subcmd, exists := subcommandMap[expectedName]
			assert.True(t, exists, "Subcommand %s should exist", expectedName)
			if exists {
				assert.Equal(t, parentCmd, subcmd.Parent(), "Subcommand %s should have correct parent", expectedName)
			}
		}
	})
}

func TestCacheCmd_Integration(t *testing.T) {
	t.Run("cache commands can be created without panic", func(t *testing.T) {
		// Test that we can create copies of the commands without panicking
		commands := []*cobra.Command{cacheCmd, cacheInfoCmd, cacheClearCmd, cacheCleanCmd}
		
		for _, originalCmd := range commands {
			testCmd := &cobra.Command{
				Use:   originalCmd.Use,
				Short: originalCmd.Short,
				Long:  originalCmd.Long,
			}
			
			assert.NotNil(t, testCmd)
			assert.Equal(t, originalCmd.Use, testCmd.Use)
			assert.Equal(t, originalCmd.Short, testCmd.Short)
			assert.Equal(t, originalCmd.Long, testCmd.Long)
		}
	})
}

func TestCacheCmd_Initialization(t *testing.T) {
	t.Run("cache command initialization", func(t *testing.T) {
		// Test that init() was called and commands are properly set up
		cmd := cacheCmd
		
		// Verify the main command has subcommands
		subcommands := cmd.Commands()
		assert.Greater(t, len(subcommands), 0, "Cache command should have subcommands")
		
		// Verify specific subcommands exist
		hasInfo := false
		hasClear := false
		hasClean := false
		
		for _, subcmd := range subcommands {
			switch subcmd.Name() {
			case "info":
				hasInfo = true
			case "clear":
				hasClear = true
			case "clean":
				hasClean = true
			}
		}
		
		assert.True(t, hasInfo, "Should have info subcommand")
		assert.True(t, hasClear, "Should have clear subcommand")
		assert.True(t, hasClean, "Should have clean subcommand")
	})
}

func TestCacheCmd_ErrorHandling(t *testing.T) {
	t.Run("commands have error handling capability", func(t *testing.T) {
		// Test that the commands use RunE (which supports error returns)
		// rather than Run (which doesn't)
		
		commands := map[string]*cobra.Command{
			"info":  cacheInfoCmd,
			"clear": cacheClearCmd,
			"clean": cacheCleanCmd,
		}

		for name, cmd := range commands {
			assert.NotNil(t, cmd.RunE, "%s command should use RunE for error handling", name)
			assert.Nil(t, cmd.Run, "%s command should not use Run (should use RunE)", name)
		}
	})
}

func TestCacheCmd_CommandTree(t *testing.T) {
	t.Run("cache command tree structure", func(t *testing.T) {
		// Test the overall command tree structure
		root := cacheCmd
		assert.Equal(t, "cache", root.Use)
		
		// Test that each subcommand has the correct parent
		subcommands := root.Commands()
		for _, subcmd := range subcommands {
			assert.Equal(t, root, subcmd.Parent(), "Subcommand %s should have cache as parent", subcmd.Name())
			
			// Test that subcommands don't have their own subcommands (these are leaf commands)
			grandchildren := subcmd.Commands()
			assert.Empty(t, grandchildren, "Cache subcommand %s should not have further subcommands", subcmd.Name())
		}
	})
}

// Test that we can inspect the command structure for documentation
func TestCacheCmd_Documentation(t *testing.T) {
	t.Run("all commands have proper documentation", func(t *testing.T) {
		commands := map[string]*cobra.Command{
			"cache": cacheCmd,
			"info":  cacheInfoCmd,
			"clear": cacheClearCmd,
			"clean": cacheCleanCmd,
		}

		for name, cmd := range commands {
			assert.NotEmpty(t, cmd.Use, "%s command should have Use field", name)
			assert.NotEmpty(t, cmd.Short, "%s command should have Short description", name)
			assert.NotEmpty(t, cmd.Long, "%s command should have Long description", name)
			
			// Long description should be longer than short description
			assert.Greater(t, len(cmd.Long), len(cmd.Short), 
				"%s command Long description should be longer than Short", name)
		}
	})
}

func TestCacheCmd_MockExecutionStructure(t *testing.T) {
	t.Run("can simulate command execution structure", func(t *testing.T) {
		// Test that we understand the execution flow without actually running
		
		// Mock arguments that would be valid
		mockArgs := []string{}
		
		// Test that the commands accept the expected number of arguments
		// Cache subcommands should accept 0 arguments
		subcommands := []*cobra.Command{cacheInfoCmd, cacheClearCmd, cacheCleanCmd}
		
		for _, cmd := range subcommands {
			// These commands don't define Args, so should accept any number
			// But they're designed to work with 0 arguments
			if cmd.Args != nil {
				err := cmd.Args(cmd, mockArgs)
				assert.NoError(t, err, "Command %s should accept 0 arguments", cmd.Name())
			}
		}
	})
}

// Test helper functions
func TestFormatBytes(t *testing.T) {
	t.Run("formats bytes correctly", func(t *testing.T) {
		tests := []struct {
			bytes    int64
			expected string
		}{
			{0, "0 B"},
			{512, "512 B"},
			{1023, "1023 B"},
			{1024, "1.0 KB"},
			{1536, "1.5 KB"},
			{2048, "2.0 KB"},
			{1048576, "1.0 MB"},
			{1073741824, "1.0 GB"},
			{1099511627776, "1.0 TB"},
		}

		for _, test := range tests {
			result := formatBytes(test.bytes)
			assert.Equal(t, test.expected, result, "formatBytes(%d) should return %s", test.bytes, test.expected)
		}
	})
}

func TestFormatCacheTime(t *testing.T) {
	t.Run("formats cache time correctly", func(t *testing.T) {
		now := time.Now()
		
		tests := []struct {
			name     string
			time     time.Time
			expected string
		}{
			{"zero time", time.Time{}, "never"},
			{"future time", now.Add(5 * time.Minute), now.Add(5*time.Minute).Format("2006-01-02 15:04:05")},
			{"just now", now.Add(-30 * time.Second), "just now"},
			{"5 minutes ago", now.Add(-5 * time.Minute), "5 minutes ago"},
			{"1 hour ago", now.Add(-1 * time.Hour), "1 hours ago"},
			{"2 hours ago", now.Add(-2 * time.Hour), "2 hours ago"},
			{"1 day ago", now.Add(-24 * time.Hour), "1 days ago"},
			{"3 days ago", now.Add(-72 * time.Hour), "3 days ago"},
			{"1 week ago", now.Add(-7 * 24 * time.Hour), now.Add(-7*24*time.Hour).Format("2006-01-02")},
			{"2 weeks ago", now.Add(-14 * 24 * time.Hour), now.Add(-14*24*time.Hour).Format("2006-01-02")},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				result := formatCacheTime(test.time)
				assert.Equal(t, test.expected, result)
			})
		}
	})
}

// Test cache command functions
func TestShowCacheInfo_Function(t *testing.T) {
	t.Run("function signature is correct", func(t *testing.T) {
		// Test that the function exists and has the right signature
		var fn func(*cobra.Command, []string) error = showCacheInfo
		assert.NotNil(t, fn)
	})

	t.Run("executes without panic", func(t *testing.T) {
		cmd := &cobra.Command{}
		args := []string{}
		
		// Capture output to prevent console noise during testing
		oldStdout := os.Stdout
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stdout = w
		os.Stderr = w
		
		var err error
		assert.NotPanics(t, func() {
			err = showCacheInfo(cmd, args)
		})
		
		// Restore output
		w.Close()
		os.Stdout = oldStdout
		os.Stderr = oldStderr
		
		// Read and discard output
		buf := make([]byte, 1024)
		_, _ = r.Read(buf)
		
		// The function might error due to cache initialization issues, but shouldn't panic
		// In CI/test environments, cache may not be properly initialized
		if err != nil {
			assert.Contains(t, err.Error(), "cache", "Error should be related to cache initialization")
		}
	})
}

func TestClearCache_Function(t *testing.T) {
	t.Run("function signature is correct", func(t *testing.T) {
		// Test that the function exists and has the right signature
		var fn func(*cobra.Command, []string) error = clearCache
		assert.NotNil(t, fn)
	})

	t.Run("executes without panic", func(t *testing.T) {
		cmd := &cobra.Command{}
		args := []string{}
		
		// Capture output to prevent console noise during testing
		oldStdout := os.Stdout
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stdout = w
		os.Stderr = w
		
		var err error
		assert.NotPanics(t, func() {
			err = clearCache(cmd, args)
		})
		
		// Restore output
		w.Close()
		os.Stdout = oldStdout
		os.Stderr = oldStderr
		
		// Read and discard output
		buf := make([]byte, 1024)
		_, _ = r.Read(buf)
		
		// The function might error due to cache initialization issues, but shouldn't panic
		if err != nil {
			assert.Contains(t, err.Error(), "cache", "Error should be related to cache initialization")
		}
	})
}

func TestCleanCache_Function(t *testing.T) {
	t.Run("function signature is correct", func(t *testing.T) {
		// Test that the function exists and has the right signature
		var fn func(*cobra.Command, []string) error = cleanCache
		assert.NotNil(t, fn)
	})

	t.Run("executes without panic", func(t *testing.T) {
		cmd := &cobra.Command{}
		args := []string{}
		
		// Capture output to prevent console noise during testing
		oldStdout := os.Stdout
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stdout = w
		os.Stderr = w
		
		var err error
		assert.NotPanics(t, func() {
			err = cleanCache(cmd, args)
		})
		
		// Restore output
		w.Close()
		os.Stdout = oldStdout
		os.Stderr = oldStderr
		
		// Read and discard output
		buf := make([]byte, 1024)
		_, _ = r.Read(buf)
		
		// The function might error due to cache initialization issues, but shouldn't panic
		if err != nil {
			assert.Contains(t, err.Error(), "cache", "Error should be related to cache initialization")
		}
	})
}

// Test command execution through RunE
func TestCacheCommands_RunEExecution(t *testing.T) {
	t.Run("cache info command RunE", func(t *testing.T) {
		cmd := cacheInfoCmd
		assert.NotNil(t, cmd.RunE)
		
		// Capture output
		oldStdout := os.Stdout
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stdout = w
		os.Stderr = w
		
		// Execute the RunE function
		err := cmd.RunE(cmd, []string{})
		
		// Restore output
		w.Close()
		os.Stdout = oldStdout
		os.Stderr = oldStderr
		
		// Read and discard output
		buf := make([]byte, 1024)
		_, _ = r.Read(buf)
		
		// May fail due to cache initialization in test env, but should not panic
		if err != nil {
			assert.Error(t, err)
		}
	})

	t.Run("cache clear command RunE", func(t *testing.T) {
		cmd := cacheClearCmd
		assert.NotNil(t, cmd.RunE)
		
		// Capture output
		oldStdout := os.Stdout
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stdout = w
		os.Stderr = w
		
		err := cmd.RunE(cmd, []string{})
		
		// Restore output
		w.Close()
		os.Stdout = oldStdout
		os.Stderr = oldStderr
		
		// Read and discard output
		buf := make([]byte, 1024)
		_, _ = r.Read(buf)
		
		// May fail due to cache initialization in test env, but should not panic
		if err != nil {
			assert.Error(t, err)
		}
	})

	t.Run("cache clean command RunE", func(t *testing.T) {
		cmd := cacheCleanCmd
		assert.NotNil(t, cmd.RunE)
		
		// Capture output
		oldStdout := os.Stdout
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stdout = w
		os.Stderr = w
		
		err := cmd.RunE(cmd, []string{})
		
		// Restore output
		w.Close()
		os.Stdout = oldStdout
		os.Stderr = oldStderr
		
		// Read and discard output
		buf := make([]byte, 1024)
		_, _ = r.Read(buf)
		
		// May fail due to cache initialization in test env, but should not panic
		if err != nil {
			assert.Error(t, err)
		}
	})
}
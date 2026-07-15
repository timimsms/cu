package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocsCmd_Structure(t *testing.T) {
	t.Run("docs command exists", func(t *testing.T) {
		cmd := docsCmd
		assert.NotNil(t, cmd)
		assert.Equal(t, "docs", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
		assert.True(t, cmd.Hidden, "docs command should be hidden")
	})

	t.Run("docs command has subcommands", func(t *testing.T) {
		cmd := docsCmd
		subcommands := cmd.Commands()
		assert.NotEmpty(t, subcommands, "docs command should have subcommands")

		// Check for markdown subcommand
		var hasMarkdown bool
		for _, subcmd := range subcommands {
			if subcmd.Name() == "markdown" {
				hasMarkdown = true
				break
			}
		}
		assert.True(t, hasMarkdown, "Should have markdown subcommand")
	})
}

func TestGenMarkdownCmd_Structure(t *testing.T) {
	t.Run("markdown command exists", func(t *testing.T) {
		cmd := genMarkdownCmd
		assert.NotNil(t, cmd)
		assert.Equal(t, "markdown", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
		assert.NotNil(t, cmd.RunE)
	})

	t.Run("markdown command has dir flag", func(t *testing.T) {
		cmd := genMarkdownCmd
		dirFlag := cmd.Flags().Lookup("dir")
		assert.NotNil(t, dirFlag)
		assert.Equal(t, "d", dirFlag.Shorthand)
		assert.Equal(t, "./docs", dirFlag.DefValue)
		assert.Contains(t, dirFlag.Usage, "Directory")
	})
}

func TestGenMarkdownCmd_Execution(t *testing.T) {
	t.Run("creates directory if it doesn't exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		docsDir := filepath.Join(tmpDir, "new-docs")

		cmd := genMarkdownCmd
		_ = cmd.Flags().Set("dir", docsDir)

		err := cmd.RunE(cmd, []string{})
		assert.NoError(t, err)

		// Check directory was created
		info, err := os.Stat(docsDir)
		assert.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("uses default directory when not specified", func(t *testing.T) {
		// Create a temporary working directory
		tmpDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer func() { _ = os.Chdir(oldWd) }()
		_ = os.Chdir(tmpDir)

		cmd := genMarkdownCmd
		// Reset flag to default
		_ = cmd.Flags().Set("dir", "")

		err := cmd.RunE(cmd, []string{})
		assert.NoError(t, err)

		// Check default ./docs directory was created
		info, err := os.Stat("./docs")
		assert.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("generates documentation files", func(t *testing.T) {
		tmpDir := t.TempDir()

		cmd := genMarkdownCmd
		_ = cmd.Flags().Set("dir", tmpDir)

		err := cmd.RunE(cmd, []string{})
		assert.NoError(t, err)

		// Check that at least one markdown file was created
		files, err := os.ReadDir(tmpDir)
		assert.NoError(t, err)

		var hasMarkdownFile bool
		for _, file := range files {
			if filepath.Ext(file.Name()) == ".md" {
				hasMarkdownFile = true
				break
			}
		}
		assert.True(t, hasMarkdownFile, "Should have generated at least one markdown file")
	})

	t.Run("handles permission errors", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("permission bits are not enforced on Windows")
		}
		if os.Getuid() == 0 {
			t.Skip("Cannot test permission errors as root")
		}

		// Create a directory with no write permissions
		tmpDir := t.TempDir()
		readOnlyDir := filepath.Join(tmpDir, "readonly")
		err := os.Mkdir(readOnlyDir, 0555)
		assert.NoError(t, err)

		cmd := genMarkdownCmd
		_ = cmd.Flags().Set("dir", filepath.Join(readOnlyDir, "docs"))

		err = cmd.RunE(cmd, []string{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create directory")
	})
}

func TestDocsCmd_Integration(t *testing.T) {
	t.Run("docs command is added to root", func(t *testing.T) {
		// Check if docs command is in root's subcommands
		var foundDocs bool
		for _, cmd := range rootCmd.Commands() {
			if cmd.Name() == "docs" {
				foundDocs = true
				break
			}
		}
		assert.True(t, foundDocs, "docs command should be added to root command")
	})

	t.Run("markdown command is added to docs", func(t *testing.T) {
		// Check if markdown command is in docs' subcommands
		var foundMarkdown bool
		for _, cmd := range docsCmd.Commands() {
			if cmd.Name() == "markdown" {
				foundMarkdown = true
				break
			}
		}
		assert.True(t, foundMarkdown, "markdown command should be added to docs command")
	})
}

func TestDocsCmd_Output(t *testing.T) {
	t.Run("prints success message", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		cmd := genMarkdownCmd
		_ = cmd.Flags().Set("dir", tmpDir)

		err := cmd.RunE(cmd, []string{})
		assert.NoError(t, err)

		// Restore stdout and read output
		w.Close()
		os.Stdout = oldStdout

		buf := make([]byte, 1024)
		n, _ := r.Read(buf)
		output := string(buf[:n])

		assert.Contains(t, output, "Documentation generated")
		assert.Contains(t, output, tmpDir)
	})
}

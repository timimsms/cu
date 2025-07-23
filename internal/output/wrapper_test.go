package output

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock config for testing
type mockConfig struct {
	values map[string]string
}

func (m *mockConfig) GetString(key string) string {
	return m.values[key]
}

func TestNewFormatter(t *testing.T) {
	t.Run("creates formatter with config", func(t *testing.T) {
		config := &mockConfig{values: map[string]string{"output": "json"}}
		formatter := NewFormatter(config)
		
		assert.NotNil(t, formatter)
		assert.Equal(t, config, formatter.config)
		assert.True(t, formatter.colorOutput, "Should default to color output")
		assert.False(t, formatter.quietMode, "Should default to non-quiet mode")
	})

	t.Run("creates formatter with nil config", func(t *testing.T) {
		formatter := NewFormatter(nil)
		
		assert.NotNil(t, formatter)
		assert.Nil(t, formatter.config)
		assert.True(t, formatter.colorOutput)
		assert.False(t, formatter.quietMode)
	})
}

func TestFormatterWrapper_Print(t *testing.T) {
	testData := map[string]string{"key": "value"}

	t.Run("uses default table format", func(t *testing.T) {
		formatter := NewFormatter(nil)
		
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		
		err := formatter.Print(testData)
		
		w.Close()
		os.Stdout = oldStdout
		
		var buf bytes.Buffer
		io.Copy(&buf, r)
		
		assert.NoError(t, err)
		assert.NotEmpty(t, buf.String())
	})

	t.Run("uses config format", func(t *testing.T) {
		config := &mockConfig{values: map[string]string{"output": "json"}}
		formatter := NewFormatter(config)
		
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		
		err := formatter.Print(testData)
		
		w.Close()
		os.Stdout = oldStdout
		
		var buf bytes.Buffer
		io.Copy(&buf, r)
		
		assert.NoError(t, err)
		assert.Contains(t, buf.String(), "key")
		assert.Contains(t, buf.String(), "value")
	})
}

func TestFormatterWrapper_PrintTo(t *testing.T) {
	testData := map[string]string{"key": "value"}

	t.Run("prints to specified writer", func(t *testing.T) {
		var buf bytes.Buffer
		formatter := NewFormatter(nil)
		
		err := formatter.PrintTo(&buf, testData)
		
		assert.NoError(t, err)
		assert.NotEmpty(t, buf.String())
	})

	t.Run("uses config format when printing to writer", func(t *testing.T) {
		var buf bytes.Buffer
		config := &mockConfig{values: map[string]string{"output": "json"}}
		formatter := NewFormatter(config)
		
		err := formatter.PrintTo(&buf, testData)
		
		assert.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "key")
		assert.Contains(t, output, "value")
	})
}

func TestFormatterWrapper_PrintInfo(t *testing.T) {
	t.Run("prints info message when not quiet", func(t *testing.T) {
		formatter := NewFormatter(nil)
		
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		
		formatter.PrintInfo("test info")
		
		w.Close()
		os.Stdout = oldStdout
		
		var buf bytes.Buffer
		io.Copy(&buf, r)
		
		assert.Contains(t, buf.String(), "test info")
	})

	t.Run("does not print when quiet mode is enabled", func(t *testing.T) {
		formatter := NewFormatter(nil)
		formatter.SetQuiet(true)
		
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		
		formatter.PrintInfo("test info")
		
		w.Close()
		os.Stdout = oldStdout
		
		var buf bytes.Buffer
		io.Copy(&buf, r)
		
		assert.Empty(t, buf.String())
	})
}

func TestFormatterWrapper_PrintSuccess(t *testing.T) {
	t.Run("prints success message with check mark", func(t *testing.T) {
		formatter := NewFormatter(nil)
		
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		
		formatter.PrintSuccess("test success")
		
		w.Close()
		os.Stdout = oldStdout
		
		var buf bytes.Buffer
		io.Copy(&buf, r)
		
		output := buf.String()
		assert.Contains(t, output, "✓")
		assert.Contains(t, output, "test success")
	})

	t.Run("does not print when quiet mode is enabled", func(t *testing.T) {
		formatter := NewFormatter(nil)
		formatter.SetQuiet(true)
		
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		
		formatter.PrintSuccess("test success")
		
		w.Close()
		os.Stdout = oldStdout
		
		var buf bytes.Buffer
		io.Copy(&buf, r)
		
		assert.Empty(t, buf.String())
	})

	t.Run("prints without color when color is disabled", func(t *testing.T) {
		formatter := NewFormatter(nil)
		formatter.SetColor(false)
		
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		
		formatter.PrintSuccess("test success")
		
		w.Close()
		os.Stdout = oldStdout
		
		var buf bytes.Buffer
		io.Copy(&buf, r)
		
		output := buf.String()
		assert.Contains(t, output, "✓")
		assert.Contains(t, output, "test success")
	})
}

func TestFormatterWrapper_PrintError(t *testing.T) {
	t.Run("prints error message with X mark", func(t *testing.T) {
		formatter := NewFormatter(nil)
		testErr := errors.New("test error")
		
		// Capture stderr
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w
		
		formatter.PrintError(testErr)
		
		w.Close()
		os.Stderr = oldStderr
		
		var buf bytes.Buffer
		io.Copy(&buf, r)
		
		output := buf.String()
		assert.Contains(t, output, "✗")
		assert.Contains(t, output, "test error")
	})

	t.Run("prints error even in quiet mode", func(t *testing.T) {
		formatter := NewFormatter(nil)
		formatter.SetQuiet(true)
		testErr := errors.New("test error")
		
		// Capture stderr
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w
		
		formatter.PrintError(testErr)
		
		w.Close()
		os.Stderr = oldStderr
		
		var buf bytes.Buffer
		io.Copy(&buf, r)
		
		output := buf.String()
		assert.Contains(t, output, "✗")
		assert.Contains(t, output, "test error")
	})

	t.Run("prints without color when color is disabled", func(t *testing.T) {
		formatter := NewFormatter(nil)
		formatter.SetColor(false)
		testErr := errors.New("test error")
		
		// Capture stderr
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w
		
		formatter.PrintError(testErr)
		
		w.Close()
		os.Stderr = oldStderr
		
		var buf bytes.Buffer
		io.Copy(&buf, r)
		
		output := buf.String()
		assert.Contains(t, output, "✗")
		assert.Contains(t, output, "test error")
	})
}

func TestFormatterWrapper_PrintWarning(t *testing.T) {
	t.Run("prints warning message with warning sign", func(t *testing.T) {
		formatter := NewFormatter(nil)
		
		// Capture stderr
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w
		
		formatter.PrintWarning("test warning")
		
		w.Close()
		os.Stderr = oldStderr
		
		var buf bytes.Buffer
		io.Copy(&buf, r)
		
		output := buf.String()
		assert.Contains(t, output, "⚠")
		assert.Contains(t, output, "test warning")
	})

	t.Run("does not print when quiet mode is enabled", func(t *testing.T) {
		formatter := NewFormatter(nil)
		formatter.SetQuiet(true)
		
		// Capture stderr
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w
		
		formatter.PrintWarning("test warning")
		
		w.Close()
		os.Stderr = oldStderr
		
		var buf bytes.Buffer
		io.Copy(&buf, r)
		
		assert.Empty(t, buf.String())
	})

	t.Run("prints without color when color is disabled", func(t *testing.T) {
		formatter := NewFormatter(nil)
		formatter.SetColor(false)
		
		// Capture stderr
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w
		
		formatter.PrintWarning("test warning")
		
		w.Close()
		os.Stderr = oldStderr
		
		var buf bytes.Buffer
		io.Copy(&buf, r)
		
		output := buf.String()
		assert.Contains(t, output, "⚠")
		assert.Contains(t, output, "test warning")
	})
}

func TestFormatterWrapper_SetQuiet(t *testing.T) {
	t.Run("sets quiet mode", func(t *testing.T) {
		formatter := NewFormatter(nil)
		
		assert.False(t, formatter.quietMode)
		
		formatter.SetQuiet(true)
		assert.True(t, formatter.quietMode)
		
		formatter.SetQuiet(false)
		assert.False(t, formatter.quietMode)
	})
}

func TestFormatterWrapper_SetColor(t *testing.T) {
	t.Run("sets color mode", func(t *testing.T) {
		formatter := NewFormatter(nil)
		
		assert.True(t, formatter.colorOutput)
		
		formatter.SetColor(false)
		assert.False(t, formatter.colorOutput)
		
		formatter.SetColor(true)
		assert.True(t, formatter.colorOutput)
	})
}

func TestFormatterWrapper_GetFormat(t *testing.T) {
	t.Run("returns default table format", func(t *testing.T) {
		formatter := NewFormatter(nil)
		assert.Equal(t, "table", formatter.GetFormat())
	})

	t.Run("returns format from config", func(t *testing.T) {
		config := &mockConfig{values: map[string]string{"output": "json"}}
		formatter := NewFormatter(config)
		assert.Equal(t, "json", formatter.GetFormat())
	})

	t.Run("returns default when config has empty format", func(t *testing.T) {
		config := &mockConfig{values: map[string]string{"output": ""}}
		formatter := NewFormatter(config)
		assert.Equal(t, "table", formatter.GetFormat())
	})
}

func TestFormatterWrapper_SetFormat(t *testing.T) {
	formatter := NewFormatter(nil)

	t.Run("accepts valid formats", func(t *testing.T) {
		validFormats := []string{"json", "yaml", "table", "csv"}
		
		for _, format := range validFormats {
			err := formatter.SetFormat(format)
			assert.NoError(t, err, "Should accept %s format", format)
		}
	})

	t.Run("rejects invalid formats", func(t *testing.T) {
		err := formatter.SetFormat("invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid format")
	})
}

func TestFormatterWrapper_SetTableHeader(t *testing.T) {
	t.Run("accepts table headers", func(t *testing.T) {
		formatter := NewFormatter(nil)
		headers := []string{"ID", "Name", "Status"}
		
		// Should not panic
		formatter.SetTableHeader(headers)
		
		// Currently a no-op, so we just test it doesn't crash
		assert.NotNil(t, formatter)
	})
}

func TestFormatterWrapper_Integration(t *testing.T) {
	t.Run("full workflow with all methods", func(t *testing.T) {
		config := &mockConfig{values: map[string]string{"output": "json"}}
		formatter := NewFormatter(config)
		
		// Configure formatter
		formatter.SetQuiet(false)
		formatter.SetColor(true)
		
		// Test format operations
		assert.Equal(t, "json", formatter.GetFormat())
		
		err := formatter.SetFormat("yaml")
		assert.NoError(t, err)
		
		// Test table headers (no-op currently)
		formatter.SetTableHeader([]string{"A", "B", "C"})
		
		// Test data printing
		testData := map[string]string{"test": "data"}
		var buf bytes.Buffer
		err = formatter.PrintTo(&buf, testData)
		assert.NoError(t, err)
		assert.NotEmpty(t, buf.String())
	})
}

func TestFormatterWrapper_EdgeCases(t *testing.T) {
	t.Run("handles nil data gracefully", func(t *testing.T) {
		formatter := NewFormatter(nil)
		var buf bytes.Buffer
		
		err := formatter.PrintTo(&buf, nil)
		// Should handle nil without crashing
		assert.NoError(t, err)
	})

	t.Run("handles empty string format in SetFormat", func(t *testing.T) {
		formatter := NewFormatter(nil)
		err := formatter.SetFormat("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid format")
	})

	t.Run("handles case variations in SetFormat", func(t *testing.T) {
		formatter := NewFormatter(nil)
		
		// These should be valid (implementation doesn't do case conversion in SetFormat)
		err := formatter.SetFormat("JSON")
		assert.Error(t, err) // Current implementation is case-sensitive
		
		err = formatter.SetFormat("json")
		assert.NoError(t, err)
	})
}
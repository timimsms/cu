package output

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
)

// FormatterWrapper wraps the output functionality to implement the OutputFormatter interface
type FormatterWrapper struct {
	config      interface{ GetString(string) string }
	quietMode   bool
	colorOutput bool
}

// NewFormatter creates a new output formatter with config
func NewFormatter(config interface{ GetString(string) string }) *FormatterWrapper {
	return &FormatterWrapper{
		config:      config,
		colorOutput: true, // Default to color output
	}
}

// Print formats and prints data according to the configured format
func (f *FormatterWrapper) Print(data interface{}) error {
	format := "table" // default
	if f.config != nil {
		if fmt := f.config.GetString("output"); fmt != "" {
			format = fmt
		}
	}

	return Format(format, data)
}

// PrintTo formats and prints data to the specified writer
func (f *FormatterWrapper) PrintTo(w io.Writer, data interface{}) error {
	format := "table" // default
	if f.config != nil {
		if fmt := f.config.GetString("output"); fmt != "" {
			format = fmt
		}
	}

	// Temporarily redirect output to the provided writer
	oldStdout := os.Stdout
	r, w2, _ := os.Pipe()
	os.Stdout = w2

	err := Format(format, data)

	// Restore stdout and copy the output
	_ = w2.Close()
	os.Stdout = oldStdout
	_, _ = io.Copy(w, r)

	return err
}

// PrintInfo prints an informational message
func (f *FormatterWrapper) PrintInfo(msg string) {
	if f.quietMode {
		return
	}

	if f.colorOutput {
		_, _ = fmt.Fprintln(os.Stdout, msg)
	} else {
		_, _ = fmt.Fprintln(os.Stdout, msg)
	}
}

// PrintSuccess prints a success message
func (f *FormatterWrapper) PrintSuccess(msg string) {
	if f.quietMode {
		return
	}

	if f.colorOutput {
		green := color.New(color.FgGreen)
		_, _ = green.Fprintf(os.Stdout, "✓ %s\n", msg)
	} else {
		_, _ = fmt.Fprintf(os.Stdout, "✓ %s\n", msg)
	}
}

// PrintError prints an error message
func (f *FormatterWrapper) PrintError(err error) {
	msg := err.Error()
	if f.colorOutput {
		red := color.New(color.FgRed)
		_, _ = red.Fprintf(os.Stderr, "✗ %s\n", msg)
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "✗ %s\n", msg)
	}
}

// PrintWarning prints a warning message
func (f *FormatterWrapper) PrintWarning(msg string) {
	if f.quietMode {
		return
	}

	if f.colorOutput {
		yellow := color.New(color.FgYellow)
		_, _ = yellow.Fprintf(os.Stderr, "⚠ %s\n", msg)
	} else {
		_, _ = fmt.Fprintf(os.Stderr, "⚠ %s\n", msg)
	}
}

// SetQuiet sets quiet mode
func (f *FormatterWrapper) SetQuiet(quiet bool) {
	f.quietMode = quiet
}

// SetColor sets color output mode
func (f *FormatterWrapper) SetColor(useColor bool) {
	f.colorOutput = useColor
}

// GetFormat returns the current output format
func (f *FormatterWrapper) GetFormat() string {
	format := "table" // default
	if f.config != nil {
		if fmt := f.config.GetString("output"); fmt != "" {
			format = fmt
		}
	}
	return format
}

// SetFormat sets the output format
func (f *FormatterWrapper) SetFormat(format string) error {
	// Validate format
	switch format {
	case "json", "yaml", "table", "csv":
		// Valid formats - store it if we have a way to persist it
		// For now, this is a no-op since we read from config
		return nil
	default:
		return fmt.Errorf("invalid format: %s", format)
	}
}

// SetTableHeader sets the table header (for table format)
func (f *FormatterWrapper) SetTableHeader(headers []string) {
	// Store table headers for later use
	// For now, this is a no-op since the Format function handles headers
}

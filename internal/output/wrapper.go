package output

import (
	"fmt"
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

// PrintInfo prints an informational message
func (f *FormatterWrapper) PrintInfo(msg string) {
	if f.quietMode {
		return
	}
	
	if f.colorOutput {
		fmt.Fprintln(os.Stdout, msg)
	} else {
		fmt.Fprintln(os.Stdout, msg)
	}
}

// PrintSuccess prints a success message
func (f *FormatterWrapper) PrintSuccess(msg string) {
	if f.quietMode {
		return
	}
	
	if f.colorOutput {
		color.Green("✓ %s", msg)
	} else {
		fmt.Fprintf(os.Stdout, "✓ %s\n", msg)
	}
}

// PrintError prints an error message
func (f *FormatterWrapper) PrintError(msg string) {
	if f.colorOutput {
		color.Red("✗ %s", msg)
	} else {
		fmt.Fprintf(os.Stderr, "✗ %s\n", msg)
	}
}

// PrintWarning prints a warning message
func (f *FormatterWrapper) PrintWarning(msg string) {
	if f.quietMode {
		return
	}
	
	if f.colorOutput {
		color.Yellow("⚠ %s", msg)
	} else {
		fmt.Fprintf(os.Stderr, "⚠ %s\n", msg)
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
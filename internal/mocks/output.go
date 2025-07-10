package mocks

import (
	"fmt"
	"io"
)

// MockOutputFormatter is a mock implementation of OutputFormatter for testing
type MockOutputFormatter struct {
	// Storage for captured outputs
	Printed      []interface{}
	Errors       []error
	SuccessMsg   []string
	WarningMsg   []string
	InfoMsg      []string
	Format       string
	ColorEnabled bool
	QuietMode    bool
	Headers      []string

	// Control behavior
	PrintErr    error  // Renamed to avoid conflict with method
	FormatError error
}

// NewMockOutputFormatter creates a new mock output formatter
func NewMockOutputFormatter() *MockOutputFormatter {
	return &MockOutputFormatter{
		Printed:    make([]interface{}, 0),
		Errors:     make([]error, 0),
		SuccessMsg: make([]string, 0),
		WarningMsg: make([]string, 0),
		InfoMsg:    make([]string, 0),
		Format:     "table", // default format
	}
}

// Print captures the data being printed
func (m *MockOutputFormatter) Print(data interface{}) error {
	if m.PrintErr != nil {
		return m.PrintErr
	}
	m.Printed = append(m.Printed, data)
	return nil
}

// PrintTo captures the data being printed to a writer
func (m *MockOutputFormatter) PrintTo(w io.Writer, data interface{}) error {
	if m.PrintErr != nil {
		return m.PrintErr
	}
	m.Printed = append(m.Printed, data)
	// Actually write to the writer for testing
	_, err := fmt.Fprintf(w, "%v", data)
	return err
}

// PrintError captures error messages
func (m *MockOutputFormatter) PrintError(err error) {
	m.Errors = append(m.Errors, err)
}

// PrintSuccess captures success messages
func (m *MockOutputFormatter) PrintSuccess(message string) {
	m.SuccessMsg = append(m.SuccessMsg, message)
}

// PrintWarning captures warning messages
func (m *MockOutputFormatter) PrintWarning(message string) {
	m.WarningMsg = append(m.WarningMsg, message)
}

// PrintInfo captures info messages
func (m *MockOutputFormatter) PrintInfo(message string) {
	m.InfoMsg = append(m.InfoMsg, message)
}

// SetFormat sets the output format
func (m *MockOutputFormatter) SetFormat(format string) error {
	if m.FormatError != nil {
		return m.FormatError
	}
	m.Format = format
	return nil
}

// GetFormat returns the current format
func (m *MockOutputFormatter) GetFormat() string {
	return m.Format
}

// SetColor sets color output
func (m *MockOutputFormatter) SetColor(enabled bool) {
	m.ColorEnabled = enabled
}

// SetQuiet sets quiet mode
func (m *MockOutputFormatter) SetQuiet(enabled bool) {
	m.QuietMode = enabled
}

// SetTableHeader sets table headers
func (m *MockOutputFormatter) SetTableHeader(headers []string) {
	m.Headers = headers
}

// Reset clears all captured data
func (m *MockOutputFormatter) Reset() {
	m.Printed = make([]interface{}, 0)
	m.Errors = make([]error, 0)
	m.SuccessMsg = make([]string, 0)
	m.WarningMsg = make([]string, 0)
	m.InfoMsg = make([]string, 0)
	m.Headers = nil
}
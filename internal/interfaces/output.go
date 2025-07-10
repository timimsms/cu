package interfaces

import "io"

// OutputFormatter defines the interface for output formatting
type OutputFormatter interface {
	// Print methods
	Print(data interface{}) error
	PrintTo(w io.Writer, data interface{}) error
	PrintError(err error)
	PrintSuccess(message string)
	PrintWarning(message string)
	PrintInfo(message string)

	// Configuration
	SetFormat(format string) error
	GetFormat() string
	SetColor(enabled bool)
	SetQuiet(enabled bool)
	
	// Table-specific (for table format)
	SetTableHeader(headers []string)
}
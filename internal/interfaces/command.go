package interfaces

import (
	"context"
	"github.com/spf13/cobra"
)

// Command defines the interface for all CLI commands
type Command interface {
	// Execute runs the command with the given context and arguments
	Execute(ctx context.Context, args []string) error
	
	// GetCobraCommand returns the underlying cobra command for integration
	GetCobraCommand() *cobra.Command
	
	// Setup initializes the command (flags, description, etc.)
	Setup()
}
package factory

import (
	"fmt"

	"github.com/tim/cu/internal/interfaces"
)

// Factory creates commands with injected dependencies
type Factory struct {
	api    interfaces.APIClient
	auth   interfaces.AuthManager
	output interfaces.OutputFormatter
	config interfaces.ConfigProvider
}

// New creates a new command factory
func New(options ...Option) *Factory {
	f := &Factory{}
	
	// Apply options
	for _, opt := range options {
		opt(f)
	}
	
	return f
}

// Option is a functional option for configuring the factory
type Option func(*Factory)

// WithAPIClient sets the API client
func WithAPIClient(client interfaces.APIClient) Option {
	return func(f *Factory) {
		f.api = client
	}
}

// WithAuthManager sets the auth manager
func WithAuthManager(auth interfaces.AuthManager) Option {
	return func(f *Factory) {
		f.auth = auth
	}
}

// WithOutputFormatter sets the output formatter
func WithOutputFormatter(output interfaces.OutputFormatter) Option {
	return func(f *Factory) {
		f.output = output
	}
}

// WithConfigProvider sets the config provider
func WithConfigProvider(config interfaces.ConfigProvider) Option {
	return func(f *Factory) {
		f.config = config
	}
}

// CreateCommand creates a command by name
func (f *Factory) CreateCommand(name string) (interfaces.Command, error) {
	switch name {
	case "version":
		return f.createVersionCommand(), nil
	case "completion":
		return f.createCompletionCommand(), nil
	case "interactive":
		return f.createInteractiveCommand(), nil
	case "config":
		return f.createConfigCommand(), nil
	case "auth":
		return f.createAuthCommand(), nil
	case "task":
		return f.createTaskCommand(), nil
	case "space":
		return f.createSpaceCommand(), nil
	case "list":
		return f.createListCommand(), nil
	default:
		return nil, fmt.Errorf("unknown command: %s", name)
	}
}

// Command creation methods will be implemented in separate files
// These are placeholder declarations that will be implemented
// when we refactor each command

func (f *Factory) createAuthCommand() interfaces.Command {
	// Will be implemented in auth.go
	return nil
}



func (f *Factory) createListCommand() interfaces.Command {
	// Will be implemented in list.go
	return nil
}
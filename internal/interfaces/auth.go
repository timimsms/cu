package interfaces

import "github.com/tim/cu/internal/auth"

// AuthManager defines the interface for authentication operations
type AuthManager interface {
	// Token management
	GetToken(workspace string) (*auth.Token, error)
	SaveToken(workspace string, token *auth.Token) error
	DeleteToken(workspace string) error
	GetCurrentToken() (*auth.Token, error)

	// Workspace operations
	ListWorkspaces() ([]string, error)
	IsAuthenticated(workspace string) bool
}
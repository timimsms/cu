package mocks

import (
	"github.com/tim/cu/internal/auth"
)

// MockAuthManager is a mock implementation of AuthManager for testing
type MockAuthManager struct {
	// SaveToken tracking
	SaveTokenCalled bool
	SavedWorkspace  string
	SavedToken      *auth.Token
	SaveTokenErr    error

	// GetToken tracking
	GetTokenCalled    bool
	GetTokenWorkspace string
	GetTokenResult    *auth.Token
	GetTokenErr       error

	// DeleteToken tracking
	DeleteTokenCalled bool
	DeletedWorkspace  string
	DeleteTokenErr    error

	// GetCurrentToken tracking
	GetCurrentTokenCalled bool
	GetCurrentTokenResult *auth.Token
	GetCurrentTokenErr    error

	// IsAuthenticated tracking
	IsAuthenticatedCalled    bool
	IsAuthenticatedWorkspace string
	IsAuthenticatedResult    bool

	// ListWorkspaces tracking
	ListWorkspacesCalled bool
	ListWorkspacesResult []string
	ListWorkspacesErr    error
}

// SaveToken saves an auth token
func (m *MockAuthManager) SaveToken(workspace string, token *auth.Token) error {
	m.SaveTokenCalled = true
	m.SavedWorkspace = workspace
	m.SavedToken = token
	return m.SaveTokenErr
}

// GetToken retrieves an auth token
func (m *MockAuthManager) GetToken(workspace string) (*auth.Token, error) {
	m.GetTokenCalled = true
	m.GetTokenWorkspace = workspace
	return m.GetTokenResult, m.GetTokenErr
}

// DeleteToken removes an auth token
func (m *MockAuthManager) DeleteToken(workspace string) error {
	m.DeleteTokenCalled = true
	m.DeletedWorkspace = workspace
	return m.DeleteTokenErr
}

// GetCurrentToken gets the current token
func (m *MockAuthManager) GetCurrentToken() (*auth.Token, error) {
	m.GetCurrentTokenCalled = true
	return m.GetCurrentTokenResult, m.GetCurrentTokenErr
}

// IsAuthenticated checks if the user is authenticated
func (m *MockAuthManager) IsAuthenticated(workspace string) bool {
	m.IsAuthenticatedCalled = true
	m.IsAuthenticatedWorkspace = workspace
	return m.IsAuthenticatedResult
}

// ListWorkspaces returns all workspaces
func (m *MockAuthManager) ListWorkspaces() ([]string, error) {
	m.ListWorkspacesCalled = true
	return m.ListWorkspacesResult, m.ListWorkspacesErr
}

// HasToken checks if a token exists
func (m *MockAuthManager) HasToken(workspace string) bool {
	return m.GetTokenResult != nil && m.GetTokenErr == nil
}

// ListTokens returns all stored tokens (legacy method)
func (m *MockAuthManager) ListTokens() (map[string]*auth.Token, error) {
	if m.GetTokenResult != nil {
		return map[string]*auth.Token{
			m.GetTokenWorkspace: m.GetTokenResult,
		}, nil
	}
	return map[string]*auth.Token{}, nil
}

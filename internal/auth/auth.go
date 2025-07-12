package auth

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tim/cu/internal/errors"
	"github.com/zalando/go-keyring"
)

const (
	// ServiceName is the keyring service name
	ServiceName = "cu-cli"
	// DefaultWorkspace is the default workspace key
	DefaultWorkspace = "default"
)

// Token represents an authentication token
type Token struct {
	Value     string `json:"value"`
	Workspace string `json:"workspace,omitempty"`
	Email     string `json:"email,omitempty"`
}

// Manager handles authentication
type Manager struct {
	service string
	config  interface{ GetString(string) string }
}

// NewManager creates a new authentication manager
func NewManager(config interface{ GetString(string) string }) *Manager {
	return &Manager{
		service: ServiceName,
		config:  config,
	}
}

// SaveToken saves a token to the keyring
func (m *Manager) SaveToken(workspace string, token *Token) error {
	if workspace == "" {
		workspace = DefaultWorkspace
	}

	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	err = keyring.Set(m.service, workspace, string(data))
	if err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	return nil
}

// GetToken retrieves a token from the keyring
func (m *Manager) GetToken(workspace string) (*Token, error) {
	if workspace == "" {
		workspace = DefaultWorkspace
	}

	data, err := keyring.Get(m.service, workspace)
	if err != nil {
		if err == keyring.ErrNotFound {
			return nil, errors.ErrNotAuthenticated
		}
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	var token Token
	if err := json.Unmarshal([]byte(data), &token); err != nil {
		// Handle legacy format (just the token string)
		if !strings.Contains(data, "{") {
			token = Token{Value: data, Workspace: workspace}
		} else {
			return nil, fmt.Errorf("failed to unmarshal token: %w", err)
		}
	}

	return &token, nil
}

// DeleteToken removes a token from the keyring
func (m *Manager) DeleteToken(workspace string) error {
	if workspace == "" {
		workspace = DefaultWorkspace
	}

	err := keyring.Delete(m.service, workspace)
	if err != nil && err != keyring.ErrNotFound {
		return fmt.Errorf("failed to delete token: %w", err)
	}

	return nil
}

// ListWorkspaces returns all workspaces with stored tokens
func (m *Manager) ListWorkspaces() ([]string, error) {
	// Note: go-keyring doesn't support listing all accounts
	// This is a limitation we'll need to work around by storing
	// workspace list separately in config
	return []string{DefaultWorkspace}, nil
}

// IsAuthenticated checks if the user is authenticated
func (m *Manager) IsAuthenticated(workspace string) bool {
	_, err := m.GetToken(workspace)
	return err == nil
}

// GetCurrentToken gets the token for the current workspace
func (m *Manager) GetCurrentToken() (*Token, error) {
	workspace := DefaultWorkspace
	if m.config != nil {
		if ws := m.config.GetString("workspace"); ws != "" {
			workspace = ws
		}
	}
	return m.GetToken(workspace)
}

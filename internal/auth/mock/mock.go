// Package mock provides mock implementations for auth testing
package mock

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/tim/cu/internal/auth"
	cuerrors "github.com/tim/cu/internal/errors"
)

// AuthProvider is a mock implementation of auth.Manager for testing
type AuthProvider struct {
	mu              sync.RWMutex
	tokens          map[string]*auth.Token
	errors          map[string]error
	authenticated   map[string]bool
	workspaces      []string
	currentWorkspace string
	
	// Behavior controls
	saveError       error
	getError        error
	deleteError     error
	listError       error
	
	// Advanced behaviors
	tokenExpiry     map[string]time.Time
	refreshBehavior func(workspace string) (*auth.Token, error)
	
	// Call tracking
	calls           []string
}

// NewAuthProvider creates a new mock auth provider
func NewAuthProvider() *AuthProvider {
	return &AuthProvider{
		tokens:          make(map[string]*auth.Token),
		errors:          make(map[string]error),
		authenticated:   make(map[string]bool),
		tokenExpiry:     make(map[string]time.Time),
		workspaces:      []string{},
		currentWorkspace: auth.DefaultWorkspace,
		calls:           []string{},
	}
}

// SaveToken mocks saving a token
func (m *AuthProvider) SaveToken(workspace string, token *auth.Token) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("SaveToken(%s)", workspace))
	
	if m.saveError != nil {
		return m.saveError
	}
	
	if workspace == "" {
		workspace = auth.DefaultWorkspace
	}
	
	m.tokens[workspace] = token
	m.authenticated[workspace] = true
	
	// Update workspaces list if new
	found := false
	for _, w := range m.workspaces {
		if w == workspace {
			found = true
			break
		}
	}
	if !found {
		m.workspaces = append(m.workspaces, workspace)
	}
	
	return nil
}

// GetToken mocks retrieving a token
func (m *AuthProvider) GetToken(workspace string) (*auth.Token, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	m.calls = append(m.calls, fmt.Sprintf("GetToken(%s)", workspace))
	
	if m.getError != nil {
		return nil, m.getError
	}
	
	if workspace == "" {
		workspace = auth.DefaultWorkspace
	}
	
	// Check for specific workspace error
	if err, ok := m.errors[workspace]; ok {
		return nil, err
	}
	
	// Check token expiry
	if expiry, ok := m.tokenExpiry[workspace]; ok {
		if time.Now().After(expiry) {
			if m.refreshBehavior != nil {
				return m.refreshBehavior(workspace)
			}
			return nil, cuerrors.ErrTokenExpired
		}
	}
	
	token, ok := m.tokens[workspace]
	if !ok {
		return nil, cuerrors.ErrNotAuthenticated
	}
	
	return token, nil
}

// DeleteToken mocks deleting a token
func (m *AuthProvider) DeleteToken(workspace string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.calls = append(m.calls, fmt.Sprintf("DeleteToken(%s)", workspace))
	
	if m.deleteError != nil {
		return m.deleteError
	}
	
	if workspace == "" {
		workspace = auth.DefaultWorkspace
	}
	
	delete(m.tokens, workspace)
	delete(m.authenticated, workspace)
	delete(m.tokenExpiry, workspace)
	
	// Remove from workspaces list
	newWorkspaces := []string{}
	for _, w := range m.workspaces {
		if w != workspace {
			newWorkspaces = append(newWorkspaces, w)
		}
	}
	m.workspaces = newWorkspaces
	
	return nil
}

// ListWorkspaces mocks listing workspaces
func (m *AuthProvider) ListWorkspaces() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	m.calls = append(m.calls, "ListWorkspaces()")
	
	if m.listError != nil {
		return nil, m.listError
	}
	
	return m.workspaces, nil
}

// IsAuthenticated mocks checking authentication status
func (m *AuthProvider) IsAuthenticated(workspace string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	m.calls = append(m.calls, fmt.Sprintf("IsAuthenticated(%s)", workspace))
	
	if workspace == "" {
		workspace = auth.DefaultWorkspace
	}
	
	// Check if token exists and not expired
	if _, err := m.GetToken(workspace); err != nil {
		return false
	}
	
	return m.authenticated[workspace]
}

// GetCurrentToken mocks getting the current workspace token
func (m *AuthProvider) GetCurrentToken() (*auth.Token, error) {
	m.mu.RLock()
	workspace := m.currentWorkspace
	m.mu.RUnlock()
	
	m.calls = append(m.calls, "GetCurrentToken()")
	
	return m.GetToken(workspace)
}

// Helper methods for test setup

// SetToken sets a token for testing
func (m *AuthProvider) SetToken(workspace string, token string, expiry time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if workspace == "" {
		workspace = auth.DefaultWorkspace
	}
	
	m.tokens[workspace] = &auth.Token{
		Value:     token,
		Workspace: workspace,
	}
	m.authenticated[workspace] = true
	
	if !expiry.IsZero() {
		m.tokenExpiry[workspace] = expiry
	}
	
	// Update workspaces list
	found := false
	for _, w := range m.workspaces {
		if w == workspace {
			found = true
			break
		}
	}
	if !found {
		m.workspaces = append(m.workspaces, workspace)
	}
}

// SetTokenWithEmail sets a token with email for testing
func (m *AuthProvider) SetTokenWithEmail(workspace, token, email string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if workspace == "" {
		workspace = auth.DefaultWorkspace
	}
	
	m.tokens[workspace] = &auth.Token{
		Value:     token,
		Workspace: workspace,
		Email:     email,
	}
	m.authenticated[workspace] = true
	
	// Update workspaces list
	found := false
	for _, w := range m.workspaces {
		if w == workspace {
			found = true
			break
		}
	}
	if !found {
		m.workspaces = append(m.workspaces, workspace)
	}
}

// SetError sets an error for a specific workspace
func (m *AuthProvider) SetError(workspace string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if workspace == "" {
		workspace = auth.DefaultWorkspace
	}
	
	m.errors[workspace] = err
}

// SetSaveError sets error for SaveToken calls
func (m *AuthProvider) SetSaveError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saveError = err
}

// SetGetError sets error for GetToken calls
func (m *AuthProvider) SetGetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getError = err
}

// SetDeleteError sets error for DeleteToken calls
func (m *AuthProvider) SetDeleteError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deleteError = err
}

// SetListError sets error for ListWorkspaces calls
func (m *AuthProvider) SetListError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listError = err
}

// SetRefreshBehavior sets custom token refresh behavior
func (m *AuthProvider) SetRefreshBehavior(fn func(workspace string) (*auth.Token, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.refreshBehavior = fn
}

// SetCurrentWorkspace sets the current workspace
func (m *AuthProvider) SetCurrentWorkspace(workspace string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentWorkspace = workspace
}

// GetCalls returns the list of method calls made
func (m *AuthProvider) GetCalls() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	calls := make([]string, len(m.calls))
	copy(calls, m.calls)
	return calls
}

// Reset clears all state and call history
func (m *AuthProvider) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.tokens = make(map[string]*auth.Token)
	m.errors = make(map[string]error)
	m.authenticated = make(map[string]bool)
	m.tokenExpiry = make(map[string]time.Time)
	m.workspaces = []string{}
	m.currentWorkspace = auth.DefaultWorkspace
	m.calls = []string{}
	
	m.saveError = nil
	m.getError = nil
	m.deleteError = nil
	m.listError = nil
	m.refreshBehavior = nil
}

// KeyringMock provides a mock implementation of the keyring interface
type KeyringMock struct {
	mu    sync.RWMutex
	store map[string]map[string]string // service -> account -> secret
	err   error
}

// NewKeyringMock creates a new keyring mock
func NewKeyringMock() *KeyringMock {
	return &KeyringMock{
		store: make(map[string]map[string]string),
	}
}

// Get retrieves a secret from the mock keyring
func (k *KeyringMock) Get(service, account string) (string, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	
	if k.err != nil {
		return "", k.err
	}
	
	if serviceStore, ok := k.store[service]; ok {
		if secret, ok := serviceStore[account]; ok {
			return secret, nil
		}
	}
	
	return "", fmt.Errorf("secret not found in keyring")
}

// Set stores a secret in the mock keyring
func (k *KeyringMock) Set(service, account, secret string) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	
	if k.err != nil {
		return k.err
	}
	
	if _, ok := k.store[service]; !ok {
		k.store[service] = make(map[string]string)
	}
	
	k.store[service][account] = secret
	return nil
}

// Delete removes a secret from the mock keyring
func (k *KeyringMock) Delete(service, account string) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	
	if k.err != nil {
		return k.err
	}
	
	if serviceStore, ok := k.store[service]; ok {
		delete(serviceStore, account)
		if len(serviceStore) == 0 {
			delete(k.store, service)
		}
	}
	
	return nil
}

// SetError sets an error to be returned by all operations
func (k *KeyringMock) SetError(err error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.err = err
}

// Reset clears all stored secrets and errors
func (k *KeyringMock) Reset() {
	k.mu.Lock()
	defer k.mu.Unlock()
	
	k.store = make(map[string]map[string]string)
	k.err = nil
}

// StoreJSON stores a JSON-serializable value
func (k *KeyringMock) StoreJSON(service, account string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return k.Set(service, account, string(data))
}
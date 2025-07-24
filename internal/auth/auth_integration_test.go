//go:build integration
// +build integration

package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration test that requires real keyring access
func TestIntegrationWithKeyring(t *testing.T) {
	config := &mockConfig{values: make(map[string]string)}
	m := NewManager(config)
	workspace := "test-workspace"

	// Clean up before test
	_ = m.DeleteToken(workspace)

	// Test save
	token := &Token{
		Value:       "test-token-123",
		WorkspaceID: workspace,
		UserID:      "user123",
		UserEmail:   "test@example.com",
	}

	err := m.SaveToken(token)
	require.NoError(t, err)

	// Test get
	retrieved, err := m.GetToken(workspace)
	require.NoError(t, err)
	assert.Equal(t, token.Value, retrieved.Value)
	assert.Equal(t, token.WorkspaceID, retrieved.WorkspaceID)
	assert.Equal(t, token.UserID, retrieved.UserID)
	assert.Equal(t, token.UserEmail, retrieved.UserEmail)

	// Test list
	workspaces := m.ListWorkspaces()
	assert.Contains(t, workspaces, workspace)

	// Test delete
	err = m.DeleteToken(workspace)
	require.NoError(t, err)

	// Verify deleted
	_, err = m.GetToken(workspace)
	assert.Error(t, err)
}
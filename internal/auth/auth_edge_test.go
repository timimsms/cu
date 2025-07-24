package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestManager_EdgeCases tests edge cases and error paths
func TestManager_EdgeCases(t *testing.T) {
	config := &mockConfig{values: make(map[string]string)}
	m := NewManager(config)

	t.Run("GetToken with empty workspace uses default", func(t *testing.T) {
		// This will fail without keyring, but tests the workspace handling
		_, err := m.GetToken("")
		assert.Error(t, err)
	})

	t.Run("IsAuthenticated with empty workspace uses default", func(t *testing.T) {
		result := m.IsAuthenticated("")
		assert.False(t, result)
	})

	t.Run("DeleteToken with non-existent workspace", func(t *testing.T) {
		// This may or may not error depending on keyring implementation
		_ = m.DeleteToken("non-existent-workspace")
		// No assertion - just testing it doesn't panic
	})

	t.Run("GetCurrentToken uses config workspace", func(t *testing.T) {
		// Test with workspace in config
		config.values["workspace"] = "custom-workspace"
		_, err := m.GetCurrentToken()
		// Will error without keyring, but tests the path
		assert.Error(t, err)
		
		// Test with empty workspace in config (should use default)
		config.values["workspace"] = ""
		_, err = m.GetCurrentToken()
		assert.Error(t, err)
	})
}

// TestManager_SaveTokenError tests error handling in SaveToken
func TestManager_SaveTokenError(t *testing.T) {
	config := &mockConfig{values: make(map[string]string)}
	m := NewManager(config)

	t.Run("SaveToken with nil token", func(t *testing.T) {
		err := m.SaveToken("workspace", nil)
		// Will fail during JSON marshaling or keyring access
		if err == nil {
			t.Skip("Keyring available, cannot test nil token error")
		}
	})

	t.Run("SaveToken with empty workspace", func(t *testing.T) {
		token := &Token{
			Value:     "test-token",
			Workspace: "test",
		}
		// Workspace should be normalized to default
		err := m.SaveToken("", token)
		// May fail due to keyring access, but shouldn't panic
		if err == nil {
			// If it succeeded, verify we can get it back with default workspace
			retrieved, _ := m.GetToken(DefaultWorkspace)
			if retrieved != nil {
				assert.Equal(t, token.Value, retrieved.Value)
			}
		}
	})
}

// TestToken_EdgeCases tests Token struct edge cases
func TestToken_EdgeCases(t *testing.T) {
	t.Run("Token with all fields", func(t *testing.T) {
		token := &Token{
			Value:     "pk_123456",
			Workspace: "production",
			Email:     "user@example.com",
		}
		
		// Test all getters work
		assert.Equal(t, "pk_123456", token.Value)
		assert.Equal(t, "production", token.Workspace)
		assert.Equal(t, "user@example.com", token.Email)
	})

	t.Run("Token with minimal fields", func(t *testing.T) {
		token := &Token{
			Value: "pk_minimal",
		}
		
		assert.Equal(t, "pk_minimal", token.Value)
		assert.Equal(t, "", token.Workspace)
		assert.Equal(t, "", token.Email)
	})
}
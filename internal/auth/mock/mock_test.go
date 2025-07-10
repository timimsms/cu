package mock

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tim/cu/internal/auth"
	cuerrors "github.com/tim/cu/internal/errors"
)

func TestNewAuthProvider(t *testing.T) {
	provider := NewAuthProvider()
	
	assert.NotNil(t, provider)
	assert.NotNil(t, provider.tokens)
	assert.NotNil(t, provider.errors)
	assert.NotNil(t, provider.authenticated)
	assert.NotNil(t, provider.tokenExpiry)
	assert.Empty(t, provider.workspaces)
	assert.Equal(t, auth.DefaultWorkspace, provider.currentWorkspace)
	assert.Empty(t, provider.calls)
}

func TestAuthProviderSaveToken(t *testing.T) {
	provider := NewAuthProvider()
	
	t.Run("successful save", func(t *testing.T) {
		token := &auth.Token{
			Value:     "test-token",
			Workspace: "prod",
			Email:     "user@example.com",
		}
		
		err := provider.SaveToken("prod", token)
		require.NoError(t, err)
		
		// Verify token was saved
		provider.mu.RLock()
		saved := provider.tokens["prod"]
		provider.mu.RUnlock()
		
		assert.Equal(t, token.Value, saved.Value)
		assert.True(t, provider.IsAuthenticated("prod"))
		assert.Contains(t, provider.GetCalls(), "SaveToken(prod)")
	})
	
	t.Run("save with error", func(t *testing.T) {
		provider.SetSaveError(errors.New("save failed"))
		
		token := &auth.Token{Value: "test"}
		err := provider.SaveToken("workspace", token)
		
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "save failed")
	})
	
	t.Run("save updates workspace list", func(t *testing.T) {
		provider := NewAuthProvider()
		
		token := &auth.Token{Value: "test"}
		_ = provider.SaveToken("workspace1", token)
		_ = provider.SaveToken("workspace2", token)
		
		workspaces, _ := provider.ListWorkspaces()
		assert.Contains(t, workspaces, "workspace1")
		assert.Contains(t, workspaces, "workspace2")
	})
}

func TestAuthProviderGetToken(t *testing.T) {
	provider := NewAuthProvider()
	
	t.Run("get existing token", func(t *testing.T) {
		token := &auth.Token{
			Value:     "test-token",
			Workspace: "prod",
			Email:     "user@example.com",
		}
		
		_ = provider.SaveToken("prod", token)
		
		retrieved, err := provider.GetToken("prod")
		require.NoError(t, err)
		assert.Equal(t, token.Value, retrieved.Value)
		assert.Contains(t, provider.GetCalls(), "GetToken(prod)")
	})
	
	t.Run("get non-existent token", func(t *testing.T) {
		_, err := provider.GetToken("nonexistent")
		assert.ErrorIs(t, err, cuerrors.ErrNotAuthenticated)
	})
	
	t.Run("get with error", func(t *testing.T) {
		provider.SetGetError(errors.New("get failed"))
		
		_, err := provider.GetToken("workspace")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "get failed")
	})
	
	t.Run("get with workspace-specific error", func(t *testing.T) {
		provider := NewAuthProvider()
		provider.SetError("prod", cuerrors.ErrTokenExpired)
		
		_, err := provider.GetToken("prod")
		assert.ErrorIs(t, err, cuerrors.ErrTokenExpired)
	})
}

func TestAuthProviderDeleteToken(t *testing.T) {
	provider := NewAuthProvider()
	
	t.Run("delete existing token", func(t *testing.T) {
		token := &auth.Token{Value: "test"}
		_ = provider.SaveToken("workspace", token)
		
		err := provider.DeleteToken("workspace")
		require.NoError(t, err)
		
		assert.False(t, provider.IsAuthenticated("workspace"))
		assert.Contains(t, provider.GetCalls(), "DeleteToken(workspace)")
	})
	
	t.Run("delete with error", func(t *testing.T) {
		provider.SetDeleteError(errors.New("delete failed"))
		
		err := provider.DeleteToken("workspace")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "delete failed")
	})
}

func TestAuthProviderListWorkspaces(t *testing.T) {
	provider := NewAuthProvider()
	
	t.Run("list with tokens", func(t *testing.T) {
		token := &auth.Token{Value: "test"}
		_ = provider.SaveToken("workspace1", token)
		_ = provider.SaveToken("workspace2", token)
		
		workspaces, err := provider.ListWorkspaces()
		require.NoError(t, err)
		assert.Len(t, workspaces, 2)
		assert.Contains(t, workspaces, "workspace1")
		assert.Contains(t, workspaces, "workspace2")
	})
	
	t.Run("list with error", func(t *testing.T) {
		provider.SetListError(errors.New("list failed"))
		
		_, err := provider.ListWorkspaces()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "list failed")
	})
}

func TestAuthProviderGetCurrentToken(t *testing.T) {
	provider := NewAuthProvider()
	
	t.Run("get current workspace token", func(t *testing.T) {
		token := &auth.Token{Value: "current-token"}
		_ = provider.SaveToken(auth.DefaultWorkspace, token)
		
		current, err := provider.GetCurrentToken()
		require.NoError(t, err)
		assert.Equal(t, token.Value, current.Value)
	})
	
	t.Run("change current workspace", func(t *testing.T) {
		provider.SetCurrentWorkspace("production")
		
		token := &auth.Token{Value: "prod-token"}
		_ = provider.SaveToken("production", token)
		
		current, err := provider.GetCurrentToken()
		require.NoError(t, err)
		assert.Equal(t, token.Value, current.Value)
	})
}

func TestAuthProviderTokenExpiry(t *testing.T) {
	provider := NewAuthProvider()
	
	t.Run("token with expiry", func(t *testing.T) {
		// SetToken method supports expiry
		expiry := time.Now().Add(-1 * time.Hour)
		provider.SetToken("workspace", "expiring-token", expiry)
		
		// Should return expired error
		_, err := provider.GetToken("workspace")
		assert.ErrorIs(t, err, cuerrors.ErrTokenExpired)
	})
	
	t.Run("token not expired", func(t *testing.T) {
		expiry := time.Now().Add(1 * time.Hour)
		provider.SetToken("workspace2", "valid-token", expiry)
		
		token, err := provider.GetToken("workspace2")
		assert.NoError(t, err)
		assert.Equal(t, "valid-token", token.Value)
	})
}

func TestAuthProviderConcurrency(t *testing.T) {
	provider := NewAuthProvider()
	
	t.Run("concurrent operations", func(t *testing.T) {
		var wg sync.WaitGroup
		
		// Concurrent saves
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				token := &auth.Token{Value: fmt.Sprintf("token-%d", i)}
				workspace := fmt.Sprintf("workspace-%d", i)
				_ = provider.SaveToken(workspace, token)
			}(i)
		}
		
		// Concurrent reads
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				workspace := fmt.Sprintf("workspace-%d", i)
				_ = provider.IsAuthenticated(workspace)
			}(i)
		}
		
		wg.Wait()
		
		// Verify all operations completed
		calls := provider.GetCalls()
		assert.GreaterOrEqual(t, len(calls), 20)
	})
}

func TestAuthProviderReset(t *testing.T) {
	provider := NewAuthProvider()
	
	// Add some data
	token := &auth.Token{Value: "test"}
	_ = provider.SaveToken("workspace", token)
	provider.SetSaveError(errors.New("error"))
	
	// Reset
	provider.Reset()
	
	// Verify everything is cleared
	assert.False(t, provider.IsAuthenticated("workspace"))
	// Check calls after IsAuthenticated adds its entry
	calls := provider.GetCalls()
	// Should only have the IsAuthenticated call from the check above
	assert.Equal(t, 1, len(calls))
	assert.Contains(t, calls[0], "IsAuthenticated")
}

func TestAuthProviderHelpers(t *testing.T) {
	provider := NewAuthProvider()
	
	t.Run("SetRefreshBehavior", func(t *testing.T) {
		called := false
		provider.SetRefreshBehavior(func(workspace string) (*auth.Token, error) {
			called = true
			assert.Equal(t, "test", workspace)
			return &auth.Token{Value: "refreshed"}, nil
		})
		
		// Set an expired token
		provider.SetToken("test", "old-token", time.Now().Add(-1*time.Hour))
		
		// Getting the token should trigger refresh
		token, err := provider.GetToken("test")
		require.NoError(t, err)
		assert.Equal(t, "refreshed", token.Value)
		assert.True(t, called)
	})
}

func TestKeyringMock(t *testing.T) {
	t.Run("basic operations", func(t *testing.T) {
		k := NewKeyringMock()
		assert.NotNil(t, k)
		
		// Test Set
		err := k.Set("service", "account", "secret")
		require.NoError(t, err)
		
		// Test Get
		secret, err := k.Get("service", "account")
		require.NoError(t, err)
		assert.Equal(t, "secret", secret)
		
		// Test Get non-existent
		_, err = k.Get("service", "nonexistent")
		assert.Error(t, err)
		
		// Test Delete
		err = k.Delete("service", "account")
		require.NoError(t, err)
		
		// Verify deleted
		_, err = k.Get("service", "account")
		assert.Error(t, err)
	})
	
	t.Run("error simulation", func(t *testing.T) {
		k := NewKeyringMock()
		k.SetError(errors.New("keyring error"))
		
		// All operations should fail
		err := k.Set("service", "account", "secret")
		assert.Error(t, err)
		
		_, err = k.Get("service", "account")
		assert.Error(t, err)
		
		err = k.Delete("service", "account")
		assert.Error(t, err)
	})
	
	t.Run("StoreJSON", func(t *testing.T) {
		k := NewKeyringMock()
		
		token := &auth.Token{
			Value:     "test-token",
			Workspace: "prod",
			Email:     "user@example.com",
		}
		
		err := k.StoreJSON("service", "account", token)
		require.NoError(t, err)
		
		// Retrieve and verify JSON
		jsonStr, err := k.Get("service", "account")
		require.NoError(t, err)
		
		var retrieved auth.Token
		err = json.Unmarshal([]byte(jsonStr), &retrieved)
		require.NoError(t, err)
		assert.Equal(t, token.Value, retrieved.Value)
	})
	
	t.Run("Reset", func(t *testing.T) {
		k := NewKeyringMock()
		k.Set("service", "account", "secret")
		k.SetError(errors.New("error"))
		
		k.Reset()
		
		// Error should be cleared
		err := k.Set("service2", "account2", "secret2")
		assert.NoError(t, err)
		
		// Original data should be cleared
		_, err = k.Get("service", "account")
		assert.Error(t, err)
	})
}

func TestScenarios(t *testing.T) {
	provider := NewAuthProvider()
	scenarios := NewScenarios(provider)
	
	t.Run("authenticated scenario", func(t *testing.T) {
		auth := scenarios.Authenticated()
		assert.True(t, auth.IsAuthenticated("default"))
		
		token, err := auth.GetToken("default")
		require.NoError(t, err)
		assert.Equal(t, ValidToken, token.Value)
	})
	
	t.Run("not authenticated scenario", func(t *testing.T) {
		auth := scenarios.NotAuthenticated()
		assert.False(t, auth.IsAuthenticated("default"))
		
		_, err := auth.GetToken("default")
		assert.ErrorIs(t, err, cuerrors.ErrNotAuthenticated)
	})
	
	t.Run("expired token scenario", func(t *testing.T) {
		auth := scenarios.ExpiredToken()
		
		_, err := auth.GetToken("default")
		assert.ErrorIs(t, err, cuerrors.ErrTokenExpired)
	})
	
	t.Run("multiple workspaces scenario", func(t *testing.T) {
		auth := scenarios.MultipleWorkspaces()
		
		// Check all workspaces are authenticated
		assert.True(t, auth.IsAuthenticated("default"))
		assert.True(t, auth.IsAuthenticated("production"))
		assert.True(t, auth.IsAuthenticated("staging"))
		
		// Check tokens have correct values
		token, _ := auth.GetToken("production")
		assert.Equal(t, ValidToken, token.Value)
		assert.Equal(t, AdminEmail, token.Email)
		
		token, _ = auth.GetToken("staging")
		assert.Equal(t, ValidToken, token.Value)
		assert.Equal(t, TestEmail, token.Email)
		
		// List workspaces
		workspaces, err := auth.ListWorkspaces()
		require.NoError(t, err)
		assert.Len(t, workspaces, 3)
	})
	
	t.Run("network error scenario", func(t *testing.T) {
		auth := scenarios.NetworkError()
		
		_, err := auth.GetToken("default")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "network error")
	})
	
	t.Run("partial error scenario", func(t *testing.T) {
		auth := scenarios.PartialError()
		
		// Default workspace works
		token, err := auth.GetToken("default")
		require.NoError(t, err)
		assert.Equal(t, ValidToken, token.Value)
		
		// Production fails
		_, err = auth.GetToken("production")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "production access denied")
	})
	
	t.Run("authenticated with email scenario", func(t *testing.T) {
		auth := scenarios.AuthenticatedWithEmail()
		
		token, err := auth.GetToken("default")
		require.NoError(t, err)
		assert.Equal(t, ValidToken, token.Value)
		assert.Equal(t, TestEmail, token.Email)
	})
	
	t.Run("expired with refresh scenario", func(t *testing.T) {
		auth := scenarios.ExpiredWithRefresh()
		
		// First get should trigger refresh
		token, err := auth.GetToken("default")
		require.NoError(t, err)
		assert.Equal(t, RefreshToken, token.Value)
		assert.Equal(t, TestEmail, token.Email)
	})
	
	t.Run("keyring error scenario", func(t *testing.T) {
		provider := scenarios.KeyringError()
		
		// Save should fail
		testToken := &auth.Token{Value: "test"}
		err := provider.SaveToken("test", testToken)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "keyring error")
		
		// Get should fail
		_, err = provider.GetToken("test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "keyring error")
	})
	
	t.Run("invalid token scenario", func(t *testing.T) {
		auth := scenarios.InvalidToken()
		
		_, err := auth.GetToken("default")
		assert.ErrorIs(t, err, cuerrors.ErrInvalidToken)
	})
	
	t.Run("legacy format scenario", func(t *testing.T) {
		auth := scenarios.LegacyFormat()
		
		token, err := auth.GetToken("default")
		require.NoError(t, err)
		assert.Equal(t, LegacyToken, token.Value)
	})
}
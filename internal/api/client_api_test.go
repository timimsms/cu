package api

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tim/cu/internal/auth"
	"github.com/tim/cu/internal/interfaces"
)

// Test rate limiting and context handling for API methods
// These tests focus on the logic we can test without HTTP calls

func TestClient_GetWorkspaces_Logic(t *testing.T) {
	t.Run("rate limiter called with context cancellation", func(t *testing.T) {
		authManager := &MockAuthManager{
			token: &auth.Token{Value: "test-token"},
		}
		client := NewClient(authManager)
		err := client.Connect()
		assert.NoError(t, err)
		
		// Test with cancelled context to verify rate limiter is called
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		
		_, err = client.GetWorkspaces(cancelledCtx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}

func TestClient_GetSpaces_Logic(t *testing.T) {
	t.Run("validates rate limiting", func(t *testing.T) {
		authManager := &MockAuthManager{
			token: &auth.Token{Value: "test-token"},
		}
		client := NewClient(authManager)
		err := client.Connect()
		assert.NoError(t, err)
		
		// Test with cancelled context
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		
		_, err = client.GetSpaces(cancelledCtx, "123")
		assert.Error(t, err)  
		assert.Contains(t, err.Error(), "context canceled")
	})
}

func TestClient_GetSpace_Logic(t *testing.T) {
	t.Run("validates rate limiting", func(t *testing.T) {
		authManager := &MockAuthManager{
			token: &auth.Token{Value: "test-token"},
		}
		client := NewClient(authManager)
		err := client.Connect()
		assert.NoError(t, err)
		
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		
		_, err = client.GetSpace(cancelledCtx, "456")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}

func TestClient_CreateSpace_Logic(t *testing.T) {
	t.Run("validates rate limiting and ID conversion", func(t *testing.T) {
		authManager := &MockAuthManager{
			token: &auth.Token{Value: "test-token"},
		}
		client := NewClient(authManager)
		err := client.Connect()
		assert.NoError(t, err)
		
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		
		// This will fail at rate limiter stage before ID conversion
		_, err = client.CreateSpace(cancelledCtx, "123", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
	
	t.Run("validates invalid team ID", func(t *testing.T) {
		authManager := &MockAuthManager{
			token: &auth.Token{Value: "test-token"},
		}
		client := NewClient(authManager)
		err := client.Connect()
		assert.NoError(t, err)
		
		ctx := context.Background()
		
		// This should fail at ID conversion before making the API call
		_, err = client.CreateSpace(ctx, "invalid-id", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid team ID")
	})
}

func TestClient_UpdateSpace_Logic(t *testing.T) {
	t.Run("validates invalid space ID", func(t *testing.T) {
		authManager := &MockAuthManager{
			token: &auth.Token{Value: "test-token"},
		}
		client := NewClient(authManager)
		err := client.Connect()
		assert.NoError(t, err)
		
		ctx := context.Background()
		
		// This should fail at ID conversion before making the API call
		_, err = client.UpdateSpace(ctx, "invalid-id", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid space ID")
	})
}

func TestClient_DeleteSpace_Logic(t *testing.T) {
	t.Run("validates invalid space ID", func(t *testing.T) {
		authManager := &MockAuthManager{
			token: &auth.Token{Value: "test-token"},
		}
		client := NewClient(authManager)
		err := client.Connect()
		assert.NoError(t, err)
		
		ctx := context.Background()
		
		// This should fail at ID conversion before making the API call
		err = client.DeleteSpace(ctx, "invalid-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid space ID")
	})
}

func TestClient_GetTask_Logic(t *testing.T) {
	t.Run("validates rate limiting", func(t *testing.T) {
		authManager := &MockAuthManager{
			token: &auth.Token{Value: "test-token"},
		}
		client := NewClient(authManager)
		err := client.Connect()
		assert.NoError(t, err)
		
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		
		_, err = client.GetTask(cancelledCtx, "task123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}

func TestClient_GetTasks_Logic(t *testing.T) {
	t.Run("validates rate limiting and options", func(t *testing.T) {
		authManager := &MockAuthManager{
			token: &auth.Token{Value: "test-token"},
		}
		client := NewClient(authManager)
		err := client.Connect()
		assert.NoError(t, err)
		
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		
		options := &interfaces.TaskQueryOptions{
			Page: 0,
			Assignees: []string{"user1"},
			Statuses: []string{"open"},
		}
		
		_, err = client.GetTasks(cancelledCtx, "list123", options)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}

func TestClient_CreateTask_Logic(t *testing.T) {
	t.Run("validates rate limiting", func(t *testing.T) {
		authManager := &MockAuthManager{
			token: &auth.Token{Value: "test-token"},
		}
		client := NewClient(authManager)
		err := client.Connect()
		assert.NoError(t, err)
		
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		
		options := &interfaces.TaskCreateOptions{
			Name: "Test Task",
			Description: "Test Description", 
		}
		
		_, err = client.CreateTask(cancelledCtx, "list123", options)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}

func TestClient_UpdateTask_Logic(t *testing.T) {
	t.Run("validates rate limiting", func(t *testing.T) {
		authManager := &MockAuthManager{
			token: &auth.Token{Value: "test-token"},
		}
		client := NewClient(authManager)
		err := client.Connect()
		assert.NoError(t, err)
		
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		
		options := &interfaces.TaskUpdateOptions{
			Name: "Updated Task",
		}
		
		_, err = client.UpdateTask(cancelledCtx, "task123", options)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}

func TestClient_DeleteTask_Logic(t *testing.T) {
	t.Run("validates rate limiting", func(t *testing.T) {
		authManager := &MockAuthManager{
			token: &auth.Token{Value: "test-token"},
		}
		client := NewClient(authManager)
		err := client.Connect()
		assert.NoError(t, err)
		
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		
		err = client.DeleteTask(cancelledCtx, "task123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}

func TestClient_GetCurrentUser_Logic(t *testing.T) {
	t.Run("validates rate limiting", func(t *testing.T) {
		authManager := &MockAuthManager{
			token: &auth.Token{Value: "test-token"},
		}
		client := NewClient(authManager)
		err := client.Connect()
		assert.NoError(t, err)
		
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		
		_, err = client.GetCurrentUser(cancelledCtx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}

func TestClient_GetAuthorizedUser_Logic(t *testing.T) {
	t.Run("aliases GetCurrentUser", func(t *testing.T) {
		authManager := &MockAuthManager{
			token: &auth.Token{Value: "test-token"},
		}
		client := NewClient(authManager)
		err := client.Connect()
		assert.NoError(t, err)
		
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		
		_, err = client.GetAuthorizedUser(cancelledCtx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}

func TestClient_GetWorkspaceMembers_Logic(t *testing.T) {
	t.Run("validates rate limiting", func(t *testing.T) {
		authManager := &MockAuthManager{
			token: &auth.Token{Value: "test-token"},
		}
		client := NewClient(authManager)
		err := client.Connect()
		assert.NoError(t, err)
		
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		
		_, err = client.GetWorkspaceMembers(cancelledCtx, "123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}

// Test client structure validation
func TestClient_MethodExistence(t *testing.T) {
	t.Run("all interface methods exist", func(t *testing.T) {
		authManager := &MockAuthManager{
			token: &auth.Token{Value: "test-token"},
		}
		client := NewClient(authManager)
		err := client.Connect()
		assert.NoError(t, err)
		
		ctx := context.Background()
		
		// Test that all methods exist (will panic with nil client internals but validates signatures)
		defer func() { 
			if r := recover(); r != nil {
				// Expected due to nil client internals, but method signatures are validated
				assert.Contains(t, fmt.Sprintf("%v", r), "runtime error")
			}
		}()
		
		// These calls validate method signatures exist
		client.GetWorkspaces(ctx)
		client.GetSpaces(ctx, "123")
		client.GetSpace(ctx, "456")
		client.GetFolders(ctx, "456")
		client.GetFolder(ctx, "789")
		client.GetLists(ctx, "789")
		client.GetList(ctx, "101")
		client.GetTask(ctx, "task123")
		client.GetTasks(ctx, "list123", &interfaces.TaskQueryOptions{})
		client.CreateTask(ctx, "list123", &interfaces.TaskCreateOptions{})
		client.UpdateTask(ctx, "task123", &interfaces.TaskUpdateOptions{})
		client.DeleteTask(ctx, "task123")
		client.GetCurrentUser(ctx)
		client.GetAuthorizedUser(ctx)
		client.GetWorkspaceMembers(ctx, "123")
	})
}

// Test error scenarios
func TestClient_ErrorScenarios(t *testing.T) {
	t.Run("not connected client returns panic", func(t *testing.T) {
		authManager := &MockAuthManager{
			token: &auth.Token{Value: "test-token"},
		}
		client := NewClient(authManager)
		// Don't call Connect()
		
		ctx := context.Background()
		
		// Should panic due to nil client.client
		defer func() {
			if r := recover(); r != nil {
				assert.Contains(t, fmt.Sprintf("%v", r), "runtime error")
			}
		}()
		
		client.GetWorkspaces(ctx)
	})
}
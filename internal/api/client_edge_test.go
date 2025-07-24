package api

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tim/cu/internal/auth"
	"github.com/tim/cu/internal/interfaces"
)

// TestClient_HandleErrorMethod tests the error handling method
func TestClient_HandleErrorMethod(t *testing.T) {
	authManager := &MockAuthManager{
		token: &auth.Token{Value: "test-token"},
	}
	client := NewClient(authManager)
	
	t.Run("nil error returns nil", func(t *testing.T) {
		err := client.handleError(nil)
		assert.NoError(t, err)
	})
	
	t.Run("non-nil error returns same error", func(t *testing.T) {
		testErr := assert.AnError
		err := client.handleError(testErr)
		assert.Equal(t, testErr, err)
	})
}

// TestClient_ErrorHandlingInMethods tests error handling paths in various methods
func TestClient_ErrorHandlingInMethods(t *testing.T) {
	authManager := &MockAuthManager{
		token: &auth.Token{Value: "test-token"},
	}
	client := NewClient(authManager)
	err := client.Connect()
	assert.NoError(t, err)
	
	// Create exhausted rate limiter for context cancellation tests
	client.rateLimiter = NewRateLimiter(1, 24*time.Hour)
	_ = client.rateLimiter.Wait(context.Background())
	
	// Create a context that's already done to test rate limiter error paths
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	
	t.Run("GetWorkspaces with cancelled context", func(t *testing.T) {
		_, err := client.GetWorkspaces(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
	
	t.Run("GetSpaces with cancelled context", func(t *testing.T) {
		_, err := client.GetSpaces(ctx, "123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
	
	t.Run("GetSpace with cancelled context", func(t *testing.T) {
		_, err := client.GetSpace(ctx, "456")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
	
	t.Run("GetFolders with cancelled context", func(t *testing.T) {
		_, err := client.GetFolders(ctx, "456")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
	
	t.Run("GetFolder with cancelled context", func(t *testing.T) {
		_, err := client.GetFolder(ctx, "789")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
	
	t.Run("GetLists with cancelled context", func(t *testing.T) {
		_, err := client.GetLists(ctx, "789")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
	
	t.Run("GetList with cancelled context", func(t *testing.T) {
		_, err := client.GetList(ctx, "101")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
	
	t.Run("GetFolderlessLists with cancelled context", func(t *testing.T) {
		_, err := client.GetFolderlessLists(ctx, "456")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}

// TestClient_TaskOperations tests task-related operations
func TestClient_TaskOperations(t *testing.T) {
	authManager := &MockAuthManager{
		token: &auth.Token{Value: "test-token"},
	}
	client := NewClient(authManager)
	err := client.Connect()
	assert.NoError(t, err)
	
	// Create exhausted rate limiter for context cancellation tests
	client.rateLimiter = NewRateLimiter(1, 24*time.Hour)
	_ = client.rateLimiter.Wait(context.Background())
	
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	
	// Test CreateTask with cancelled context
	t.Run("CreateTask with cancelled context", func(t *testing.T) {
		taskOpts := &interfaces.TaskCreateOptions{
			Name: "Test Task",
		}
		_, err := client.CreateTask(ctx, "list123", taskOpts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
	
	// Test GetTasks with cancelled context
	t.Run("GetTasks with cancelled context", func(t *testing.T) {
		options := &interfaces.TaskQueryOptions{
			Page: 1,
		}
		_, err := client.GetTasks(ctx, "list123", options)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}
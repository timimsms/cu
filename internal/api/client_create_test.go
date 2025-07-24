package api

import (
	"context"
	"testing"
	"time"

	"github.com/raksul/go-clickup/clickup"
	"github.com/stretchr/testify/assert"
	"github.com/tim/cu/internal/auth"
)

// TestClient_CreateMethods tests all create methods with various scenarios
func TestClient_CreateMethods(t *testing.T) {
	authManager := &MockAuthManager{
		token: &auth.Token{Value: "test-token"},
	}
	client := NewClient(authManager)
	err := client.Connect()
	assert.NoError(t, err)
	
	// Create exhausted rate limiter for context cancellation tests
	client.rateLimiter = NewRateLimiter(1, 24*time.Hour)
	_ = client.rateLimiter.Wait(context.Background())

	t.Run("CreateSpace with cancelled context", func(t *testing.T) {
		// Test with context cancellation using exhausted rate limiter
		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel()
		
		spaceReq := &clickup.SpaceRequest{
			Name: "Test Space",
		}
		_, err := client.CreateSpace(cancelCtx, "123456", spaceReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})

	t.Run("CreateFolder with cancelled context", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel()
		folderReq := &clickup.FolderRequest{
			Name: "Test Folder",
		}
		
		_, err := client.CreateFolder(cancelCtx, "123456", folderReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})

	t.Run("CreateList with cancelled context", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel()
		listReq := &clickup.ListRequest{
			Name: "Test List",
		}
		
		_, err := client.CreateList(cancelCtx, "123456", listReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})

	t.Run("CreateFolderlessList with cancelled context", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel()
		listReq := &clickup.ListRequest{
			Name: "Test Folderless List",
		}
		
		_, err := client.CreateFolderlessList(cancelCtx, "123456", listReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}

// TestClient_UpdateMethods tests update methods for better coverage
func TestClient_UpdateMethods(t *testing.T) {
	authManager := &MockAuthManager{
		token: &auth.Token{Value: "test-token"},
	}
	client := NewClient(authManager)
	err := client.Connect()
	assert.NoError(t, err)
	
	// Create exhausted rate limiter for context cancellation tests
	client.rateLimiter = NewRateLimiter(1, 24*time.Hour)
	_ = client.rateLimiter.Wait(context.Background())

	t.Run("UpdateSpace with cancelled context", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel()
		spaceReq := &clickup.SpaceRequest{
			Name: "Updated Space",
		}
		
		_, err := client.UpdateSpace(cancelCtx, "123456", spaceReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})

	t.Run("UpdateFolder with cancelled context", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel()
		folderReq := &clickup.FolderRequest{
			Name: "Updated Folder",
		}
		
		_, err := client.UpdateFolder(cancelCtx, "123456", folderReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}

// TestClient_DeleteMethods tests delete methods for better coverage
func TestClient_DeleteMethods(t *testing.T) {
	authManager := &MockAuthManager{
		token: &auth.Token{Value: "test-token"},
	}
	client := NewClient(authManager)
	err := client.Connect()
	assert.NoError(t, err)
	
	// Create exhausted rate limiter for context cancellation tests
	client.rateLimiter = NewRateLimiter(1, 24*time.Hour)
	_ = client.rateLimiter.Wait(context.Background())

	t.Run("DeleteSpace with cancelled context", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel()
		
		err := client.DeleteSpace(cancelCtx, "123456")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})

	t.Run("DeleteFolder with cancelled context", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(context.Background())
		cancel()
		
		err := client.DeleteFolder(cancelCtx, "123456")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}
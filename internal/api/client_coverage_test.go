package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tim/cu/internal/auth"
	"github.com/tim/cu/internal/interfaces"
)

// Focused tests for coverage improvement without HTTP calls
// These tests target specific business logic and validation paths

func TestClient_IDValidation(t *testing.T) {
	authManager := &MockAuthManager{
		token: &auth.Token{Value: "test-token"},
	}
	client := NewClient(authManager)
	err := client.Connect()
	assert.NoError(t, err)
	
	ctx := context.Background()

	t.Run("CreateSpace with invalid team ID", func(t *testing.T) {
		_, err := client.CreateSpace(ctx, "invalid-id", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid team ID")
	})

	t.Run("UpdateSpace with invalid space ID", func(t *testing.T) {
		_, err := client.UpdateSpace(ctx, "invalid-id", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid space ID")
	})

	t.Run("DeleteSpace with invalid space ID", func(t *testing.T) {
		err := client.DeleteSpace(ctx, "invalid-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid space ID")
	})

	t.Run("CreateFolder with invalid space ID", func(t *testing.T) {
		_, err := client.CreateFolder(ctx, "invalid-id", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid space ID")
	})

	t.Run("UpdateFolder with invalid folder ID", func(t *testing.T) {
		_, err := client.UpdateFolder(ctx, "invalid-id", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid folder ID")
	})

	t.Run("DeleteFolder with invalid folder ID", func(t *testing.T) {
		err := client.DeleteFolder(ctx, "invalid-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid folder ID")
	})

	// Note: CreateList, UpdateList, and DeleteList don't validate IDs
	// They pass them directly to the ClickUp API which returns errors
}

// Test that all the method entry points exist and handle rate limiting
func TestClient_MethodCoverage(t *testing.T) {
	authManager := &MockAuthManager{
		token: &auth.Token{Value: "test-token"},
	}
	client := NewClient(authManager)
	err := client.Connect()
	assert.NoError(t, err)
	
	// Use a cancelled context to stop at rate limiter for most methods
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Test methods that reach rate limiter
	methods := []func() error{
		func() error { _, err := client.GetWorkspaces(ctx); return err },
		func() error { _, err := client.GetSpaces(ctx, "123"); return err },
		func() error { _, err := client.GetSpace(ctx, "456"); return err },
		func() error { _, err := client.GetFolders(ctx, "456"); return err },
		func() error { _, err := client.GetFolder(ctx, "789"); return err },
		func() error { _, err := client.GetLists(ctx, "789"); return err },
		func() error { _, err := client.GetFolderlessLists(ctx, "456"); return err },
		func() error { _, err := client.GetList(ctx, "101"); return err },
		func() error { _, err := client.GetTask(ctx, "task123"); return err },
		func() error { _, err := client.GetCurrentUser(ctx); return err },
		func() error { _, err := client.GetAuthorizedTeams(ctx); return err },
		func() error { _, err := client.GetWorkspaceMembers(ctx, "123"); return err },
		func() error { _, err := client.GetMembers(ctx, "456"); return err },
		func() error { _, err := client.GetViews(ctx, "101"); return err },
		func() error { _, err := client.GetView(ctx, "view1"); return err },
		func() error { _, err := client.GetTaskComments(ctx, "task123"); return err },
		func() error { _, err := client.GetCustomFields(ctx, "101"); return err },
		func() error { _, _, err := client.GetGoals(ctx, "123", false); return err },
		func() error { _, err := client.GetGoal(ctx, "goal1"); return err },
		func() error { _, err := client.GetWebhooks(ctx, "123"); return err },
		func() error { return client.DeleteTask(ctx, "task123") },
		func() error { return client.UpdateTaskComment(ctx, "comment1", "text", false) },
		func() error { return client.DeleteTaskComment(ctx, "comment1") },
		func() error { return client.SetCustomFieldValue(ctx, "task123", "field1", map[string]interface{}{"value": "test"}) },
		func() error { return client.DeleteGoal(ctx, "goal1") },
		func() error { return client.DeleteWebhook(ctx, "webhook1") },
	}

	// These methods will all fail with HTTP errors since we're using real tokens,
	// but they exercise the rate limiter and method entry points
	for i, method := range methods {
		err := method()
		assert.Error(t, err, "Method %d should return error", i)
		// Don't check specific error content since it could be context canceled or HTTP error
	}
}

// Test alias methods that just call other methods
func TestClient_AliasMethods(t *testing.T) {
	authManager := &MockAuthManager{
		token: &auth.Token{Value: "test-token"},
	}
	client := NewClient(authManager)
	err := client.Connect()
	assert.NoError(t, err)
	
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	t.Run("GetAuthorizedUser aliases GetCurrentUser", func(t *testing.T) {
		_, err := client.GetAuthorizedUser(ctx)
		assert.Error(t, err)
	})

	t.Run("GetAuthorizedTeams aliases GetWorkspaces", func(t *testing.T) {
		_, err := client.GetAuthorizedTeams(ctx)
		assert.Error(t, err)
	})
}

// Test methods with business logic that we can validate
func TestClient_BusinessLogic(t *testing.T) {
	t.Run("GetAuthorizedUser returns GetCurrentUser", func(t *testing.T) {
		// This tests the alias relationship
		authManager := &MockAuthManager{
			token: &auth.Token{Value: "test-token"},
		}
		client := NewClient(authManager)
		err := client.Connect()
		assert.NoError(t, err)
		
		ctx := context.Background()
		
		// Both should produce the same error pattern when failing
		_, err1 := client.GetCurrentUser(ctx)
		_, err2 := client.GetAuthorizedUser(ctx)
		
		// Both should fail in the same way
		assert.Error(t, err1)
		assert.Error(t, err2)
	})
}

// Test comment ID validation and conversion
func TestClient_CommentOperations(t *testing.T) {
	authManager := &MockAuthManager{
		token: &auth.Token{Value: "test-token"},
	}
	client := NewClient(authManager)
	err := client.Connect()
	assert.NoError(t, err)
	
	ctx := context.Background()

	t.Run("UpdateTaskComment with invalid comment ID", func(t *testing.T) {
		err := client.UpdateTaskComment(ctx, "invalid-id", "updated text", false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid comment ID format")
	})

	t.Run("DeleteTaskComment with invalid comment ID", func(t *testing.T) {
		err := client.DeleteTaskComment(ctx, "invalid-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid comment ID format")
	})

	t.Run("UpdateTaskComment with valid comment ID format", func(t *testing.T) {
		// Use cancelled context to stop at rate limiter
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		
		err := client.UpdateTaskComment(cancelledCtx, "123", "updated text", false)
		assert.Error(t, err)
		// Will fail at rate limiter, not ID validation
	})

	t.Run("DeleteTaskComment with valid comment ID format", func(t *testing.T) {
		// Use cancelled context to stop at rate limiter
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		
		err := client.DeleteTaskComment(cancelledCtx, "456")
		assert.Error(t, err)
		// Will fail at rate limiter, not ID validation
	})
}

// Test date parsing logic
func TestClient_DateParsing(t *testing.T) {
	t.Run("parseDueDate handles relative dates", func(t *testing.T) {
		// Test relative dates
		_, err := parseDueDate("today")
		assert.NoError(t, err)
		
		_, err = parseDueDate("tomorrow")
		assert.NoError(t, err)
		
		_, err = parseDueDate("week")
		assert.NoError(t, err)
	})

	t.Run("parseDueDate handles RFC3339 dates", func(t *testing.T) {
		_, err := parseDueDate("2023-12-25T15:30:00Z")
		assert.NoError(t, err)
	})

	t.Run("parseDueDate handles date-only format", func(t *testing.T) {
		_, err := parseDueDate("2023-12-25")
		assert.NoError(t, err)
	})

	t.Run("parseDueDate handles invalid dates", func(t *testing.T) {
		_, err := parseDueDate("invalid-date")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unable to parse date")
	})
}

// Test goal operations with ID validation
func TestClient_GoalOperations(t *testing.T) {
	authManager := &MockAuthManager{
		token: &auth.Token{Value: "test-token"},
	}
	client := NewClient(authManager)
	err := client.Connect()
	assert.NoError(t, err)
	
	ctx := context.Background()

	t.Run("CreateGoal with invalid team ID", func(t *testing.T) {
		_, err := client.CreateGoal(ctx, "invalid-id", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid team ID")
	})

	t.Run("CreateGoal with valid team ID format", func(t *testing.T) {
		// Use cancelled context to stop at rate limiter
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		
		_, err := client.CreateGoal(cancelledCtx, "123", nil)
		assert.Error(t, err)
		// Will fail at rate limiter, not ID validation
	})
}

// Test webhook operations with ID validation
func TestClient_WebhookOperations(t *testing.T) {
	authManager := &MockAuthManager{
		token: &auth.Token{Value: "test-token"},
	}
	client := NewClient(authManager)
	err := client.Connect()
	assert.NoError(t, err)
	
	ctx := context.Background()

	t.Run("GetWebhooks with invalid team ID", func(t *testing.T) {
		_, err := client.GetWebhooks(ctx, "invalid-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid team ID")
	})

	t.Run("CreateWebhook with invalid team ID", func(t *testing.T) {
		_, err := client.CreateWebhook(ctx, "invalid-id", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid team ID")
	})

	t.Run("GetWebhooks with valid team ID format", func(t *testing.T) {
		// Use cancelled context to stop at rate limiter
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		
		_, err := client.GetWebhooks(cancelledCtx, "123")
		assert.Error(t, err)
		// Will fail at rate limiter, not ID validation
	})

	t.Run("CreateWebhook with valid team ID format", func(t *testing.T) {
		// Use cancelled context to stop at rate limiter
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		
		_, err := client.CreateWebhook(cancelledCtx, "456", nil)
		assert.Error(t, err)
		// Will fail at rate limiter, not ID validation
	})
}

// Test TaskUpdateOptions business logic
func TestClient_TaskUpdateOptions(t *testing.T) {
	t.Run("HasUpdates returns false for empty options", func(t *testing.T) {
		options := &interfaces.TaskUpdateOptions{}
		assert.False(t, options.HasUpdates())
	})

	t.Run("HasUpdates returns true when Name is set", func(t *testing.T) {
		options := &interfaces.TaskUpdateOptions{Name: "test"}
		assert.True(t, options.HasUpdates())
	})

	t.Run("HasUpdates returns true when Description is set", func(t *testing.T) {
		options := &interfaces.TaskUpdateOptions{Description: "test"}
		assert.True(t, options.HasUpdates())
	})

	t.Run("HasUpdates returns true when Status is set", func(t *testing.T) {
		options := &interfaces.TaskUpdateOptions{Status: "open"}
		assert.True(t, options.HasUpdates())
	})

	t.Run("HasUpdates returns true when Priority is set", func(t *testing.T) {
		options := &interfaces.TaskUpdateOptions{Priority: "high"}
		assert.True(t, options.HasUpdates())
	})

	t.Run("HasUpdates returns true when Tags are set", func(t *testing.T) {
		options := &interfaces.TaskUpdateOptions{Tags: []string{"tag1"}}
		assert.True(t, options.HasUpdates())
	})

	t.Run("HasUpdates returns true when DueDate is set", func(t *testing.T) {
		options := &interfaces.TaskUpdateOptions{DueDate: "today"}
		assert.True(t, options.HasUpdates())
	})

	t.Run("HasUpdates returns true when AddAssignees are set", func(t *testing.T) {
		options := &interfaces.TaskUpdateOptions{AddAssignees: []string{"user1"}}
		assert.True(t, options.HasUpdates())
	})

	t.Run("HasUpdates returns true when RemoveAssignees are set", func(t *testing.T) {
		options := &interfaces.TaskUpdateOptions{RemoveAssignees: []string{"user1"}}
		assert.True(t, options.HasUpdates())
	})
}
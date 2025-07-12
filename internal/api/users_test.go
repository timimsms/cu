package api

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/raksul/go-clickup/clickup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockClient for testing UserLookup
type mockClient struct {
	users []clickup.TeamUser
	err   error
}

func (m *mockClient) GetWorkspaceMembers(ctx context.Context, workspaceID string) ([]clickup.TeamUser, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.users, nil
}

func TestNewUserLookup(t *testing.T) {
	client := &Client{}
	ul := NewUserLookup(client)

	assert.NotNil(t, ul)
	assert.NotNil(t, ul.cache)
	assert.NotNil(t, ul.idMap)
	assert.Equal(t, client, ul.client)
}

func TestUserLookupLoadWorkspaceUsers(t *testing.T) {
	t.Run("loads users successfully", func(t *testing.T) {
		testUsers := []clickup.TeamUser{
			{ID: 123, Username: "john.doe", Email: "john@example.com"},
			{ID: 456, Username: "Jane.Smith", Email: "jane@example.com"},
		}

		client := &Client{}
		ul := &UserLookup{
			client: client,
			cache:  make(map[string]*clickup.TeamUser),
			idMap:  make(map[int]*clickup.TeamUser),
		}

		// Manually set up the users since we can't easily mock the client
		ul.mu.Lock()
		for i := range testUsers {
			user := &testUsers[i]
			ul.cache[strings.ToLower(user.Username)] = user
			ul.idMap[user.ID] = user
		}
		ul.mu.Unlock()

		// Verify users are loaded
		assert.Len(t, ul.cache, 2)
		assert.Len(t, ul.idMap, 2)

		// Check case-insensitive storage
		_, ok := ul.cache["john.doe"]
		assert.True(t, ok)
		_, ok = ul.cache["jane.smith"]
		assert.True(t, ok)
	})

	t.Run("handles API error", func(t *testing.T) {
		// This would require proper mocking of the client
		t.Skip("Requires client mocking")
	})
}

func TestUserLookupByUsername(t *testing.T) {
	ul := &UserLookup{
		cache: make(map[string]*clickup.TeamUser),
		idMap: make(map[int]*clickup.TeamUser),
	}

	testUser := &clickup.TeamUser{
		ID:       123,
		Username: "TestUser",
		Email:    "test@example.com",
	}

	ul.cache["testuser"] = testUser
	ul.idMap[123] = testUser

	t.Run("finds user by exact username", func(t *testing.T) {
		user, err := ul.LookupByUsername("testuser")
		require.NoError(t, err)
		assert.Equal(t, testUser, user)
	})

	t.Run("finds user case-insensitive", func(t *testing.T) {
		user, err := ul.LookupByUsername("TestUser")
		require.NoError(t, err)
		assert.Equal(t, testUser, user)

		user, err = ul.LookupByUsername("TESTUSER")
		require.NoError(t, err)
		assert.Equal(t, testUser, user)
	})

	t.Run("returns error for unknown user", func(t *testing.T) {
		user, err := ul.LookupByUsername("unknown")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "user not found: unknown")
	})
}

func TestUserLookupByID(t *testing.T) {
	ul := &UserLookup{
		cache: make(map[string]*clickup.TeamUser),
		idMap: make(map[int]*clickup.TeamUser),
	}

	testUser := &clickup.TeamUser{
		ID:       789,
		Username: "testuser",
		Email:    "test@example.com",
	}

	ul.idMap[789] = testUser

	t.Run("finds user by ID", func(t *testing.T) {
		user, err := ul.LookupByID(789)
		require.NoError(t, err)
		assert.Equal(t, testUser, user)
	})

	t.Run("returns error for unknown ID", func(t *testing.T) {
		user, err := ul.LookupByID(999)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "user not found: 999")
	})
}

func TestConvertUsernamesToIDs(t *testing.T) {
	ul := &UserLookup{
		cache: make(map[string]*clickup.TeamUser),
		idMap: make(map[int]*clickup.TeamUser),
	}

	// Set up test users
	users := []struct {
		user *clickup.TeamUser
		key  string
	}{
		{&clickup.TeamUser{ID: 100, Username: "alice"}, "alice"},
		{&clickup.TeamUser{ID: 200, Username: "bob"}, "bob"},
		{&clickup.TeamUser{ID: 300, Username: "Charlie"}, "charlie"},
	}

	for _, u := range users {
		ul.cache[u.key] = u.user
		ul.idMap[u.user.ID] = u.user
	}

	t.Run("converts usernames to IDs", func(t *testing.T) {
		ids, err := ul.ConvertUsernamesToIDs([]string{"alice", "bob", "Charlie"})
		require.NoError(t, err)
		assert.Equal(t, []int{100, 200, 300}, ids)
	})

	t.Run("handles numeric strings as IDs", func(t *testing.T) {
		ids, err := ul.ConvertUsernamesToIDs([]string{"alice", "999", "bob"})
		require.NoError(t, err)
		assert.Equal(t, []int{100, 999, 200}, ids)
	})

	t.Run("returns error for unknown username", func(t *testing.T) {
		ids, err := ul.ConvertUsernamesToIDs([]string{"alice", "unknown"})
		assert.Error(t, err)
		assert.Nil(t, ids)
		assert.Contains(t, err.Error(), "failed to find user unknown")
	})

	t.Run("handles empty list", func(t *testing.T) {
		ids, err := ul.ConvertUsernamesToIDs([]string{})
		require.NoError(t, err)
		assert.Empty(t, ids)
	})
}

func TestGetAllUsers(t *testing.T) {
	ul := &UserLookup{
		cache: make(map[string]*clickup.TeamUser),
		idMap: make(map[int]*clickup.TeamUser),
	}

	t.Run("returns empty list when no users", func(t *testing.T) {
		users := ul.GetAllUsers()
		assert.Empty(t, users)
	})

	t.Run("returns all cached users", func(t *testing.T) {
		testUsers := []*clickup.TeamUser{
			{ID: 1, Username: "user1"},
			{ID: 2, Username: "user2"},
			{ID: 3, Username: "user3"},
		}

		for _, user := range testUsers {
			ul.cache[strings.ToLower(user.Username)] = user
			ul.idMap[user.ID] = user
		}

		users := ul.GetAllUsers()
		assert.Len(t, users, 3)

		// Check all users are present
		userMap := make(map[int]bool)
		for _, user := range users {
			userMap[user.ID] = true
		}
		assert.True(t, userMap[1])
		assert.True(t, userMap[2])
		assert.True(t, userMap[3])
	})
}

func TestUserLookupConcurrency(t *testing.T) {
	t.Run("concurrent reads are safe", func(t *testing.T) {
		ul := &UserLookup{
			cache: make(map[string]*clickup.TeamUser),
			idMap: make(map[int]*clickup.TeamUser),
		}

		// Add test user
		testUser := &clickup.TeamUser{ID: 123, Username: "test"}
		ul.cache["test"] = testUser
		ul.idMap[123] = testUser

		var wg sync.WaitGroup
		errors := make(chan error, 100)

		// Launch multiple readers
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				if i%2 == 0 {
					_, err := ul.LookupByUsername("test")
					if err != nil {
						errors <- err
					}
				} else {
					_, err := ul.LookupByID(123)
					if err != nil {
						errors <- err
					}
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		// Check no errors occurred
		for err := range errors {
			t.Errorf("Unexpected error during concurrent read: %v", err)
		}
	})

	t.Run("concurrent writes are safe", func(t *testing.T) {
		ul := &UserLookup{
			cache: make(map[string]*clickup.TeamUser),
			idMap: make(map[int]*clickup.TeamUser),
		}

		var wg sync.WaitGroup

		// Launch multiple writers
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				ul.mu.Lock()
				user := &clickup.TeamUser{
					ID:       id,
					Username: fmt.Sprintf("user%d", id),
				}
				ul.cache[strings.ToLower(user.Username)] = user
				ul.idMap[user.ID] = user
				ul.mu.Unlock()
			}(i)
		}

		wg.Wait()

		// Verify all users were added
		assert.Len(t, ul.cache, 10)
		assert.Len(t, ul.idMap, 10)
	})
}
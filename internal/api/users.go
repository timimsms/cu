package api

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/raksul/go-clickup/clickup"
	"github.com/tim/cu/internal/cache"
)

// UserLookup provides user ID lookup functionality
type UserLookup struct {
	client *Client
	mu     sync.RWMutex
	cache  map[string]*clickup.TeamUser // username -> user
	idMap  map[int]*clickup.TeamUser    // id -> user
}

// NewUserLookup creates a new user lookup service
func NewUserLookup(client *Client) *UserLookup {
	return &UserLookup{
		client: client,
		cache:  make(map[string]*clickup.TeamUser),
		idMap:  make(map[int]*clickup.TeamUser),
	}
}

// LoadWorkspaceUsers loads all users from a workspace into cache
func (ul *UserLookup) LoadWorkspaceUsers(ctx context.Context, workspaceID string) error {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("users_%s", workspaceID)
	if cache.UserCache != nil {
		var users []clickup.TeamUser
		if err := cache.UserCache.Get(cacheKey, &users); err == nil {
			// Load from cache
			ul.mu.Lock()
			for i := range users {
				user := &users[i]
				ul.cache[strings.ToLower(user.Username)] = user
				ul.idMap[user.ID] = user
			}
			ul.mu.Unlock()
			return nil
		}
	}

	// Get users from API
	users, err := ul.client.GetWorkspaceMembers(ctx, workspaceID)
	if err != nil {
		return err
	}

	// Cache the result
	if cache.UserCache != nil {
		_ = cache.UserCache.Set(cacheKey, users)
	}

	// Store in memory
	ul.mu.Lock()
	for i := range users {
		user := &users[i]
		ul.cache[strings.ToLower(user.Username)] = user
		ul.idMap[user.ID] = user
	}
	ul.mu.Unlock()

	return nil
}

// LookupByUsername finds a user by username (case-insensitive)
func (ul *UserLookup) LookupByUsername(username string) (*clickup.TeamUser, error) {
	ul.mu.RLock()
	defer ul.mu.RUnlock()

	user, ok := ul.cache[strings.ToLower(username)]
	if !ok {
		return nil, fmt.Errorf("user not found: %s", username)
	}

	return user, nil
}

// LookupByID finds a user by ID
func (ul *UserLookup) LookupByID(userID int) (*clickup.TeamUser, error) {
	ul.mu.RLock()
	defer ul.mu.RUnlock()

	user, ok := ul.idMap[userID]
	if !ok {
		return nil, fmt.Errorf("user not found: %d", userID)
	}

	return user, nil
}

// ConvertUsernamesToIDs converts a list of usernames to user IDs
func (ul *UserLookup) ConvertUsernamesToIDs(usernames []string) ([]int, error) {
	ids := make([]int, 0, len(usernames))

	for _, username := range usernames {
		// Check if it's already an ID
		if id, err := strconv.Atoi(username); err == nil {
			ids = append(ids, id)
			continue
		}

		// Look up by username
		user, err := ul.LookupByUsername(username)
		if err != nil {
			return nil, fmt.Errorf("failed to find user %s: %w", username, err)
		}
		ids = append(ids, user.ID)
	}

	return ids, nil
}

// GetAllUsers returns all cached users
func (ul *UserLookup) GetAllUsers() []*clickup.TeamUser {
	ul.mu.RLock()
	defer ul.mu.RUnlock()

	users := make([]*clickup.TeamUser, 0, len(ul.cache))
	for _, user := range ul.cache {
		users = append(users, user)
	}

	return users
}

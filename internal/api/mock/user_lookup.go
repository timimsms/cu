package mock

import (
	"context"
	"fmt"
	"sync"

	clickup "github.com/raksul/go-clickup"
)

// UserLookup is a mock implementation of the UserLookup service
type UserLookup struct {
	mu sync.RWMutex

	// Data
	workspaceUsers map[string][]*clickup.TeamUser
	usersByUsername map[string]*clickup.TeamUser
	usersByID      map[int]*clickup.TeamUser

	// Errors
	loadErr error

	// Call tracking
	calls []string
}

// NewUserLookup creates a new mock UserLookup
func NewUserLookup() *UserLookup {
	return &UserLookup{
		workspaceUsers:  make(map[string][]*clickup.TeamUser),
		usersByUsername: make(map[string]*clickup.TeamUser),
		usersByID:       make(map[int]*clickup.TeamUser),
		calls:           []string{},
	}
}

// LoadWorkspaceUsers mocks loading workspace users
func (u *UserLookup) LoadWorkspaceUsers(ctx context.Context, workspaceID string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	
	u.calls = append(u.calls, fmt.Sprintf("LoadWorkspaceUsers(%s)", workspaceID))
	
	if u.loadErr != nil {
		return u.loadErr
	}
	
	// Simulate loading by populating lookup maps
	if users, ok := u.workspaceUsers[workspaceID]; ok {
		for _, user := range users {
			u.usersByUsername[user.Username] = user
			u.usersByID[user.ID] = user
		}
	}
	
	return nil
}

// LookupByUsername returns a user by username
func (u *UserLookup) LookupByUsername(username string) (*clickup.TeamUser, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()
	
	u.calls = append(u.calls, fmt.Sprintf("LookupByUsername(%s)", username))
	
	user, ok := u.usersByUsername[username]
	if !ok {
		return nil, fmt.Errorf("user not found: %s", username)
	}
	return user, nil
}

// LookupByID returns a user by ID
func (u *UserLookup) LookupByID(userID int) (*clickup.TeamUser, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()
	
	u.calls = append(u.calls, fmt.Sprintf("LookupByID(%d)", userID))
	
	user, ok := u.usersByID[userID]
	if !ok {
		return nil, fmt.Errorf("user not found: %d", userID)
	}
	return user, nil
}

// ConvertUsernamesToIDs converts usernames to user IDs
func (u *UserLookup) ConvertUsernamesToIDs(usernames []string) ([]int, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()
	
	u.calls = append(u.calls, fmt.Sprintf("ConvertUsernamesToIDs(%v)", usernames))
	
	var ids []int
	for _, username := range usernames {
		user, ok := u.usersByUsername[username]
		if !ok {
			return nil, fmt.Errorf("user not found: %s", username)
		}
		ids = append(ids, user.ID)
	}
	return ids, nil
}

// GetAllUsers returns all loaded users
func (u *UserLookup) GetAllUsers() []*clickup.TeamUser {
	u.mu.RLock()
	defer u.mu.RUnlock()
	
	u.calls = append(u.calls, "GetAllUsers()")
	
	var users []*clickup.TeamUser
	for _, user := range u.usersByID {
		users = append(users, user)
	}
	return users
}

// Helper methods for test setup

// SetWorkspaceUsers sets users for a workspace
func (u *UserLookup) SetWorkspaceUsers(workspaceID string, users []*clickup.TeamUser) {
	u.mu.Lock()
	defer u.mu.Unlock()
	
	u.workspaceUsers[workspaceID] = users
}

// AddUser adds a user to the lookup maps
func (u *UserLookup) AddUser(user *clickup.TeamUser) {
	u.mu.Lock()
	defer u.mu.Unlock()
	
	u.usersByUsername[user.Username] = user
	u.usersByID[user.ID] = user
}

// SetLoadError sets an error for LoadWorkspaceUsers
func (u *UserLookup) SetLoadError(err error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.loadErr = err
}

// GetCalls returns the list of method calls made
func (u *UserLookup) GetCalls() []string {
	u.mu.RLock()
	defer u.mu.RUnlock()
	
	calls := make([]string, len(u.calls))
	copy(calls, u.calls)
	return calls
}

// Reset clears all data and call history
func (u *UserLookup) Reset() {
	u.mu.Lock()
	defer u.mu.Unlock()
	
	u.workspaceUsers = make(map[string][]*clickup.TeamUser)
	u.usersByUsername = make(map[string]*clickup.TeamUser)
	u.usersByID = make(map[int]*clickup.TeamUser)
	u.loadErr = nil
	u.calls = []string{}
}
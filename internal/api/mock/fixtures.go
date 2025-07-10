package mock

import (
	"time"

	clickup "github.com/raksul/go-clickup"
)

// Test IDs
const (
	TestUserID      = 12345
	TestWorkspaceID = "98765"
	TestSpaceID     = "sp_123"
	TestFolderID    = "fl_456"
	TestListID      = "li_789"
	TestTaskID      = "tk_abc"
	TestCommentID   = "cm_xyz"
)

// UserFixtures provides pre-configured users for testing
var UserFixtures = struct {
	CurrentUser *clickup.User
	TeamMember1 *clickup.TeamUser
	TeamMember2 *clickup.TeamUser
}{
	CurrentUser: &clickup.User{
		ID:       TestUserID,
		Username: "testuser",
		Email:    "test@example.com",
		Color:    "#FF5733",
	},
	TeamMember1: &clickup.TeamUser{
		User: clickup.User{
			ID:       12346,
			Username: "alice",
			Email:    "alice@example.com",
			Color:    "#33FF57",
		},
		Role: 3, // Member
	},
	TeamMember2: &clickup.TeamUser{
		User: clickup.User{
			ID:       12347,
			Username: "bob",
			Email:    "bob@example.com",
			Color:    "#5733FF",
		},
		Role: 2, // Admin
	},
}

// WorkspaceFixtures provides pre-configured workspaces
var WorkspaceFixtures = struct {
	DefaultWorkspace *clickup.Team
	SecondWorkspace  *clickup.Team
}{
	DefaultWorkspace: &clickup.Team{
		ID:      TestWorkspaceID,
		Name:    "Test Workspace",
		Color:   "#FF5733",
		Members: []clickup.TeamUser{*UserFixtures.TeamMember1, *UserFixtures.TeamMember2},
	},
	SecondWorkspace: &clickup.Team{
		ID:      "98766",
		Name:    "Secondary Workspace",
		Color:   "#33FF57",
		Members: []clickup.TeamUser{},
	},
}

// HierarchyFixtures provides pre-configured spaces, folders, and lists
var HierarchyFixtures = struct {
	Space1  *clickup.Space
	Space2  *clickup.Space
	Folder1 *clickup.Folder
	List1   *clickup.List
	List2   *clickup.List
}{
	Space1: &clickup.Space{
		ID:       TestSpaceID,
		Name:     "Test Space",
		Private:  false,
		Statuses: []clickup.Status{
			{Status: "to do", Color: "#ff0000", OrderIndex: 0},
			{Status: "in progress", Color: "#ffff00", OrderIndex: 1},
			{Status: "done", Color: "#00ff00", OrderIndex: 2},
		},
	},
	Space2: &clickup.Space{
		ID:       "sp_124",
		Name:     "Private Space",
		Private:  true,
		Statuses: []clickup.Status{
			{Status: "open", Color: "#ff0000", OrderIndex: 0},
			{Status: "closed", Color: "#00ff00", OrderIndex: 1},
		},
	},
	Folder1: &clickup.Folder{
		ID:               TestFolderID,
		Name:             "Test Folder",
		OrderIndex:       0,
		OverrideStatuses: false,
		Hidden:           false,
	},
	List1: &clickup.List{
		ID:               TestListID,
		Name:             "Test List",
		OrderIndex:       0,
		Status:           clickup.Status{Status: "active", Color: "#00ff00"},
		Priority:         clickup.Priority{Priority: "high", Color: "#ff0000"},
		Assignee:         nil,
		TaskCount:        5,
		DueDate:          "",
		DueDateTimestamp: 0,
		StartDate:        "",
		Archived:         false,
	},
	List2: &clickup.List{
		ID:               "li_790",
		Name:             "Archived List",
		OrderIndex:       1,
		Status:           clickup.Status{Status: "inactive", Color: "#999999"},
		Priority:         clickup.Priority{Priority: "low", Color: "#0000ff"},
		TaskCount:        0,
		Archived:         true,
	},
}

// TaskFixtures provides pre-configured tasks
var TaskFixtures = struct {
	SimpleTask    *clickup.Task
	CompleteTask  *clickup.Task
	OverdueTask   *clickup.Task
	AssignedTask  *clickup.Task
}{
	SimpleTask: &clickup.Task{
		ID:          TestTaskID,
		Name:        "Test Task",
		Description: "This is a test task",
		Status:      clickup.Status{Status: "to do", Color: "#ff0000"},
		OrderIndex:  "1",
		DateCreated: "1640995200000", // 2022-01-01
		DateUpdated: "1640995200000",
		Creator:     UserFixtures.CurrentUser,
		Assignees:   []clickup.User{},
		Checklists:  []clickup.Checklist{},
		Tags:        []clickup.Tag{},
		Parent:      "",
		Priority:    &clickup.TaskPriority{Priority: "normal"},
		URL:         "https://app.clickup.com/t/tk_abc",
	},
	CompleteTask: &clickup.Task{
		ID:          "tk_abd",
		Name:        "Completed Task",
		Description: "This task is done",
		Status:      clickup.Status{Status: "done", Color: "#00ff00"},
		OrderIndex:  "2",
		DateCreated: "1640995200000",
		DateUpdated: "1641081600000", // 2022-01-02
		DateClosed:  "1641081600000",
		Creator:     UserFixtures.CurrentUser,
		Assignees:   []clickup.User{},
		TimeSpent:   3600000, // 1 hour in milliseconds
	},
	OverdueTask: &clickup.Task{
		ID:          "tk_abe",
		Name:        "Overdue Task",
		Description: "This task is overdue",
		Status:      clickup.Status{Status: "in progress", Color: "#ffff00"},
		OrderIndex:  "3",
		DateCreated: "1640995200000",
		DateUpdated: "1640995200000",
		Creator:     UserFixtures.CurrentUser,
		Assignees:   []clickup.User{*UserFixtures.TeamMember1.User},
		DueDate:     "1640908800000", // 2021-12-31 (past date)
		Priority:    &clickup.TaskPriority{Priority: "high"},
	},
	AssignedTask: &clickup.Task{
		ID:          "tk_abf",
		Name:        "Assigned Task",
		Description: "This task is assigned to multiple users",
		Status:      clickup.Status{Status: "to do", Color: "#ff0000"},
		OrderIndex:  "4",
		DateCreated: "1640995200000",
		DateUpdated: "1640995200000",
		Creator:     UserFixtures.CurrentUser,
		Assignees: []clickup.User{
			*UserFixtures.TeamMember1.User,
			*UserFixtures.TeamMember2.User,
		},
		Tags: []clickup.Tag{
			{Name: "bug", TagFg: "#ffffff", TagBg: "#ff0000"},
			{Name: "urgent", TagFg: "#000000", TagBg: "#ffff00"},
		},
	},
}

// CommentFixtures provides pre-configured comments
var CommentFixtures = struct {
	SimpleComment   *clickup.Comment
	ResolvedComment *clickup.Comment
}{
	SimpleComment: &clickup.Comment{
		ID:          TestCommentID,
		Comment:     []clickup.CommentItem{{Text: "This is a test comment"}},
		CommentText: "This is a test comment",
		User:        UserFixtures.CurrentUser,
		Resolved:    false,
		Date:        time.Now().Unix() * 1000,
	},
	ResolvedComment: &clickup.Comment{
		ID:          "cm_xy2",
		Comment:     []clickup.CommentItem{{Text: "This issue is resolved"}},
		CommentText: "This issue is resolved",
		User:        UserFixtures.TeamMember1.User,
		Resolved:    true,
		Date:        time.Now().Unix() * 1000,
	},
}

// Scenarios provides pre-configured API scenarios
type Scenarios struct {
	client *Client
}

// NewScenarios creates a new scenarios helper
func NewScenarios(client *Client) *Scenarios {
	return &Scenarios{client: client}
}

// EmptyWorkspace sets up an empty workspace scenario
func (s *Scenarios) EmptyWorkspace() *Client {
	s.client.Reset()
	s.client.SetCurrentUser(UserFixtures.CurrentUser, nil)
	s.client.SetWorkspaces([]clickup.Team{*WorkspaceFixtures.DefaultWorkspace}, nil)
	s.client.SetSpaces(TestWorkspaceID, []clickup.Space{})
	return s.client
}

// PopulatedWorkspace sets up a workspace with full hierarchy
func (s *Scenarios) PopulatedWorkspace() *Client {
	s.client.Reset()
	s.client.SetCurrentUser(UserFixtures.CurrentUser, nil)
	s.client.SetWorkspaces([]clickup.Team{*WorkspaceFixtures.DefaultWorkspace}, nil)
	s.client.SetSpaces(TestWorkspaceID, []clickup.Space{*HierarchyFixtures.Space1, *HierarchyFixtures.Space2})
	s.client.SetFolders(TestSpaceID, []clickup.Folder{*HierarchyFixtures.Folder1})
	s.client.SetLists(TestFolderID, []clickup.List{*HierarchyFixtures.List1})
	s.client.SetFolderlessLists(TestSpaceID, []clickup.List{*HierarchyFixtures.List2})
	return s.client
}

// TaskListWithTasks sets up a list with various tasks
func (s *Scenarios) TaskListWithTasks() *Client {
	s.client.Reset()
	s.client.SetCurrentUser(UserFixtures.CurrentUser, nil)
	
	// Set up hierarchy
	s.PopulatedWorkspace()
	
	// Add tasks
	tasks := []clickup.Task{
		*TaskFixtures.SimpleTask,
		*TaskFixtures.CompleteTask,
		*TaskFixtures.OverdueTask,
		*TaskFixtures.AssignedTask,
	}
	s.client.SetTaskList(TestListID, tasks)
	
	// Also set individual tasks for GetTask
	for i := range tasks {
		s.client.SetTask(&tasks[i])
	}
	
	return s.client
}

// APIError sets up a scenario with API errors
func (s *Scenarios) APIError() *Client {
	s.client.Reset()
	apiErr := fmt.Errorf("API error: rate limit exceeded")
	s.client.SetCurrentUser(nil, apiErr)
	s.client.SetWorkspaces(nil, apiErr)
	return s.client
}

// Helper methods for test data

// SetFolders sets folders for a space
func (c *Client) SetFolders(spaceID string, folders []clickup.Folder) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.folders[spaceID] = folders
}

// SetLists sets lists for a folder
func (c *Client) SetLists(folderID string, lists []clickup.List) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lists[folderID] = lists
}

// SetFolderlessLists sets folderless lists for a space
func (c *Client) SetFolderlessLists(spaceID string, lists []clickup.List) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.folderlessLists[spaceID] = lists
}
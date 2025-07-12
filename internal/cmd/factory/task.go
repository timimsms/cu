package factory

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/raksul/go-clickup/clickup"
	"github.com/spf13/cobra"
	"github.com/tim/cu/internal/cmd/base"
	"github.com/tim/cu/internal/interfaces"
)

// TaskCommand implements the task command with dependency injection
type TaskCommand struct {
	*base.Command
	subcommands map[string]func(context.Context, []string) error
	
	// Flags
	listID      string
	spaceID     string
	folderID    string
	assignee    string
	status      string
	tag         string
	priority    string
	due         string
	sortBy      string
	order       string
	limit       int
	page        int
	name        string
	description string
	assignees   []string
	tags        []string
}

// createTaskCommand creates a new task command
func (f *Factory) createTaskCommand() interfaces.Command {
	cmd := &TaskCommand{
		Command: &base.Command{
			Use:   "task",
			Short: "Manage tasks",
			Long:  `Create, view, update, and manage ClickUp tasks.`,
			API:    f.api,
			Auth:   f.auth,
			Output: f.output,
			Config: f.config,
		},
		subcommands: make(map[string]func(context.Context, []string) error),
	}

	// Register subcommands
	cmd.subcommands["list"] = cmd.runList
	cmd.subcommands["create"] = cmd.runCreate
	cmd.subcommands["view"] = cmd.runView
	cmd.subcommands["update"] = cmd.runUpdate
	cmd.subcommands["close"] = cmd.runClose
	cmd.subcommands["reopen"] = cmd.runReopen
	cmd.subcommands["search"] = cmd.runSearch

	// Set the execution function
	cmd.Command.RunFunc = cmd.run

	return cmd
}

// run executes the task command
func (c *TaskCommand) run(ctx context.Context, args []string) error {
	// If no subcommand, show usage
	if len(args) == 0 {
		return fmt.Errorf("no subcommand specified. Available subcommands: list, create, view, update, close, reopen, search")
	}

	subcommand := args[0]
	handler, exists := c.subcommands[subcommand]
	if !exists {
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}

	// Execute subcommand with remaining args
	return handler(ctx, args[1:])
}

// runList executes the task list subcommand
func (c *TaskCommand) runList(ctx context.Context, args []string) error {
	// Ensure API client is connected
	if c.API == nil {
		return fmt.Errorf("API client not initialized")
	}

	// If no list is specified, try to use default from config
	if c.listID == "" && c.spaceID == "" && c.folderID == "" {
		c.listID = c.Config.GetString("default_list")
		if c.listID == "" {
			return fmt.Errorf("no list specified. Use --list, --space, or --folder flag, or set a default list with 'cu list default'")
		}
	}

	// TODO: Implement space/folder to list resolution
	// For now, require a list ID
	if c.listID == "" {
		return fmt.Errorf("list ID is required for now. Space/folder resolution coming soon")
	}

	// Build query options
	queryOpts := &interfaces.TaskQueryOptions{
		Page: c.page,
	}

	if c.assignee != "" {
		queryOpts.Assignees = []string{c.assignee}
	}
	if c.status != "" {
		queryOpts.Statuses = []string{c.status}
	}
	if c.tag != "" {
		queryOpts.Tags = []string{c.tag}
	}

	// Get tasks
	tasks, err := c.API.GetTasks(ctx, c.listID, queryOpts)
	if err != nil {
		return fmt.Errorf("failed to get tasks: %w", err)
	}

	// Convert to pointer slice for filtering and sorting
	var taskPtrs []*clickup.Task
	for i := range tasks {
		taskPtrs = append(taskPtrs, &tasks[i])
	}

	// Apply client-side filtering
	taskPtrs = c.filterTasks(taskPtrs, c.priority, c.due)

	// Apply sorting
	c.sortTasks(taskPtrs, c.sortBy, c.order)

	// Apply limit
	if c.limit > 0 && len(taskPtrs) > c.limit {
		taskPtrs = taskPtrs[:c.limit]
	}

	// Format output
	format := c.Config.GetString("output")
	if c.Output != nil {
		if outputFlag := c.getOutputFormat(); outputFlag != "" {
			format = outputFlag
		}
	}

	if format == "table" {
		// Prepare table data
		type taskRow struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Status   string `json:"status"`
			Assignee string `json:"assignee"`
			Priority string `json:"priority"`
			Due      string `json:"due"`
		}

		var rows []taskRow
		for _, task := range taskPtrs {
			row := taskRow{
				ID:       task.ID,
				Name:     truncate(task.Name, 50),
				Status:   c.getTaskStatus(task),
				Assignee: c.getTaskAssignee(task),
				Priority: c.getTaskPriority(task),
				Due:      c.getTaskDueDate(task),
			}
			rows = append(rows, row)
		}

		return c.Output.Print(rows)
	}

	// For other formats, output raw task data
	return c.Output.Print(taskPtrs)
}

// runCreate executes the task create subcommand
func (c *TaskCommand) runCreate(ctx context.Context, args []string) error {
	// Ensure API client is connected
	if c.API == nil {
		return fmt.Errorf("API client not initialized")
	}

	// Get task name from args or flag
	var taskName string
	if len(args) > 0 {
		taskName = args[0]
	} else {
		taskName = c.name
	}

	if taskName == "" {
		return fmt.Errorf("task name is required. Provide it as an argument or use --name flag")
	}

	// If no list is specified, try to use default from config
	if c.listID == "" {
		c.listID = c.Config.GetString("default_list")
		if c.listID == "" {
			return fmt.Errorf("no list specified. Use --list flag or set a default list with 'cu list default'")
		}
	}

	// Build task creation options
	createOpts := &interfaces.TaskCreateOptions{
		Name:        taskName,
		Description: c.description,
		Status:      c.status,
		Priority:    c.priority,
		Tags:        c.tags,
	}

	// Handle assignees
	if len(c.assignees) > 0 {
		createOpts.Assignees = c.assignees
	}

	// Handle due date
	if c.due != "" {
		dueTime, err := parseDueDate(c.due)
		if err != nil {
			return fmt.Errorf("invalid due date format: %w", err)
		}
		// Convert to milliseconds string as expected by ClickUp API
		createOpts.DueDate = fmt.Sprintf("%d", dueTime.Unix()*1000)
	}

	// Create task
	task, err := c.API.CreateTask(ctx, c.listID, createOpts)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	c.Output.PrintSuccess(fmt.Sprintf("Created task: %s (%s)", task.Name, task.ID))
	
	// Output task details if requested
	format := c.Config.GetString("output")
	if format != "table" {
		return c.Output.Print(task)
	}
	
	return nil
}

// runView executes the task view subcommand
func (c *TaskCommand) runView(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task ID is required")
	}

	taskID := args[0]

	// Ensure API client is connected
	if c.API == nil {
		return fmt.Errorf("API client not initialized")
	}

	// Get task
	task, err := c.API.GetTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Format output
	format := c.Config.GetString("output")
	if format == "table" {
		// Display task details in a readable format
		c.Output.PrintInfo(fmt.Sprintf("Task: %s", task.Name))
		c.Output.PrintInfo(fmt.Sprintf("ID: %s", task.ID))
		c.Output.PrintInfo(fmt.Sprintf("Status: %s", c.getTaskStatus(task)))
		c.Output.PrintInfo(fmt.Sprintf("Priority: %s", c.getTaskPriority(task)))
		
		if task.Description != "" {
			c.Output.PrintInfo(fmt.Sprintf("\nDescription:\n%s", task.Description))
		}
		
		if len(task.Assignees) > 0 {
			c.Output.PrintInfo(fmt.Sprintf("\nAssignees: %s", c.getTaskAssignee(task)))
		}
		
		if task.DueDate != nil {
			c.Output.PrintInfo(fmt.Sprintf("Due: %s", c.getTaskDueDate(task)))
		}
		
		return nil
	}

	// For other formats, output raw task data
	return c.Output.Print(task)
}

// runUpdate executes the task update subcommand
func (c *TaskCommand) runUpdate(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task ID is required")
	}

	taskID := args[0]

	// Ensure API client is connected
	if c.API == nil {
		return fmt.Errorf("API client not initialized")
	}

	// Build update options
	updateOpts := &interfaces.TaskUpdateOptions{}
	hasUpdates := false

	if c.name != "" {
		updateOpts.Name = c.name
		hasUpdates = true
	}
	if c.description != "" {
		updateOpts.Description = c.description
		hasUpdates = true
	}
	if c.status != "" {
		updateOpts.Status = c.status
		hasUpdates = true
	}
	if c.priority != "" {
		updateOpts.Priority = c.priority
		hasUpdates = true
	}
	if len(c.assignees) > 0 {
		updateOpts.AddAssignees = c.assignees
		hasUpdates = true
	}
	if c.due != "" {
		dueTime, err := parseDueDate(c.due)
		if err != nil {
			return fmt.Errorf("invalid due date format: %w", err)
		}
		// Convert to milliseconds string as expected by ClickUp API
		updateOpts.DueDate = fmt.Sprintf("%d", dueTime.Unix()*1000)
		hasUpdates = true
	}

	if !hasUpdates {
		return fmt.Errorf("no updates specified")
	}

	// Update task
	task, err := c.API.UpdateTask(ctx, taskID, updateOpts)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	c.Output.PrintSuccess(fmt.Sprintf("Updated task: %s", task.Name))
	return nil
}

// runClose executes the task close subcommand
func (c *TaskCommand) runClose(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task ID is required")
	}

	taskID := args[0]

	// Ensure API client is connected
	if c.API == nil {
		return fmt.Errorf("API client not initialized")
	}

	// Close task by setting status to "closed"
	updateOpts := &interfaces.TaskUpdateOptions{
		Status: "closed",
	}

	task, err := c.API.UpdateTask(ctx, taskID, updateOpts)
	if err != nil {
		return fmt.Errorf("failed to close task: %w", err)
	}

	c.Output.PrintSuccess(fmt.Sprintf("Closed task: %s", task.Name))
	return nil
}

// runReopen executes the task reopen subcommand
func (c *TaskCommand) runReopen(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task ID is required")
	}

	taskID := args[0]

	// Ensure API client is connected
	if c.API == nil {
		return fmt.Errorf("API client not initialized")
	}

	// Reopen task by setting status to "open"
	updateOpts := &interfaces.TaskUpdateOptions{
		Status: "open",
	}

	task, err := c.API.UpdateTask(ctx, taskID, updateOpts)
	if err != nil {
		return fmt.Errorf("failed to reopen task: %w", err)
	}

	c.Output.PrintSuccess(fmt.Sprintf("Reopened task: %s", task.Name))
	return nil
}

// runSearch executes the task search subcommand
func (c *TaskCommand) runSearch(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("search query is required")
	}

	// query := strings.Join(args, " ")

	// Ensure API client is connected
	if c.API == nil {
		return fmt.Errorf("API client not initialized")
	}

	// For now, we'll use the list endpoint with filtering
	// TODO: Implement proper search when API supports it
	return fmt.Errorf("search functionality not yet implemented")
}

// Helper methods

func (c *TaskCommand) getTaskStatus(task *clickup.Task) string {
	return task.Status.Status
}

func (c *TaskCommand) getTaskAssignee(task *clickup.Task) string {
	if len(task.Assignees) > 0 {
		return task.Assignees[0].Username
	}
	return "unassigned"
}

func (c *TaskCommand) getTaskPriority(task *clickup.Task) string {
	switch task.Priority.Priority {
	case "1":
		return "urgent"
	case "2":
		return "high"
	case "3":
		return "normal"
	case "4":
		return "low"
	}
	return "none"
}

func (c *TaskCommand) getTaskDueDate(task *clickup.Task) string {
	if task.DueDate != nil {
		if t := task.DueDate.Time(); t != nil {
			return t.Format("2006-01-02")
		}
	}
	return ""
}

func (c *TaskCommand) filterTasks(tasks []*clickup.Task, priority, due string) []*clickup.Task {
	var filtered []*clickup.Task

	for _, task := range tasks {
		// Filter by priority
		if priority != "" && c.getTaskPriority(task) != priority {
			continue
		}

		// Filter by due date
		if due != "" {
			taskDue := c.getTaskDueDate(task)
			if !matchesDueFilter(taskDue, due) {
				continue
			}
		}

		filtered = append(filtered, task)
	}

	return filtered
}

func (c *TaskCommand) sortTasks(tasks []*clickup.Task, sortBy, order string) {
	if sortBy == "" {
		sortBy = "created"
	}
	if order == "" {
		order = "desc"
	}

	sort.Slice(tasks, func(i, j int) bool {
		var less bool

		switch sortBy {
		case "name":
			less = tasks[i].Name < tasks[j].Name
		case "status":
			less = c.getTaskStatus(tasks[i]) < c.getTaskStatus(tasks[j])
		case "priority":
			// Priority is reversed (1 is highest)
			iPri := getPriorityValue(tasks[i])
			jPri := getPriorityValue(tasks[j])
			less = iPri < jPri
		case "due":
			iDue := getTaskDueTime(tasks[i])
			jDue := getTaskDueTime(tasks[j])
			less = iDue.Before(jDue)
		case "created":
			fallthrough
		default:
			iCreated := getTaskCreatedTime(tasks[i])
			jCreated := getTaskCreatedTime(tasks[j])
			less = iCreated.Before(jCreated)
		}

		if order == "desc" {
			return !less
		}
		return less
	})
}

func (c *TaskCommand) getOutputFormat() string {
	// This would be set by cobra flags
	return ""
}

// GetCobraCommand returns the cobra command with subcommands
func (c *TaskCommand) GetCobraCommand() *cobra.Command {
	cmd := c.Command.GetCobraCommand()

	// Add subcommands
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		Long:  `List tasks from ClickUp with various filtering and sorting options.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Set flags from cobra command
			c.listID, _ = cmd.Flags().GetString("list")
			c.spaceID, _ = cmd.Flags().GetString("space")
			c.folderID, _ = cmd.Flags().GetString("folder")
			c.assignee, _ = cmd.Flags().GetString("assignee")
			c.status, _ = cmd.Flags().GetString("status")
			c.tag, _ = cmd.Flags().GetString("tag")
			c.priority, _ = cmd.Flags().GetString("priority")
			c.due, _ = cmd.Flags().GetString("due")
			c.sortBy, _ = cmd.Flags().GetString("sort")
			c.order, _ = cmd.Flags().GetString("order")
			c.limit, _ = cmd.Flags().GetInt("limit")
			c.page, _ = cmd.Flags().GetInt("page")
			
			return c.runList(cmd.Context(), args)
		},
	}

	// Add flags to list subcommand
	listCmd.Flags().StringP("list", "l", "", "List ID to fetch tasks from")
	listCmd.Flags().StringP("space", "s", "", "Space ID to fetch tasks from")
	listCmd.Flags().StringP("folder", "f", "", "Folder ID to fetch tasks from")
	listCmd.Flags().StringP("assignee", "a", "", "Filter by assignee (email or ID)")
	listCmd.Flags().String("status", "", "Filter by status")
	listCmd.Flags().String("tag", "", "Filter by tag")
	listCmd.Flags().String("priority", "", "Filter by priority (urgent, high, normal, low)")
	listCmd.Flags().String("due", "", "Filter by due date (today, tomorrow, week, overdue)")
	listCmd.Flags().String("sort", "created", "Sort by field (name, status, priority, due, created)")
	listCmd.Flags().String("order", "desc", "Sort order (asc, desc)")
	listCmd.Flags().Int("limit", 20, "Maximum number of tasks to display")
	listCmd.Flags().Int("page", 0, "Page number for pagination")

	createCmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create a new task",
		Long:  `Create a new task in ClickUp with the specified name and optional properties.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Set flags from cobra command
			c.name, _ = cmd.Flags().GetString("name")
			c.listID, _ = cmd.Flags().GetString("list")
			c.description, _ = cmd.Flags().GetString("description")
			c.assignees, _ = cmd.Flags().GetStringSlice("assignee")
			c.status, _ = cmd.Flags().GetString("status")
			c.priority, _ = cmd.Flags().GetString("priority")
			c.due, _ = cmd.Flags().GetString("due")
			c.tags, _ = cmd.Flags().GetStringSlice("tag")
			
			return c.runCreate(cmd.Context(), args)
		},
	}

	// Add flags to create subcommand
	createCmd.Flags().StringP("name", "n", "", "Task name (alternative to providing as argument)")
	createCmd.Flags().StringP("list", "l", "", "List ID to create task in")
	createCmd.Flags().StringP("description", "d", "", "Task description")
	createCmd.Flags().StringSliceP("assignee", "a", nil, "Assignees (email or ID, can be repeated)")
	createCmd.Flags().String("status", "", "Initial status")
	createCmd.Flags().String("priority", "", "Priority (urgent, high, normal, low)")
	createCmd.Flags().String("due", "", "Due date (YYYY-MM-DD or relative like 'tomorrow')")
	createCmd.Flags().StringSlice("tag", nil, "Tags to add (can be repeated)")

	viewCmd := &cobra.Command{
		Use:   "view <task-id>",
		Short: "View task details",
		Long:  `Display detailed information about a specific task.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runView(cmd.Context(), args)
		},
	}

	updateCmd := &cobra.Command{
		Use:   "update <task-id>",
		Short: "Update a task",
		Long:  `Update properties of an existing task.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Set flags from cobra command
			c.name, _ = cmd.Flags().GetString("name")
			c.description, _ = cmd.Flags().GetString("description")
			c.assignees, _ = cmd.Flags().GetStringSlice("assignee")
			c.status, _ = cmd.Flags().GetString("status")
			c.priority, _ = cmd.Flags().GetString("priority")
			c.due, _ = cmd.Flags().GetString("due")
			
			return c.runUpdate(cmd.Context(), args)
		},
	}

	// Add flags to update subcommand
	updateCmd.Flags().StringP("name", "n", "", "New task name")
	updateCmd.Flags().StringP("description", "d", "", "New task description")
	updateCmd.Flags().StringSliceP("assignee", "a", nil, "New assignees (replaces existing)")
	updateCmd.Flags().String("status", "", "New status")
	updateCmd.Flags().String("priority", "", "New priority (urgent, high, normal, low)")
	updateCmd.Flags().String("due", "", "New due date (YYYY-MM-DD or relative)")

	closeCmd := &cobra.Command{
		Use:   "close <task-id>",
		Short: "Close a task",
		Long:  `Mark a task as closed/completed.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runClose(cmd.Context(), args)
		},
	}

	reopenCmd := &cobra.Command{
		Use:   "reopen <task-id>",
		Short: "Reopen a task",
		Long:  `Reopen a previously closed task.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runReopen(cmd.Context(), args)
		},
	}

	searchCmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search for tasks",
		Long:  `Search for tasks by name or description.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runSearch(cmd.Context(), args)
		},
	}

	cmd.AddCommand(listCmd, createCmd, viewCmd, updateCmd, closeCmd, reopenCmd, searchCmd)

	return cmd
}

// Utility functions

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func parseDueDate(due string) (time.Time, error) {
	// Handle relative dates
	now := time.Now()
	switch strings.ToLower(due) {
	case "today":
		return time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location()), nil
	case "tomorrow":
		return now.AddDate(0, 0, 1), nil
	case "week":
		return now.AddDate(0, 0, 7), nil
	}

	// Try to parse as date
	t, err := time.Parse("2006-01-02", due)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format. Use YYYY-MM-DD or relative dates (today, tomorrow, week)")
	}
	return t, nil
}

func parseClickUpTime(timeStr string) (time.Time, error) {
	// ClickUp uses milliseconds since epoch
	// First, try to parse as milliseconds
	var ts int64
	if _, err := fmt.Sscanf(timeStr, "%d", &ts); err == nil {
		return time.Unix(ts/1000, (ts%1000)*1000000), nil
	}
	
	// Fallback to RFC3339
	return time.Parse(time.RFC3339, timeStr)
}

func matchesDueFilter(taskDue, filter string) bool {
	if taskDue == "" {
		return false
	}

	taskTime, err := time.Parse("2006-01-02", taskDue)
	if err != nil {
		return false
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	switch strings.ToLower(filter) {
	case "today":
		return taskTime.Equal(today)
	case "tomorrow":
		tomorrow := today.AddDate(0, 0, 1)
		return taskTime.Equal(tomorrow)
	case "week":
		weekFromNow := today.AddDate(0, 0, 7)
		return taskTime.After(today) && taskTime.Before(weekFromNow)
	case "overdue":
		return taskTime.Before(today)
	}

	return false
}

func getPriorityValue(task *clickup.Task) int {
	switch task.Priority.Priority {
	case "1":
		return 1
	case "2":
		return 2
	case "3":
		return 3
	case "4":
		return 4
	}
	return 5 // No priority
}

func getTaskDueTime(task *clickup.Task) time.Time {
	if task.DueDate != nil {
		if t := task.DueDate.Time(); t != nil {
			return *t
		}
	}
	return time.Time{}
}

func getTaskCreatedTime(task *clickup.Task) time.Time {
	if task.DateCreated != "" {
		if t, err := parseClickUpTime(task.DateCreated); err == nil {
			return t
		}
	}
	return time.Time{}
}
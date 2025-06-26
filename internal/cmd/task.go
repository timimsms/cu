package cmd

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/raksul/go-clickup/clickup"
	"github.com/spf13/cobra"
	"github.com/tim/cu/internal/api"
	"github.com/tim/cu/internal/config"
	"github.com/tim/cu/internal/output"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks",
	Long:  `Create, view, update, and manage ClickUp tasks.`,
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	Long:  `List tasks from ClickUp with various filtering and sorting options.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		// Create API client
		client, err := api.NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create API client: %v\n", err)
			os.Exit(1)
		}

		// Get flags
		listID, _ := cmd.Flags().GetString("list")
		spaceID, _ := cmd.Flags().GetString("space")
		folderID, _ := cmd.Flags().GetString("folder")
		assignee, _ := cmd.Flags().GetString("assignee")
		status, _ := cmd.Flags().GetString("status")
		tag, _ := cmd.Flags().GetString("tag")
		priority, _ := cmd.Flags().GetString("priority")
		due, _ := cmd.Flags().GetString("due")
		sortBy, _ := cmd.Flags().GetString("sort")
		order, _ := cmd.Flags().GetString("order")
		limit, _ := cmd.Flags().GetInt("limit")
		page, _ := cmd.Flags().GetInt("page")

		// If no list is specified, try to use default from config
		if listID == "" && spaceID == "" && folderID == "" {
			listID = config.GetString("default_list")
			if listID == "" {
				fmt.Fprintln(os.Stderr, "No list specified. Use --list, --space, or --folder flag, or set a default list with 'cu list default'")
				os.Exit(1)
			}
		}

		// TODO: Implement space/folder to list resolution
		// For now, require a list ID
		if listID == "" {
			fmt.Fprintln(os.Stderr, "List ID is required for now. Space/folder resolution coming soon.")
			os.Exit(1)
		}

		// Build query options
		queryOpts := &api.TaskQueryOptions{
			Page: page,
		}

		if assignee != "" {
			queryOpts.Assignees = []string{assignee}
		}
		if status != "" {
			queryOpts.Statuses = []string{status}
		}
		if tag != "" {
			queryOpts.Tags = []string{tag}
		}

		// Get tasks
		tasks, err := client.GetTasks(ctx, listID, queryOpts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get tasks: %v\n", err)
			os.Exit(1)
		}

		// Apply client-side filtering
		tasks = filterTasks(tasks, priority, due)

		// Apply sorting
		sortTasks(tasks, sortBy, order)

		// Apply limit
		if limit > 0 && len(tasks) > limit {
			tasks = tasks[:limit]
		}

		// Format output
		format := cmd.Flag("output").Value.String()

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
			for _, task := range tasks {
				row := taskRow{
					ID:       task.ID,
					Name:     truncate(task.Name, 50),
					Status:   getTaskStatus(task),
					Assignee: getTaskAssignee(task),
					Priority: getTaskPriority(task),
					Due:      getTaskDueDate(task),
				}
				rows = append(rows, row)
			}

			if err := output.Format(format, rows); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
		} else {
			// For other formats, output raw task data
			if err := output.Format(format, tasks); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

var taskCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new task",
	Long:  `Create a new task in ClickUp with the specified name and optional properties.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		// Get task name from args or flag
		var name string
		if len(args) > 0 {
			name = args[0]
		} else {
			name, _ = cmd.Flags().GetString("name")
		}

		if name == "" {
			fmt.Fprintln(os.Stderr, "Task name is required. Provide it as an argument or use --name flag")
			os.Exit(1)
		}

		// Create API client
		client, err := api.NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create API client: %v\n", err)
			os.Exit(1)
		}

		// Get flags
		listID, _ := cmd.Flags().GetString("list")
		description, _ := cmd.Flags().GetString("description")
		assignees, _ := cmd.Flags().GetStringSlice("assignee")
		status, _ := cmd.Flags().GetString("status")
		priority, _ := cmd.Flags().GetString("priority")
		dueDate, _ := cmd.Flags().GetString("due")
		tags, _ := cmd.Flags().GetStringSlice("tag")

		// If no list is specified, try to use default from config
		if listID == "" {
			listID = config.GetString("default_list")
			if listID == "" {
				fmt.Fprintln(os.Stderr, "No list specified. Use --list flag or set a default list with 'cu list default'")
				os.Exit(1)
			}
		}

		// Build task creation options
		createOpts := &api.TaskCreateOptions{
			Name:        name,
			Description: description,
			Status:      status,
			Priority:    priority,
			Tags:        tags,
		}

		// Handle assignees
		if len(assignees) > 0 {
			createOpts.Assignees = assignees
		}

		// Handle due date
		if dueDate != "" {
			// TODO: Parse due date strings like "today", "tomorrow", etc.
			// For now, expect ISO date format
			createOpts.DueDate = dueDate
		}

		// Create task
		task, err := client.CreateTask(ctx, listID, createOpts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create task: %v\n", err)
			os.Exit(1)
		}

		// Format output
		format := cmd.Flag("output").Value.String()

		if format == "table" {
			// Simple success message with task details
			fmt.Printf("✓ Created task %s: %s\n", task.ID, task.Name)
			if task.URL != "" {
				fmt.Printf("  View in ClickUp: %s\n", task.URL)
			}
		} else {
			// For other formats, output the created task
			if err := output.Format(format, task); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

var taskViewCmd = &cobra.Command{
	Use:   "view [task-id]",
	Short: "View task details",
	Long:  `View detailed information about a specific task.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		taskID := args[0]

		// Create API client
		client, err := api.NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create API client: %v\n", err)
			os.Exit(1)
		}

		// Get task
		task, err := client.GetTask(ctx, taskID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get task: %v\n", err)
			os.Exit(1)
		}

		// Format output
		format := cmd.Flag("output").Value.String()

		if format == "table" {
			// Display detailed task information
			fmt.Printf("Task: %s\n", task.Name)
			fmt.Printf("ID: %s\n", task.ID)
			if task.URL != "" {
				fmt.Printf("URL: %s\n", task.URL)
			}
			fmt.Printf("Status: %s\n", getTaskStatus(*task))
			fmt.Printf("Priority: %s\n", getTaskPriority(*task))

			if len(task.Assignees) > 0 {
				fmt.Printf("Assignees: ")
				for i, assignee := range task.Assignees {
					if i > 0 {
						fmt.Printf(", ")
					}
					fmt.Printf("%s", assignee.Username)
				}
				fmt.Println()
			}

			if task.Description != "" {
				fmt.Printf("\nDescription:\n%s\n", task.Description)
			}

			if task.DueDate != nil {
				fmt.Printf("Due: %s\n", getTaskDueDate(*task))
			}

			if task.DateCreated != "" {
				fmt.Printf("Created: %s\n", task.DateCreated)
			}

			if task.DateUpdated != "" {
				fmt.Printf("Updated: %s\n", task.DateUpdated)
			}

			if len(task.Tags) > 0 {
				fmt.Printf("Tags: ")
				for i, tag := range task.Tags {
					if i > 0 {
						fmt.Printf(", ")
					}
					fmt.Printf("%s", tag.Name)
				}
				fmt.Println()
			}
		} else {
			// For other formats, output the raw task
			if err := output.Format(format, task); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

var taskUpdateCmd = &cobra.Command{
	Use:   "update [task-id]",
	Short: "Update a task",
	Long:  `Update an existing task with new properties.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		taskID := args[0]

		// Create API client
		client, err := api.NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create API client: %v\n", err)
			os.Exit(1)
		}

		// Get flags
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		status, _ := cmd.Flags().GetString("status")
		priority, _ := cmd.Flags().GetString("priority")
		dueDate, _ := cmd.Flags().GetString("due")
		addAssignees, _ := cmd.Flags().GetStringSlice("add-assignee")
		removeAssignees, _ := cmd.Flags().GetStringSlice("remove-assignee")
		tags, _ := cmd.Flags().GetStringSlice("tag")

		// Build update options
		updateOpts := &api.TaskUpdateOptions{
			Name:            name,
			Description:     description,
			Status:          status,
			Priority:        priority,
			DueDate:         dueDate,
			Tags:            tags,
			AddAssignees:    addAssignees,
			RemoveAssignees: removeAssignees,
		}

		// Check if any updates were specified
		if !updateOpts.HasUpdates() {
			fmt.Fprintln(os.Stderr, "No updates specified. Use flags like --name, --status, --priority, etc.")
			os.Exit(1)
		}

		// Update task
		task, err := client.UpdateTask(ctx, taskID, updateOpts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to update task: %v\n", err)
			os.Exit(1)
		}

		// Format output
		format := cmd.Flag("output").Value.String()

		if format == "table" {
			fmt.Printf("✓ Updated task %s: %s\n", task.ID, task.Name)
			if task.URL != "" {
				fmt.Printf("  View in ClickUp: %s\n", task.URL)
			}
		} else {
			// For other formats, output the updated task
			if err := output.Format(format, task); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

var taskCloseCmd = &cobra.Command{
	Use:   "close [task-id]",
	Short: "Close a task",
	Long:  `Close a task by marking it as complete.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		taskID := args[0]

		// Create API client
		client, err := api.NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create API client: %v\n", err)
			os.Exit(1)
		}

		// Find a closed status in the same list
		// For now, we'll use "complete" as the closed status
		// TODO: Query the list's statuses to find the actual closed status
		updateOpts := &api.TaskUpdateOptions{
			Status: "complete",
		}

		// Update task
		updatedTask, err := client.UpdateTask(ctx, taskID, updateOpts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to close task: %v\n", err)
			os.Exit(1)
		}

		// Format output
		format := cmd.Flag("output").Value.String()

		if format == "table" {
			fmt.Printf("✓ Closed task %s: %s\n", updatedTask.ID, updatedTask.Name)
			if updatedTask.URL != "" {
				fmt.Printf("  View in ClickUp: %s\n", updatedTask.URL)
			}
		} else {
			// For other formats, output the updated task
			if err := output.Format(format, updatedTask); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

var taskReopenCmd = &cobra.Command{
	Use:   "reopen [task-id]",
	Short: "Reopen a task",
	Long:  `Reopen a closed task by marking it as open/in progress.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		taskID := args[0]

		// Create API client
		client, err := api.NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create API client: %v\n", err)
			os.Exit(1)
		}

		// Get the status flag or use default
		status, _ := cmd.Flags().GetString("status")
		if status == "" {
			status = "open" // Default to "open"
		}

		// Update task
		updateOpts := &api.TaskUpdateOptions{
			Status: status,
		}

		updatedTask, err := client.UpdateTask(ctx, taskID, updateOpts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to reopen task: %v\n", err)
			os.Exit(1)
		}

		// Format output
		format := cmd.Flag("output").Value.String()

		if format == "table" {
			fmt.Printf("✓ Reopened task %s: %s\n", updatedTask.ID, updatedTask.Name)
			fmt.Printf("  New status: %s\n", updatedTask.Status.Status)
			if updatedTask.URL != "" {
				fmt.Printf("  View in ClickUp: %s\n", updatedTask.URL)
			}
		} else {
			// For other formats, output the updated task
			if err := output.Format(format, updatedTask); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

var taskSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for tasks",
	Long:  `Search for tasks across all lists in your workspace. Searches in task names and descriptions.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		query := strings.Join(args, " ")

		// Create API client
		client, err := api.NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create API client: %v\n", err)
			os.Exit(1)
		}

		// Get search scope from flags
		spaceID, _ := cmd.Flags().GetString("space")
		listID, _ := cmd.Flags().GetString("list")
		searchDescription, _ := cmd.Flags().GetBool("include-description")
		limit, _ := cmd.Flags().GetInt("limit")

		// Get workspaces to search
		workspaces, err := client.GetWorkspaces(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get workspaces: %v\n", err)
			os.Exit(1)
		}

		if len(workspaces) == 0 {
			fmt.Fprintln(os.Stderr, "No workspaces found")
			os.Exit(1)
		}

		var allTasks []clickup.Task
		var searchErrors []string

		// If specific list is provided, search only that list
		if listID != "" {
			tasks, err := client.GetTasks(ctx, listID, &api.TaskQueryOptions{})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to get tasks from list %s: %v\n", listID, err)
				os.Exit(1)
			}
			allTasks = tasks
		} else {
			// Search across all lists in workspace or space
			for _, workspace := range workspaces {
				spaces, err := client.GetSpaces(ctx, workspace.ID)
				if err != nil {
					searchErrors = append(searchErrors, fmt.Sprintf("Failed to get spaces for workspace %s: %v", workspace.Name, err))
					continue
				}

				for _, space := range spaces {
					// Skip if specific space is requested and this isn't it
					if spaceID != "" && space.ID != spaceID && space.Name != spaceID {
						continue
					}

					// Get folders in space
					folders, err := client.GetFolders(ctx, space.ID)
					if err != nil {
						searchErrors = append(searchErrors, fmt.Sprintf("Failed to get folders for space %s: %v", space.Name, err))
						continue
					}

					// Get tasks from folders
					for _, folder := range folders {
						lists, err := client.GetLists(ctx, folder.ID)
						if err != nil {
							searchErrors = append(searchErrors, fmt.Sprintf("Failed to get lists for folder %s: %v", folder.Name, err))
							continue
						}

						for _, list := range lists {
							tasks, err := client.GetTasks(ctx, list.ID, &api.TaskQueryOptions{})
							if err != nil {
								searchErrors = append(searchErrors, fmt.Sprintf("Failed to get tasks for list %s: %v", list.Name, err))
								continue
							}
							allTasks = append(allTasks, tasks...)
						}
					}

					// Get folderless lists
					lists, err := client.GetFolderlessLists(ctx, space.ID)
					if err != nil {
						searchErrors = append(searchErrors, fmt.Sprintf("Failed to get folderless lists for space %s: %v", space.Name, err))
						continue
					}

					for _, list := range lists {
						tasks, err := client.GetTasks(ctx, list.ID, &api.TaskQueryOptions{})
						if err != nil {
							searchErrors = append(searchErrors, fmt.Sprintf("Failed to get tasks for list %s: %v", list.Name, err))
							continue
						}
						allTasks = append(allTasks, tasks...)
					}
				}
			}
		}

		// Print any errors encountered during search
		if len(searchErrors) > 0 {
			fmt.Fprintln(os.Stderr, "Some errors occurred during search:")
			for _, err := range searchErrors {
				fmt.Fprintf(os.Stderr, "  - %s\n", err)
			}
		}

		// Filter tasks based on search query
		query = strings.ToLower(query)
		var matchedTasks []clickup.Task

		for _, task := range allTasks {
			nameMatch := strings.Contains(strings.ToLower(task.Name), query)
			descMatch := searchDescription && strings.Contains(strings.ToLower(task.Description), query)

			if nameMatch || descMatch {
				matchedTasks = append(matchedTasks, task)
			}
		}

		// Apply limit
		if limit > 0 && len(matchedTasks) > limit {
			matchedTasks = matchedTasks[:limit]
		}

		// Format output
		format := cmd.Flag("output").Value.String()

		if format == "table" {
			if len(matchedTasks) == 0 {
				fmt.Printf("No tasks found matching '%s'\n", strings.Join(args, " "))
				return
			}

			fmt.Printf("Found %d task(s) matching '%s':\n\n", len(matchedTasks), strings.Join(args, " "))

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
			for _, task := range matchedTasks {
				row := taskRow{
					ID:       task.ID,
					Name:     truncate(task.Name, 50),
					Status:   getTaskStatus(task),
					Assignee: getTaskAssignee(task),
					Priority: getTaskPriority(task),
					Due:      getTaskDueDate(task),
				}
				rows = append(rows, row)
			}

			if err := output.Format(format, rows); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
		} else {
			// For other formats, output raw task data
			if err := output.Format(format, matchedTasks); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskCreateCmd)
	taskCmd.AddCommand(taskViewCmd)
	taskCmd.AddCommand(taskUpdateCmd)
	taskCmd.AddCommand(taskCloseCmd)
	taskCmd.AddCommand(taskReopenCmd)
	taskCmd.AddCommand(taskSearchCmd)

	// List command flags
	taskListCmd.Flags().StringP("list", "l", "", "List ID or name")
	taskListCmd.Flags().StringP("space", "s", "", "Space ID or name")
	taskListCmd.Flags().StringP("folder", "f", "", "Folder ID or name")
	taskListCmd.Flags().String("assignee", "", "Filter by assignee (username or ID)")
	taskListCmd.Flags().String("status", "", "Filter by status")
	taskListCmd.Flags().String("tag", "", "Filter by tag")
	taskListCmd.Flags().String("priority", "", "Filter by priority")
	taskListCmd.Flags().String("due", "", "Filter by due date (today, tomorrow, week, overdue)")
	taskListCmd.Flags().Int("limit", 30, "Maximum number of tasks to return")
	taskListCmd.Flags().Int("page", 0, "Page number for pagination")
	taskListCmd.Flags().String("sort", "", "Sort by field (created, updated, due, priority)")
	taskListCmd.Flags().String("order", "asc", "Sort order (asc, desc)")

	// Create command flags
	taskCreateCmd.Flags().StringP("name", "n", "", "Task name (alternative to providing as argument)")
	taskCreateCmd.Flags().StringP("list", "l", "", "List ID to create task in")
	taskCreateCmd.Flags().StringP("description", "d", "", "Task description")
	taskCreateCmd.Flags().StringSliceP("assignee", "a", []string{}, "Assignees (username or ID)")
	taskCreateCmd.Flags().StringP("status", "s", "", "Task status")
	taskCreateCmd.Flags().StringP("priority", "p", "", "Task priority (urgent, high, normal, low)")
	taskCreateCmd.Flags().String("due", "", "Due date (ISO format or 'today', 'tomorrow')")
	taskCreateCmd.Flags().StringSlice("tag", []string{}, "Tags to add to the task")

	// Update command flags
	taskUpdateCmd.Flags().StringP("name", "n", "", "New task name")
	taskUpdateCmd.Flags().StringP("description", "d", "", "New task description")
	taskUpdateCmd.Flags().StringP("status", "s", "", "New task status")
	taskUpdateCmd.Flags().StringP("priority", "p", "", "New task priority (urgent, high, normal, low)")
	taskUpdateCmd.Flags().String("due", "", "New due date (ISO format or 'today', 'tomorrow')")
	taskUpdateCmd.Flags().StringSlice("tag", []string{}, "Replace tags with these tags")
	taskUpdateCmd.Flags().StringSlice("add-assignee", []string{}, "Add assignees (username or ID)")
	taskUpdateCmd.Flags().StringSlice("remove-assignee", []string{}, "Remove assignees (username or ID)")

	// Reopen command flags
	taskReopenCmd.Flags().StringP("status", "s", "", "Status to set when reopening (default: open)")

	// Search command flags
	taskSearchCmd.Flags().StringP("space", "s", "", "Limit search to specific space")
	taskSearchCmd.Flags().StringP("list", "l", "", "Limit search to specific list")
	taskSearchCmd.Flags().Bool("include-description", false, "Search in task descriptions as well as names")
	taskSearchCmd.Flags().Int("limit", 50, "Maximum number of results to return")
}

// Helper functions

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func getTaskStatus(task clickup.Task) string {
	return task.Status.Status
}

func getTaskAssignee(task clickup.Task) string {
	if len(task.Assignees) > 0 {
		return task.Assignees[0].Username
	}
	return ""
}

func getTaskPriority(task clickup.Task) string {
	// Priority is a struct, check if it has a value
	if task.Priority.Priority != "" {
		return task.Priority.Priority
	}
	return "Normal"
}

func getTaskDueDate(task clickup.Task) string {
	if task.DueDate == nil {
		return ""
	}

	// DueDate is a *Date type, convert to string
	t := task.DueDate.Time()
	if t != nil {
		return formatRelativeTime(*t)
	}
	return ""
}

// formatRelativeTime formats a time as relative to now
func formatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := t.Sub(now)

	if diff < 0 {
		// Past
		diff = -diff
		switch {
		case diff < time.Hour:
			return fmt.Sprintf("%d minutes ago", int(diff.Minutes()))
		case diff < 24*time.Hour:
			return fmt.Sprintf("%d hours ago", int(diff.Hours()))
		case diff < 7*24*time.Hour:
			return fmt.Sprintf("%d days ago", int(diff.Hours()/24))
		default:
			return t.Format("Jan 2, 2006")
		}
	} else {
		// Future
		switch {
		case diff < time.Hour:
			return fmt.Sprintf("in %d minutes", int(diff.Minutes()))
		case diff < 24*time.Hour:
			return fmt.Sprintf("in %d hours", int(diff.Hours()))
		case diff < 7*24*time.Hour:
			days := int(diff.Hours() / 24)
			if days == 1 {
				return "tomorrow"
			}
			return fmt.Sprintf("in %d days", days)
		default:
			return t.Format("Jan 2, 2006")
		}
	}
}

// filterTasks applies client-side filtering for priority and due date
func filterTasks(tasks []clickup.Task, priority, due string) []clickup.Task {
	if priority == "" && due == "" {
		return tasks
	}

	filtered := make([]clickup.Task, 0, len(tasks))
	now := time.Now()

	for _, task := range tasks {
		// Filter by priority
		if priority != "" {
			taskPriority := strings.ToLower(getTaskPriority(task))
			filterPriority := strings.ToLower(priority)
			if taskPriority != filterPriority {
				continue
			}
		}

		// Filter by due date
		if due != "" && task.DueDate != nil {
			dueTime := task.DueDate.Time()
			if dueTime == nil {
				continue
			}

			switch due {
			case "today":
				if !isToday(*dueTime) {
					continue
				}
			case "tomorrow":
				if !isTomorrow(*dueTime) {
					continue
				}
			case "week":
				if !isThisWeek(*dueTime) {
					continue
				}
			case "overdue":
				if !dueTime.Before(now) {
					continue
				}
			}
		} else if due != "" && task.DueDate == nil {
			// Skip tasks without due dates if filtering by due date
			continue
		}

		filtered = append(filtered, task)
	}

	return filtered
}

// sortTasks sorts tasks by the specified field and order
func sortTasks(tasks []clickup.Task, sortBy, order string) {
	if sortBy == "" {
		return
	}

	sort.Slice(tasks, func(i, j int) bool {
		var less bool

		switch sortBy {
		case "created":
			less = tasks[i].DateCreated < tasks[j].DateCreated
		case "updated":
			less = tasks[i].DateUpdated < tasks[j].DateUpdated
		case "due":
			// Tasks without due dates go to the end
			if tasks[i].DueDate == nil && tasks[j].DueDate == nil {
				less = false
			} else if tasks[i].DueDate == nil {
				less = false
			} else if tasks[j].DueDate == nil {
				less = true
			} else {
				iTime := tasks[i].DueDate.Time()
				jTime := tasks[j].DueDate.Time()
				if iTime != nil && jTime != nil {
					less = iTime.Before(*jTime)
				}
			}
		case "priority":
			// Convert priority to number for comparison (lower number = higher priority)
			iPriority := getPriorityValue(getTaskPriority(tasks[i]))
			jPriority := getPriorityValue(getTaskPriority(tasks[j]))
			less = iPriority < jPriority
		default:
			// Default to sorting by name
			less = tasks[i].Name < tasks[j].Name
		}

		if order == "desc" {
			return !less
		}
		return less
	})
}

// Helper functions for date filtering
func isToday(t time.Time) bool {
	now := time.Now()
	y1, m1, d1 := now.Date()
	y2, m2, d2 := t.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func isTomorrow(t time.Time) bool {
	tomorrow := time.Now().AddDate(0, 0, 1)
	y1, m1, d1 := tomorrow.Date()
	y2, m2, d2 := t.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func isThisWeek(t time.Time) bool {
	now := time.Now()
	weekFromNow := now.AddDate(0, 0, 7)
	return t.After(now) && t.Before(weekFromNow)
}

func getPriorityValue(priority string) int {
	switch strings.ToLower(priority) {
	case "urgent":
		return 1
	case "high":
		return 2
	case "normal":
		return 3
	case "low":
		return 4
	default:
		return 3 // Default to normal
	}
}

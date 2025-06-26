package cmd

import (
	"context"
	"fmt"
	"os"
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
		// priority, _ := cmd.Flags().GetString("priority") // TODO: implement priority filtering
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
					ID:     task.ID,
					Name:   truncate(task.Name, 50),
					Status: getTaskStatus(task),
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

func init() {
	taskCmd.AddCommand(taskListCmd)
	
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
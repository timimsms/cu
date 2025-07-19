package factory

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/raksul/go-clickup/clickup"
	"github.com/spf13/cobra"
	"github.com/tim/cu/internal/cmd/base"
	"github.com/tim/cu/internal/interfaces"
)

// ExportCommand implements the export command with dependency injection
type ExportCommand struct {
	*base.Command
	subcommands map[string]func(context.Context, []string) error

	// Output writer for testing
	outputWriter io.Writer

	// Flags
	listID     string
	spaceID    string
	format     string
	outputFile string
	status     string
	priority   string
	assignee   string
}

// createExportCommand creates a new export command
func (f *Factory) createExportCommand() interfaces.Command {
	cmd := &ExportCommand{
		Command: &base.Command{
			Use:    "export",
			Short:  "Export data to various formats",
			Long:   `Export ClickUp data to CSV, JSON, or Markdown formats.`,
			API:    f.api,
			Auth:   f.auth,
			Output: f.output,
			Config: f.config,
		},
		subcommands:  make(map[string]func(context.Context, []string) error),
		outputWriter: os.Stdout,
	}

	// Register subcommands
	cmd.subcommands["tasks"] = cmd.runExportTasks

	// Set the execution function
	cmd.RunFunc = cmd.run

	return cmd
}

// run executes the export command
func (c *ExportCommand) run(ctx context.Context, args []string) error {
	// Export command requires a subcommand
	if len(args) == 0 {
		return fmt.Errorf("no subcommand specified. Available subcommands: tasks")
	}

	subcommand := args[0]
	handler, exists := c.subcommands[subcommand]
	if !exists {
		return fmt.Errorf("unknown subcommand: %s. Available subcommands: tasks", subcommand)
	}

	// Execute subcommand with remaining args
	return handler(ctx, args[1:])
}

// runExportTasks executes the export tasks subcommand
func (c *ExportCommand) runExportTasks(ctx context.Context, args []string) error {
	// Ensure API client is available
	if c.API == nil {
		return fmt.Errorf("API client not initialized")
	}

	// Validate format
	c.format = strings.ToLower(c.format)
	if c.format != "csv" && c.format != "json" && c.format != "markdown" && c.format != "md" {
		return fmt.Errorf("invalid format: %s. Must be csv, json, or markdown", c.format)
	}
	if c.format == "md" {
		c.format = "markdown"
	}

	// Get tasks based on parameters
	tasks, err := c.getTasks(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tasks: %w", err)
	}

	// Open output file or use stdout
	var output io.Writer
	var outputCloser io.Closer

	if c.outputFile != "" {
		// Sanitize the file path to prevent directory traversal
		cleanPath := filepath.Clean(c.outputFile)
		if filepath.IsAbs(cleanPath) || strings.Contains(cleanPath, "..") {
			return fmt.Errorf("invalid output file path: %s", c.outputFile)
		}

		file, err := os.Create(cleanPath)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		output = file
		outputCloser = file
		defer func() { _ = outputCloser.Close() }()
	} else {
		output = c.outputWriter
	}

	// Export based on format
	switch c.format {
	case "csv":
		err = c.exportTasksToCSV(output, tasks)
	case "json":
		err = c.exportTasksToJSON(output, tasks)
	case "markdown":
		err = c.exportTasksToMarkdown(output, tasks)
	}

	if err != nil {
		return fmt.Errorf("failed to export tasks: %w", err)
	}

	if c.outputFile != "" {
		c.Output.PrintSuccess(fmt.Sprintf("Exported %d task(s) to %s", len(tasks), c.outputFile))
	}

	return nil
}

// getTasks retrieves tasks based on export parameters
func (c *ExportCommand) getTasks(ctx context.Context) ([]clickup.Task, error) {
	var tasks []clickup.Task

	if c.listID != "" {
		// Get tasks from specific list
		queryOpts := &interfaces.TaskQueryOptions{}
		if c.status != "" {
			queryOpts.Statuses = []string{c.status}
		}
		if c.assignee != "" {
			queryOpts.Assignees = []string{c.assignee}
		}
		if c.priority != "" {
			p, err := c.parsePriority(c.priority)
			if err != nil {
				return nil, err
			}
			queryOpts.Priority = &p
		}

		var err error
		tasks, err = c.API.GetTasks(ctx, c.listID, queryOpts)
		if err != nil {
			return nil, err
		}
	} else {
		// Get all tasks from workspace or space
		workspaces, err := c.API.GetWorkspaces(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get workspaces: %w", err)
		}

		for _, workspace := range workspaces {
			spaces, err := c.API.GetSpaces(ctx, workspace.ID)
			if err != nil {
				continue
			}

			for _, space := range spaces {
				if c.spaceID != "" && space.ID != c.spaceID && space.Name != c.spaceID {
					continue
				}

				// Get tasks from all lists in space
				folders, _ := c.API.GetFolders(ctx, space.ID)
				for _, folder := range folders {
					lists, _ := c.API.GetLists(ctx, folder.ID)
					for _, list := range lists {
						listTasks, err := c.API.GetTasks(ctx, list.ID, &interfaces.TaskQueryOptions{})
						if err == nil {
							tasks = append(tasks, listTasks...)
						}
					}
				}

				// Get folderless lists
				lists, _ := c.API.GetFolderlessLists(ctx, space.ID)
				for _, list := range lists {
					listTasks, err := c.API.GetTasks(ctx, list.ID, &interfaces.TaskQueryOptions{})
					if err == nil {
						tasks = append(tasks, listTasks...)
					}
				}
			}
		}

		// Client-side filtering
		tasks = c.filterTasks(tasks)
	}

	return tasks, nil
}

// filterTasks applies client-side filtering
func (c *ExportCommand) filterTasks(tasks []clickup.Task) []clickup.Task {
	var filtered []clickup.Task

	for _, task := range tasks {
		// Filter by status
		if c.status != "" && task.Status.Status != c.status {
			continue
		}

		// Filter by priority
		if c.priority != "" {
			taskPriority := c.getTaskPriority(task)
			if taskPriority != c.priority {
				continue
			}
		}

		// Filter by assignee
		if c.assignee != "" {
			hasAssignee := false
			for _, a := range task.Assignees {
				if a.Username == c.assignee || fmt.Sprint(a.ID) == c.assignee {
					hasAssignee = true
					break
				}
			}
			if !hasAssignee {
				continue
			}
		}

		filtered = append(filtered, task)
	}

	return filtered
}

// parsePriority converts priority string to int
func (c *ExportCommand) parsePriority(priority string) (int, error) {
	switch priority {
	case "urgent":
		return 1, nil
	case "high":
		return 2, nil
	case "normal":
		return 3, nil
	case "low":
		return 4, nil
	default:
		return 0, fmt.Errorf("invalid priority: %s", priority)
	}
}

// getTaskPriority returns the task priority as a string
func (c *ExportCommand) getTaskPriority(task clickup.Task) string {
	// Priority is a struct with Priority field containing the text
	if task.Priority.Priority == "" {
		return ""
	}

	// The Priority field contains text like "urgent", "high", etc.
	return strings.ToLower(task.Priority.Priority)
}

// getTaskDueDate returns the task due date as a string
func (c *ExportCommand) getTaskDueDate(task clickup.Task) string {
	if task.DueDate == nil {
		return ""
	}

	// Convert millisecond timestamp to time
	if task.DueDate.Time().IsZero() {
		return ""
	}

	return task.DueDate.Time().Format(time.RFC3339)
}

// formatTimestamp formats a timestamp for display
func (c *ExportCommand) formatTimestamp(ms string) string {
	// Convert millisecond timestamp to readable format
	// ClickUp timestamps are in milliseconds
	if ms == "" {
		return ""
	}

	// The ClickUp API might return timestamps in different formats
	// For now, just return the raw value
	return ms
}

// exportTasksToCSV exports tasks to CSV format
func (c *ExportCommand) exportTasksToCSV(output io.Writer, tasks []clickup.Task) error {
	writer := csv.NewWriter(output)
	defer writer.Flush()

	// Write header
	header := []string{"ID", "Name", "Status", "Priority", "Assignees", "Due Date", "Created", "Updated", "URL"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write tasks
	for _, task := range tasks {
		assignees := make([]string, 0, len(task.Assignees))
		for _, a := range task.Assignees {
			assignees = append(assignees, a.Username)
		}

		row := []string{
			task.ID,
			task.Name,
			task.Status.Status,
			c.getTaskPriority(task),
			strings.Join(assignees, ", "),
			c.getTaskDueDate(task),
			c.formatTimestamp(task.DateCreated),
			c.formatTimestamp(task.DateUpdated),
			task.URL,
		}

		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// exportTasksToJSON exports tasks to JSON format
func (c *ExportCommand) exportTasksToJSON(output io.Writer, tasks []clickup.Task) error {
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(tasks)
}

// exportTasksToMarkdown exports tasks to Markdown format
func (c *ExportCommand) exportTasksToMarkdown(output io.Writer, tasks []clickup.Task) error {
	// Group tasks by status
	tasksByStatus := make(map[string][]clickup.Task)
	for _, task := range tasks {
		status := task.Status.Status
		tasksByStatus[status] = append(tasksByStatus[status], task)
	}

	// Write markdown
	fmt.Fprintf(output, "# Task Report\n\n")
	fmt.Fprintf(output, "Generated: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(output, "Total tasks: %d\n\n", len(tasks))

	// Write summary
	fmt.Fprintf(output, "## Summary by Status\n\n")
	for status, statusTasks := range tasksByStatus {
		fmt.Fprintf(output, "- **%s**: %d tasks\n", status, len(statusTasks))
	}
	fmt.Fprintln(output)

	// Write tasks by status
	for status, statusTasks := range tasksByStatus {
		// Simple title case - capitalize first letter
		titleStatus := status
		if len(status) > 0 {
			titleStatus = strings.ToUpper(string(status[0])) + status[1:]
		}
		fmt.Fprintf(output, "## %s (%d)\n\n", titleStatus, len(statusTasks))

		for _, task := range statusTasks {
			// Task header
			fmt.Fprintf(output, "### %s\n", task.Name)
			fmt.Fprintf(output, "- **ID**: %s\n", task.ID)
			fmt.Fprintf(output, "- **Priority**: %s\n", c.getTaskPriority(task))

			// Assignees
			if len(task.Assignees) > 0 {
				assignees := make([]string, 0, len(task.Assignees))
				for _, a := range task.Assignees {
					assignees = append(assignees, a.Username)
				}
				fmt.Fprintf(output, "- **Assignees**: %s\n", strings.Join(assignees, ", "))
			}

			// Due date
			if due := c.getTaskDueDate(task); due != "" {
				fmt.Fprintf(output, "- **Due**: %s\n", due)
			}

			// Description
			if task.Description != "" {
				fmt.Fprintf(output, "\n%s\n", task.Description)
			}

			// Link
			if task.URL != "" {
				fmt.Fprintf(output, "\n[View in ClickUp](%s)\n", task.URL)
			}

			fmt.Fprintln(output)
		}
	}

	return nil
}

// GetCobraCommand returns the cobra command with subcommands
func (c *ExportCommand) GetCobraCommand() *cobra.Command {
	cmd := c.Command.GetCobraCommand()

	// Add tasks subcommand
	tasksCmd := &cobra.Command{
		Use:   "tasks",
		Short: "Export tasks to file",
		Long: `Export tasks to CSV, JSON, or Markdown format.

Examples:
  # Export all tasks from a list to CSV
  cu export tasks --list mylist --format csv --output tasks.csv
  
  # Export tasks with specific status to JSON
  cu export tasks --list mylist --status open --format json > open-tasks.json
  
  # Generate a Markdown report of high priority tasks
  cu export tasks --priority high --format markdown --output report.md`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Set flags from cobra command
			c.listID, _ = cmd.Flags().GetString("list")
			c.spaceID, _ = cmd.Flags().GetString("space")
			c.format, _ = cmd.Flags().GetString("format")
			c.outputFile, _ = cmd.Flags().GetString("output")
			c.status, _ = cmd.Flags().GetString("status")
			c.priority, _ = cmd.Flags().GetString("priority")
			c.assignee, _ = cmd.Flags().GetString("assignee")

			return c.runExportTasks(cmd.Context(), args)
		},
	}

	// Add flags to tasks subcommand
	tasksCmd.Flags().StringP("list", "l", "", "List ID to export tasks from")
	tasksCmd.Flags().StringP("space", "s", "", "Space ID to export tasks from")
	tasksCmd.Flags().StringP("format", "f", "csv", "Export format (csv, json, markdown)")
	tasksCmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")
	tasksCmd.Flags().String("status", "", "Filter by status")
	tasksCmd.Flags().String("priority", "", "Filter by priority")
	tasksCmd.Flags().String("assignee", "", "Filter by assignee")

	cmd.AddCommand(tasksCmd)

	return cmd
}

// SetOutputWriter sets the output writer for testing
func (c *ExportCommand) SetOutputWriter(w io.Writer) {
	c.outputWriter = w
}

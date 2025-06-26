package cmd

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/raksul/go-clickup/clickup"
	"github.com/spf13/cobra"
	"github.com/tim/cu/internal/api"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export data to various formats",
	Long:  `Export ClickUp data to CSV, JSON, or Markdown formats.`,
}

var exportTasksCmd = &cobra.Command{
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
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		// Get flags
		listID, _ := cmd.Flags().GetString("list")
		spaceID, _ := cmd.Flags().GetString("space")
		format, _ := cmd.Flags().GetString("format")
		outputFile, _ := cmd.Flags().GetString("output")
		status, _ := cmd.Flags().GetString("status")
		priority, _ := cmd.Flags().GetString("priority")
		assignee, _ := cmd.Flags().GetString("assignee")

		// Validate format
		format = strings.ToLower(format)
		if format != "csv" && format != "json" && format != "markdown" && format != "md" {
			fmt.Fprintf(os.Stderr, "Invalid format: %s. Must be csv, json, or markdown\n", format)
			os.Exit(1)
		}
		if format == "md" {
			format = "markdown"
		}

		// Create API client
		client, err := api.NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create API client: %v\n", err)
			os.Exit(1)
		}

		// Get tasks based on parameters
		var tasks []clickup.Task

		if listID != "" {
			// Get tasks from specific list
			queryOpts := &api.TaskQueryOptions{}
			if status != "" {
				queryOpts.Statuses = []string{status}
			}
			if assignee != "" {
				queryOpts.Assignees = []string{assignee}
			}
			if priority != "" {
				switch priority {
				case "urgent":
					p := 1
					queryOpts.Priority = &p
				case "high":
					p := 2
					queryOpts.Priority = &p
				case "normal":
					p := 3
					queryOpts.Priority = &p
				case "low":
					p := 4
					queryOpts.Priority = &p
				}
			}

			tasks, err = client.GetTasks(ctx, listID, queryOpts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to get tasks: %v\n", err)
				os.Exit(1)
			}
		} else {
			// Get all tasks from workspace or space
			workspaces, err := client.GetWorkspaces(ctx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to get workspaces: %v\n", err)
				os.Exit(1)
			}

			for _, workspace := range workspaces {
				spaces, err := client.GetSpaces(ctx, workspace.ID)
				if err != nil {
					continue
				}

				for _, space := range spaces {
					if spaceID != "" && space.ID != spaceID && space.Name != spaceID {
						continue
					}

					// Get tasks from all lists in space
					folders, _ := client.GetFolders(ctx, space.ID)
					for _, folder := range folders {
						lists, _ := client.GetLists(ctx, folder.ID)
						for _, list := range lists {
							listTasks, err := client.GetTasks(ctx, list.ID, &api.TaskQueryOptions{})
							if err == nil {
								tasks = append(tasks, listTasks...)
							}
						}
					}

					// Get folderless lists
					lists, _ := client.GetFolderlessLists(ctx, space.ID)
					for _, list := range lists {
						listTasks, err := client.GetTasks(ctx, list.ID, &api.TaskQueryOptions{})
						if err == nil {
							tasks = append(tasks, listTasks...)
						}
					}
				}
			}

			// Client-side filtering
			tasks = filterTasksForExport(tasks, status, priority, assignee)
		}

		// Open output file or use stdout
		var output *os.File
		if outputFile != "" {
			file, err := os.Create(outputFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to create output file: %v\n", err)
				os.Exit(1)
			}
			defer file.Close()
			output = file
		} else {
			output = os.Stdout
		}

		// Export based on format
		switch format {
		case "csv":
			err = exportTasksToCSV(output, tasks)
		case "json":
			err = exportTasksToJSON(output, tasks)
		case "markdown":
			err = exportTasksToMarkdown(output, tasks)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to export tasks: %v\n", err)
			os.Exit(1)
		}

		if outputFile != "" {
			fmt.Printf("✓ Exported %d task(s) to %s\n", len(tasks), outputFile)
		}
	},
}

func filterTasksForExport(tasks []clickup.Task, status, priority, assignee string) []clickup.Task {
	var filtered []clickup.Task

	for _, task := range tasks {
		// Filter by status
		if status != "" && task.Status.Status != status {
			continue
		}

		// Filter by priority
		if priority != "" {
			taskPriority := getTaskPriority(task)
			if taskPriority != priority {
				continue
			}
		}

		// Filter by assignee
		if assignee != "" {
			hasAssignee := false
			for _, a := range task.Assignees {
				if a.Username == assignee || fmt.Sprint(a.ID) == assignee {
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

func exportTasksToCSV(output *os.File, tasks []clickup.Task) error {
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
			getTaskPriority(task),
			strings.Join(assignees, ", "),
			getTaskDueDate(task),
			formatTimestamp(task.DateCreated),
			formatTimestamp(task.DateUpdated),
			task.URL,
		}

		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

func exportTasksToJSON(output *os.File, tasks []clickup.Task) error {
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(tasks)
}

func exportTasksToMarkdown(output *os.File, tasks []clickup.Task) error {
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
		fmt.Fprintf(output, "## %s (%d)\n\n", strings.Title(status), len(statusTasks))

		for _, task := range statusTasks {
			// Task header
			fmt.Fprintf(output, "### %s\n", task.Name)
			fmt.Fprintf(output, "- **ID**: %s\n", task.ID)
			fmt.Fprintf(output, "- **Priority**: %s\n", getTaskPriority(task))

			// Assignees
			if len(task.Assignees) > 0 {
				assignees := make([]string, 0, len(task.Assignees))
				for _, a := range task.Assignees {
					assignees = append(assignees, a.Username)
				}
				fmt.Fprintf(output, "- **Assignees**: %s\n", strings.Join(assignees, ", "))
			}

			// Due date
			if due := getTaskDueDate(task); due != "" {
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

func formatTimestamp(ms string) string {
	// Convert millisecond timestamp to readable format
	// ClickUp timestamps are in milliseconds
	if ms == "" {
		return ""
	}

	// The ClickUp API might return timestamps in different formats
	// For now, just return the raw value
	return ms
}

func init() {
	exportCmd.AddCommand(exportTasksCmd)

	// Export tasks flags
	exportTasksCmd.Flags().StringP("list", "l", "", "List ID to export tasks from")
	exportTasksCmd.Flags().StringP("space", "s", "", "Space ID to export tasks from")
	exportTasksCmd.Flags().StringP("format", "f", "csv", "Export format (csv, json, markdown)")
	exportTasksCmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")
	exportTasksCmd.Flags().String("status", "", "Filter by status")
	exportTasksCmd.Flags().String("priority", "", "Filter by priority")
	exportTasksCmd.Flags().String("assignee", "", "Filter by assignee")
}

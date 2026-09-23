package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/timimsms/cu/internal/api"
	"github.com/timimsms/cu/internal/output"
)

// bulkOutcome is the per-task result of a bulk operation, and bulkSummary the
// whole run. Progress lines are for a human watching; a scripted caller needs
// to know which ids failed and why, which the progress output cannot express.
type bulkOutcome struct {
	TaskID string `json:"task_id"`
	OK     bool   `json:"ok"`
	Error  string `json:"error,omitempty"`
}

type bulkSummary struct {
	Operation string        `json:"operation"`
	Succeeded int           `json:"succeeded"`
	Failed    int           `json:"failed"`
	Results   []bulkOutcome `json:"results"`
}

// record appends one task's outcome and prints the human progress line. The
// line is suppressed outside table output, where it would otherwise interleave
// with the structured document on stdout and make it unparseable.
func (b *bulkSummary) record(taskID string, err error) {
	if err != nil {
		b.Failed++
		b.Results = append(b.Results, bulkOutcome{TaskID: taskID, OK: false, Error: err.Error()})
		human("  ✗ %s: %v", taskID, err)
		return
	}
	b.Succeeded++
	b.Results = append(b.Results, bulkOutcome{TaskID: taskID, OK: true})
	human("  ✓ %s", taskID)
}

// finish emits the summary in whichever format was requested and exits
// non-zero if any task failed.
func (b bulkSummary) finish() {
	if outputFormat != "table" {
		if err := output.Format(outputFormat, b); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("\nSummary:\n")
		fmt.Printf("  Success: %d\n", b.Succeeded)
		fmt.Printf("  Failed:  %d\n", b.Failed)
	}

	if b.Failed > 0 {
		os.Exit(1)
	}
}

// human prints progress intended for a person watching, and only then.
func human(format string, a ...interface{}) {
	if outputFormat != "table" {
		return
	}
	fmt.Printf(format+"\n", a...)
}

var bulkCmd = &cobra.Command{
	Use:   "bulk",
	Short: "Perform bulk operations on tasks",
	Long:  `Perform bulk operations on multiple tasks at once.`,
}

var bulkUpdateCmd = &cobra.Command{
	Use:   "update [task-ids...]",
	Short: "Update multiple tasks",
	Long: `Update multiple tasks at once. Task IDs can be provided as arguments or from stdin.

Examples:
  # Update status for multiple tasks
  cu bulk update task1 task2 task3 --status done
  
  # Update priority from a file
  cat task-ids.txt | cu bulk update --priority high
  
  # Add assignee to multiple tasks
  cu bulk update task1 task2 --add-assignee @john`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		// Get task IDs from args or stdin
		taskIDs := args
		if len(taskIDs) == 0 {
			// Read from stdin
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line != "" {
					taskIDs = append(taskIDs, line)
				}
			}
			if err := scanner.Err(); err != nil {
				fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
				os.Exit(1)
			}
		}

		if len(taskIDs) == 0 {
			fmt.Fprintln(os.Stderr, "No task IDs provided")
			os.Exit(1)
		}

		// Get update options from flags
		status, _ := cmd.Flags().GetString("status")
		priority, _ := cmd.Flags().GetString("priority")
		addAssignees, _ := cmd.Flags().GetStringSlice("add-assignee")
		removeAssignees, _ := cmd.Flags().GetStringSlice("remove-assignee")
		tags, _ := cmd.Flags().GetStringSlice("tag")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		// Build update options
		updateOpts := &api.TaskUpdateOptions{
			Status:          status,
			Priority:        priority,
			Tags:            tags,
			AddAssignees:    addAssignees,
			RemoveAssignees: removeAssignees,
		}

		// Check if any updates were specified
		if !updateOpts.HasUpdates() {
			fmt.Fprintln(os.Stderr, "No updates specified. Use flags like --status, --priority, etc.")
			os.Exit(1)
		}

		// Show what will be updated
		fmt.Printf("Updating %d task(s):\n", len(taskIDs))
		if status != "" {
			fmt.Printf("  Status: %s\n", status)
		}
		if priority != "" {
			fmt.Printf("  Priority: %s\n", priority)
		}
		if len(tags) > 0 {
			fmt.Printf("  Tags: %s\n", strings.Join(tags, ", "))
		}
		if len(addAssignees) > 0 {
			fmt.Printf("  Add assignees: %s\n", strings.Join(addAssignees, ", "))
		}
		if len(removeAssignees) > 0 {
			fmt.Printf("  Remove assignees: %s\n", strings.Join(removeAssignees, ", "))
		}

		if dryRun {
			fmt.Println("\nDry run - no changes will be made")
			fmt.Printf("Would update tasks: %s\n", strings.Join(taskIDs, ", "))
			return
		}

		// Confirmation prompt unless --yes flag is set
		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("\nAre you sure you want to update %d task(s)? [y/N] ", len(taskIDs))
			var response string
			_, _ = fmt.Scanln(&response)
			if strings.ToLower(response) != "y" {
				fmt.Println("Cancelled")
				return
			}
		}

		// Create API client
		client, err := api.NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create API client: %v\n", err)
			os.Exit(1)
		}

		// Update tasks
		summary := bulkSummary{Operation: "update"}

		human("\nUpdating tasks...")
		for _, taskID := range taskIDs {
			_, err := client.UpdateTask(ctx, taskID, updateOpts)
			summary.record(taskID, err)
		}

		summary.finish()
	},
}

var bulkCloseCmd = &cobra.Command{
	Use:   "close [task-ids...]",
	Short: "Close multiple tasks",
	Long: `Close multiple tasks at once by marking them as complete.

Examples:
  # Close multiple tasks
  cu bulk close task1 task2 task3
  
  # Close tasks from a file
  cat completed-tasks.txt | cu bulk close`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		// Get task IDs from args or stdin
		taskIDs := args
		if len(taskIDs) == 0 {
			// Read from stdin
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line != "" {
					taskIDs = append(taskIDs, line)
				}
			}
			if err := scanner.Err(); err != nil {
				fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
				os.Exit(1)
			}
		}

		if len(taskIDs) == 0 {
			fmt.Fprintln(os.Stderr, "No task IDs provided")
			os.Exit(1)
		}

		// Confirmation prompt unless --yes flag is set
		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("Are you sure you want to close %d task(s)? [y/N] ", len(taskIDs))
			var response string
			_, _ = fmt.Scanln(&response)
			if strings.ToLower(response) != "y" {
				fmt.Println("Cancelled")
				return
			}
		}

		// Create API client
		client, err := api.NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create API client: %v\n", err)
			os.Exit(1)
		}

		// Close tasks. The done status is resolved per task from its own list
		// — a bulk close can span lists with different status sets — and the
		// resolver caches per list so one list costs one lookup.
		status, _ := cmd.Flags().GetString("status")
		resolver := api.NewStatusResolver(client)

		summary := bulkSummary{Operation: "close"}

		human("Closing tasks...")
		for _, taskID := range taskIDs {
			taskStatus := status
			if taskStatus == "" {
				resolved, err := resolver.ClosedStatusForTask(ctx, taskID)
				if err != nil {
					summary.record(taskID, err)
					continue
				}
				taskStatus = resolved
			}

			_, err := client.UpdateTask(ctx, taskID, &api.TaskUpdateOptions{Status: taskStatus})
			summary.record(taskID, err)
		}

		summary.finish()
	},
}

var bulkDeleteCmd = &cobra.Command{
	Use:   "delete [task-ids...]",
	Short: "Delete multiple tasks",
	Long: `Delete multiple tasks at once. This action cannot be undone.

Examples:
  # Delete multiple tasks
  cu bulk delete task1 task2 task3
  
  # Delete tasks from a file
  cat obsolete-tasks.txt | cu bulk delete --yes`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		// Get task IDs from args or stdin
		taskIDs := args
		if len(taskIDs) == 0 {
			// Read from stdin
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line != "" {
					taskIDs = append(taskIDs, line)
				}
			}
			if err := scanner.Err(); err != nil {
				fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
				os.Exit(1)
			}
		}

		if len(taskIDs) == 0 {
			fmt.Fprintln(os.Stderr, "No task IDs provided")
			os.Exit(1)
		}

		// Strong confirmation for delete
		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			fmt.Printf("⚠️  WARNING: This will permanently delete %d task(s).\n", len(taskIDs))
			fmt.Printf("Are you absolutely sure? Type 'delete' to confirm: ")
			var response string
			_, _ = fmt.Scanln(&response)
			if response != "delete" {
				fmt.Println("Cancelled")
				return
			}
		}

		// Create API client
		client, err := api.NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create API client: %v\n", err)
			os.Exit(1)
		}

		// Delete tasks
		summary := bulkSummary{Operation: "delete"}

		human("Deleting tasks...")
		for _, taskID := range taskIDs {
			summary.record(taskID, client.DeleteTask(ctx, taskID))
		}

		summary.finish()
	},
}

func init() {
	bulkCmd.AddCommand(bulkUpdateCmd)
	bulkCmd.AddCommand(bulkCloseCmd)
	bulkCmd.AddCommand(bulkDeleteCmd)

	// Bulk update flags
	bulkUpdateCmd.Flags().StringP("status", "s", "", "New task status")
	bulkUpdateCmd.Flags().StringP("priority", "p", "", "New task priority (urgent, high, normal, low)")
	bulkUpdateCmd.Flags().StringSlice("tag", []string{}, "Replace tags with these tags")
	bulkUpdateCmd.Flags().StringSlice("add-assignee", []string{}, "Add assignees (username or ID)")
	bulkUpdateCmd.Flags().StringSlice("remove-assignee", []string{}, "Remove assignees (username or ID)")
	bulkUpdateCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	bulkUpdateCmd.Flags().Bool("dry-run", false, "Show what would be updated without making changes")

	// Bulk close flags
	bulkCloseCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	bulkCloseCmd.Flags().StringP("status", "s", "", "Status to set (default: each list's done status)")

	// Bulk delete flags
	bulkDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
}

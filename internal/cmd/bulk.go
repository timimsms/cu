package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tim/cu/internal/api"
	"github.com/tim/cu/internal/output"
)

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
		var successCount, errorCount int

		fmt.Println("\nUpdating tasks...")
		for _, taskID := range taskIDs {
			_, err := client.UpdateTask(ctx, taskID, updateOpts)
			if err != nil {
				errorCount++
				fmt.Printf("  ✗ %s: %v\n", taskID, err)
			} else {
				successCount++
				fmt.Printf("  ✓ %s\n", taskID)
			}
		}

		// Summary
		fmt.Printf("\nSummary:\n")
		fmt.Printf("  Success: %d\n", successCount)
		fmt.Printf("  Failed:  %d\n", errorCount)

		if errorCount > 0 {
			os.Exit(1)
		}
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

		// Close tasks
		updateOpts := &api.TaskUpdateOptions{
			Status: "complete",
		}

		var successCount, errorCount int

		fmt.Println("Closing tasks...")
		for _, taskID := range taskIDs {
			_, err := client.UpdateTask(ctx, taskID, updateOpts)
			if err != nil {
				errorCount++
				fmt.Printf("  ✗ %s: %v\n", taskID, err)
			} else {
				successCount++
				fmt.Printf("  ✓ %s\n", taskID)
			}
		}

		// Summary
		fmt.Printf("\nSummary:\n")
		fmt.Printf("  Success: %d\n", successCount)
		fmt.Printf("  Failed:  %d\n", errorCount)

		if errorCount > 0 {
			os.Exit(1)
		}
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
		var successCount, errorCount int
		var deletedTasks []string

		fmt.Println("Deleting tasks...")
		for _, taskID := range taskIDs {
			err := client.DeleteTask(ctx, taskID)
			if err != nil {
				errorCount++
				fmt.Printf("  ✗ %s: %v\n", taskID, err)
			} else {
				successCount++
				deletedTasks = append(deletedTasks, taskID)
				fmt.Printf("  ✓ %s\n", taskID)
			}
		}

		// Summary
		fmt.Printf("\nSummary:\n")
		fmt.Printf("  Deleted: %d\n", successCount)
		fmt.Printf("  Failed:  %d\n", errorCount)

		// Output deleted task IDs for potential recovery scripts
		format := cmd.Flag("output").Value.String()
		if format != "table" && len(deletedTasks) > 0 {
			if err := output.Format(format, deletedTasks); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
			}
		}

		if errorCount > 0 {
			os.Exit(1)
		}
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

	// Bulk delete flags
	bulkDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
}

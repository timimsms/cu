package factory

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tim/cu/internal/cmd/base"
	"github.com/tim/cu/internal/interfaces"
)

// BulkCommand implements the bulk command with dependency injection
type BulkCommand struct {
	*base.Command
	subcommands map[string]func(context.Context, []string) error

	// Input/output dependencies for testing
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer

	// Flags
	status          string
	priority        string
	tags            []string
	addAssignees    []string
	removeAssignees []string
	yes             bool
	dryRun          bool
}

// createBulkCommand creates a new bulk command
func (f *Factory) createBulkCommand() interfaces.Command {
	cmd := &BulkCommand{
		Command: &base.Command{
			Use:    "bulk",
			Short:  "Perform bulk operations on tasks",
			Long:   `Perform bulk operations on multiple tasks at once.`,
			API:    f.api,
			Auth:   f.auth,
			Output: f.output,
			Config: f.config,
		},
		subcommands: make(map[string]func(context.Context, []string) error),
		stdin:       os.Stdin,
		stdout:      os.Stdout,
		stderr:      os.Stderr,
	}

	// Register subcommands
	cmd.subcommands["update"] = cmd.runUpdate
	cmd.subcommands["close"] = cmd.runClose
	cmd.subcommands["delete"] = cmd.runDelete

	// Set the execution function
	cmd.Command.RunFunc = cmd.run

	return cmd
}

// run executes the bulk command
func (c *BulkCommand) run(ctx context.Context, args []string) error {
	// Bulk command requires a subcommand
	if len(args) == 0 {
		return fmt.Errorf("no subcommand specified. Available subcommands: update, close, delete")
	}

	subcommand := args[0]
	handler, exists := c.subcommands[subcommand]
	if !exists {
		return fmt.Errorf("unknown subcommand: %s. Available subcommands: update, close, delete", subcommand)
	}

	// Execute subcommand with remaining args
	return handler(ctx, args[1:])
}

// runUpdate executes the bulk update subcommand
func (c *BulkCommand) runUpdate(ctx context.Context, args []string) error {
	// Ensure API client is available
	if c.API == nil {
		return fmt.Errorf("API client not initialized")
	}

	// Get task IDs from args or stdin
	taskIDs, err := c.getTaskIDs(args)
	if err != nil {
		return err
	}

	if len(taskIDs) == 0 {
		return fmt.Errorf("no task IDs provided")
	}

	// Build update options
	updateOpts := &interfaces.TaskUpdateOptions{
		Status:          c.status,
		Priority:        c.priority,
		Tags:            c.tags,
		AddAssignees:    c.addAssignees,
		RemoveAssignees: c.removeAssignees,
	}

	// Check if any updates were specified
	if !c.hasUpdates(updateOpts) {
		return fmt.Errorf("no updates specified. Use flags like --status, --priority, etc.")
	}

	// Show what will be updated
	c.Output.PrintInfo(fmt.Sprintf("Updating %d task(s):", len(taskIDs)))
	if c.status != "" {
		c.Output.PrintInfo(fmt.Sprintf("  Status: %s", c.status))
	}
	if c.priority != "" {
		c.Output.PrintInfo(fmt.Sprintf("  Priority: %s", c.priority))
	}
	if len(c.tags) > 0 {
		c.Output.PrintInfo(fmt.Sprintf("  Tags: %s", strings.Join(c.tags, ", ")))
	}
	if len(c.addAssignees) > 0 {
		c.Output.PrintInfo(fmt.Sprintf("  Add assignees: %s", strings.Join(c.addAssignees, ", ")))
	}
	if len(c.removeAssignees) > 0 {
		c.Output.PrintInfo(fmt.Sprintf("  Remove assignees: %s", strings.Join(c.removeAssignees, ", ")))
	}

	if c.dryRun {
		fmt.Fprintln(c.stdout)
		c.Output.PrintInfo("Dry run - no changes will be made")
		c.Output.PrintInfo(fmt.Sprintf("Would update tasks: %s", strings.Join(taskIDs, ", ")))
		return nil
	}

	// Confirmation prompt unless --yes flag is set
	if !c.yes {
		confirmed, err := c.confirmAction(fmt.Sprintf("update %d task(s)", len(taskIDs)))
		if err != nil {
			return err
		}
		if !confirmed {
			c.Output.PrintInfo("Cancelled")
			return nil
		}
	}

	// Update tasks
	var successCount, errorCount int

	fmt.Fprintln(c.stdout)
	c.Output.PrintInfo("Updating tasks...")
	for _, taskID := range taskIDs {
		_, err := c.API.UpdateTask(ctx, taskID, updateOpts)
		if err != nil {
			errorCount++
			c.Output.PrintError(fmt.Errorf("%s: %v", taskID, err))
		} else {
			successCount++
			c.Output.PrintSuccess(taskID)
		}
	}

	// Summary
	fmt.Fprintln(c.stdout)
	c.Output.PrintInfo("Summary:")
	c.Output.PrintInfo(fmt.Sprintf("  Success: %d", successCount))
	c.Output.PrintInfo(fmt.Sprintf("  Failed:  %d", errorCount))

	if errorCount > 0 {
		return fmt.Errorf("failed to update %d task(s)", errorCount)
	}

	return nil
}

// runClose executes the bulk close subcommand
func (c *BulkCommand) runClose(ctx context.Context, args []string) error {
	// Ensure API client is available
	if c.API == nil {
		return fmt.Errorf("API client not initialized")
	}

	// Get task IDs from args or stdin
	taskIDs, err := c.getTaskIDs(args)
	if err != nil {
		return err
	}

	if len(taskIDs) == 0 {
		return fmt.Errorf("no task IDs provided")
	}

	// Confirmation prompt unless --yes flag is set
	if !c.yes {
		confirmed, err := c.confirmAction(fmt.Sprintf("close %d task(s)", len(taskIDs)))
		if err != nil {
			return err
		}
		if !confirmed {
			c.Output.PrintInfo("Cancelled")
			return nil
		}
	}

	// Close tasks
	updateOpts := &interfaces.TaskUpdateOptions{
		Status: "complete",
	}

	var successCount, errorCount int

	c.Output.PrintInfo("Closing tasks...")
	for _, taskID := range taskIDs {
		_, err := c.API.UpdateTask(ctx, taskID, updateOpts)
		if err != nil {
			errorCount++
			c.Output.PrintError(fmt.Errorf("%s: %v", taskID, err))
		} else {
			successCount++
			c.Output.PrintSuccess(taskID)
		}
	}

	// Summary
	fmt.Fprintln(c.stdout)
	c.Output.PrintInfo("Summary:")
	c.Output.PrintInfo(fmt.Sprintf("  Success: %d", successCount))
	c.Output.PrintInfo(fmt.Sprintf("  Failed:  %d", errorCount))

	if errorCount > 0 {
		return fmt.Errorf("failed to close %d task(s)", errorCount)
	}

	return nil
}

// runDelete executes the bulk delete subcommand
func (c *BulkCommand) runDelete(ctx context.Context, args []string) error {
	// Ensure API client is available
	if c.API == nil {
		return fmt.Errorf("API client not initialized")
	}

	// Get task IDs from args or stdin
	taskIDs, err := c.getTaskIDs(args)
	if err != nil {
		return err
	}

	if len(taskIDs) == 0 {
		return fmt.Errorf("no task IDs provided")
	}

	// Strong confirmation for delete
	if !c.yes {
		c.Output.PrintWarning(fmt.Sprintf("WARNING: This will permanently delete %d task(s).", len(taskIDs)))
		fmt.Fprint(c.stdout, "Are you absolutely sure? Type 'delete' to confirm: ")

		reader := bufio.NewReader(c.stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}

		if strings.TrimSpace(response) != "delete" {
			c.Output.PrintInfo("Cancelled")
			return nil
		}
	}

	// Delete tasks
	var successCount, errorCount int
	var deletedTasks []string

	c.Output.PrintInfo("Deleting tasks...")
	for _, taskID := range taskIDs {
		err := c.API.DeleteTask(ctx, taskID)
		if err != nil {
			errorCount++
			c.Output.PrintError(fmt.Errorf("%s: %v", taskID, err))
		} else {
			successCount++
			deletedTasks = append(deletedTasks, taskID)
			c.Output.PrintSuccess(taskID)
		}
	}

	// Summary
	fmt.Fprintln(c.stdout)
	c.Output.PrintInfo("Summary:")
	c.Output.PrintInfo(fmt.Sprintf("  Deleted: %d", successCount))
	c.Output.PrintInfo(fmt.Sprintf("  Failed:  %d", errorCount))

	// Output deleted task IDs for potential recovery scripts
	format := c.Config.GetString("output")
	if format != "table" && len(deletedTasks) > 0 {
		if err := c.Output.Print(deletedTasks); err != nil {
			c.Output.PrintWarning(fmt.Sprintf("Failed to format output: %v", err))
		}
	}

	if errorCount > 0 {
		return fmt.Errorf("failed to delete %d task(s)", errorCount)
	}

	return nil
}

// getTaskIDs gets task IDs from arguments or stdin
func (c *BulkCommand) getTaskIDs(args []string) ([]string, error) {
	taskIDs := args

	if len(taskIDs) == 0 {
		// Read from stdin
		scanner := bufio.NewScanner(c.stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				taskIDs = append(taskIDs, line)
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("error reading from stdin: %w", err)
		}
	}

	return taskIDs, nil
}

// confirmAction prompts the user for confirmation
func (c *BulkCommand) confirmAction(action string) (bool, error) {
	fmt.Fprintf(c.stdout, "Are you sure you want to %s? [y/N] ", action)

	reader := bufio.NewReader(c.stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read confirmation: %w", err)
	}

	return strings.ToLower(strings.TrimSpace(response)) == "y", nil
}

// hasUpdates checks if any updates were specified
func (c *BulkCommand) hasUpdates(opts *interfaces.TaskUpdateOptions) bool {
	return opts.Status != "" ||
		opts.Priority != "" ||
		len(opts.Tags) > 0 ||
		len(opts.AddAssignees) > 0 ||
		len(opts.RemoveAssignees) > 0
}

// GetCobraCommand returns the cobra command with subcommands
func (c *BulkCommand) GetCobraCommand() *cobra.Command {
	cmd := c.Command.GetCobraCommand()

	// Add update subcommand
	updateCmd := &cobra.Command{
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
		RunE: func(cmd *cobra.Command, args []string) error {
			// Set flags from cobra command
			c.status, _ = cmd.Flags().GetString("status")
			c.priority, _ = cmd.Flags().GetString("priority")
			c.tags, _ = cmd.Flags().GetStringSlice("tag")
			c.addAssignees, _ = cmd.Flags().GetStringSlice("add-assignee")
			c.removeAssignees, _ = cmd.Flags().GetStringSlice("remove-assignee")
			c.yes, _ = cmd.Flags().GetBool("yes")
			c.dryRun, _ = cmd.Flags().GetBool("dry-run")

			return c.runUpdate(cmd.Context(), args)
		},
	}

	// Add close subcommand
	closeCmd := &cobra.Command{
		Use:   "close [task-ids...]",
		Short: "Close multiple tasks",
		Long: `Close multiple tasks at once by marking them as complete.

Examples:
  # Close multiple tasks
  cu bulk close task1 task2 task3
  
  # Close tasks from a file
  cat completed-tasks.txt | cu bulk close`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Set flags from cobra command
			c.yes, _ = cmd.Flags().GetBool("yes")

			return c.runClose(cmd.Context(), args)
		},
	}

	// Add delete subcommand
	deleteCmd := &cobra.Command{
		Use:   "delete [task-ids...]",
		Short: "Delete multiple tasks",
		Long: `Delete multiple tasks at once. This action cannot be undone.

Examples:
  # Delete multiple tasks
  cu bulk delete task1 task2 task3
  
  # Delete tasks from a file
  cat obsolete-tasks.txt | cu bulk delete --yes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Set flags from cobra command
			c.yes, _ = cmd.Flags().GetBool("yes")

			return c.runDelete(cmd.Context(), args)
		},
	}

	// Add flags to update subcommand
	updateCmd.Flags().StringP("status", "s", "", "New task status")
	updateCmd.Flags().StringP("priority", "p", "", "New task priority (urgent, high, normal, low)")
	updateCmd.Flags().StringSlice("tag", []string{}, "Replace tags with these tags")
	updateCmd.Flags().StringSlice("add-assignee", []string{}, "Add assignees (username or ID)")
	updateCmd.Flags().StringSlice("remove-assignee", []string{}, "Remove assignees (username or ID)")
	updateCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	updateCmd.Flags().Bool("dry-run", false, "Show what would be updated without making changes")

	// Add flags to close subcommand
	closeCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")

	// Add flags to delete subcommand
	deleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")

	cmd.AddCommand(updateCmd, closeCmd, deleteCmd)

	return cmd
}

// SetStdin sets the stdin for testing
func (c *BulkCommand) SetStdin(stdin io.Reader) {
	c.stdin = stdin
}

// SetStdout sets the stdout for testing
func (c *BulkCommand) SetStdout(stdout io.Writer) {
	c.stdout = stdout
}

// SetStderr sets the stderr for testing
func (c *BulkCommand) SetStderr(stderr io.Writer) {
	c.stderr = stderr
}

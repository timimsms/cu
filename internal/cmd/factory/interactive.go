package factory

import (
	"context"
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/raksul/go-clickup/clickup"
	"github.com/tim/cu/internal/cmd/base"
	"github.com/tim/cu/internal/interfaces"
)

// InteractiveCommand implements the interactive command using dependency injection
type InteractiveCommand struct {
	*base.Command
	// Allow injection of promptui for testing
	selectPrompt func(label string, items []string) (int, string, error)
	inputPrompt  func(label string) (string, error)
	confirmPrompt func(label string) (string, error)
}

// createInteractiveCommand creates a new interactive command
func (f *Factory) createInteractiveCommand() interfaces.Command {
	cmd := &InteractiveCommand{
		Command: &base.Command{
			Use:   "interactive",
			Short: "Interactive mode for task management",
			Long:  `Enter interactive mode to browse and manage tasks with a user-friendly interface.`,
			API:    f.api,
			Auth:   f.auth,
			Output: f.output,
			Config: f.config,
		},
	}

	// Set default prompt implementations
	cmd.selectPrompt = defaultSelectPrompt
	cmd.inputPrompt = defaultInputPrompt
	cmd.confirmPrompt = defaultConfirmPrompt

	// Set the execution function
	cmd.Command.RunFunc = cmd.run

	return cmd
}

// run executes the interactive command
func (c *InteractiveCommand) run(ctx context.Context, args []string) error {
	for {
		_, result, err := c.selectPrompt("What would you like to do?", []string{
			"Browse Tasks",
			"Create Task",
			"Switch Workspace",
			"Exit",
		})
		
		if err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}

		switch result {
		case "Browse Tasks":
			if err := c.runTaskBrowser(ctx); err != nil {
				c.Output.PrintError(err)
			}
		case "Create Task":
			if err := c.runCreateTask(ctx); err != nil {
				c.Output.PrintError(err)
			}
		case "Switch Workspace":
			c.Output.PrintWarning("Workspace switching not yet implemented")
		case "Exit":
			return nil
		}
	}
}

// runTaskBrowser handles the task browsing interface
func (c *InteractiveCommand) runTaskBrowser(ctx context.Context) error {
	// Get default list or error
	listID := c.Config.GetString("default_list")
	if listID == "" {
		return fmt.Errorf("no default list set. Please set one with 'cu list default' or use 'cu task list --list <id>'")
	}

	// Get tasks
	tasks, err := c.API.GetTasks(ctx, listID, &interfaces.TaskQueryOptions{})
	if err != nil {
		return fmt.Errorf("failed to get tasks: %w", err)
	}

	if len(tasks) == 0 {
		c.Output.PrintInfo("No tasks found")
		return nil
	}

	// Create task selection prompt
	taskNames := make([]string, len(tasks))
	for i, task := range tasks {
		taskNames[i] = fmt.Sprintf("%s (%s)", task.Name, task.Status.Status)
	}

	index, _, err := c.selectPrompt("Select a task", taskNames)
	if err != nil {
		if err == promptui.ErrInterrupt {
			return nil
		}
		return fmt.Errorf("task selection failed: %w", err)
	}

	selectedTask := tasks[index]
	return c.runTaskActions(ctx, selectedTask)
}

// runTaskActions handles actions for a selected task
func (c *InteractiveCommand) runTaskActions(ctx context.Context, task clickup.Task) error {
	for {
		_, action, err := c.selectPrompt(fmt.Sprintf("Task: %s", task.Name), []string{
			"View Details",
			"Update Status",
			"Update Priority",
			"Close Task",
			"Open in Browser",
			"Back",
		})

		if err != nil {
			return err
		}

		switch action {
		case "View Details":
			c.displayTaskDetails(task)
		case "Update Status":
			return c.updateTaskStatus(ctx, task)
		case "Update Priority":
			return c.updateTaskPriority(ctx, task)
		case "Close Task":
			return c.closeTask(ctx, task)
		case "Open in Browser":
			if task.URL != "" {
				c.Output.PrintInfo(fmt.Sprintf("Task URL: %s", task.URL))
			}
		case "Back":
			return nil
		}
	}
}

// displayTaskDetails shows task details
func (c *InteractiveCommand) displayTaskDetails(task clickup.Task) {
	details := fmt.Sprintf(`
=== Task Details ===
ID: %s
Name: %s
Status: %s
Priority: %s
`, task.ID, task.Name, task.Status.Status, c.getTaskPriority(task))

	if len(task.Assignees) > 0 {
		assignees := make([]string, len(task.Assignees))
		for i, assignee := range task.Assignees {
			assignees[i] = assignee.Username
		}
		details += fmt.Sprintf("Assignees: %s\n", strings.Join(assignees, ", "))
	}

	if task.Description != "" {
		details += fmt.Sprintf("\nDescription:\n%s\n", task.Description)
	}

	if task.DueDate != nil {
		details += fmt.Sprintf("Due: %s\n", task.DueDate.String())
	}

	c.Output.PrintInfo(details)
	
	// Wait for user acknowledgment
	_, _ = c.inputPrompt("Press Enter to continue...")
}

// updateTaskStatus updates a task's status
func (c *InteractiveCommand) updateTaskStatus(ctx context.Context, task clickup.Task) error {
	statuses := []string{"open", "in progress", "review", "complete", "closed"}

	_, status, err := c.selectPrompt("Select new status", statuses)
	if err != nil {
		return err
	}

	updateOpts := &interfaces.TaskUpdateOptions{
		Status: status,
	}

	updatedTask, err := c.API.UpdateTask(ctx, task.ID, updateOpts)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	c.Output.PrintSuccess(fmt.Sprintf("Updated task status to: %s", updatedTask.Status.Status))
	return nil
}

// updateTaskPriority updates a task's priority
func (c *InteractiveCommand) updateTaskPriority(ctx context.Context, task clickup.Task) error {
	priorities := []string{"urgent", "high", "normal", "low"}

	_, priority, err := c.selectPrompt("Select new priority", priorities)
	if err != nil {
		return err
	}

	updateOpts := &interfaces.TaskUpdateOptions{
		Priority: priority,
	}

	_, err = c.API.UpdateTask(ctx, task.ID, updateOpts)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	c.Output.PrintSuccess(fmt.Sprintf("Updated task priority to: %s", priority))
	return nil
}

// closeTask closes a task
func (c *InteractiveCommand) closeTask(ctx context.Context, task clickup.Task) error {
	_, err := c.confirmPrompt("Are you sure you want to close this task")
	if err != nil {
		return nil // User cancelled
	}

	updateOpts := &interfaces.TaskUpdateOptions{
		Status: "complete",
	}

	_, err = c.API.UpdateTask(ctx, task.ID, updateOpts)
	if err != nil {
		return fmt.Errorf("failed to close task: %w", err)
	}

	c.Output.PrintSuccess("Task closed successfully")
	return nil
}

// runCreateTask handles interactive task creation
func (c *InteractiveCommand) runCreateTask(ctx context.Context) error {
	// Task name
	name, err := c.inputPrompt("Task name")
	if err != nil {
		return err
	}

	// Description (optional)
	description, _ := c.inputPrompt("Description (optional)")

	// Priority
	_, priority, _ := c.selectPrompt("Priority", []string{"urgent", "high", "normal", "low"})

	// Get default list
	listID := c.Config.GetString("default_list")
	if listID == "" {
		return fmt.Errorf("no default list set. Please set one with 'cu list default'")
	}

	// Create task
	createOpts := &interfaces.TaskCreateOptions{
		Name:        name,
		Description: description,
		Priority:    priority,
	}

	task, err := c.API.CreateTask(ctx, listID, createOpts)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	c.Output.PrintSuccess(fmt.Sprintf("Created task: %s", task.Name))
	if task.URL != "" {
		c.Output.PrintInfo(fmt.Sprintf("View in ClickUp: %s", task.URL))
	}

	return nil
}

// getTaskPriority returns a readable priority string
func (c *InteractiveCommand) getTaskPriority(task clickup.Task) string {
	// TaskPriority is not a pointer, check if it's empty
	if task.Priority.Priority == "" {
		return "Normal"
	}
	
	switch task.Priority.Priority {
	case "urgent":
		return "Urgent"
	case "high":
		return "High"
	case "normal":
		return "Normal"
	case "low":
		return "Low"
	default:
		return "Normal"
	}
}

// Default prompt implementations using promptui
func defaultSelectPrompt(label string, items []string) (int, string, error) {
	prompt := promptui.Select{
		Label: label,
		Items: items,
	}
	return prompt.Run()
}

func defaultInputPrompt(label string) (string, error) {
	prompt := promptui.Prompt{
		Label: label,
	}
	return prompt.Run()
}

func defaultConfirmPrompt(label string) (string, error) {
	prompt := promptui.Prompt{
		Label:     label,
		IsConfirm: true,
	}
	return prompt.Run()
}
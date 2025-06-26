package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/raksul/go-clickup/clickup"
	"github.com/spf13/cobra"
	"github.com/tim/cu/internal/api"
	"github.com/tim/cu/internal/config"
)

var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Interactive mode for task management",
	Long:  `Enter interactive mode to browse and manage tasks with a user-friendly interface.`,
	Run: func(cmd *cobra.Command, args []string) {
		runInteractiveMode()
	},
}

var taskInteractiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Interactive task browser",
	Long:  `Browse and manage tasks interactively.`,
	Run: func(cmd *cobra.Command, args []string) {
		runTaskInteractive()
	},
}

func init() {
	taskCmd.AddCommand(taskInteractiveCmd)
}

func runInteractiveMode() {
	for {
		prompt := promptui.Select{
			Label: "What would you like to do?",
			Items: []string{
				"Browse Tasks",
				"Create Task",
				"Switch Workspace",
				"Exit",
			},
		}

		_, result, err := prompt.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Prompt failed: %v\n", err)
			return
		}

		switch result {
		case "Browse Tasks":
			runTaskInteractive()
		case "Create Task":
			runCreateTaskInteractive()
		case "Switch Workspace":
			fmt.Println("Workspace switching not yet implemented")
		case "Exit":
			return
		}
	}
}

func runTaskInteractive() {
	ctx := context.Background()

	// Create API client
	client, err := api.NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create API client: %v\n", err)
		os.Exit(1)
	}

	// Get default list or prompt for one
	listID := config.GetString("default_list")
	if listID == "" {
		fmt.Fprintln(os.Stderr, "No default list set. Please set one with 'cu list default' or use 'cu task list --list <id>'")
		return
	}

	// Get tasks
	tasks, err := client.GetTasks(ctx, listID, &api.TaskQueryOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get tasks: %v\n", err)
		return
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks found")
		return
	}

	// Create task items for prompt
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "▸ {{ .Name | cyan }} ({{ .Status.Status | yellow }})",
		Inactive: "  {{ .Name }} ({{ .Status.Status }})",
		Selected: "✓ {{ .Name | green }}",
		Details: `
--------- Task Details ----------
{{ "ID:" | faint }}      {{ .ID }}
{{ "Status:" | faint }}  {{ .Status.Status }}
{{ "Priority:" | faint }} {{ .Priority.Priority | default "Normal" }}
{{ "Assignees:" | faint }} {{ if .Assignees }}{{ range .Assignees }}{{ .Username }} {{ end }}{{ else }}Unassigned{{ end }}
{{ "Due Date:" | faint }} {{ if .DueDate }}{{ .DueDate }}{{ else }}No due date{{ end }}`,
	}

	searcher := func(input string, index int) bool {
		task := tasks[index]
		name := strings.Replace(strings.ToLower(task.Name), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)
		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:     "Select a task",
		Items:     tasks,
		Templates: templates,
		Size:      10,
		Searcher:  searcher,
	}

	index, _, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrInterrupt {
			return
		}
		fmt.Fprintf(os.Stderr, "Prompt failed: %v\n", err)
		return
	}

	selectedTask := tasks[index]
	runTaskActions(selectedTask)
}

func runTaskActions(task clickup.Task) {
	for {
		prompt := promptui.Select{
			Label: fmt.Sprintf("Task: %s", task.Name),
			Items: []string{
				"View Details",
				"Update Status",
				"Update Priority",
				"Close Task",
				"Open in Browser",
				"Back",
			},
		}

		_, action, err := prompt.Run()
		if err != nil {
			return
		}

		switch action {
		case "View Details":
			displayTaskDetails(task)
		case "Update Status":
			updateTaskStatusInteractive(task)
			return
		case "Update Priority":
			updateTaskPriorityInteractive(task)
			return
		case "Close Task":
			closeTaskInteractive(task)
			return
		case "Open in Browser":
			if task.URL != "" {
				fmt.Printf("Task URL: %s\n", task.URL)
			}
		case "Back":
			return
		}
	}
}

func displayTaskDetails(task clickup.Task) {
	fmt.Printf("\n=== Task Details ===\n")
	fmt.Printf("ID: %s\n", task.ID)
	fmt.Printf("Name: %s\n", task.Name)
	fmt.Printf("Status: %s\n", task.Status.Status)
	fmt.Printf("Priority: %s\n", getTaskPriority(task))

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
		fmt.Printf("Due: %s\n", getTaskDueDate(task))
	}

	fmt.Printf("\nPress Enter to continue...")
	_, _ = fmt.Scanln()
}

func updateTaskStatusInteractive(task clickup.Task) {
	statuses := []string{"open", "in progress", "review", "complete", "closed"}

	prompt := promptui.Select{
		Label: "Select new status",
		Items: statuses,
	}

	_, status, err := prompt.Run()
	if err != nil {
		return
	}

	ctx := context.Background()
	client, _ := api.NewClient()

	updateOpts := &api.TaskUpdateOptions{
		Status: status,
	}

	updatedTask, err := client.UpdateTask(ctx, task.ID, updateOpts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to update task: %v\n", err)
		return
	}

	fmt.Printf("✓ Updated task status to: %s\n", updatedTask.Status.Status)
}

func updateTaskPriorityInteractive(task clickup.Task) {
	priorities := []string{"urgent", "high", "normal", "low"}

	prompt := promptui.Select{
		Label: "Select new priority",
		Items: priorities,
	}

	_, priority, err := prompt.Run()
	if err != nil {
		return
	}

	ctx := context.Background()
	client, _ := api.NewClient()

	updateOpts := &api.TaskUpdateOptions{
		Priority: priority,
	}

	_, err = client.UpdateTask(ctx, task.ID, updateOpts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to update task: %v\n", err)
		return
	}

	fmt.Printf("✓ Updated task priority to: %s\n", priority)
}

func closeTaskInteractive(task clickup.Task) {
	prompt := promptui.Prompt{
		Label:     "Are you sure you want to close this task",
		IsConfirm: true,
	}

	_, err := prompt.Run()
	if err != nil {
		return
	}

	ctx := context.Background()
	client, _ := api.NewClient()

	updateOpts := &api.TaskUpdateOptions{
		Status: "complete",
	}

	_, err = client.UpdateTask(ctx, task.ID, updateOpts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to close task: %v\n", err)
		return
	}

	fmt.Println("✓ Task closed successfully")
}

func runCreateTaskInteractive() {
	// Task name
	namePrompt := promptui.Prompt{
		Label: "Task name",
	}

	name, err := namePrompt.Run()
	if err != nil {
		return
	}

	// Description (optional)
	descPrompt := promptui.Prompt{
		Label: "Description (optional)",
	}

	description, _ := descPrompt.Run()

	// Priority
	priorityPrompt := promptui.Select{
		Label: "Priority",
		Items: []string{"urgent", "high", "normal", "low"},
	}

	_, priority, _ := priorityPrompt.Run()

	// Get default list
	listID := config.GetString("default_list")
	if listID == "" {
		fmt.Fprintln(os.Stderr, "No default list set. Please set one with 'cu list default'")
		return
	}

	// Create task
	ctx := context.Background()
	client, _ := api.NewClient()

	createOpts := &api.TaskCreateOptions{
		Name:        name,
		Description: description,
		Priority:    priority,
	}

	task, err := client.CreateTask(ctx, listID, createOpts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create task: %v\n", err)
		return
	}

	fmt.Printf("✓ Created task: %s\n", task.Name)
	if task.URL != "" {
		fmt.Printf("  View in ClickUp: %s\n", task.URL)
	}
}


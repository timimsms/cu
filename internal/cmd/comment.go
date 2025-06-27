package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/raksul/go-clickup/clickup"
	"github.com/spf13/cobra"
	"github.com/tim/cu/internal/api"
	"github.com/tim/cu/internal/output"
)

var commentCmd = &cobra.Command{
	Use:   "comment <task-id>",
	Short: "Manage task comments",
	Long: `Manage comments on ClickUp tasks.

Without subcommands, adds a comment to the specified task.`,
	Args: cobra.ExactArgs(1),
	RunE: addComment,
}

var (
	commentMessage string
	commentAssignee string
	notifyAll      bool
	listComments   bool
	deleteComment  string
	resolveComment bool
	yesFlag        bool
)

func init() {
	rootCmd.AddCommand(commentCmd)

	// Flags for adding comments
	commentCmd.Flags().StringVarP(&commentMessage, "message", "m", "", "Comment text (opens editor if not provided)")
	commentCmd.Flags().StringVar(&commentAssignee, "assignee", "", "Assign comment to user")
	commentCmd.Flags().BoolVar(&notifyAll, "notify-all", false, "Notify all task watchers")
	
	// List comments flag
	commentCmd.Flags().BoolVarP(&listComments, "list", "l", false, "List all comments on the task")
	
	// Delete comment flag
	commentCmd.Flags().StringVarP(&deleteComment, "delete", "d", "", "Delete comment by ID")
	
	// Subcommands
	commentCmd.AddCommand(listCommentsCmd)
	commentCmd.AddCommand(deleteCommentCmd)
	
	// Add yes flag to delete subcommand
	deleteCommentCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Skip confirmation prompt")
}

// addComment adds a new comment to a task
func addComment(cmd *cobra.Command, args []string) error {
	taskID := args[0]
	
	// If listing comments, delegate to list function
	if listComments {
		return listTaskComments(cmd, []string{taskID})
	}
	
	// If deleting comment, delegate to delete function
	if deleteComment != "" {
		return deleteTaskComment(cmd, []string{deleteComment})
	}
	
	// Get comment text
	var text string
	if commentMessage != "" {
		text = commentMessage
	} else {
		// Prompt for comment text
		fmt.Print("Enter comment text (press Enter twice to finish):\n> ")
		scanner := bufio.NewScanner(os.Stdin)
		var lines []string
		emptyLineCount := 0
		
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				emptyLineCount++
				if emptyLineCount >= 1 {
					break
				}
			} else {
				emptyLineCount = 0
			}
			lines = append(lines, line)
			fmt.Print("> ")
		}
		
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("failed to read comment: %w", err)
		}
		
		text = strings.TrimSpace(strings.Join(lines, "\n"))
		if text == "" {
			return fmt.Errorf("comment text cannot be empty")
		}
	}
	
	// Create API client
	client, err := api.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}
	
	ctx := context.Background()
	
	// Create comment
	comment, err := client.CreateTaskComment(ctx, taskID, text, commentAssignee, notifyAll)
	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}
	
	// Display result
	if outputFormat == "json" || outputFormat == "yaml" || outputFormat == "csv" {
		if err := output.Format(outputFormat, comment); err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		return nil
	}
	
	// Human-readable output
	fmt.Printf("Comment added successfully!\n")
	fmt.Printf("ID: %d\n", comment.ID)
	if comment.Date != nil {
		fmt.Printf("Date: %s\n", comment.Date.String())
	}
	
	return nil
}

var listCommentsCmd = &cobra.Command{
	Use:   "list <task-id>",
	Short: "List all comments on a task",
	Args:  cobra.ExactArgs(1),
	RunE:  listTaskComments,
}

func listTaskComments(cmd *cobra.Command, args []string) error {
	taskID := args[0]
	
	// Create API client
	client, err := api.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}
	
	ctx := context.Background()
	
	// Get comments
	comments, err := client.GetTaskComments(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get comments: %w", err)
	}
	
	// Display results
	if outputFormat == "json" || outputFormat == "yaml" || outputFormat == "csv" {
		if err := output.Format(outputFormat, comments); err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
		return nil
	}
	
	// Table output
	var rows [][]string
	
	for _, comment := range comments {
		text := comment.CommentText
		if len(text) > 50 {
			text = text[:47] + "..."
		}
		// Replace newlines with spaces for table display
		text = strings.ReplaceAll(text, "\n", " ")
		
		resolved := ""
		if comment.Resolved {
			resolved = "✓"
		}
		
		assignee := ""
		if comment.Assignee.ID != 0 {
			assignee = getUserDisplay(comment.Assignee)
		}
		
		rows = append(rows, []string{
			fmt.Sprintf("%d", comment.ID),
			getUserDisplay(comment.User),
			formatCommentDate(comment.Date),
			text,
			resolved,
			assignee,
		})
	}
	
	// Print table
	if len(rows) > 0 {
		// Print header
		fmt.Printf("%-10s %-20s %-16s %-50s %-8s %-20s\n", "ID", "User", "Date", "Text", "Resolved", "Assignee")
		fmt.Println(strings.Repeat("-", 134))
		
		// Print rows
		for _, row := range rows {
			fmt.Printf("%-10s %-20s %-16s %-50s %-8s %-20s\n", row[0], row[1], row[2], row[3], row[4], row[5])
		}
	} else {
		fmt.Println("No comments found")
	}
	
	fmt.Printf("\nTotal comments: %d\n", len(comments))
	
	return nil
}

var deleteCommentCmd = &cobra.Command{
	Use:   "delete <comment-id>",
	Short: "Delete a comment",
	Args:  cobra.ExactArgs(1),
	RunE:  deleteTaskComment,
}

func deleteTaskComment(cmd *cobra.Command, args []string) error {
	commentID := args[0]
	
	// Confirm deletion
	if !yesFlag {
		fmt.Printf("Are you sure you want to delete comment %s? (y/N): ", commentID)
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}
		
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Deletion cancelled")
			return nil
		}
	}
	
	// Create API client
	client, err := api.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}
	
	ctx := context.Background()
	
	// Delete comment
	if err := client.DeleteTaskComment(ctx, commentID); err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}
	
	fmt.Printf("Comment %s deleted successfully\n", commentID)
	
	return nil
}

// Helper functions

func getUserDisplay(user interface{}) string {
	// Handle different user types from the API
	switch u := user.(type) {
	case clickup.User:
		if u.Username != "" {
			return u.Username
		}
		if u.Email != "" {
			return u.Email
		}
		return fmt.Sprintf("User %d", u.ID)
	case *clickup.User:
		if u != nil {
			if u.Username != "" {
				return u.Username
			}
			if u.Email != "" {
				return u.Email
			}
			return fmt.Sprintf("User %d", u.ID)
		}
	case map[string]interface{}:
		if username, ok := u["username"].(string); ok && username != "" {
			return username
		}
		if email, ok := u["email"].(string); ok && email != "" {
			return email
		}
		if id, ok := u["id"].(float64); ok {
			return fmt.Sprintf("User %.0f", id)
		}
	case string:
		return u
	}
	return "Unknown"
}

func formatCommentDate(dateStr string) string {
	// Try to parse the date
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		// Try Unix timestamp in milliseconds
		var timestamp int64
		if _, err := fmt.Sscanf(dateStr, "%d", &timestamp); err == nil && timestamp > 0 {
			t = time.Unix(timestamp/1000, 0)
		} else {
			return dateStr // Return as-is if we can't parse it
		}
	}
	
	// Format relative time
	now := time.Now()
	diff := now.Sub(t)
	
	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		return fmt.Sprintf("%d minutes ago", int(diff.Minutes()))
	case diff < 24*time.Hour:
		return fmt.Sprintf("%d hours ago", int(diff.Hours()))
	case diff < 7*24*time.Hour:
		return fmt.Sprintf("%d days ago", int(diff.Hours()/24))
	default:
		return t.Format("2006-01-02 15:04")
	}
}
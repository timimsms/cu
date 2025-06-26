package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tim/cu/internal/api"
	"github.com/tim/cu/internal/cache"
	"github.com/tim/cu/internal/output"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users",
	Long:  `View and manage workspace users.`,
}

var userListCmd = &cobra.Command{
	Use:   "list",
	Short: "List workspace users",
	Long:  `List all users in your ClickUp workspace.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		// Initialize caches if not already done
		if cache.UserCache == nil {
			if err := cache.InitCaches(); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to initialize cache: %v\n", err)
			}
		}

		// Create API client
		client, err := api.NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create API client: %v\n", err)
			os.Exit(1)
		}

		// Get workspaces first
		workspaces, err := client.GetWorkspaces(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get workspaces: %v\n", err)
			os.Exit(1)
		}

		if len(workspaces) == 0 {
			fmt.Fprintln(os.Stderr, "No workspaces found")
			os.Exit(1)
		}

		// For now, use the first workspace
		workspace := workspaces[0]

		// Load users
		err = client.UserLookup().LoadWorkspaceUsers(ctx, workspace.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load users: %v\n", err)
			os.Exit(1)
		}

		// Get all users
		users := client.UserLookup().GetAllUsers()

		// Format output
		format := cmd.Flag("output").Value.String()

		if format == "table" {
			// Prepare table data
			type userRow struct {
				ID       int    `json:"id"`
				Username string `json:"username"`
				Email    string `json:"email"`
				Role     string `json:"role"`
			}

			var rows []userRow
			for _, user := range users {
				row := userRow{
					ID:       user.ID,
					Username: user.Username,
					Email:    user.Email,
					Role:     fmt.Sprintf("%d", user.Role), // Convert role to string
				}
				rows = append(rows, row)
			}

			if err := output.Format(format, rows); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
		} else {
			// For other formats, output raw user data
			if err := output.Format(format, users); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	userCmd.AddCommand(userListCmd)
}

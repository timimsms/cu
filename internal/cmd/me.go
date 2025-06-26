package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tim/cu/internal/api"
	"github.com/tim/cu/internal/output"
)

var meCmd = &cobra.Command{
	Use:   "me",
	Short: "Show current user information",
	Long: `Display information about the currently authenticated ClickUp user,
including workspace membership and API rate limit status.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		
		// Create API client
		client, err := api.NewClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create API client: %v\n", err)
			os.Exit(1)
		}
		
		// Get current user
		user, err := client.GetCurrentUser(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get user info: %v\n", err)
			os.Exit(1)
		}
		
		// Get workspaces
		workspaces, err := client.GetWorkspaces(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get workspaces: %v\n", err)
			os.Exit(1)
		}
		
		// Prepare output data
		outputData := map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"color":    user.Color,
			"initials": user.Initials,
		}
		
		// Add workspace info
		var workspaceNames []string
		for _, ws := range workspaces {
			workspaceNames = append(workspaceNames, ws.Name)
		}
		outputData["workspaces"] = workspaceNames
		outputData["workspace_count"] = len(workspaces)
		
		// Output based on format flag
		format := cmd.Flag("output").Value.String()
		if err := output.Format(format, outputData); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
			os.Exit(1)
		}
	},
}
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

var spaceCmd = &cobra.Command{
	Use:   "space",
	Short: "Manage spaces",
	Long:  `View and manage ClickUp spaces within your workspace.`,
}

var spaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all spaces",
	Long:  `List all spaces in your ClickUp workspace.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		// Initialize caches if not already done
		if cache.WorkspaceCache == nil {
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
		// TODO: Add workspace selection
		workspace := workspaces[0]

		// Try cache first
		var spaces []interface{}
		cacheKey := fmt.Sprintf("spaces_%s", workspace.ID)
		if cache.WorkspaceCache != nil {
			if err := cache.WorkspaceCache.Get(cacheKey, &spaces); err == nil {
				// Cache hit - format and return
				format := cmd.Flag("output").Value.String()
				if err := output.Format(format, spaces); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
					os.Exit(1)
				}
				return
			}
		}

		// Cache miss - fetch from API
		apiSpaces, err := client.GetSpaces(ctx, workspace.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get spaces: %v\n", err)
			os.Exit(1)
		}

		// Cache the result
		if cache.WorkspaceCache != nil {
			_ = cache.WorkspaceCache.Set(cacheKey, apiSpaces)
		}

		// Format output
		format := cmd.Flag("output").Value.String()

		if format == "table" {
			// Prepare table data
			type spaceRow struct {
				ID       string `json:"id"`
				Name     string `json:"name"`
				Private  bool   `json:"private"`
				Archived bool   `json:"archived"`
			}

			var rows []spaceRow
			for _, space := range apiSpaces {
				row := spaceRow{
					ID:       space.ID,
					Name:     space.Name,
					Private:  space.Private,
					Archived: space.Archived,
				}
				rows = append(rows, row)
			}

			if err := output.Format(format, rows); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
		} else {
			// For other formats, output raw space data
			if err := output.Format(format, apiSpaces); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	spaceCmd.AddCommand(spaceListCmd)
}

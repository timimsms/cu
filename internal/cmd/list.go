package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/raksul/go-clickup/clickup"
	"github.com/spf13/cobra"
	"github.com/timimsms/cu/internal/api"
	"github.com/timimsms/cu/internal/cache"
	"github.com/timimsms/cu/internal/config"
	"github.com/timimsms/cu/internal/output"
)

var (
	isProjectFlag bool
)

// listRow is the table view of a clickup.List.
type listRow struct {
	ID       string `json:"id" yaml:"id"`
	Name     string `json:"name" yaml:"name"`
	Default  string `json:"default" yaml:"default"`
	Tasks    int    `json:"tasks" yaml:"tasks"`
	Archived bool   `json:"archived" yaml:"archived"`
}

// listTableRows converts lists into table rows, marking the default list with "*".
func listTableRows(lists []clickup.List, defaultListID string) []listRow {
	rows := make([]listRow, 0, len(lists))
	for _, l := range lists {
		defaultMarker := ""
		if l.ID == defaultListID {
			defaultMarker = "*"
		}
		// task_count may be absent; leave Tasks at zero in that case
		tasks, _ := l.TaskCount.Int64()
		rows = append(rows, listRow{
			ID:       l.ID,
			Name:     l.Name,
			Default:  defaultMarker,
			Tasks:    int(tasks),
			Archived: l.Archived,
		})
	}
	return rows
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Manage lists",
	Long:  `View and manage ClickUp lists.`,
}

var listListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all lists",
	Long:  `List all lists in a space or folder.`,
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

		// Get flags
		spaceID, _ := cmd.Flags().GetString("space")
		folderID, _ := cmd.Flags().GetString("folder")
		includeArchived, _ := cmd.Flags().GetBool("archived")

		// If neither space nor folder specified, show error
		if spaceID == "" && folderID == "" {
			fmt.Fprintln(os.Stderr, "Please specify either --space or --folder")
			os.Exit(1)
		}

		var allLists []clickup.List

		if folderID != "" {
			// Get lists from folder
			lists, err := client.GetLists(ctx, folderID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to get lists from folder: %v\n", err)
				os.Exit(1)
			}
			for _, list := range lists {
				if !list.Archived || includeArchived {
					allLists = append(allLists, list)
				}
			}
		} else if spaceID != "" {
			// Get folderless lists from space
			lists, err := client.GetFolderlessLists(ctx, spaceID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to get folderless lists: %v\n", err)
				os.Exit(1)
			}
			for _, list := range lists {
				if !list.Archived || includeArchived {
					allLists = append(allLists, list)
				}
			}

			// Also get lists from folders in the space
			folders, err := client.GetFolders(ctx, spaceID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to get folders: %v\n", err)
				os.Exit(1)
			}

			for _, folder := range folders {
				lists, err := client.GetLists(ctx, folder.ID)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to get lists from folder %s: %v\n", folder.Name, err)
					continue
				}
				for _, list := range lists {
					if !list.Archived || includeArchived {
						allLists = append(allLists, list)
					}
				}
			}
		}

		// Get default list ID for highlighting
		defaultListID := config.GetString("default_list")

		// Format output
		format := cmd.Flag("output").Value.String()

		if format == "table" {
			rows := listTableRows(allLists, defaultListID)
			if err := output.Format(format, rows); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
		} else {
			// For other formats, output raw list data
			if err := output.Format(format, allLists); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

var listDefaultCmd = &cobra.Command{
	Use:   "default <list-id>",
	Short: "Set default list",
	Long:  `Set the default list for task operations.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		listID := args[0]

		// TODO: Validate that the list exists and is accessible

		// Save to project config if in a project, otherwise global config
		if config.HasProjectConfig() || isProjectFlag {
			// Save to project config
			settings := map[string]interface{}{
				"default_list": listID,
			}
			if err := config.SaveProjectConfig(settings); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to save project configuration: %v\n", err)
				os.Exit(1)
			}

			configPath := config.GetProjectConfigPath()
			if configPath == "" {
				configPath = ".cu.yml"
			}
			fmt.Printf("Default list set to: %s\n", listID)
			fmt.Printf("Saved to project config: %s\n", configPath)
		} else {
			// Save to global config
			config.Set("default_list", listID)
			if err := config.Save(); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to save configuration: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Default list set to: %s (global)\n", listID)
			fmt.Println("Tip: Use --project flag to save to project-specific config")
		}
	},
}

func init() {
	listCmd.AddCommand(listListCmd)
	listCmd.AddCommand(listDefaultCmd)

	// Add --project flag to list default command
	listDefaultCmd.Flags().BoolVarP(&isProjectFlag, "project", "p", false, "Save to project config instead of global config")

	// List command flags
	listListCmd.Flags().StringP("space", "s", "", "Space ID or name")
	listListCmd.Flags().StringP("folder", "f", "", "Folder ID or name")
	listListCmd.Flags().Bool("archived", false, "Include archived lists")
}

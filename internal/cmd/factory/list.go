package factory

import (
	"context"
	"fmt"

	"github.com/raksul/go-clickup/clickup"
	"github.com/spf13/cobra"
	"github.com/tim/cu/internal/cmd/base"
	"github.com/tim/cu/internal/interfaces"
)

// ListCommand implements the list command with dependency injection
type ListCommand struct {
	*base.Command
	subcommands map[string]func(context.Context, []string) error
	
	// Flags
	spaceID         string
	folderID        string
	includeArchived bool
	isProjectFlag   bool
}

// createListCommand creates a new list command
func (f *Factory) createListCommand() interfaces.Command {
	cmd := &ListCommand{
		Command: &base.Command{
			Use:   "list",
			Short: "Manage lists",
			Long:  `View and manage ClickUp lists.`,
			API:    f.api,
			Auth:   f.auth,
			Output: f.output,
			Config: f.config,
		},
		subcommands: make(map[string]func(context.Context, []string) error),
	}

	// Register subcommands
	cmd.subcommands["list"] = cmd.runList
	cmd.subcommands["default"] = cmd.runDefault

	// Set the execution function
	cmd.Command.RunFunc = cmd.run

	return cmd
}

// run executes the list command
func (c *ListCommand) run(ctx context.Context, args []string) error {
	// If no subcommand, default to list
	if len(args) == 0 {
		return c.runList(ctx, args)
	}

	subcommand := args[0]
	handler, exists := c.subcommands[subcommand]
	if !exists {
		return fmt.Errorf("unknown subcommand: %s. Available subcommands: list, default", subcommand)
	}

	// Execute subcommand with remaining args
	return handler(ctx, args[1:])
}

// runList executes the list list subcommand
func (c *ListCommand) runList(ctx context.Context, args []string) error {
	// Ensure API client is connected
	if c.API == nil {
		return fmt.Errorf("API client not initialized")
	}

	// Validate flags
	if c.spaceID == "" && c.folderID == "" {
		return fmt.Errorf("please specify either --space or --folder")
	}

	var allLists []clickup.List

	if c.folderID != "" {
		// Get lists from folder
		lists, err := c.API.GetLists(ctx, c.folderID)
		if err != nil {
			return fmt.Errorf("failed to get lists from folder: %w", err)
		}
		for _, list := range lists {
			if !list.Archived || c.includeArchived {
				allLists = append(allLists, list)
			}
		}
	} else if c.spaceID != "" {
		// Get folderless lists from space
		lists, err := c.API.GetFolderlessLists(ctx, c.spaceID)
		if err != nil {
			return fmt.Errorf("failed to get folderless lists: %w", err)
		}
		for _, list := range lists {
			if !list.Archived || c.includeArchived {
				allLists = append(allLists, list)
			}
		}

		// Also get lists from folders in the space
		folders, err := c.API.GetFolders(ctx, c.spaceID)
		if err != nil {
			return fmt.Errorf("failed to get folders: %w", err)
		}

		for _, folder := range folders {
			lists, err := c.API.GetLists(ctx, folder.ID)
			if err != nil {
				// Log warning but continue with other folders
				c.Output.PrintWarning(fmt.Sprintf("Failed to get lists from folder %s: %v", folder.Name, err))
				continue
			}
			for _, list := range lists {
				if !list.Archived || c.includeArchived {
					allLists = append(allLists, list)
				}
			}
		}
	}

	// Get default list ID for highlighting
	defaultListID := c.Config.GetString("default_list")

	// Format output
	format := c.Config.GetString("output")
	if format == "" {
		format = "table"
	}

	if format == "table" {
		// Prepare table data
		type listRow struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Default  string `json:"default"`
			Tasks    int    `json:"tasks"`
			Archived bool   `json:"archived"`
		}

		var rows []listRow
		for _, list := range allLists {
			defaultMarker := ""
			if list.ID == defaultListID {
				defaultMarker = "*"
			}

			row := listRow{
				ID:       list.ID,
				Name:     list.Name,
				Default:  defaultMarker,
				Tasks:    list.TaskCount,
				Archived: list.Archived,
			}
			rows = append(rows, row)
		}

		return c.Output.Print(rows)
	}

	// For other formats, output raw list data
	return c.Output.Print(allLists)
}

// runDefault executes the list default subcommand
func (c *ListCommand) runDefault(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("list ID is required")
	}

	listID := args[0]

	// TODO: Validate that the list exists and is accessible
	// This would require adding GetList method to API interface

	// Save to project config if in a project, otherwise global config
	if c.hasProjectConfig() || c.isProjectFlag {
		// Save to project config
		if projectSaver, ok := c.Config.(interface {
			SaveProjectConfig(map[string]interface{}) error
			GetProjectConfigPath() string
		}); ok {
			settings := map[string]interface{}{
				"default_list": listID,
			}
			if err := projectSaver.SaveProjectConfig(settings); err != nil {
				return fmt.Errorf("failed to save project configuration: %w", err)
			}

			configPath := projectSaver.GetProjectConfigPath()
			if configPath == "" {
				configPath = ".cu.yml"
			}
			c.Output.PrintSuccess(fmt.Sprintf("Default list set to: %s", listID))
			c.Output.PrintInfo(fmt.Sprintf("Saved to project config: %s", configPath))
		} else {
			return fmt.Errorf("project config not supported")
		}
	} else {
		// Save to global config
		c.Config.Set("default_list", listID)
		if saver, ok := c.Config.(interface{ Save() error }); ok {
			if err := saver.Save(); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}
		}
		c.Output.PrintSuccess(fmt.Sprintf("Default list set to: %s (global)", listID))
		c.Output.PrintInfo("Tip: Use --project flag to save to project-specific config")
	}

	return nil
}

// hasProjectConfig checks if project config exists
func (c *ListCommand) hasProjectConfig() bool {
	if checker, ok := c.Config.(interface{ HasProjectConfig() bool }); ok {
		return checker.HasProjectConfig()
	}
	return false
}

// GetCobraCommand returns the cobra command with subcommands
func (c *ListCommand) GetCobraCommand() *cobra.Command {
	cmd := c.Command.GetCobraCommand()

	// Add list subcommand
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all lists",
		Long:  `List all lists in a space or folder.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Set flags from cobra command
			c.spaceID, _ = cmd.Flags().GetString("space")
			c.folderID, _ = cmd.Flags().GetString("folder")
			c.includeArchived, _ = cmd.Flags().GetBool("archived")
			
			return c.runList(cmd.Context(), args)
		},
	}

	// Add flags to list subcommand
	listCmd.Flags().StringP("space", "s", "", "Space ID or name")
	listCmd.Flags().StringP("folder", "f", "", "Folder ID or name")
	listCmd.Flags().Bool("archived", false, "Include archived lists")

	// Add default subcommand
	defaultCmd := &cobra.Command{
		Use:   "default <list-id>",
		Short: "Set default list",
		Long:  `Set the default list for task operations.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Set flags from cobra command
			c.isProjectFlag, _ = cmd.Flags().GetBool("project")
			
			return c.runDefault(cmd.Context(), args)
		},
	}

	// Add flags to default subcommand
	defaultCmd.Flags().BoolP("project", "p", false, "Save to project config instead of global config")

	cmd.AddCommand(listCmd, defaultCmd)

	return cmd
}
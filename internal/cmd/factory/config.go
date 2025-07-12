package factory

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tim/cu/internal/cmd/base"
	"github.com/tim/cu/internal/interfaces"
)

// ConfigCommand implements the config command using dependency injection
type ConfigCommand struct {
	*base.Command
	subcommands map[string]func(context.Context, []string) error
}

// createConfigCommand creates a new config command
func (f *Factory) createConfigCommand() interfaces.Command {
	cmd := &ConfigCommand{
		Command: &base.Command{
			Use:   "config",
			Short: "Manage cu configuration",
			Long:  `View and modify cu configuration settings.`,
			Output: f.output,
			Config: f.config,
		},
		subcommands: make(map[string]func(context.Context, []string) error),
	}

	// Register subcommands
	cmd.subcommands["list"] = cmd.runList
	cmd.subcommands["get"] = cmd.runGet
	cmd.subcommands["set"] = cmd.runSet
	cmd.subcommands["init"] = cmd.runInit
	cmd.subcommands["show"] = cmd.runShow

	// Set the execution function
	cmd.Command.RunFunc = cmd.run

	return cmd
}

// run executes the config command
func (c *ConfigCommand) run(ctx context.Context, args []string) error {
	// If no subcommand, show usage
	if len(args) == 0 {
		return fmt.Errorf("no subcommand specified. Available subcommands: list, get, set, init, show")
	}

	subcommand := args[0]
	handler, exists := c.subcommands[subcommand]
	if !exists {
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}

	// Execute subcommand with remaining args
	return handler(ctx, args[1:])
}

// runList lists all configuration settings
func (c *ConfigCommand) runList(ctx context.Context, args []string) error {
	settings := c.Config.AllSettings()
	
	// Sort keys for consistent output
	keys := make([]string, 0, len(settings))
	for k := range settings {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build output
	var output strings.Builder
	for _, key := range keys {
		output.WriteString(fmt.Sprintf("%s=%v\n", key, settings[key]))
	}

	c.Output.PrintInfo(output.String())
	return nil
}

// runGet gets a configuration value
func (c *ConfigCommand) runGet(ctx context.Context, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("exactly one argument required: key")
	}

	key := args[0]
	value := c.Config.Get(key)
	if value == nil {
		return fmt.Errorf("configuration key '%s' not found", key)
	}

	c.Output.PrintInfo(fmt.Sprintf("%v", value))
	return nil
}

// runSet sets a configuration value
func (c *ConfigCommand) runSet(ctx context.Context, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("exactly two arguments required: key value")
	}

	key := args[0]
	value := args[1]

	// Handle boolean values
	if strings.ToLower(value) == "true" || strings.ToLower(value) == "false" {
		boolValue := strings.ToLower(value) == "true"
		c.Config.Set(key, boolValue)
	} else {
		c.Config.Set(key, value)
	}

	// Save configuration
	if saver, ok := c.Config.(interface{ Save() error }); ok {
		if err := saver.Save(); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}
	}

	c.Output.PrintSuccess(fmt.Sprintf("Set %s to %s", key, value))
	return nil
}

// runInit initializes project configuration
func (c *ConfigCommand) runInit(ctx context.Context, args []string) error {
	// Check if project config already exists
	if checker, ok := c.Config.(interface{ HasProjectConfig() bool }); ok {
		if checker.HasProjectConfig() {
			if pathGetter, ok := c.Config.(interface{ GetProjectConfigPath() string }); ok {
				return fmt.Errorf("project config already exists at: %s", pathGetter.GetProjectConfigPath())
			}
			return fmt.Errorf("project config already exists")
		}
	}

	// Initialize project config
	if initializer, ok := c.Config.(interface{ InitProjectConfig() error }); ok {
		if err := initializer.InitProjectConfig(); err != nil {
			return fmt.Errorf("failed to initialize project config: %w", err)
		}
	} else {
		return fmt.Errorf("project config initialization not supported")
	}

	c.Output.PrintSuccess("Initialized project configuration: .cu.yml")
	c.Output.PrintInfo(`
You can now use project-specific settings such as:
  - Default list for this project
  - Default space for this project
  - Team member aliases

Edit .cu.yml to customize your project settings.`)
	
	return nil
}

// runShow shows current configuration
func (c *ConfigCommand) runShow(ctx context.Context, args []string) error {
	// Build config data
	configData := map[string]interface{}{
		"global": map[string]interface{}{
			"default_space":  c.Config.GetString("default_space"),
			"default_folder": c.Config.GetString("default_folder"),
			"default_list":   c.Config.GetString("default_list"),
			"output":         c.Config.GetString("output"),
			"debug":          c.Config.GetBool("debug"),
		},
	}

	// Add project config if present
	if checker, ok := c.Config.(interface{ HasProjectConfig() bool }); ok && checker.HasProjectConfig() {
		projectData := map[string]interface{}{
			"default_space": c.Config.GetString("default_space"),
			"default_list":  c.Config.GetString("default_list"),
			"output":        c.Config.GetString("output"),
		}
		
		if pathGetter, ok := c.Config.(interface{ GetProjectConfigPath() string }); ok {
			projectData["config_path"] = pathGetter.GetProjectConfigPath()
		}
		
		configData["project"] = projectData
	}

	// Output based on format
	format := c.Config.GetString("output")
	if format == "json" || format == "yaml" {
		return c.Output.Print(configData)
	}

	// Default table/text output
	var output strings.Builder
	output.WriteString("=== Global Configuration ===\n")
	global := configData["global"].(map[string]interface{})
	for k, v := range global {
		output.WriteString(fmt.Sprintf("%s: %v\n", k, v))
	}

	if project, exists := configData["project"]; exists {
		output.WriteString("\n=== Project Configuration ===\n")
		projectMap := project.(map[string]interface{})
		for k, v := range projectMap {
			output.WriteString(fmt.Sprintf("%s: %v\n", k, v))
		}
	}

	c.Output.PrintInfo(output.String())
	return nil
}

// GetCobraCommand returns the cobra command with subcommands
func (c *ConfigCommand) GetCobraCommand() *cobra.Command {
	cmd := c.Command.GetCobraCommand()
	
	// Add subcommands
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all configuration settings",
		Long:  `Display all current configuration settings.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runList(cmd.Context(), args)
		},
	}
	
	getCmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Long:  `Retrieve the value of a specific configuration setting.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runGet(cmd.Context(), args)
		},
	}
	
	setCmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Long:  `Set the value of a specific configuration setting.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runSet(cmd.Context(), args)
		},
	}
	
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize project configuration",
		Long:  `Initialize a project-specific configuration file (.cu.yml) in the current directory.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runInit(cmd.Context(), args)
		},
	}
	
	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Long:  `Display current configuration values from both global and project configs.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runShow(cmd.Context(), args)
		},
	}
	
	cmd.AddCommand(listCmd, getCmd, setCmd, initCmd, showCmd)
	
	return cmd
}
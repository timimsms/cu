package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tim/cu/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage cu configuration",
	Long:  `View and modify cu configuration settings.`,
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration settings",
	Long:  `Display all current configuration settings.`,
	Run: func(cmd *cobra.Command, args []string) {
		settings := viper.AllSettings()
		keys := make([]string, 0, len(settings))
		for k := range settings {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {
			fmt.Printf("%s=%v\n", key, settings[key])
		}
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Long:  `Retrieve the value of a specific configuration setting.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := config.Get(key)
		if value == nil {
			fmt.Fprintf(os.Stderr, "Configuration key '%s' not found\n", key)
			os.Exit(1)
		}
		fmt.Println(value)
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long:  `Set the value of a specific configuration setting.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]

		// Handle boolean values
		if strings.ToLower(value) == "true" || strings.ToLower(value) == "false" {
			config.Set(key, strings.ToLower(value) == "true")
		} else {
			config.Set(key, value)
		}

		if err := config.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save configuration: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Set %s to %s\n", key, value)
	},
}

func init() {
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
}

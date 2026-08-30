package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/timimsms/cu/internal/config"
	"github.com/timimsms/cu/internal/output"
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
			// A token here predates the refusal in `config set`, or was added
			// by hand. Never print it: `config list` output gets pasted into
			// issues and terminals far more casually than an explicit
			// `config get <key>` does.
			if config.IsCredentialKey(key) {
				fmt.Printf("%s=%s\n", key, config.RedactedValue)
				continue
			}
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

		// Redacted for the same reason `config list` is: what this defends
		// against is incidental disclosure — pasted terminal output, a
		// screen-share, a script whose stdout lands in a CI log — and `get` is
		// the spelling most likely to be captured by one. It was never a way to
		// reach a secret cu uses, since authentication reads the keyring; a
		// value here is an unused plaintext leftover. The pointer goes to
		// stderr so it reaches a person without joining piped output.
		if config.IsCredentialKey(key) {
			fmt.Println(config.RedactedValue)
			fmt.Fprintf(os.Stderr,
				"%q is not printed. cu authenticates via the system keyring; this value is an unused plaintext leftover.\nTo read or remove it, edit %s directly.\n",
				key, config.GlobalConfigPath())
			return
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

		// Credentials do not belong in a plaintext config file, and cu would
		// not read one back if they were — authentication uses the keyring.
		if config.IsCredentialKey(key) {
			fmt.Fprintf(os.Stderr, "Refusing to write %q to the config file — it would be stored in plaintext and never used.\n", key)
			fmt.Fprintln(os.Stderr, "Authenticate with 'cu auth login' instead; the token is kept in your system keyring.")
			os.Exit(1)
		}

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

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize project configuration",
	Long:  `Initialize a project-specific configuration file (.cu.yml) in the current directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if config already exists
		if config.HasProjectConfig() {
			fmt.Fprintf(os.Stderr, "Project config already exists at: %s\n", config.GetProjectConfigPath())
			os.Exit(1)
		}

		// Initialize project config
		if err := config.InitProjectConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize project config: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Initialized project configuration: .cu.yml")
		fmt.Println("\nYou can now use project-specific settings such as:")
		fmt.Println("  - Default list for this project")
		fmt.Println("  - Default space for this project")
		fmt.Println("  - Team member aliases")
		fmt.Println("\nEdit .cu.yml to customize your project settings.")
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  `Display current configuration values from both global and project configs.`,
	Run: func(cmd *cobra.Command, args []string) {
		format := cmd.Flag("output").Value.String()

		// Build config data
		configData := map[string]interface{}{
			"global": map[string]interface{}{
				"default_space":  config.GetString("default_space"),
				"default_folder": config.GetString("default_folder"),
				"default_list":   config.GetString("default_list"),
				"output":         config.GetString("output"),
				"debug":          config.GetBool("debug"),
			},
		}

		// Add project config if present
		if config.HasProjectConfig() {
			configData["project"] = map[string]interface{}{
				"config_path":   config.GetProjectConfigPath(),
				"default_space": config.GetString("default_space"),
				"default_list":  config.GetString("default_list"),
				"output":        config.GetString("output"),
			}
		}

		// Format output
		if err := output.Format(format, configData); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to format output: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
}

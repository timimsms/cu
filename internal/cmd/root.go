package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tim/cu/internal/config"
	"github.com/tim/cu/internal/version"
)

var (
	cfgFile      string
	debug        bool
	outputFormat string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cu",
	Short: "A GitHub CLI-inspired command-line interface for ClickUp",
	Long: `cu is a command-line interface for ClickUp that provides GitHub CLI-like
functionality for managing tasks, lists, spaces, and other ClickUp resources.

It allows developers and teams to interact with ClickUp directly from the terminal,
enabling efficient task management and seamless integration with development workflows.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize configuration
		if err := config.Init(cfgFile); err != nil {
			return fmt.Errorf("failed to initialize config: %w", err)
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/cu/config.yml)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug mode")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "output format (table|json|yaml|csv)")

	// Bind flags to viper
	if err := viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug")); err != nil {
		// Log error but don't fail - this is non-critical
		fmt.Fprintf(os.Stderr, "Warning: failed to bind debug flag: %v\n", err)
	}
	if err := viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output")); err != nil {
		// Log error but don't fail - this is non-critical
		fmt.Fprintf(os.Stderr, "Warning: failed to bind output flag: %v\n", err)
	}

	// Version flag
	rootCmd.Version = version.Version
	rootCmd.SetVersionTemplate(version.FullVersion())

	// Add subcommands
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(meCmd)
	rootCmd.AddCommand(taskCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(spaceCmd)
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".cu" (without extension)
		viper.AddConfigPath(home + "/.config/cu")
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	// Read in environment variables that match
	viper.SetEnvPrefix("CU")
	viper.AutomaticEnv()

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil && debug {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

package cmd

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/tim/cu/internal/api"
	"github.com/tim/cu/internal/auth"
	"github.com/tim/cu/internal/cmd/factory"
	"github.com/tim/cu/internal/config"
	"github.com/tim/cu/internal/output"
)

// ExecuteWithFactory creates and runs the CLI using the factory pattern
func ExecuteWithFactory() error {
	// Initialize configuration
	cfg, err := initializeConfig()
	if err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	// Create dependencies
	authManager := auth.NewManager(cfg)
	apiClient := api.NewClient(authManager)
	// Initialize API connection
	if err := apiClient.Connect(); err != nil {
		// It's okay if not authenticated yet, commands will handle it
	}
	outputFormatter := output.NewFormatter(cfg)

	// Create factory with dependencies
	cmdFactory := factory.New(
		factory.WithAPIClient(apiClient),
		factory.WithAuthManager(authManager),
		factory.WithOutputFormatter(outputFormatter),
		factory.WithConfigProvider(cfg),
	)

	// Create root command
	rootCmd, err := factory.NewRootCommand(cmdFactory)
	if err != nil {
		return fmt.Errorf("failed to create root command: %w", err)
	}

	// Execute
	return rootCmd.Execute()
}

// initializeConfig sets up the configuration system
func initializeConfig() (*config.Provider, error) {
	// Set defaults
	viper.SetDefault("output", "table")
	viper.SetDefault("debug", false)

	// Set config search paths
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.config/cu")
	viper.AddConfigPath(".")

	// Read environment variables
	viper.SetEnvPrefix("CU")
	viper.AutomaticEnv()

	// Try to read config file
	if err := viper.ReadInConfig(); err != nil {
		// It's okay if config file doesn't exist
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config: %w", err)
		}
	}

	// Create config provider instance
	return config.New(), nil
}
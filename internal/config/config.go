package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	DefaultSpace  string            `mapstructure:"default_space"`
	DefaultFolder string            `mapstructure:"default_folder"`
	DefaultList   string            `mapstructure:"default_list"`
	Output        string            `mapstructure:"output"`
	Debug         bool              `mapstructure:"debug"`
	APIToken      string            `mapstructure:"api_token"`
	Workspaces    map[string]string `mapstructure:"workspaces"`
}

var (
	// DefaultConfigDir is the default configuration directory
	DefaultConfigDir = filepath.Join(os.Getenv("HOME"), ".config", "cu")
	// ConfigFileName is the name of the config file
	ConfigFileName = "config"
	// ConfigType is the type of the config file
	ConfigType = "yaml"
	// ProjectConfigFileName is the name of the project config file
	ProjectConfigFileName = ".cu.yml"

	// Track if we're in a project with config
	hasProjectConfig  bool
	projectConfigPath string
)

// Init initializes the configuration
func Init(cfgFile string) error {
	// Create config directory if it doesn't exist
	if err := os.MkdirAll(DefaultConfigDir, 0750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Set default values
	viper.SetDefault("output", "table")
	viper.SetDefault("debug", false)

	// Look for project config file in current directory and parent directories
	projectConfigPath = findProjectConfig()
	if projectConfigPath != "" {
		hasProjectConfig = true
		projectViper := viper.New()
		projectViper.SetConfigFile(projectConfigPath)

		// Read project config
		if err := projectViper.ReadInConfig(); err == nil {
			// Merge project config with main config
			// Project config takes precedence
			for k, v := range projectViper.AllSettings() {
				viper.Set(k, v)
			}
		}
	}

	return nil
}

// Load loads the configuration from file
func Load() (*Config, error) {
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return &cfg, nil
}

// Save saves the current configuration to file
func Save() error {
	configPath := filepath.Join(DefaultConfigDir, ConfigFileName+"."+ConfigType)
	return viper.WriteConfigAs(configPath)
}

// Get returns a configuration value
func Get(key string) interface{} {
	return viper.Get(key)
}

// Set sets a configuration value
func Set(key string, value interface{}) {
	viper.Set(key, value)
}

// GetString returns a string configuration value
func GetString(key string) string {
	return viper.GetString(key)
}

// GetBool returns a boolean configuration value
func GetBool(key string) bool {
	return viper.GetBool(key)
}

// findProjectConfig looks for .cu.yml in current directory and parent directories
func findProjectConfig() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Get the absolute path to ensure we're working with real paths
	dir, err = filepath.Abs(dir)
	if err != nil {
		return ""
	}

	// Look up to 10 levels up
	for i := 0; i < 10; i++ {
		configPath := filepath.Join(dir, ProjectConfigFileName)
		configPath = filepath.Clean(configPath)

		// Check if file exists and is a regular file (not a symlink)
		if info, err := os.Lstat(configPath); err == nil {
			if info.Mode().IsRegular() {
				return configPath
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			break
		}
		dir = parent
	}

	return ""
}

// HasProjectConfig returns true if a project config file was found
func HasProjectConfig() bool {
	return hasProjectConfig
}

// GetProjectConfigPath returns the path to the project config file
func GetProjectConfigPath() string {
	return projectConfigPath
}

// SaveProjectConfig saves configuration to the project config file
func SaveProjectConfig(settings map[string]interface{}) error {
	// If no project config exists, create one in current directory
	if projectConfigPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		projectConfigPath = filepath.Join(cwd, ProjectConfigFileName)
	}

	// Clean and validate the config path
	projectConfigPath = filepath.Clean(projectConfigPath)

	// Get absolute path for validation
	absPath, err := filepath.Abs(projectConfigPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Ensure the config file is within or above the current directory (for traversal)
	// but not in system directories
	if strings.Contains(absPath, "..") || !strings.HasPrefix(absPath, "/") {
		return fmt.Errorf("invalid config path: contains invalid characters")
	}

	// Create a new viper instance for project config
	projectViper := viper.New()
	projectViper.SetConfigFile(projectConfigPath)

	// If file exists, read current content
	if _, err := os.Stat(projectConfigPath); err == nil {
		if err := projectViper.ReadInConfig(); err != nil {
			return fmt.Errorf("failed to read existing project config: %w", err)
		}
	}

	// Update with new settings
	for k, v := range settings {
		projectViper.Set(k, v)
		// Also update main viper
		viper.Set(k, v)
	}

	// Write the file
	if err := projectViper.WriteConfig(); err != nil {
		// If file doesn't exist, create it
		if os.IsNotExist(err) {
			if err := projectViper.SafeWriteConfig(); err != nil {
				return fmt.Errorf("failed to create project config: %w", err)
			}
		} else {
			return fmt.Errorf("failed to write project config: %w", err)
		}
	}

	hasProjectConfig = true
	return nil
}

// InitProjectConfig creates a new project config file in the current directory
func InitProjectConfig() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Ensure we're using a clean, safe filename
	configPath := filepath.Join(cwd, ProjectConfigFileName)
	configPath = filepath.Clean(configPath)

	// Verify the path is within the current directory
	if !strings.HasPrefix(configPath, cwd) {
		return fmt.Errorf("invalid config path: must be within current directory")
	}

	// Check if already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("project config already exists at %s", configPath)
	}

	// Create with default content
	projectViper := viper.New()
	projectViper.SetConfigFile(configPath)

	// Set some default project settings
	projectViper.Set("project_name", filepath.Base(cwd))
	projectViper.SetDefault("default_list", "")
	projectViper.SetDefault("default_space", "")
	projectViper.SetDefault("output", "table")

	// Add helpful comments by writing a template
	template := `# ClickUp CLI Project Configuration
# This file contains project-specific settings for the cu CLI

# Project name
project_name: %s

# Default space for this project
# default_space: "My Space"

# Default list for task operations
# default_list: "abc123"

# Default output format (table|json|yaml|csv)
output: table

# Team member aliases for easier assignment
# aliases:
#   john: john.doe@example.com
#   jane: jane.smith@example.com
`

	content := fmt.Sprintf(template, filepath.Base(cwd))
	// #nosec G304 - configPath is validated to be within current directory
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write project config: %w", err)
	}

	projectConfigPath = configPath
	hasProjectConfig = true

	return nil
}

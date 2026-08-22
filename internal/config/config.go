package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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

	// globalConfigPath is the global config file discovered by Init, if any.
	globalConfigPath string
	// explicitConfigFile records a --config path, which always wins.
	explicitConfigFile string
	// staged holds values written through Set. Save applies these on top of
	// whatever is already on disk, so project .cu.yml values, environment
	// variables and flags can never be baked into ~/.config/cu/config.yaml.
	staged = map[string]interface{}{}
)

// credentialKeys are never accepted from a project .cu.yml. That file is
// committed and reviewed like code, so honouring a token there would let any
// repository you clone substitute the credential used for API calls.
var credentialKeys = []string{"api_token"}

// globalPath returns the global config file to write. An explicit --config
// always wins; otherwise a discovered file is used only while it still lives
// under the configured directory, since DefaultConfigDir is a variable that
// tests and tooling repoint.
func globalPath() string {
	if explicitConfigFile != "" {
		return explicitConfigFile
	}
	fallback := filepath.Join(DefaultConfigDir, ConfigFileName+"."+ConfigType)
	if globalConfigPath != "" && filepath.Dir(globalConfigPath) == filepath.Clean(DefaultConfigDir) {
		return globalConfigPath
	}
	return fallback
}

// Init initializes the configuration
func Init(cfgFile string) error {
	// Create config directory if it doesn't exist
	if err := os.MkdirAll(DefaultConfigDir, 0750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Set default values
	viper.SetDefault("output", "table")
	viper.SetDefault("debug", false)

	// Environment variables outrank both config layers below.
	viper.SetEnvPrefix("CU")
	viper.AutomaticEnv()

	// --- global config layer -------------------------------------------------
	staged = map[string]interface{}{}
	explicitConfigFile = cfgFile
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(DefaultConfigDir)
		viper.AddConfigPath(".")
		viper.SetConfigType(ConfigType)
		viper.SetConfigName(ConfigFileName)
	}
	// A missing global config is normal on a fresh machine.
	_ = viper.ReadInConfig()

	globalConfigPath = viper.ConfigFileUsed()

	// --- project overlay -----------------------------------------------------
	// Look for project config file in current directory and parent directories
	projectConfigPath = findProjectConfig()
	if projectConfigPath != "" {
		hasProjectConfig = true
		projectViper := viper.New()
		projectViper.SetConfigFile(projectConfigPath)

		// Read project config
		if err := projectViper.ReadInConfig(); err == nil {
			settings := projectViper.AllSettings()
			for _, k := range credentialKeys {
				if _, present := settings[k]; present {
					delete(settings, k)
					fmt.Fprintf(os.Stderr,
						"cu: ignoring %q in %s — credentials come from the keyring, environment, or your global config\n",
						k, projectConfigPath)
				}
			}
			// MergeConfigMap merges into viper's *config* layer, so project
			// values override the global file while still losing to
			// environment variables and command-line flags. Using viper.Set
			// here would place them in the override slot, which outranks
			// everything — the inversion this replaces.
			if err := viper.MergeConfigMap(settings); err != nil {
				return fmt.Errorf("failed to merge project config %s: %w", projectConfigPath, err)
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

// Save writes the global config file. Only values that came from that file or
// were written through Set are persisted — project .cu.yml values, environment
// variables and flags are deliberately excluded, so running `cu config set`
// inside a project can no longer bake that project's settings into the global
// config.
func Save() error {
	path := globalPath()

	// Start from what is already on disk so a write can never truncate
	// settings this process did not load, then apply only explicit Sets.
	gv := viper.New()
	gv.SetConfigFile(path)
	_ = gv.ReadInConfig()
	for k, v := range staged {
		gv.Set(k, v)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	return gv.WriteConfigAs(path)
}

// Get returns a configuration value
func Get(key string) interface{} {
	return viper.Get(key)
}

// Set sets a configuration value for this process and stages it for the global
// config file, so a following Save persists it there.
func Set(key string, value interface{}) {
	viper.Set(key, value)
	staged[key] = value
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

	// Ensure the config file is safe - check if it's trying to escape current directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Convert to absolute for comparison
	absCwd, _ := filepath.Abs(cwd)

	// The config should be within the current directory tree
	if !strings.HasPrefix(absPath, absCwd) {
		return fmt.Errorf("invalid config path: outside current directory")
	}

	// Check for absolute path based on OS
	if runtime.GOOS == "windows" {
		// On Windows, absolute paths start with drive letter (e.g., C:\)
		if len(absPath) < 3 || absPath[1] != ':' || absPath[2] != '\\' {
			return fmt.Errorf("invalid config path: must be absolute path")
		}
	} else {
		// On Unix-like systems, absolute paths start with /
		if !strings.HasPrefix(absPath, "/") {
			return fmt.Errorf("invalid config path: must be absolute path")
		}
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
	}
	// Reflect them in the running process at project precedence — below env and
	// flags, above the global file — matching how Init loads them.
	if err := viper.MergeConfigMap(settings); err != nil {
		return fmt.Errorf("failed to apply project config: %w", err)
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

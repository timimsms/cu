package config

import (
	"fmt"
	"os"
	"path/filepath"

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
)

// Init initializes the configuration
func Init(cfgFile string) error {
	// Create config directory if it doesn't exist
	if err := os.MkdirAll(DefaultConfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Set default values
	viper.SetDefault("output", "table")
	viper.SetDefault("debug", false)

	// Look for project config file
	projectViper := viper.New()
	projectViper.SetConfigName(".cu")
	projectViper.SetConfigType("yml")
	projectViper.AddConfigPath(".")
	
	// Read project config if it exists
	if err := projectViper.ReadInConfig(); err == nil {
		// Merge project config with main config
		for k, v := range projectViper.AllSettings() {
			viper.Set(k, v)
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
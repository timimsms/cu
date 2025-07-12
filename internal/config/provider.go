package config

import (
	"github.com/spf13/viper"
)

// Provider wraps viper to implement the ConfigProvider interface
type Provider struct {
	viper *viper.Viper
}

// New creates a new config provider using the global viper instance
func New() *Provider {
	return &Provider{
		viper: viper.GetViper(),
	}
}

// NewWithViper creates a new config provider with a specific viper instance
func NewWithViper(v *viper.Viper) *Provider {
	return &Provider{
		viper: v,
	}
}

// Get returns a configuration value
func (p *Provider) Get(key string) interface{} {
	return p.viper.Get(key)
}

// GetString returns a string configuration value
func (p *Provider) GetString(key string) string {
	return p.viper.GetString(key)
}

// GetBool returns a boolean configuration value
func (p *Provider) GetBool(key string) bool {
	return p.viper.GetBool(key)
}

// GetInt returns an integer configuration value
func (p *Provider) GetInt(key string) int {
	return p.viper.GetInt(key)
}

// GetStringSlice returns a string slice configuration value
func (p *Provider) GetStringSlice(key string) []string {
	return p.viper.GetStringSlice(key)
}

// GetStringMap returns a string map configuration value
func (p *Provider) GetStringMap(key string) map[string]interface{} {
	return p.viper.GetStringMap(key)
}

// Set sets a configuration value
func (p *Provider) Set(key string, value interface{}) {
	p.viper.Set(key, value)
}

// IsSet checks if a key exists
func (p *Provider) IsSet(key string) bool {
	return p.viper.IsSet(key)
}

// AllSettings returns all settings
func (p *Provider) AllSettings() map[string]interface{} {
	return p.viper.AllSettings()
}

// Save saves the configuration
func (p *Provider) Save() error {
	return Save()
}

// HasProjectConfig returns true if a project config file was found
func (p *Provider) HasProjectConfig() bool {
	return HasProjectConfig()
}

// GetProjectConfigPath returns the path to the project config file
func (p *Provider) GetProjectConfigPath() string {
	return GetProjectConfigPath()
}

// InitProjectConfig creates a new project config file
func (p *Provider) InitProjectConfig() error {
	return InitProjectConfig()
}
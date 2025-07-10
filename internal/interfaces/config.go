package interfaces

// ConfigProvider defines the interface for configuration access
type ConfigProvider interface {
	// Get configuration values
	Get(key string) interface{}
	GetString(key string) string
	GetBool(key string) bool
	GetInt(key string) int
	GetStringSlice(key string) []string
	GetStringMap(key string) map[string]interface{}

	// Set configuration values
	Set(key string, value interface{})

	// Check if key exists
	IsSet(key string) bool

	// Get all settings
	AllSettings() map[string]interface{}
}
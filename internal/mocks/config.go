package mocks

// MockConfigProvider is a mock implementation of ConfigProvider for testing
type MockConfigProvider struct {
	values map[string]interface{}
}

// NewMockConfigProvider creates a new mock config provider
func NewMockConfigProvider() *MockConfigProvider {
	return &MockConfigProvider{
		values: make(map[string]interface{}),
	}
}

// Get returns a configuration value
func (m *MockConfigProvider) Get(key string) interface{} {
	return m.values[key]
}

// GetString returns a string configuration value
func (m *MockConfigProvider) GetString(key string) string {
	if val, ok := m.values[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// GetBool returns a boolean configuration value
func (m *MockConfigProvider) GetBool(key string) bool {
	if val, ok := m.values[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// GetInt returns an integer configuration value
func (m *MockConfigProvider) GetInt(key string) int {
	if val, ok := m.values[key]; ok {
		if i, ok := val.(int); ok {
			return i
		}
	}
	return 0
}

// GetStringSlice returns a string slice configuration value
func (m *MockConfigProvider) GetStringSlice(key string) []string {
	if val, ok := m.values[key]; ok {
		if slice, ok := val.([]string); ok {
			return slice
		}
	}
	return nil
}

// GetStringMap returns a string map configuration value
func (m *MockConfigProvider) GetStringMap(key string) map[string]interface{} {
	if val, ok := m.values[key]; ok {
		if m, ok := val.(map[string]interface{}); ok {
			return m
		}
	}
	return nil
}

// Set sets a configuration value
func (m *MockConfigProvider) Set(key string, value interface{}) {
	m.values[key] = value
}

// IsSet checks if a key exists
func (m *MockConfigProvider) IsSet(key string) bool {
	_, ok := m.values[key]
	return ok
}

// AllSettings returns all settings
func (m *MockConfigProvider) AllSettings() map[string]interface{} {
	// Return a copy to prevent external modification
	result := make(map[string]interface{})
	for k, v := range m.values {
		result[k] = v
	}
	return result
}

// MockConfigWithProject is a mock config that supports project config operations
type MockConfigWithProject struct {
	*MockConfigProvider
	HasProjectConfigVal   bool
	ProjectConfigSaved    bool
	ProjectSettings       map[string]interface{}
	SaveProjectConfigErr  error
	ProjectConfigPath     string
}

func (m *MockConfigWithProject) HasProjectConfig() bool {
	return m.HasProjectConfigVal
}

func (m *MockConfigWithProject) SaveProjectConfig(settings map[string]interface{}) error {
	if m.SaveProjectConfigErr != nil {
		return m.SaveProjectConfigErr
	}
	m.ProjectConfigSaved = true
	if m.ProjectSettings == nil {
		m.ProjectSettings = make(map[string]interface{})
	}
	for k, v := range settings {
		m.ProjectSettings[k] = v
	}
	return nil
}

func (m *MockConfigWithProject) GetProjectConfigPath() string {
	if m.ProjectConfigPath != "" {
		return m.ProjectConfigPath
	}
	return ".cu.yml"
}

// MockConfigWithSaveError is a mock config that returns save errors
type MockConfigWithSaveError struct {
	*MockConfigProvider
	SaveErr error
}

func (m *MockConfigWithSaveError) Save() error {
	return m.SaveErr
}
package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// Simple tests that don't involve os.Exit

func TestConfigCommand_Basic(t *testing.T) {
	cmd := configCmd
	assert.NotNil(t, cmd)
	assert.Equal(t, "config", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	
	// Verify subcommands
	subcommands := map[string]bool{
		"list": false,
		"get":  false,
		"set":  false,
		"init": false,
		"show": false,
	}
	
	for _, child := range cmd.Commands() {
		name := strings.Split(child.Use, " ")[0]
		if _, ok := subcommands[name]; ok {
			subcommands[name] = true
		}
	}
	
	for name, found := range subcommands {
		assert.True(t, found, "Subcommand %s should exist", name)
	}
}

func TestConfigListCmd_Metadata(t *testing.T) {
	cmd := configListCmd
	assert.NotNil(t, cmd)
	assert.Equal(t, "list", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotNil(t, cmd.Run)
}

func TestConfigGetCmd_Metadata(t *testing.T) {
	cmd := configGetCmd
	assert.NotNil(t, cmd)
	assert.Equal(t, "get <key>", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotNil(t, cmd.Run)
	assert.NotNil(t, cmd.Args)
}

func TestConfigSetCmd_Metadata(t *testing.T) {
	cmd := configSetCmd
	assert.NotNil(t, cmd)
	assert.Equal(t, "set <key> <value>", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotNil(t, cmd.Run)
	assert.NotNil(t, cmd.Args)
}

func TestConfigInitCmd_Metadata(t *testing.T) {
	cmd := configInitCmd
	assert.NotNil(t, cmd)
	assert.Equal(t, "init", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotNil(t, cmd.Run)
}

func TestConfigShowCmd_Metadata(t *testing.T) {
	cmd := configShowCmd
	assert.NotNil(t, cmd)
	assert.Equal(t, "show", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotNil(t, cmd.Run)
}

// Test config value handling (without executing commands)
func TestConfigValueHandling(t *testing.T) {
	// Save viper state
	originalViper := viper.New()
	for _, key := range viper.AllKeys() {
		originalViper.Set(key, viper.Get(key))
	}
	defer func() {
		viper.Reset()
		for _, key := range originalViper.AllKeys() {
			viper.Set(key, originalViper.Get(key))
		}
	}()

	tests := []struct {
		name     string
		setup    func()
		key      string
		expected interface{}
	}{
		{
			name: "string value",
			setup: func() {
				viper.Set("test_string", "hello")
			},
			key:      "test_string",
			expected: "hello",
		},
		{
			name: "boolean value",
			setup: func() {
				viper.Set("test_bool", true)
			},
			key:      "test_bool",
			expected: true,
		},
		{
			name: "integer value",
			setup: func() {
				viper.Set("test_int", 42)
			},
			key:      "test_int",
			expected: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			tt.setup()
			
			value := viper.Get(tt.key)
			assert.Equal(t, tt.expected, value)
		})
	}
}
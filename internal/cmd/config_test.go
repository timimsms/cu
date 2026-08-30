package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/timimsms/cu/internal/config"
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

// captureStdout runs fn with os.Stdout redirected and returns what it wrote.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	orig := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	fn()
	_ = w.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("read captured stdout: %v", err)
	}
	return buf.String()
}

func TestConfigGetRedactsCredentials(t *testing.T) {
	// The value is never printed even though `get` names the key explicitly:
	// what redaction defends against is incidental disclosure, and `get` is the
	// spelling most likely to be captured into a log or a pasted transcript.
	t.Run("credential key is redacted", func(t *testing.T) {
		viper.Reset()
		t.Cleanup(viper.Reset)
		viper.Set("api_token", "sk-must-not-be-printed")

		out := captureStdout(t, func() { configGetCmd.Run(configGetCmd, []string{"api_token"}) })

		assert.NotContains(t, out, "sk-must-not-be-printed", "the token must not reach stdout")
		assert.Contains(t, out, config.RedactedValue)
	})

	t.Run("ordinary key still prints its value", func(t *testing.T) {
		viper.Reset()
		t.Cleanup(viper.Reset)
		viper.Set("default_list", "abc123")

		out := captureStdout(t, func() { configGetCmd.Run(configGetCmd, []string{"default_list"}) })

		assert.Contains(t, out, "abc123")
	})
}

func TestConfigListRedactsCredentials(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	viper.Set("api_token", "sk-must-not-be-printed")
	viper.Set("default_list", "abc123")

	out := captureStdout(t, func() { configListCmd.Run(configListCmd, nil) })

	assert.NotContains(t, out, "sk-must-not-be-printed")
	assert.Contains(t, out, "api_token="+config.RedactedValue)
	assert.Contains(t, out, "default_list=abc123", "ordinary keys are unaffected")
}

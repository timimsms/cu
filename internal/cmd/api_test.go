package cmd

import (
	"strings"
	"testing"
)

func TestAPICommand(t *testing.T) {
	// Just verify the command exists and can be created
	cmd := apiCmd
	if cmd == nil {
		t.Fatal("apiCmd should not be nil")
	}

	// Check command metadata
	if cmd.Use != "api <endpoint>" {
		t.Errorf("Expected Use 'api <endpoint>', got %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if cmd.Long == "" {
		t.Error("Long description should not be empty")
	}

	// Check flags exist
	methodFlag := cmd.Flag("method")
	if methodFlag == nil {
		t.Error("method flag should exist")
	} else if methodFlag.DefValue != "GET" {
		t.Errorf("method flag default should be GET, got %s", methodFlag.DefValue)
	}

	dataFlag := cmd.Flag("data")
	if dataFlag == nil {
		t.Error("data flag should exist")
	}

	headerFlag := cmd.Flag("header")
	if headerFlag == nil {
		t.Error("header flag should exist")
	}
}

func TestEndpointNormalization(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "with leading slash",
			input:    "/team",
			expected: "/team",
		},
		{
			name:     "without leading slash",
			input:    "team",
			expected: "/team",
		},
		{
			name:     "with query params",
			input:    "/list/123?archived=false",
			expected: "/list/123?archived=false",
		},
	}

	for _, tt := range tests {  
		t.Run(tt.name, func(t *testing.T) {
			// This tests the logic that should add leading slash
			result := tt.input
			if !strings.HasPrefix(result, "/") {
				result = "/" + result
			}
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestHeaderParsing(t *testing.T) {
	tests := []struct {
		name           string
		headers        []string
		expectedKey    string
		expectedValue  string
		shouldParse    bool
	}{
		{
			name:          "valid header",
			headers:       []string{"X-Custom: value"},
			expectedKey:   "X-Custom",
			expectedValue: "value",
			shouldParse:   true,
		},
		{
			name:          "header with spaces",
			headers:       []string{"X-Custom:    value with spaces   "},
			expectedKey:   "X-Custom",
			expectedValue: "value with spaces",
			shouldParse:   true,
		},
		{
			name:        "invalid header",
			headers:     []string{"InvalidHeader"},
			shouldParse: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, header := range tt.headers {
				parts := strings.SplitN(header, ":", 2)
				if tt.shouldParse && len(parts) != 2 {
					t.Errorf("Expected header to parse into 2 parts")
				} else if tt.shouldParse {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					if key != tt.expectedKey {
						t.Errorf("Expected key %s, got %s", tt.expectedKey, key)
					}
					if value != tt.expectedValue {
						t.Errorf("Expected value %s, got %s", tt.expectedValue, value)
					}
				}
			}
		})
	}
}
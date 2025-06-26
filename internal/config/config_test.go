package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestInit(t *testing.T) {
	// Create temp directory for config
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	defer func() { _ = os.Setenv("HOME", oldHome) }()
	
	// Update DefaultConfigDir for test
	oldConfigDir := DefaultConfigDir
	DefaultConfigDir = filepath.Join(tmpDir, ".config", "cu")
	defer func() { DefaultConfigDir = oldConfigDir }()
	
	err := Init("")
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	
	// Check if config directory was created
	if _, err := os.Stat(DefaultConfigDir); os.IsNotExist(err) {
		t.Error("Config directory was not created")
	}
}

func TestGetSet(t *testing.T) {
	// Reset viper for clean test
	viper.Reset()
	
	// Test Set and Get
	Set("test_key", "test_value")
	value := GetString("test_key")
	if value != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", value)
	}
	
	// Test GetBool
	Set("bool_key", true)
	boolValue := GetBool("bool_key")
	if !boolValue {
		t.Error("Expected true, got false")
	}
}

func TestLoad(t *testing.T) {
	// Reset viper
	viper.Reset()
	
	// Set some test values
	viper.Set("default_space", "TestSpace")
	viper.Set("output", "json")
	viper.Set("debug", true)
	
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	
	if cfg.DefaultSpace != "TestSpace" {
		t.Errorf("Expected DefaultSpace 'TestSpace', got '%s'", cfg.DefaultSpace)
	}
	if cfg.Output != "json" {
		t.Errorf("Expected Output 'json', got '%s'", cfg.Output)
	}
	if !cfg.Debug {
		t.Error("Expected Debug true, got false")
	}
}
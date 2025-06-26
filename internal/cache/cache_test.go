package cache

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create a cache with a short TTL
	c := &Cache{
		dir: tmpDir,
		ttl: 1 * time.Second,
	}

	// Test data
	type testData struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	original := testData{Name: "test", Value: 42}

	// Test Set
	err := c.Set("test-key", original)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	// Verify file was created with hashed name
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read cache directory: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("Expected 1 cache file, got %d", len(files))
	}

	// Test Get - should succeed
	var retrieved testData
	err = c.Get("test-key", &retrieved)
	if err != nil {
		t.Fatalf("Failed to get from cache: %v", err)
	}

	if retrieved.Name != original.Name || retrieved.Value != original.Value {
		t.Errorf("Retrieved data doesn't match: got %+v, want %+v", retrieved, original)
	}

	// Test expiration
	time.Sleep(2 * time.Second)
	err = c.Get("test-key", &retrieved)
	if err == nil {
		t.Error("Expected cache to be expired, but Get succeeded")
	}

	// Test Delete
	err = c.Set("delete-test", original)
	if err != nil {
		t.Fatalf("Failed to set cache for delete test: %v", err)
	}

	err = c.Delete("delete-test")
	if err != nil {
		t.Fatalf("Failed to delete from cache: %v", err)
	}

	err = c.Get("delete-test", &retrieved)
	if err == nil {
		t.Error("Expected Get to fail after Delete, but it succeeded")
	}
}

func TestCacheClear(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	c := &Cache{
		dir: tmpDir,
		ttl: 1 * time.Hour,
	}

	// Add multiple items
	for i := 0; i < 5; i++ {
		err := c.Set(string(rune('a'+i)), i)
		if err != nil {
			t.Fatalf("Failed to set cache item %d: %v", i, err)
		}
	}

	// Verify files were created
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read cache directory: %v", err)
	}
	if len(files) != 5 {
		t.Errorf("Expected 5 cache files, got %d", len(files))
	}

	// Clear cache
	err = c.Clear()
	if err != nil {
		t.Fatalf("Failed to clear cache: %v", err)
	}

	// Verify all files were removed
	files, err = os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read cache directory after clear: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("Expected 0 cache files after clear, got %d", len(files))
	}
}

func TestCacheFilename(t *testing.T) {
	c := &Cache{dir: "/tmp"}

	// Test that different keys produce different filenames
	name1 := c.filename("key1")
	name2 := c.filename("key2")

	if name1 == name2 {
		t.Error("Different keys should produce different filenames")
	}

	// Test that the same key produces the same filename
	name1Again := c.filename("key1")
	if name1 != name1Again {
		t.Error("Same key should produce same filename")
	}

	// Test that filename is in the correct directory
	if !strings.HasPrefix(name1, "/tmp/") {
		t.Errorf("Filename should be in cache directory: %s", name1)
	}

	// Test that filename ends with .json
	if filepath.Ext(name1) != ".json" {
		t.Errorf("Filename should end with .json: %s", name1)
	}
}

func TestCacheSafety(t *testing.T) {
	c := &Cache{dir: "/tmp"}

	// Test potentially dangerous keys
	dangerousKeys := []string{
		"../../../etc/passwd",
		"..\\..\\..\\windows\\system32\\config\\sam",
		"/etc/passwd",
		"C:\\Windows\\System32\\config\\SAM",
		"../../sensitive-file",
		"key with spaces",
		"key:with:colons",
		"key|with|pipes",
		"key*with*asterisks",
	}

	filenames := make(map[string]bool)

	for _, key := range dangerousKeys {
		filename := c.filename(key)

		// Check that filename is within cache directory
		if !strings.HasPrefix(filename, "/tmp/") {
			t.Errorf("Dangerous key %q produced filename outside cache dir: %s", key, filename)
		}

		// Check that filename doesn't contain path traversal
		if filepath.Clean(filename) != filename {
			t.Errorf("Filename for key %q is not clean: %s", key, filename)
		}

		// Check for uniqueness
		if filenames[filename] {
			t.Errorf("Duplicate filename generated for key %q: %s", key, filename)
		}
		filenames[filename] = true
	}
}

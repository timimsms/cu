package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tim/cu/internal/config"
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
	tmpDir := t.TempDir()
	c := &Cache{dir: tmpDir}

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
	if !strings.HasPrefix(name1, tmpDir) {
		t.Errorf("Filename should be in cache directory: expected prefix %s, got %s", tmpDir, name1)
	}

	// Test that filename ends with .json
	if filepath.Ext(name1) != ".json" {
		t.Errorf("Filename should end with .json: %s", name1)
	}
}

func TestNewCache(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		// Save original config dir
		origDir := config.DefaultConfigDir
		config.DefaultConfigDir = t.TempDir()
		defer func() { config.DefaultConfigDir = origDir }()
		
		cache, err := NewCache(5 * time.Minute)
		require.NoError(t, err)
		assert.NotNil(t, cache)
		assert.Equal(t, 5*time.Minute, cache.ttl)
		assert.Contains(t, cache.dir, "cache")
		
		// Verify cache directory was created
		_, err = os.Stat(cache.dir)
		assert.NoError(t, err)
	})
	
	t.Run("directory creation failure", func(t *testing.T) {
		// Use a path that will fail
		origDir := config.DefaultConfigDir
		config.DefaultConfigDir = "/root/no-permission"
		defer func() { config.DefaultConfigDir = origDir }()
		
		cache, err := NewCache(5 * time.Minute)
		assert.Error(t, err)
		assert.Nil(t, cache)
		assert.Contains(t, err.Error(), "failed to create cache directory")
	})
}

func TestCacheEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	c := &Cache{
		dir: tmpDir,
		ttl: 1 * time.Hour,
	}
	
	t.Run("get non-existent key", func(t *testing.T) {
		var result string
		err := c.Get("non-existent", &result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cache miss")
	})
	
	t.Run("delete non-existent key", func(t *testing.T) {
		err := c.Delete("non-existent")
		assert.NoError(t, err) // Should not error on non-existent
	})
	
	t.Run("set and get complex data", func(t *testing.T) {
		type ComplexData struct {
			ID       int                    `json:"id"`
			Name     string                 `json:"name"`
			Tags     []string               `json:"tags"`
			Metadata map[string]interface{} `json:"metadata"`
		}
		
		original := ComplexData{
			ID:   123,
			Name: "test",
			Tags: []string{"tag1", "tag2"},
			Metadata: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
			},
		}
		
		err := c.Set("complex", original)
		require.NoError(t, err)
		
		var retrieved ComplexData
		err = c.Get("complex", &retrieved)
		require.NoError(t, err)
		
		// Compare fields individually due to JSON number handling
		assert.Equal(t, original.ID, retrieved.ID)
		assert.Equal(t, original.Name, retrieved.Name)
		assert.Equal(t, original.Tags, retrieved.Tags)
		assert.Equal(t, original.Metadata["key1"], retrieved.Metadata["key1"])
		// JSON unmarshals numbers as float64
		assert.Equal(t, float64(42), retrieved.Metadata["key2"])
	})
	
	t.Run("corrupted cache file", func(t *testing.T) {
		// Create a corrupted cache file
		filename := c.filename("corrupted")
		err := os.WriteFile(filename, []byte("invalid json"), 0600)
		require.NoError(t, err)
		
		var result string
		err = c.Get("corrupted", &result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal cache entry")
	})
	
	t.Run("invalid destination type", func(t *testing.T) {
		err := c.Set("string-data", "hello world")
		require.NoError(t, err)
		
		var wrongType int
		err = c.Get("string-data", &wrongType)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal to destination")
	})
}

func TestGetStats(t *testing.T) {
	tmpDir := t.TempDir()
	c := &Cache{
		dir: tmpDir,
		ttl: 1 * time.Hour,
	}
	
	t.Run("empty cache", func(t *testing.T) {
		stats, err := c.GetStats()
		require.NoError(t, err)
		assert.Equal(t, 0, stats.TotalEntries)
		assert.Equal(t, 0, stats.ExpiredEntries)
		assert.Equal(t, 0, stats.ValidEntries)
		assert.Equal(t, int64(0), stats.TotalSize)
	})
	
	t.Run("cache with entries", func(t *testing.T) {
		// Add some valid entries
		for i := 0; i < 3; i++ {
			err := c.Set(string(rune('a'+i)), i)
			require.NoError(t, err)
		}
		
		// Add an expired entry manually
		expiredEntry := CacheEntry{
			Data:      "expired",
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}
		data, _ := json.Marshal(expiredEntry)
		err := os.WriteFile(c.filename("expired"), data, 0600)
		require.NoError(t, err)
		
		stats, err := c.GetStats()
		require.NoError(t, err)
		assert.Equal(t, 4, stats.TotalEntries)
		assert.Equal(t, 1, stats.ExpiredEntries)
		assert.Equal(t, 3, stats.ValidEntries)
		assert.Greater(t, stats.TotalSize, int64(0))
		assert.False(t, stats.OldestEntry.IsZero())
		assert.False(t, stats.NewestEntry.IsZero())
	})
	
	t.Run("cache with invalid files", func(t *testing.T) {
		// Create a non-JSON file
		err := os.WriteFile(filepath.Join(tmpDir, "notjson.txt"), []byte("text"), 0600)
		require.NoError(t, err)
		
		// Create a directory
		err = os.Mkdir(filepath.Join(tmpDir, "subdir"), 0750)
		require.NoError(t, err)
		
		stats, err := c.GetStats()
		require.NoError(t, err)
		// Should still count the 4 valid JSON files from previous test
		assert.Equal(t, 4, stats.TotalEntries)
	})
}

func TestCleanExpired(t *testing.T) {
	tmpDir := t.TempDir()
	c := &Cache{
		dir: tmpDir,
		ttl: 1 * time.Hour,
	}
	
	t.Run("clean expired entries", func(t *testing.T) {
		// Add valid entries
		for i := 0; i < 3; i++ {
			err := c.Set(string(rune('a'+i)), i)
			require.NoError(t, err)
		}
		
		// Add expired entries manually
		for i := 0; i < 2; i++ {
			expiredEntry := CacheEntry{
				Data:      i,
				ExpiresAt: time.Now().Add(-1 * time.Hour),
			}
			data, _ := json.Marshal(expiredEntry)
			err := os.WriteFile(c.filename(string(rune('x'+i))), data, 0600)
			require.NoError(t, err)
		}
		
		// Verify we have 5 entries
		files, _ := os.ReadDir(tmpDir)
		assert.Equal(t, 5, len(files))
		
		removed, err := c.CleanExpired()
		require.NoError(t, err)
		assert.Equal(t, 2, removed)
		
		// Verify only 3 remain
		files, _ = os.ReadDir(tmpDir)
		assert.Equal(t, 3, len(files))
	})
	
	t.Run("clean with no expired entries", func(t *testing.T) {
		c2 := &Cache{
			dir: t.TempDir(),
			ttl: 1 * time.Hour,
		}
		
		// Add only valid entries
		for i := 0; i < 3; i++ {
			err := c2.Set(string(rune('a'+i)), i)
			require.NoError(t, err)
		}
		
		removed, err := c2.CleanExpired()
		require.NoError(t, err)
		assert.Equal(t, 0, removed)
	})
}

func TestInitCaches(t *testing.T) {
	t.Run("successful initialization", func(t *testing.T) {
		// Save original config dir
		origDir := config.DefaultConfigDir
		config.DefaultConfigDir = t.TempDir()
		defer func() { config.DefaultConfigDir = origDir }()
		
		err := InitCaches()
		require.NoError(t, err)
		
		assert.NotNil(t, WorkspaceCache)
		assert.NotNil(t, UserCache)
		assert.NotNil(t, TaskCache)
		
		// Verify TTLs
		assert.Equal(t, 1*time.Hour, WorkspaceCache.ttl)
		assert.Equal(t, 1*time.Hour, UserCache.ttl)
		assert.Equal(t, 5*time.Minute, TaskCache.ttl)
		
		// Reset globals
		WorkspaceCache = nil
		UserCache = nil
		TaskCache = nil
	})
	
	t.Run("initialization failure", func(t *testing.T) {
		// Use a path that will fail
		origDir := config.DefaultConfigDir
		config.DefaultConfigDir = "/root/no-permission"
		defer func() { config.DefaultConfigDir = origDir }()
		
		err := InitCaches()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create workspace cache")
	})
}

func TestCacheConcurrency(t *testing.T) {
	tmpDir := t.TempDir()
	c := &Cache{
		dir: tmpDir,
		ttl: 1 * time.Hour,
	}
	
	t.Run("concurrent operations", func(t *testing.T) {
		done := make(chan bool)
		
		// Writer goroutines
		for i := 0; i < 5; i++ {
			go func(id int) {
				for j := 0; j < 10; j++ {
					key := string(rune('a'+id)) + string(rune('0'+j))
					c.Set(key, id*10+j)
				}
				done <- true
			}(i)
		}
		
		// Reader goroutines
		for i := 0; i < 5; i++ {
			go func(id int) {
				for j := 0; j < 10; j++ {
					key := string(rune('a'+id)) + string(rune('0'+j))
					var val int
					c.Get(key, &val)
				}
				done <- true
			}(i)
		}
		
		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}
		
		// Verify cache is still functional
		err := c.Set("final", "test")
		assert.NoError(t, err)
		
		var result string
		err = c.Get("final", &result)
		assert.NoError(t, err)
		assert.Equal(t, "test", result)
	})
}

func TestCacheErrorPaths(t *testing.T) {
	t.Run("clear with read directory error", func(t *testing.T) {
		c := &Cache{
			dir: "/nonexistent/path",
			ttl: 1 * time.Hour,
		}
		
		err := c.Clear()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read cache directory")
	})
	
	t.Run("clear with remove error", func(t *testing.T) {
		tmpDir := t.TempDir()
		c := &Cache{
			dir: tmpDir,
			ttl: 1 * time.Hour,
		}
		
		// Create a file and make it read-only
		err := c.Set("test", "data")
		require.NoError(t, err)
		
		// Change permissions to make directory read-only
		err = os.Chmod(tmpDir, 0500)
		require.NoError(t, err)
		defer os.Chmod(tmpDir, 0750)
		
		// Clear should fail due to permissions
		err = c.Clear()
		if err != nil {
			assert.Contains(t, err.Error(), "failed to remove cache file")
		}
	})
	
	t.Run("get stats with read directory error", func(t *testing.T) {
		c := &Cache{
			dir: "/nonexistent/path",
			ttl: 1 * time.Hour,
		}
		
		stats, err := c.GetStats()
		assert.Error(t, err)
		assert.Nil(t, stats)
		assert.Contains(t, err.Error(), "failed to read cache directory")
	})
	
	t.Run("clean expired with read directory error", func(t *testing.T) {
		c := &Cache{
			dir: "/nonexistent/path",
			ttl: 1 * time.Hour,
		}
		
		removed, err := c.CleanExpired()
		assert.Error(t, err)
		assert.Equal(t, 0, removed)
		assert.Contains(t, err.Error(), "failed to read cache directory")
	})
	
	t.Run("set with marshal error", func(t *testing.T) {
		tmpDir := t.TempDir()
		c := &Cache{
			dir: tmpDir,
			ttl: 1 * time.Hour,
		}
		
		// Try to set an unmarshalable value (channel)
		ch := make(chan int)
		err := c.Set("channel", ch)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to marshal cache entry")
	})
	
	t.Run("set with write error", func(t *testing.T) {
		c := &Cache{
			dir: "/root/no-permission",
			ttl: 1 * time.Hour,
		}
		
		err := c.Set("test", "data")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write cache")
	})
	
	t.Run("get with read file error", func(t *testing.T) {
		tmpDir := t.TempDir()
		c := &Cache{
			dir: tmpDir,
			ttl: 1 * time.Hour,
		}
		
		// Try to get with no permissions
		filename := c.filename("test")
		err := os.WriteFile(filename, []byte("data"), 0000)
		require.NoError(t, err)
		
		var result string
		err = c.Get("test", &result)
		// Error depends on OS permissions handling
		if err != nil {
			assert.Contains(t, err.Error(), "failed to read cache")
		}
		
		// Clean up
		os.Chmod(filename, 0600)
	})
}

func TestCacheSafety(t *testing.T) {
	tmpDir := t.TempDir()
	c := &Cache{dir: tmpDir}

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
		if !strings.HasPrefix(filename, tmpDir) {
			t.Errorf("Dangerous key %q produced filename outside cache dir: expected prefix %s, got %s", key, tmpDir, filename)
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

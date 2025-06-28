package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/tim/cu/internal/config"
)

// Cache represents a simple file-based cache
type Cache struct {
	mu  sync.RWMutex
	dir string
	ttl time.Duration
}

// CacheEntry represents a cached item with metadata
type CacheEntry struct {
	Data      interface{} `json:"data"`
	ExpiresAt time.Time   `json:"expires_at"`
}

// NewCache creates a new cache instance
func NewCache(ttl time.Duration) (*Cache, error) {
	cacheDir := filepath.Join(config.DefaultConfigDir, "cache")
	if err := os.MkdirAll(cacheDir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &Cache{
		dir: cacheDir,
		ttl: ttl,
	}, nil
}

// Get retrieves an item from the cache
func (c *Cache) Get(key string, dest interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	filename := c.filename(key)
	data, err := os.ReadFile(filename) // #nosec G304 - filename is generated from SHA256 hash
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("cache miss: %s", key)
		}
		return fmt.Errorf("failed to read cache: %w", err)
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return fmt.Errorf("failed to unmarshal cache entry: %w", err)
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		return fmt.Errorf("cache expired: %s", key)
	}

	// Marshal data to JSON then unmarshal to destination
	jsonData, err := json.Marshal(entry.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal cached data: %w", err)
	}

	if err := json.Unmarshal(jsonData, dest); err != nil {
		return fmt.Errorf("failed to unmarshal to destination: %w", err)
	}

	return nil
}

// Set stores an item in the cache
func (c *Cache) Set(key string, value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := CacheEntry{
		Data:      value,
		ExpiresAt: time.Now().Add(c.ttl),
	}

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}

	filename := c.filename(key)
	if err := os.WriteFile(filename, data, 0600); err != nil {
		return fmt.Errorf("failed to write cache: %w", err)
	}

	return nil
}

// Delete removes an item from the cache
func (c *Cache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	filename := c.filename(key)
	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete cache: %w", err)
	}

	return nil
}

// Clear removes all items from the cache
func (c *Cache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			path := filepath.Join(c.dir, entry.Name())
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("failed to remove cache file %s: %w", entry.Name(), err)
			}
		}
	}

	return nil
}

// filename generates a safe filename for a cache key using SHA256
func (c *Cache) filename(key string) string {
	// Use SHA256 to generate a safe filename from the key
	hash := sha256.Sum256([]byte(key))
	filename := hex.EncodeToString(hash[:])
	return filepath.Join(c.dir, filename+".json")
}

// Stats represents cache statistics
type Stats struct {
	TotalEntries   int
	ExpiredEntries int
	ValidEntries   int
	TotalSize      int64
	OldestEntry    time.Time
	NewestEntry    time.Time
}

// GetStats returns statistics about the cache
func (c *Cache) GetStats() (*Stats, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := &Stats{}
	
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache directory: %w", err)
	}

	now := time.Now()
	
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			stats.TotalEntries++
			
			info, err := entry.Info()
			if err != nil {
				continue
			}
			
			stats.TotalSize += info.Size()
			modTime := info.ModTime()
			
			// Track oldest and newest
			if stats.OldestEntry.IsZero() || modTime.Before(stats.OldestEntry) {
				stats.OldestEntry = modTime
			}
			if modTime.After(stats.NewestEntry) {
				stats.NewestEntry = modTime
			}
			
			// Check if expired
			path := filepath.Join(c.dir, entry.Name())
			data, err := os.ReadFile(path) // #nosec G304 - path is constructed from directory listing
			if err != nil {
				continue
			}
			
			var cacheEntry CacheEntry
			if err := json.Unmarshal(data, &cacheEntry); err != nil {
				continue
			}
			
			if now.After(cacheEntry.ExpiresAt) {
				stats.ExpiredEntries++
			} else {
				stats.ValidEntries++
			}
		}
	}
	
	return stats, nil
}

// CleanExpired removes expired entries from the cache
func (c *Cache) CleanExpired() (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return 0, fmt.Errorf("failed to read cache directory: %w", err)
	}

	now := time.Now()
	removed := 0
	
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			path := filepath.Join(c.dir, entry.Name())
			
			data, err := os.ReadFile(path) // #nosec G304 - path is constructed from directory listing
			if err != nil {
				continue
			}
			
			var cacheEntry CacheEntry
			if err := json.Unmarshal(data, &cacheEntry); err != nil {
				continue
			}
			
			if now.After(cacheEntry.ExpiresAt) {
				if err := os.Remove(path); err == nil {
					removed++
				}
			}
		}
	}
	
	return removed, nil
}

// Global cache instances with different TTLs
var (
	// WorkspaceCache for workspace structure (1 hour)
	WorkspaceCache *Cache
	// UserCache for user list (1 hour)
	UserCache *Cache
	// TaskCache for recent tasks (5 minutes)
	TaskCache *Cache
)

// InitCaches initializes the global cache instances
func InitCaches() error {
	var err error

	WorkspaceCache, err = NewCache(1 * time.Hour)
	if err != nil {
		return fmt.Errorf("failed to create workspace cache: %w", err)
	}

	UserCache, err = NewCache(1 * time.Hour)
	if err != nil {
		return fmt.Errorf("failed to create user cache: %w", err)
	}

	TaskCache, err = NewCache(5 * time.Minute)
	if err != nil {
		return fmt.Errorf("failed to create task cache: %w", err)
	}

	return nil
}

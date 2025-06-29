package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/tim/cu/internal/cache"
	"github.com/tim/cu/internal/output"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage local cache",
	Long:  `Manage the local cache used to improve performance.`,
}

var cacheInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show cache information and statistics",
	Long:  `Display detailed information about cache usage, including size, entry count, and expiration status.`,
	RunE:  showCacheInfo,
}

var cacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all cache entries",
	Long:  `Remove all cached data. This will force fresh API calls on next use.`,
	RunE:  clearCache,
}

var cacheCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove expired cache entries",
	Long:  `Remove only expired cache entries, keeping valid cached data.`,
	RunE:  cleanCache,
}

func init() {
	rootCmd.AddCommand(cacheCmd)
	cacheCmd.AddCommand(cacheInfoCmd)
	cacheCmd.AddCommand(cacheClearCmd)
	cacheCmd.AddCommand(cacheCleanCmd)
}

func showCacheInfo(cmd *cobra.Command, args []string) error {
	// Initialize caches if not already done
	if err := cache.InitCaches(); err != nil {
		return fmt.Errorf("failed to initialize caches: %w", err)
	}

	type cacheInfo struct {
		Name           string        `json:"name" yaml:"name"`
		TTL            time.Duration `json:"ttl" yaml:"ttl"`
		TotalEntries   int           `json:"total_entries" yaml:"total_entries"`
		ValidEntries   int           `json:"valid_entries" yaml:"valid_entries"`
		ExpiredEntries int           `json:"expired_entries" yaml:"expired_entries"`
		TotalSize      int64         `json:"total_size_bytes" yaml:"total_size_bytes"`
		SizeHuman      string        `json:"size_human" yaml:"size_human"`
		OldestEntry    string        `json:"oldest_entry" yaml:"oldest_entry"`
		NewestEntry    string        `json:"newest_entry" yaml:"newest_entry"`
	}

	var allCacheInfo []cacheInfo
	var totalSize int64
	var totalEntries, totalValid, totalExpired int

	// Get stats for each cache type
	caches := []struct {
		name  string
		cache *cache.Cache
		ttl   time.Duration
	}{
		{"Workspace", cache.WorkspaceCache, 1 * time.Hour},
		{"User", cache.UserCache, 1 * time.Hour},
		{"Task", cache.TaskCache, 5 * time.Minute},
	}

	for _, c := range caches {
		stats, err := c.cache.GetStats()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to get stats for %s cache: %v\n", c.name, err)
			continue
		}

		info := cacheInfo{
			Name:           c.name,
			TTL:            c.ttl,
			TotalEntries:   stats.TotalEntries,
			ValidEntries:   stats.ValidEntries,
			ExpiredEntries: stats.ExpiredEntries,
			TotalSize:      stats.TotalSize,
			SizeHuman:      formatBytes(stats.TotalSize),
		}

		if !stats.OldestEntry.IsZero() {
			info.OldestEntry = formatCacheTime(stats.OldestEntry)
		}
		if !stats.NewestEntry.IsZero() {
			info.NewestEntry = formatCacheTime(stats.NewestEntry)
		}

		allCacheInfo = append(allCacheInfo, info)
		
		totalSize += stats.TotalSize
		totalEntries += stats.TotalEntries
		totalValid += stats.ValidEntries
		totalExpired += stats.ExpiredEntries
	}

	// Output based on format
	if outputFormat == "json" || outputFormat == "yaml" {
		result := map[string]interface{}{
			"caches":         allCacheInfo,
			"total_size":     totalSize,
			"total_entries":  totalEntries,
			"valid_entries":  totalValid,
			"expired_entries": totalExpired,
		}
		return output.Format(outputFormat, result)
	}

	// Table output
	fmt.Println("Cache Information")
	fmt.Println("=================")
	fmt.Println()

	for _, info := range allCacheInfo {
		fmt.Printf("%s Cache:\n", info.Name)
		fmt.Printf("  TTL:             %v\n", info.TTL)
		fmt.Printf("  Total Entries:   %d\n", info.TotalEntries)
		fmt.Printf("  Valid Entries:   %d\n", info.ValidEntries)
		fmt.Printf("  Expired Entries: %d\n", info.ExpiredEntries)
		fmt.Printf("  Size:            %s\n", info.SizeHuman)
		if info.OldestEntry != "" {
			fmt.Printf("  Oldest Entry:    %s\n", info.OldestEntry)
		}
		if info.NewestEntry != "" {
			fmt.Printf("  Newest Entry:    %s\n", info.NewestEntry)
		}
		fmt.Println()
	}

	fmt.Printf("Total Cache Usage:\n")
	fmt.Printf("  Total Entries:   %d\n", totalEntries)
	fmt.Printf("  Valid Entries:   %d\n", totalValid)
	fmt.Printf("  Expired Entries: %d\n", totalExpired)
	fmt.Printf("  Total Size:      %s\n", formatBytes(totalSize))

	return nil
}

func clearCache(cmd *cobra.Command, args []string) error {
	// Initialize caches if not already done
	if err := cache.InitCaches(); err != nil {
		return fmt.Errorf("failed to initialize caches: %w", err)
	}

	// Clear each cache
	caches := []struct {
		name  string
		cache *cache.Cache
	}{
		{"Workspace", cache.WorkspaceCache},
		{"User", cache.UserCache},
		{"Task", cache.TaskCache},
	}

	var totalCleared int
	for _, c := range caches {
		stats, _ := c.cache.GetStats()
		before := 0
		if stats != nil {
			before = stats.TotalEntries
		}

		if err := c.cache.Clear(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to clear %s cache: %v\n", c.name, err)
			continue
		}

		if before > 0 {
			fmt.Printf("Cleared %d entries from %s cache\n", before, c.name)
			totalCleared += before
		}
	}

	if totalCleared > 0 {
		fmt.Printf("\nTotal cleared: %d cache entries\n", totalCleared)
	} else {
		fmt.Println("No cache entries to clear")
	}

	return nil
}

func cleanCache(cmd *cobra.Command, args []string) error {
	// Initialize caches if not already done
	if err := cache.InitCaches(); err != nil {
		return fmt.Errorf("failed to initialize caches: %w", err)
	}

	// Clean expired entries from each cache
	caches := []struct {
		name  string
		cache *cache.Cache
	}{
		{"Workspace", cache.WorkspaceCache},
		{"User", cache.UserCache},
		{"Task", cache.TaskCache},
	}

	var totalCleaned int
	for _, c := range caches {
		removed, err := c.cache.CleanExpired()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to clean %s cache: %v\n", c.name, err)
			continue
		}

		if removed > 0 {
			fmt.Printf("Removed %d expired entries from %s cache\n", removed, c.name)
			totalCleaned += removed
		}
	}

	if totalCleaned > 0 {
		fmt.Printf("\nTotal removed: %d expired entries\n", totalCleaned)
	} else {
		fmt.Println("No expired entries to remove")
	}

	return nil
}

// Helper functions

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatCacheTime(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	
	duration := time.Since(t)
	if duration < 0 {
		return t.Format("2006-01-02 15:04:05")
	}
	
	switch {
	case duration < time.Minute:
		return "just now"
	case duration < time.Hour:
		return fmt.Sprintf("%d minutes ago", int(duration.Minutes()))
	case duration < 24*time.Hour:
		return fmt.Sprintf("%d hours ago", int(duration.Hours()))
	case duration < 7*24*time.Hour:
		return fmt.Sprintf("%d days ago", int(duration.Hours()/24))
	default:
		return t.Format("2006-01-02")
	}
}
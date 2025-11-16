package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/zeromicro/mcp-zero/internal/analyzer"
	"github.com/zeromicro/mcp-zero/internal/responses"
)

type AnalyzeProjectParams struct {
	ProjectPath string `json:"project_path"`
}

// Cache for project analysis results with automatic cleanup
type analysisCache struct {
	mu      sync.RWMutex
	entries map[string]*cacheEntry
}

type cacheEntry struct {
	analysis  *analyzer.ProjectAnalysis
	timestamp time.Time
	hits      int // Track cache hits for optimization metrics
}

var cache = &analysisCache{
	entries: make(map[string]*cacheEntry),
}

const (
	cacheTTL             = 5 * time.Minute
	cacheCleanupInterval = 10 * time.Minute
	maxCacheEntries      = 100 // Prevent unbounded memory growth
)

func init() {
	// Start background cache cleanup goroutine
	go func() {
		ticker := time.NewTicker(cacheCleanupInterval)
		defer ticker.Stop()

		for range ticker.C {
			cleanupExpiredEntries()
		}
	}()
}

func AnalyzeProject(ctx context.Context, req *mcp.CallToolRequest, params AnalyzeProjectParams) (*mcp.CallToolResult, any, error) {
	projectPath := params.ProjectPath
	if projectPath == "" {
		cwd, _ := os.Getwd()
		projectPath = cwd
	}

	if !filepath.IsAbs(projectPath) {
		absPath, err := filepath.Abs(projectPath)
		if err != nil {
			return responses.FormatError(fmt.Sprintf("failed to resolve project path: %v", err))
		}
		projectPath = absPath
	}

	// Check cache
	if cachedAnalysis := getCachedAnalysis(projectPath); cachedAnalysis != nil {
		return formatAnalysisResult(cachedAnalysis, true)
	}

	// Perform analysis
	analysis, err := analyzer.ScanProject(projectPath)
	if err != nil {
		return responses.FormatError(fmt.Sprintf("failed to analyze project: %v", err))
	}

	// Cache the result
	cacheAnalysis(projectPath, analysis)

	return formatAnalysisResult(analysis, false)
}

func getCachedAnalysis(projectPath string) *analyzer.ProjectAnalysis {
	cache.mu.Lock() // Use write lock to update hits counter
	defer cache.mu.Unlock()

	entry, exists := cache.entries[projectPath]
	if !exists {
		return nil
	}

	// Check if cache is still valid
	if time.Since(entry.timestamp) > cacheTTL {
		// Remove expired entry immediately
		delete(cache.entries, projectPath)
		return nil
	}

	// Increment hit counter for metrics
	entry.hits++

	return entry.analysis
}

func cacheAnalysis(projectPath string, analysis *analyzer.ProjectAnalysis) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	// Enforce max cache size - evict least recently used entries
	if len(cache.entries) >= maxCacheEntries {
		evictOldestEntry()
	}

	cache.entries[projectPath] = &cacheEntry{
		analysis:  analysis,
		timestamp: time.Now(),
		hits:      0,
	}
}

// cleanupExpiredEntries removes stale cache entries in background
func cleanupExpiredEntries() {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	now := time.Now()
	for path, entry := range cache.entries {
		if now.Sub(entry.timestamp) > cacheTTL {
			delete(cache.entries, path)
		}
	}
}

// evictOldestEntry removes the oldest or least used cache entry
// Must be called with cache.mu locked
func evictOldestEntry() {
	if len(cache.entries) == 0 {
		return
	}

	var oldestPath string
	var oldestTime time.Time
	var minHits = -1

	// Find entry with lowest hits, or oldest if tied
	for path, entry := range cache.entries {
		if minHits == -1 || entry.hits < minHits ||
			(entry.hits == minHits && (oldestTime.IsZero() || entry.timestamp.Before(oldestTime))) {
			oldestPath = path
			oldestTime = entry.timestamp
			minHits = entry.hits
		}
	}

	if oldestPath != "" {
		delete(cache.entries, oldestPath)
	}
}

// GetCacheStats returns cache statistics for monitoring
func GetCacheStats() map[string]interface{} {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	totalHits := 0
	oldestEntry := time.Now()

	for _, entry := range cache.entries {
		totalHits += entry.hits
		if entry.timestamp.Before(oldestEntry) {
			oldestEntry = entry.timestamp
		}
	}

	return map[string]interface{}{
		"total_entries": len(cache.entries),
		"total_hits":    totalHits,
		"oldest_entry":  oldestEntry,
		"max_capacity":  maxCacheEntries,
	}
}

// ClearCache removes all cached entries (useful for testing)
func ClearCache() {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.entries = make(map[string]*cacheEntry)
}

func formatAnalysisResult(analysis *analyzer.ProjectAnalysis, fromCache bool) (*mcp.CallToolResult, any, error) {
	var message strings.Builder

	message.WriteString(fmt.Sprintf("Project Analysis: %s\n\n", analysis.ProjectPath))

	if fromCache {
		message.WriteString("(Results from cache)\n\n")
	}

	// Summary section
	message.WriteString("=== Summary ===\n")
	message.WriteString(fmt.Sprintf("Total Services: %d\n", analysis.Summary.TotalServices))
	message.WriteString(fmt.Sprintf("  - API Services: %d\n", analysis.Summary.APIServices))
	message.WriteString(fmt.Sprintf("  - RPC Services: %d\n", analysis.Summary.RPCServices))
	message.WriteString(fmt.Sprintf("Total Endpoints: %d\n", analysis.Summary.TotalEndpoints))
	message.WriteString(fmt.Sprintf("Total RPC Methods: %d\n", analysis.Summary.TotalRPCMethods))
	message.WriteString(fmt.Sprintf("Dependencies: %d\n", analysis.Summary.TotalDependencies))
	if analysis.Summary.GoZeroVersion != "" {
		message.WriteString(fmt.Sprintf("Go-Zero Version: %s\n", analysis.Summary.GoZeroVersion))
	}
	message.WriteString("\n")

	// Services section
	if len(analysis.Services) > 0 {
		message.WriteString("=== Services ===\n")
		for i, service := range analysis.Services {
			message.WriteString(fmt.Sprintf("\n%d. %s (%s)\n", i+1, service.Name, service.Type))
			message.WriteString(fmt.Sprintf("   Path: %s\n", service.Path))
			message.WriteString(fmt.Sprintf("   Spec: %s\n", service.SpecFile))

			if service.Type == "api" && len(service.Endpoints) > 0 {
				message.WriteString("   Endpoints:\n")
				for _, endpoint := range service.Endpoints {
					message.WriteString(fmt.Sprintf("     - %s %s (%s)\n",
						strings.ToUpper(endpoint.Method), endpoint.Path, endpoint.Handler))
				}
			}

			if service.Type == "rpc" && len(service.RPCMethods) > 0 {
				message.WriteString("   RPC Methods:\n")
				for _, method := range service.RPCMethods {
					streamInfo := ""
					if method.Stream {
						streamInfo = " [stream]"
					}
					message.WriteString(fmt.Sprintf("     - %s(%s) returns %s%s\n",
						method.Name, method.Request, method.Response, streamInfo))
				}
			}
		}
		message.WriteString("\n")
	}

	// Dependencies section
	if len(analysis.Dependencies) > 0 {
		message.WriteString("=== Key Dependencies ===\n")
		directDeps := 0
		for _, dep := range analysis.Dependencies {
			if dep.Type == "direct" {
				directDeps++
				if directDeps <= 10 { // Show only first 10 direct deps
					message.WriteString(fmt.Sprintf("  - %s %s\n", dep.Name, dep.Version))
				}
			}
		}
		if directDeps > 10 {
			message.WriteString(fmt.Sprintf("  ... and %d more direct dependencies\n", directDeps-10))
		}
		message.WriteString("\n")
	}

	// Config files section
	if len(analysis.Configs) > 0 {
		message.WriteString("=== Configuration Files ===\n")
		for _, config := range analysis.Configs {
			relPath, _ := filepath.Rel(analysis.ProjectPath, config.Path)
			message.WriteString(fmt.Sprintf("  - %s (%s)\n", relPath, config.Type))
		}
		message.WriteString("\n")
	}

	// Next steps
	message.WriteString("=== Next Steps ===\n")
	message.WriteString("  - Use generate_api_from_spec to update API services\n")
	message.WriteString("  - Use create_rpc_service to add new RPC services\n")
	message.WriteString("  - Use generate_model to add database models\n")

	data := map[string]any{
		"project_path":      analysis.ProjectPath,
		"total_services":    analysis.Summary.TotalServices,
		"api_services":      analysis.Summary.APIServices,
		"rpc_services":      analysis.Summary.RPCServices,
		"total_endpoints":   analysis.Summary.TotalEndpoints,
		"total_rpc_methods": analysis.Summary.TotalRPCMethods,
		"dependencies":      analysis.Summary.TotalDependencies,
		"go_zero_version":   analysis.Summary.GoZeroVersion,
		"from_cache":        fromCache,
	}

	return responses.FormatSuccessWithData(message.String(), data)
}

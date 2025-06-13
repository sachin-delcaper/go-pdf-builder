package cache

import (
	"crypto/md5"
	"fmt"
	"os"
	"sync"
	"time"

	"pdf-gen-simple/internal/models"
)

// TemplateCache provides caching for parsed CSV templates
type TemplateCache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
	maxSize int
	ttl     time.Duration
}

// CacheEntry represents a cached template entry
type CacheEntry struct {
	Elements    []models.PDFElement
	Hash        string
	CreatedAt   time.Time
	AccessedAt  time.Time
	FileModTime time.Time
}

// FontCache provides caching for font resources
type FontCache struct {
	mu     sync.RWMutex
	fonts  map[string]bool
	loaded bool
}

var (
	defaultTemplateCache *TemplateCache
	defaultFontCache     *FontCache
	once                 sync.Once
)

// GetTemplateCache returns the global template cache instance
func GetTemplateCache() *TemplateCache {
	once.Do(func() {
		defaultTemplateCache = NewTemplateCache(100, 30*time.Minute)
		defaultFontCache = NewFontCache()
	})
	return defaultTemplateCache
}

// GetFontCache returns the global font cache instance
func GetFontCache() *FontCache {
	once.Do(func() {
		defaultTemplateCache = NewTemplateCache(100, 30*time.Minute)
		defaultFontCache = NewFontCache()
	})
	return defaultFontCache
}

// NewTemplateCache creates a new template cache
func NewTemplateCache(maxSize int, ttl time.Duration) *TemplateCache {
	cache := &TemplateCache{
		entries: make(map[string]*CacheEntry),
		maxSize: maxSize,
		ttl:     ttl,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Get retrieves a template from cache if valid
func (tc *TemplateCache) Get(filePath string) ([]models.PDFElement, bool) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	entry, exists := tc.entries[filePath]
	if !exists {
		return nil, false
	}

	// Check if file has been modified
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		// File doesn't exist anymore, remove from cache
		delete(tc.entries, filePath)
		return nil, false
	}

	if fileInfo.ModTime().After(entry.FileModTime) {
		// File has been modified, invalidate cache
		delete(tc.entries, filePath)
		return nil, false
	}

	// Check TTL
	if time.Since(entry.CreatedAt) > tc.ttl {
		delete(tc.entries, filePath)
		return nil, false
	}

	// Update access time
	entry.AccessedAt = time.Now()

	return entry.Elements, true
}

// Set stores a template in cache
func (tc *TemplateCache) Set(filePath string, elements []models.PDFElement) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return // Can't cache if we can't stat the file
	}

	// Calculate hash for integrity checking
	hash := tc.calculateHash(elements)

	entry := &CacheEntry{
		Elements:    elements,
		Hash:        hash,
		CreatedAt:   time.Now(),
		AccessedAt:  time.Now(),
		FileModTime: fileInfo.ModTime(),
	}

	tc.entries[filePath] = entry

	// Evict oldest entries if cache is full
	if len(tc.entries) > tc.maxSize {
		tc.evictOldest()
	}
}

// calculateHash creates a hash of the elements for integrity checking
func (tc *TemplateCache) calculateHash(elements []models.PDFElement) string {
	h := md5.New()
	for _, elem := range elements {
		h.Write([]byte(elem.String()))
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// evictOldest removes the least recently accessed entry
func (tc *TemplateCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range tc.entries {
		if oldestKey == "" || entry.AccessedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.AccessedAt
		}
	}

	if oldestKey != "" {
		delete(tc.entries, oldestKey)
	}
}

// cleanup periodically removes expired entries
func (tc *TemplateCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		tc.mu.Lock()
		now := time.Now()

		for key, entry := range tc.entries {
			if now.Sub(entry.CreatedAt) > tc.ttl {
				delete(tc.entries, key)
			}
		}

		tc.mu.Unlock()
	}
}

// Clear removes all entries from cache
func (tc *TemplateCache) Clear() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.entries = make(map[string]*CacheEntry)
}

// Stats returns cache statistics
func (tc *TemplateCache) Stats() map[string]interface{} {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	return map[string]interface{}{
		"entries": len(tc.entries),
		"maxSize": tc.maxSize,
		"ttl":     tc.ttl.String(),
	}
}

// NewFontCache creates a new font cache
func NewFontCache() *FontCache {
	return &FontCache{
		fonts: make(map[string]bool),
	}
}

// IsLoaded checks if a font is already loaded
func (fc *FontCache) IsLoaded(fontName string) bool {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return fc.fonts[fontName]
}

// MarkLoaded marks a font as loaded
func (fc *FontCache) MarkLoaded(fontName string) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.fonts[fontName] = true
}

// Clear clears the font cache
func (fc *FontCache) Clear() {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.fonts = make(map[string]bool)
	fc.loaded = false
}

// IsSystemLoaded checks if the font system is loaded
func (fc *FontCache) IsSystemLoaded() bool {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return fc.loaded
}

// MarkSystemLoaded marks the font system as loaded
func (fc *FontCache) MarkSystemLoaded() {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.loaded = true
}

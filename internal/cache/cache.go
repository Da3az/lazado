package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// DiskCache provides a TTL-based disk cache for JSON-serializable data.
type DiskCache struct {
	dir string
	ttl time.Duration
}

// New creates a DiskCache with the given directory and TTL in seconds.
func New(dir string, ttlSeconds int) *DiskCache {
	return &DiskCache{
		dir: dir,
		ttl: time.Duration(ttlSeconds) * time.Second,
	}
}

// Get reads a cached value. Returns false if the key doesn't exist or is expired.
// If allowStale is true, returns expired entries as a fallback.
func (c *DiskCache) Get(key string, dest interface{}, allowStale bool) bool {
	path := c.path(key)
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	expired := time.Since(info.ModTime()) > c.ttl
	if expired && !allowStale {
		return false
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	return json.Unmarshal(data, dest) == nil
}

// Set writes a value to the cache.
func (c *DiskCache) Set(key string, value interface{}) error {
	if err := os.MkdirAll(c.dir, 0700); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return os.WriteFile(c.path(key), data, 0600)
}

// IsFresh checks if a cached key exists and is not expired.
func (c *DiskCache) IsFresh(key string) bool {
	info, err := os.Stat(c.path(key))
	if err != nil {
		return false
	}
	return time.Since(info.ModTime()) <= c.ttl
}

// Delete removes a cached key.
func (c *DiskCache) Delete(key string) {
	os.Remove(c.path(key))
}

func (c *DiskCache) path(key string) string {
	return filepath.Join(c.dir, key+".json")
}

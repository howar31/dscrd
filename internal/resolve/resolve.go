// Package resolve maps Discord snowflake IDs to human-readable names (and
// caches the reverse lookups commands perform) with a local disk cache.
package resolve

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Lookup fetches the display name for an ID of the given kind
// (user|channel|guild|role).
type Lookup interface {
	Name(kind, id string) (string, error)
}

// Resolver caches kind:id -> name lookups on disk.
type Resolver struct {
	path   string
	lookup Lookup
	mu     sync.Mutex
	cache  map[string]string
}

// New returns a Resolver backed by cachePath and the given Lookup.
func New(cachePath string, lookup Lookup) *Resolver {
	r := &Resolver{path: cachePath, lookup: lookup, cache: map[string]string{}}
	if data, err := os.ReadFile(cachePath); err == nil {
		_ = json.Unmarshal(data, &r.cache)
	}
	return r
}

// Resolve returns the name for id, or id itself if lookup fails.
func (r *Resolver) Resolve(kind, id string) string {
	key := kind + ":" + id
	r.mu.Lock()
	defer r.mu.Unlock()
	if name, ok := r.cache[key]; ok {
		return name
	}
	name, err := r.lookup.Name(kind, id)
	if err != nil {
		return id // graceful degradation
	}
	r.cache[key] = name
	r.flush()
	return name
}

// flush persists the in-memory cache to disk; errors are silently ignored
// because the cache is an optional best-effort optimization.
func (r *Resolver) flush() {
	if err := os.MkdirAll(filepath.Dir(r.path), 0o700); err != nil {
		return
	}
	data, err := json.Marshal(r.cache)
	if err != nil {
		return
	}
	_ = os.WriteFile(r.path, data, 0o600)
}

// IsSnowflake reports whether s looks like a Discord snowflake ID
// (17-20 decimal digits).
func IsSnowflake(s string) bool {
	if len(s) < 17 || len(s) > 20 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

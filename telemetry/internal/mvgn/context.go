package mvgn

import (
	"encoding/json"
	"os"
	"time"
)

// ProjectContext holds the current state of an MVGN project read from .mvgn-context.json.
type ProjectContext struct {
	State      string `json:"state"`
	Progress   string `json:"progress"`
	ActiveTask string `json:"active_task"`
	Blockers   string `json:"blockers"`
}

// RuntimeManifest describes the runtime capabilities and profile read from mvgn-runtime.json.
type RuntimeManifest struct {
	RuntimeVersion string            `json:"runtime_version"`
	Profile        string            `json:"profile"`
	Capabilities   map[string]bool   `json:"capabilities"`
	Actors         []ActorRef        `json:"actors"`
	Skills         []string          `json:"skills"`
}

// ActorRef represents a single actor declared in the runtime manifest.
type ActorRef struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Model    string `json:"model"`
	Provider string `json:"provider"`
}

// cacheEntry holds a cached value with its expiry time.
type cacheEntry struct {
	data      interface{}
	expiresAt time.Time
}

// ContextReader reads MVGN project context, runtime manifest, and state reports
// with TTL-based caching (default 5s). All read methods degrade gracefully on
// missing files by returning sensible defaults.
type ContextReader struct {
	projectPath string
	cache       map[string]cacheEntry
	ttl         time.Duration
}

// NewContextReader creates a ContextReader rooted at projectPath.
func NewContextReader(projectPath string) *ContextReader {
	return &ContextReader{
		projectPath: projectPath,
		cache:       make(map[string]cacheEntry),
		ttl:         5 * time.Second,
	}
}

// ReadContext reads .mvgn-context.json and returns the project context.
// Returns defaults on missing or malformed file.
func (r *ContextReader) ReadContext() (*ProjectContext, error) {
	if cached, ok := r.getCache("context"); ok {
		return cached.(*ProjectContext), nil
	}

	data, err := os.ReadFile(r.projectPath + "/.mvgn-context.json")
	if err != nil {
		return &ProjectContext{State: "unknown", Progress: "0/0"}, nil
	}

	var ctx ProjectContext
	if err := json.Unmarshal(data, &ctx); err != nil {
		return &ProjectContext{State: "unknown"}, nil
	}

	r.setCache("context", &ctx)
	return &ctx, nil
}

// ReadRuntime returns the runtime manifest from mvgn-runtime.json.
// Tries multiple known locations (root, runtime-state/).
// Returns defaults when missing or malformed.
func (r *ContextReader) ReadRuntime() (*RuntimeManifest, error) {
	if cached, ok := r.getCache("runtime"); ok {
		return cached.(*RuntimeManifest), nil
	}

	// Try multiple locations for mvgn-runtime.json
	locations := []string{
		r.projectPath + "/runtime-state/mvgn-runtime.json",
		r.projectPath + "/mvgn-runtime.json",
	}
	var data []byte
	var err error
	for _, loc := range locations {
		data, err = os.ReadFile(loc)
		if err == nil {
			break
		}
	}
	if err != nil {
		return &RuntimeManifest{Profile: "unknown"}, nil
	}

	var rt RuntimeManifest
	if err := json.Unmarshal(data, &rt); err != nil {
		return &RuntimeManifest{Profile: "unknown"}, nil
	}

	r.setCache("runtime", &rt)
	return &rt, nil
}

func (r *ContextReader) getCache(key string) (interface{}, bool) {
	entry, ok := r.cache[key]
	if !ok || time.Now().After(entry.expiresAt) {
		delete(r.cache, key)
		return nil, false
	}
	return entry.data, true
}

func (r *ContextReader) setCache(key string, data interface{}) {
	r.cache[key] = cacheEntry{data: data, expiresAt: time.Now().Add(r.ttl)}
}
// Package fragkeys implements fragmented cache keys: composite keys built from
// independently versioned tags, where bumping one tag's version invalidates
// every key that includes it without deleting anything.
//
// It is the Go member of the fragmented-keys family (PHP, Python, Django,
// TypeScript, Elixir, Java, Swift) and follows the shared specification in
// docs/fractional-key-definition.md. Key derivation is byte-compatible with
// the PHP, Python, Elixir, and TypeScript ports: the same tags, instances,
// versions, and ordering produce the same MD5 in every one of them.
//
// The compatibility that matters most here is with Django. A Go service and a
// Django service pointed at the same Redis will compute identical keys and can
// read and write each other's entries; see the djangoredis subpackage.
package fragkeys

import (
	"context"
	"sync"
	"time"
)

// CacheHandler is a pluggable backend for tag version storage.
//
// Versions are exchanged as strings rather than floats because that is what
// every port stores — the exact decimal spelling is part of the key input, so
// re-parsing and re-formatting a version must be lossless. See the pyfloat
// package for why the spelling is load-bearing.
type CacheHandler interface {
	// GroupName uniquely identifies this handler for bulk-fetch grouping.
	// Tags sharing a group name are fetched in one GetMulti call.
	GroupName() string

	// Get returns the stored value and whether it was present.
	Get(ctx context.Context, key string) (string, bool, error)

	// Set stores a value. A zero ttl means the backend's default.
	Set(ctx context.Context, key, value string, ttl time.Duration) error

	// GetMulti fetches many keys at once, omitting those not found.
	GetMulti(ctx context.Context, keys []string) (map[string]string, error)
}

// MemoryHandler is an in-process CacheHandler for tests and single-process
// use. It honours TTLs, which the specification permits a memory handler to
// ignore; enforcing them keeps test behaviour closer to production.
//
// It is safe for concurrent use.
type MemoryHandler struct {
	name string

	mu      sync.RWMutex
	entries map[string]memoryEntry
}

type memoryEntry struct {
	value     string
	expiresAt time.Time // zero means no expiry
}

// NewMemoryHandler returns an empty in-memory handler. The name distinguishes
// it from other handlers during bulk-fetch grouping; pass distinct names when
// a single key mixes handlers.
func NewMemoryHandler(name string) *MemoryHandler {
	if name == "" {
		name = "MemoryHandler"
	}
	return &MemoryHandler{name: name, entries: map[string]memoryEntry{}}
}

// GroupName implements CacheHandler.
func (h *MemoryHandler) GroupName() string { return h.name }

// Get implements CacheHandler.
func (h *MemoryHandler) Get(_ context.Context, key string) (string, bool, error) {
	h.mu.RLock()
	e, ok := h.entries[key]
	h.mu.RUnlock()
	if !ok || e.expired() {
		return "", false, nil
	}
	return e.value, true, nil
}

// Set implements CacheHandler.
func (h *MemoryHandler) Set(_ context.Context, key, value string, ttl time.Duration) error {
	e := memoryEntry{value: value}
	if ttl > 0 {
		e.expiresAt = time.Now().Add(ttl)
	}
	h.mu.Lock()
	h.entries[key] = e
	h.mu.Unlock()
	return nil
}

// GetMulti implements CacheHandler.
func (h *MemoryHandler) GetMulti(_ context.Context, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, k := range keys {
		if e, ok := h.entries[k]; ok && !e.expired() {
			out[k] = e.value
		}
	}
	return out, nil
}

// Reset drops every entry. Intended for test teardown.
func (h *MemoryHandler) Reset() {
	h.mu.Lock()
	h.entries = map[string]memoryEntry{}
	h.mu.Unlock()
}

func (e memoryEntry) expired() bool {
	return !e.expiresAt.IsZero() && time.Now().After(e.expiresAt)
}

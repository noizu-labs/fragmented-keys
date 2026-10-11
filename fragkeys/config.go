package fragkeys

import (
	"errors"
	"sync"
)

// DefaultGlobalPrefix matches the specification's default and the value every
// sibling port uses when none is configured.
const DefaultGlobalPrefix = "DefaultPrefix"

// ErrNoDefaultHandler is returned when a tag needs the global handler and none
// has been configured.
var ErrNoDefaultHandler = errors.New("fragkeys: no default cache handler configured")

// Configuration holds the process-wide defaults the specification requires:
// a fallback cache handler and the prefix appended to every tag name.
//
// The prefix is part of the key input, so it must match the value the other
// runtimes use. Django's fragmented_keys_django sets it via
// configure_fragmented_keys(prefix=...); gnp-backend passes "GNP".
//
// A package-level singleton mirrors the other ports, but every tag, key, and
// ring can be given an explicit handler and prefix instead, which is what
// tests and multi-tenant callers should do.
type Configuration struct {
	mu             sync.RWMutex
	defaultHandler CacheHandler
	globalPrefix   string
	epoch          float64
}

// Config is the process-wide configuration singleton.
var Config = &Configuration{globalPrefix: DefaultGlobalPrefix, epoch: DefaultEpoch}

// SetDefaultCacheHandler installs the fallback handler used by tags created
// without an explicit one.
func (c *Configuration) SetDefaultCacheHandler(h CacheHandler) {
	c.mu.Lock()
	c.defaultHandler = h
	c.mu.Unlock()
}

// DefaultCacheHandler returns the fallback handler, or ErrNoDefaultHandler.
func (c *Configuration) DefaultCacheHandler() (CacheHandler, error) {
	c.mu.RLock()
	h := c.defaultHandler
	c.mu.RUnlock()
	if h == nil {
		return nil, ErrNoDefaultHandler
	}
	return h, nil
}

// SetGlobalPrefix sets the suffix appended to every tag name.
//
// Despite the name — kept for cross-port consistency — the value is appended,
// not prepended: a tag name is "{tag}_{instance}{prefix}". Changing it makes
// every previously stored version unreachable, which is a full cache bust.
func (c *Configuration) SetGlobalPrefix(p string) {
	c.mu.Lock()
	c.globalPrefix = p
	c.mu.Unlock()
}

// GlobalPrefix returns the current prefix.
func (c *Configuration) GlobalPrefix() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.globalPrefix
}

// Reset restores the defaults. Intended for test teardown.
func (c *Configuration) Reset() {
	c.mu.Lock()
	c.defaultHandler = nil
	c.globalPrefix = DefaultGlobalPrefix
	c.epoch = DefaultEpoch
	c.mu.Unlock()
}

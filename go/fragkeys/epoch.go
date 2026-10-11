package fragkeys

// Global cache epoch — one constant dimension carried by every composite key.
//
// Because the epoch is a ConstantTag it never touches the cache backend and
// costs nothing to resolve, but changing its version changes every key at once.
// That makes it a single-value global cache bust, for the cases narrow
// per-entity invalidation cannot express:
//
//   - A key-derivation change. When fragmented_keys_django's double-prefix
//     defect was fixed, two previously separate tag namespaces merged; entries
//     written before the fix remained reachable under the merged naming but
//     had been computed under different assumptions. Raising the epoch as the
//     fix ships makes them unreachable instead.
//   - Promoting a build from stage to prod against a shared or restored cache.
//   - A payload schema change that leaves the key shape untouched.
//
// The epoch must be configured identically on every runtime sharing the cache.
// In Django that is settings.CACHES["…"]["EPOCH"] or
// configure_fragmented_keys(epoch=…); here it is Config.SetEpoch or
// djangoredis.WithEpoch.
const (
	// EpochTagName and EpochTagInstance are part of the cross-implementation
	// key contract and must match fragmented_keys_django's
	// cache_utils/epoch.py. Do not change them.
	EpochTagName     = "epoch"
	EpochTagInstance = "global"

	// DefaultEpoch is the version used when none is configured.
	DefaultEpoch = 1.0
)

// SetEpoch sets the process-wide cache epoch. Every key built afterwards
// resolves differently from every key built before.
func (c *Configuration) SetEpoch(epoch float64) {
	c.mu.Lock()
	c.epoch = epoch
	c.mu.Unlock()
}

// Epoch returns the process-wide cache epoch.
func (c *Configuration) Epoch() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.epoch
}

// EpochTag builds the epoch tag at the currently configured epoch.
func EpochTag(opts ...TagOption) *ConstantTag {
	all := append([]TagOption{WithVersion(Config.Epoch())}, opts...)
	return NewConstantTag(EpochTagName, EpochTagInstance, all...)
}

// EpochTagAt builds an epoch tag at an explicit version, ignoring the global
// configuration. Useful for reading a previous generation of entries.
func EpochTagAt(epoch float64, opts ...TagOption) *ConstantTag {
	all := append([]TagOption{WithVersion(epoch)}, opts...)
	return NewConstantTag(EpochTagName, EpochTagInstance, all...)
}

// IsEpochTag reports whether t is an epoch tag.
func IsEpochTag(t Tag) bool {
	switch v := t.(type) {
	case *ConstantTag:
		return v.tag == EpochTagName && v.instance == EpochTagInstance
	case *StandardTag:
		return v.tag == EpochTagName && v.instance == EpochTagInstance
	}
	return false
}

// WithEpoch returns tags with the epoch tag prepended, idempotently.
//
// The epoch must come first: tag order is part of the key, so every
// implementation has to place it in the same position or the keys diverge.
// Passing a list that already starts with an epoch tag returns it unchanged,
// so it is safe to call at more than one layer.
func WithEpoch(tags []Tag, opts ...TagOption) []Tag {
	if len(tags) > 0 && IsEpochTag(tags[0]) {
		return tags
	}
	out := make([]Tag, 0, len(tags)+1)
	out = append(out, EpochTag(opts...))
	return append(out, tags...)
}

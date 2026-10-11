package fragkeys

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/noizu-labs/fragmented-keys-go/pyfloat"
)

// IncrementDelta is added to a tag's version on Increment.
//
// It is 0.1, not 1, to match the PHP, Python, Elixir, and TypeScript ports.
// Binary floating point makes repeated addition drift — 1748000000000.0
// becomes 1748000000000.2002 after two increments, not ...0.2 — and that
// drifted spelling is what goes into the key. Every port drifts identically
// because they all use IEEE-754 doubles, so the keys stay in agreement.
const IncrementDelta = 0.1

// ConstantDefaultVersion is the version a ConstantTag takes when none is given.
const ConstantDefaultVersion = 1.0

// Tag is a versioned key component. Bumping its version changes every
// composite key that includes it.
type Tag interface {
	// Name returns "{tag}_{instance}{prefix}", the cache key under which the
	// version is stored and the string embedded in the composite key.
	Name() string

	// Version returns the current version, lazily loading and seeding it for
	// a standard tag.
	Version(ctx context.Context) (float64, error)

	// FullTag returns "{name}@{version}", a human-readable form for logs.
	FullTag(ctx context.Context) (string, error)

	// Increment bumps the version by IncrementDelta and persists it.
	Increment(ctx context.Context) error

	// Reset replaces the version with a fresh timestamp seed and persists it.
	Reset(ctx context.Context) error

	// SetVersion sets an explicit version, persisting it when persist is true.
	SetVersion(ctx context.Context, version float64, persist bool) error

	// Handler returns the tag's cache handler, falling back to the global default.
	Handler() (CacheHandler, error)

	// DelegateCacheQuery reports whether this tag takes part in a bulk fetch
	// for the given handler group. Constant tags always return false.
	DelegateCacheQuery(group string) bool
}

// tagConfig carries construction options.
type tagConfig struct {
	handler    CacheHandler
	prefix     *string
	version    *float64
	ttl        time.Duration
	hasVersion bool
}

// TagOption customises a tag at construction.
type TagOption func(*tagConfig)

// WithHandler gives the tag its own cache handler instead of the global default.
func WithHandler(h CacheHandler) TagOption {
	return func(c *tagConfig) { c.handler = h }
}

// WithPrefix overrides the global prefix for this tag.
func WithPrefix(p string) TagOption {
	return func(c *tagConfig) { c.prefix = &p }
}

// WithVersion pins an initial version, skipping the lazy load for a standard
// tag and setting the fixed value for a constant one.
func WithVersion(v float64) TagOption {
	return func(c *tagConfig) { c.version = &v; c.hasVersion = true }
}

// WithTTL sets the expiry applied when the tag writes its version. Zero means
// the backend default, which is what every other port uses.
func WithTTL(d time.Duration) TagOption {
	return func(c *tagConfig) { c.ttl = d }
}

// baseTag holds the state shared by standard and constant tags.
type baseTag struct {
	tag      string
	instance string
	prefix   string
	handler  CacheHandler
	ttl      time.Duration

	mu      sync.Mutex
	version *float64 // nil until loaded
}

// initBase fills b in place. It takes a pointer rather than returning a value
// because baseTag carries a mutex, which must not be copied.
func initBase(b *baseTag, tag, instance string, opts []TagOption) {
	cfg := &tagConfig{}
	for _, o := range opts {
		o(cfg)
	}
	// The prefix is resolved once, at construction, exactly as the Python port
	// does. A later SetGlobalPrefix does not retroactively rename live tags.
	prefix := Config.GlobalPrefix()
	if cfg.prefix != nil {
		prefix = *cfg.prefix
	}
	b.tag = tag
	b.instance = instance
	b.prefix = prefix
	b.handler = cfg.handler
	b.ttl = cfg.ttl
	b.version = cfg.version
}

// Name implements Tag.
func (b *baseTag) Name() string {
	return b.tag + "_" + b.instance + b.prefix
}

// Handler implements Tag.
func (b *baseTag) Handler() (CacheHandler, error) {
	b.mu.Lock()
	h := b.handler
	b.mu.Unlock()
	if h != nil {
		return h, nil
	}
	return Config.DefaultCacheHandler()
}

// SetHandler overrides the tag's cache handler after construction.
func (b *baseTag) SetHandler(h CacheHandler) {
	b.mu.Lock()
	b.handler = h
	b.mu.Unlock()
}

// StandardTag is a mutable tag whose version lives in a cache backend.
type StandardTag struct {
	baseTag
}

// NewStandardTag creates a mutable tag. Its version is loaded from the cache
// on first use, and seeded with a millisecond timestamp if absent.
func NewStandardTag(tag, instance string, opts ...TagOption) *StandardTag {
	t := &StandardTag{}
	initBase(&t.baseTag, tag, instance, opts)
	return t
}

// Version implements Tag. The loaded value is cached on the tag, so repeated
// calls do not re-query; use a fresh tag (or Key resolution, which bulk-fetches
// afresh each time) to observe an external increment.
func (t *StandardTag) Version(ctx context.Context) (float64, error) {
	t.mu.Lock()
	if t.version != nil {
		v := *t.version
		t.mu.Unlock()
		return v, nil
	}
	t.mu.Unlock()

	v, err := t.load(ctx)
	if err != nil {
		return 0, err
	}

	t.mu.Lock()
	t.version = &v
	t.mu.Unlock()
	return v, nil
}

// load fetches the stored version, seeding one when the key is absent.
func (t *StandardTag) load(ctx context.Context) (float64, error) {
	h, err := t.Handler()
	if err != nil {
		return 0, err
	}
	raw, found, err := h.Get(ctx, t.Name())
	if err != nil {
		return 0, fmt.Errorf("fragkeys: load version for %s: %w", t.Name(), err)
	}
	if found {
		v, err := pyfloat.Parse(raw)
		if err != nil {
			return 0, fmt.Errorf("fragkeys: version for %s is not a number (%q): %w", t.Name(), raw, err)
		}
		return v, nil
	}

	// Absent: seed with the current time in milliseconds, matching Python's
	// time.time() * 1000. Two processes seeding concurrently will write
	// different values and one will win; both then agree on the winner. Every
	// port has this same benign race.
	seed := nowMillis()
	if err := h.Set(ctx, t.Name(), pyfloat.Format(seed), t.ttl); err != nil {
		return 0, fmt.Errorf("fragkeys: seed version for %s: %w", t.Name(), err)
	}
	return seed, nil
}

// FullTag implements Tag.
func (t *StandardTag) FullTag(ctx context.Context) (string, error) {
	return fullTag(ctx, t)
}

// Increment implements Tag.
func (t *StandardTag) Increment(ctx context.Context) error {
	cur, err := t.Version(ctx)
	if err != nil {
		return err
	}
	return t.SetVersion(ctx, cur+IncrementDelta, true)
}

// Reset implements Tag.
func (t *StandardTag) Reset(ctx context.Context) error {
	return t.SetVersion(ctx, nowMillis(), true)
}

// SetVersion implements Tag.
func (t *StandardTag) SetVersion(ctx context.Context, version float64, persist bool) error {
	t.mu.Lock()
	t.version = &version
	t.mu.Unlock()

	if !persist {
		return nil
	}
	h, err := t.Handler()
	if err != nil {
		return err
	}
	if err := h.Set(ctx, t.Name(), pyfloat.Format(version), t.ttl); err != nil {
		return fmt.Errorf("fragkeys: persist version for %s: %w", t.Name(), err)
	}
	return nil
}

// DelegateCacheQuery implements Tag.
func (t *StandardTag) DelegateCacheQuery(group string) bool {
	h, err := t.Handler()
	if err != nil {
		return false
	}
	return h.GroupName() == group
}

// ConstantTag is an immutable tag. Its version is fixed at construction and it
// never touches the cache backend, which is what makes it suitable for schema
// or API version dimensions.
type ConstantTag struct {
	baseTag
}

// NewConstantTag creates an immutable tag. Without WithVersion the version is
// ConstantDefaultVersion.
func NewConstantTag(tag, instance string, opts ...TagOption) *ConstantTag {
	t := &ConstantTag{}
	initBase(&t.baseTag, tag, instance, opts)
	if t.version == nil {
		v := ConstantDefaultVersion
		t.version = &v
	}
	return t
}

// Version implements Tag. It never reaches the cache.
func (t *ConstantTag) Version(context.Context) (float64, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return *t.version, nil
}

// FullTag implements Tag.
func (t *ConstantTag) FullTag(ctx context.Context) (string, error) {
	return fullTag(ctx, t)
}

// Increment implements Tag as a no-op.
func (t *ConstantTag) Increment(context.Context) error { return nil }

// Reset implements Tag as a no-op.
func (t *ConstantTag) Reset(context.Context) error { return nil }

// SetVersion implements Tag as a no-op. A constant tag that could be reassigned
// would silently break the keys that depend on its stability.
func (t *ConstantTag) SetVersion(context.Context, float64, bool) error { return nil }

// DelegateCacheQuery implements Tag. Always false: a constant tag is excluded
// from bulk version fetches because it has nothing stored to fetch.
func (t *ConstantTag) DelegateCacheQuery(string) bool { return false }

func fullTag(ctx context.Context, t Tag) (string, error) {
	v, err := t.Version(ctx)
	if err != nil {
		return "", err
	}
	return t.Name() + "@" + pyfloat.Format(v), nil
}

// nowMillis returns the current Unix time in milliseconds as a float64, the
// same quantity Python's time.time() * 1000 produces.
func nowMillis() float64 {
	return float64(time.Now().UnixNano()) / 1e6
}

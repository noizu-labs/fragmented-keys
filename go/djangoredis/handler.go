package djangoredis

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/noizu-labs/fragmented-keys-go/fragkeys"
	"github.com/noizu-labs/fragmented-keys-go/pyfloat"
	"github.com/redis/go-redis/v9"
)

// Handler is a fragkeys.CacheHandler backed by the same Redis a Django
// application uses, speaking django-redis's key naming and value encoding.
//
// It satisfies fragkeys.CacheHandler for tag-version storage, and adds
// GetValue/SetValue and the *WithTags helpers for cache payloads — the
// equivalent of FragmentedKeysCacheBackend's get_with_tags/set_with_tags.
type Handler struct {
	rdb redis.UniversalClient

	// keyPrefix is Django's KEY_PREFIX. Empty is the Django default, and is
	// what a settings file that only specifies "PREFIX" ends up with.
	keyPrefix string
	// cacheVersion is Django's cache VERSION, 1 by default.
	cacheVersion int

	groupName  string
	defaultTTL time.Duration
	codec      codec

	// keyFunc, when set, replaces the default naming entirely.
	keyFunc func(string) string

	// epoch, when set, is prepended to every tagged key as a constant
	// dimension. nil means inherit fragkeys.Config.Epoch().
	epoch *float64
}

// Option customises a Handler.
type Option func(*Handler)

// WithKeyPrefix sets Django's KEY_PREFIX. Only set this if the Django settings
// actually specify KEY_PREFIX; note that django-redis ignores a "PREFIX" entry
// for this purpose.
func WithKeyPrefix(p string) Option {
	return func(h *Handler) { h.keyPrefix = p }
}

// WithCacheVersion sets Django's cache VERSION component. Default 1.
func WithCacheVersion(v int) Option {
	return func(h *Handler) { h.cacheVersion = v }
}

// WithGroupName overrides the bulk-fetch group identifier.
//
// The group name is process-local — it decides which tags share a GetMulti
// round trip and is never stored — so it does not have to match Python's
// "DjangoCacheHandler:RedisCache:default". Set it only when one key mixes
// several handlers and they must not be merged.
func WithGroupName(n string) Option {
	return func(h *Handler) { h.groupName = n }
}

// WithDefaultTTL sets the expiry used when a caller passes ttl == 0. Zero
// means no expiry, matching a Django TIMEOUT of None.
func WithDefaultTTL(d time.Duration) Option {
	return func(h *Handler) { h.defaultTTL = d }
}

// WithCompression controls whether this process compresses the values it
// writes. Enable it when Django's COMPRESSOR is ZlibCompressor — as
// gnp-backend's is — so entry sizes stay consistent across writers. Reads
// handle both forms either way.
func WithCompression(on bool) Option {
	return func(h *Handler) { h.codec.compress = on }
}

// WithEpoch pins the global cache epoch this handler applies to tagged keys,
// overriding fragkeys.Config.Epoch().
//
// The value must match the Django side's EPOCH setting (or its
// configure_fragmented_keys(epoch=...) argument) or the two runtimes will
// compute different keys for the same data. Raising it on both sides at once
// strands every existing entry — which is the point when a key-derivation or
// schema change means the previous generation should not be trusted.
func WithEpoch(epoch float64) Option {
	return func(h *Handler) { h.epoch = &epoch }
}

// New builds a Handler over an existing Redis client.
//
// Pass whichever redis.UniversalClient matches the Django deployment: a plain
// Client, or a FailoverClient when Django is configured with
// django_redis.client.SentinelClient. The client must point at the same
// logical database Django uses — django-redis takes it from the LOCATION URL.
func New(rdb redis.UniversalClient, opts ...Option) *Handler {
	h := &Handler{
		rdb:          rdb,
		cacheVersion: 1,
		groupName:    "djangoredis",
		codec:        codec{compress: true},
	}
	for _, o := range opts {
		o(h)
	}
	return h
}

// WithKeyFunc replaces physical-key derivation wholesale. Use it when the
// Django side names keys in a way the KEY_PREFIX/version pair cannot express —
// see DoublePrefixedKeyFunc.
func WithKeyFunc(f func(logicalKey string) string) Option {
	return func(h *Handler) { h.keyFunc = f }
}

// DoublePrefixedKeyFunc reproduces the naming produced when a tag or value
// passes through fragmented_keys_django's FragmentedKeysCacheBackend twice.
//
// That backend inherits Django's BaseCache.make_key, which prefixes a key
// once, and then hands the already-prefixed key to the wrapped django-redis
// backend, which prefixes it again — so a tag named "Page_1GNP" is stored at
// ":1::1:Page_1GNP" rather than ":1:Page_1GNP".
//
// Which naming a given entry uses depends on how the tag was constructed on
// the Django side:
//
//   - Tags built by the backend's own set_with_tags/get_with_tags/
//     invalidate_tags carry the backend's inner handler and are singly
//     prefixed. Use the default naming for those.
//
//   - Tags built anywhere else — model_tag(), ModelTagManager, the
//     @tagged_cache decorator, invalidate_scope(), or a bare StandardTag() —
//     fall back to the global handler configured by
//     configure_fragmented_keys(), which wraps the backend and so is doubly
//     prefixed. Use this key function for those.
//
// Both namespaces can be live in one Django process at once, holding different
// versions for the same logical tag. Confirm which one the Django code you
// need to interoperate with actually writes before choosing.
func DoublePrefixedKeyFunc(keyPrefix string, cacheVersion int) func(string) string {
	inner := keyPrefix + ":" + strconv.Itoa(cacheVersion) + ":"
	return func(key string) string {
		return inner + inner + key
	}
}

// RedisKey returns the physical Redis key for a logical cache key, applying
// django-redis's "{KEY_PREFIX}:{version}:{key}" naming.
//
// Exported because it is the first thing to check when Go and Django disagree
// about whether an entry exists.
func (h *Handler) RedisKey(key string) string {
	if h.keyFunc != nil {
		return h.keyFunc(key)
	}
	return h.keyPrefix + ":" + strconv.Itoa(h.cacheVersion) + ":" + key
}

// GroupName implements fragkeys.CacheHandler.
func (h *Handler) GroupName() string { return h.groupName }

// Get implements fragkeys.CacheHandler, returning a tag version as the string
// the other ports would have produced.
func (h *Handler) Get(ctx context.Context, key string) (string, bool, error) {
	raw, err := h.rdb.Get(ctx, h.RedisKey(key)).Bytes()
	if errors.Is(err, redis.Nil) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("djangoredis: get %s: %w", key, err)
	}
	v, err := h.codec.decode(raw)
	if err != nil {
		return "", false, err
	}
	s, err := versionToString(v)
	if err != nil {
		return "", false, fmt.Errorf("djangoredis: key %s: %w", key, err)
	}
	return s, true, nil
}

// Set implements fragkeys.CacheHandler. Versions are written as Python str
// values — that is, pickled strings — because that is what
// fragmented_keys_django's DjangoCacheHandler.set() stores.
func (h *Handler) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return h.SetValue(ctx, key, value, ttl)
}

// GetMulti implements fragkeys.CacheHandler using a single MGET.
func (h *Handler) GetMulti(ctx context.Context, keys []string) (map[string]string, error) {
	if len(keys) == 0 {
		return map[string]string{}, nil
	}
	physical := make([]string, len(keys))
	for i, k := range keys {
		physical[i] = h.RedisKey(k)
	}
	vals, err := h.rdb.MGet(ctx, physical...).Result()
	if err != nil {
		return nil, fmt.Errorf("djangoredis: mget: %w", err)
	}

	out := make(map[string]string, len(keys))
	for i, raw := range vals {
		if raw == nil || i >= len(keys) {
			continue
		}
		b, ok := toBytes(raw)
		if !ok {
			continue
		}
		v, err := h.codec.decode(b)
		if err != nil {
			// A value Go cannot decode is treated as absent, so the tag
			// re-seeds rather than failing the whole key resolution.
			continue
		}
		s, err := versionToString(v)
		if err != nil {
			continue
		}
		out[keys[i]] = s
	}
	return out, nil
}

// GetValue reads an arbitrary cache payload written by either runtime.
//
// A value Django stored that has no Go representation — a model instance, a
// datetime, a Decimal — comes back as an error wrapping
// pypickle.ErrUnsupportedOpcode. Treat that as a miss and recompute.
func (h *Handler) GetValue(ctx context.Context, key string) (any, bool, error) {
	raw, err := h.rdb.Get(ctx, h.RedisKey(key)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("djangoredis: get %s: %w", key, err)
	}
	v, err := h.codec.decode(raw)
	if err != nil {
		return nil, false, err
	}
	return v, true, nil
}

// SetValue writes an arbitrary cache payload in Django's format. A zero ttl
// uses the handler's default.
func (h *Handler) SetValue(ctx context.Context, key string, value any, ttl time.Duration) error {
	enc, err := h.codec.encode(value)
	if err != nil {
		return err
	}
	if ttl == 0 {
		ttl = h.defaultTTL
	}
	if err := h.rdb.Set(ctx, h.RedisKey(key), enc, ttl).Err(); err != nil {
		return fmt.Errorf("djangoredis: set %s: %w", key, err)
	}
	return nil
}

// Delete removes a key. Reports whether it existed.
func (h *Handler) Delete(ctx context.Context, key string) (bool, error) {
	n, err := h.rdb.Del(ctx, h.RedisKey(key)).Result()
	if err != nil {
		return false, fmt.Errorf("djangoredis: del %s: %w", key, err)
	}
	return n > 0, nil
}

// ---- tagged payload API ---------------------------------------------------

// GetWithTags resolves a fragmented key and reads the value behind it. This is
// the Go counterpart of FragmentedKeysCacheBackend.get_with_tags.
func (h *Handler) GetWithTags(ctx context.Context, name string, tags []fragkeys.Tag) (any, bool, error) {
	k, err := h.resolve(ctx, name, tags)
	if err != nil {
		return nil, false, err
	}
	return h.GetValue(ctx, k)
}

// SetWithTags resolves a fragmented key and writes the value behind it, the
// counterpart of FragmentedKeysCacheBackend.set_with_tags.
func (h *Handler) SetWithTags(ctx context.Context, name string, value any, tags []fragkeys.Tag, ttl time.Duration) error {
	k, err := h.resolve(ctx, name, tags)
	if err != nil {
		return err
	}
	return h.SetValue(ctx, k, value, ttl)
}

// InvalidateTags increments each tag, making every key that includes any of
// them resolve differently from the next read onward. Nothing is deleted; the
// orphaned entries age out under their own TTL.
//
// On a partial failure the tags already incremented stay incremented. That is
// safe: over-invalidation costs a recompute, whereas rolling back would risk
// serving stale data.
func (h *Handler) InvalidateTags(ctx context.Context, tags []fragkeys.Tag) error {
	var errs []error
	for _, t := range tags {
		if err := t.Increment(ctx); err != nil {
			errs = append(errs, fmt.Errorf("invalidate %s: %w", t.Name(), err))
		}
	}
	return errors.Join(errs...)
}

// resolve builds the composite key for a tagged operation, prepending the
// epoch dimension so Go and Django agree on the key for the same data.
func (h *Handler) resolve(ctx context.Context, name string, tags []fragkeys.Tag) (string, error) {
	var withEpoch []fragkeys.Tag
	if h.epoch != nil {
		withEpoch = prependEpochAt(*h.epoch, tags)
	} else {
		withEpoch = fragkeys.WithEpoch(tags)
	}
	return fragkeys.NewKey(name, withEpoch...).String(ctx)
}

// prependEpochAt is fragkeys.WithEpoch pinned to an explicit epoch.
func prependEpochAt(epoch float64, tags []fragkeys.Tag) []fragkeys.Tag {
	if len(tags) > 0 && fragkeys.IsEpochTag(tags[0]) {
		return tags
	}
	out := make([]fragkeys.Tag, 0, len(tags)+1)
	out = append(out, fragkeys.EpochTagAt(epoch))
	return append(out, tags...)
}

// ---- helpers --------------------------------------------------------------

// versionToString renders a decoded version the way Python's
// DjangoCacheHandler.get() would: strings pass through, numbers go through
// str(). It is the bridge between "whatever was pickled" and the exact decimal
// spelling the key formula needs.
func versionToString(v any) (string, error) {
	switch t := v.(type) {
	case string:
		return t, nil
	case []byte:
		return string(t), nil
	case float64:
		return pyfloat.Format(t), nil
	case int64:
		return strconv.FormatInt(t, 10), nil
	case bool:
		// str(True) in Python; nonsensical as a version, but preserve the
		// behaviour rather than inventing one.
		if t {
			return "True", nil
		}
		return "False", nil
	default:
		return "", fmt.Errorf("version has unusable type %T", v)
	}
}

func toBytes(v any) ([]byte, bool) {
	switch t := v.(type) {
	case string:
		return []byte(t), true
	case []byte:
		return t, true
	}
	return nil, false
}

// Compile-time proof that Handler satisfies the interface fragkeys expects.
var _ fragkeys.CacheHandler = (*Handler)(nil)

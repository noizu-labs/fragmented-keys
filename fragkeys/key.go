package fragkeys

import (
	"context"
	"crypto/md5" //nolint:gosec // Not a security primitive; MD5 is the cross-port key format.
	"encoding/hex"
	"strings"

	"github.com/noizu-labs/fragmented-keys-go/pyfloat"
)

// Key is a composite cache key: a base name plus an ordered list of tags.
//
// The resolved key is
//
//	md5("{name}_{groupID}" + ":t{tagName}:v{version}" for each tag, in order)
//
// Tag order is significant and is the order tags were added — the same tags
// added in a different order produce a different key. That matches the PHP,
// Python, Elixir, and TypeScript ports (the Java and Swift ports sort instead,
// and so are not key-compatible with this one).
type Key struct {
	name    string
	groupID string
	tags    []Tag
}

// NewKey builds a key from a base name and an ordered tag list.
func NewKey(name string, tags ...Tag) *Key {
	return &Key{name: name, tags: append([]Tag(nil), tags...)}
}

// WithGroupID returns a copy namespaced by groupID, which is appended to the
// base name as "{name}_{groupID}".
func (k *Key) WithGroupID(groupID string) *Key {
	out := *k
	out.groupID = groupID
	out.tags = append([]Tag(nil), k.tags...)
	return &out
}

// AddTag appends a tag. Because order matters, appending mid-life changes the
// resolved key.
func (k *Key) AddTag(t Tag) { k.tags = append(k.tags, t) }

// Tags returns the tag list.
func (k *Key) Tags() []Tag { return k.tags }

// Raw resolves the unhashed key string. Useful for debugging a key mismatch
// against another runtime: compare Raw output before comparing hashes.
func (k *Key) Raw(ctx context.Context) (string, error) {
	versions, err := k.gatherVersions(ctx)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString(k.name)
	b.WriteByte('_')
	b.WriteString(k.groupID)

	for _, t := range k.tags {
		name := t.Name()
		v, ok := versions[name]
		if !ok {
			// Not returned by the bulk fetch — either a constant tag, or a
			// standard tag whose key is absent. Fall back to the tag itself,
			// which seeds and stores a version if needed. This mirrors the
			// Python port's versions.get(name, tag.get_tag_version()).
			v, err = t.Version(ctx)
			if err != nil {
				return "", err
			}
		}
		b.WriteString(":t")
		b.WriteString(name)
		b.WriteString(":v")
		b.WriteString(pyfloat.Format(v))
	}
	return b.String(), nil
}

// String resolves the key and returns its MD5 hex digest, which is what the
// cache is keyed by.
func (k *Key) String(ctx context.Context) (string, error) {
	raw, err := k.Raw(ctx)
	if err != nil {
		return "", err
	}
	return HashRaw(raw), nil
}

// HashRaw applies the family's hash to an already-built raw key string:
// lowercase MD5 hex, 32 characters.
//
// MD5 is a key-derivation choice inherited from the original PHP
// implementation, not a security control. Changing it would desynchronise
// every port.
func HashRaw(raw string) string {
	sum := md5.Sum([]byte(raw)) //nolint:gosec // See above.
	return hex.EncodeToString(sum[:])
}

// gatherVersions bulk-fetches tag versions, one round trip per handler group.
//
// The tags are not mutated, so each resolution reflects current cache state and
// an increment made elsewhere is picked up by the next call on the same Key.
func (k *Key) gatherVersions(ctx context.Context) (map[string]float64, error) {
	// Partition by handler group, preserving first-seen order so the round
	// trips are deterministic.
	type group struct {
		handler CacheHandler
		names   []string
	}
	var order []string
	groups := map[string]*group{}

	for _, t := range k.tags {
		h, err := t.Handler()
		if err != nil {
			// A constant tag with no handler is legitimate: it never queries.
			if t.DelegateCacheQuery("") {
				return nil, err
			}
			continue
		}
		name := h.GroupName()
		if !t.DelegateCacheQuery(name) {
			continue
		}
		g, ok := groups[name]
		if !ok {
			g = &group{handler: h}
			groups[name] = g
			order = append(order, name)
		}
		g.names = append(g.names, t.Name())
	}

	versions := make(map[string]float64, len(k.tags))
	for _, name := range order {
		g := groups[name]
		found, err := g.handler.GetMulti(ctx, g.names)
		if err != nil {
			return nil, err
		}
		for tagName, raw := range found {
			v, err := pyfloat.Parse(raw)
			if err != nil {
				// A malformed stored version is not a reason to fail the
				// request: skip it and let the tag re-seed below.
				continue
			}
			versions[tagName] = v
		}
	}
	return versions, nil
}

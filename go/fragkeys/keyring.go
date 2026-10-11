package fragkeys

import (
	"fmt"
	"time"
)

// TagType selects which tag implementation a KeyRing instantiates.
type TagType string

const (
	// TagStandard is the mutable, cache-backed tag. It is the default.
	TagStandard TagType = "standard"
	// TagConstant is the immutable tag.
	TagConstant TagType = "constant"
)

// TagOptions configures a tag created through a KeyRing.
//
// Pointer and empty-value fields mean "not set", so options can be merged from
// general to specific without a zero value clobbering a real one.
type TagOptions struct {
	// Type selects standard or constant. Empty means standard.
	Type TagType
	// Version pins the version. For a constant tag this is its fixed value.
	Version *float64
	// HandlerName selects a handler from the ring's registry.
	HandlerName string
	// Handler sets a handler directly, taking precedence over HandlerName.
	Handler CacheHandler
	// Prefix overrides the tag-name suffix.
	Prefix *string
	// TTL sets the expiry used when the tag writes its version.
	TTL time.Duration
}

// merge layers other on top of o, with other winning on every field it sets.
func (o TagOptions) merge(other TagOptions) TagOptions {
	out := o
	if other.Type != "" {
		out.Type = other.Type
	}
	if other.Version != nil {
		out.Version = other.Version
	}
	if other.HandlerName != "" {
		out.HandlerName = other.HandlerName
	}
	if other.Handler != nil {
		out.Handler = other.Handler
	}
	if other.Prefix != nil {
		out.Prefix = other.Prefix
	}
	if other.TTL != 0 {
		out.TTL = other.TTL
	}
	return out
}

// Param is one tag slot in a key template: the tag's logical name plus any
// overrides that apply only in this key.
type Param struct {
	Tag     string
	Options TagOptions
}

// P is shorthand for a Param with no overrides.
func P(tag string) Param { return Param{Tag: tag} }

type keyDefinition struct {
	params  []Param
	globals TagOptions
}

// KeyRing is a factory for reusable key templates.
//
// Define a key once with the tag slots it needs, then instantiate it with
// concrete tag instances. Options resolve from general to specific:
//
//	globalOptions -> globalTagOptions[tag] -> per-key globals -> per-param
//
// which is the same merge chain the PHP, Python, Elixir, and TypeScript ports
// use.
//
// The dynamic accessor those ports offer (ring.get_users_key_obj(...) via
// __call__, __getattr__, or a Proxy) has no Go equivalent and is deliberately
// omitted; the specification marks it optional. Use KeyObj with the key name.
type KeyRing struct {
	handlers         map[string]CacheHandler
	defaultHandler   string
	defaultPrefix    *string
	globalOptions    TagOptions
	globalTagOptions map[string]TagOptions
	definitions      map[string]keyDefinition
}

// KeyRingOption customises a ring at construction.
type KeyRingOption func(*KeyRing)

// WithHandlers registers the named handler pool a ring's tags can draw from.
func WithHandlers(h map[string]CacheHandler) KeyRingOption {
	return func(r *KeyRing) {
		for k, v := range h {
			r.handlers[k] = v
		}
	}
}

// WithDefaultHandler names the handler used when a tag specifies none.
func WithDefaultHandler(name string) KeyRingOption {
	return func(r *KeyRing) { r.defaultHandler = name }
}

// WithDefaultPrefix sets the tag-name suffix for every tag from this ring,
// overriding the global configuration.
func WithDefaultPrefix(p string) KeyRingOption {
	return func(r *KeyRing) { r.defaultPrefix = &p }
}

// WithGlobalOptions sets the base of the option merge chain.
func WithGlobalOptions(o TagOptions) KeyRingOption {
	return func(r *KeyRing) { r.globalOptions = o }
}

// WithGlobalTagOptions sets per-tag-name defaults, the second link in the
// merge chain. Typical use is marking a schema-version tag constant once:
//
//	WithGlobalTagOptions(map[string]TagOptions{
//	    "schema": {Type: TagConstant, Version: Float(0.1)},
//	})
func WithGlobalTagOptions(m map[string]TagOptions) KeyRingOption {
	return func(r *KeyRing) {
		for k, v := range m {
			r.globalTagOptions[k] = v
		}
	}
}

// NewKeyRing creates an empty ring.
func NewKeyRing(opts ...KeyRingOption) *KeyRing {
	r := &KeyRing{
		handlers:         map[string]CacheHandler{},
		globalTagOptions: map[string]TagOptions{},
		definitions:      map[string]keyDefinition{},
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

// DefineKey registers a key template: an ordered list of tag slots, plus
// options applying to all of them within this key.
func (r *KeyRing) DefineKey(name string, params []Param, globals TagOptions) {
	r.definitions[name] = keyDefinition{
		params:  append([]Param(nil), params...),
		globals: globals,
	}
}

// KeyObj instantiates a defined key. One instance value is required per
// declared param, in declaration order.
func (r *KeyRing) KeyObj(name string, values ...string) (*Key, error) {
	return r.KeyObjWithGroup(name, "", values...)
}

// KeyObjWithGroup is KeyObj with a group ID namespacing the base key name.
func (r *KeyRing) KeyObjWithGroup(name, groupID string, values ...string) (*Key, error) {
	def, ok := r.definitions[name]
	if !ok {
		return nil, fmt.Errorf("fragkeys: no key definition named %q", name)
	}
	if len(values) != len(def.params) {
		return nil, fmt.Errorf(
			"fragkeys: key %q takes %d tag instance(s), got %d",
			name, len(def.params), len(values),
		)
	}

	tags := make([]Tag, 0, len(def.params))
	for i, p := range def.params {
		opts := r.globalOptions.
			merge(r.globalTagOptions[p.Tag]).
			merge(def.globals).
			merge(p.Options)
		t, err := r.Tag(p.Tag, values[i], opts)
		if err != nil {
			return nil, fmt.Errorf("fragkeys: key %q tag %q: %w", name, p.Tag, err)
		}
		tags = append(tags, t)
	}

	k := NewKey(name, tags...)
	if groupID != "" {
		k = k.WithGroupID(groupID)
	}
	return k, nil
}

// Tag builds a single tag with this ring's defaults applied. Options passed
// here are the most specific and win over the ring's own.
func (r *KeyRing) Tag(tagName, instance string, opts TagOptions) (Tag, error) {
	merged := r.globalOptions.
		merge(r.globalTagOptions[tagName]).
		merge(opts)

	handler, err := r.resolveHandler(merged)
	if err != nil {
		return nil, err
	}

	tagOpts := make([]TagOption, 0, 4)
	if handler != nil {
		tagOpts = append(tagOpts, WithHandler(handler))
	}
	switch {
	case merged.Prefix != nil:
		tagOpts = append(tagOpts, WithPrefix(*merged.Prefix))
	case r.defaultPrefix != nil:
		tagOpts = append(tagOpts, WithPrefix(*r.defaultPrefix))
	}
	if merged.Version != nil {
		tagOpts = append(tagOpts, WithVersion(*merged.Version))
	}
	if merged.TTL != 0 {
		tagOpts = append(tagOpts, WithTTL(merged.TTL))
	}

	if merged.Type == TagConstant {
		return NewConstantTag(tagName, instance, tagOpts...), nil
	}
	return NewStandardTag(tagName, instance, tagOpts...), nil
}

// resolveHandler picks the handler for a tag: an explicit instance first, then
// a named one from the registry, then the ring's default. A nil result means
// the tag falls back to the global Configuration default.
func (r *KeyRing) resolveHandler(o TagOptions) (CacheHandler, error) {
	if o.Handler != nil {
		return o.Handler, nil
	}
	name := o.HandlerName
	if name == "" {
		name = r.defaultHandler
	}
	if name == "" {
		return nil, nil
	}
	h, ok := r.handlers[name]
	if !ok {
		return nil, fmt.Errorf("no handler named %q registered on this key ring", name)
	}
	return h, nil
}

// Float is a helper for the pointer fields on TagOptions.
func Float(v float64) *float64 { return &v }

// String is a helper for the pointer fields on TagOptions.
func String(v string) *string { return &v }

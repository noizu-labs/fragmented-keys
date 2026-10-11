package fragkeys

import (
	"context"
	"strings"
	"testing"
	"time"
)

func testEnv(t *testing.T) (*MemoryHandler, context.Context) {
	t.Helper()
	h := NewMemoryHandler("MemoryHandler")
	Config.SetDefaultCacheHandler(h)
	Config.SetGlobalPrefix("GNP")
	t.Cleanup(Config.Reset)
	return h, context.Background()
}

func TestTagNameFormat(t *testing.T) {
	testEnv(t)
	cases := []struct {
		tag, instance, prefix, want string
	}{
		{"User", "42", "GNP", "User_42GNP"},
		{"Global", "", "GNP", "Global_GNP"},
		{"City", "chicago", "", "City_chicago"},
	}
	for _, c := range cases {
		got := NewStandardTag(c.tag, c.instance, WithPrefix(c.prefix)).Name()
		if got != c.want {
			t.Errorf("Name() = %q, want %q", got, c.want)
		}
	}
}

func TestTagSeedsOnFirstAccessAndPersists(t *testing.T) {
	h, ctx := testEnv(t)

	tag := NewStandardTag("User", "42")
	v, err := tag.Version(ctx)
	if err != nil {
		t.Fatalf("Version: %v", err)
	}
	if v <= 0 {
		t.Fatalf("seeded version = %v, want a positive millisecond timestamp", v)
	}

	// The seed must be written through, so a second process sees the same one.
	stored, found, err := h.Get(ctx, "User_42GNP")
	if err != nil || !found {
		t.Fatalf("seed was not persisted (found=%v, err=%v)", found, err)
	}
	fresh := NewStandardTag("User", "42")
	v2, err := fresh.Version(ctx)
	if err != nil {
		t.Fatalf("Version on fresh tag: %v", err)
	}
	if v2 != v {
		t.Errorf("a second tag read %v, want the persisted %v (stored %q)", v2, v, stored)
	}
}

func TestIncrementChangesKeyAndPersists(t *testing.T) {
	_, ctx := testEnv(t)

	tag := NewStandardTag("User", "42", WithVersion(1.0))
	key := NewKey("profile", tag)

	before, err := key.String(ctx)
	if err != nil {
		t.Fatalf("String: %v", err)
	}
	// Resolving twice without an increment must be stable.
	again, _ := key.String(ctx)
	if before != again {
		t.Errorf("key changed without an increment: %s then %s", before, again)
	}

	if err := tag.Increment(ctx); err != nil {
		t.Fatalf("Increment: %v", err)
	}
	after, _ := key.String(ctx)
	if after == before {
		t.Error("key did not change after Increment")
	}

	// The bump must be visible to a tag built independently — that is what
	// makes invalidation work across processes.
	other := NewKey("profile", NewStandardTag("User", "42"))
	fromScratch, err := other.String(ctx)
	if err != nil {
		t.Fatalf("String on independent tag: %v", err)
	}
	if fromScratch != after {
		t.Errorf("independent tag resolved %s, want the incremented %s", fromScratch, after)
	}
}

func TestIncrementUsesDelta(t *testing.T) {
	_, ctx := testEnv(t)
	tag := NewStandardTag("User", "42", WithVersion(1.0))
	if err := tag.Increment(ctx); err != nil {
		t.Fatalf("Increment: %v", err)
	}
	v, _ := tag.Version(ctx)
	if v != 1.0+IncrementDelta {
		t.Errorf("version = %v, want %v", v, 1.0+IncrementDelta)
	}
}

func TestConstantTagIsImmutableAndNeverQueries(t *testing.T) {
	h, ctx := testEnv(t)

	tag := NewConstantTag("schema", "v1")
	v, _ := tag.Version(ctx)
	if v != ConstantDefaultVersion {
		t.Errorf("default version = %v, want %v", v, ConstantDefaultVersion)
	}

	for _, op := range []struct {
		name string
		fn   func() error
	}{
		{"Increment", func() error { return tag.Increment(ctx) }},
		{"Reset", func() error { return tag.Reset(ctx) }},
		{"SetVersion", func() error { return tag.SetVersion(ctx, 99, true) }},
	} {
		if err := op.fn(); err != nil {
			t.Errorf("%s returned %v, want nil (no-op)", op.name, err)
		}
		if got, _ := tag.Version(ctx); got != ConstantDefaultVersion {
			t.Errorf("%s changed the version to %v", op.name, got)
		}
	}

	if tag.DelegateCacheQuery("MemoryHandler") {
		t.Error("DelegateCacheQuery = true; a constant tag must be excluded from bulk fetch")
	}
	// Nothing should have been written for it.
	if _, found, _ := h.Get(ctx, tag.Name()); found {
		t.Error("constant tag wrote a version to the cache backend")
	}
}

func TestKeyRawFormat(t *testing.T) {
	_, ctx := testEnv(t)
	key := NewKey("Dashboard",
		NewStandardTag("User", "42", WithVersion(1748000000000.0)),
		NewStandardTag("City", "chicago", WithVersion(5.0)),
	)
	want := "Dashboard_:tUser_42GNP:v1748000000000.0:tCity_chicagoGNP:v5.0"
	got, err := key.Raw(ctx)
	if err != nil {
		t.Fatalf("Raw: %v", err)
	}
	if got != want {
		t.Errorf("Raw() = %q, want %q", got, want)
	}
}

func TestGroupIDNamespacesKey(t *testing.T) {
	_, ctx := testEnv(t)
	base := NewKey("list", NewStandardTag("Org", "7", WithVersion(1.0)))
	scoped := NewKey("list", NewStandardTag("Org", "7", WithVersion(1.0))).WithGroupID("us")

	a, _ := base.String(ctx)
	b, _ := scoped.String(ctx)
	if a == b {
		t.Error("group_id did not change the key")
	}
	raw, _ := scoped.Raw(ctx)
	if !strings.HasPrefix(raw, "list_us:") {
		t.Errorf("raw = %q, want it to start with %q", raw, "list_us:")
	}
}

// countingHandler records how many bulk fetches were issued, to prove key
// resolution does one round trip per handler group rather than one per tag.
type countingHandler struct {
	*MemoryHandler
	multiCalls int
	keysSeen   int
}

func (c *countingHandler) GetMulti(ctx context.Context, keys []string) (map[string]string, error) {
	c.multiCalls++
	c.keysSeen += len(keys)
	return c.MemoryHandler.GetMulti(ctx, keys)
}

func TestKeyResolutionBulkFetchesOncePerGroup(t *testing.T) {
	ctx := context.Background()
	h := &countingHandler{MemoryHandler: NewMemoryHandler("counting")}
	Config.SetDefaultCacheHandler(h)
	Config.SetGlobalPrefix("GNP")
	t.Cleanup(Config.Reset)

	key := NewKey("wide",
		NewStandardTag("A", "1", WithHandler(h)),
		NewStandardTag("B", "2", WithHandler(h)),
		NewStandardTag("C", "3", WithHandler(h)),
		NewConstantTag("schema", "v1", WithHandler(h)), // excluded
	)
	if _, err := key.String(ctx); err != nil {
		t.Fatalf("String: %v", err)
	}
	if h.multiCalls != 1 {
		t.Errorf("GetMulti called %d times, want 1 per handler group", h.multiCalls)
	}
	if h.keysSeen != 3 {
		t.Errorf("bulk fetch asked for %d keys, want 3 (the constant tag must be excluded)", h.keysSeen)
	}
}

func TestMixedHandlersUseSeparateRoundTrips(t *testing.T) {
	ctx := context.Background()
	a := &countingHandler{MemoryHandler: NewMemoryHandler("A")}
	b := &countingHandler{MemoryHandler: NewMemoryHandler("B")}
	Config.SetGlobalPrefix("GNP")
	t.Cleanup(Config.Reset)

	key := NewKey("mixed",
		NewStandardTag("X", "1", WithHandler(a)),
		NewStandardTag("Y", "2", WithHandler(b)),
		NewStandardTag("Z", "3", WithHandler(a)),
	)
	if _, err := key.String(ctx); err != nil {
		t.Fatalf("String: %v", err)
	}
	if a.multiCalls != 1 || a.keysSeen != 2 {
		t.Errorf("handler A: %d calls / %d keys, want 1 / 2", a.multiCalls, a.keysSeen)
	}
	if b.multiCalls != 1 || b.keysSeen != 1 {
		t.Errorf("handler B: %d calls / %d keys, want 1 / 1", b.multiCalls, b.keysSeen)
	}
}

func TestTagWithoutHandlerFallsBackToGlobal(t *testing.T) {
	h, ctx := testEnv(t)
	tag := NewStandardTag("User", "1")
	got, err := tag.Handler()
	if err != nil {
		t.Fatalf("Handler: %v", err)
	}
	if got != CacheHandler(h) {
		t.Error("tag did not fall back to the global default handler")
	}

	Config.Reset()
	if _, err := NewStandardTag("User", "2").Version(ctx); err == nil {
		t.Error("expected ErrNoDefaultHandler with no handler configured")
	}
}

func TestKeyRingMatchesManualConstruction(t *testing.T) {
	h, ctx := testEnv(t)

	ring := NewKeyRing(
		WithHandlers(map[string]CacheHandler{"mem": h}),
		WithDefaultHandler("mem"),
		WithDefaultPrefix("GNP"),
		WithGlobalTagOptions(map[string]TagOptions{
			"schema": {Type: TagConstant, Version: Float(1.0)},
		}),
	)
	ring.DefineKey("org_list", []Param{P("schema"), P("Org")}, TagOptions{})

	fromRing, err := ring.KeyObj("org_list", "v1", "7")
	if err != nil {
		t.Fatalf("KeyObj: %v", err)
	}
	manual := NewKey("org_list",
		NewConstantTag("schema", "v1", WithVersion(1.0), WithHandler(h)),
		NewStandardTag("Org", "7", WithHandler(h)),
	)

	a, err := fromRing.String(ctx)
	if err != nil {
		t.Fatalf("ring key: %v", err)
	}
	b, err := manual.String(ctx)
	if err != nil {
		t.Fatalf("manual key: %v", err)
	}
	if a != b {
		ra, _ := fromRing.Raw(ctx)
		rb, _ := manual.Raw(ctx)
		t.Errorf("ring and manual keys differ\n ring: %s\nmanual: %s", ra, rb)
	}
}

func TestKeyRingOptionMergeOrder(t *testing.T) {
	h, _ := testEnv(t)

	ring := NewKeyRing(
		WithHandlers(map[string]CacheHandler{"mem": h}),
		WithDefaultHandler("mem"),
		WithGlobalOptions(TagOptions{Version: Float(1)}),
		WithGlobalTagOptions(map[string]TagOptions{"t": {Version: Float(2)}}),
	)
	// Per-key globals beat global tag options; per-param beats everything.
	ring.DefineKey("k1", []Param{P("t")}, TagOptions{Version: Float(3)})
	ring.DefineKey("k2", []Param{{Tag: "t", Options: TagOptions{Version: Float(4)}}}, TagOptions{Version: Float(3)})

	ctx := context.Background()
	for _, c := range []struct {
		key  string
		want float64
	}{{"k1", 3}, {"k2", 4}} {
		k, err := ring.KeyObj(c.key, "i")
		if err != nil {
			t.Fatalf("KeyObj(%s): %v", c.key, err)
		}
		got, err := k.Tags()[0].Version(ctx)
		if err != nil {
			t.Fatalf("Version: %v", err)
		}
		if got != c.want {
			t.Errorf("%s resolved version %v, want %v", c.key, got, c.want)
		}
	}
}

func TestKeyRingArityIsChecked(t *testing.T) {
	h, _ := testEnv(t)
	ring := NewKeyRing(WithHandlers(map[string]CacheHandler{"mem": h}), WithDefaultHandler("mem"))
	ring.DefineKey("two", []Param{P("a"), P("b")}, TagOptions{})

	if _, err := ring.KeyObj("two", "only-one"); err == nil {
		t.Error("expected an arity error, got nil")
	}
	if _, err := ring.KeyObj("undefined", "x"); err == nil {
		t.Error("expected an error for an undefined key, got nil")
	}
}

func TestMemoryHandlerHonoursTTL(t *testing.T) {
	h := NewMemoryHandler("mem")
	ctx := context.Background()
	if err := h.Set(ctx, "k", "v", time.Millisecond); err != nil {
		t.Fatalf("Set: %v", err)
	}
	time.Sleep(5 * time.Millisecond)
	if _, found, _ := h.Get(ctx, "k"); found {
		t.Error("entry survived its TTL")
	}
}

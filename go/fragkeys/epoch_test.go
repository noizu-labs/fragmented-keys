package fragkeys

import (
	"context"
	"testing"
)

// TestEpochMatchesGoldenVectors is the cross-runtime check for the epoch
// dimension: WithEpoch must produce exactly the key the Python side produces
// when fragmented_keys_django's with_epoch() prepends its constant tag.
//
// If this diverges, Go and Django compute different keys for the same data and
// silently stop sharing cache entries — the failure mode the epoch exists to
// prevent, reintroduced by the epoch itself.
func TestEpochMatchesGoldenVectors(t *testing.T) {
	g := loadGolden(t)

	handler := NewMemoryHandler("MemoryHandler")
	Config.SetDefaultCacheHandler(handler)
	Config.SetGlobalPrefix(g.GlobalPrefix)
	t.Cleanup(Config.Reset)

	ctx := context.Background()
	byName := map[string]goldenCase{}
	for _, c := range g.Cases {
		byName[c.Name] = c
	}

	for _, name := range []string{"epoch_1", "epoch_2"} {
		c, ok := byName[name]
		if !ok {
			t.Fatalf("golden vectors are missing case %q; regenerate testdata/golden_keys.json", name)
		}
		t.Run(name, func(t *testing.T) {
			// The epoch value lives in the first tag descriptor.
			Config.SetEpoch(c.Tags[0].Version)

			// Build only the non-epoch tags, then let WithEpoch prepend —
			// exercising the same call path production uses.
			var tags []Tag
			for _, td := range c.Tags[1:] {
				tags = append(tags, NewStandardTag(td.Tag, td.Instance,
					WithVersion(td.Version), WithHandler(handler)))
			}

			key := NewKey(c.Key, WithEpoch(tags)...)
			raw, err := key.Raw(ctx)
			if err != nil {
				t.Fatalf("Raw: %v", err)
			}
			if raw != c.Raw {
				t.Fatalf("raw differs from %s\n got: %q\nwant: %q", g.Source, raw, c.Raw)
			}
			got, err := key.String(ctx)
			if err != nil {
				t.Fatalf("String: %v", err)
			}
			if got != c.MD5 {
				t.Errorf("md5 = %s, want %s", got, c.MD5)
			}
		})
	}

	// Raising the epoch must strand the previous generation, and an epoch-less
	// key must never collide with an epoch-bearing one.
	e1, e2, base := byName["epoch_1"].MD5, byName["epoch_2"].MD5, byName["no_epoch_baseline"].MD5
	if e1 == e2 {
		t.Error("epoch 1.0 and 2.0 produced the same key; raising the epoch would not bust the cache")
	}
	if e1 == base || e2 == base {
		t.Error("an epoch-bearing key collided with the epoch-less baseline; pre-cutover entries would be reused")
	}
}

func TestEpochTagShape(t *testing.T) {
	Config.SetGlobalPrefix("GNP")
	Config.SetEpoch(2.0)
	t.Cleanup(Config.Reset)

	tag := EpochTag()
	if got := tag.Name(); got != "epoch_globalGNP" {
		t.Errorf("Name() = %q, want %q — this name is part of the cross-port contract", got, "epoch_globalGNP")
	}
	v, err := tag.Version(context.Background())
	if err != nil {
		t.Fatalf("Version: %v", err)
	}
	if v != 2.0 {
		t.Errorf("Version() = %v, want the configured epoch 2.0", v)
	}
	if tag.DelegateCacheQuery("anything") {
		t.Error("the epoch tag must be excluded from bulk fetch; it is constant")
	}
	if !IsEpochTag(tag) {
		t.Error("IsEpochTag did not recognise its own tag")
	}
}

func TestWithEpochPrependsAndIsIdempotent(t *testing.T) {
	Config.SetGlobalPrefix("GNP")
	t.Cleanup(Config.Reset)

	tags := []Tag{NewStandardTag("Page", "careers", WithVersion(1.0))}

	once := WithEpoch(tags)
	if len(once) != 2 {
		t.Fatalf("WithEpoch returned %d tags, want 2", len(once))
	}
	if !IsEpochTag(once[0]) {
		t.Error("the epoch must be the first tag; order is part of the key")
	}

	twice := WithEpoch(once)
	if len(twice) != 2 {
		t.Errorf("WithEpoch stacked a second epoch tag (%d tags); it must be idempotent", len(twice))
	}

	// The caller's slice must not be aliased or mutated.
	if len(tags) != 1 {
		t.Error("WithEpoch mutated the caller's slice")
	}
}

func TestEpochTagAtIgnoresGlobalConfig(t *testing.T) {
	Config.SetEpoch(5.0)
	t.Cleanup(Config.Reset)

	v, err := EpochTagAt(1.0).Version(context.Background())
	if err != nil {
		t.Fatalf("Version: %v", err)
	}
	if v != 1.0 {
		t.Errorf("EpochTagAt(1.0) = %v, want 1.0 — it must ignore the global epoch so a "+
			"previous generation of entries can still be read", v)
	}
}

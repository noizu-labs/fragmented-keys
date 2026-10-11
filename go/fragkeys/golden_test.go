package fragkeys

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// goldenFile is the cross-port contract: raw strings and MD5 digests produced
// by running the reference Python implementation. Regenerate with
// testdata/gen_golden_keys.py — see that script for the command.
type goldenFile struct {
	GlobalPrefix string       `json:"global_prefix"`
	Source       string       `json:"source"`
	Cases        []goldenCase `json:"cases"`
}

type goldenCase struct {
	Name    string      `json:"name"`
	Key     string      `json:"key"`
	GroupID string      `json:"group_id"`
	Tags    []goldenTag `json:"tags"`
	Raw     string      `json:"raw"`
	MD5     string      `json:"md5"`
}

type goldenTag struct {
	Kind     string  `json:"kind"`
	Tag      string  `json:"tag"`
	Instance string  `json:"instance"`
	Version  float64 `json:"version"`
}

func loadGolden(t *testing.T) goldenFile {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "testdata", "golden_keys.json"))
	if err != nil {
		t.Fatalf("read golden vectors: %v", err)
	}
	var g goldenFile
	if err := json.Unmarshal(b, &g); err != nil {
		t.Fatalf("parse golden vectors: %v", err)
	}
	if len(g.Cases) == 0 {
		t.Fatal("golden vector file has no cases")
	}
	return g
}

// TestGoldenKeysMatchPython is the compatibility test this port exists to pass.
// If it fails, a Go service and a Django service pointed at the same Redis are
// computing different keys and silently not sharing cache entries.
func TestGoldenKeysMatchPython(t *testing.T) {
	g := loadGolden(t)

	handler := NewMemoryHandler("MemoryHandler")
	prev := Config.GlobalPrefix()
	Config.SetDefaultCacheHandler(handler)
	Config.SetGlobalPrefix(g.GlobalPrefix)
	t.Cleanup(func() {
		Config.SetGlobalPrefix(prev)
		Config.Reset()
	})

	ctx := context.Background()
	for _, c := range g.Cases {
		t.Run(c.Name, func(t *testing.T) {
			key := NewKey(c.Key)
			if c.GroupID != "" {
				key = key.WithGroupID(c.GroupID)
			}
			for _, td := range c.Tags {
				switch td.Kind {
				case "constant":
					key.AddTag(NewConstantTag(td.Tag, td.Instance,
						WithVersion(td.Version), WithHandler(handler)))
				case "standard":
					key.AddTag(NewStandardTag(td.Tag, td.Instance,
						WithVersion(td.Version), WithHandler(handler)))
				default:
					t.Fatalf("unknown tag kind %q in golden vectors", td.Kind)
				}
			}

			raw, err := key.Raw(ctx)
			if err != nil {
				t.Fatalf("Raw: %v", err)
			}
			if raw != c.Raw {
				// Compare raw before hash: a raw mismatch localises the bug to
				// name construction or version formatting, not the digest.
				t.Fatalf("raw string differs from %s\n got: %q\nwant: %q", g.Source, raw, c.Raw)
			}

			got, err := key.String(ctx)
			if err != nil {
				t.Fatalf("String: %v", err)
			}
			if got != c.MD5 {
				t.Errorf("md5 = %s, want %s (raw matched, so this is a hashing bug)", got, c.MD5)
			}
		})
	}
	t.Logf("verified %d key vectors against %s", len(g.Cases), g.Source)
}

// TestTagOrderChangesKey pins the insertion-order semantics this port shares
// with PHP, Python, Elixir, and TypeScript — and deliberately does not share
// with Java and Swift, which sort tags.
func TestTagOrderChangesKey(t *testing.T) {
	g := loadGolden(t)
	var forward, reversed string
	for _, c := range g.Cases {
		switch c.Name {
		case "two_standard":
			forward = c.MD5
		case "two_standard_reversed":
			reversed = c.MD5
		}
	}
	if forward == "" || reversed == "" {
		t.Skip("golden vectors lack the ordering pair")
	}
	if forward == reversed {
		t.Error("reversing tag order produced the same key; insertion order is not being honoured")
	}
}

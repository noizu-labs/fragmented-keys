package djangoredis

import (
	"encoding/hex"
	"reflect"
	"strconv"
	"testing"
)

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("bad fixture hex: %v", err)
	}
	return b
}

// TestDecodeRealDjangoBytes decodes values captured verbatim out of Redis
// after a Django process wrote them through
// FragmentedKeysCacheBackend -> django_redis.RedisCache with
// COMPRESSOR=ZlibCompressor. These are not synthesised: they are the bytes
// production stores.
func TestDecodeRealDjangoBytes(t *testing.T) {
	c := codec{compress: true}

	cases := []struct {
		name  string
		bytes []byte
		want  any
	}{
		{
			// A tag version: Python str, pickled, then zlib'd because the
			// pickled form exceeds the 15-byte compression floor.
			"tag_version",
			mustHex(t, "789c6b60992aca00013d8286e6166606868696a60646e6267aa696a653f40058860602"),
			"1786011950274.595",
		},
		{
			// A short string. Still compressed: the pickle envelope pushes even
			// "hi" past the floor.
			"short_string",
			mustHex(t, "789c6b6099cac600013d4c199953f40019630341"),
			"hi",
		},
		{
			// An int, which django-redis stores as raw decimal text so that
			// Redis INCR works against it — no pickle, no compression.
			"plain_int",
			mustHex(t, "3432"),
			int64(42),
		},
		{
			"float_with_drift",
			mustHex(t, "789c6b6099cac50001ee4e95f9e7b634309be801002d8804fc"),
			1748000000000.2002,
		},
		{
			// A nested section document, the payload shape this codec exists for.
			"section_document",
			mustHex(t, "789c258c310a02311045351b592bcf6125780d3d8348d81d92b09a844c52082e6c6931e5f7be46fde57bbcbfe8f773f5df8cbd68be550be9079389324374792482ecee264f547cb0d7646c035ba6a1f81818177c43e54788e2235c27da518e904dca31f1cff68eccd8e2f6778a3e40d601e74e549cf0aad5d4c307f99d32c4"),
			map[string]any{
				"slug": "careers",
				"type": "marketing_page",
				"sections": []any{
					map[string]any{
						"id":    "s1",
						"type":  "hero",
						"props": map[string]any{"heading": "Join", "n": int64(3), "ok": true},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := c.decode(tc.bytes)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("decode = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	for _, compress := range []bool{true, false} {
		c := codec{compress: compress}
		values := []any{
			nil, true, false,
			int64(0), int64(42), int64(-7),
			1.5, 1748000000000.901,
			"", "hi", "a longer string that will certainly exceed the compression floor",
			[]any{int64(1), "two", nil},
			map[string]any{"a": int64(1), "b": []any{true, 2.5}},
		}
		for _, v := range values {
			enc, err := c.encode(v)
			if err != nil {
				t.Fatalf("compress=%v encode(%#v): %v", compress, v, err)
			}
			got, err := c.decode(enc)
			if err != nil {
				t.Fatalf("compress=%v decode(%#v): %v", compress, v, err)
			}
			if !reflect.DeepEqual(got, v) {
				t.Errorf("compress=%v round trip = %#v, want %#v", compress, got, v)
			}
		}
	}
}

// TestIntsBypassPickling pins the behaviour that makes Django's INCR work.
func TestIntsBypassPickling(t *testing.T) {
	c := codec{compress: true}
	enc, err := c.encode(42)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if string(enc) != "42" {
		t.Errorf("encode(42) = %q, want the raw decimal %q", enc, "42")
	}

	// Booleans must NOT take that path, even though Python's isinstance(True,
	// int) is true — django-redis pickles them.
	enc, err = c.encode(true)
	if err != nil {
		t.Fatalf("encode(true): %v", err)
	}
	if _, err := strconv.ParseInt(string(enc), 10, 64); err == nil {
		t.Errorf("encode(true) = %q, which parses as an int; booleans must be pickled", enc)
	}
}

// TestUncompressedIsStillReadable covers django-redis's read path, which
// tolerates a value that was never compressed.
func TestUncompressedIsStillReadable(t *testing.T) {
	writer := codec{compress: false}
	reader := codec{compress: true}

	enc, err := writer.encode(map[string]any{"written": "uncompressed"})
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	got, err := reader.decode(enc)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !reflect.DeepEqual(got, map[string]any{"written": "uncompressed"}) {
		t.Errorf("decode = %#v", got)
	}
}

func TestRedisKeyNaming(t *testing.T) {
	h := New(nil)
	if got := h.RedisKey("abc"); got != ":1:abc" {
		t.Errorf("default RedisKey = %q, want %q", got, ":1:abc")
	}

	h = New(nil, WithKeyPrefix("myapp"), WithCacheVersion(2))
	if got := h.RedisKey("abc"); got != "myapp:2:abc" {
		t.Errorf("prefixed RedisKey = %q, want %q", got, "myapp:2:abc")
	}

	// The double-prefixed namespace produced by tags that route through
	// FragmentedKeysCacheBackend's pass-through methods.
	h = New(nil, WithKeyFunc(DoublePrefixedKeyFunc("", 1)))
	if got := h.RedisKey("Page_1GNP"); got != ":1::1:Page_1GNP" {
		t.Errorf("double-prefixed RedisKey = %q, want %q", got, ":1::1:Page_1GNP")
	}
}

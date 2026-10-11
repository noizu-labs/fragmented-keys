// Package djangoredis is a fragkeys.CacheHandler that reads and writes the
// exact byte format django-redis uses, so a Go service can share a Redis cache
// with a Django service — same keys, same values, in both directions.
//
// # What Django actually stores
//
// Two transformations sit between a Django cache.set() call and Redis, and
// both have to be reproduced:
//
//  1. Key naming. django-redis rewrites every key as
//     "{KEY_PREFIX}:{version}:{key}". Both parts default to "" and 1, so a
//     key named "abc" lands in Redis as ":1:abc". Note that Django reads
//     KEY_PREFIX, not PREFIX — a settings file specifying "PREFIX" (as
//     gnp-backend's does, for the fragmented-keys tag suffix) leaves the
//     Django key prefix empty.
//
//  2. Value encoding. Non-bool ints are stored as their raw decimal text, so
//     Redis INCR works on them. Everything else is pickled with
//     pickle.DEFAULT_PROTOCOL and then run through the configured compressor.
//     With ZlibCompressor that means zlib at level 6, applied only when the
//     pickled form exceeds 15 bytes.
//
// On read, django-redis tries int() first, then decompresses (tolerating an
// uncompressed value), then unpickles. This package does the same, which means
// values it writes uncompressed are still readable by Django.
package djangoredis

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"strconv"

	"github.com/noizu-labs/fragmented-keys-go/pypickle"
)

const (
	// compressMinLength matches django_redis.compressors.zlib.ZlibCompressor's
	// min_length: values at or below this are stored uncompressed.
	compressMinLength = 15

	// compressLevel matches that compressor's preset.
	compressLevel = 6

	// maxDecompressed caps a decompressed value, so a malicious or corrupt
	// entry cannot exhaust memory.
	maxDecompressed = 128 << 20
)

// codec encodes and decodes cache values the way django-redis does.
type codec struct {
	// compress mirrors the COMPRESSOR setting. When false, values are written
	// as plain pickle. Reads tolerate either regardless, so this only affects
	// what this process writes.
	compress bool
}

// encode converts a Go value into the bytes django-redis would store.
func (c codec) encode(v any) ([]byte, error) {
	// django_redis.client.default.encode: non-bool ints bypass pickling
	// entirely so that INCR/DECR work against them.
	if n, ok := asPlainInt(v); ok {
		return []byte(strconv.FormatInt(n, 10)), nil
	}

	raw, err := pypickle.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("djangoredis: pickle: %w", err)
	}
	if !c.compress || len(raw) <= compressMinLength {
		return raw, nil
	}

	var buf bytes.Buffer
	zw, err := zlib.NewWriterLevel(&buf, compressLevel)
	if err != nil {
		return nil, fmt.Errorf("djangoredis: zlib writer: %w", err)
	}
	if _, err := zw.Write(raw); err != nil {
		return nil, fmt.Errorf("djangoredis: zlib write: %w", err)
	}
	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("djangoredis: zlib close: %w", err)
	}
	return buf.Bytes(), nil
}

// decode converts stored bytes back into a Go value, following django-redis's
// int-then-decompress-then-unpickle order.
func (c codec) decode(b []byte) (any, error) {
	if n, err := strconv.ParseInt(string(b), 10, 64); err == nil {
		return n, nil
	}

	raw := b
	if plain, ok := tryDecompress(b); ok {
		raw = plain
	}
	v, err := pypickle.Unmarshal(raw)
	if err != nil {
		return nil, fmt.Errorf("djangoredis: unpickle: %w", err)
	}
	return v, nil
}

// tryDecompress reports whether b was a zlib stream, and if so returns the
// inflated bytes. A non-zlib value is not an error: django-redis stores short
// values uncompressed and swallows the resulting decompression failure.
func tryDecompress(b []byte) ([]byte, bool) {
	zr, err := zlib.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, false
	}
	defer zr.Close()

	out, err := io.ReadAll(io.LimitReader(zr, maxDecompressed))
	if err != nil {
		return nil, false
	}
	return out, true
}

// asPlainInt reports whether v is an integer Django would store unpickled.
// Booleans are excluded because Python's isinstance(True, int) is true but
// django-redis special-cases them back into the pickle path.
func asPlainInt(v any) (int64, bool) {
	switch n := v.(type) {
	case int:
		return int64(n), true
	case int8:
		return int64(n), true
	case int16:
		return int64(n), true
	case int32:
		return int64(n), true
	case int64:
		return n, true
	case uint8:
		return int64(n), true
	case uint16:
		return int64(n), true
	case uint32:
		return int64(n), true
	case uint:
		if uint64(n) <= 1<<62 {
			return int64(n), true
		}
	case uint64:
		if n <= 1<<62 {
			return int64(n), true
		}
	}
	return 0, false
}

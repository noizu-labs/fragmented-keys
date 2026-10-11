package djangoredis

import (
	"context"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/noizu-labs/fragmented-keys-go/fragkeys"
	"github.com/redis/go-redis/v9"
)

// requireRedis skips unless FK_REDIS_ADDR names a reachable Redis.
//
//	docker compose up -d redis          # from the fragmented-keys root
//	FK_REDIS_ADDR=127.0.0.1:6380 go test ./djangoredis/
//
// The tests write under a dedicated DB (FK_REDIS_DB, default 15) and flush it
// on entry, so point this at a throwaway instance — never at a shared cache.
func requireRedis(t *testing.T) (*redis.Client, context.Context) {
	t.Helper()
	addr := os.Getenv("FK_REDIS_ADDR")
	if addr == "" {
		t.Skip("set FK_REDIS_ADDR to run the Redis integration tests")
	}
	db := 15
	if s := os.Getenv("FK_REDIS_DB"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil {
			t.Fatalf("FK_REDIS_DB=%q is not a number", s)
		}
		db = n
	}

	rdb := redis.NewClient(&redis.Options{Addr: addr, DB: db})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis at %s is not reachable: %v", addr, err)
	}
	if err := rdb.FlushDB(ctx).Err(); err != nil {
		t.Fatalf("flush test db: %v", err)
	}
	t.Cleanup(func() { _ = rdb.Close() })
	return rdb, ctx
}

func TestIntegrationTagLifecycle(t *testing.T) {
	rdb, ctx := requireRedis(t)

	h := New(rdb, WithCompression(true))
	fragkeys.Config.SetDefaultCacheHandler(h)
	fragkeys.Config.SetGlobalPrefix("GNP")
	t.Cleanup(fragkeys.Config.Reset)

	tag := fragkeys.NewStandardTag("Page", "careers", fragkeys.WithHandler(h))
	key := fragkeys.NewKey("content_page", tag)

	before, err := key.String(ctx)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	// The seeded version must have been written where Django would look.
	physical := h.RedisKey("Page_careersGNP")
	if n, err := rdb.Exists(ctx, physical).Result(); err != nil || n != 1 {
		t.Fatalf("expected the tag version at %q (exists=%d, err=%v)", physical, n, err)
	}

	if err := h.SetWithTags(ctx, "content_page", map[string]any{"slug": "careers"}, []fragkeys.Tag{tag}, time.Minute); err != nil {
		t.Fatalf("SetWithTags: %v", err)
	}
	got, found, err := h.GetWithTags(ctx, "content_page", []fragkeys.Tag{tag})
	if err != nil || !found {
		t.Fatalf("GetWithTags: found=%v err=%v", found, err)
	}
	if !reflect.DeepEqual(got, map[string]any{"slug": "careers"}) {
		t.Errorf("GetWithTags = %#v", got)
	}

	// Invalidate, then confirm a freshly built tag resolves elsewhere and the
	// old entry is simply orphaned rather than deleted.
	if err := h.InvalidateTags(ctx, []fragkeys.Tag{tag}); err != nil {
		t.Fatalf("InvalidateTags: %v", err)
	}
	fresh := fragkeys.NewStandardTag("Page", "careers", fragkeys.WithHandler(h))
	after, err := fragkeys.NewKey("content_page", fresh).String(ctx)
	if err != nil {
		t.Fatalf("resolve after invalidation: %v", err)
	}
	if after == before {
		t.Error("key did not change after InvalidateTags")
	}
	if _, found, _ := h.GetWithTags(ctx, "content_page", []fragkeys.Tag{fresh}); found {
		t.Error("the new key already has a value; invalidation did not produce a miss")
	}
	if n, _ := rdb.Exists(ctx, h.RedisKey(before)).Result(); n != 1 {
		t.Error("the old entry was deleted; fragmented keys must orphan it and let TTL reclaim it")
	}
}

func TestIntegrationTTLIsApplied(t *testing.T) {
	rdb, ctx := requireRedis(t)
	h := New(rdb)

	if err := h.SetValue(ctx, "ttl-probe", "v", 30*time.Second); err != nil {
		t.Fatalf("SetValue: %v", err)
	}
	ttl, err := rdb.TTL(ctx, h.RedisKey("ttl-probe")).Result()
	if err != nil {
		t.Fatalf("TTL: %v", err)
	}
	if ttl <= 0 || ttl > 30*time.Second {
		t.Errorf("TTL = %v, want (0, 30s]", ttl)
	}
}

func TestIntegrationGetMultiBulkFetch(t *testing.T) {
	rdb, ctx := requireRedis(t)
	h := New(rdb)

	for i, v := range []string{"1.0", "2.0", "3.0"} {
		if err := h.Set(ctx, "T_"+strconv.Itoa(i)+"GNP", v, 0); err != nil {
			t.Fatalf("Set: %v", err)
		}
	}
	got, err := h.GetMulti(ctx, []string{"T_0GNP", "T_1GNP", "T_2GNP", "T_missingGNP"})
	if err != nil {
		t.Fatalf("GetMulti: %v", err)
	}
	want := map[string]string{"T_0GNP": "1.0", "T_1GNP": "2.0", "T_2GNP": "3.0"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetMulti = %#v, want %#v (missing keys must be omitted, not nil)", got, want)
	}
}

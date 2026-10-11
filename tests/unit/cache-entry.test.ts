import { describe, it, expect } from "vitest";
import {
  createCacheEntry,
  serializeCacheEntry,
  deserializeCacheEntry,
} from "../../src/cache-entry/cache-entry";

describe("CacheEntry", () => {
  it("creates entry with defaults", () => {
    const entry = createCacheEntry("hello");
    expect(entry.status).toBe("green");
    expect(entry.meta).toEqual({});
    expect(entry.entry).toBe("hello");
    expect(entry.cached).toBeGreaterThan(0);
    expect(entry.dataType).toBeUndefined();
    expect(entry.dataTypeVersion).toBeUndefined();
  });

  it("creates entry with overrides", () => {
    const entry = createCacheEntry({ id: 1 }, {
      status: "error",
      meta: { reason: "timeout" },
      dataType: "User",
      dataTypeVersion: "2.0",
    });
    expect(entry.status).toBe("error");
    expect(entry.meta).toEqual({ reason: "timeout" });
    expect(entry.dataType).toBe("User");
    expect(entry.dataTypeVersion).toBe("2.0");
    expect(entry.entry).toEqual({ id: 1 });
  });

  it("round-trips through serialization", () => {
    const original = createCacheEntry({ name: "test" }, {
      dataType: "Widget",
      dataTypeVersion: "1.0",
    });
    const json = serializeCacheEntry(original);
    const restored = deserializeCacheEntry(json);
    expect(restored.entry).toEqual({ name: "test" });
    expect(restored.status).toBe("green");
    expect(restored.dataType).toBe("Widget");
    expect(restored.cached).toBe(original.cached);
  });

  it("creates empty entry", () => {
    const entry = createCacheEntry(null, { status: "empty" });
    expect(entry.status).toBe("empty");
    expect(entry.entry).toBeNull();
  });
});

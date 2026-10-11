import { describe, it, expect, beforeEach, afterAll } from "vitest";
import Redis from "ioredis";
import { RedisHandler } from "../../src/cache-handler/redis-handler";

const TEST_PREFIX = "fk-test:";

describe("RedisHandler", () => {
  const redis = new Redis({ host: "localhost", port: 6379, lazyConnect: true });
  let handler: RedisHandler;

  beforeEach(async () => {
    await redis.connect().catch(() => {});
    const keys = await redis.keys(`${TEST_PREFIX}*`);
    if (keys.length > 0) await redis.del(...keys);
    handler = new RedisHandler(redis, "test-redis");
  });

  afterAll(async () => {
    const keys = await redis.keys(`${TEST_PREFIX}*`);
    if (keys.length > 0) await redis.del(...keys);
    await redis.quit();
  });

  it("returns group name", () => {
    expect(handler.groupName()).toBe("test-redis");
  });

  it("returns null for missing key", async () => {
    expect(await handler.get(`${TEST_PREFIX}nonexistent`)).toBeNull();
  });

  it("round-trips set and get", async () => {
    await handler.set(`${TEST_PREFIX}key1`, "value1");
    expect(await handler.get(`${TEST_PREFIX}key1`)).toBe("value1");
  });

  it("overwrites existing key", async () => {
    await handler.set(`${TEST_PREFIX}key1`, "v1");
    await handler.set(`${TEST_PREFIX}key1`, "v2");
    expect(await handler.get(`${TEST_PREFIX}key1`)).toBe("v2");
  });

  it("respects TTL", async () => {
    await handler.set(`${TEST_PREFIX}ttl-key`, "val", 1);
    expect(await handler.get(`${TEST_PREFIX}ttl-key`)).toBe("val");
    const ttl = await redis.ttl(`${TEST_PREFIX}ttl-key`);
    expect(ttl).toBeGreaterThan(0);
    expect(ttl).toBeLessThanOrEqual(1);
  });

  it("getMulti returns only existing keys", async () => {
    await handler.set(`${TEST_PREFIX}a`, "1");
    await handler.set(`${TEST_PREFIX}c`, "3");
    const result = await handler.getMulti([
      `${TEST_PREFIX}a`,
      `${TEST_PREFIX}b`,
      `${TEST_PREFIX}c`,
    ]);
    expect(result).toEqual({
      [`${TEST_PREFIX}a`]: "1",
      [`${TEST_PREFIX}c`]: "3",
    });
  });

  it("getMulti returns empty for empty input", async () => {
    const result = await handler.getMulti([]);
    expect(result).toEqual({});
  });

  it("getMulti returns empty when all keys missing", async () => {
    const result = await handler.getMulti([`${TEST_PREFIX}x`, `${TEST_PREFIX}y`]);
    expect(result).toEqual({});
  });
});

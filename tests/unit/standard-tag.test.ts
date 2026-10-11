import { describe, it, expect, beforeEach } from "vitest";
import { StandardTag } from "../../src/tag/standard-tag";
import { MemoryHandler } from "../../src/cache-handler/memory-handler";
import { Configuration } from "../../src/configuration";

describe("StandardTag", () => {
  let handler: MemoryHandler;

  beforeEach(() => {
    handler = new MemoryHandler();
    Configuration.setDefaultCacheHandler(handler);
  });

  it("returns consistent version without increment", async () => {
    const tag = new StandardTag("User", "123");
    const v1 = await tag.getTagVersion();
    const v2 = await tag.getTagVersion();
    expect(v1).toBe(v2);
  });

  it("seeds version from Date.now() on first access", async () => {
    const before = Date.now();
    const tag = new StandardTag("User", "456");
    const version = await tag.getTagVersion();
    const after = Date.now();
    expect(version).toBeGreaterThanOrEqual(before);
    expect(version).toBeLessThanOrEqual(after);
  });

  it("uses explicit version when provided", async () => {
    const tag = new StandardTag("User", "789", 42.0);
    expect(await tag.getTagVersion()).toBe(42.0);
  });

  it("increment changes version by 0.1", async () => {
    const tag = new StandardTag("User", "100", 10.0);
    await tag.increment();
    const version = await tag.getTagVersion();
    expect(version).toBeCloseTo(10.1, 5);
  });

  it("increment persists to cache", async () => {
    const tag = new StandardTag("User", "101", 5.0);
    await tag.increment();
    const raw = await handler.get(tag.getTagName());
    expect(raw).not.toBeNull();
    expect(Number.parseFloat(raw!)).toBeCloseTo(5.1, 5);
  });

  it("different instances have different tag names", () => {
    const t1 = new StandardTag("User", "aaa");
    const t2 = new StandardTag("User", "bbb");
    expect(t1.getTagName()).not.toBe(t2.getTagName());
  });

  it("different tag types have different tag names", () => {
    const t1 = new StandardTag("User", "123");
    const t2 = new StandardTag("Site", "123");
    expect(t1.getTagName()).not.toBe(t2.getTagName());
  });

  it("incrementing one instance does not affect another", async () => {
    const t1 = new StandardTag("User", "a", 1.0);
    const t2 = new StandardTag("User", "b", 1.0);
    await t1.increment();
    expect(await t1.getTagVersion()).toBeCloseTo(1.1, 5);
    expect(await t2.getTagVersion()).toBe(1.0);
  });

  it("incrementing one tag type does not affect another", async () => {
    const t1 = new StandardTag("User", "x", 1.0);
    const t2 = new StandardTag("Site", "x", 1.0);
    await t1.increment();
    expect(await t1.getTagVersion()).toBeCloseTo(1.1, 5);
    expect(await t2.getTagVersion()).toBe(1.0);
  });

  it("resetTagVersion produces new timestamp-based value", async () => {
    const tag = new StandardTag("User", "rst", 1.0);
    const before = Date.now();
    await tag.resetTagVersion();
    const version = await tag.getTagVersion();
    const after = Date.now();
    expect(version).toBeGreaterThanOrEqual(before);
    expect(version).toBeLessThanOrEqual(after);
  });

  it("delegateCacheQuery returns true for matching group", () => {
    const tag = new StandardTag("User", "dq");
    expect(tag.delegateCacheQuery("MemoryHandler")).toBe(true);
  });

  it("delegateCacheQuery returns false for non-matching group", () => {
    const tag = new StandardTag("User", "dq");
    expect(tag.delegateCacheQuery("RedisHandler")).toBe(false);
  });

  it("getTagName formats correctly", () => {
    const tag = new StandardTag("User", "123", undefined, undefined, "Test");
    expect(tag.getTagName()).toBe("User_123Test");
  });

  it("getFullTag formats correctly", async () => {
    const tag = new StandardTag("User", "123", 3.5);
    expect(await tag.getFullTag()).toBe("User_123DefaultPrefix@3.5");
  });
});

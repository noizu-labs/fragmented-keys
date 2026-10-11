import { describe, it, expect, beforeEach } from "vitest";
import { ConstantTag } from "../../src/tag/constant-tag";
import { MemoryHandler } from "../../src/cache-handler/memory-handler";
import { Configuration } from "../../src/configuration";

describe("ConstantTag", () => {
  let handler: MemoryHandler;

  beforeEach(() => {
    handler = new MemoryHandler();
    Configuration.setDefaultCacheHandler(handler);
  });

  it("returns default version 1.0", async () => {
    const tag = new ConstantTag("User", "123");
    expect(await tag.getTagVersion()).toBe(1.0);
  });

  it("returns custom version", async () => {
    const tag = new ConstantTag("User", "123", 5.0);
    expect(await tag.getTagVersion()).toBe(5.0);
  });

  it("increment is a no-op", async () => {
    const tag = new ConstantTag("User", "123", 2.0);
    await tag.increment();
    expect(await tag.getTagVersion()).toBe(2.0);
  });

  it("resetTagVersion is a no-op", async () => {
    const tag = new ConstantTag("User", "123", 3.0);
    await tag.resetTagVersion();
    expect(await tag.getTagVersion()).toBe(3.0);
  });

  it("setTagVersion is a no-op", async () => {
    const tag = new ConstantTag("User", "123", 4.0);
    await tag.setTagVersion(99.0, true);
    expect(await tag.getTagVersion()).toBe(4.0);
  });

  it("delegateCacheQuery always returns false", () => {
    const tag = new ConstantTag("User", "123");
    expect(tag.delegateCacheQuery("MemoryHandler")).toBe(false);
    expect(tag.delegateCacheQuery("RedisHandler")).toBe(false);
  });

  it("does not read from or write to cache", async () => {
    const tag = new ConstantTag("User", "123", 7.0);
    await tag.getTagVersion();
    await tag.increment();
    expect(await handler.get(tag.getTagName())).toBeNull();
  });

  it("getTagName formats correctly", () => {
    const tag = new ConstantTag("User", "123", 1.0, undefined, "MyPrefix");
    expect(tag.getTagName()).toBe("User_123MyPrefix");
  });

  it("getFullTag formats correctly", async () => {
    const tag = new ConstantTag("User", "123", 2.5);
    expect(await tag.getFullTag()).toBe("User_123DefaultPrefix@2.5");
  });
});

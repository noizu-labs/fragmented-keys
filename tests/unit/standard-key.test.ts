import { describe, it, expect, beforeEach } from "vitest";
import { StandardKey } from "../../src/key/standard-key";
import { StandardTag } from "../../src/tag/standard-tag";
import { ConstantTag } from "../../src/tag/constant-tag";
import { MemoryHandler } from "../../src/cache-handler/memory-handler";
import { Configuration } from "../../src/configuration";

describe("StandardKey", () => {
  let handler: MemoryHandler;

  beforeEach(() => {
    handler = new MemoryHandler();
    Configuration.setDefaultCacheHandler(handler);
  });

  it("returns consistent key without increment", async () => {
    const tag = new StandardTag("User", "123", 1.0);
    const key = new StandardKey("Profile", [tag]);
    const k1 = await key.getKeyStr();
    const k2 = await key.getKeyStr();
    expect(k1).toBe(k2);
  });

  it("returns 32-char MD5 hex", async () => {
    const tag = new StandardTag("User", "123", 1.0);
    const key = new StandardKey("Profile", [tag]);
    const hash = await key.getKeyStr();
    expect(hash).toMatch(/^[a-f0-9]{32}$/);
  });

  it("returns different key after increment", async () => {
    const tag = new StandardTag("User", "123", 1.0);
    const key = new StandardKey("Profile", [tag]);
    const before = await key.getKeyStr();
    await tag.increment();
    const after = await new StandardKey("Profile", [tag]).getKeyStr();
    expect(before).not.toBe(after);
  });

  it("composes multiple tags", async () => {
    const t1 = new StandardTag("User", "123", 1.0);
    const t2 = new StandardTag("Site", "us", 2.0);
    const key = new StandardKey("Dashboard", [t1, t2]);
    const hash = await key.getKeyStr();
    expect(hash).toMatch(/^[a-f0-9]{32}$/);
  });

  it("constant tags produce stable keys", async () => {
    const t1 = new StandardTag("User", "123", 1.0);
    const t2 = new ConstantTag("Version", "", 5.0);
    const key = new StandardKey("Config", [t1, t2]);
    const k1 = await key.getKeyStr();
    await t2.increment();
    const k2 = await new StandardKey("Config", [t1, t2]).getKeyStr();
    expect(k1).toBe(k2);
  });

  it("raw key string contains expected substrings", async () => {
    const t1 = new StandardTag("User", "123", 1.0);
    const key = new StandardKey("Profile", [t1], "grp");
    const raw = await key.getKeyStr(false);
    expect(raw).toContain("Profile_grp");
    expect(raw).toContain(":t");
    expect(raw).toContain(":v1");
  });

  it("addTag appends to composition", async () => {
    const t1 = new StandardTag("User", "123", 1.0);
    const key = new StandardKey("Profile", [t1]);
    const before = await key.getKeyStr(false);

    const t2 = new StandardTag("Site", "us", 2.0);
    key.addTag(t2);
    const after = await key.getKeyStr(false);
    expect(after.length).toBeGreaterThan(before.length);
    expect(after).toContain("Site");
  });

  it("includes groupId in raw key", async () => {
    const tag = new StandardTag("User", "123", 1.0);
    const key = new StandardKey("Profile", [tag], "mygroup");
    const raw = await key.getKeyStr(false);
    expect(raw).toContain("Profile_mygroup");
  });

  it("works with mixed handlers", async () => {
    const handler2 = new MemoryHandler();
    const t1 = new StandardTag("User", "123", 1.0, handler);
    const t2 = new StandardTag("Site", "us", 2.0, handler2);
    const key = new StandardKey("Mixed", [t1, t2]);
    const k1 = await key.getKeyStr();
    const k2 = await key.getKeyStr();
    expect(k1).toBe(k2);
  });

  it("different after increment with mixed handlers", async () => {
    const handler2 = new MemoryHandler();
    const t1 = new StandardTag("User", "123", 1.0, handler);
    const t2 = new StandardTag("Site", "us", 2.0, handler2);
    const before = await new StandardKey("Mixed", [t1, t2]).getKeyStr();
    await t1.increment();
    const after = await new StandardKey("Mixed", [t1, t2]).getKeyStr();
    expect(before).not.toBe(after);
  });
});

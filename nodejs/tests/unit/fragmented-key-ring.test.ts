import { describe, it, expect, beforeEach } from "vitest";
import { FragmentedKeyRing, createKeyRing } from "../../src/key-ring/fragmented-key-ring";
import { StandardTag } from "../../src/tag/standard-tag";
import { ConstantTag } from "../../src/tag/constant-tag";
import { MemoryHandler } from "../../src/cache-handler/memory-handler";
import { Configuration } from "../../src/configuration";

describe("FragmentedKeyRing", () => {
  let handler: MemoryHandler;
  let ring: FragmentedKeyRing;

  beforeEach(() => {
    handler = new MemoryHandler();
    Configuration.setDefaultCacheHandler(handler);
    ring = new FragmentedKeyRing({
      defaultCacheHandler: "memory",
      cacheHandlers: { memory: handler },
    });
  });

  it("defineKey + getKeyObj produces valid 32-char hash", async () => {
    ring.defineKey("UserProfile", ["user", "site"]);
    const key = ring.getKeyObj("UserProfile", ["123", "us"]);
    const hash = await key.getKeyStr();
    expect(hash).toMatch(/^[a-f0-9]{32}$/);
  });

  it("getKeyObj produces consistent results", async () => {
    ring.defineKey("UserProfile", ["user", "site"]);
    const k1 = ring.getKeyObj("UserProfile", ["123", "us"]);
    const k2 = ring.getKeyObj("UserProfile", ["123", "us"]);
    expect(await k1.getKeyStr()).toBe(await k2.getKeyStr());
  });

  it("keyring key matches manually constructed key", async () => {
    ring.defineKey("Simple", ["user"]);
    const ringKey = ring.getKeyObj("Simple", ["abc"]);

    const manualTag = ring.tag("user", "abc");
    const { StandardKey } = await import("../../src/key/standard-key");
    const manualKey = new StandardKey("Simple", [manualTag]);

    expect(await ringKey.getKeyStr()).toBe(await manualKey.getKeyStr());
  });

  it("constant tag via globalTagOptions type override", async () => {
    const ringWithConst = new FragmentedKeyRing({
      defaultCacheHandler: "memory",
      cacheHandlers: { memory: handler },
      globalTagOptions: { version: { type: "constant", version: 5.0 } },
    });
    ringWithConst.defineKey("Versioned", ["version", "user"]);
    const key = ringWithConst.getKeyObj("Versioned", ["", "123"]);
    const h1 = await key.getKeyStr();

    const tag = ringWithConst.tag("version", "");
    await tag.increment();
    const key2 = ringWithConst.getKeyObj("Versioned", ["", "123"]);
    const h2 = await key2.getKeyStr();
    expect(h1).toBe(h2);
  });

  it("handler override via param dict", async () => {
    const handler2 = new MemoryHandler();
    const ringMulti = new FragmentedKeyRing({
      defaultCacheHandler: "memory",
      cacheHandlers: { memory: handler, alt: handler2 },
    });
    ringMulti.defineKey("Multi", [
      "user",
      { tag: "site", cacheHandler: "alt" },
    ]);
    const key = ringMulti.getKeyObj("Multi", ["123", "us"]);
    const hash = await key.getKeyStr();
    expect(hash).toMatch(/^[a-f0-9]{32}$/);
  });

  it("listDefinedKeys empty initially", () => {
    const empty = new FragmentedKeyRing({
      defaultCacheHandler: "memory",
      cacheHandlers: { memory: handler },
    });
    expect(empty.listDefinedKeys()).toEqual({});
  });

  it("listDefinedKeys returns populated definitions", () => {
    ring.defineKey("A", ["x"]);
    ring.defineKey("B", ["y", "z"]);
    const defs = ring.listDefinedKeys();
    expect(Object.keys(defs)).toEqual(["A", "B"]);
  });

  it("listDefinedKeys returns deep copy", () => {
    ring.defineKey("A", ["x"]);
    const d1 = ring.listDefinedKeys();
    const d2 = ring.listDefinedKeys();
    expect(d1).toEqual(d2);
    expect(d1).not.toBe(d2);
  });

  it("throws for undefined key", () => {
    expect(() => ring.getKeyObj("NotDefined", ["a"])).toThrow();
  });

  it("throws for wrong argument count", () => {
    ring.defineKey("TwoArgs", ["x", "y"]);
    expect(() => ring.getKeyObj("TwoArgs", ["only_one"])).toThrow();
  });

  it("throws for unknown handler", () => {
    expect(
      () =>
        new FragmentedKeyRing({
          defaultCacheHandler: "nonexistent",
          cacheHandlers: { memory: handler },
        }),
    ).toThrow();
  });

  it("setTagOptions / getTagOptions round-trip", () => {
    ring.setTagOptions("user", { type: "constant", version: 3.0 });
    const opts = ring.getTagOptions("user");
    expect(opts.type).toBe("constant");
    expect(opts.version).toBe(3.0);
  });

  it("setGlobalOptions / getGlobalOptions round-trip", () => {
    ring.setGlobalOptions({ type: "constant" });
    expect(ring.getGlobalOptions().type).toBe("constant");
  });

  it("3-tier option merging: inline > per-tag > global", async () => {
    const ringMerge = new FragmentedKeyRing({
      defaultCacheHandler: "memory",
      cacheHandlers: { memory: handler },
      globalOptions: { type: "standard" },
      globalTagOptions: { special: { version: 10.0 } },
    });
    ringMerge.defineKey("Merged", [
      { tag: "special", version: 99.0 },
    ]);
    const key = ringMerge.getKeyObj("Merged", [""]);
    const raw = await key.getKeyStr(false);
    expect(raw).toContain(":v99");
  });
});

describe("createKeyRing (Proxy)", () => {
  let handler: MemoryHandler;

  beforeEach(() => {
    handler = new MemoryHandler();
    Configuration.setDefaultCacheHandler(handler);
  });

  it("dynamic accessor getUserProfileKeyObj works", async () => {
    const ring = createKeyRing({
      defaultCacheHandler: "memory",
      cacheHandlers: { memory: handler },
    });
    ring.defineKey("UserProfile", ["user", "site"]);
    const key = ring.getUserProfileKeyObj("123", "us");
    const hash = await key.getKeyStr();
    expect(hash).toMatch(/^[a-f0-9]{32}$/);
  });

  it("dynamic accessor matches getKeyObj", async () => {
    const ring = createKeyRing({
      defaultCacheHandler: "memory",
      cacheHandlers: { memory: handler },
    });
    ring.defineKey("UserProfile", ["user", "site"]);
    const k1 = ring.getKeyObj("UserProfile", ["abc", "eu"]);
    const k2 = ring.getUserProfileKeyObj("abc", "eu");
    expect(await k1.getKeyStr()).toBe(await k2.getKeyStr());
  });
});

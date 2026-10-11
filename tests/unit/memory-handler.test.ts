import { describe, it, expect, beforeEach } from "vitest";
import { MemoryHandler } from "../../src/cache-handler/memory-handler";

describe("MemoryHandler", () => {
  let handler: MemoryHandler;

  beforeEach(() => {
    handler = new MemoryHandler();
  });

  it("returns MemoryHandler as group name", () => {
    expect(handler.groupName()).toBe("MemoryHandler");
  });

  it("returns null for missing key", async () => {
    expect(await handler.get("nonexistent")).toBeNull();
  });

  it("round-trips set and get", async () => {
    await handler.set("key1", "value1");
    expect(await handler.get("key1")).toBe("value1");
  });

  it("overwrites existing key", async () => {
    await handler.set("key1", "v1");
    await handler.set("key1", "v2");
    expect(await handler.get("key1")).toBe("v2");
  });

  it("getMulti returns only existing keys", async () => {
    await handler.set("a", "1");
    await handler.set("c", "3");
    const result = await handler.getMulti(["a", "b", "c"]);
    expect(result).toEqual({ a: "1", c: "3" });
  });

  it("getMulti returns empty for all missing", async () => {
    const result = await handler.getMulti(["x", "y"]);
    expect(result).toEqual({});
  });

  it("clear empties all entries", async () => {
    await handler.set("k1", "v1");
    await handler.set("k2", "v2");
    handler.clear();
    expect(await handler.get("k1")).toBeNull();
    expect(await handler.get("k2")).toBeNull();
  });

  it("sync get/set round-trips", () => {
    handler.setSync("key1", "value1");
    expect(handler.getSync("key1")).toBe("value1");
  });

  it("sync getMulti works", () => {
    handler.setSync("a", "1");
    handler.setSync("b", "2");
    const result = handler.getMultiSync(["a", "b", "missing"]);
    expect(result).toEqual({ a: "1", b: "2" });
  });
});

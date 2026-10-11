import { describe, it, expect, beforeEach } from "vitest";
import { Configuration } from "../../src/configuration";
import { MemoryHandler } from "../../src/cache-handler/memory-handler";

describe("Configuration", () => {
  beforeEach(() => {
    Configuration.reset();
  });

  it("throws when getDefaultCacheHandler called before set", () => {
    expect(() => Configuration.getDefaultCacheHandler()).toThrow();
  });

  it("round-trips default cache handler", () => {
    const handler = new MemoryHandler();
    Configuration.setDefaultCacheHandler(handler);
    expect(Configuration.getDefaultCacheHandler()).toBe(handler);
  });

  it("returns default global prefix", () => {
    expect(Configuration.getGlobalPrefix()).toBe("DefaultPrefix");
  });

  it("round-trips global prefix", () => {
    Configuration.setGlobalPrefix("MyApp");
    expect(Configuration.getGlobalPrefix()).toBe("MyApp");
  });

  it("returns empty cache prefix by default", () => {
    expect(Configuration.getCachePrefix()).toBe("");
  });

  it("round-trips cache prefix", () => {
    Configuration.setCachePrefix("myapp:staging:");
    expect(Configuration.getCachePrefix()).toBe("myapp:staging:");
  });

  it("reset clears handler and restores defaults", () => {
    const handler = new MemoryHandler();
    Configuration.setDefaultCacheHandler(handler);
    Configuration.setGlobalPrefix("Custom");
    Configuration.setCachePrefix("ns:");
    Configuration.reset();

    expect(() => Configuration.getDefaultCacheHandler()).toThrow();
    expect(Configuration.getGlobalPrefix()).toBe("DefaultPrefix");
    expect(Configuration.getCachePrefix()).toBe("");
  });
});

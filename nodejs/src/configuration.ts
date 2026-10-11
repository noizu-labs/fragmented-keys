import type { CacheHandler } from "./interfaces/cache-handler";

export class Configuration {
  private static _defaultCacheHandler: CacheHandler | null = null;
  private static _globalPrefix = "DefaultPrefix";
  private static _cachePrefix = "";

  static setDefaultCacheHandler(handler: CacheHandler): void {
    Configuration._defaultCacheHandler = handler;
  }

  static getDefaultCacheHandler(): CacheHandler {
    if (!Configuration._defaultCacheHandler) {
      throw new Error(
        "Default cache handler not configured. Call Configuration.setDefaultCacheHandler() first.",
      );
    }
    return Configuration._defaultCacheHandler;
  }

  static setGlobalPrefix(prefix: string): void {
    Configuration._globalPrefix = prefix;
  }

  static getGlobalPrefix(): string {
    return Configuration._globalPrefix;
  }

  static setCachePrefix(prefix: string): void {
    Configuration._cachePrefix = prefix;
  }

  static getCachePrefix(): string {
    return Configuration._cachePrefix;
  }

  static reset(): void {
    Configuration._defaultCacheHandler = null;
    Configuration._globalPrefix = "DefaultPrefix";
    Configuration._cachePrefix = "";
  }
}

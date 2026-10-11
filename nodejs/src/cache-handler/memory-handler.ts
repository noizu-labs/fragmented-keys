import type { CacheHandler } from "../interfaces/cache-handler";
import type { SyncCacheHandler } from "../interfaces/cache-handler-sync";

export class MemoryHandler implements CacheHandler, SyncCacheHandler {
  private _cache = new Map<string, string>();

  groupName(): string {
    return "MemoryHandler";
  }

  async get(key: string): Promise<string | null> {
    return this.getSync(key);
  }

  async set(key: string, value: string, _ttl?: number): Promise<void> {
    this.setSync(key, value, _ttl);
  }

  async getMulti(keys: string[]): Promise<Record<string, string>> {
    return this.getMultiSync(keys);
  }

  getSync(key: string): string | null {
    return this._cache.get(key) ?? null;
  }

  setSync(key: string, value: string, _ttl?: number): void {
    this._cache.set(key, value);
  }

  getMultiSync(keys: string[]): Record<string, string> {
    const result: Record<string, string> = {};
    for (const key of keys) {
      const value = this._cache.get(key);
      if (value !== undefined) {
        result[key] = value;
      }
    }
    return result;
  }

  clear(): void {
    this._cache.clear();
  }
}

import type { CacheHandler } from "../interfaces/cache-handler";
import type Redis from "ioredis";

export class RedisHandler implements CacheHandler {
  private _client: Redis;
  private _group: string;

  constructor(client: Redis, group = "RedisHandler") {
    this._client = client;
    this._group = group;
  }

  groupName(): string {
    return this._group;
  }

  async get(key: string): Promise<string | null> {
    return this._client.get(key);
  }

  async set(key: string, value: string, ttl?: number): Promise<void> {
    if (ttl && ttl > 0) {
      await this._client.set(key, value, "EX", ttl);
    } else {
      await this._client.set(key, value);
    }
  }

  async getMulti(keys: string[]): Promise<Record<string, string>> {
    if (keys.length === 0) return {};
    const values = await this._client.mget(...keys);
    const result: Record<string, string> = {};
    for (let i = 0; i < keys.length; i++) {
      const v = values[i];
      const k = keys[i];
      if (v !== null && v !== undefined && k !== undefined) {
        result[k] = v;
      }
    }
    return result;
  }
}

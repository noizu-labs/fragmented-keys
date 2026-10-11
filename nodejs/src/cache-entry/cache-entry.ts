import type { CacheEntryData, CacheEntryStatus } from "../interfaces/types";

export interface CreateCacheEntryOptions {
  status?: CacheEntryStatus;
  meta?: Record<string, unknown>;
  dataType?: string;
  dataTypeVersion?: string;
  cached?: number;
}

export function createCacheEntry<T>(
  entry: T,
  options: CreateCacheEntryOptions = {},
): CacheEntryData<T> {
  return {
    cached: options.cached ?? Date.now(),
    status: options.status ?? "green",
    meta: options.meta ?? {},
    dataType: options.dataType,
    dataTypeVersion: options.dataTypeVersion,
    entry,
  };
}

export function serializeCacheEntry<T>(entry: CacheEntryData<T>): string {
  return JSON.stringify(entry);
}

export function deserializeCacheEntry<T = unknown>(json: string): CacheEntryData<T> {
  return JSON.parse(json) as CacheEntryData<T>;
}

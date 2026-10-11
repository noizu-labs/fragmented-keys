import type { CacheHandler } from "./cache-handler";

export interface Tag {
  getTagName(): string;
  getTagVersion(): Promise<number>;
  getFullTag(): Promise<string>;
  increment(): Promise<void>;
  resetTagVersion(): Promise<void>;
  setTagVersion(version: number, update?: boolean): Promise<void>;
  setCacheHandler(handler: CacheHandler): void;
  getCacheHandler(): CacheHandler;
  delegateCacheQuery(group: string): boolean;
}

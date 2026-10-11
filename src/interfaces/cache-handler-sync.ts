export interface SyncCacheHandler {
  groupName(): string;
  getSync(key: string): string | null;
  setSync(key: string, value: string, ttl?: number): void;
  getMultiSync(keys: string[]): Record<string, string>;
}

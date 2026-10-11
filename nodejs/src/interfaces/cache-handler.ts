export interface CacheHandler {
  groupName(): string;
  get(key: string): Promise<string | null>;
  set(key: string, value: string, ttl?: number): Promise<void>;
  getMulti(keys: string[]): Promise<Record<string, string>>;
}

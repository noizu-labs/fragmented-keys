export type TagType = "standard" | "constant";

export type CacheEntryStatus = "green" | "empty" | "error" | "unknown";

export interface TagOptions {
  type?: TagType;
  version?: number;
  cacheHandler?: string;
  prefix?: string;
  shardKey?: string;
}

export interface KeyDefinitionParam {
  tag: string;
  [key: string]: unknown;
}

export type KeyParam = string | KeyDefinitionParam;

export interface KeyDefinition {
  params: KeyParam[];
  globals: TagOptions;
}

export interface ShardConfig {
  name: string;
  handler: import("./cache-handler").CacheHandler;
  tags?: string[];
  default?: boolean;
}

export interface CacheEntryData<T = unknown> {
  cached: number;
  status: CacheEntryStatus;
  meta: Record<string, unknown>;
  dataType?: string;
  dataTypeVersion?: string;
  entry: T;
}

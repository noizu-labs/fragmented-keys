import type { Key } from "./key";
import type { Tag } from "./tag";
import type { KeyDefinition, KeyParam, TagOptions } from "./types";

export interface KeyRing {
  defineKey(key: string, params: KeyParam[], globals?: TagOptions): void;
  getKeyObj(key: string, tagValues: string[]): Key;
  tag(tag: string, instance: string, options?: TagOptions): Tag;
  listDefinedKeys(): Record<string, KeyDefinition>;
  setTagOptions(tag: string, options: TagOptions): void;
  getTagOptions(tag: string, extra?: TagOptions): TagOptions;
  setGlobalOptions(options: TagOptions): void;
  getGlobalOptions(): TagOptions;
}

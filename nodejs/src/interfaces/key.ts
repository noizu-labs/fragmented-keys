import type { Tag } from "./tag";

export interface Key {
  getKeyStr(hash?: boolean): Promise<string>;
  addTag(tag: Tag): void;
}

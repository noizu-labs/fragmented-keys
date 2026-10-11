import type { CacheHandler } from "../interfaces/cache-handler";
import { BaseTag } from "./base-tag";

export class ConstantTag extends BaseTag {
  constructor(
    tag: string,
    instance = "",
    version = 1.0,
    handler?: CacheHandler | null,
    prefix?: string,
  ) {
    super(tag, instance, version, handler, prefix);
  }

  async getTagVersion(): Promise<number> {
    return this._version!;
  }

  async increment(): Promise<void> {}
  async resetTagVersion(): Promise<void> {}

  async setTagVersion(_version: number, _update?: boolean): Promise<void> {}

  delegateCacheQuery(_group: string): boolean {
    return false;
  }
}

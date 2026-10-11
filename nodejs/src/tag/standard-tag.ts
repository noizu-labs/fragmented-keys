import type { CacheHandler } from "../interfaces/cache-handler";
import { BaseTag } from "./base-tag";

export class StandardTag extends BaseTag {
  constructor(
    tag: string,
    instance = "",
    version?: number,
    handler?: CacheHandler | null,
    prefix?: string,
  ) {
    super(tag, instance, version ?? null, handler, prefix);
  }

  async increment(): Promise<void> {
    const current = await this.getTagVersion();
    const next = current + 0.1;
    this._version = next;
    await this._storeVersion(next);
  }

  async resetTagVersion(): Promise<void> {
    const seed = Date.now();
    this._version = seed;
    await this._storeVersion(seed);
  }
}

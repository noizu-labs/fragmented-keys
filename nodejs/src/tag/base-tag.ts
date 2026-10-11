import type { CacheHandler } from "../interfaces/cache-handler";
import type { Tag } from "../interfaces/tag";
import { Configuration } from "../configuration";

export abstract class BaseTag implements Tag {
  protected _tag: string;
  protected _instance: string;
  protected _version: number | null;
  protected _handler: CacheHandler | null;
  protected _prefix: string;

  constructor(
    tag: string,
    instance = "",
    version?: number | null,
    handler?: CacheHandler | null,
    prefix?: string,
  ) {
    this._tag = tag;
    this._instance = instance;
    this._version = version ?? null;
    this._handler = handler ?? null;
    this._prefix = prefix ?? Configuration.getGlobalPrefix();
  }

  getTagName(): string {
    return `${this._tag}_${this._instance}${this._prefix}`;
  }

  async getTagVersion(): Promise<number> {
    if (this._version === null) {
      this._version = await this._getVersion();
    }
    return this._version;
  }

  async getFullTag(): Promise<string> {
    const version = await this.getTagVersion();
    return `${this.getTagName()}@${version}`;
  }

  getCacheHandler(): CacheHandler {
    return this._handler ?? Configuration.getDefaultCacheHandler();
  }

  setCacheHandler(handler: CacheHandler): void {
    this._handler = handler;
  }

  delegateCacheQuery(group: string): boolean {
    return this.getCacheHandler().groupName() === group;
  }

  protected async _getVersion(): Promise<number> {
    const handler = this.getCacheHandler();
    const raw = await handler.get(this.getTagName());
    if (raw !== null) {
      return Number.parseFloat(raw);
    }
    const seed = Date.now();
    await this._storeVersion(seed);
    return seed;
  }

  protected async _storeVersion(version: number): Promise<void> {
    const handler = this.getCacheHandler();
    await handler.set(this.getTagName(), String(version));
  }

  async setTagVersion(version: number, update = false): Promise<void> {
    this._version = version;
    if (update) {
      await this._storeVersion(version);
    }
  }

  abstract increment(): Promise<void>;
  abstract resetTagVersion(): Promise<void>;
}

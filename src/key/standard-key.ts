import { createHash } from "node:crypto";
import type { Key } from "../interfaces/key";
import type { Tag } from "../interfaces/tag";

export class StandardKey implements Key {
  private _key: string;
  private _tags: Tag[];
  private _groupId: string;

  constructor(key: string, tags: Tag[] = [], groupId = "") {
    this._key = key;
    this._tags = [...tags];
    this._groupId = groupId;
  }

  addTag(tag: Tag): void {
    this._tags.push(tag);
  }

  async getKeyStr(hash = true): Promise<string> {
    await this._gatherGroupVersions();
    const parts = [`${this._key}_${this._groupId}`];
    for (const tag of this._tags) {
      const version = await tag.getTagVersion();
      parts.push(`:t${tag.getTagName()}:v${version}`);
    }
    const raw = parts.join("");
    if (!hash) {
      return raw;
    }
    return createHash("md5").update(raw).digest("hex");
  }

  private async _gatherGroupVersions(): Promise<void> {
    const groups = new Map<string, Tag[]>();
    for (const tag of this._tags) {
      const groupName = tag.getCacheHandler().groupName();
      if (!tag.delegateCacheQuery(groupName)) {
        continue;
      }
      const existing = groups.get(groupName);
      if (existing) {
        existing.push(tag);
      } else {
        groups.set(groupName, [tag]);
      }
    }

    for (const [, tags] of groups) {
      if (tags.length === 0) continue;
      const handler = tags[0]!.getCacheHandler();
      const keys = tags.map((t) => t.getTagName());
      const versions = await handler.getMulti(keys);
      for (const tag of tags) {
        const raw = versions[tag.getTagName()];
        if (raw !== undefined) {
          await tag.setTagVersion(Number.parseFloat(raw));
        }
      }
    }
  }
}

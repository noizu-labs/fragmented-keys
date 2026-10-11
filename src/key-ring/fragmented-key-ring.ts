import type { CacheHandler } from "../interfaces/cache-handler";
import type { Key } from "../interfaces/key";
import type { KeyRing } from "../interfaces/key-ring";
import type { Tag } from "../interfaces/tag";
import type { KeyDefinition, KeyParam, TagOptions } from "../interfaces/types";
import { StandardKey } from "../key/standard-key";
import { ConstantTag } from "../tag/constant-tag";
import { StandardTag } from "../tag/standard-tag";

export interface KeyRingOptions {
  globalOptions?: TagOptions;
  globalTagOptions?: Record<string, TagOptions>;
  defaultCacheHandler?: string;
  cacheHandlers?: Record<string, CacheHandler>;
  defaultPrefix?: string;
}

export class FragmentedKeyRing implements KeyRing {
  private _globalOptions: TagOptions;
  private _globalTagOptions: Record<string, TagOptions>;
  private _defaultHandlerName: string;
  private _cacheHandlers: Record<string, CacheHandler>;
  private _defaultPrefix: string;
  private _keyDefinitions: Record<string, KeyDefinition> = {};

  constructor(options: KeyRingOptions = {}) {
    this._globalOptions = options.globalOptions ?? {};
    this._globalTagOptions = options.globalTagOptions ?? {};
    this._defaultHandlerName = options.defaultCacheHandler ?? "memory";
    this._cacheHandlers = options.cacheHandlers ?? {};
    this._defaultPrefix = options.defaultPrefix ?? "DefaultPrefix";

    this._resolveHandler(this._defaultHandlerName);
  }

  private _resolveHandler(name?: string): CacheHandler {
    const handlerName = name ?? this._defaultHandlerName;
    const handler = this._cacheHandlers[handlerName];
    if (!handler) {
      throw new Error(`Unknown cache handler: "${handlerName}"`);
    }
    return handler;
  }

  private _mergeOptions(tagName: string, extra?: TagOptions): TagOptions {
    const global = this._globalOptions;
    const perTag = this._globalTagOptions[tagName] ?? {};
    return { ...global, ...perTag, ...extra };
  }

  setTagOptions(tag: string, options: TagOptions): void {
    this._globalTagOptions[tag] = options;
  }

  getTagOptions(tag: string, extra?: TagOptions): TagOptions {
    return this._mergeOptions(tag, extra);
  }

  setGlobalOptions(options: TagOptions): void {
    this._globalOptions = options;
  }

  getGlobalOptions(): TagOptions {
    return { ...this._globalOptions };
  }

  tag(tagName: string, instance: string, options?: TagOptions): Tag {
    const merged = this._mergeOptions(tagName, options);
    const handler = this._resolveHandler(merged.cacheHandler);
    const prefix = merged.prefix ?? this._defaultPrefix;

    if (merged.type === "constant") {
      return new ConstantTag(tagName, instance, merged.version, handler, prefix);
    }
    return new StandardTag(tagName, instance, merged.version, handler, prefix);
  }

  defineKey(key: string, params: KeyParam[], globals?: TagOptions): void {
    this._keyDefinitions[key] = {
      params,
      globals: globals ?? {},
    };
  }

  getKeyObj(key: string, tagValues: string[]): Key {
    const definition = this._keyDefinitions[key];
    if (!definition) {
      throw new Error(`Key "${key}" is not defined`);
    }
    if (definition.params.length !== tagValues.length) {
      throw new Error(
        `Key "${key}" expects ${definition.params.length} tag values, got ${tagValues.length}`,
      );
    }

    const tags: Tag[] = [];
    for (let i = 0; i < definition.params.length; i++) {
      const param = definition.params[i]!;
      const instance = tagValues[i]!;

      let tagName: string;
      let inlineOptions: TagOptions = {};

      if (typeof param === "string") {
        tagName = param;
      } else {
        tagName = param.tag;
        const { tag: _tag, ...rest } = param;
        inlineOptions = rest as TagOptions;
      }

      const merged = this._mergeOptions(tagName, {
        ...definition.globals,
        ...inlineOptions,
      });
      const handler = this._resolveHandler(merged.cacheHandler);
      const prefix = merged.prefix ?? this._defaultPrefix;

      let createdTag: Tag;
      if (merged.type === "constant") {
        createdTag = new ConstantTag(tagName, instance, merged.version, handler, prefix);
      } else {
        createdTag = new StandardTag(tagName, instance, merged.version, handler, prefix);
      }
      tags.push(createdTag);
    }

    return new StandardKey(key, tags);
  }

  listDefinedKeys(): Record<string, KeyDefinition> {
    return structuredClone(this._keyDefinitions);
  }
}

type DynamicKeyRing = FragmentedKeyRing & Record<string, (...args: string[]) => Key>;

export function createKeyRing(options?: KeyRingOptions): DynamicKeyRing {
  const ring = new FragmentedKeyRing(options);
  return new Proxy(ring, {
    get(target, prop, receiver) {
      if (
        typeof prop === "string" &&
        prop.startsWith("get") &&
        prop.endsWith("KeyObj")
      ) {
        const rawName = prop.slice(3, -6);
        if (!rawName) return Reflect.get(target, prop, receiver);
        return (...args: string[]) => {
          const definitions = target.listDefinedKeys();
          const match = Object.keys(definitions).find(
            (k) => k.toLowerCase() === rawName.toLowerCase(),
          );
          if (!match) {
            throw new Error(`Key "${rawName}" is not defined`);
          }
          return target.getKeyObj(match, args);
        };
      }
      return Reflect.get(target, prop, receiver);
    },
  }) as DynamicKeyRing;
}

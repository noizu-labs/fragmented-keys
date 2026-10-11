//
//  File.swift
//  
//
//  Created by Keith Brings on 5/2/24.
//

import Foundation



protocol DataStore {
    associatedtype Config: DataStoreConfig
    
    func incrementTag(_ key: RecordHandle) -> TagVersion
    func currentTag(_ key: RecordHandle) -> TagVersion
    func get<T: Serializable>(_ key: RecordHandle) -> T?
    func getCacheEntryItem(_ key: RecordHandle) -> CacheEntryItem
    func getCacheEntry(_ key: RecordHandle) -> CacheEntry?
    
    func set(_ key: RecordHandle, to: Serializable?, options: DataStoreSetOptions?) -> Void
    func erase(_ key: RecordHandle) -> Void
    
    func multiGet(_ keys: [RecordHandle]) -> [MultiGetResponse]
}

extension DataStore {
    
        
    func get<T: Serializable>(_ key: RecordHandle) -> T? {
        switch getCacheEntryItem(key) {
        case .empty:
            break;
        case .raw(let d):
            return T.deserialize(data: d)
        case .entity(let o):
            return o as? T
        }
        return nil
    }
    
    func getCacheEntryItem(_ key: RecordHandle) -> CacheEntryItem {
        guard let c = getCacheEntry(key) else {
            return .empty
        }
        
        switch (c.ttl) {
        case .infinity:
            return c.entry
        case .expiry(let i):
            if (c.storedAt + i) >= Date().timeIntervalSince1970 {
                return c.entry
            } else {
                set(key, to: nil, options: SetOptions(ttl: .infinity))
                return .empty
            }
        }
    }
    
    func incrementTag(_ key: RecordHandle) -> TagVersion {
        if let c = getCacheEntry(key) {
            switch c.entry {
            case .empty:
                break;
            case .raw(let d):
                if var t: TagVersion = TagVersion.deserialize(data: d) {
                    _ = t.incrementTagVersion()
                    set(key, to: t, options: SetOptions(ttl: c.ttl))
                    return t
                }
            case .entity(let o):
                if var t: TagVersion = o as? TagVersion {
                    _ = t.incrementTagVersion()
                    set(key, to: t, options: SetOptions(ttl: c.ttl))
                    return t
                }
            }
        }
        var y = TagVersion(version: 1);
        _ = y.incrementTagVersion()
        set(key, to: y, options: SetOptions(ttl: .infinity))
        return y
    }
    
    func currentTag(_ key: RecordHandle) -> TagVersion {
        if let c = getCacheEntry(key) {
            switch c.entry {
            case .empty:
                break;
            case .raw(let d):
                if let t: TagVersion = TagVersion.deserialize(data: d) {
                    return t
                }
            case .entity(let o):
                if let t: TagVersion = o as? TagVersion {
                    return t
                }
            }
        }
        var t = TagVersion(version: 1);
        _ = t.incrementTagVersion()
        set(key, to: t, options: SetOptions(ttl: .infinity))
        return t
    }
    
    func multiGet(_ keys: [RecordHandle]) -> [MultiGetResponse] {
        return keys.map {
            key in
            switch getCacheEntryItem(key) {
            case .empty:
                return .miss(key)
            case .raw(_):
                return .miss(key) // return to do this
            case .entity(let o):
                return .hit(key, o)
            }
        }
    }
}

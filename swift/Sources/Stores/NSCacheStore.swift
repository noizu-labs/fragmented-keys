//
//  File.swift
//  
//
//  Created by Keith Brings on 5/2/24.
//

import Foundation


/// MARK: - NSCache
struct NSCacheStoreConfig : DataStoreConfig {
    var name: String? = nil
    var costLimit: Int? = nil
    var countLimit: Int? = nil
    var delegate: NSCacheDelegate?
    var ttl: TTL = .infinity
    
    init(name: String? = nil, costLimit: Int? = nil, countLimit: Int? = nil, delegate: NSCacheDelegate? = nil, ttl: TTL = .infinity) {
        self.name = name
        self.costLimit = costLimit
        self.countLimit = countLimit
        self.delegate = delegate
        self.ttl = ttl
    }
    init?(from: [String : Any?]) {
        if let name = from["name"] as? String {
            self.name = name
        }
        if let cost = from["costLimit"] as? Int {
            self.costLimit = cost
        }
        if let count = from["countLimit"] as? Int {
            self.countLimit = count
        }
        if let d = from["delegate"] as? NSCacheDelegate {
            self.delegate = d
        }
        if let ttl = from["ttl"] as? TTL {
            self.ttl = ttl
        }
    }
    
}


class NSCacheStore : DataStore {
    
    
    typealias Config = NSCacheStoreConfig
    private let cache = NSCache<NSString, AnyObject>()
    private var config: Config
    
    init(config: Config? = nil) {
        if let c = config {
            self.config = c
            if let n = c.name {
                cache.name = n
            }
            if let cost = c.costLimit {
                cache.totalCostLimit = cost
            }
            if let count = c.countLimit {
                cache.countLimit = count
            }
            if let d = c.delegate {
                cache.delegate = d
            }
        } else {
            self.config = Config()
        }
    }
    
    
    func getCacheEntry(_ key: any RecordHandle) -> CacheEntry? {
        guard let c = cache.object(forKey: key.handleName() as NSString) as? CacheEntry else {
            return nil
        }
        return c
    }

    func set(_ key: RecordHandle, to: Serializable?, options: DataStoreSetOptions? = nil) {
        if let o = options {
            let entry = CacheEntry(entry: to, ttl: o.ttl ?? config.ttl)
            cache.setObject(entry as AnyObject, forKey: key.handleName() as NSString)
        } else {
            let entry = CacheEntry(entry: to, ttl: config.ttl)
            self.cache.setObject(entry as AnyObject, forKey: key.handleName() as NSString)
        }
    }
    
    func erase(_ key: RecordHandle) {
        cache.removeObject(forKey: key.handleName() as NSString)
    }
    
    
}

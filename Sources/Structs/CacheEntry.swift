//
//  File.swift
//  
//
//  Created by Keith Brings on 5/2/24.
//

import Foundation

enum CacheEntryItem {
    case empty
    case raw(_ data: Data)
    case entity(_ object: Serializable)
}

struct CacheEntry: Codable {
    var entry: CacheEntryItem
    var storedAt: TimeInterval
    var ttl: TTL
    init(entry: Serializable?, ttl: TTL = .infinity) {
        if let e = entry {
            self.entry = .entity(e)
        } else {
            self.entry = .empty
        }
        self.ttl = ttl
        self.storedAt = Date().timeIntervalSince1970
    }
    
    enum CodingKeys: String, CodingKey {
            case entryData
            case storedAt
            case ttl
        }
        
    func encode(to encoder: Encoder) throws {
        var container = encoder.container(keyedBy: CodingKeys.self)
        // Encode `storedAt` and `ttl` normally
        try container.encode(storedAt, forKey: .storedAt)
        try container.encode(ttl, forKey: .ttl)
        // Encode `entry` by converting it to Data via the `serialize` method
        switch entry {
        case .empty:
            try container.encodeNil(forKey: .entryData)
        case .raw(let raw):
            try container.encode(raw, forKey: .entryData)
        case .entity(let o):
            let data = o.serialize()
            try container.encode(data, forKey: .entryData)
        }
    }
        
        init(from decoder: Decoder) throws {
            let container = try decoder.container(keyedBy: CodingKeys.self)
            storedAt = try container.decode(TimeInterval.self, forKey: .storedAt)
            ttl = try container.decode(TTL.self, forKey: .ttl)
            if let entryData = try container.decodeIfPresent(Data.self, forKey: .entryData) {
                entry = .raw(entryData)
            } else {
                entry = .empty
            }
        }
}

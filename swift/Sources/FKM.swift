//
//  File.swift
//  
//
//  Created by Keith Brings on 4/24/24.
//

import Foundation


protocol Telemetry {
    
}

class StandardTelemetry : Telemetry {
    
}

class FKM {
    //---------------
    // Static
    //---------------
    static var _instance: FKM = FKM()
    static var instance: FKM {
        get {
            return _instance
        }
    }
    static func configure(tagStore: (any DataStore)? = nil, recordStore: (any DataStore)? = nil) {
        _instance = FKM(tagStore: tagStore, recordStore: recordStore)
    }
    
    //----------------
    // Init
    //----------------
    init(
        tagStore: (any DataStore)? = nil,
        recordStore: (any DataStore)? = nil
    ) {
        if let ts = tagStore {
            self.tagStore = ts
        } else {
            self.tagStore = NSCacheStore()
        }
        
        if let rs = recordStore {
            self.recordStore = rs
        } else {
            self.recordStore = NSCacheStore()
        }
    }
    
    //----------------
    // Members
    //----------------
    let tagStore: any DataStore
    let recordStore: any DataStore
    var tags: [KeyRingTag] = []
    var subjects: [String:Taggable] = [:]

    
}


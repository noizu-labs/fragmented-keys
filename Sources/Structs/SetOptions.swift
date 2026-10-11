//
//  File.swift
//  
//
//  Created by Keith Brings on 5/2/24.
//

import Foundation

struct SetOptions : DataStoreSetOptions {
    var ttl: TTL?
    
    init(ttl: TTL? = nil) {
        self.ttl = ttl
    }
    
    init?(from: [String : Any?]) {
        if let t = from["ttl"] as? TTL {
            self.ttl = t
        }
    }
}

//
//  File.swift
//  
//
//  Created by Keith Brings on 5/2/24.
//

import Foundation

protocol Settings {
    init?(from: [String: Any?]) throws
}


protocol DataStoreConfig : Settings {
    
}

protocol DataStoreOptions : Settings {
    
}

protocol DataStoreSetOptions : DataStoreOptions {
    var ttl: TTL? {get}
}

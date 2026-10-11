//
//  File.swift
//  
//
//  Created by Keith Brings on 5/2/24.
//

import Foundation


protocol KeyProtocol : RecordHandle {
//    var store: any DataStore {get}
    
    func get<T: Serializable>() -> T?
    func set(to: Serializable, options: DataStoreSetOptions?) -> Void
}

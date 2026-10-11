//
//  File.swift
//  
//
//  Created by Keith Brings on 5/2/24.
//

import Foundation

class Key : KeyProtocol {
    
    //--------------
    // Init
    //--------------
    init(_ name: String, forType: Any? = nil, tags: [TagProtocol], store: (any DataStore)? = nil) {
        self.name = name
        self.tags = tags
        self.forType = forType
        if let s = store {
            self.store = s
        } else {
            self.store = FKM.instance.recordStore
        }
    }
    
    //--------------
    // Members
    //--------------
    let store: any DataStore
    let name: String
    let tags: [TagProtocol]
    let forType: Any?
    //--------------
    // Methods
    //--------------
    func handleName() -> String {
        /// @TODO MultiFetch
        /// TEMP - iterate tags and tag tagVersion to insure populated.
        let f: [String] = tags.map {
            tag in
            return tag.tagFragment()
        }
        let frags = Array(Set(f)).sorted().joined(separator: ",")
        return "\(name)::\(frags)"
    }
    func recordType() -> Any?
    {
        return forType
    }

    func get<T: Serializable>() -> T? {
        return store.get(self)
    }
    
    func set(to: Serializable, options: DataStoreSetOptions? = nil) -> Void {
        store.set(self, to: to, options: options)
    }

}

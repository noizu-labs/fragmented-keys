//
//  File.swift
//  
//
//  Created by Keith Brings on 5/2/24.
//

import Foundation

class KeyRing : KeyRingProtocol {
    //--------------
    // Init
    //--------------
    init(
        tags: [KeyRingTag] = [],
        bind: [String:Taggable] = [:],
        recordStore: (any DataStore)? = nil,
        tagStore: (any DataStore)? = nil) {
        if let s = recordStore {
            self.recordStore = s
        } else {
            self.recordStore = FKM.instance.recordStore
        }
        if let ts = tagStore {
            self.tagStore = ts
        } else {
            self.tagStore = FKM.instance.tagStore
        }
        self.tags = tags
        self.subjects = bind
    }
    
    //--------------
    // Members
    //--------------
    let recordStore: any DataStore
    let tagStore: any DataStore
 
    let tags: [KeyRingTag]
    let subjects: [String:Taggable]
    
    //--------------
    // Methods
    //--------------
    func key(name: String, forType: Any? = nil, tags: [KeyRingTag], subjects: [String:Taggable]? = nil, store: (any DataStore)? = nil) throws -> Key {
        let t = tags + self.tags + FKM.instance.tags
        let t2 = try t.map {
            tag in
            return try tag.bind(keyRingObjects: self.subjects, keyObjects: subjects, store: self.tagStore)
        }
        if let s = store {
            return Key(name, forType: forType, tags: t2, store: s)
        } else {
            return Key(name, forType: forType, tags: t2, store: self.recordStore)
        }
    }
}

extension KeyRing {
    class Bound {
        class FixedTag : KeyRingTag {
            let name: String
            let subject: String
            let version: TagVersion
            init(name: String, bindToSubject: String, version: TagVersion = TagVersion(version: 1)) {
                self.name = name
                self.subject = bindToSubject
                self.version = version
            }
            func bind(keyRingObjects: [String:Taggable], keyObjects: [String:Taggable]?, store: (any DataStore)) throws -> TagProtocol {
                if let o = keyObjects {
                    if let k = o[subject] {
                        return FragmentedKeys.FixedTag(name: name, subject: k, version: self.version)
                    }
                }
                if let r = keyRingObjects[subject] {
                    return FragmentedKeys.FixedTag(name: name, subject: r, version: self.version)
                }
                if let f = FKM.instance.subjects[subject] {
                    return FragmentedKeys.FixedTag(name: name, subject: f, version: self.version)
                }
                // FKM bind
                throw FragmentedKeyError.error(message: "Required Object \(subject) not available on keyRing")
            }
        }
    }
}

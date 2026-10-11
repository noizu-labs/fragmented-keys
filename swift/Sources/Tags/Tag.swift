//
//  File.swift
//  
//
//  Created by Keith Brings on 5/2/24.
//

import Foundation


class Tag : TagProtocol {
    private let _store: any DataStore
    private let name: String
    private let subject: String?
    private var version: TagVersion? = nil
    init(name: String, subject: Taggable? = nil, store: (any DataStore)? = nil) {
        self.name = name
        if let s = subject {
            self.subject = s.tagHandle()
        } else {
            self.subject = nil
        }
        if let s = store {
            self._store = s
        } else {
            // get from global scope
            var s: (any DataStore)
            do {
                s = try FKM.instance.tagStore
            } catch {
                s = NSCacheStore()
            }
            self._store = s
        }
    }
    
    var store: (any DataStore)? {
        get {
            return _store
        }
    }
    
    func handleName() -> String {
        if let subject = self.subject {
            return "\(self.name).\(subject)"
        } else {
            return self.name
        }
    }
    func recordType() -> Any?
    {
        return TagVersion.self
    }
    
    func tagFragment() -> String {
        return "\(handleName())@\(tagVersion().value)"
    }
    func tagVersion() -> TagVersion {
        guard let v = version else {
            version = _store.currentTag(self)
            return version!
        }
        return v
    }
    func increment() -> Void {
        version = _store.incrementTag(self)
    }
}

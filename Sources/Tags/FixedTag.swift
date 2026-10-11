//
//  File.swift
//  
//
//  Created by Keith Brings on 5/2/24.
//

import Foundation


class FixedTag : TagProtocol {
    
    private let name: String
    private let subject: String?
    private let version: TagVersion
    
    convenience init(name: String, subject: Taggable? = nil) {
        self.init(name: name, subject: subject, version: TagVersion(version: 1))
    }
    convenience init(name: String, subject: Taggable? = nil, version: UInt) {
        self.init(name: name, subject: subject, version: TagVersion(version: version))
    }
    init(name: String, subject: Taggable? = nil, version: TagVersion) {
        self.name = name
        if let s = subject {
            self.subject = s.tagHandle()
        } else {
            self.subject = nil
        }
        self.version = version
    }
    
    var store: (any DataStore)? {
        get {
            return nil
        }
    }
    
    func handleName() -> String {
        if let subject = self.subject {
            return "\(self.name):\(subject)"
        } else {
            return self.name
        }
    }
    func recordType() -> Any?
    {
        return nil
    }

    func tagFragment() -> String {
        return "\(handleName())@\(version.value)"
    }
    func tagVersion() -> TagVersion {
        return self.version
    }
}


//
//  File.swift
//  
//
//  Created by Keith Brings on 5/2/24.
//

import Foundation

protocol KeyRingProtocol {
}

protocol KeyRingTag {
    func bind(keyRingObjects: [String:Taggable], keyObjects: [String:Taggable]?, store: (any DataStore)) throws -> TagProtocol
}

extension FixedTag: KeyRingTag {
    func bind(keyRingObjects: [String:Taggable], keyObjects: [String:Taggable]? = nil, store: (any DataStore)) -> TagProtocol {
        return self
    }
}


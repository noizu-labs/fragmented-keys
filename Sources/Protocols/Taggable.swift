//
//  File.swift
//  
//
//  Created by Keith Brings on 5/2/24.
//

import Foundation

protocol Taggable {
    func tagHandle() -> String
}

extension String: Taggable {
    func tagHandle() -> String {
        return self
    }
}

//
//  File.swift
//  
//
//  Created by Keith Brings on 5/2/24.
//

import Foundation


protocol TagProtocol : RecordHandle {
    //init?(name: String, subject: Taggable?, store: (any DataStore)?)
    var store: (any DataStore)? {get}
    func tagFragment() -> String
    func tagVersion() -> TagVersion
}


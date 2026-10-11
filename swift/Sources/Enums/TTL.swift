//
//  File.swift
//  
//
//  Created by Keith Brings on 5/2/24.
//

import Foundation


enum TTL : Codable {
    case infinity
    case expiry(TimeInterval)
}

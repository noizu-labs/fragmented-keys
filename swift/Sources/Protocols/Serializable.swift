//
//  File.swift
//  
//
//  Created by Keith Brings on 5/2/24.
//

import Foundation





protocol Serializable {    
    func serialize() -> Data
    static func deserialize(data: Data) -> Self?
}

extension String: Serializable {
    func serialize() -> Data {
        guard let x = self.data(using: .utf32) else {
            return "[FKM:nil]".data(using: .utf32)!
        }
        return x
    }
    static func deserialize(data: Data) -> String? {
        return String(data: data, encoding: .utf32)
    }
}

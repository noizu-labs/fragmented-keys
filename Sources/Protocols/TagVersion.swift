//
//  File.swift
//  
//
//  Created by Keith Brings on 5/2/24.
//

import Foundation



struct TagVersion : Serializable {
    var value: UInt

    init(version: UInt) {
        self.value = version
    }
    
    static func deserialize(data: Data) -> TagVersion? {
        guard data.count == MemoryLayout<UInt>.size else {
            return nil
        }
        var intValue: UInt = 0
        _ = withUnsafeMutableBytes(of: &intValue) { data.copyBytes(to: $0) }
        return TagVersion(version: UInt(littleEndian: intValue))
    }
    
    func serialize() -> Data {
        var littleEndianValue = value.littleEndian
        return Data(bytes: &littleEndianValue, count: MemoryLayout<Int>.size)
    }

    mutating func incrementTagVersion() -> UInt {
        let t = Date().timeIntervalSince1970
        self.value = UInt(t * 10000000)
        return self.value
    }
}
//
//extension Int: Serializable {
//    func serialize() -> Data {
//        var bigEndian = self.bigEndian
//        return Data(bytes: &bigEndian, count: MemoryLayout<Int>.size)
//    }
//}
//
//extension String: TagVersion {
//    func incrementTagVersion() -> String {
//        do {
//            let regex = try NSRegularExpression(pattern: "-(\\d+)$", options: [])
//            let range = NSRange(location: 0, length: self.utf16.count)
//            
//            if let match = regex.firstMatch(in: self, options: [], range: range) {
//                let lastNumberRange = match.range(at: 1)
//                if let swiftRange = Range(lastNumberRange, in: self) {
//                    let lastNumber = Int(self[swiftRange])!
//                    let incrementedNumber = lastNumber + 1
//                    return self.replacingCharacters(in: swiftRange, with: String(incrementedNumber))
//                }
//            }
//        } catch {
//            
//        }
//        // If the string does not end with a number, append "-1"
//        return self + "-1"
//    }
//}
//
//extension UUID: TagVersion {
//    func incrementTagVersion() -> UUID {
//        return UUID()
//    }
//}

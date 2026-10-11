//
//  File.swift
//  
//
//  Created by Keith Brings on 5/2/24.
//

import Foundation


enum MultiGetResponse {
    case hit(RecordHandle, Serializable)
    case miss(RecordHandle)
}

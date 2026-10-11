//
//  RecordHandle.swift
//  
//
//  Created by Keith Brings on 5/2/24.
//

import Foundation


protocol RecordHandle {
    func handleName() -> String
    func recordType() -> Any?
}

extension String : RecordHandle {
    func handleName() -> String {
        return self;
    }
    func recordType() -> Any?
    {
        return nil
    }
}

//
//  File.swift
//
//
//  Created by Keith Brings on 5/2/24.
//

import XCTest
@testable import FragmentedKeys


class KeyTests: XCTestCase {
    //    func testExample() throws {
    //        // XCTest Documentation
    //        // https://developer.apple.com/documentation/xctest
    //
    //        // Defining Test Cases and Test Methods
    //        // https://developer.apple.com/documentation/xctest/defining_test_cases_and_test_methods
    //    }
    //
    
    override func setUp() {
//        FKMTest.reset()
        FKM.configure()
    }
    
    func testKeyRing() throws {
        let k = Key("Moop", forType: String.self, tags: [FixedTag(name: "Glob", subject: "123", version: 7)])
        k.set(to: "Bannana")
        let k2 = Key("Boop", forType: String.self, tags: [FixedTag(name: "Glob", subject: "123", version: 7)])
        k2.set(to: "Bonnana")

        let sut: String = k.get()!
        XCTAssertEqual(sut, "Bannana")
        let sut2: String = k2.get()!
        XCTAssertEqual(sut2, "Bonnana")

    }
}

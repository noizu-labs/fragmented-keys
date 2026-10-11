//
//  File.swift
//  
//
//  Created by Keith Brings on 5/2/24.
//

import XCTest
@testable import FragmentedKeys


class KeyRingTests: XCTestCase {
    //    func testExample() throws {
    //        // XCTest Documentation
    //        // https://developer.apple.com/documentation/xctest
    //
    //        // Defining Test Cases and Test Methods
    //        // https://developer.apple.com/documentation/xctest/defining_test_cases_and_test_methods
    //    }
    //
    
    override func setUp() {
        //FKMTest.reset()
        FKM.configure()
        let g = FooBip(identifier: 5, value: "Global'sFooBip")
        FKM.instance.subjects["FooBip"] = g
        super.setUp()
    }
    
    func testKeyRing() throws {
        let x = FizBop(identifier: 7, value: "Ohayoo")
        let b = BooBop(identifier: 12, value: "k")
        let sut = try x.keyRing.key(
            name: "MyKey",
            forType: FizBop.self,
            tags: [
                FixedTag(name: "Glob", subject: "123", version: 7),
                KeyRing.Bound.FixedTag(name: "FizBop", bindToSubject: "@")
            ],
            subjects: ["BooBop": b]
        )
        let key = sut.handleName()
        XCTAssertEqual(key, "MyKey::FizBop:ref.fiz-bop.7@1,Glob:123@7,Global@1,GlobalFooBip:ref.foo-bip.5@1,KeyBooBip:ref.boo-bop.12@1")
    }
}

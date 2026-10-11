import XCTest
@testable import FragmentedKeys



class FragmentedKeysTests: XCTestCase {
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
//        do {
//            try FKMDeprecated.initialize()
//        } catch {}
//        
        FKM.configure()
        super.setUp()
    }
    func testTagIncrement() {
        let b = BizBop(identifier: 1, value: "Apple")
        let sut = Tag(name: "Global", subject: b)
        let f = sut.tagFragment()
        sut.increment()
        let f2 = sut.tagFragment()
        XCTAssertNotEqual(f,f2)
        let f3 = sut.tagFragment()
        XCTAssertEqual(f2,f3)
    }
    
}

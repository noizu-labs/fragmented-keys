import XCTest
@testable import FragmentedKeys

class FileStoreTests: XCTestCase {

    func testMultiGet() throws {
        let ds = try FileStore(config: FileStoreConfig())
        
        var a = Tag(name: "FKM:FST:A", store: ds)
        var b = Tag(name: "FKM:FST:B", store: ds)
        var c = Tag(name: "FKM:FST:C", store: ds)

        let tags =  [a,b,c]
        _ = tags.map {
            tag in
            tag.increment()
        }
        let o = ds.multiGet(tags)
        switch o.last {
        case .hit(let t, let v):
            let t2 = t as! Tag
            let v2 = v as! TagVersion
            XCTAssertEqual(t2.handleName(), c.handleName())
            XCTAssertEqual(v2.value, c.tagVersion().value)
            break;
        default:
            XCTFail()
        }
    }
    
    
    func testSetGet() throws {
        let ds = try FileStore(config: FileStoreConfig())
        let f = Foo(bop: "Apple")
        ds.set("bAnna", to: f)
        guard let sut: Foo = ds.get("bAnna") else {
            XCTFail("Get from Cache Failed")
            return
        }
        XCTAssertEqual(sut.bop, "Apple")
        
    }
    
    
    func testExpiry() throws {
        let ds = try FileStore(config: FileStoreConfig())
        //let f = Foo(bop: "Apple")
        //ds.set("Anna", to: f, options: SetOptions(ttl: .expiry(0.0001)))
        ///usleep(100)
        let sut: Foo? = ds.get("Anna")
        XCTAssertNil(sut)
    }
    
}

//
//  File.swift
//  
//
//  Created by Keith Brings on 5/2/24.
//

import Foundation
@testable import FragmentedKeys

//class FKMTest : FKM {
//    static func reset() {
//        instance = nil
//    }
//}


struct Foo : Serializable {
    func serialize() -> Data {
        return bop.serialize()
    }
    static func deserialize(data: Data) -> Foo? {
        guard let s = String.deserialize(data: data) else {
            return nil
        }
        return Foo(bop: s)
    }
    
    
    init?(data: Data) {
        fatalError()
    }
    init(bop: String) {
        self.bop = bop
    }
    var bop: String
}



class BizBop : Taggable {
    let identifier: Int
    let value: String
    init(identifier: Int, value: String) {
        self.identifier = identifier
        self.value = value
    }
    func tagHandle() -> String {
        "ref.biz-bop.\(identifier)"
    }

}



class BooBop : Taggable {
    let identifier: Int
    let value: String
    init(identifier: Int, value: String) {
        self.identifier = identifier
        self.value = value
    }
    func tagHandle() -> String {
        "ref.boo-bop.\(identifier)"
    }

}


class FooBip : Taggable {
    let identifier: Int
    let value: String
    init(identifier: Int, value: String) {
        self.identifier = identifier
        self.value = value
    }
    func tagHandle() -> String {
        "ref.foo-bip.\(identifier)"
    }

}

class FizBop : Taggable {
    let identifier: Int
    let value: String
    var _keyRing: FizBop.KeyRing?
    var keyRing: FizBop.KeyRing {
        if nil == _keyRing {
            _keyRing = FizBop.KeyRing.init(self)
        }
        return _keyRing!
    }
    init(identifier: Int, value: String) {
        self.identifier = identifier
        self.value = value
    }
    func tagHandle() -> String {
        "ref.fiz-bop.\(identifier)"
    }

    class KeyRing: FragmentedKeys.KeyRing {
        init(
            _ me: FizBop,
            recordStore: (any DataStore)? = nil,
            tagStore: (any DataStore)? = nil
        ) {
            let t : [KeyRingTag] = [
                FixedTag(name: "Global"),
                FragmentedKeys.KeyRing.Bound.FixedTag(name: "KeyBooBip", bindToSubject: "BooBop"),
                FragmentedKeys.KeyRing.Bound.FixedTag(name: "GlobalFooBip", bindToSubject: "FooBip")
            ]
            let b = BizBop(identifier: 5, value: "FizBop'sBizBop")
            let s: [String: Taggable] = ["BizBop": b, "@": me]
            super.init(tags: t, bind: s, recordStore: recordStore, tagStore: tagStore)
        }
    }
}


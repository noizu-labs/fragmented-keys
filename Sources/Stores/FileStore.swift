//
//  File.swift
//  
//
//  Created by Keith Brings on 5/2/24.
//
import Foundation
import CryptoKit

/// MARK: - File Cache Store
struct FileStoreConfig: DataStoreConfig {
    init?(from: [String : Any?]) throws {
        let d: String? = from["cacheDirectory"] as? String
        let s: Bool = from["secure"] as? Bool ?? false
        let ttl: TTL = from["ttl"] as? TTL ?? .infinity
        try self.init(cacheDirectory: d, secure: s, ttl: ttl)
    }
    
    static func prepareCache(_ subfolder: String) throws -> URL {
        let fileManager = FileManager.default
        guard var cd = fileManager.urls(for: .cachesDirectory, in: .userDomainMask).first else {
            throw FragmentedKeyError.error(message: "Unable To Load File Cache")
        }
        
        if #available(iOS 16.0, macOS 13.0, *) {
            cd.append(component: "FKM", directoryHint: .isDirectory)
            cd.append(component: subfolder, directoryHint: .isDirectory)
        } else {
            cd = cd.appendingPathComponent("FKM", isDirectory: true)
            cd = cd.appendingPathComponent(subfolder, isDirectory: true)
        }
        
        //let cd = d.appending(path: "FKM").appending(path: subfolder)
        if !fileManager.fileExists(atPath: cd.path) {
            do {
                try fileManager.createDirectory(at: cd, withIntermediateDirectories: true)
            } catch {
                throw FragmentedKeyError.error(message: "Unable To Setup Cahde Dir \(cd.path)")
            }
        }
        return cd
    }

    // today need a unique/configurable version of this
    let namespace: UUID = UUID(uuidString: "6ba7b810-9dad-11d1-80b4-00c04fd430c8")!
    let cacheDirectory: URL
    let ttl: TTL
    let secure: Bool
    init(cacheDirectory: String? = nil, secure: Bool = false, ttl: TTL = .infinity) throws {
        self.secure = secure
        self.ttl = ttl
        if let c = cacheDirectory {
            try self.cacheDirectory = Self.prepareCache(c)
        } else {
            if secure {
                try self.cacheDirectory = Self.prepareCache("secure")
            } else {
                try self.cacheDirectory = Self.prepareCache("cache")
            }
        }
    }
}

extension UUID {
    @available(macOS 10.15, iOS 13, *)
    static func uuid5(namespace: UUID, name: String) -> UUID {

        var namespaceBytes = withUnsafeBytes(of: namespace.uuid) { Array($0) }
        let nameBytes = Array(name.utf8)
        namespaceBytes.append(contentsOf: nameBytes)
        
        let hash = Insecure.SHA1.hash(data: namespaceBytes)
        let hashBytes = Array(hash.prefix(16))
        
        // Set the version (5) and the variant (DCE 1.1)
        var result = hashBytes
        result[6] = (result[6] & 0x0F) | 0x50 // version 5
        result[8] = (result[8] & 0x3F) | 0x80 // variant is 10xxxxxx
        
        let uuid = result.withUnsafeBufferPointer { buffer -> UUID in
            let data = Data(buffer)
            return data.withUnsafeBytes { ptr -> UUID in
                ptr.load(as: UUID.self)
            }
        }
        
        return uuid
    }
}

@available(macOS 10.15, *)
class FileStore: DataStore {
    
    
    typealias Config = FileStoreConfig
    private var config: Config
    
    init(config: Config) {
        self.config = config
    }
    
    private func filePath(for key: RecordHandle) -> URL {
        let uuid = UUID.uuid5(namespace: config.namespace, name: key.handleName())
        let parts = uuid.uuidString.split(separator: "-").map(String.init)
        
        if #available(iOS 16.0, macOS 13.0, *) {
            let p = parts.reduce(config.cacheDirectory) { $0.appending(component: $1, directoryHint: .isDirectory) }
            return p.appending(component: "\(uuid).cache", directoryHint: .notDirectory)
        } else {
            let p = parts.reduce(config.cacheDirectory) { $0.appendingPathComponent($1, isDirectory: true) }
            return p.appendingPathComponent("\(uuid).cache", isDirectory: false)
        }
    }
    
    func getCacheEntry(_ key: RecordHandle) -> CacheEntry? {
        let fileURL = filePath(for: key)
        let decoder = JSONDecoder()
        //encoder.dateEncodingStrategy = .iso8601
        
        guard let d = try? Data(contentsOf: fileURL),
              //let c = try? NSKeyedUnarchiver.unarchiveTopLevelObjectWithData(d) as? CacheEntry
              let c = try? decoder.decode(CacheEntry.self, from: d)
        else {
            return nil
        }
        return c
    }
    
    func set(_ key: RecordHandle, to: Serializable?, options: DataStoreSetOptions? = nil) {
        do {
            let fileURL = filePath(for: key)
            let ttl = options?.ttl ?? config.ttl
            let entry = CacheEntry(entry: to, ttl: ttl)
            try FileManager.default.createDirectory(at: fileURL.deletingLastPathComponent(), withIntermediateDirectories: true)
            let encoder = JSONEncoder()
            encoder.dateEncodingStrategy = .iso8601
//            if let data = try? NSKeyedArchiver.archivedData(withRootObject: entry, requiringSecureCoding: config.secure) {
//                try? data.write(to: fileURL, options: .atomic)
//            }
            let data = try encoder.encode(entry)
            print("WriteTo \(fileURL.path)")
            try data.write(to: fileURL, options: .atomic)
        } catch {
            print("Swallow")
            // swallow
        }
    }
    
    func erase(_ key: RecordHandle) {
        let fileURL = filePath(for: key)
        try? FileManager.default.removeItem(at: fileURL)
    }

    
    func multiGet(_ keys: [RecordHandle]) -> [MultiGetResponse] {
        return keys.map {
            key in
            switch getCacheEntryItem(key) {
            case .empty:
                return .miss(key)
            case .raw(let d):
                guard let t = key.recordType() as? Serializable.Type,
                      let o = t.deserialize(data: d) else {
                    return .miss(key)
                }
                return .hit(key, o)
            case .entity(let o):
                return .hit(key, o)
            }
        }
    }
    
}

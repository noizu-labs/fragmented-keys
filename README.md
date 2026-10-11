# FragmentedKeys Library

## Overview

FragmentedKeys is a Swift library designed to streamline the management of cache invalidation by leveraging tag-value pair versioning. It's ideal for applications requiring sophisticated cache control mechanisms, such as web and mobile apps with dynamic content. The library supports multiple storage backends including `NSCache`, Disk, and SQL.

## Features

* **Tag-Based Versioning**: Efficiently manage cache invalidation using tag-instance versioning.
* **Flexible Storage Options**: Choose from various persistence stores like `NSCache`, Disk, or SQL through the `PersistenceManager`.
* **Advanced Tag Types**: Utilize a variety of tag types, including `StaticTag`, `LocaleTag`, `AccountTag`, `CacheTag`, and `IntervalTag` for diverse caching strategies.
* **Asynchronous Operations**: Fetch and populate cache entries asynchronously for optimal performance.
* **Thread-Safe Implementation**: Ensure safe use across multiple threads with built-in thread safety.

## Installation

### Swift Package Manager

To integrate FragmentedKeys into your Swift project using SPM, add the following to your `Package.swift` file:

```swift
dependencies: [
    .package(url: "https://github.com/yourusername/FragmentedKeys.git", .upToNextMajor(from: "1.0.0"))
]
```

Replace `https://github.com/yourusername/FragmentedKeys.git` with the actual URL of your library.

## Usage

### Setting Up

First, import FragmentedKeys in your Swift file.

```swift
import FragmentedKeys
```

### Initializing Tags

Create tags for cache management:

```swift
let userTag = CacheTag(name: "User", subject: userID)
let globalTag = StaticTag("GlobalSettings")
```

### Creating a KeyRing

Instantiate a `KeyRing` with the desired tags and a persistence manager:

```swift
let keyRing = FragmentedKeyManager.shared.declareKeyRing(withTags: [userTag, globalTag])
```

### Fetching and Storing Data

Use the `KeyRing` to fetch and store data in the cache:

```swift
let userData = keyRing.fetch(default: { self.loadUserData() }, persist: true)
```

### Invalidating Cache

Invalidate cache related to a specific tag:

```swift
userTag.incrementVersion()
```

## Advanced Usage

For more advanced scenarios, like using interval-based tags or subscribing to specific changes, refer to the detailed documentation provided in the library.

## Contributing

Contributions are welcome. Please read the contributing guide and code of conduct on our repository.

## License

FragmentedKeys is released under the [MIT License](https://opensource.org/licenses/MIT).

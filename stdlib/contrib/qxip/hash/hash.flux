// Package hash provides functions that convert string values to hashes.
//
// ## Metadata
// introduced: 0.192.0
// contributors: **GitHub**: [@lmangani](https://github.com/lmangani)
//
package hash


// sha256 converts a string value to a hexadecimal hash using the SHA 256 hash algorithm.
//
// ## Parameters
//
// - v: String to hash.
//
// ## Examples
// ### Convert a string to a SHA 256 hash
// ```no_run
// import "contrib/qxip/hash"
//
// hash.sha256(v: "Hello, world!")
//
// // Returns 315f5bdb76d078c43b8ac0064e4a0164612b1fce77c869345bfc94c75894edd3
// ```
//
// ## Metadata
// tag: type-conversion
builtin sha256 : (v: A) => string

// sha1 converts a string value to a hexadecimal hash using the SHA-1 hash algorithm.
//
// ## Parameters
//
// - v: String to hash.
//
// ## Examples
// ### Convert a string to a SHA-1 hash
// ```no_run
// import "contrib/qxip/hash"
//
// hash.sha1(v: "Hello, world!")
//
// // Returns 315f5bdb76d078c43b8ac0064e4a0164612b1fce77c869345bfc94c75894edd3
// ```
//
// ## Metadata
// tag: type-conversion
builtin sha1 : (v: A) => string

// xxhash64 converts a string value to a 64-bit hexadecimal hash using the xxHash algorithm.
//
// ## Parameters
//
// - v: String to hash.
//
// ## Examples
// ### Convert a string to 64-bit hash using xxHash
// ```no_run
// import "contrib/qxip/hash"
//
// hash.xxhash64(v: "Hello, world!")
//
// // Returns 17691043854468224118
// ```
//
// ## Metadata
// tag: type-conversion
builtin xxhash64 : (v: A) => string

// cityhash64 converts a string value to a 64-bit hexadecimal hash using the CityHash64 algorithm.
//
// ## Parameters
//
// - v: String to hash.
//
// ## Examples
// ### Convert a string to a 64-bit hash using CityHash64
// ```no_run
// import "contrib/qxip/hash"
//
// hash.cityhash64(v: "Hello, world!")
//
// // Returns 2359500134450972198
// ```
//
// ## Metadata
// tag: type-conversion
builtin cityhash64 : (v: A) => string

// b64 converts a string value to a Base64 string.
//
// ## Parameters
//
// - v: String to hash.
//
// ## Examples
// ### Convert a string to a Base64 string
// ```no_run
// import "contrib/qxip/hash"
//
// hash.b64(v: "Hello, world!")
//
// // Returns 2359500134450972198
// ```
//
// ## Metadata
// tag: type-conversion
builtin b64 : (v: A) => string

// md5 converts a string value to an MD5 hash.
//
// ## Parameters
//
// - v: String to hash.
//
// ## Examples
// ### Convert a string to an MD5 hash
// ```no_run
// import "contrib/qxip/hash"
//
// hash.md5(v: "Hello, world!")
//
// // Returns 2359500134450972198
// ```
//
// ## Metadata
// tag: type-conversion
builtin md5 : (v: A) => string

// hmac converts a string value to an MD5-signed SHA-1 hash.
//
// ## Parameters
//
// - v: String to hash.
// - k: Key to sign hash.
//
// ## Examples
// ### Convert a string and key to a base64-signed hash
// ```no_run
// import "contrib/qxip/hash"
//
// hash.hmac(v: "helloworld", k: "123456")
//
// // Returns 75B5ueLnnGepYvh+KoevTzXCrjc=
// ```
//
// ## Metadata
// tag: type-conversion
builtin hmac : (v: A, k: A) => string

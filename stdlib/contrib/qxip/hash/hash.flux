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

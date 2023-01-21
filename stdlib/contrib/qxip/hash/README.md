# Hash Package

The Flux Hash Package provides functions that perform hash conversion of `string` values.

## hash.sha256
The `hash.sha256()` function converts a single string to a hash using sha256.

Example:

```
    import "contrib/qxip/hash"

    a = hash.sha256(v: "Hello, world!")
```

## hash.sha1
The `hash.sha256()` function converts a single string to a hash using sha256.

Example:

```
    import "contrib/qxip/hash"

    a = hash.sha1(v: "Hello, world!")
```

## hash.xxhash64
The `hash.xxhash64()` function converts a single string to a hash using xxhash64.

Example:

```
    import "contrib/qxip/hash"

    a = hash.xxhash64(v: "Hello, world!")
```

## hash.cityhash64
The `hash.cityhash64()` function converts a single string to hash using cityhash64.

Example:

```
    import "contrib/qxip/hash"

    a = hash.cityhash64(v: "Hello, world!")
```

## hash.md5
The `hash.md5()` function converts a single string to hash using MD5.

Example:

```
    import "contrib/qxip/hash"

    a = hash.md5(v: "Hello, world!")
```

## hash.b64
The `hash.b64()` function converts a single string to a Base64 string.

Example:

```
    import "contrib/qxip/hash"

    a = hash.b64(v: "Hello, world!")
```

## Contact
- Author: Lorenzo Mangani
- Email: lorenzo.mangani@gmail.com
- Github: [@lmangani](https://github.com/lmangani)
- Website: [@qryn](https://qryn.dev)

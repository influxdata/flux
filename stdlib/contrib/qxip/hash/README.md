# Hash Package

The Flux Hash Package provides functions that perform hash conversion of `string` values.

## hash.sha256
The `hash.sha256()` function converts a single string to a hash using sha256.

Example:

```
    import "contrib/qxip/hash"

    a = hash.sha256("Hello, world!")
    // a is "315f5bdb76d078c43b8ac0064e4a0164612b1fce77c869345bfc94c75894edd3"
```
## hash.xxhash64
The `hash.xxhash64()` function converts a single string to a hash using xxhash64.

Example:

```
    import "contrib/qxip/hash"

    a = hash.xxhash64("Hello, world!")
    // a is "17691043854468224118"
```
## hash.cityhash64
The `hash.cityhash64()` function converts a single string to hash using cityhash64.

Example:

```
    import "contrib/qxip/hash"

    a = hash.cityhash64("Hello, world!")
    // a is "2359500134450972198"
```

## Contact
- Author: Lorenzo Mangani
- Email: lorenzo.mangani@gmail.com
- Github: [@lmangani](https://github.com/lmangani)
- Website: [@qryn](https://qryn.dev)

# Hex Package

The Flux Hex Package provides functions that perform hexadecimal conversion of `int`, `uint` or `bytes` values to/from `string` values.

## hex.string
The `hex.string()` function converts a single value to a string. It is like a standard [string() function](https://docs.influxdata.com/influxdb/v2.0/reference/flux/stdlib/built-in/transformations/type-conversions/string/), but it encodes `int`, `uint` and `bytes` to hexadecimal lowercase characters.

Example:

    import "contrib/bonitoo-io/hex"

    a = hex.string(v: 12)
    // a is "c"
    b = hex.string(v: bytes(v: "hi"))
    // b is "6869"


## hex.int

The `hex.int()` function converts a hexadecimal string representation of a number value to an integer.
An input value can be optionally prefixed by `0x`. 

Example:

    import "contrib/bonitoo-io/hex"

    a = hex.int(v: "c")
    // a is 12
    b = hex.int(v: "-d")
    // b is -13
    c = hex.int(v: "0xe")
    // c is 14
    c = hex.int(v: "-0xF")
    // d is -15



## hex.uint

The `hex.uint()` function converts a hexadecimal string representation of a number to an unsigned integer.
An input value can be optionally prefixed by `0x`. 

Example:

    import "contrib/bonitoo-io/hex"

    a = hex.uint(v: "C")
    // a is uint(12)
    b = hex.int(v: "0xd")
    // b is uint(13)


## hex.bytes

The `hex.bytes()` function decodes a string of hexadecimal characters into a flux bytes value. 

Example:

    import "contrib/bonitoo-io/hex"

    a = hex.bytes(v: "6869")
    // a is bytes("hi")

## Contact

- Author: Pavel Zavora
- Email: pavel.zavora@bonitoo.io
- Github: [@sranka](https://github.com/sranka)
- Influx Slack: [@sranka](https://influxdata.com/slack)

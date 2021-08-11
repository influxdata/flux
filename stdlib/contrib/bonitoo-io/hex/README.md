# Hex Package

The Flux Hex Package provides functions that perform hexadecimal conversion of `int`, `uint` or `bytes` values to/from `string` values.

## hex.string
The `hex.string()` function converts a single value to a string. It is like a standard [string() function](https://docs.influxdata.com/influxdb/v2.0/reference/flux/stdlib/built-in/transformations/type-conversions/string/), but it encodes `int`, `uint` and `bytes` to hexadecimal lowercase characters.

Example:

    import "contrib/bonitoo-io/hex"

    a = hex.string(v: 12)
    // a is "c"
    b = hex.string(v: bytes("hi"))
    // b is "6869"


## hex.int

The `hex.int()` function converts a single value to an integer. It is like a standard [int() function](https://docs.influxdata.com/influxdb/v2.0/reference/flux/stdlib/built-in/transformations/type-conversions/int/), but it assumes that a string argument is a hexadecimal representation of an integer, which can be optionally prefixed by `0x`. 

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

The `hex.uint()` function converts a single value to an unsigned integer. It is like a standard [uint() function](https://docs.influxdata.com/influxdb/v2.0/reference/flux/stdlib/built-in/transformations/type-conversions/uint/), but it assumes that a string argument is a hexadecimal representation of an number, which can be optionally prefixed by `0x`. 

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

## hex.toString

The `hex.toString()` function converts all values in the `_value` column to strings using `hex.string()` function. It is like a standard [toString() function](https://docs.influxdata.com/influxdb/v2.0/reference/flux/stdlib/built-in/transformations/type-conversions/toString/), but it encodes `int`, `uint` and `bytes` to hexadecimal lowercase characters.

Example:

    import "contrib/bonitoo-io/hex"

    from(bucket: "example")
      |> filter(fn:(r) =>
        r._field == "txId"
      )
      |> hex.toString()

## hex.toInt

The `hex.toInt()` function converts all values in the `_value` column to integers using `hex.int()` function. It is like a standard [toInt() function](https://docs.influxdata.com/influxdb/v2.0/reference/flux/stdlib/built-in/transformations/type-conversions/toInt/), but string input values are parsed from hexadecimal characters. 

Example:

    import "contrib/bonitoo-io/hex"

    from(bucket: "example")
      |> filter(fn:(r) =>
        r._field == "corIdStr"
      )
      |> hex.toInt()

## hex.toUInt

The `hex.toUInt()` function converts all values in the `_value` column to unsigned integers using `hex.uint()` function. It is like a standard [toUInt() function](https://docs.influxdata.com/influxdb/v2.0/reference/flux/stdlib/built-in/transformations/type-conversions/toUInt/), but string input values are parsed from hexadecimal characters. 

Example:

    import "contrib/bonitoo-io/hex"

    from(bucket: "example")
      |> filter(fn:(r) =>
        r._field == "corIdStr"
      )
      |> hex.toUInt()

## Contact

- Author: Pavel Zavora
- Email: pavel.zavora@bonitoo.io
- Github: [@sranka](https://github.com/sranka)
- Influx Slack: [@sranka](https://influxdata.com/slack)

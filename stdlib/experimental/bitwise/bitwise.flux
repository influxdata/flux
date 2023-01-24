// Package bitwise provides functions for performing bitwise operations on integers.
//
// **Deprecated**: This package is deprecated in favor of [`bitwise`](https://docs.influxdata.com/flux/v0.x/stdlib/bitwise/).
//
// All integers are 64 bit integers.
//
// Functions prefixed with s operate on signed integers (int).
// Functions prefixed with u operate on unsigned integers (uint).
//
// ## Metadata
// introduced: 0.138.0
// deprecated: 0.173.0
// tags: bitwise
//
package bitwise


import "bitwise"

// uand performs the bitwise operation, `a AND b`, with unsigned integers.
//
// **Deprecated**: Experimental `bitwise.uand` is deprecated in favor of
// [`bitwise.uand`](https://docs.influxdata.com/flux/v0.x/stdlib/bitwise/uand/).
//
// ## Parameters
// - a: Left hand operand.
// - b: Right hand operand.
//
// ## Examples
// ### Perform a bitwise AND operation
// ```no_run
// import "experimental/bitwise"
//
// bitwise.uand(a: uint(v: 1234), b: uint(v: 4567))
//
// // Returns 210 (uint)
// ```
//
// ### Perform a bitwise AND operation on a stream of tables
// ```
// import "experimental/bitwise"
// import "sampledata"
//
// < sampledata.uint()
// >    |> map(fn: (r) => ({ r with _value: bitwise.uand(a: r._value, b: uint(v: 3))}))
// ```
//
uand = bitwise.uand

// uor performs the bitwise operation, `a OR b`, with unsigned integers.
//
// **Deprecated**: Experimental `bitwise.uor` is deprecated in favor of
// [`bitwise.uor`](https://docs.influxdata.com/flux/v0.x/stdlib/bitwise/uor/).
//
// ## Parameters
// - a: Left hand operand.
// - b: Right hand operand.
//
// ## Examples
// ### Perform a bitwise OR operation
// ```no_run
// import "experimental/bitwise"
//
// bitwise.uor(a: uint(v: 1234), b: uint(v: 4567))
//
// // Returns 5591 (uint)
// ```
//
// ### Perform a bitwise OR operation on a stream of tables
// ```
// import "experimental/bitwise"
// import "sampledata"
//
// < sampledata.uint()
// >    |> map(fn: (r) => ({ r with _value: bitwise.uor(a: r._value, b: uint(v: 3))}))
// ```
//
uor = bitwise.uor

// unot inverts every bit in `a`, an unsigned integer.
//
// **Deprecated**: Experimental `bitwise.unot` is deprecated in favor of
// [`bitwise.unot`](https://docs.influxdata.com/flux/v0.x/stdlib/bitwise/unot/).
//
// ## Parameters
// - a: Unsigned integer to invert.
//
// ## Examples
// ### Invert bits in an unsigned integer
// ```no_run
// import "experimental/bitwise"
//
// bitwise.unot(a: uint(v: 1234))
//
// // Returns 18446744073709550381 (uint)
// ```
//
// ### Invert bits in unsigned integers in a stream of tables
// ```
// import "experimental/bitwise"
// import "sampledata"
//
// < sampledata.uint()
// >    |> map(fn: (r) => ({ r with _value: bitwise.unot(a: r._value)}))
// ```
//
unot = bitwise.unot

// uxor performs the bitwise operation, `a XOR b`, with unsigned integers.
//
// **Deprecated**: Experimental `bitwise.uxor` is deprecated in favor of
// [`bitwise.uxor`](https://docs.influxdata.com/flux/v0.x/stdlib/bitwise/uxor/).
//
// ## Parameters
// - a: Left hand operand.
// - b: Right hand operand.
//
// ## Examples
// ### Perform a bitwise XOR operation
// ```no_run
// import "experimental/bitwise"
//
// bitwise.uxor(a: uint(v: 1234), b: uint(v: 4567))
//
// // Returns 5381 (uint)
// ```
//
// ### Perform a bitwise XOR operation on a stream of tables
// ```
// import "experimental/bitwise"
// import "sampledata"
//
// < sampledata.uint()
// >    |> map(fn: (r) => ({ r with _value: bitwise.uxor(a: r._value, b: uint(v: 3))}))
// ```
//
uxor = bitwise.uxor

// uclear performs the bitwise operation `a AND NOT b`, with unsigned integers.
//
// **Deprecated**: Experimental `bitwise.uclear` is deprecated in favor of
// [`bitwise.uclear`](https://docs.influxdata.com/flux/v0.x/stdlib/bitwise/uclear/).
//
// ## Parameters
// - a: Left hand operand.
// - b: Bits to clear.
//
// ## Examples
// ### Perform a bitwise AND NOT operation
// ```no_run
// import "experimental/bitwise"
//
// bitwise.uclear(a: uint(v: 1234), b: uint(v: 4567))
//
// // Returns 1024 (uint)
// ```
//
// ### Perform a bitwise AND NOT operation on a stream of tables
// ```
// import "experimental/bitwise"
// import "sampledata"
//
// < sampledata.uint()
// >    |> map(fn: (r) => ({ r with _value: bitwise.uclear(a: r._value, b: uint(v: 3))}))
// ```
//
uclear = bitwise.uclear

// ulshift shifts the bits in `a` left by `b` bits.
// Both `a` and `b` are unsigned integers.
//
// **Deprecated**: Experimental `bitwise.ulshift` is deprecated in favor of
// [`bitwise.ulshift`](https://docs.influxdata.com/flux/v0.x/stdlib/bitwise/ulshift/).
//
// ## Parameters
// - a: Left hand operand.
// - b: Number of bits to shift.
//
// ## Examples
// ### Shift bits left in an unsigned integer
// ```no_run
// import "experimental/bitwise"
//
// bitwise.ulshift(a: uint(v: 1234), b: uint(v: 2))
//
// // Returns 4936 (uint)
// ```
//
// ### Shift bits left in unsigned integers in a stream of tables
// ```
// import "experimental/bitwise"
// import "sampledata"
//
// < sampledata.uint()
// >    |> map(fn: (r) => ({ r with _value: bitwise.ulshift(a: r._value, b: uint(v: 3))}))
// ```
//
ulshift = bitwise.ulshift

// urshift shifts the bits in `a` right by `b` bits.
// Both `a` and `b` are unsigned integers.
//
// **Deprecated**: Experimental `bitwise.urshift` is deprecated in favor of
// [`bitwise.urshift`](https://docs.influxdata.com/flux/v0.x/stdlib/bitwise/urshift/).
//
// ## Parameters
// - a: Left hand operand.
// - b: Number of bits to shift.
//
// ## Examples
// ### Shift bits right in an unsigned integer
// ```no_run
// import "experimental/bitwise"
//
// bitwise.urshift(a: uint(v: 1234), b: uint(v: 2))
//
// // Returns 308 (uint)
// ```
//
// ### Shift bits right in unsigned integers in a stream of tables
// ```
// import "experimental/bitwise"
// import "sampledata"
//
// < sampledata.uint()
// >    |> map(fn: (r) => ({ r with _value: bitwise.urshift(a: r._value, b: uint(v: 3))}))
// ```
//
urshift = bitwise.urshift

// sand performs the bitwise operation, `a AND b`, with integers.
//
// **Deprecated**: Experimental `bitwise.sand` is deprecated in favor of
// [`bitwise.sand`](https://docs.influxdata.com/flux/v0.x/stdlib/bitwise/sand/).
//
// ## Parameters
// - a: Left hand operand.
// - b: Right hand operand.
//
// ## Examples
// ### Perform a bitwise AND operation
// ```no_run
// import "experimental/bitwise"
//
// bitwise.sand(a: 1234, b: 4567)
//
// // Returns 210
// ```
//
// ### Perform a bitwise AND operation on a stream of tables
// ```
// import "experimental/bitwise"
// import "sampledata"
//
// < sampledata.int()
// >    |> map(fn: (r) => ({ r with _value: bitwise.sand(a: r._value, b: 3)}))
// ```
//
sand = bitwise.sand

// sor performs the bitwise operation, `a OR b`, with integers.
//
// **Deprecated**: Experimental `bitwise.sor` is deprecated in favor of
// [`bitwise.sor`](https://docs.influxdata.com/flux/v0.x/stdlib/bitwise/sor/).
//
// ## Parameters
// - a: Left hand operand.
// - b: Right hand operand.
//
// ## Examples
// ### Perform a bitwise OR operation
// ```no_run
// import "experimental/bitwise"
//
// bitwise.sor(a: 1234, b: 4567)
//
// // Returns 5591
// ```
//
// ### Perform a bitwise OR operation on a stream of tables
// ```
// import "experimental/bitwise"
// import "sampledata"
//
// < sampledata.int()
// >    |> map(fn: (r) => ({ r with _value: bitwise.sor(a: r._value, b: 3)}))
// ```
//
sor = bitwise.sor

// snot inverts every bit in `a`, an integer.
//
// **Deprecated**: Experimental `bitwise.snot` is deprecated in favor of
// [`bitwise.snot`](https://docs.influxdata.com/flux/v0.x/stdlib/bitwise/snot/).
//
// ## Parameters
// - a: Integer to invert.
//
// ## Examples
// ### Invert bits in an integer
// ```no_run
// import "experimental/bitwise"
//
// bitwise.snot(a: 1234)
//
// // Returns -1235
// ```
//
// ### Invert bits in integers in a stream of tables
// ```
// import "experimental/bitwise"
// import "sampledata"
//
// < sampledata.int()
// >    |> map(fn: (r) => ({ r with _value: bitwise.snot(a: r._value)}))
// ```
//
snot = bitwise.snot

// sxor performs the bitwise operation, `a XOR b`, with integers.
//
// **Deprecated**: Experimental `bitwise.sxor` is deprecated in favor of
// [`bitwise.sxor`](https://docs.influxdata.com/flux/v0.x/stdlib/bitwise/sxor/).
//
// ## Parameters
// - a: Left hand operand.
// - b: Right hand operand.
//
// ## Examples
// ### Perform a bitwise XOR operation
// ```no_run
// import "experimental/bitwise"
//
// bitwise.sxor(a: 1234, b: 4567)
//
// // Returns 5381
// ```
//
// ### Perform a bitwise XOR operation on a stream of tables
// ```
// import "experimental/bitwise"
// import "sampledata"
//
// < sampledata.int()
// >    |> map(fn: (r) => ({ r with _value: bitwise.sxor(a: r._value, b: 3)}))
// ```
//
sxor = bitwise.sxor

// sclear performs the bitwise operation `a AND NOT b`.
// Both `a` and `b` are integers.
//
// **Deprecated**: Experimental `bitwise.sclear` is deprecated in favor of
// [`bitwise.sclear`](https://docs.influxdata.com/flux/v0.x/stdlib/bitwise/sclear/).
//
// ## Parameters
// - a: Left hand operand.
// - b: Bits to clear.
//
// ## Examples
// ### Perform a bitwise AND NOT operation
// ```no_run
// import "experimental/bitwise"
//
// bitwise.sclear(a: 1234, b: 4567)
//
// // Returns 1024
// ```
//
// ### Perform a bitwise AND NOT operation on a stream of tables
// ```
// import "experimental/bitwise"
// import "sampledata"
//
// < sampledata.int()
// >    |> map(fn: (r) => ({ r with _value: bitwise.sclear(a: r._value, b: 3)}))
// ```
//
sclear = bitwise.sclear

// slshift shifts the bits in `a` left by `b` bits.
// Both `a` and `b` are integers.
//
// **Deprecated**: Experimental `bitwise.slshift` is deprecated in favor of
// [`bitwise.slshift`](https://docs.influxdata.com/flux/v0.x/stdlib/bitwise/slshift/).
//
// ## Parameters
// - a: Left hand operand.
// - b: Number of bits to shift.
//
// ## Examples
// ### Shift bits left in an integer
// ```no_run
// import "experimental/bitwise"
//
// bitwise.slshift(a: 1234, b: 2)
//
// // Returns 4936
// ```
//
// ### Shift bits left in integers in a stream of tables
// ```
// import "experimental/bitwise"
// import "sampledata"
//
// < sampledata.int()
// >    |> map(fn: (r) => ({ r with _value: bitwise.slshift(a: r._value, b: 3)}))
// ```
//
slshift = bitwise.slshift

// srshift shifts the bits in `a` right by `b` bits.
// Both `a` and `b` are integers.
//
// **Deprecated**: Experimental `bitwise.srshift` is deprecated in favor of
// [`bitwise.srshift`](https://docs.influxdata.com/flux/v0.x/stdlib/bitwise/srshift/).
//
// ## Parameters
// - a: Left hand operand.
// - b: Number of bits to shift.
//
// ## Examples
// ### Shift bits right in an integer
// ```no_run
// import "experimental/bitwise"
//
// bitwise.srshift(a: 1234, b: 2)
//
// // Returns 308
// ```
//
// ### Shift bits right in integers in a stream of tables
// ```
// import "experimental/bitwise"
// import "sampledata"
//
// < sampledata.int()
// >    |> map(fn: (r) => ({ r with _value: bitwise.srshift(a: r._value, b: 3)}))
// ```
//
srshift = bitwise.srshift

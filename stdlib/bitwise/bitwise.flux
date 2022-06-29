// Package bitwise provides functions for performing bitwise operations on integers.
//
// All integers are 64 bit integers.
//
// Functions prefixed with s operate on signed integers (int).
// Functions prefixed with u operate on unsigned integers (uint).
//
// ## Metadata
// introduced: 0.173.0
// tags: bitwise
//
package bitwise


// uand performs the bitwise operation, `a AND b`, with unsigned integers.
//
// ## Parameters
// - a: Left hand operand.
// - b: Right hand operand.
//
// ## Examples
// ### Perform a bitwise AND operation
// ```no_run
// import "bitwise"
//
// bitwise.uand(a: uint(v: 1234), b: uint(v: 4567))
//
// // Returns 210 (uint)
// ```
//
// ### Perform a bitwise AND operation on a stream of tables
// ```
// import "bitwise"
// import "sampledata"
//
// < sampledata.uint()
// >    |> map(fn: (r) => ({ r with _value: bitwise.uand(a: r._value, b: uint(v: 3))}))
// ```
//
builtin uand : (a: uint, b: uint) => uint

// uor performs the bitwise operation, `a OR b`, with unsigned integers.
//
// ## Parameters
// - a: Left hand operand.
// - b: Right hand operand.
//
// ## Examples
// ### Perform a bitwise OR operation
// ```no_run
// import "bitwise"
//
// bitwise.uor(a: uint(v: 1234), b: uint(v: 4567))
//
// // Returns 5591 (uint)
// ```
//
// ### Perform a bitwise OR operation on a stream of tables
// ```
// import "bitwise"
// import "sampledata"
//
// < sampledata.uint()
// >    |> map(fn: (r) => ({ r with _value: bitwise.uor(a: r._value, b: uint(v: 3))}))
// ```
//
builtin uor : (a: uint, b: uint) => uint

// unot inverts every bit in `a`, an unsigned integer.
//
// ## Parameters
// - a: Unsigned integer to invert.
//
// ## Examples
// ### Invert bits in an unsigned integer
// ```no_run
// import "bitwise"
//
// bitwise.unot(a: uint(v: 1234))
//
// // Returns 18446744073709550381 (uint)
// ```
//
// ### Invert bits in unsigned integers in a stream of tables
// ```
// import "bitwise"
// import "sampledata"
//
// < sampledata.uint()
// >    |> map(fn: (r) => ({ r with _value: bitwise.unot(a: r._value)}))
// ```
//
builtin unot : (a: uint) => uint

// uxor performs the bitwise operation, `a XOR b`, with unsigned integers.
//
// ## Parameters
// - a: Left hand operand.
// - b: Right hand operand.
//
// ## Examples
// ### Perform a bitwise XOR operation
// ```no_run
// import "bitwise"
//
// bitwise.uxor(a: uint(v: 1234), b: uint(v: 4567))
//
// // Returns 5381 (uint)
// ```
//
// ### Perform a bitwise XOR operation on a stream of tables
// ```
// import "bitwise"
// import "sampledata"
//
// < sampledata.uint()
// >    |> map(fn: (r) => ({ r with _value: bitwise.uxor(a: r._value, b: uint(v: 3))}))
// ```
//
builtin uxor : (a: uint, b: uint) => uint

// uclear performs the bitwise operation `a AND NOT b`, with unsigned integers.
//
// ## Parameters
// - a: Left hand operand.
// - b: Bits to clear.
//
// ## Examples
// ### Perform a bitwise AND NOT operation
// ```no_run
// import "bitwise"
//
// bitwise.uclear(a: uint(v: 1234), b: uint(v: 4567))
//
// // Returns 1024 (uint)
// ```
//
// ### Perform a bitwise AND NOT operation on a stream of tables
// ```
// import "bitwise"
// import "sampledata"
//
// < sampledata.uint()
// >    |> map(fn: (r) => ({ r with _value: bitwise.uclear(a: r._value, b: uint(v: 3))}))
// ```
//
builtin uclear : (a: uint, b: uint) => uint

// ulshift shifts the bits in `a` left by `b` bits.
// Both `a` and `b` are unsigned integers.
//
// ## Parameters
// - a: Left hand operand.
// - b: Number of bits to shift.
//
// ## Examples
// ### Shift bits left in an unsigned integer
// ```no_run
// import "bitwise"
//
// bitwise.ulshift(a: uint(v: 1234), b: uint(v: 2))
//
// // Returns 4936 (uint)
// ```
//
// ### Shift bits left in unsigned integers in a stream of tables
// ```
// import "bitwise"
// import "sampledata"
//
// < sampledata.uint()
// >    |> map(fn: (r) => ({ r with _value: bitwise.ulshift(a: r._value, b: uint(v: 3))}))
// ```
//
builtin ulshift : (a: uint, b: uint) => uint

// urshift shifts the bits in `a` right by `b` bits.
// Both `a` and `b` are unsigned integers.
//
// ## Parameters
// - a: Left hand operand.
// - b: Number of bits to shift.
//
// ## Examples
// ### Shift bits right in an unsigned integer
// ```no_run
// import "bitwise"
//
// bitwise.urshift(a: uint(v: 1234), b: uint(v: 2))
//
// // Returns 308 (uint)
// ```
//
// ### Shift bits right in unsigned integers in a stream of tables
// ```
// import "bitwise"
// import "sampledata"
//
// < sampledata.uint()
// >    |> map(fn: (r) => ({ r with _value: bitwise.urshift(a: r._value, b: uint(v: 3))}))
// ```
//
builtin urshift : (a: uint, b: uint) => uint

// sand performs the bitwise operation, `a AND b`, with integers.
//
// ## Parameters
// - a: Left hand operand.
// - b: Right hand operand.
//
// ## Examples
// ### Perform a bitwise AND operation
// ```no_run
// import "bitwise"
//
// bitwise.sand(a: 1234, b: 4567)
//
// // Returns 210
// ```
//
// ### Perform a bitwise AND operation on a stream of tables
// ```
// import "bitwise"
// import "sampledata"
//
// < sampledata.int()
// >    |> map(fn: (r) => ({ r with _value: bitwise.sand(a: r._value, b: 3)}))
// ```
//
builtin sand : (a: int, b: int) => int

// sor performs the bitwise operation, `a OR b`, with integers.
//
// ## Parameters
// - a: Left hand operand.
// - b: Right hand operand.
//
// ## Examples
// ### Perform a bitwise OR operation
// ```no_run
// import "bitwise"
//
// bitwise.sor(a: 1234, b: 4567)
//
// // Returns 5591
// ```
//
// ### Perform a bitwise OR operation on a stream of tables
// ```
// import "bitwise"
// import "sampledata"
//
// < sampledata.int()
// >    |> map(fn: (r) => ({ r with _value: bitwise.sor(a: r._value, b: 3)}))
// ```
//
builtin sor : (a: int, b: int) => int

// snot inverts every bit in `a`, an integer.
//
// ## Parameters
// - a: Integer to invert.
//
// ## Examples
// ### Invert bits in an integer
// ```no_run
// import "bitwise"
//
// bitwise.snot(a: 1234)
//
// // Returns -1235
// ```
//
// ### Invert bits in integers in a stream of tables
// ```
// import "bitwise"
// import "sampledata"
//
// < sampledata.int()
// >    |> map(fn: (r) => ({ r with _value: bitwise.snot(a: r._value)}))
// ```
//
builtin snot : (a: int) => int

// sxor performs the bitwise operation, `a XOR b`, with integers.
//
// ## Parameters
// - a: Left hand operand.
// - b: Right hand operand.
//
// ## Examples
// ### Perform a bitwise XOR operation
// ```no_run
// import "bitwise"
//
// bitwise.sxor(a: 1234, b: 4567)
//
// // Returns 5381
// ```
//
// ### Perform a bitwise XOR operation on a stream of tables
// ```
// import "bitwise"
// import "sampledata"
//
// < sampledata.int()
// >    |> map(fn: (r) => ({ r with _value: bitwise.sxor(a: r._value, b: 3)}))
// ```
//
builtin sxor : (a: int, b: int) => int

// sclear performs the bitwise operation `a AND NOT b`.
// Both `a` and `b` are integers.
//
// ## Parameters
// - a: Left hand operand.
// - b: Bits to clear.
//
// ## Examples
// ### Perform a bitwise AND NOT operation
// ```no_run
// import "bitwise"
//
// bitwise.sclear(a: 1234, b: 4567)
//
// // Returns 1024
// ```
//
// ### Perform a bitwise AND NOT operation on a stream of tables
// ```
// import "bitwise"
// import "sampledata"
//
// < sampledata.int()
// >    |> map(fn: (r) => ({ r with _value: bitwise.sclear(a: r._value, b: 3)}))
// ```
//
builtin sclear : (a: int, b: int) => int

// slshift shifts the bits in `a` left by `b` bits.
// Both `a` and `b` are integers.
//
// ## Parameters
// - a: Left hand operand.
// - b: Number of bits to shift.
//
// ## Examples
// ### Shift bits left in an integer
// ```no_run
// import "bitwise"
//
// bitwise.slshift(a: 1234, b: 2)
//
// // Returns 4936
// ```
//
// ### Shift bits left in integers in a stream of tables
// ```
// import "bitwise"
// import "sampledata"
//
// < sampledata.int()
// >    |> map(fn: (r) => ({ r with _value: bitwise.slshift(a: r._value, b: 3)}))
// ```
//
builtin slshift : (a: int, b: int) => int

// srshift shifts the bits in `a` right by `b` bits.
// Both `a` and `b` are integers.
//
// ## Parameters
// - a: Left hand operand.
// - b: Number of bits to shift.
//
// ## Examples
// ### Shift bits right in an integer
// ```no_run
// import "bitwise"
//
// bitwise.srshift(a: 1234, b: 2)
//
// // Returns 308
// ```
//
// ### Shift bits right in integers in a stream of tables
// ```
// import "bitwise"
// import "sampledata"
//
// < sampledata.int()
// >    |> map(fn: (r) => ({ r with _value: bitwise.srshift(a: r._value, b: 3)}))
// ```
//
builtin srshift : (a: int, b: int) => int

// Package bitwise provides functions for performing bitwise operations on integers.
//
// All integers are 64 bit integers.
//
// Functions prefixed with 'u' operate on unsigned integers while functions
// prefixed with 's' operate on signed integers.
package bitwise


// uand performs the bitwise operation a AND b.
//
// ## Parameters
// - a: is the left hand operand
// - b: is the right hand operand
//
builtin uand : (a: uint, b: uint) => uint

// uor performs the bitwise operation a OR b.
//
// ## Parameters
// - a: is the left hand operand
// - b: is the right hand operand
//
builtin uor : (a: uint, b: uint) => uint

// unot inverts every bit in a.
//
// ## Parameters
// - a: is the integer to invert.
//
builtin unot : (a: uint) => uint

// uxor performs the bitwise operation a XOR b.
//
// ## Parameters
// - a: is the left hand operand
// - b: is the right hand operand
//
builtin uxor : (a: uint, b: uint) => uint

// uclear performs the bitwise operation a AND NOT b.
//
// ## Parameters
// - a: is the left hand operand
// - b: indicates which bits of a will be cleared.
//
builtin uclear : (a: uint, b: uint) => uint

// ulshift shifts the bits in a left by b bits.
//
// ## Parameters
// - a: is the left hand operand
// - b: the number of bits to shift
//
builtin ulshift : (a: uint, b: uint) => uint

// urshift shifts the bits in a right by b bits.
//
// ## Parameters
// - a: is the left hand operand
// - b: the number of bits to shift
//
builtin urshift : (a: uint, b: uint) => uint

// sand performs the bitwise operation a AND b.
//
// ## Parameters
// - a: is the left hand operand
// - b: is the right hand operand
//
builtin sand : (a: int, b: int) => int

// sor performs the bitwise operation a OR b.
//
// ## Parameters
// - a: is the left hand operand
// - b: is the right hand operand
//
builtin sor : (a: int, b: int) => int

// snot inverts every bit in a.
//
// ## Parameters
// - a: is the integer to invert.
//
builtin snot : (a: int) => int

// sxor performs the bitwise operation a XOR b.
//
// ## Parameters
// - a: is the left hand operand
// - b: is the right hand operand
//
builtin sxor : (a: int, b: int) => int

// sclear performs the bitwise operation a AND NOT b.
//
// ## Parameters
// - a: is the left hand operand
// - b: indicates which bits of a will be cleared.
//
builtin sclear : (a: int, b: int) => int

// slshift shifts the bits in a left by b bits.
//
// ## Parameters
// - a: is the left hand operand
// - b: the number of bits to shift
//
builtin slshift : (a: int, b: int) => int

// srshift shifts the bits in a right by b bits.
//
// ## Parameters
// - a: is the left hand operand
// - b: the number of bits to shift
//
builtin srshift : (a: int, b: int) => int

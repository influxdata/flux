// Package math provides basic constants and mathematical functions.
//
// ## Metadata
// introduced: 0.22.0
package math


// pi represents pi (π).
builtin pi : float

// e represents the base of the natural logarithm, also known as Euler's number.
builtin e : float

// phi represents the [Golden Ratio](https://www.britannica.com/science/golden-ratio).
builtin phi : float

// sqrt2 represents the square root of 2.
builtin sqrt2 : float

// sqrte represents the square root of **e** (`math.e`).
builtin sqrte : float

// sqrtpi represents the square root of pi (π).
builtin sqrtpi : float

// sqrtphi represents the square root of phi (`math.phi`), the Golden Ratio.
builtin sqrtphi : float

// ln2 represents the natural logarithm of 2.
builtin ln2 : float

// log2e represents the base 2 logarithm of **e** (`math.e`).
builtin log2e : float

// ln10 represents the natural logarithm of 10.
builtin ln10 : float

// log10e represents the base 10 logarithm of **e** (`math.e`).
builtin log10e : float

// maxfloat represents the maximum float value.
builtin maxfloat : float

// smallestNonzeroFloat represents the smallest nonzero float value.
builtin smallestNonzeroFloat : float

// maxint represents the maximum integer value (`2^63 - 1`).
builtin maxint : int

// minint represents the minimum integer value (`-2^63`).
builtin minint : int

// maxuint representes the maximum unsigned integer value  (`2^64 - 1`).
builtin maxuint : uint

// abs returns the absolute value of `x`.
//
// ## Parameters
// - x: Value to operate on.
//
// ## Examples
//
// ### Return the absolute value
// ```no_run
// # import "math"
//
// math.abs(x: -1.22) // 1.22
// ```
//
// ### Use math.abs in map
// ```
// # import "math"
// # import "sampledata"
//
// < sampledata.float()
// >     |> map(fn: (r) => ({r with _value: math.abs(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.abs(x: ±Inf) // Returns +Inf
// math.abs(x: NaN) // Returns NaN
// ```
//
builtin abs : (x: float) => float

// acos returns the acosine of `x` in radians.
//
// ## Parameters
// - x: Value to operate on.
//
//   `x` should be greater than -1 and less than 1. Otherwise, the operation
//   will return `NaN`.
//
// ## Examples
//
// ### Return the acosine of a value
// ```no_run
// import "math"
//
// math.acos(x: 0.22) // 1.3489818562981022
// ```
//
// ### Use math.acos in map
// ```
// import "math"
// # import "sampledata"
// #
// # data = sampledata.float() |> map(fn: (r) => ({r with _value: r._value * .01}))
//
// < data
// >     |> map(fn: (r) => ({r with _value: math.acos(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.acos(x: <-1) // Returns NaN
// math.acos(x: >1) // Returns NaN
// ```
//
builtin acos : (x: float) => float

// acosh returns the inverse hyperbolic cosine of `x`.
//
// ## Parameters
// - x: Value to operate on.
//
//   `x` should be greater than 1. If less than 1 the operation will return `NaN`.
//
// ## Examples
//
// ### Return the inverse hyperbolic cosine of a value
// ```no_run
// import "math"
//
// math.acosh(x: 1.22)
// ```
//
// ### Use math.acosh in map
// ```
// import "math"
// # import "sampledata"
// #
// # data = sampledata.float() |> map(fn: (r) => ({r with _value: r._value * 0.1}))
//
// < data
// >     |> map(fn: (r) => ({r with _value: math.acosh(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.acosh(x: +Inf) // Returns +Inf
// math.acosh(x: <1) // Returns NaN
// math.acosh(x: NaN) // Returns NaN
// ```
//
builtin acosh : (x: float) => float

// asin returns the arcsine of `x` in radians.
//
// ## Parameters
// - x: Value to operate on.
//
//   `x` should be greater than -1 and less than 1. Otherwise the function will
//   return `NaN`.
//
// ## Examples
//
// ### Return the arcsine of a value
// ```no_run
// import "math"
//
// math.asin(x: 0.22)
// ```
//
// ### Use math.asin in map
// ```
// import "math"
// # import "sampledata"
// #
// # data = sampledata.float() |> map(fn: (r) => ({r with _value: r._value * .01}))
//
// < data
// >     |> map(fn: (r) => ({r with _value: math.asin(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.asin(x: ±0) // Returns ±0
// math.asin(x: <-1) // Returns NaN
// math.asin(x: >1) // Returns NaN
// ```
//
builtin asin : (x: float) => float

// asinh returns the inverse hyperbolic sine of `x`.
//
// ## Parameters
// - x: Value to operate on.
//
// ## Examples
//
// ### Return the inverse hyperbolic sine of a value
// ```no_run
// import "math"
//
// math.asinh(x: 3.14) // 1.8618125572133835
// ```
//
// ### Use math.asinh in map
// ```
// import "math"
// import "sampledata"
//
// < sampledata.float()
// >     |> map(fn: (r) => ({r with _value: math.asinh(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.asinh(x: ±0) // Returns ±0
// math.asinh(x: ±Inf) // Returns ±Inf
// math.asinh(x: NaN) // Returns NaN
// ```
//
builtin asinh : (x: float) => float

// atan returns the arctangent of `x` in radians.
//
// ## Parameters
// - x: Value to operate on.
//
// ## Examples
//
// ### Return the arctangent of a value
// ```no_run
// import "math"
//
// math.atan(x: 3.14) // 1.262480664599468
// ```
//
// ### Use math.atan in map
// ```
// import "math"
// import "sampledata"
//
// < sampledata.float()
// >     |> map(fn: (r) => ({r with _value: math.atan(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.atan(x: ±0) // Returns ±0
// math.atan(x: ±Inf) // Returns ±Pi/2
// ```
//
builtin atan : (x: float) => float

// atan2 returns the artangent of `x/y`, using the signs
// of the two to determine the quadrant of the return value.
//
// ## Parameters
// - y: y-coordinate to use in the operation.
// - x: x-corrdinate to use in the operation.
//
// ## Examples
//
// Return the arctangent of two values
// ```no_run
// import "math"
//
// math.atan2(y: 1.22, x: 3.14) // 0.3705838802763881
// ```
//
// Use math.atan2 in map
// ```
// # import "array"
// import "math"
// #
// # data = array.from(
// #     rows: [
// #         {_time: 2021-01-01T00:00:00Z, x: 1.2, y: 3.9},
// #         {_time: 2021-01-01T01:00:00Z, x: 2.4, y: 4.2},
// #         {_time: 2021-01-01T02:00:00Z, x: 3.6, y: 5.3},
// #         {_time: 2021-01-01T03:00:00Z, x: 4.8, y: 6.8},
// #         {_time: 2021-01-01T04:00:00Z, x: 5.1, y: 7.5},
// #     ],
// # )
//
// < data
// >     |> map(fn: (r) => ({_time: r._time, _value: math.atan2(x: r.x, y: r.y)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.atan2(y:y, x:NaN)        // Returns NaN
// math.atan2(y: NaN, x:x)       // Returns NaN
// math.atan2(y: +0, x: >=0)     // Returns +0
// math.atan2(y: -0, x: >=0)     // Returns -0
// math.atan2(y: +0, x: <=-0)    // Returns +Pi
// math.atan2(y: -0, x: <=-0)    // Returns -Pi
// math.atan2(y: >0, x: 0)       // Returns +Pi/2
// math.atan2(y: <0, x: 0)       // Returns -Pi/2
// math.atan2(y: +Inf, x: +Inf)  // Returns +Pi/4
// math.atan2(y: -Inf, x: +Inf)  // Returns -Pi/4
// math.atan2(y: +Inf, x: -Inf)  // Returns 3Pi/4
// math.atan2(y: -Inf, x: -Inf)  // Returns -3Pi/4
// math.atan2(y:y, x: +Inf)      // Returns 0
// math.atan2(y: >0, x: -Inf)    // Returns +Pi
// math.atan2(y: <0, x: -Inf)    // Returns -Pi
// math.atan2(y: +Inf, x:x)      // Returns +Pi/2
// math.atan2(y: -Inf, x:x)      // Returns -Pi/2
// ```
//
builtin atan2 : (y: float, x: float) => float

// atanh returns the inverse hyperbolic tangent of `x`.
//
// ## Parameters
// - x: Value to operate on.
//
//   `x` should be greater than -1 and less than 1. Otherwise the operation
//   will return `NaN`.
//
// ## Examples
//
// ### Return the hyperbolic tangent of a value
// ```no_run
// import "math"
//
// math.atanh(x: 0.22) // 0.22365610902183242
// ```
//
// ### Use math.atanh in map
// ```
// import "math"
// # import "sampledata"
// #
// # data = sampledata.float() |> map(fn: (r) => ({r with _value: r._value * .01}))
//
// < data
// >     |> map(fn: (r) => ({r with _value: math.atanh(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.atanh(x: 1)   // Returns +Inf
// math.atanh(x: ±0)  // Returns ±0
// math.atanh(x: -1)  // Returns -Inf
// math.atanh(x: <-1) // Returns NaN
// math.atanh(x: >1)  // Returns NaN
// math.atanh(x: NaN) // Returns NaN
// ```
//
builtin atanh : (x: float) => float

// cbrt returns the cube root of x.
//
// ## Parameters
// - x: Value to operate on.
//
// ## Examples
//
// ### Return the cube root of a value
// ```no_run
// import "math"
//
// math.cbrt(x: 1728.0) // 12.0
// ```
//
// ### Use math.cbrt in map
// ```
// import "math"
// import "sampledata"
//
// < sampledata.float()
// >     |> map(fn: (r) => ({r with _value: math.cbrt(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.cbrt(±0)   // Returns ±0
// math.cbrt(±Inf) // Returns ±Inf
// math.cbrt(NaN)  // Returns NaN
// ```
//
builtin cbrt : (x: float) => float

// ceil returns the least integer value greater than or equal to `x`.
//
// ## Parameters
// - x: Value to operate on.
//
// ## Examples
//
// ### Round a value up to the nearest integer
// ```no_run
// import "math"
//
// math.ceil(x: 3.14) // 4.0
// ```
//
// ### Use math.ceil in map
// ```
// import "math"
// import "sampledata"
//
// < sampledata.float()
// >     |> map(fn: (r) => ({r with _value: math.ceil(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.ceil(±0)   // Returns ±0
// math.ceil(±Inf) // Returns ±Inf
// math.ceil(NaN)  // Returns NaN
// ```
//
builtin ceil : (x: float) => float

// copysign returns a value with the magnitude `x` and the sign of `y`.
//
// ## Parameters
// - x: Magnitude to use in the operation.
// - y: Sign to use in the operation.
//
// ## Examples
//
// ### Return the copysign of two columns
// ```no_run
// import "math"
//
// math.copysign(x: 1.0, y: 2.0)
// ```
//
// ### Use math.copysign in map
// ```
// # import "array"
// import "math"
// #
// # data = array.from(
// #     rows: [
// #         {_time: 2021-01-01T00:00:00Z, x: 1.2, y: 3.9},
// #         {_time: 2021-01-01T01:00:00Z, x: 2.4, y: 4.2},
// #         {_time: 2021-01-01T02:00:00Z, x: 3.6, y: 5.3},
// #         {_time: 2021-01-01T03:00:00Z, x: 4.8, y: 6.8},
// #         {_time: 2021-01-01T04:00:00Z, x: 5.1, y: 7.5},
// #     ],
// # )
//
// < data
// >     |> map(fn: (r) => ({_time: r._time, _value: math.copysign(x: r.x, y: r.y)}))
// ```
//
builtin copysign : (x: float, y: float) => float

// cos returns the cosine of the radian argument `x`.
//
// ## Parameters
// - x: Value to operate on.
//
// ## Examples
//
// Return the cosine of a radian value
// ```no_run
// import "math"
//
// math.cos(x: 3.14) // -0.9999987317275396
// ```
//
// ### Use math.cos in map
// ```
// import "math"
// import "sampledata"
//
// sampledata.float()
//     |> map(fn: (r) => ({_time: r._time, _value: math.cos(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.cos(±Inf) // Returns NaN
// math.cos(NaN)  // Returns NaN
// ```
//
builtin cos : (x: float) => float

// cosh returns the hyperbolic cosine of `x`.
//
// ## Parameters
// - x: Value to operate on.
//
// ## Examples
//
// Return the hyperbolic cosine of a value
// ```no_run
// import "math"
//
// math.cosh(x: 1.22) // 1.8412089502726745
// ```
//
// ### Use math.cosh in map
// ```
// import "math"
// import "sampledata"
//
// sampledata.float()
//     |> map(fn: (r) => ({_time: r._time, _value: math.cosh(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.cosh(±0)   // Returns 1
// math.cosh(±Inf) // Returns +Inf
// math.cosh(NaN)  // Returns NaN
// ```
//
builtin cosh : (x: float) => float

// dim returns the maximum of `x - y` or `0`.
//
// ## Parameters
// - x: x-value to use in the operation.
// - y: y-value to use in the operation.
//
// ## Examples
//
// ### Return the maximum difference betwee two values
// ```no_run
// import "math"
//
// math.dim(x: 12.2, y: 8.1) // 4.1
// ```
//
// ### Use math.dim in map
// ```
// # import "array"
// import "math"
// #
// # data = array.from(
// #     rows: [
// #         {_time: 2021-01-01T00:00:00Z, x: 3.9, y: 1.2},
// #         {_time: 2021-01-01T01:00:00Z, x: 4.2, y: 2.4},
// #         {_time: 2021-01-01T02:00:00Z, x: 5.3, y: 3.6},
// #         {_time: 2021-01-01T03:00:00Z, x: 6.8, y: 4.8},
// #         {_time: 2021-01-01T04:00:00Z, x: 7.5, y: 5.1},
// #     ],
// # )
//
// < data
// >     |> map(fn: (r) => ({_time: r._time, _value: math.dim(x: r.x, y: r.y)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.dim(x: +Inf, y: +Inf) // Returns NaN
// math.dim(x: -Inf, y: -Inf) // Returns NaN
// math.dim(x: x, y: NaN)  // Returns NaN
// math.dim(x: NaN, y: y)     // Returns NaN
// ```
//
builtin dim : (x: float, y: float) => float

// erf returns the error function of `x`.
//
// ## Parameters
// - x: Value to operate on.
//
// ## Examples
//
// ### Return the error function of a value.
// ```no_run
// import "math"
//
// math.erf(x: 22.6) // 1.0
// ```
//
// ### Use math.erf in map
// ```
// import "math"
// import "sampledata"
//
// < sampledata.float()
// >     |> map(fn: (r) => ({r with _value: math.erf(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.erf(+Inf) // Returns 1
// math.erf(-Inf) // Returns -1
// math.erf(NaN)  // Returns NaN
// ```
//
builtin erf : (x: float) => float

// erfc returns the complementary error function of `x`.
//
// ## Parameters
// - x: Value to operate on.
//
// ## Examples
//
// ### Return the complementary error function of a value
// ```no_run
// import "math"
//
// math.erfc(x: 22.6) // 3.772618913849058e-224
// ```
//
// ### Use math.erfc in map
// ```
// import "math"
// import "sampledata"
//
// < sampledata.float()
// >     |> map(fn: (r) => ({r with _value: math.erfc(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.erfc(+Inf) // Returns 0
// math.erfc(-Inf) // Returns 2
// math.erfc(NaN)  // Returns NaN
// ```
//
builtin erfc : (x: float) => float

// erfcinv returns the inverse of `math.erfc()`.
//
// ## Parameters
// - x: Value to operate on.
//
//   `x` should be greater than 0 and less than 2. Otherwise the operation
//   will return `NaN`.
//
// ## Examples
//
// ### Return the inverse complimentary error function
// ```no_run
// import "math"
//
// math.erfcinv(x: 0.42345) // 0.5660037715858239
// ```
//
// ### Use math.erfcinv in map
// ```
// # import "sampledata"
// import "math"
// #
// # data = sampledata.float()
// #     |> map(fn: (r) => ({r with _value: math.erfc(x: r._value)}))
//
// < data
// >    |> map(fn: (r) => ({r with _value: math.erfcinv(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.erfcinv(x: 0)   // Returns +Inf
// math.erfcinv(x: 2)   // Returns -Inf
// math.erfcinv(x: <0)  // Returns NaN
// math.erfcinv(x: >2)  // Returns NaN
// math.erfcinv(x: NaN) // Returns NaN
// ```
//
builtin erfcinv : (x: float) => float

// erfinv returns the inverse error function of `x`.
//
// ## Parameters
// - x: Value to operate on.
//
//   `x` should be greater than -1 and less than 1. Otherwise, the operation will
//   return `NaN`.
//
// ## Examples
//
// ### Return the inverse error function of a value
// ```no_run
// import "math"
//
// math.erfinv(x: 0.22) // 0.19750838337227364
// ```
//
// ### Use math.erfinv in map
// ```
// # import "sampledata"
// import "math"
// #
// # data = sampledata.float()
// #     |> map(fn: (r) => ({r with _value: math.erf(x: r._value)}))
//
// < data
// >    |> map(fn: (r) => ({r with _value: math.erfinv(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.erfinv(x: 1)   // Returns +Inf
// math.erfinv(x: -1)  // Returns -Inf
// math.erfinv(x: <-1) // Returns NaN
// math.erfinv(x: > 1) // Returns NaN
// math.erfinv(x: NaN) // Returns NaN
// ```
//
builtin erfinv : (x: float) => float

// exp returns `e**x`, the base-e exponential of `x`.
//
// ## Parameters
// - x: Value to operate on.
//
// ## Examples
//
// ### Return the base-e exponential of a value
// ```no_run
// import "math"
//
// math.exp(x: 21.0) // 1.3188157344832146e+09
// ```
//
// ### Use math.exp in map
// ```
// import "sampledata"
// import "math"
//
// < sampledata.float()
// >    |> map(fn: (r) => ({r with _value: math.exp(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.exp(x: +Inf) // Returns +Inf
// math.exp(x: NaN)  // Returns NaN
// ```
//
builtin exp : (x: float) => float

// exp2 returns `2**x`, the base-2 exponential of `x`.
//
// ## Parameters
// - x: Value to operate on.
//
// ## Examples
//
// ### Return the base-2 exponential of a value
// ```no_run
// import "math"
//
// math.exp2(x: 21.0) // 2.097152e+06
// ```
//
// ### Use math.exp2 in map
// ```
// import "sampledata"
// import "math"
//
// < sampledata.float()
// >    |> map(fn: (r) => ({r with _value: math.exp2(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.exp2(x: +Inf) // Returns +Inf
// math.exp2(x: NaN)  // Returns NaN
// ```
//
// Very large values overflow to 0 or +Inf. Very small values overflow to 1.
//
builtin exp2 : (x: float) => float

// expm1 returns `e**x - 1`, the base-e exponential of `x` minus 1.
// It is more accurate than `math.exp(x:x) - 1` when `x` is near zero.
//
// ## Parameters
// - x: Value to operate on.
//
// ## Examples
//
// ### Get more accurate base-e exponentials for values near zero
// ```no_run
// import "math"
//
// math.expm1(x: 0.022) // 0.022243784470438233
// ```
//
// ### Use math.expm1 in map
// ```
// # import "sampledata"
// import "math"
// #
// # data = sampledata.float() |> map(fn: (r) => ({r with _value: r._value * .01}))
//
// < data
// >    |> map(fn: (r) => ({r with _value: math.expm1(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.expm1(+Inf) // Returns +Inf
// math.expm1(-Inf) // Returns -1
// math.expm1(NaN)  // Returns NaN
// ```
//
// Very large values overflow to -1 or +Inf.
//
builtin expm1 : (x: float) => float

// float64bits returns the IEEE 754 binary representation of `f`,
// with the sign bit of `f` and the result in the same bit position.
//
// ## Parameters
// - f: Value to operate on.
//
// ## Examples
//
// ### Return the binary expression of a value
// ```no_run
// import "math"
//
// math.float64bits(f: 1234.56) // 4653144467747100426
// ```
//
// ### Use math.float64bits in map
// ```
// import "sampledata"
// import "math"
//
// < sampledata.float()
// >    |> map(fn: (r) => ({r with _value: math.float64bits(f: r._value)}))
// ```
//
builtin float64bits : (f: float) => uint

// float64frombits returns the floating-point number corresponding to the IEE
// 754 binary representation `b`, with the sign bit of `b` and the result in the
// same bit position.
//
// ## Parameters
// - b: Value to operate on.
//
// ## Examples
//
// ### Convert bits into a float value
// ```no_run
// import "math"
//
// math.float64frombits(b: uint(v: 4)) // 2e-323
// ```
//
// ### Use math.float64frombits in map
// ```
// # import "sampledata"
// import "math"
// #
// # data = sampledata.float() |> map(fn: (r) => ({r with _value: math.float64bits(f: r._value)}))
//
// < data
// >    |> map(fn: (r) => ({r with _value: math.float64frombits(b: r._value)}))
// ```
//
builtin float64frombits : (b: uint) => float

// floor returns the greatest integer value less than or equal to `x`.
//
// ## Parameters
// - x: Value to operate on.
//
// ## Examples
//
// ### Return the nearest integer less than a value
// ```no_run
// import "math"
//
// math.floor(x: 1.22) // 1.0
// ```
//
// ### Use math.floor in map
// ```
// import "sampledata"
// import "math"
//
// < sampledata.float()
// >    |> map(fn: (r) => ({r with _value: math.floor(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.floor(±0)   // Returns ±0
// math.floor(±Inf) // Returns ±Inf
// math.floor(NaN)  // Returns NaN
// ```
//
builtin floor : (x: float) => float

// frexp breaks `f` into a normalized fraction and an integral part of two.
//
// It returns **frac** and **exp** satisfying `f == frac x 2**exp`,
// with the absolute value of **frac** in the interval [1/2, 1).
//
// ## Parameters
// - f: Value to operate on.
//
// ## Examples
//
// ### Return the normalize fraction and integral of a value
// ```no_run
// import "math"
//
// math.frexp(f: 22.0) // {exp: 5, frac: 0.6875}
// ```
//
// ### Use math.frexp in map
// ```
// import "sampledata"
// import "math"
//
// < sampledata.float()
//     |> map(
//         fn: (r) => {
//             result = math.frexp(f: r._value)
//
//             return {r with exp: result.exp, frac: result.frac}
//         },
// >     )
// ```
//
// ## Special cases
//
// ```no_run
// math.frexp(f: ±0)   // Returns {frac: ±0, exp: 0}
// math.frexp(f: ±Inf) // Returns {frac: ±Inf, exp: 0}
// math.frexp(f: NaN)  // Returns {frac: NaN, exp: 0}
// ```
//
builtin frexp : (f: float) => {frac: float, exp: int}

// gamma returns the gamma function of `x`.
//
// ## Parameters
// - x: Value to operate on.
//
// ## Examples
//
// ### Return the gamma function of a value
// ```no_run
// import "math"
//
// math.gamma(x: 2.12) // 1.056821007887572
// ```
//
// ### Use math.gamma in map
// ```
// import "sampledata"
// import "math"
//
// < sampledata.float()
// >    |> map(fn: (r) => ({r with _value: math.gamma(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.gamma(x: +Inf) = +Inf
// math.gamma(x: +0) = +Inf
// math.gamma(x: -0) = -Inf
// math.gamma(x: <0) = NaN for integer x < 0
// math.gamma(x: -Inf) = NaN
// math.gamma(x: NaN) = NaN
// ```
//
builtin gamma : (x: float) => float

// hypot returns the square root of `p*p + q*q`, taking care to avoid overflow
// and underflow.
//
// ## Parameters
// - p: p-value to use in the operation.
// - q: q-value to use in the operation.
//
// ## Examples
//
// ### Return the hypotenuse of two values
// ```no_run
// import "math"
//
// math.hypot(p: 2.0, q: 5.0) // 5.385164807134505
// ```
//
// ### Use math.hypot in map
// ```
// # import "array"
// import "math"
// #
// # data = array.from(
// #     rows: [
// #         {triangle: "t1", a: 12.3, b: 11.7},
// #         {triangle: "t2", a: 109.6, b: 23.3},
// #         {triangle: "t3", a: 8.2, b: 34.2},
// #         {triangle: "t4", a: 33.9, b: 28.0},
// #         {triangle: "t5", a: 25.0, b: 25.0},
// #     ],
// # )
//
// < data
// >     |> map(fn: (r) => ({r with _value: math.hypot(p: r.a, q: r.b)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.hypot(p: ±Inf, q:q) // Returns +Inf
// math.hypot(p:p, q: ±Inf) // Returns +Inf
// math.hypot(p: NaN, q:q)  // Returns NaN
// math.hypot(p:p, q: NaN)  // Returns NaN
// ```
//
builtin hypot : (p: float, q: float) => float

// ilogb returns the binary exponent of `x` as an integer.
//
// ## Parameters
// - x: Value to operate on.
//
// ## Examples
//
// ### Return the binary exponent of a value
// ```no_run
// import "math"
//
// math.ilogb(x: 123.45) // 6
// ```
//
// ### Use math.ilogb in map
// ```
// import "sampledata"
// import "math"
//
// < sampledata.float()
// >    |> map(fn: (r) => ({r with _value: math.ilogb(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.ilogb(x: ±Inf) // Returns MaxInt32
// math.ilogb(x: 0)    // Returns MinInt32
// math.ilogb(x: NaN)  // Returns MaxInt32
// ```
//
builtin ilogb : (x: float) => int

// mInf returns positive infinity if `sign >= 0`, negative infinity
// if `sign < 0`.
//
// ## Parameters
// - sign: Value to operate on.
//
// ## Examples
//
// ### Return an infinity float value from a positive or negative sign value
// ```no_run
// import "math"
//
// math.mInf(sign: 1) // +Inf
// ```
//
// ### Use math.mInf in map
// ```
// import "sampledata"
// import "math"
//
// < sampledata.int()
// >    |> map(fn: (r) => ({r with _value: math.mInf(sign: r._value)}))
// ```
//
builtin mInf : (sign: int) => float

// isInf reports whether `f` is an infinity, according to `sign`.
//
// If `sign > 0`, math.isInf reports whether `f` is positive infinity.
// If `sign < 0`, math.isInf reports whether `f` is negative infinity.
// If `sign  == 0`, math.isInf reports whether `f` is either infinity.
//
// ## Parameters
// - f: is the value used in the evaluation.
// - sign: is the sign used in the eveluation.
//
// ## Examples
//
// ### Test if a value is an infinity value
// ```no_run
// import "math"
//
// math.isInf(f: 2.12, sign: 3) // false
// ```
//
// ### Use math.isInf in map
// ```
// # import "sampledata"
// import "math"
// #
// # data = sampledata.float(includeNull: true)
// #     |> fill(value: float(v: "+Inf"))
//
// < data
// >     |> map(fn: (r) => ({r with _value: math.isInf(f: r._value, sign: 1)}))
// ```
//
builtin isInf : (f: float, sign: int) => bool

// isNaN reports whether `f` is an IEEE 754 "not-a-number" value.
//
// ## Parameters
// - f: Value to operate on.
//
// ## Examples
//
// ### Check if a value is a NaN float value
// ```no_run
// import "math"
//
// math.isNaN(f: 12.345) // false
// ```
//
// ### Use math.isNaN in map
// ```
// # import "sampledata"
// import "math"
// #
// # data = sampledata.float(includeNull: true)
// #     |> fill(value: float(v: "NaN"))
//
// < data
// >     |> map(fn: (r) => ({r with _value: math.isNaN(f: r._value)}))
// ```
//
builtin isNaN : (f: float) => bool

// j0 returns the order-zero Bessel function of the first kind.
//
// ## Parameters
// - x: Value to operate on.
//
// ## Examples
//
// ### Return the order-zero Bessel function of a value
// ```no_run
// import "math"
//
// math.j0(x: 1.23) // 0.656070571706025
// ```
//
// ### Use math.j0 in map
// ```
// import "sampledata"
// import "math"
//
// < sampledata.float()
// >     |> map(fn: (r) => ({r with _value: math.j0(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.j0(x: ±Inf) // Returns 0
// math.j0(x: 0)    // Returns 1
// math.j0(x: NaN)  // Returns NaN
// ```
//
builtin j0 : (x: float) => float

// j1 is a funciton that returns the order-one Bessel function for the first kind.
//
// ## Parameters
// - x: Value to operate on.
//
// ## Examples
//
// ### Return the order-one Bessel function of a value
// ```no_run
// import "math"
//
// math.j1(x: 1.23) // 0.5058005726280961
// ```
//
// ### Use math.j1 in map
// ```
// import "sampledata"
// import "math"
//
// < sampledata.float()
// >     |> map(fn: (r) => ({r with _value: math.j1(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.j1(±Inf) // Returns 0
// math.j1(NaN)  // Returns NaN
// ```
//
builtin j1 : (x: float) => float

// jn returns the order-n Bessel funciton of the first kind.
//
// ## Parameters
// - n: Order number.
// - x: Value to operate on.
//
// ## Examples
//
// ### Return the order-n Bessel function of a value
// ```no_run
// import "math"
//
// math.jn(n: 2, x: 1.23) // 0.16636938378681407
// ```
//
// ### Use math.jn in map
// ```
// import "sampledata"
// import "math"
//
// < sampledata.float()
// >     |> map(fn: (r) => ({r with _value: math.jn(n: 4, x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.jn(n:n, x: ±Inf) // Returns 0
// math.jn(n:n, x: NaN)  // Returns NaN
// ```
//
builtin jn : (n: int, x: float) => float

// ldexp is the inverse of `math.frexp()`. It returns `frac x 2**exp`.
//
// ## Parameters
// - frac: Fraction to use in the operation.
// - exp: Exponent to use in the operation.
//
// ## Examples
//
// ### Return the inverse of math.frexp
// ```no_run
// import "math"
//
// math.ldexp(frac: 0.5, exp: 6) // 32.0
// ```
//
// ### Use math.ldexp in map
// ```
// # import "array"
// import "math"
// #
// # data = array.from(
// #     rows: [
// #         {tag: "t1", _time: 2021-01-01T00:00:00Z, exp: 2, frac: -0.545},
// #         {tag: "t1", _time: 2021-01-01T00:00:10Z, exp: 4, frac: 0.6825},
// #         {tag: "t1", _time: 2021-01-01T00:00:20Z, exp: 3, frac: 0.91875},
// #         {tag: "t1", _time: 2021-01-01T00:00:30Z, exp: 5, frac: 0.5478125},
// #         {tag: "t1", _time: 2021-01-01T00:00:40Z, exp: 4, frac: 0.951875},
// #         {tag: "t1", _time: 2021-01-01T00:00:50Z, exp: 3, frac: 0.55375},
// #         {tag: "t2", _time: 2021-01-01T00:00:00Z, exp: 5, frac: 0.6203125},
// #         {tag: "t2", _time: 2021-01-01T00:00:10Z, exp: 3, frac: 0.62125},
// #         {tag: "t2", _time: 2021-01-01T00:00:20Z, exp: 2, frac: -0.9375},
// #         {tag: "t2", _time: 2021-01-01T00:00:30Z, exp: 5, frac: 0.6178125},
// #         {tag: "t2", _time: 2021-01-01T00:00:40Z, exp: 4, frac: 0.86625},
// #         {tag: "t2", _time: 2021-01-01T00:00:50Z, exp: 1, frac: 0.93},
// #     ],
// # )
// #     |> group(columns: ["tag"])
//
// < data
// >     |> map(fn: (r) => ({_time: r._time, tag: r.tag, _value: math.ldexp(frac: r.frac, exp: r.exp)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.ldexp(frac: ±0, exp:exp)   // Returns ±0
// math.ldexp(frac: ±Inf, exp:exp) // Returns ±Inf
// math.ldexp(frac: NaN, exp:exp)  // Returns NaN
// ```
//
builtin ldexp : (frac: float, exp: int) => float

// lgamma returns the natural logarithm and sign (-1 or +1) of `math.gamma(x:x)`.
//
// ## Parameters
// - x: Value to operate on.
//
// ## Examples
//
// ### Return the natural logarithm and sign of a gamma function
// ```no_run
// import "math"
//
// math.lgamma(x: 3.14) // {lgamma: 0.8261387047770286, sign: 1}
// ```
//
// ### Use math.lgamma in map
// ```
// import "sampledata"
// import "math"
//
// < sampledata.float()
//     |> map(
//         fn: (r) => {
//             result = math.lgamma(x: r._value)
//
//             return {r with lgamma: result.lgamma, sign: result.sign}
//         },
// >     )
// ```
//
// ## Special cases
//
// ```no_run
// math.lgamma(x: +Inf)     // Returns +Inf
// math.lgamma(x: 0)        // Returns +Inf
// math.lgamma(x: -integer) // Returns +Inf
// math.lgamma(x: -Inf)     // Returns -Inf
// math.lgamma(x: NaN)      // Returns NaN
// ```
//
builtin lgamma : (x: float) => {lgamma: float, sign: int}

// log returns the natural logarithm of `x`.
//
// ## Parameters
// - x: Value to operate on.
//
// ## Examples
//
// ### Return the natural logarithm of a value
// ```no_run
// import "math"
//
// math.log(x: 3.14) // 1.144222799920162
// ```
//
// ### Use math.log in map
// ```
// import "sampledata"
// import "math"
//
// sampledata.float()
//     |> map(fn: (r) => ({r with _value: math.log(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.log(x: +Inf) // Returns +Inf
// math.log(x: 0)    // Returns -Inf
// math.log(x: <0)   // Returns NaN
// math.log(x: NaN)  // Returns NaN
// ```
//
builtin log : (x: float) => float

// log10 returns the decimal logarithm of x.
//
// ## Parameters
// - x: Value to operate on.
//
// ## Examples
//
// ### Return the decimal lagarithm of a value
// ```no_run
// import "math"
//
// math.log10(x: 3.14) // 0.4969296480732149
// ```
//
// ### Use math.log10 in map
// ```
// import "sampledata"
// import "math"
//
// sampledata.float()
//     |> map(fn: (r) => ({r with _value: math.log10(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.log10(x: +Inf) // Returns +Inf
// math.log10(x: 0)    // Returns -Inf
// math.log10(x: <0)   // Returns NaN
// math.log10(x: NaN)  // Returns NaN
// ```
//
builtin log10 : (x: float) => float

// log1p returns the natural logarithm of 1 plus `x`.
// This operation is more accurate than `math.log(x: 1 + x)` when `x` is
// near zero.
//
// ## Parameters
// - x: Value to operate on.
//
// ## Examples
//
// ### Return the natural logarithm of values near zero
// ```no_run
// import "math"
//
// math.log1p(x: 0.56) // 0.44468582126144574
// ```
//
// ### Use math.log1p in map
// ```
// # import "sampledata"
// import "math"
// #
// # data = sampledata.float() |> map(fn: (r) => ({r with _value: r._value * .01}))
//
// < data
// >     |> map(fn: (r) => ({r with _value: math.log1p(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// import "math"
//
// math.log1p(x: +Inf) // Returns +Inf
// math.log1p(x: ±0)   // Returns ±0
// math.log1p(x: -1)   // Returns -Inf
// math.log1p(x: <-1)  // Returns NaN
// math.log1p(x: NaN)  // Returns NaN
// ```
//
builtin log1p : (x: float) => float

// log2 is a function returns the binary logarithm of `x`.
//
// ## Parameters
// - x: the value used in the operation.
//
// ## Examples
//
// ### Return the binary logarithm of a value
// ```no_run
// import "math"
//
// math.log2(x: 3.14) // 1.6507645591169022
// ```
//
// ### Use math.log2 in map
// ```
// import "sampledata"
// import "math"
//
// < sampledata.float()
// >     |> map(fn: (r) => ({r with _value: math.log2(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.log2(x: +Inf) // Returns +Inf
// math.log2(x: 0)    // Returns -Inf
// math.log2(x: <0)   // Returns NaN
// math.log2(x: NaN)  // Returns NaN
// ```
//
builtin log2 : (x: float) => float

// logb returns the binary exponent of `x`.
//
// ## Parameters
// - x: Value to operate on.
//
// ## Examples
//
// ### Return the binary exponent of a value
// ```no_run
// import "math"
//
// math.logb(x: 3.14) // 1
// ```
//
// ### Use math.logb in map
// ```
// import "sampledata"
// import "math"
//
// < sampledata.float()
// >     |> map(fn: (r) => ({r with _value: math.logb(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.logb(x: ±Inf) // Returns +Inf
// math.logb(x: 0)    // Returns -Inf
// math.logb(x: NaN)  // Returns NaN
// ```
//
builtin logb : (x: float) => float

// mMax returns the larger of `x` or `y`.
//
// ## Parameters
// - x: x-value to use in the operation.
// - y: y-value to use in the operation.
//
// ## Examples
//
// ### Return the larger of two values
// ```no_run
// import "math"
//
// math.mMax(x: 1.23, y: 4.56) // 4.56
// ```
//
// ### Use math.mMax in map
// ```
// import "math"
// # import "sampledata"
// #
// # data = sampledata.float()
// #     |> pivot(rowKey: ["_time"], columnKey: ["tag"], valueColumn: "_value")
//
// < data
// >     |> map(fn: (r) => ({_time: r._time, _value: math.mMax(x: r.t1, y: r.t2)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.mMax(x:x, y:+Inf)  // Returns +Inf
// math.mMax(x: +Inf, y:y) // Returns +Inf
// math.mMax(x:x, y: NaN)  // Returns NaN
// math.mMax(x: NaN, y:y)  // Returns NaN
// math.mMax(x: +0, y: ±0) // Returns +0
// math.mMax(x: ±0, y: +0) // Returns +0
// math.mMax(x: -0, y: -0) // Returns -0
// ```
//
builtin mMax : (x: float, y: float) => float

// mMin is a function taht returns the lessser of `x` or `y`.
//
// ## Parameters
// - x: x-value to use in the operation.
// - y: y-value to use in the operation.
//
// ## Examples
//
// ### Return the lesser of two values
// ```no_run
// import "math"
//
// math.mMin(x: 1.23, y: 4.56) // 1.23
// ```
//
// ### Use math.mMin in map
// ```
// import "math"
// # import "sampledata"
// #
// # data = sampledata.float()
// #     |> pivot(rowKey: ["_time"], columnKey: ["tag"], valueColumn: "_value")
//
// < data
// >     |> map(fn: (r) => ({_time: r._time, _value: math.mMin(x: r.t1, y: r.t2)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.mMin(x:x, y: -Inf) // Returns -Inf
// math.mMin(x: -Inf, y:y) // Returns -Inf
// math.mMin(x:x, y: NaN)  // Returns NaN
// math.mMin(x: NaN, y:y)  // Returns NaN
// math.mMin(x: -0, y: ±0) // Returns -0
// math.mMin(x: ±0, y: -0) // Returns -0
// ```
//
builtin mMin : (x: float, y: float) => float

// mod returns a floating-point remainder of `x/y`.
//
// The magnitude of the result is less than `y` and its sign agrees
// with that of `x`.
//
// **Note**: `math.mod()` performs the same operation as the modulo operator (`%`).
// For example: `4.56 % 1.23`
//
// ## Parameters
// - x: x-value to use in the operation.
// - y: y-value to use in the operation.
//
// ## Examples
//
// ### Return the modulo of two values
// ```no_run
// import "math"
//
// math.mod(x: 4.56, y: 1.23) // 0.8699999999999997
// ```
//
// ### Use math.mod in map
// ```
// import "math"
// # import "sampledata"
// #
// # data = sampledata.float()
// #     |> pivot(rowKey: ["_time"], columnKey: ["tag"], valueColumn: "_value")
//
// < data
// >     |> map(fn: (r) => ({_time: r._time, _value: math.mod(x: r.t1, y: r.t2)}))
// ```
//
// ## Special cases
// ```no_run
// math.mod(x: ±Inf, y:y)  // Returns NaN
// math.mod(x: NaN, y:y)   // Returns NaN
// math.mod(x:x, y: 0)     // Returns NaN
// math.mod(x:x, y: ±Inf)  // Returns x
// math.mod(x:x, y: NaN)   // Returns NaN
// ```
//
builtin mod : (x: float, y: float) => float

// modf returns integer and fractional floating-point numbers that sum to `f`.
//
// Both values have the same sign as `f`.
//
// ## Parameters
// - f: Value to operate on.
//
// ## Examples
//
// ### Return the integer and float that sum to a value
// ```no_run
// import "math"
//
// math.modf(f: 3.14) // {frac: 0.14000000000000012, int: 3}
// ```
//
// ### Use math.modf in map
// ```
// import "math"
// import "sampledata"
//
// < sampledata.float()
//     |> map(
//         fn: (r) => {
//             result = math.modf(f: r._value)
//
//             return {_time: r._time, int: result.int, frac: result.frac}
//         }
// >     )
// ```
//
// ## Special cases
//
// ```no_run
// math.modf(f: ±Inf) // Returns {int: ±Inf, frac: NaN}
// math.modf(f: NaN)  // Returns {int: NaN, frac: NaN}
// ```
//
builtin modf : (f: float) => {int: float, frac: float}

// NaN returns a IEEE 754 "not-a-number" value.
//
// ## Examples
//
// ### Return a NaN value
// ```no_run
// import "math"
//
// math.NaN()
// ```
//
builtin NaN : () => float

// nextafter returns the next representable float value after `x` towards `y`.
//
// ## Parameters
// - x: x-value to use in the operation.
// - y: y-value to use in the operation.
//
// ## Examples
//
// ### Return the next possible float value
// ```no_run
// import "math"
//
// math.nextafter(x: 1.23, y: 4.56) // 1.2300000000000002
// ```
//
// ### Use math.nextafter in map
// ```
// import "math"
// # import "sampledata"
// #
// # data = sampledata.float()
// #     |> pivot(rowKey: ["_time"], columnKey: ["tag"], valueColumn: "_value")
//
// < data
// >     |> map(fn: (r) => ({_time: r._time, _value: math.nextafter(x: r.t1, y: r.t2)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.nextafter(x:x, y:x)    // Returns x
// math.nextafter(x: NaN, y:y) // Returns NaN
// math.nextafter(x:x, y:NaN)  // Returns NaN
// ```
//
builtin nextafter : (x: float, y: float) => float

// pow returns `x**y`, the base-x exponential of `y`.
//
// ## Parameters
// - x: Base value to operate on.
// - y: Exponent value.
//
// ## Examples
//
// ### Return the base-x exponential of a value
// ```no_run
// import "math"
//
// math.pow(x: 2.0, y: 3.0) // 8.0
// ```
//
// ### Use math.pow in map
// ```
// import "math"
// # import "sampledata"
// #
// # data = sampledata.float()
// #     |> pivot(rowKey: ["_time"], columnKey: ["tag"], valueColumn: "_value")
//
// < data
// >     |> map(fn: (r) => ({_time: r._time, _value: math.pow(x: r.t1, y: r.t2)}))
// ```
//
// ## Special cases
//
// ```no_run
// // In order of priority
// math.pow(x:x, y:±0)     // Returns 1 for any x
// math.pow(x:1, y:y)      // Returns 1 for any y
// math.pow(x:X, y:1)      // Returns x for any x
// math.pow(x:NaN, y:y)    // Returns NaN
// math.pow(x:x, y:NaN)    // Returns NaN
// math.pow(x:±0, y:y)     // Returns ±Inf for y an odd integer < 0
// math.pow(x:±0, y:-Inf)  // Returns +Inf
// math.pow(x:±0, y:+Inf)  // Returns +0
// math.pow(x:±0, y:y)     // Returns +Inf for finite y < 0 and not an odd integer
// math.pow(x:±0, y:y)     // Returns ±0 for y an odd integer > 0
// math.pow(x:±0, y:y)     // Returns +0 for finite y > 0 and not an odd integer
// math.pow(x:-1, y:±Inf)  // Returns 1
// math.pow(x:x, y:+Inf)   // Returns +Inf for |x| > 1
// math.pow(x:x, y:-Inf)   // Returns +0 for |x| > 1
// math.pow(x:x, y:+Inf)   // Returns +0 for |x| < 1
// math.pow(x:x, y:-Inf)   // Returns +Inf for |x| < 1
// math.pow(x:+Inf, y:y)   // Returns +Inf for y > 0
// math.pow(x:+Inf, y:y)   // Returns +0 for y < 0
// math.pow(x:-Inf, y:y)   // Returns math.pow(-0, -y)
// math.pow(x:x, y:y)      // Returns NaN for finite x < 0 and finite non-integer y
// ```
//
builtin pow : (x: float, y: float) => float

// pow10 returns 10**n, the base-10 exponential of `n`.
//
// ## Parameters
// - n: Exponent value.
//
// ## Examples
//
// ### Return the base-10 exponential of n
// ```no_run
// import "math"
//
// math.pow10(n: 3) // 1000.0
// ```
//
// ### Use math.pow10 in map
// ```
// import "math"
// import "sampledata"
//
// < sampledata.int()
// >     |> map(fn: (r) => ({r with _value: math.pow10(n: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.pow10(n: <-323) // Returns 0
// math.pow10(n: >308)  // Returns +Inf
// ```
//
builtin pow10 : (n: int) => float

// remainder returns the IEEE 754 floating-point remainder of `x/y`.
//
// ## Parameters
// - x: Numerator to use in the operation.
// - y: Denominator to use in the operation.
//
// ## Examples
//
// ### Return the remainder of division between two values
// ```no_run
// import "math"
//
// math.remainder(x: 21.0, y: 4.0) // 1.0
// ```
//
// ### Use math.remainder in map
// ```
// import "math"
// # import "sampledata"
// #
// # data = sampledata.float()
// #     |> pivot(rowKey: ["_time"], columnKey: ["tag"], valueColumn: "_value")
//
// < data
// >     |> map(fn: (r) => ({_time: r._time, _value: math.remainder(x: r.t1, y: r.t2)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.remainder(x: ±Inf, y:y)  // Returns NaN
// math.remainder(x: NaN, y:y)   // Returns NaN
// math.remainder(x:x, y: 0)     // Returns NaN
// math.remainder(x:x, y: ±Inf)  // Returns x
// math.remainder(x:x, y: NaN)   // Returns NaN
// ```
//
builtin remainder : (x: float, y: float) => float

// round returns the nearest integer, rounding half away from zero.
//
// ## Parameters
// - x: Value to operate on.
//
// ## Examples
//
// ### Round a value to the nearest whole number
// ```no_run
// import "math"
//
// math.round(x: 2.12) // 2.0
// ```
//
// ### Use math.round in map
// ```
// import "math"
// import "sampledata"
//
// < sampledata.float()
// >     |> map(fn: (r) => ({r with _value: math.round(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.round(x: ±0)   // Returns ±0
// math.round(x: ±Inf) // Returns ±Inf
// math.round(x: NaN)  // Returns NaN
// ```
//
builtin round : (x: float) => float

// roundtoeven returns the nearest integer, rounding ties to even.
//
// ## Parameters
// - x: Value to operate on.
//
// ## Examples
//
// ### Round a value to the nearest integer
// ```no_run
// import "math"
//
// math.roundtoeven(x: 3.14) // 3.0
// math.roundtoeven(x: 3.5) // 4.0
// ```
//
// ### Use math.roundtoeven in map
// ```
// import "math"
// import "sampledata"
//
// < sampledata.float()
// >     |> map(fn: (r) => ({r with _value: math.roundtoeven(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.roundtoeven(x: ±0)   // Returns ±0
// math.roundtoeven(x: ±Inf) // Returns ±Inf
// math.roundtoeven(x: NaN)  // Returns NaN
// ```
//
builtin roundtoeven : (x: float) => float

// signbit reports whether `x` is negative or negative zero.
//
// ## Parameters
// - x: Value to evaluate.
//
// ## Examples
//
// ### Test if a value is negative
// ```no_run
// import "math"
//
// math.signbit(x: -1.2) // true
// ```
//
// ### Use math.signbit in map
// ```
// import "math"
// # import "sampledata"
// #
// # data = sampledata.float(includeNull: true) |> fill(value: -0.0)
//
// < data
// >     |> map(fn: (r) => ({r with _value: math.signbit(x: r._value)}))
// ```
//
builtin signbit : (x: float) => bool

// sin returns the sine of the radian argument `x`.
//
// ## Parameters
// - x: Radian value to use in the operation.
//
// ## Examples
//
// ### Return the sine of a radian value
// ```no_run
// import "math"
//
// math.sin(x: 3.14) // 0.0015926529164868282
// ```
//
// ### Use math.sin in map
// ```
// import "math"
// import "sampledata"
//
// < sampledata.float()
// >     |> map(fn: (r) => ({r with _value: math.sin(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.sin(x: ±0)   // Returns ±0
// math.sin(x: ±Inf) // Returns NaN
// math.sin(x: NaN)  // Returns NaN
// ```
//
builtin sin : (x: float) => float

// sincos returns the values of `math.sin(x:x)` and `math.cos(x:x)`.
//
// ## Parameters
// - x: Value to operate on.
//
// ## Examples
//
// ### Return the sine and cosine of a value
// ```no_run
// import "math"
//
// math.sincos(x: 1.23) // {cos: 0.3342377271245026, sin: 0.9424888019316975}
// ```
//
// ### Use math.sincos in map
// ```
// import "math"
// import "sampledata"
//
// < sampledata.float()
//     |> map(
//         fn: (r) => {
//             result = math.sincos(x: r._value)
//
//             return {_time: r._time, tag: r._tag, sin: result.sin, cos: result.cos}
//         }
// >     )
// ```
//
// ## Special cases
//
// ```no_run
// math.sincos(x: ±0)   // Returns {sin: ±0, cos: 1}
// math.sincos(x: ±Inf) // Returns {sin: NaN, cos: NaN}
// math.sincos(x: NaN)  // Returns {sin: NaN, cos:  NaN}
// ```
//
builtin sincos : (x: float) => {sin: float, cos: float}

// sinh returns the hyperbolic sine of `x`.
//
// ## Parameters
// - x: Value to operate on.
//
// ## Examples
//
// ### Return the hyperbolic sine of a value
// ```no_run
// import "math"
//
// math.sinh(x: 1.23) // 1.564468479304407
// ```
//
// ### Use math.sinh in map
// ```
// import "math"
// import "sampledata"
//
// < sampledata.float()
// >     |> map(fn: (r) => ({r with _value: math.sinh(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.sinh(x: ±0)   // Returns ±0
// math.sinh(x: ±Inf) // Returns ±Inf
// math.sinh(x: NaN)  // Returns NaN
// ```
//
builtin sinh : (x: float) => float

// sqrt returns the square root of `x`.
//
// ## Parameters
// - x: Value to operate on.
//
// ## Examples
//
// ### Return the square root of a value
// ```no_run
// import "math"
//
// math.sqrt(x: 4.0) // 2.0
// ```
//
// ### Use math.sqrt in map
// ```
// import "math"
// import "sampledata"
//
// < sampledata.float()
// >     |> map(fn: (r) => ({r with _value: math.sqrt(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.sqrt(x: +Inf) // Returns +Inf
// math.sqrt(x: ±0)   // Returns ±0
// math.sqrt(x: <0)   // Returns NaN
// math.sqrt(x: NaN)  // Returns NaN
// ```
//
builtin sqrt : (x: float) => float

// tan returns the tangent of the radian argument `x`.
//
// ## Parameters
// - x: Value to operate on.
//
// ## Examples
//
// ### Return the tangent of a radian value
// ```no_run
// import "math"
//
// math.tan(x: 3.14) // -0.001592654936407223
// ```
//
// ### Use math.tan in map
// ```
// import "math"
// import "sampledata"
//
// < sampledata.float()
// >     |> map(fn: (r) => ({r with _value: math.tan(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.tan(x: ±0)   // Returns ±0
// math.tan(x: ±Inf) // Returns NaN
// math.tan(x: NaN)  // Returns NaN
// ```
//
builtin tan : (x: float) => float

// tanh returns the hyperbolic tangent of `x`.
//
// ## Parameters
// - x: Value to operate on.
//
// ## Examples
//
// ### Return the hyperbolic tangent of a value
// ```no_run
// import "math"
//
// math.tanh(x: 1.23) // 0.8425793256589296
// ```
//
// ### Use math.tanh in map
// ```
// import "math"
// import "sampledata"
//
// < sampledata.float()
// >     |> map(fn: (r) => ({r with _value: math.tanh(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.tanh(x: ±0)   // Returns ±0
// math.tanh(x: ±Inf) // Returns ±1
// math.tanh(x: NaN)  // Returns NaN
// ```
//
builtin tanh : (x: float) => float

// trunc returns the integer value of `x`.
//
// ## Parameters
// - x: Value to operate on.
//
// ## Examples
//
// ### Truncate a value at the decimal
// ```no_run
// import "math"
//
// math.trunc(x: 3.14) // 3.0
// ```
//
// ### Use math.trunc in map
// ```
// import "math"
// import "sampledata"
//
// < sampledata.float()
// >     |> map(fn: (r) => ({r with _value: math.trunc(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.trunc(x: ±0)   // Returns ±0
// math.trunc(x: ±Inf) // Returns ±Inf
// math.trunc(x: NaN)  // Returns NaN
// ```
//
builtin trunc : (x: float) => float

// y0 returns the order-zero Bessel function of the second kind.
//
// ## Parameters
// - x: Value to operate on.
//
// ## Examples
//
// ### Return the order-zero Bessel function of a value
// ```no_run
// import "math"
//
// math.y0(x: 3.14) // 0.3289375969127807
// ```
//
// ### Use math.y0 in map
// ```
// import "math"
// import "sampledata"
//
// < sampledata.float()
// >     |> map(fn: (r) => ({r with _value: math.y0(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.y0(x: +Inf) // Returns 0
// math.y0(x: 0)    // Returns -Inf
// math.y0(x: <0)   // Returns NaN
// math.y0(x: NaN)  // Returns NaN
// ```
//
builtin y0 : (x: float) => float

// y1 returns the order-one Bessel function of the second kind.
//
// ## Parameters
// - x: Value to operate on.
//
// ## Examples
//
// ### Return the order-one Bessel function of a value
// ```no_run
// import "math"
//
// math.y1(x: 3.14) // 0.35853138083924085
// ```
//
// ### Use math.y1 in map
// ```
// import "math"
// import "sampledata"
//
// < sampledata.float()
// >     |> map(fn: (r) => ({r with _value: math.y1(x: r._value)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.y1(x: +Inf) // Returns 0
// math.y1(x: 0)    // Returns -Inf
// math.y1(x: <0)   // Returns NaN
// math.y1(x: NaN)  // Returns NaN
// ```
//
builtin y1 : (x: float) => float

// yn returns the order-n Bessel function of the second kind.
//
// ## Parameters
// - n: Order number to use in the operation.
// - x: Value to operate on.
//
// ## Examples
//
// ### Return the order-n Bessel function of a value
// ```no_run
// import "math"
//
// math.yn(n: 3, x: 3.14) // -0.4866506930335083
// ```
//
// ### Use math.yn in map
// ```
// # import "array"
// import "math"
// #
// # data = array.from(
// #     rows: [
// #         {_time: 2021-01-01T00:00:00Z, x: 1.2, n: 3},
// #         {_time: 2021-01-01T01:00:00Z, x: 2.4, n: 4},
// #         {_time: 2021-01-01T02:00:00Z, x: 3.6, n: 5},
// #         {_time: 2021-01-01T03:00:00Z, x: 4.8, n: 6},
// #         {_time: 2021-01-01T04:00:00Z, x: 5.1, n: 7},
// #     ],
// # )
//
// < data
// >     |> map(fn: (r) => ({_time: r._time, _value: math.yn(n: r.n, x: r.x)}))
// ```
//
// ## Special cases
//
// ```no_run
// math.yn(n:n, x: +Inf) // Returns 0
// math.yn(n: ≥0, x: 0)  // Returns -Inf
// math.yn(n: <0, x: 0)  // Returns +Inf if n is odd, -Inf if n is even
// math.yn(n:n, x: <0)   // Returns NaN
// math.yn(n:n, x:NaN)   // Returns NaN
// ```
//
builtin yn : (n: int, x: float) => float

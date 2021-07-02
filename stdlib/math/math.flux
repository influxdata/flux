// Package math provides basic constants and mathematical functions
package math


// on floating point numbers.
builtin pi : float
builtin e : float
builtin phi : float
builtin sqrt2 : float
builtin sqrte : float
builtin sqrtpi : float
builtin sqrtphi : float
builtin ln2 : float
builtin log2e : float
builtin ln10 : float
builtin log10e : float
builtin maxfloat : float
builtin smallestNonzeroFloat : float
builtin maxint : int
builtin minint : int
builtin maxuint : uint

// abs is a function that returns the absolute value of x.
//
// ## Parameters
// - `x` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.abs(x: -1.22)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.abs(x: ±Inf) // returns +Inf
// math.abs(x: NaN) // returns NaN
// ```
builtin abs : (x: float) => float

// acos is a funciton that returns the acosine of x in radians.
//
// ## Parameters
// - `x` is the value used in the operation.
//
//   x should be greater than -1 and less than 1. Otherwise, the operation
//   will return NaN.
//
// ## Example
//
// ```
// import "math"
//
// math.acos(x: 0.22)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.acos(x: <-1) // returns NaN
// math.acos(x: >1) // returns NaN
// ```
builtin acos : (x: float) => float

// acosh is a function that returns the inverse hyperbolic cosine of x.
//
// ## Parameters
// - `x` is the value used in the operation.
//
//   x should be greater than 1. If less than 1 the operation will return NaN.
//
// ## Example
//
// ```
// import "math"
//
// math.acosh(x: 1.22)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.acosh(x: +Inf) // returns +Inf
// math.acosh(x: <1) // returns NaN
// math.acosh(x: NaN) // returns NaN
// ```
builtin acosh : (x: float) => float

// asin is a function that returns the arcsine of x in radians.
//
// ## Parameters
// - `x` is is value used in the operation.
//
//   x should be greater than -1 and less than 1. Otherwise the function will
//   return NaN.
//
// ## Example
//
// ```
// import "math"
//
// math.asin(x: 0.22)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.asin(x: ±0) // returns ±0
// math.asin(x: <-1) // returns NaN
// math.asin(x: >1) // returns NaN
// ```
builtin asin : (x: float) => float

// asinh is a function that returns the inverse hyperbolic sine of x.
//
// ## Parameters
// - `x` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.asinh(x: 3.14)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.asinh(x: ±0) // returns ±0
// math.asinh(x: ±Inf) // returns ±Inf
// math.asinh(x: NaN) // returns NaN
// ```
builtin asinh : (x: float) => float

// atan is a function that returns the arctangent of x in radians.
//
// ## Parameters
// - `x` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.atan(x: 3.14)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.atan(x: ±0) // returns ±0
// math.atan(x: ±Inf) // returns ±Pi/2
// ```
builtin atan : (x: float) => float

// atan2 is a function that returns the artangent of x/y, using the signs
//  of the two to determine the quadrant of the return value.
//
// ## Parameters
// - `y` is the y-coordinate used in the operation.
// - `x` is the x-corrdinate used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.atan2(y: 1.22, x: 3.14)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
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
builtin atan2 : (y: float, x: float) => float

// atanh is a function that returns the inverse hyperbolic tangent of x.
//
// ## Parameters
// - `x` is the value used in the operation.
//
//   x should be greater than -1 and less than 1, otherwise the operation
//   will return NaN.
//
// ## Example
//
// ```
// import "math"
//
// math.atanh(x: 0.22)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.atanh(x: 1)   // Returns +Inf
// math.atanh(x: ±0)  // Returns ±0
// math.atanh(x: -1)  // Returns -Inf
// math.atanh(x: <-1) // Returns NaN
// math.atanh(x: >1)  // Returns NaN
// math.atanh(x: NaN) // Returns NaN
// ```
builtin atanh : (x: float) => float

// cbrt is a function that returns the cube root of x.
//
// ## Parameters
// - `x` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.cbrt(x: 1728.0)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.cbrt(±0)   // Returns ±0
// math.cbrt(±Inf) // Returns ±Inf
// math.cbrt(NaN)  // Returns NaN
// ```
builtin cbrt : (x: float) => float

// ceil is a function that returns the least integer value greater than
//  or equal to x.
//
// ## Parameters
// - `x` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.ceil(x: 3.14)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.ceil(±0)   // Returns ±0
// math.ceil(±Inf) // Returns ±Inf
// math.ceil(NaN)  // Returns NaN
// ```
builtin ceil : (x: float) => float

// copysign is a function that returns a value with the magnitude x and
//  the sign of y.
//
// ## Parameters
// - `x` is the magnitude used in the operation.
// - `y` is the sign used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.copysign(x: 1.0, y: 2.0)
// ```
builtin copysign : (x: float, y: float) => float

// cos is a function that returns the cosine of the radian argument x.
//
// ## Parameters
// - `x` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.cos(x: 3.14)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.cos(±Inf) // Returns NaN
// math.cos(NaN)  // Returns NaN
// ```
builtin cos : (x: float) => float

// cosh is a function that returns the hyperbolic cosine of x.
//
// ## Parameters
// - `x` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.cosh(x: 1.22)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.cosh(±0)   // Returns 1
// math.cosh(±Inf) // Returns +Inf
// math.cosh(NaN)  // Returns NaN
// ```
builtin cosh : (x: float) => float

// dim is a function that returns the maximum of x - y or 0.
//
// ## Parameters
// - `x` is the X-value used in the operation .
// - 'y' is the Y-value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.dim(x: 12.2, y: 8.1)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.dim(x: +Inf, y: +Inf) // Returns NaN
// math.dim(x: -Inf, y: -Inf) // Returns NaN
// math.dim(x:x, y    : NaN)  // Returns NaN
// math.dim(x: NaN, y :y)     // Returns NaN
// ```
builtin dim : (x: float, y: float) => float

// erf is a function that returns the error function of x
//
// ## Parameters
// - `x` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.erf(x: 22.6)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.erf(+Inf) // Returns 1
// math.erf(-Inf) // Returns -1
// math.erf(NaN)  // Returns NaN
// ```
builtin erf : (x: float) => float

// erfc is a function that returns the complementary error function of x.
//
// ## Parameters
// - `x` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.erfc(x: 22.6)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.erfc(+Inf) // Returns 0
// math.erfc(-Inf) // Returns 2
// math.erfc(NaN)  // Returns NaN
// ```
builtin erfc : (x: float) => float

// erfcinv is a function that returns the inverse of math.erfc().
//
// ## Parameters
// - `x` is the value used in the operation.
//
//   x should be greater than 0 and less than 2. Otherwise the operation
//   will return NaN.
//
// ## Example
//
// ```
// import "math"
//
// math.erfcinv(x: 0.42345)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.erfcinv(x: 0)   // Returns +Inf
// math.erfcinv(x: 2)   // Returns -Inf
// math.erfcinv(x: <0)  // Returns NaN
// math.erfcinv(x: >2)  // Returns NaN
// math.erfcinv(x: NaN) // Returns NaN
// ```
builtin erfcinv : (x: float) => float

// erfinv is a function that returns the inverse error function of x.
//
// ## Parameter
// - `x` is the value used in the operation.
//
//   x should be greater than -1 and less than 1. Otherwise, the operation will
//   return NaN.
//
// ## Example
//
// ```
// import "math"
//
// math.erfinv(x: 0.22)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.erfinv(x: 1)   // Returns +Inf
// math.erfinv(x: -1)  // Returns -Inf
// math.erfinv(x: <-1) // Returns NaN
// math.erfinv(x: > 1) // Returns NaN
// math.erfinv(x: NaN) // Returns NaN
// ```
builtin erfinv : (x: float) => float

// exp is a function that returns `e**x`, the base-e exponential of x.
//
// ## Parameters
// - `x` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.exp(x: 21.0)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.exp(x: +Inf) // Returns +Inf
// math.exp(x: NaN)  // Returns NaN
// ```
builtin exp : (x: float) => float

// exp2 is a function that returns `2**x`, the base-2 exponential of x.
//
// ## Parameters
// - `x` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.exp2(x: 21.0)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.exp2(x: +Inf) // Returns +Inf
// math.exp2(x: NaN)  // Returns NaN
// ```
//
// Very large values overflow to 0 or +Inf. Very small values overflow to 1.
builtin exp2 : (x: float) => float

// expm1 is a function that returns `e**x - 1`, the base-e exponential of x minus
//  1. It is more accurate than `math.exp(x:x) - 1` when x is near zero.
//
// ## Parameters
// - `x` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.expm1(x: 1.22)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.expm1(+Inf) // Returns +Inf
// math.expm1(-Inf) // Returns -1
// math.expm1(NaN)  // Returns NaN
// ```
//
// Very large values overflow to -1 or +Inf.
builtin expm1 : (x: float) => float

// float64bits is a function that returns the IEEE 754 binary representation of f,
//  with the sign bit of f and the result in the same bit position.
//
// ## Parameters
// - `f` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.float64bits(f: 1234.56)
// ```
builtin float64bits : (f: float) => uint

// float64frombits is a function that returns the floating-point number corresponding
//  to the IEE 754 binary representation b, with the sign bit of b and the result in the
//  same bit position.
//
// ## Parameters
// - `b` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.float64frombits(b: 4)
// ```
builtin float64frombits : (b: uint) => float

// floor is a function that returns the greatest integer value less than or
//  equal to x.
//
// ## Parameters
// - `x` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.floor(x: 1.22)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.floor(±0)   // Returns ±0
// math.floor(±Inf) // Returns ±Inf
// math.floor(NaN)  // Returns NaN
// ```
builtin floor : (x: float) => float

// frexp is a function that breaks f into a normalized fraction and an
//  integral part of two.
//
//  It returns frac and exp satisfying `f == frac x 2**exp`,
//  with the absolute value of frac in the interval [1/2, 1).
//
// ## Parameters
// - `f` the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.frexp(f: 22.0)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.frexp(f: ±0)   // Returns {frac: ±0, exp: 0}
// math.frexp(f: ±Inf) // Returns {frac: ±Inf, exp: 0}
// math.frexp(f: NaN)  // Returns {frac: NaN, exp: 0}
// ```
builtin frexp : (f: float) => {frac: float, exp: int}

// gamma is a function that returns the gamma function of x.
//
// ## Parameters
// - `x` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.gamma(x: 2.12)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.gamma(x: +Inf) = +Inf
// math.gamma(x: +0) = +Inf
// math.gamma(x: -0) = -Inf
// math.gamma(x: <0) = NaN for integer x < 0
// math.gamma(x: -Inf) = NaN
// math.gamma(x: NaN) = NaN
// ```
builtin gamma : (x: float) => float

// hypot is a function that returns the square root of `p*p + q*q`, taking
//  care to avoid overflow and underflow.
//
// ## Params
// - `p` is the p-value used in the operation.
// - `q` is the q-value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.hypot(p: 2.0, q: 5.0)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.hypot(p: ±Inf, q:q) // Returns +Inf
// math.hypot(p:p, q: ±Inf) // Returns +Inf
// math.hypot(p: NaN, q:q)  // Returns NaN
// math.hypot(p:p, q: NaN)  // Returns NaN
// ```
builtin hypot : (x: float) => float

// ilogb is a function that returns the binary exponent of x as an integer.
//
// ## Parameters
// - `x` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.ilogb(x: 123.45)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.ilogb(x: ±Inf) // Returns MaxInt32
// math.ilogb(x: 0)    // Returns MinInt32
// math.ilogb(x: NaN)  // Returns MaxInt32
// ```
builtin ilogb : (x: float) => float

// mInf is a function that returns positive infinity if `sign >= 0`, negative infinity
// if `sign < 0`
//
// ## Parameters
// - `sign` is the sign value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.mInf(sign: 1)
// ```
builtin mInf : (sign: int) => float

// isInf is a function that reports whether f is an infinity, according to sign.
//
// If `sign > 0`, math.isInf reports whether f is positive infinity.
// If `sign < 0`, math.isInf reports whether f is negative infinity.
// If `sign  == 0`, math.isInf reports whether f is either infinity.
//
// ## Parameters
// - `f` is the value used in the evaluation.
// - `sign` is the sign used in the eveluation.
//
// ## Example
//
// ```
// import "math"
//
// math.isInf(f: 2.12, sign: 3)
// ```
builtin isInf : (f: float, sign: int) => bool

// isNaN is a function that reports whether f is an IEEE 754 "not-a-number" value.
//
// ## Parameters
// - `f` is the value used in the evaluation.
//
// ## Example
//
// ```
// import "math"
//
// math.isNaN(f: 12.345)
// ```
builtin isNaN : (f: float) => bool

// j0 is a function that returns the order-zero Bessel function of the first kind.
//
// ## Parameters
// - `x` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.j0(x: 1.23)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.j0(x: ±Inf) // Returns 0
// math.j0(x: 0)    // Returns 1
// math.j0(x: NaN)  // Returns NaN
// ```
builtin j0 : (x: float) => float

// j1 is a funciton that returns the order-one Bessel function for the first kind.
//
// ## Parameters
// - `x` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.j1(x: 1.23)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.j1(±Inf) // Returns 0
// math.j1(NaN)  // Returns NaN
// ```
builtin j1 : (x: float) => float

// jn is a function that returns the order-n Bessel funciton of the first kind.
//
// ## Parameters
// - `n` is the order number.
// - `x` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.jn(n: 2, x: 1.23)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.jn(n:n, x: ±Inf) // Returns 0
// math.jn(n:n, x: NaN)  // Returns NaN
// ```
builtin jn : (n: int, x: float) => float

// ldexp is a function that is the inverse of math.frexp(). It returns
//  `frac x 2**exp`. 
//
// ## Parameters
// - `frac` is the fraction used in the operation.
// - `exp` is the exponent used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.ldexp(frac: 0.5, exp: 6)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.ldexp(frac: ±0, exp:exp)   // Returns ±0
// math.ldexp(frac: ±Inf, exp:exp) // Returns ±Inf
// math.ldexp(frac: NaN, exp:exp)  // Returns NaN
// ```
builtin ldexp : (frac: float, exp: int) => float

// lgamma is a function that returns the natural logarithm and sign
//  (-1 or +1) of math.gamma(x:x).
//
// ## Parameters
// - `x` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.lgamma(x: 3.14)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.lgamma(x: +Inf)     // Returns +Inf
// math.lgamma(x: 0)        // Returns +Inf
// math.lgamma(x: -integer) // Returns +Inf
// math.lgamma(x: -Inf)     // Returns -Inf
// math.lgamma(x: NaN)      // Returns NaN
// ```
builtin lgamma : (x: float) => {lgamma: float, sign: int}

// log is a function that returns the natural logarithm of x.
//
// ## Parameters
// - `x` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
// 
// math.log(x: 3.14)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.log(x: +Inf) // Returns +Inf
// math.log(x: 0)    // Returns -Inf
// math.log(x: <0)   // Returns NaN
// math.log(x: NaN)  // Returns NaN
// ```
builtin log : (x: float) => float

// log10 is a function that returns the decimal logarithm of x.
//
// ## Params
// - `x` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.log10(x: 3.14)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.log10(x: +Inf) // Returns +Inf
// math.log10(x: 0)    // Returns -Inf
// math.log10(x: <0)   // Returns NaN
// math.log10(x: NaN)  // Returns NaN
// ```
builtin log10 : (x: float) => float

// log1p is a function that returns the natural logarithm of 1 plus the
//  argument x. it is more accurate than `math.log(x: 1 + x)` when x is
//  near zero.
//
// ## Parameters
// - `x` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.log1p(x: 0.56)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
//math.log1p(x: +Inf) // Returns +Inf
// math.log1p(x: ±0)   // Returns ±0
// math.log1p(x: -1)   // Returns -Inf
// math.log1p(x: <-1)  // Returns NaN
// math.log1p(x: NaN)  // Returns NaN
// ```
builtin log1p : (x: float) => float

// log2 is a function returns the binary logarithm of x.
//
// ## Parameters
// - `x` the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.log2(X: 3.14)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.log2(x: +Inf) // Returns +Inf
// math.log2(x: 0)    // Returns -Inf
// math.log2(x: <0)   // Returns NaN
// math.log2(x: NaN)  // Returns NaN
// ```
builtin log2 : (x: float) => float

// logb is a function that returns the binary exponent of x.
//
// ## Parameters
// - `x` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.logb(x: 3.14)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.logb(x: ±Inf) // Returns +Inf
// math.logb(x: 0)    // Returns -Inf
// math.logb(x: NaN)  // Returns NaN
// ```
builtin logb : (x: float) => float

// mMax is a function that returns the larger of x or y.
//
// ## Parameters
// - `x` is the x-value used in the operation.
// - `y` is the y-value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.mMax(x: 1.23, y: 4.56)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.mMax(x:x, y:+Inf)  // Returns +Inf
// math.mMax(x: +Inf, y:y) // Returns +Inf
// math.mMax(x:x, y: NaN)  // Returns NaN
// math.mMax(x: NaN, y:y)  // Returns NaN
// math.mMax(x: +0, y: ±0) // Returns +0
// math.mMax(x: ±0, y: +0) // Returns +0
// math.mMax(x: -0, y: -0) // Returns -0
// ```
builtin mMax : (x: float, y: float) => float

// mMin is a function taht returns the lessser of x or y.
//
// ## Parameters
// - `x` is the x-value used in the operation.
// - `y` is the y-value used in the operation.
//
// ## Example
// ```
// import "math"
//
// math.mMin(x: 1.23, y: 4.56)
// ```
//
// ## Special Cases
// ```
// import "math"
//
// math.mMin(x:x, y: -Inf) // Returns -Inf
// math.mMin(x: -Inf, y:y) // Returns -Inf
// math.mMin(x:x, y: NaN)  // Returns NaN
// math.mMin(x: NaN, y:y)  // Returns NaN
// math.mMin(x: -0, y: ±0) // Returns -0
// math.mMin(x: ±0, y: -0) // Returns -0
// ```
builtin mMin : (x: float, y: float) => float

// mod is a function that returns a floating-point remainder of x/y.
//
//  The magnitude of the result is less than y and its sign agrees
//  with that of x.
//
// ## Parameters
// - `x` is the x-value used in the operation.
// - `y` is the y-value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.mod(x: 1.23, y: 4.56)
// ```
//
// ## Special Cases
//
// ```
// math.mod(x: ±Inf, y:y)  // Returns NaN
// math.mod(x: NaN, y:y)   // Returns NaN
// math.mod(x:x, y: 0)     // Returns NaN
// math.mod(x:x, y: ±Inf)  // Returns x
// math.mod(x:x, y: NaN)   // Returns NaN
// ```
builtin mod : (x: float, y: float) => float

// modf is a function that returns integer and fractional floating-point numbers
//  that sum to f. 
//
//  Both values have the same sign as f.
//
// ## Parameters
// - `f` is the value used in the operation
//
// ## Example
//
// ```
// import "math"
//
// math.modf(f: 3.14)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.modf(f: ±Inf) // Returns {int: ±Inf, frac: NaN}
// math.modf(f: NaN)  // Returns {int: NaN, frac: NaN}
// ```
builtin modf : (f: float) => {int: float, frac: float}

// NaN is a function that returns a IEEE 754 "not-a-number" value.
//
// ## Example
//
// ```
// import "math"
//
// math.NaN()
// ```
builtin NaN : () => float

// nextafter is a function that returns the next representable float value after
//  x towards y.
//
// ## Parameters
// - `x` is the x-vaue used in the operation.
// - `y` is the y-value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.nextafter(x: 1.23, y: 4.56)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.nextafter(x:x, y:x)    // Returns x
// math.nextafter(x: NaN, y:y) // Returns NaN
// math.nextafter(x:x, y:NaN)  // Returns NaN
// ```
builtin nextafter : (x: float, y: float) => float

// pow is a function that returns x**y, the base-x exponential of y.
//
// ## Example
//
// ```
// import "math"
//
// math.pow(x: 2.0, y: 3.0)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
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
builtin pow : (x: float, y: float) => float

// pow10 is a function that returns 10**n, the base-10 exponential of n.
//
// ## Parameters
// - `n` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.pow10(n: 3)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.pow10(n: <-323) // Returns 0
// math.pow10(n: >308)  // Returns +Inf
// ```
builtin pow10 : (n: int) => float

// remainder is a function that returns the IEEE 754 floating-point remainder
//  of x / y.
//
// ## Parameters
// - `x` is the numerator used in the operation.
// - `y` is the denominator used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.remainder(x: 21.0, y: 4.0)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.remainder(x: ±Inf, y:y)  // Returns NaN
// math.remainder(x: NaN, y:y)   // Returns NaN
// math.remainder(x:x, y: 0)     // Returns NaN
// math.remainder(x:x, y: ±Inf)  // Returns x
// math.remainder(x:x, y: NaN)   // Returns NaN
// ```
builtin remainder : (x: float, y: float) => float

// round is a function that returns the nearest integer, rounding half away
//  from zero.
//
// ## Parameters
// - `x` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.round(x: 2.12)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.round(x: ±0)   // Returns ±0
// math.round(x: ±Inf) // Returns ±Inf
// math.round(x: NaN)  // Returns NaN
// ```
builtin round : (x: float) => float

// roundtoeven is a function that returns the nearest integer, rounding
//  ties to even.
//
// ## Parameters
// - `x` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.roundtoeven(x: 3.14)
// math.roundtoeven(x: 3.5)
// ```
//
// ## Special Cases
//
// ```
// math.roundtoeven(x: ±0)   // Returns ±0
// math.roundtoeven(x: ±Inf) // Returns ±Inf
// math.roundtoeven(x: NaN)  // Returns NaN
// ```
builtin roundtoeven : (x: float) => float

// signbit is a function that reports whether x is negative of negative zero.
//
// ## Parameters
// - `x` is the value used in the evaluation.
//
// ## Example
//
// ```
// import "math"
//
// math.signbit(x: -1.2)
// ```
builtin signbit : (x: float) => bool

// sin is a function that returns the sine of the radian argument x.
//
// ## Parameters
// - `x` is the radian value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.sin(x: 3.14)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.sin(x: ±0)   // Returns ±0
// math.sin(x: ±Inf) // Returns NaN
// math.sin(x: NaN)  // Returns NaN
// ```
builtin sin : (x: float) => float

// sincos is a function that returns the values of math.sin(x:x) and
//  math.cos(x:x).
//
// ## Parameters
// - `x` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.sincos(x: 1.23)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.sincos(x: ±0)   // Returns {sin: ±0, cos: 1}
// math.sincos(x: ±Inf) // Returns {sin: NaN, cos: NaN}
// math.sincos(x: NaN)  // Returns {sin: NaN, cos:  NaN}
// ```
builtin sincos : (x: float) => {sin: float, cos: float}

// sinh is a function that returns the hyperbolic sine of x.
//
// ## Parameters
// - `x` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.sinh(x: 1.23)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.sinh(x: ±0)   // Returns ±0
// math.sinh(x: ±Inf) // Returns ±Inf
// math.sinh(x: NaN)  // Returns NaN
// ```
builtin sinh : (x: float) => float

// sqrt is a function that returns the square root of x.
//
// ## Parameters
// - `x` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.sqrt(x: 4.0)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.sqrt(x: +Inf) // Returns +Inf
// math.sqrt(x: ±0)   // Returns ±0
// math.sqrt(x: <0)   // Returns NaN
// math.sqrt(x: NaN)  // Returns NaN
// ```
builtin sqrt : (x: float) => float

// tan is a function that returns the tangent of the radian argument.
//
// ## Parameters
// - `x` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.tan(x: 3.14)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.tan(x: ±0)   // Returns ±0
// math.tan(x: ±Inf) // Returns NaN
// math.tan(x: NaN)  // Returns NaN
// ```
builtin tan : (x: float) => float

// tanh is a function that returns the hyperbolic tangent of x.
//
// ## Parameters
// - `x` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.tanh(x: 1.23)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.tanh(x: ±0)   // Returns ±0
// math.tanh(x: ±Inf) // Returns ±1
// math.tanh(x: NaN)  // Returns NaN
// ```
builtin tanh : (x: float) => float

// trunc is a function that returns the integer value of x.
//
// ## Parameters
// - `x` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.trunc(x: 3.14)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.trunc(x: ±0)   // Returns ±0
// math.trunc(x: ±Inf) // Returns ±Inf
// math.trunc(x: NaN)  // Returns NaN
// ```
builtin trunc : (x: float) => float

// y0 is a function that returns the order-zero Bessel function of the
//  second kind.
//
// ## Parameters
// - `x` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.y0(x: 3.14)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.y0(x: +Inf) // Returns 0
// math.y0(x: 0)    // Returns -Inf
// math.y0(x: <0)   // Returns NaN
// math.y0(x: NaN)  // Returns NaN
// ```
builtin y0 : (x: float) => float

// y1 is a function that returns the order-one Bessel function of
//  the second kind.
//
// ## Parameters
// - `x` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.y1(x: 3.14)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.y1(x: +Inf) // Returns 0
// math.y1(x: 0)    // Returns -Inf
// math.y1(x: <0)   // Returns NaN
// math.y1(x: NaN)  // Returns NaN
// ```
builtin y1 : (x: float) => float

// yn is a function that returns the order-n Bessel function of
//  the second kind.
//
// ## Parameters
// - `n` is the order number used in the operation.
// - `x` is the value used in the operation.
//
// ## Example
//
// ```
// import "math"
//
// math.yn(n: 3, x: 3.14)
// ```
//
// ## Special Cases
//
// ```
// import "math"
//
// math.yn(n:n, x: +Inf) // Returns 0
// math.yn(n: ≥0, x: 0)  // Returns -Inf
// math.yn(n: <0, x: 0)  // Returns +Inf if n is odd, -Inf if n is even
// math.yn(n:n, x: <0)   // Returns NaN
// math.yn(n:n, x:NaN)   // Returns NaN
// ```
builtin yn : (n: int, x: float) => float

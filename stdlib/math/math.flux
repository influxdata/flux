// `math`
//
// The Flux math package provides basic constants and mathematical functions.
package math


// 3.14159265358979323846264338327950288419716939937510582097494459
builtin pi : float

// 2.71828182845904523536028747135266249775724709369995957496696763
builtin e : float

// 1.61803398874989484820458683436563811772030917980576286213544862
builtin phi : float

// 1.41421356237309504880168872420969807856967187537694807317667974
builtin sqrt2 : float

// 1.64872127070012814684865078781416357165377610071014801157507931
builtin sqrte : float

// 1.77245385090551602729816748334114518279754945612238712821380779
builtin sqrtpi : float

// 1.27201964951406896425242246173749149171560804184009624861664038
builtin sqrtphi : float

// 0.693147180559945309417232121458176568075500134360255254120680009
builtin ln2 : float

// 1 ÷ math.ln2
builtin log2e : float

// 2.30258509299404568401799145468436420760110148862877297603332790
builtin ln10 : float

// 1 ÷ math.ln10
builtin log10e : float

// 1.797693134862315708145274237317043567981e+308
builtin maxfloat : float

// 4.940656458412465441765687928682213723651e-324
builtin smallestNonzeroFloat : float

// 1<<63 - 1
builtin maxint : int

// -1 << 63
builtin minint : int

// 1<<64 - 1
builtin maxuint : uint


// Abs returns x as a positive value.
//
// Example
//
//    import "math"
//    math.abs(x: -10.42) // 10.42
builtin abs : (x: float) => float
builtin acos : (x: float) => float
builtin acosh : (x: float) => float
builtin asin : (x: float) => float
builtin asinh : (x: float) => float
builtin atan : (x: float) => float
builtin atan2 : (x: float, y: float) => float
builtin atanh : (x: float) => float
builtin cbrt : (x: float) => float
builtin ceil : (x: float) => float
builtin copysign : (x: float, y: float) => float
builtin cos : (x: float) => float
builtin cosh : (x: float) => float
builtin dim : (x: float, y: float) => float
builtin erf : (x: float) => float
builtin erfc : (x: float) => float
builtin erfcinv : (x: float) => float
builtin erfinv : (x: float) => float
builtin exp : (x: float) => float
builtin exp2 : (x: float) => float
builtin expm1 : (x: float) => float
builtin float64bits : (f: float) => uint
builtin float64frombits : (b: uint) => float
builtin floor : (x: float) => float
builtin frexp : (f: float) => {frac: float, exp: int}
builtin gamma : (x: float) => float
builtin hypot : (x: float) => float
builtin ilogb : (x: float) => float
builtin mInf : (sign: int) => float
builtin isInf : (f: float, sign: int) => bool
builtin isNaN : (f: float) => bool
builtin j0 : (x: float) => float
builtin j1 : (x: float) => float
builtin jn : (n: int, x: float) => float
builtin ldexp : (frac: float, exp: int) => float
builtin lgamma : (x: float) => {lgamma: float, sign: int}
builtin log : (x: float) => float
builtin log10 : (x: float) => float
builtin log1p : (x: float) => float
builtin log2 : (x: float) => float
builtin logb : (x: float) => float
builtin mMax : (x: float, y: float) => float
builtin mMin : (x: float, y: float) => float
builtin mod : (x: float, y: float) => float
builtin modf : (f: float) => {int: float, frac: float}
builtin NaN : () => float
builtin nextafter : (x: float, y: float) => float
builtin pow : (x: float, y: float) => float
builtin pow10 : (n: int) => float
builtin remainder : (x: float, y: float) => float
builtin round : (x: float) => float
builtin roundtoeven : (x: float) => float
builtin signbit : (x: float) => bool
builtin sin : (x: float) => float
builtin sincos : (x: float) => {sin: float, cos: float}
builtin sinh : (x: float) => float
builtin sqrt : (x: float) => float
builtin tan : (x: float) => float
builtin tanh : (x: float) => float
builtin trunc : (x: float) => float
builtin y0 : (x: float) => float
builtin y1 : (x: float) => float
builtin yn : (n: int, x: float) => float

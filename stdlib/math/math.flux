package math

// builtin constants
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

// builtin functions
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
builtin frexp : (f: float) => {frac: float , exp: int}
builtin gamma : (x: float) => float
builtin hypot : (x: float) => float
builtin ilogb : (x: float) => float
builtin mInf : (x: int) => float
builtin isInf : (f: float, sign: int) => bool
builtin isNaN : (f: float) => bool
builtin j0 : (x: float) => float
builtin j1 : (x: float) => float
builtin jn : (n: int, x: float) => float
builtin ldexp : (frac: float, exp: int) => float
builtin lgamma : (x: float) => {lgamma: float , sign: int}
builtin log : (x: float) => float
builtin log10 : (x: float) => float
builtin log1p : (x: float) => float
builtin log2 : (x: float) => float
builtin logb : (x: float) => float
builtin mMax : (x: float, y: float) => float
builtin mMin : (x: float, y: float) => float
builtin mod : (x: float, y: float) => float
builtin modf : (f: float) => {int: float , frac: float}
builtin NaN : () => float
builtin nextafter : (x: float, y: float) => float
builtin pow : (x: float, y: float) => float
builtin pow10 : (n: int) => float
builtin remainder : (x: float, y: float) => float
builtin round : (x: float) => float
builtin roundtoeven : (x: float) => float
builtin signbit : (x: float) => bool
builtin sin : (x: float) => float
builtin sincos : (x: float) => {sin: float , cos: float}
builtin sinh : (x: float) => float
builtin sqrt : (x: float) => float
builtin tan : (x: float) => float
builtin tanh : (x: float) => float
builtin trunc : (x: float) => float
builtin y0 : (x: float) => float
builtin y1 : (x: float) => float
builtin yn : (n: int, x: float) => float

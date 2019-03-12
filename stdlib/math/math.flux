package math

// builtin constants
builtin pi
builtin e
builtin phi
builtin sqrt2
builtin sqrte
builtin sqrtpi
builtin sqrtphi
builtin ln2
builtin log2e
builtin ln10
builtin log10e
builtin maxfloat
builtin smallestNonzeroFloat
builtin maxint
builtin minint
builtin maxuint

// builtin functions
builtin abs
builtin acos
builtin acosh
builtin asin
builtin asinh
builtin atan
builtin atan2
builtin atanh
builtin cbrt
builtin ceil
builtin copysign
builtin cos
builtin cosh
builtin dim
builtin erf
builtin erfc
builtin erfcinv
builtin erfinv
builtin exp
builtin exp2
builtin expm1
builtin float64bits
builtin float64frombits
builtin floor
builtin frexp
builtin gamma
builtin hypot
builtin ilogb
builtin mInf
builtin isInf
builtin isNaN
builtin j0
builtin j1
builtin jn
builtin ldexp
builtin lgamma
builtin log
builtin log10
builtin log1p
builtin log2
builtin logb
builtin mMax
builtin mMin
builtin mod
builtin modf
builtin NaN
builtin nextafter
builtin pow
builtin pow10
builtin remainder
builtin round
builtin roundtoeven
builtin signbit
builtin sin
builtin sincos
builtin sinh
builtin sqrt
builtin tan
builtin tanh
builtin trunc
builtin y0
builtin y1
builtin yn

// hack to simulate an imported math package
math = {
pi:pi
e:e
phi:phi
sqrt2:sqrt2
sqrte:sqrte
sqrtpi:sqrtpi
sqrtphi:sqrtphi
ln2:ln2
log2e:log2e
ln10:ln10
log10e:log10e
maxfloat:maxfloat
smallestNonzeroFloat:smallestNonzeroFloat
maxint:maxint
minint:minint
maxuint:maxuint
abs:abs
acos:acos
acosh:acosh
asin:asin
asinh:asinh
atan:atan
atan2:atan2
atanh:atanh
cbrt:cbrt
ceil:ceil
copysign:copysign
cos:cos
cosh:cosh
dim:dim
erf:erf
erfc:erfc
erfcinv:erfcinv
erfinv:erfinv
exp:exp
exp2:exp2
expm1:expm1
float64bits:float64bits
floor:floor
frexp:frexp
gamma:gamma
hypot:hypot
ilogb:ilogb
mInf:mInf
isInf:isInf
isNaN:isNaN
j0:j0
j1:j1
jn:jn
ldexp:ldexp
lgamma:lgamma
log:log
log10:log10
log1p:log1p
log2:log2
logb:logb
mMax:mMax
mMin:mMin
mod:mod
modf:modf
NaN:NaN
nextafter:nextafter
pow:pow
pow10:pow10
remainder:remainder
round:round
roundtoeven:roundtoeven
signbit:signbit
sin:sin
sincos:sincos
sinh:sinh
sqrt:sqrt
tan:tan
tanh:tanh
trunc:trunc
y0:y0
y1:y1
yn:yn
}

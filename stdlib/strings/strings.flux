package strings

// Transformation functions
builtin title : (v: string) => string
builtin toUpper : (v: string) => string
builtin toLower : (v: string) => string
builtin trim : (v: string, cutset: string) => string
builtin trimPrefix : (v: string, prefix: string) => string
builtin trimSpace : (v: string) => string
builtin trimSuffix : (v: string, suffix: string) => string
builtin trimRight : (v: string, cutset: string) => string
builtin trimLeft : (v: string, cutset: string) => string
builtin toTitle : (v: string) => string
builtin hasPrefix : (v: string, prefix: string) => bool
builtin hasSuffix : (v: string, suffix: string) => bool
builtin containsStr : (v: string, substr: string) => bool
builtin containsAny : (v: string, chars: string) => bool
builtin equalFold : (v: string, t: string) => bool
builtin compare : (v: string, t: string) => int
builtin countStr : (v: string, substr: string) => int
builtin index : (v: string, substr: string) => int
builtin indexAny : (v: string, chars: string) => int
builtin lastIndex : (v: string, substr: string) => int
builtin lastIndexAny : (v: string, chars: string) => int
builtin isDigit : (v: string) => bool
builtin isLetter : (v: string) => bool
builtin isLower : (v: string) => bool
builtin isUpper : (v: string) => bool
builtin repeat : (v: string, i: int) => string
builtin replace : (v: string, t: string, u: string, i: int) => string
builtin replaceAll : (v: string, t: string, u: string) => string
builtin split : (v: string, t: string) => [string]
builtin splitAfter : (v: string, t: string) => [string]
builtin splitN : (v: string, t: string, n: int) => [string]
builtin splitAfterN : (v: string, t: string, i: int) => [string]
builtin joinStr : (arr: [string], v: string) => string
builtin strlen : (v: string) => int
builtin substring : (v: string, start: int, end: int) => string

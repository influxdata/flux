package regexp

builtin compile : (v: string) => regexp
builtin quoteMeta : (v: string) => string
builtin findString : (r: regexp, v: string) => string
builtin findStringIndex : (r: regexp, v: string) => [int]
builtin matchRegexpString : (r: regexp, v: string) => bool
builtin replaceAllString : (r: regexp, v: string, t: string) => string
builtin splitRegexp : (r: regexp, v: string, i: int) => [string]
builtin getString : (r: regexp) => string

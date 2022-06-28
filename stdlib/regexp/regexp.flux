// Package regexp provides tools for working with regular expressions.
//
// ## Metadata
// introduced: 0.33.0
//
package regexp


// compile parses a string into a regular expression and returns a regexp type
// that can be used to match against strings.
//
// ## Parameters
// - v: String value to parse into a regular expression.
//
// ## Examples
//
// ### Convert a string into a regular expression
// ```no_run
// import "regexp"
//
// regexp.compile(v: "abcd")
// // Returns the regexp object /abcd/
// ```
//
// ## Metadata
// tags: type-conversions
//
builtin compile : (v: string) => regexp

// quoteMeta escapes all regular expression metacharacters in a string.
//
// ## Parameters
// - v: String that contains regular expression metacharacters to escape.
//
// ## Examples
//
// ### Escape regular expression metacharacters in a string
// ```no_run
// import "regexp"
//
// regexp.quoteMeta(v: ".+*?()|[]{}^$")
// // Returns "\.\+\*\?\(\)\|\[\]\{\}\^\$"
// ```
//
builtin quoteMeta : (v: string) => string

// findString returns the left-most regular expression match in a string.
//
// ## Parameters
// - r: Regular expression used to search `v`.
// - v: String value to search.
//
// ## Examples
//
// ### Return the first regular expression match in a string
// ```no_run
// import "regexp"
//
// regexp.findString(r: /foo.?/, v: "seafood fool")
// // Returns "food"
// ```
//
// ### Find the first regular expression match in each row
// ```
// import "regexp"
// import "sampledata"
//
// regex = /.{6}$/
//
// < sampledata.string()
// >     |> map(fn: (r) => ({r with _value: regexp.findString(v: r._value, r: regex)}))
// ```
//
builtin findString : (r: regexp, v: string) => string

// findStringIndex returns a two-element array of integers that represent the
// beginning and ending indexes of the first regular expression match in a string.
//
// ## Parameters
// - r: Regular expression used to search `v`.
// - v: String value to search.
//
// ## Examples
//
// ### Index the bounds of first regular expression match in each row
// ```no_run
// import "regexp"
//
// regexp.findStringIndex(r: /ab?/, v: "tablet")
// // Returns [1, 3]
// ```
//
builtin findStringIndex : (r: regexp, v: string) => [int]

// matchRegexpString tests if a string contains any match to a regular expression.
//
// ## Parameters
// - r: Regular expression used to search `v`.
// - v: String value to search.
//
// ## Examples
//
// ### Test if a string contains a regular expression match
// ```no_run
// import "regexp"
//
// regexp.matchRegexpString(r: /(gopher){2}/, v: "gophergophergopher")
// // Returns true
// ```
//
// ### Filter by rows that contain matches to a regular expression
// ```
// import "regexp"
// import "sampledata"
//
// sampledata.string()
//   |> filter(fn: (r) => regexp.matchRegexpString(r: /_\d/, v: r._value))
// ```
builtin matchRegexpString : (r: regexp, v: string) => bool

// replaceAllString replaces all reguar expression matches in a string with a
// specified replacement.
//
// ## Parameters
// - r: Regular expression used to search `v`.
// - v: String value to search.
// - t: Replacement for matches to `r`.
//
// ## Examples
//
// ### Replace regular expression matches in a string
// ```no_run
// import "regexp"
//
// regexp.replaceAllString(r: /a(x*)b/, v: "-ab-axxb-", t: "T")
// // Returns "-T-T-"
// ```
//
// ### Replace regular expression matches in string column values
// ```
// import "regexp"
// import "sampledata"
//
// < sampledata.string()
// >     |> map(fn: (r) => ({r with _value: regexp.replaceAllString(r: /smpl_/, v: r._value, t: "")}))
// ```
//
builtin replaceAllString : (r: regexp, v: string, t: string) => string

// splitRegexp splits a string into substrings separated by regular expression
// matches and returns an array of `i` substrings between matches.
//
// ## Parameters
// - r: Regular expression used to search `v`.
// - v: String value to be searched.
// - i: Maximum number of substrings to return.
//
//   -1 returns all matching substrings.
//
// ## Examples
//
// ### Return an array of regular expression matches
// ```no_run
// import "regexp"
//
// regexp.splitRegexp(r: /a*/, v: "abaabaccadaaae", i: -1)
// // Returns ["", "b", "b", "c", "c", "d", "e"]
// ```
//
builtin splitRegexp : (r: regexp, v: string, i: int) => [string]

// getString returns the source string used to compile a regular expression.
//
// ## Parameters
// - r: Regular expression object to convert to a string.
//
// ## Example
//
// ### Convert a regular expression to a string
// ```no_run
// import "regexp"
//
// regexp.getString(r: /[a-zA-Z]/)
// // Returns "[a-zA-Z]"
// ```
//
// ## Metadata
// tags: type-conversions
//
builtin getString : (r: regexp) => string

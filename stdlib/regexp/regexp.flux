// Flux regular expressions package includes functions that provide enhanced
// regular expression functionality.
package regexp


// compile is a function that parses a regular expression and,
//  if successful, returns a Regexp object that can be used to
//  match against text.
//
// ## Parameters
// - `v` is the string value to parse into regular expression.
//
// ## Example
//
// ```
// import "regexp"
//
// regexp.compile(v: "abcd")
// // Returns the regexp object `abcd`
// ```
//
// ## Use a string value as a regular expression
//
// ```
// import "regexp"
//
// data
//   |> map(fn: (r) => ({
//       r with
//       regexStr: r.regexStr,
//       _value: r._value,
//       firstRegexMatch: findString(
//         r: regexp.compile(v: regexStr),
//         v: r._value
//       )
//     })
//   )
// ```
builtin compile : (v: string) => regexp

// quoteMeta is a function that escapes all regular expression
//  metacharacters inside of a string.
//
// ## Parameters
// - `v` is the string that contains regular expression metacharacters
//   to escape
//
// ## Example
//
// ```
// import "regexp"
//
// regexp.quoteMeta(v: ".+*?()|[]{}^$")
// // Returns "\.\+\*\?\(\)\|\[\]\{\}\^\$"
// ```
//
// ## Escape regular expression meta characters in column values
//
// ```
// import "regexp"
//
// data
//   |> map(fn: (r) => ({
//       r with
//       notes: r.notes,
//       notes_escaped: regexp.quoteMeta(v: r.notes)
//     })
//   )
// ```
builtin quoteMeta : (v: string) => string

// findString is a function that returns the left-most regular expression
//  match in a string.
//
// ## Parameters
// - `r` is the regular expression used to search v.
// - `v` is the string value to search.
//
// ## Example
//
// ```
// import "regexp"
//
// regexp.findString(r: /foo.?/, v: "seafood fool")
// // Returns "food"
// ```
//
// ## Find the first regular expression match in each row
//
// ```
// import "regexp"
//
// data
//   |> map(fn: (r) => ({
//       r with
//       message: r.message,
//       regexp: r.regexp,
//       match: regexp.findString(r: r.regexp, v: r.message)
//     })
//   )
// ```
builtin findString : (r: regexp, v: string) => string

// findStringIdex is a function that returns a two-element array of integers
//  defining the beginning and ending indexes of the left-most regular
//  expression match in a string.
//
// ## Parameters
// - 'r' is the regular expression used to search v.
// - `v` is the string value to search.
//
// ## Example
//
// ```
// import "regexp"
//
// regexp.findStringIndex(r: /ab?/, v: "tablet")
// // Returns [1, 3]
// ```
//
// ## Index the bounds of first regular expression match in each row
//
// ```
// import "regexp"
//
// data
//   |> map(fn: (r) => ({
//       r with
//       regexStr: r.regexStr,
//       _value: r._value,
//       matchIndex: regexp.findStringIndex(
//         r: regexp.compile(r.regexStr),
//         v: r._value
//       )
//     })
//   )
// ```
builtin findStringIndex : (r: regexp, v: string) => [int]

// matchRegexpString is a function that tests if a string contains any
//  match to a regular expression.
//
// ## Parameters
// - `r` is the regular expression used to search v.
// - `v` is the string value to search.
//
// ## Example
//
// ```
// import "regexp"
//
// regexp.matchRegexpString(r: /(gopher){2}/, v: "gophergophergopher")
// // Returns true
// ```
//
// ## Filter by columns that contain matches to a regular expression
//
// ```
// import "regexp"
//
// data
//   |> filter(fn: (r) =>
//     regexp.matchRegexpString(
//       r: /Alert\:/,
//       v: r.message
//     )
//   )
// ```
builtin matchRegexpString : (r: regexp, v: string) => bool

// replaceAllString is a function that replaces all reguar expression matches
//  in a string with a specified replacement.
//
// ## Parameters
// - `r` is the regular expression used to search v.
// - `v` is the string value to search.
// - `t` is the replacement for matches to r.
//
// ## Example
//
// ```
// import "regexp"
//
// regexp.replaceAllString(r: /a(x*)b/, v: "-ab-axxb-", t: "T")
// // Returns "-T-T-"
// ```
//
// ## Replace regular expression matches in string column values
//
// ```
// import "regexp"
//
// data
//   |> map(fn: (r) => ({
//       r with
//       message: r.message,
//       updated_message: regexp.replaceAllString(
//         r: /cat|bird|ferret/,
//         v: r.message,
//         t: "dog"
//       )
//   }))
// ```
builtin replaceAllString : (r: regexp, v: string, t: string) => string

// splitRegexp is a function that splits a string into substrings separated
//  by regular expression matches and return an array of i substrings
//  between matches.
//
// ## Parameters
// - `r` is the regular expression used to search v.
// - `v` is the string value to be searched.
// - `i` is the maximum number of substrings to return.
//
//   -1 returns all matching substrings.
//
// ## Example
//
// ```
// import "regexp"
//
// regexp.splitRegexp(r: /a*/, v: "abaabaccadaaae", i: 5)
// // Returns ["", "b", "b", "c", "cadaaae"]
// ```
builtin splitRegexp : (r: regexp, v: string, i: int) => [string]

// getString is a function that returns the source string used to compile
//  a regular expression.
//
// ## Parameters
// - `r` is the regular expression object to convert to a string.
//
// ## Example
//
// ```
// import "regexp"
//
// regexp.getString(r: /[a-zA-Z]/)
// // Returns "[a-zA-Z]"
// ```
//
// ## Convert regular expressions into strings in each row
//
// ```
// import "regexp"
//
// data
//   |> map(fn: (r) => ({
//       r with
//       regex: r.regex,
//       regexStr: regexp.getString(r: r.regex)
//     })
//   )
// ```
builtin getString : (r: regexp) => string

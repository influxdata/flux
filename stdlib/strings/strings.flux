// Package strings provides functions to operate on UTF-8 encoded strings.
//
// ## Metadata
// introduced: 0.18.0
//
package strings


// title converts a string to title case.
//
// ## Parameters
//
// - v: String value to convert.
//
// ## Examples
//
// ### Convert all values of a column to title case
// ```
// import "sampledata"
// import "strings"
//
// < sampledata.string()
// >     |> map(fn: (r) => ({ r with _value: strings.title(v: r._value) }))
// ```
//
builtin title : (v: string) => string

// toUpper converts a string to uppercase.
//
// #### toUpper vs toTitle
// The results of `toUpper()` and `toTitle()` are often the same, however the
// difference is visible when using special characters:
//
// ```no_run
// str = "ǳ"
//
// strings.toUpper(v: str) // Returns Ǳ
// strings.toTitle(v: str) // Returns ǲ
// ```
//
// ## Parameters
//
// - v: String value to convert.
//
// ## Examples
//
// ### Convert all values of a column to upper case
// ```
// import "sampledata"
// import "strings"
//
// < sampledata.string()
// >     |> map(fn: (r) => ({ r with _value: strings.toUpper(v: r._value) }))
// ```
//
builtin toUpper : (v: string) => string

// toLower converts a string to lowercase.
//
// ## Parameters
//
// - v: String value to convert.
//
// ## Examples
//
// ### Convert all values of a column to lower case
// ```
// import "sampledata"
// import "strings"
//
// < sampledata.string()
// >     |> map(fn: (r) => ({r with _value: strings.toLower(v: r._value)}))
// ```
//
builtin toLower : (v: string) => string

// trim removes leading and trailing characters specified in the cutset from a string.
//
// ## Parameters
//
// - v: String to remove characters from.
// - cutset: Leading and trailing characters to remove from the string.
//
//   Only characters that match the cutset string exactly are trimmed.
//
// ## Examples
//
// ### Trim leading and trailing periods from all values in a column
// ```
// # import "sampledata"
// import "strings"
// #
// # data = sampledata.string() |> map(fn: (r) => ({r with _value: ".${r._value}."}))
//
// < data
// >     |> map(fn: (r) => ({r with _value: strings.trim(v: r._value, cutset: "smpl_")}))
// ```
//
builtin trim : (v: string, cutset: string) => string

// trimPrefix removes a prefix from a string. Strings that do not start with the prefix are returned unchanged.
//
// ## Parameters
//
// - v: String to trim.
// - prefix: Prefix to remove.
//
// ## Examples
//
// ### Trim a prefix from all values in a column
// ```
// import "sampledata"
// import "strings"
//
// < sampledata.string()
// >     |> map(fn: (r) => ({r with _value: strings.trimPrefix(v: r._value, prefix: "smpl_")}))
// ```
//
builtin trimPrefix : (v: string, prefix: string) => string

// trimSpace removes leading and trailing spaces from a string.
//
// ## Parameters
//
// - v: String to remove spaces from.
//
// ## Examples
//
// ### Trim leading and trailing spaces from all values in a column
// ```
// # import "sampledata"
// import "strings"
//
// # data = sampledata.string() |> map(fn: (r) => ({r with _value: "   ${r._value}   "}))
//
// < data
// >     |> map(fn: (r) => ({r with _value: strings.trimSpace(v: r._value)}))
// ```
//
builtin trimSpace : (v: string) => string

// trimSuffix removes a suffix from a string.
//
// Strings that do not end with the suffix are returned unchanged.
//
// ## Parameters
//
// - v: String to trim.
// - suffix: Suffix to remove.
//
// ## Examples
//
// ### Remove a suffix from all values in a column
// ```
// # import "sampledata"
// import "strings"
//
// # data = sampledata.string() |> map(fn: (r) => ({r with _value: "${r._value}_ex1"}))
//
// < data
// >     |> map(fn: (r) => ({ r with _value: strings.trimSuffix(v: r._value, suffix: "_ex1") }))
// ```
//
builtin trimSuffix : (v: string, suffix: string) => string

// trimRight removes trailing characters specified in the cutset from a string.
//
// ## Parameters
//
// - v: String to to remove characters from.
// - cutset: Trailing characters to trim from the string.
//
//   Only characters that match the cutset string exactly are trimmed.
//
// ## Examples
//
// ### Trim trailing periods from all values in a column
// ```
// # import "sampledata"
// import "strings"
// #
// # data = sampledata.string() |> map(fn: (r) => ({r with _value: "${r._value}..."}))
//
// < data
// >     |> map(fn: (r) => ({ r with _value: strings.trimRight(v: r._value, cutset: ".")}))
// ```
//
builtin trimRight : (v: string, cutset: string) => string

// trimLeft removes specified leading characters from a string.
//
// ## Parameters
//
// - v: String to to remove characters from.
// - cutset: Leading characters to trim from the string.
//
// ## Examples
//
// ### Trim leading periods from all values in a column
// ```
// # import "sampledata"
// import "strings"
// #
// # data = sampledata.string() |> map(fn: (r) => ({r with _value: "...${r._value}"}))
//
// < data
// >     |> map(fn: (r) => ({ r with _value: strings.trimLeft(v: r._value, cutset: ".")}))
// ```
//
builtin trimLeft : (v: string, cutset: string) => string

// toTitle converts all characters in a string to title case.
//
// #### toTitle vs toUpper
// The results of `toTitle()` and `toUpper()` are often the same, however the
// difference is visible when using special characters:
//
// ```no_run
// str = "ǳ"
//
// strings.toTitle(v: str) // Returns ǲ
// strings.toUpper(v: str) // Returns Ǳ
// ```
//
// ## Parameters
//
// - v: String value to convert.
//
// ## Examples
//
// ### Convert characters in a string to title case
// ```
// import "sampledata"
// import "strings"
//
// sampledata.string()
//     |> map(fn: (r) => ({r with _value: strings.toTitle(v: r._value)}))
// ```
//
builtin toTitle : (v: string) => string

// hasPrefix indicates if a string begins with a specified prefix.
//
// ## Parameters
//
// - v: String value to search.
// - prefix: Prefix to search for.
//
// ## Examples
//
// ### Filter based on the presence of a prefix in a column value
// ```
// import "sampledata"
// import "strings"
//
// < sampledata.string()
// >     |> filter(fn:(r) => strings.hasPrefix(v: r._value, prefix: "smpl_5" ))
// ```
//
builtin hasPrefix : (v: string, prefix: string) => bool

// hasSuffix indicates if a string ends with a specified suffix.
//
// ## Parameters
//
// - v: String value to search.
// - suffix: Suffix to search for.
//
// ## Examples
//
// ### Filter based on the presence of a suffix in a column value
// ```
// import "sampledata"
// import "strings"
//
// < sampledata.string()
// >     |> filter(fn:(r) => strings.hasSuffix(v: r._value, suffix: "4" ))
// ```
//
builtin hasSuffix : (v: string, suffix: string) => bool

// containsStr reports whether a string contains a specified substring.
//
// ## Parameters
//
// - v: String value to search.
// - substr: Substring value to search for.
//
// ## Examples
//
// ### Filter based on the presence of a substring in a column value
// ```
// import "sampledata"
// import "strings"
//
// < sampledata.string()
// >     |> filter(fn: (r) => strings.containsStr(v: r._value, substr: "5"))
// ```
//
builtin containsStr : (v: string, substr: string) => bool

// containsAny reports whether a specified string contains characters from another string.
//
// ## Parameters
//
// - v: String value to search.
// - chars: Characters to search for.
//
// ## Examples
//
// ### Filter based on the presence of a specific characters in a column value
// ```
// import "sampledata"
// import "strings"
//
// < sampledata.string()
// >     |> filter(fn: (r) => strings.containsAny(v: r._value, chars: "a79"))
// ```
//
builtin containsAny : (v: string, chars: string) => bool

// equalFold reports whether two UTF-8 strings are equal under Unicode case-folding.
//
// ## Parameters
//
// - v: String value to compare.
// - t: String value to compare against.
//
// ## Examples
//
// ### Ignore case when comparing two strings
// ```
// # import "array"
// import "strings"
// #
// # data = array.from(
// #     rows: [
// #         {time: 2022-01-01T00:00:00Z, string1: "RJqcVGNlcJ", string2: "rjQCvGNLCj"},
// #         {time: 2022-01-01T00:01:00Z, string1: "hBumdSljCQ", string2: "unfbcNAXUA"},
// #         {time: 2022-01-01T00:02:00Z, string1: "ITcHyLZuqu", string2: "KKtCcRHsKj"},
// #         {time: 2022-01-01T00:03:00Z, string1: "HyXdjvrjgp", string2: "hyxDJvrJGP"},
// #         {time: 2022-01-01T00:04:00Z, string1: "SVepvUBAVx", string2: "GuKKjuGsyI"},
// #     ],
// # )
//
// < data
// >     |> map(fn: (r) => ({r with same: strings.equalFold(v: r.string1, t: r.string2)}))
// ```
//
builtin equalFold : (v: string, t: string) => bool

// compare compares the lexicographical order of two strings.
//
// #### Return values
// | Comparison | Return value |
// | :--------- | -----------: |
// | v < t      |           -1 |
// | v == t     |            0 |
// | v > t      |            1 |
//
// ## Parameters
//
// - v: String value to compare.
// - t: String value to compare against.
//
// ## Examples
//
// ### Compare the lexicographical order of column values
// ```
// # import "array"
// import "strings"
// #
// # data = array.from(
// #     rows: [
// #         {time: 2022-01-01T00:00:00Z, string1: "RJqcVGNlcJ", string2: "rjQCvGNLCj"},
// #         {time: 2022-01-01T00:01:00Z, string1: "unfbcNAXUA", string2: "hBumdSljCQ"},
// #         {time: 2022-01-01T00:02:00Z, string1: "ITcHyLZuqu", string2: "ITcHyLZuqu"},
// #         {time: 2022-01-01T00:03:00Z, string1: "HyXdjvrjgp", string2: "hyxDJvrJGP"},
// #         {time: 2022-01-01T00:04:00Z, string1: "SVepvUBAVx", string2: "GuKKjuGsyI"},
// #     ],
// # )
//
// < data
// >     |> map(fn: (r) => ({r with same: strings.compare(v: r.string1, t: r.string2)}))
// ```
//
builtin compare : (v: string, t: string) => int

// countStr counts the number of non-overlapping instances of a substring appears in a string.
//
// ## Parameters
//
// - v: String value to search.
// - substr: Substring to count occurences of.
//
//   The function counts only non-overlapping instances of `substr`.
//
// ## Examples
//
// ### Count instances of a substring within a string
// ```
// import "sampledata"
// import "strings"
//
// < sampledata.string()
// >     |> map(fn: (r) => ({r with _value: strings.countStr(v: r._value, substr: "p")}))
// ```
//
builtin countStr : (v: string, substr: string) => int

// index returns the index of the first instance of a substring in a string.
// If the substring is not present, it returns `-1`.
//
// ## Parameters
//
// - v: String value to search.
// - substr: Substring to search for.
//
// ## Examples
//
// ### Find the index of the first occurrence of a substring
// ```
// import "sampledata"
// import "strings"
//
// < sampledata.string()
// >     |> map(fn: (r) => ({r with _value: strings.index(v: r._value, substr: "g")}))
// ```
//
builtin index : (v: string, substr: string) => int

// indexAny returns the index of the first instance of specified characters in a string.
// If none of the specified characters are present, it returns `-1`.
//
// ## Parameters
//
// - v: String value to search.
// - chars: Characters to search for.
//
// ## Examples
//
// ### Find the index of the first occurrence of characters from a string
// ```
// import "sampledata"
// import "strings"
//
// < sampledata.string()
// >     |> map(fn: (r) => ({r with _value: strings.indexAny(v: r._value, chars: "g7t")}))
// ```
//
builtin indexAny : (v: string, chars: string) => int

// lastIndex returns the index of the last instance of a substring in a string.
// If the substring is not present, the function returns -1.
//
// ## Parameters
//
// - v: String value to search.
// - substr: Substring to search for.
//
// ## Examples
//
// ### Find the index of the last occurrence of a substring
// ```
// import "sampledata"
// import "strings"
//
// < sampledata.string()
// >     |> map(fn: (r) => ({r with _value: strings.lastIndex(v: r._value, substr: "g")}))
// ```
//
builtin lastIndex : (v: string, substr: string) => int

// lastIndexAny returns the index of the last instance of any specified
// characters in a string.
// If none of the specified characters are present, the function returns `-1`.
//
// ## Parameters
//
// - v: String value to search.
// - chars: Characters to search for.
//
// ## Examples
//
// ### Find the index of the last occurrence of characters from a string
// ```
// import "sampledata"
// import "strings"
//
// < sampledata.string()
// >     |> map(fn: (r) => ({r with _value: strings.lastIndexAny(v: r._value, chars: "g7t")}))
// ```
//
builtin lastIndexAny : (v: string, chars: string) => int

// isDigit tests if a single-character string is a digit (0-9).
//
// ## Parameters
//
// - v: Single-character string to test.
//
// ## Examples
//
// ### Filter by columns with digits as values
// ```
// # import "regexp"
// # import "sampledata"
// import "strings"
// #
// # data = sampledata.string() |> map(fn: (r) => ({ r with _value: regexp.findString(r: /\S{1}$/, v: r._value)}))
//
// < data
// >     |> filter(fn: (r) => strings.isDigit(v: r._value))
// ```
//
builtin isDigit : (v: string) => bool

// isLetter tests if a single character string is a letter (a-z, A-Z).
//
// ## Parameters
//
// - v: Single-character string to test.
//
// ## Examples
//
// ### Filter by columns with digits as values
// ```
// # import "regexp"
// # import "sampledata"
// import "strings"
// #
// # data = sampledata.string() |> map(fn: (r) => ({ r with _value: regexp.findString(r: /\S{1}$/, v: r._value)}))
//
// < data
// >     |> filter(fn: (r) => strings.isLetter(v: r._value))
// ```
//
builtin isLetter : (v: string) => bool

// isLower tests if a single-character string is lowercase.
//
// ## Parameters
//
// - v: Single-character string value to test.
//
// ## Examples
//
// ### Filter by columns with single-letter lowercase values
// ```
// # import "array"
// import "strings"
// #
// # data = array.from(
// #     rows: [
// #         {_time: 2022-01-01T00:00:00Z, tag: "t1", _value: "a"},
// #         {_time: 2022-01-01T00:01:00Z, tag: "t1", _value: "B"},
// #         {_time: 2022-01-01T00:02:00Z, tag: "t1", _value: "C"},
// #         {_time: 2022-01-01T00:03:00Z, tag: "t1", _value: "d"},
// #         {_time: 2022-01-01T00:04:00Z, tag: "t1", _value: "e"},
// #         {_time: 2022-01-01T00:00:00Z, tag: "t2", _value: "F"},
// #         {_time: 2022-01-01T00:01:00Z, tag: "t2", _value: "g"},
// #         {_time: 2022-01-01T00:02:00Z, tag: "t2", _value: "H"},
// #         {_time: 2022-01-01T00:03:00Z, tag: "t2", _value: "i"},
// #         {_time: 2022-01-01T00:04:00Z, tag: "t2", _value: "J"},
// #     ]
// # )  |> group(columns: ["tag"])
//
// < data
// >    |> filter(fn: (r) => strings.isLower(v: r._value))
// ```
//
builtin isLower : (v: string) => bool

// isUpper tests if a single character string is uppercase.
//
// ## Parameters
//
// - v: Single-character string value to test.
//
// ## Examples
//
// ### Filter by columns with single-letter uppercase values
// ```
// # import "array"
// import "strings"
// #
// # data = array.from(
// #     rows: [
// #         {_time: 2022-01-01T00:00:00Z, tag: "t1", _value: "a"},
// #         {_time: 2022-01-01T00:01:00Z, tag: "t1", _value: "B"},
// #         {_time: 2022-01-01T00:02:00Z, tag: "t1", _value: "C"},
// #         {_time: 2022-01-01T00:03:00Z, tag: "t1", _value: "d"},
// #         {_time: 2022-01-01T00:04:00Z, tag: "t1", _value: "e"},
// #         {_time: 2022-01-01T00:00:00Z, tag: "t2", _value: "F"},
// #         {_time: 2022-01-01T00:01:00Z, tag: "t2", _value: "g"},
// #         {_time: 2022-01-01T00:02:00Z, tag: "t2", _value: "H"},
// #         {_time: 2022-01-01T00:03:00Z, tag: "t2", _value: "i"},
// #         {_time: 2022-01-01T00:04:00Z, tag: "t2", _value: "J"},
// #     ]
// # )  |> group(columns: ["tag"])
//
// < data
// >    |> filter(fn: (r) => strings.isUpper(v: r._value))
// ```
//
builtin isUpper : (v: string) => bool

// repeat returns a string consisting of `i` copies of a specified string.
//
// ## Parameters
//
// - v: String value to repeat.
// - i: Number of times to repeat `v`.
//
// ## Examples
//
// ### Repeat a string based on existing columns
// ```
// # import "math"
// # import "sampledata"
// import "strings"
// #
// # data = sampledata.float() |> map(fn: (r) => ({r with _value: math.abs(x: r._value) / 2.0})) |> toInt()
//
// < data
// >     |> map(fn: (r) => ({r with _value: strings.repeat(v: "ha", i: r._value)}))
// ```
//
builtin repeat : (v: string, i: int) => string

// replace replaces the first `i` non-overlapping instances of a substring with
// a specified replacement.
//
// ## Parameters
//
// - v: String value to search.
// - t: Substring value to replace.
// - u: Replacement for `i` instances of `t`.
// - i: Number of non-overlapping `t` matches to replace.
//
// ## Examples
//
// ### Replace a specific number of string matches
// ```
// import "sampledata"
// import "strings"
// < sampledata.string()
// >     |> map(fn: (r) => ({r with _value: strings.replace(v: r._value, t: "p", u: "XX", i: 2)}))
// ```
//
builtin replace : (v: string, t: string, u: string, i: int) => string

// replaceAll replaces all non-overlapping instances of a substring with a specified replacement.
//
// ## Parameters
//
// - v: String value to search.
// - t: Substring to replace.
// - u: Replacement for all instances of `t`.
//
// ## Examples
//
// ### Replace string matches
// ```
// import "sampledata"
// import "strings"
//
// < sampledata.string()
// >     |> map(fn: (r) => ({r with _value: strings.replaceAll(v: r._value, t: "p", u: "XX")}))
// ```
//
builtin replaceAll : (v: string, t: string, u: string) => string

// split splits a string on a specified separator and returns an array of substrings.
//
// ## Parameters
//
// - v: String value to split.
// - t: String value that acts as the separator.
//
// ## Examples
//
// ### Split a string into an array of substrings
// ```no_run
// import "strings"
//
// strings.split(v: "foo, bar, baz, quz", t: ", ")
// // Returns ["foo", "bar", "baz", "quz"]
// ```
//
builtin split : (v: string, t: string) => [string]

// splitAfter splits a string after a specified separator and returns an array of substrings.
// Split substrings include the separator, `t`.
//
// ## Parameters
//
// - v: String value to split.
// - t: String value that acts as the separator.
//
// ## Examples
//
// ### Split a string into an array of substrings
// ```no_run
// import "strings"
//
// strings.splitAfter(v: "foo, bar, baz, quz", t: ", ")
// // Returns ["foo, ", "bar, ", "baz, ", "quz"]
// ```
//
builtin splitAfter : (v: string, t: string) => [string]

// splitN splits a string on a specified separator and returns an array of `i` substrings.
//
// ## Parameters
//
// - v: String value to split.
// - t: String value that acts as the separator.
// - i: Maximum number of split substrings to return.
//
//      `-1` returns all matching substrings.
//       The last substring is the unsplit remainder.
//
// ## Examples
//
// ### Split a string into an array of substrings
// ```no_run
// import "strings"
//
// strings.splitN(v: "foo, bar, baz, quz", t: ", ", i: 3)
// // Returns ["foo", "bar", "baz, quz"]
// ```
//
builtin splitN : (v: string, t: string, i: int) => [string]

// splitAfterN splits a string after a specified separator and returns an array of `i` substrings.
// Split substrings include the separator, `t`.
//
// ## Parameters
//
// - v: String value to split.
// - t: String value that acts as the separator.
// - i: Maximum number of split substrings to return.
//
//     `-1` returns all matching substrings.
//     The last substring is the unsplit remainder.
//
// ## Examples
//
// ### Split a string into an array of substrings
// ```no_run
// import "strings"
//
// strings.splitAfterN(v: "foo, bar, baz, quz", t: ", ", i: 3)
// // Returns ["foo, ", "bar, ", "baz, quz"]
// ```
//
builtin splitAfterN : (v: string, t: string, i: int) => [string]

// joinStr concatenates elements of a string array into a single string using a specified separator.
//
// ## Parameters
//
// - arr: Array of strings to concatenate.
// - v: Separator to use in the concatenated value.
//
// ## Examples
//
// ### Join a list of strings into a single string
// ```no_run
// import "strings"
//
// strings.joinStr(arr: ["foo", "bar", "baz", "quz"], v: ", ")
// // Returns "foo, bar, baz, quz"
// ```
//
builtin joinStr : (arr: [string], v: string) => string

// strlen returns the length of a string. String length is determined by the number of UTF code points a string contains.
//
// ## Parameters
//
// - v: String value to measure.
//
// ## Examples
//
// ### Filter based on string value length
// ```
// # import "regexp"
// # import "sampledata"
// import "strings"
// #
// # data = sampledata.string() |> map(fn: (r) => ({r with _value: regexp.replaceAllString(r: /[sm]|\d/, v: r._value, t: "")}))
//
// < data
// >     |> filter(fn: (r) => strings.strlen(v: r._value) <= 6)
// ```
//
// ### Store the length of string values
// ```
// # import "regexp"
// # import "sampledata"
// import "strings"
// #
// # data = sampledata.string() |> map(fn: (r) => ({r with _value: regexp.replaceAllString(r: /[sm]|\d/, v: r._value, t: "")}))
//
// < data
// >  |> map(fn: (r) => ({ r with length: strings.strlen(v: r._value)}))
// ```
//
builtin strlen : (v: string) => int

// substring returns a substring based on start and end parameters. These parameters are represent indices of UTF code points in the string.
//
// ## Parameters
//
// - v: String value to search for.
// - start: Starting inclusive index of the substring.
// - end: Ending exclusive index of the substring.
//
// When start or end are past the bounds of the string, respecitvely the start or end of the string
// is assumed. When end is less than or equal to start an empty string is returned.
//
// ## Examples
//
// ### Return part of a string based on character index
// ```
// import "sampledata"
// import "strings"
//
// < sampledata.string()
// >     |> map(fn: (r) => ({r with _value: strings.substring(v: r._value, start: 5, end: 9)}))
// ```
builtin substring : (v: string, start: int, end: int) => string

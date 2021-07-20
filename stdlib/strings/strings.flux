// Package strings provides functions to manipulate UTF-8 encoded strings.
package strings


// title converts a string to title case.
//
// ## Parameters
//
// - `v` is the string value to convert.
//
// ## Convert all values of a column to title case
//
// ```
//  import "strings"
//
//  data
//      |> map(fn: (r) => ({ r with pageTitle: strings.title(v: r.pageTitle) }))
//
builtin title : (v: string) => string

// toUpper converts a string to uppercase.
//
// ## Parameters
//
// - `v` is the string value to convert.
//
// ## Convert all values of a column to upper case
//
// ```
// import "strings"
//
// data
//   |> map(fn: (r) => ({ r with envVars: strings.toUpper(v: r.envVars) }))
// ```
//
// The difference between toTitle and toUpper
//
//      - The results of toUpper() and toTitle are often the same, however the difference is visible when using special characters:
//
//      - str = "ǳ"
//
//      - strings.toUpper(v: str) // Returns Ǳ
//      - strings.toTitle(v: str) // Returns ǲ
//
builtin toUpper : (v: string) => string

// toLower converts a string to lowercase.
//
// ## Parameters
//
// - `v` is the string value to convert.
//
// ## Convert all values of a column to lower case
//
// ```
// import "strings"
//
// data
//   |> map(fn: (r) => ({
//        r with exclamation: strings.toLower(v: r.exclamation)
//      })
//    )
// ```
//
builtin toLower : (v: string) => string

// trim removes leading and trailing characters specified in the cutset from a string.
//
// ## Parameters
//
// - `v` is the string to remove characters from.
// - `cutset` is the  leading and trailing characters to remove from the string.
//
//      Only characters that match the cutset string exactly are trimmed.
//
// ## Trim leading and trailing periods from all values in a column
//
// ```
// import "strings"
//
// data
//   |> map(fn: (r) => ({
//       r with
//       variables: strings.trim(v: r.variables, cutset: ".")
//     })
//   )
//
builtin trim : (v: string, cutset: string) => string

// trimPrefix removes a prefix from a string. Strings that do not start with the prefix are returned unchanged.
//
// ## Parameters
//
// - `v` is the string to trim
// - `prefix` is the prefix to remove
//
// ## Trim leading and trailing periods from all values in a column
//
// ```
// import "strings"
//
// data
//   |> map(fn: (r) => ({
//       r with
//       sensorID: strings.trimPrefix(v: r.sensorId, prefix: "s12_")
//     })
//   )
// ```
//
builtin trimPrefix : (v: string, prefix: string) => string

// trimSpace removes leading and trailing spaces from a string.
//
// ## Parameters
//
// - `v` is the string to remove spaces from
//
// ## Trim leading and trailing spaces from all values in a column
//
// ```
// import "strings"
//
// data
//   |> map(fn: (r) => ({ r with userInput: strings.trimSpace(v: r.userInput) }))
// ```
builtin trimSpace : (v: string) => string

// The trimSuffix removes a suffix from a string. Strings that do not end with the suffix are returned unchanged.
//
// ## Parameters
//
// - `v` is the string to trim
// - `suffix` is the suffix to remove.
//
// ## Remove a suffix from all values in a column
//
// ```
// import "strings"
//
// data
//   |> map(fn: (r) => ({
//       r with
//       sensorID: strings.trimSuffix(v: r.sensorId, suffix: "_s12")
//     })
//   )
// ```
//
builtin trimSuffix : (v: string, suffix: string) => string

// trimRight removes trailing characters specified in the cutset from a string.
//
// ## Parameters
//
// - `v` is the string to to remove characters from
// - `cutset` is the trailing characters to trim from the string.
//
//      Only characters that match the cutset string exactly are trimmed.
//
// ## Trim trailing periods from all values in a column
//
// ```
// import "strings"
//
// data
//   |> map(fn: (r) => ({
//       r with
//       variables: strings.trimRight(v: r.variables, cutset: ".")
//     })
//   )
// ```
//
builtin trimRight : (v: string, cutset: string) => string

// trimLeft removes specified leading characters from a string.
//
// ## Parameters
//
// - `v` is the string to to remove characters from
// - `cutset` is the trailing characters to trim from the string.
//
// ## Trim leading periods from all values in a column
//
// ```
// import "strings"
//
// data
//   |> map(fn: (r) => ({
//       r with
//       variables: strings.trimLeft(v: r.variables, cutset: ".")
//     })
//   )
// ```
//
builtin trimLeft : (v: string, cutset: string) => string

// toTitle converts all characters in a string to title case.
//
// ## Parameters
//
// - `v` is the string value to convert.
//
// ## Convert characters in a string to title case
//
// ```
// import "strings"
//
// data
//   |> map(fn: (r) => ({ r with pageTitle: strings.toTitle(v: r.pageTitle) }))
// ```
//
builtin toTitle : (v: string) => string

// hasPrefix indicates if a string begins with a specified prefix.
//
// ## Parameters
//
// - `v` is the string value to search.
// - `prefix` is the string prefix to search for.
//
// ## Filter based on the presence of a prefix in a column value
//
// ```
// import "strings"
//
// data
//   |> filter(fn:(r) => strings.hasPrefix(v: r.metric, prefix: "int_" ))
// ```
//
builtin hasPrefix : (v: string, prefix: string) => bool

// hasSuffix indicates if a string ends with a specified suffix.
//
// ## Parameters
//
// - `v` is the string value to search.
// - `prefix` is the string suffix to search for.
//
// ## Filter based on the presence of a suffix in a column value
//
// ```
// import "strings"
//
// data
//   |> filter(fn:(r) => strings.hasSuffix(v: r.metric, suffix: "_count" ))
// ```
//
builtin hasSuffix : (v: string, suffix: string) => bool

// containsStr reports whether a string contains a specified substring.
//
// ## Parameters
//
// - `v` is the string value to search
// - `substr` is the substring value to search for
//
// ## Report if a string contains a specific substring
//
// ```
// import "strings"
//
// data
//   |> map(fn: (r) => ({
//       r with
//       _value: strings.containsStr(v: r.author, substr: "John")
//     })
//   )
// ```
//
builtin containsStr : (v: string, substr: string) => bool

// containsAny reports whether a specified string contains characters from another string.
//
// ## Parameters
//
// - `v` is the string value to search
// - `chars` is the character to search for
//
// ## Report if a string contains specific characters
//
// ```
// import "strings"
//
// data
//   |> map(fn: (r) => ({
//       r with
//       _value: strings.containsAny(v: r.price, chars: "£$¢")
//     })
//   )
// ```
//
builtin containsAny : (v: string, chars: string) => bool

// equalFold reports whether two UTF-8 strings are equal under Unicode case-folding.
//
// ## Parameters
//
// - `v` is the string value to compare
// - `t` is the string value to compare against
//
// ## Ignore case when testing if two strings are the same
//
// ```
// import "strings"
//
// data
//   |> map(fn: (r) => ({
//       r with
//       string1: r.string1,
//       string2: r.string2,
//       same: strings.equalFold(v: r.string1, t: r.string2)
//     })
//   )
// ```
//
builtin equalFold : (v: string, t: string) => bool

// compare compares the lexicographical order of two strings.
//
//      Return values
//      Comparison	Return value
//      v < t	    -1
//      v == t	    0
//      v > t	    1
//
// ## Parameters
//
// - `v` is the string value to compare
// - `t` is the string value to compare against
//
// ## Compare the lexicographical order of column values
//
// ```
// import "strings"
//
// data
//   |> map(fn: (r) => ({
//       r with
//       _value: strings.compare(v: r.tag1, t: r.tag2)
//     })
//   )
// ```
//
builtin compare : (v: string, t: string) => int

//countStr counts the number of non-overlapping instances of a substring appears in a string.
//
// ## Parameters
//
// - `v` is the string value to search
// - `substr` is the substr value to count
//
//      The function counts only non-overlapping instances of substr. For example:
//      strings.coutnStr(v: "ooooo", substr: "oo")
//      // Returns 2 -- (oo)(oo)o
//
// ## Count instances of a substring within a string
//
// ```
// import "strings"
//
// data
//   |> map(fn: (r) => ({
//       r with
//        _value: strings.countStr(v: r.message, substr: "uh")
//     })
//   )
// ```
//
builtin countStr : (v: string, substr: string) => int

// index returns the index of the first instance of a substring in a string. If the substring is not present, it returns -1.
//
// ## Parameters
//
// - `v` is the string value to search
// - `substr` is the substring to search for
//
// ## Find the first occurrence of a substring
//
// ```
// import "strings"
//
// data
//   |> map(fn: (r) => ({
//       r with
//       the_index: strings.index(v: r.pageTitle, substr: "the")
//     })
//   )
// ```
//
builtin index : (v: string, substr: string) => int

// indexAny returns the index of the first instance of specified characters in a string. If none of the specified characters are present, it returns -1.
//
// ## Parameters
//
// - `v` is the string value to search
// - `chars` are the chars to search for
//
// ## Find the first occurrence of characters from a string
//
// ```
// import "strings"
//
// data
//   |> map(fn: (r) => ({
//       r with
//       charIndex: strings.indexAny(v: r._field, chars: "_-")
//     })
//   )
// ```
//
builtin indexAny : (v: string, chars: string) => int

// lastIndex returns the index of the last instance of a substring in a string. If the substring is not present, the function returns -1.
//
// ## Parameters
//
// - `v` is the string value to search
// - `substr` is the substring to search for
//
// ## Find the last occurrence of a substring
//
// ```
// import "strings"
//
// data
//   |> map(fn: (r) => ({
//       r with
//       the_index: strings.lastIndex(v: r.pageTitle, substr: "the")
//     })
//   )
// ```
//
builtin lastIndex : (v: string, substr: string) => int

// lastIndexAny returns the index of the last instance of any specified characters in a string. If none of the specified characters are present, the function returns -1.
//
// ## Parameters
//
// - `v` is the string value to search
// - `chars` are the characters to search for
//
// ## Find the last occurrence of characters from a string
//
// ```
// import "strings"
//
// data
//   |> map(fn: (r) => ({
//       r with
//       charLastIndex: strings.lastIndexAny(v: r._field, chars: "_-")
//     })
//   )
// ```
//
builtin lastIndexAny : (v: string, chars: string) => int

// isDigit tests if a single-character string is a digit (0-9).
//
// ## Parameters
//
// - `v` is the single-character string to test.
//
// ## Filter by columns with digits as values
//
// ```
// import "strings"
//
// data
//   |> filter(fn: (r) => strings.isDigit(v: r.serverRef))
// ```
//
builtin isDigit : (v: string) => bool

// isLetter tests if a single character string is a letter (a-z, A-Z).
//
// ## Parameters
//
// - `v` is the single-character string to test.
//
// ## Filter by columns with digits as values
//
// ```
// import "strings"
//
// data
//   |> filter(fn: (r) => strings.isLetter(v: r.serverRef))
// ```
//
builtin isLetter : (v: string) => bool

// isLower tests if a single-character string is lowercase.
//
// ## Parameters
//
// - `v` is the single-character string value to test.
//
// ## Filter by columns with single-letter lowercase values
//
// ```
// import "strings"
//
// data
//   |> filter(fn: (r) => strings.isLower(v: r.host))
// ```
//
builtin isLower : (v: string) => bool

// isUpper tests if a single character string is uppercase.
//
// ## Parameters
//
// - `v` is the single-character string value to test.
//
// ## Filter by columns with single-letter uppercase values
//
// ```
// import "strings"
//
// data
//   |> filter(fn: (r) => strings.isUpper(v: r.host))
// ```
//
builtin isUpper : (v: string) => bool

// repeat returns a string consisting of i copies of a specified string.
//
// ## Parameters
//
// - `v` is the string value to repeat.
// - `i` is the number of times to repeat v.
//
// ## Repeat a string based on existing columns
//
// ```
// import "strings"
//
// data
//   |> map(fn: (r) => ({
//       laugh: r.laugh
//       intensity: r.intensity
//       laughter: strings.repeat(v: r.laugh, i: r.intensity)
//     })
//   )
// ```
//
builtin repeat : (v: string, i: int) => string

// replace replaces the first i non-overlapping instances of a substring with a specified replacement.
//
// ## Parameters
//
// - `v` is the string value to search.
// - `t` is the substring value to replace.
// - `u` is the replacement for i instances of t.
// - `i` is the number of non-overlapping t matches to replace.
//
// ## Replace a specific number of string matches
//
// ```
// import "strings"
//
// data
//   |> map(fn: (r) => ({
//       r with
//       content: strings.replace(v: r.content, t: "he", u: "her", i: 3)
//     })
//   )
// ```
//
builtin replace : (v: string, t: string, u: string, i: int) => string

// replaceAll replaces all non-overlapping instances of a substring with a specified replacement.
//
// ## Parameters
//
// - `v` is the string value to search.
// - `t` is the substring to replace.
// - `u` is the replacement for all instances of t.
//
// ## Replace string matches
//
// ```
// import "strings"
//
// data
//   |> map(fn: (r) => ({
//       r with
//       content: strings.replaceAll(v: r.content, t: "he", u: "her")
//     })
//   )
// ```
//
builtin replaceAll : (v: string, t: string, u: string) => string

// split splits a string on a specified separator and returns an array of substrings.
//
// ## Parameters
//
// - `v` is the string value to split.
// - `t` is the string value that acts as the separator.
//
// ## Split a string into an array of substrings
//
// ```
// import "strings"
//
// data
//   |> map (fn:(r) => strings.split(v: r.searchTags, t: ","))
// ```
//
builtin split : (v: string, t: string) => [string]

// splitAfter splits a string after a specified separator and returns an array of substrings. Split substrings include the separator, t.
//
// ## Parameters
//
// - `v` is the string value to split.
// - `t` is the string value that acts as the separator.
//
// ## Split a string into an array of substrings
//
// ```
// import "strings"
//
// data
//    |> map (fn:(r) => strings.splitAfter(v: r.searchTags, t: ","))
// ```
//
builtin splitAfter : (v: string, t: string) => [string]

// splitN splits a string on a specified separator and returns an array of i substrings.
//
// ## Parameters
//
// - `v` is the string value to split.
// - `t` is the string value that acts as the separator.
// - `i` is the maximum number of split substrings to return. -1 returns all matching substrings.
//
//       - The last substring is the unsplit remainder.
//
// ## Split a string into an array of substrings
//
// ```
// import "strings"
//
// data
//    |> map (fn:(r) => strings.splitN(v: r.searchTags, t: ","))
// ```
//
builtin splitN : (v: string, t: string, n: int) => [string]

// splitAfterN splits a string after a specified separator and returns an array of i substrings. Split substrings include the separator t.
//
// ## Parameters
//
// - `v` is the string value to split.
// - `t` is the string value that acts as the separator.
// - `i` is the maximum number of split substrings to return. -1 returns all matching substrings.
//
//       - The last substring is the unsplit remainder.
//
// ## Split a string into an array of substrings
//
// ```
// import "strings"
//
// data
//    |> map (fn:(r) => strings.splitAfterN(v: r.searchTags, t: ","))
// ```
//
builtin splitAfterN : (v: string, t: string, i: int) => [string]

// joinStr concatenates elements of a string array into a single string using a specified separator.
//
// ## Parameters
//
// - `arr` is the array of strings to concatenate.
// - `t` is the separator to use in the concatenated value.
//
// ## Join a list of strings into a single string
//
// ```
// import "strings"
//
// searchTags = ["tag1", "tag2", "tag3"]
//
// strings.joinStr(arr: searchTags, v: ","))
// ```
//
builtin joinStr : (arr: [string], v: string) => string

// strlen returns the length of a string. String length is determined by the number of UTF code points a string contains.
//
// ## Parameters
//
// - `v` is the string value to measure.
//
// ## Filter based on string value length
//
// ```
// import "strings"
//
// data
//    |> filter(fn: (r) => strings.strlen(v: r._measurement) <= 4)
// ```
//
// ## Store the length of string values
//
// ```
// import "strings"
//
// data
//   |> map(fn: (r) => ({
//       r with
//       length: strings.strlen(v: r._value)
//     })
//   )
// ```
//
builtin strlen : (v: string) => int

// substring returns a substring based on start and end parameters. These parameters are represent indices of UTF code points in the string.
//
// ## Parameters
//
// - `v` is the string value to search for.
// - `start` is the starting inclusive index of the substring.
// - `end` is the ending exclusive index of the substring.
//
// ## Store the first four characters of a string
//
// ```
// import "strings"
//
// data
//   |> map(fn: (r) => ({
//       r with
//       abbr: strings.substring(v: r.name, start: 0, end: 4)
//     })
//   )
// ```
builtin substring : (v: string, start: int, end: int) => string

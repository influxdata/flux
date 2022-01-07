// Package array provides functions for building tables from Flux arrays.
//
// introduced: 0.103.0
// tags: text,tables
package textreader


// from constructs a table from an arbitrary sample of text
//
// Each line of text is converted into an output row or record.
// All lines must have the same dimensionality/types
//
// ## Parameters
// - txt: text to process
// - header: array of strings for column headers.  Default is an empty array which indicates using the first
//           of text as the header
//
// ## Examples
//
// ### Build an arbitrary table
// ```
// import "array"
//
// text = "
// a|b|c
// 1|2|3
// "
//
// > textreader.from(txt: text, header: [)
// outputs table with single record {a:1, b:2, c:3}
// ```
//
builtin from : (txt: string, parseFn: (line: string) => A ) => [A] where A: Record

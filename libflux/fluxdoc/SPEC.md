# Writing Flux Doc Comments

Flux will treat any uninterrupted series of line comments immediately preceding a function definition, package clause, builtin statement, or any declaration of an option or constant, as a doc comment.

A flux doc comment consists of multiple sections, which are defined in this document. Each section should be separated from other sections by a single blank line comment. Unless otherwise noted, each section is required, and the sections must be arranged in the order in which they are presented in this document.

## Documenting functions

### Name and Description

Doc comments for functions should begin with the function name, wrapped in backticks, followed by a brief description of what the function does or how it is used. The name and description should read as one or more full sentences.

_Example:_

```
// `join` merges two input streams into a single output stream based on columns
// with equal values.
```

### Extra Details

Some function documentation may require explanation beyond what is provided in the `Name and Description` section described above. Any markdown-formatted text between the end of the initial function description and the beginning of the parameter list will be treated as additional explanatory notes, and will be read and stored separately from the initial description in the final data structure.

Separating the main function description from this longer, more detailed explanation allows consumers who only need a condensed version of the docs (e.g. the InfluxDB UI) to do so, while other consumers are free to use the full docs.

This section may be omitted.

### Parameters

Parameters should take the form of a markdown unordered list. Each list item should start with the name of the parameter wrapped in backticks, followed by a brief, one-line description.

The parameters in the list should be ordered exactly as they are in the function signature, and no parameter should be omitted.

Each list item should read as a complete sentence.

_Example:_

```
// - `tables` is a stream of tables
// - `method` is the method to use when joining
// - `on` is a list of column names on which to join
```

### Code Examples

Code examples consist of a heading, an optional preamble, and a markdown-formatted code block.

The heading should be a markdown H1 heading, and should be immediately followed by a blank line comment.

Any text between the heading and the code block will be treated as part of the preamble. The preamble should be separated from the code block by a blank line comment. The preamble may be omitted.

Code blocks should contain valid flux code that readers can copy and paste into the InfluxDB data explorer and run with no modifications. Any sample data used in a code example should come from an `array.from()` call in the example code itself, rather than being represented as a markdown- or html-formatted table.

This section may include any number of code examples.

_Example:_

```
// # Joining two tables
// 
// Run this in the data explorer to see the results!
//
// ```
// import "array"
//
// sf_temp = array.from(
//     rows: [
//         {_time: 2021-06-01T01:00:00Z, _field: "temp", _value: 70},
//         {_time: 2021-06-01T02:00:00Z, _field: "temp", _value: 75},
//         {_time: 2021-06-01T03:00:00Z, _field: "temp", _value: 72},
//     ],
// )
//
// ny_temp = array.from(
//     rows: [
//         {_time: 2021-06-01T01:00:00Z, _field: "temp", _value: 55},
//         {_time: 2021-06-01T02:00:00Z, _field: "temp", _value: 56},
//         {_time: 2021-06-01T03:00:00Z, _field: "temp", _value: 57},
//     ],
// )
//
// join(
//   tables: {sf: sf_temp, ny: ny_temp},
//   on: ["_time", "_field"]
// )
// ```
```

### Other Information

There are other details about Flux functions that we will want to document, like the return type, the types of the parameters, and whether or not a parameter is required. We should be able to get all of that information from the function signature, and therefore shouldn't need to include it in the doc comment.

### Full Example

For some examples of what a complete doc comment looks like for a flux function, see the doc comments for `join()` and `aggregateWindow()` in the `universe` package.

## Documenting Packages

### Import path

Doc comments for packages should start with the full import path for the package, wrapped in backticks. Nothing else should be included on this line.

_Example:_

```
// `experimental/geo`
```

### Description

Include a brief description of the package. This should read as one or more full setences.

_Example:_

```
// `math`
//
// The Flux math package provides basic constants and mathematical functions.
package math
```

## Documenting Flux Options

### Name and Description

The first line of a doc comment for a flux option should start with the name of the option wrapped in backticks, followed by a brief explanation of what the option is, or how it is used.

### Extra Details

Additional notes explaining how users can use the option in their code. This section should follow the same formatting rules as the `Extra Details` section for Flux functions described earlier in this document.

This section may be omitted.

### Code Example

A markdown-formatted code block that users can copy and paste into the InfluxDB data explorer and run to see the option in action.

### Full Example

For an example of a complete doc comment for a Flux option, see the `enabledProfilers` option in the `profiler` package.

## Documenting Builtin Constants

Flux packages often come with one or more predefined constants that users can import and use in their code. Some of these constants are defined with `builtin` statements, which means their actual value is defined in Go code rather than Flux code. To document a such constant, include a line comment directly above its declaration with a string representation of the constant's value.

Because this value will only ever be treated as a string in the docs, writing a description of the constant's value, rather than the literal value, is acceptable if it makes things easier to read or understand.

If a constant is not defined with a builtin statement, the doc comment parser should be able to get its value from the flux code itself, and no documentation is required.

_Example:_

```
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
```

## Documenting Groups of Constants

Some packages contain multiple constants that are meaningfully related to one another. In these cases, it may be helpful to document them as a group, rather than document each one of them separately.

A doc comment for a group of constants consists of a heading and a description. The heading should be a markdown H1 heading, and should be immediately followed by a blank line comment. The description should be a single line of text explaining what the constants represent and/or how they are related.

There should be no blank lines or lines containing code between the declaration/definition of one constant in a group and the declaration of the next constant in the same group.

If one of the constants in a group is defined using a `builtin` statement, a doc comment should be added on the line immediately above its declaration (See the `Documenting Builtin Constants` section above).

The end of a constant group is deliniated by a blank, uncommented line.

_Example:_

```
// # Days of the week
//
// The days of the week are represented as integers in the range `[0-6]`
Sunday = 0
Monday = 1
Tuesday = 2
Wednesday = 3
Thursday = 4
Friday = 5
Saturday = 6

// # Months of the year
//
// Months are represented as integers in the range `[1-12]`
January = 1
February = 2
March = 3
April = 4
May = 5
June = 6
July = 7
August = 8
September = 9
October = 10
November = 11
December = 12
```

## Docs as Rust Data Structures

The overarching goal for this project is to be able to parse the new docs into rust structures, which we will then serialize into JSON for broader consumption.

Here's a proprosal for what those rust data structures could look like:

```rust
enum Doc {
	Package(Box<PackageDoc>),
	Function(Box<FunctionDoc>),
	Opt(Box<OptionDoc>),
	Constant(Box<ConstDoc>),
	ConstGroup(Box<ConstGroupDoc>),
}

struct PackageDoc {
	name: String,
	path: String,
	desc: String,
	members: HashMap<String, Doc>
}

struct OptionDoc {
	name: String,
	notes: Option<String>,
	examples: String,
}

struct ConstDoc {
	name: String,
	value: String,
}

struct ConstGroupDoc {
	head: String,
	desc: String,
	consts: Vec<ConstDoc>,
}

struct FunctionDoc {
	name: String,
	description: String,
	notes: Option<String>,
	parameters: Vec<Parameter>,
	examples: Vec<CodeExample>,
	return_type: String,
}

struct Parameter {
	name: String,
	description: String,
	flux_type: String,
	required: bool,
}

struct CodeExample {
	heading: String,
	preamble: Option<String>,
	code: String,
}
```

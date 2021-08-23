# Flux doc comments

## Project background and goals

In order to establish a single source for Flux documentation,
Flux will soon support doc comments in `stdlib`. Work is underway to integrate the
documentation generated from doc comments into the `flux` crate, and expose them
as JSON data for broader consumption.

The end goals of this project are as follows:

1. Continuously populate the Flux documentation search bar in the InfluxDB data
   explorer with the latest version of the docs.
2. Expose documentation via the the Flux Language Server Protocol.
3. Use the docs to generate a static site that will become the new home of the official Flux documentation.

This document defines how Flux doc comments should be written and formatted.

## Writing Flux doc comments

A Flux doc comment is any uninterrupted series of line comments immediately preceding
one of the following:

- [package clause](#packages)
- [package value (assignment or builtin statement)](#package-values)
- [function declaration](#functions)
- [option declaration](#options)

Flux doc comments support Markdown formatting in accordance with the
[CommonMark specification](https://spec.commonmark.org/0.29/).

### Headlines and descriptions

Flux docs need to support different formats for different consumers.
For example, the Flux LSP and the InfluxDB UI will need to produce a condensed
version of the documentation for any given identifier in the Flux standard library.
In contrast, the docs site will need to present more detailed information
to readers. Consumers should be able to decide which format they want to use by
selecting only the parts of the docs that are required for their use case.

To allow for this, Flux doc comments consist of a _headline_ and a _description_.
A description may contain one or more _code examples_.

### Packages

#### Headline

This is a one-sentence description of the package. The beginning of the sentence
should read `Package <package-name> provides...`, and should be on its own line.

There must be a blank line comment between the headline and description.

#### Description

This section provides a detailed description of the package and its
contents.

If the package provides any constants that require documentation, that documentation
should be included here. Otherwise, constants are listed by name in the
generated package documentation, with no further elaboration.

Code examples should be contained in a Markdown-formatted code block, and may be
included anywhere in the description. Readers should be able to copy and paste any
any example code into the InfluxDB data explorer and run it with no modifications.
Any sample data used in a code example should come from an `array.from()` call in
the example code itself, rather than being represented as a Markdown- or html-formatted table.

> Future iterations on this project may add extra functionality around code examples
> (for instance, allowing users to run them interactively on the docs site). For now,
> they will just be treated as plain Markdown.

### Package values

Packages expose values via assignment statements and builtin statements.
These values should be documented with a headline and optional description.

#### Headline

Documentation for values should begin with the headline: a one-line description
of the value that begins with the value name. The value name should be
written exactly as it appears in the assignment statement, and should be followed
by a brief explanation of the purpose of the value or how it can be used. The entire
headline should read as one complete sentence.

Follow the headline with a blank line comment.

_Example:_

Here's what the doc comment might look like for the pi constant in the `math` package.

```
// pi is a floating point representation of the mathematical constant pi.
builtin pi : float
```

#### Description

Package value descriptions are the same as [package descriptions](#description) above.

_Example:_

Here's what the doc comment might look like for a fictional constant `fortyTwo` from the math package.

```
// fortyTwo is the integer 42.
//
// The constant fortyTwo can be used to deteremine the answer to important questions.
//
// ```
// import "math"
//
// question = (answer) => answer
// question(answer: math.fortyTwo)
// ```
fortyTwo = 42
```

### Functions

When a package exposes a value that is a function we want to include additional information
about the function's parameters. Documentation for Flux functions consists of a short headline, a detailed description,
and a list of the function's parameters.

#### Headline

Documentation for functions should begin with the headline: a one-line description
of the function that begins with the function name. The function name should be
written exactly as it appears in the function signature, and should be followed
by a brief explanation of what the function does or how it can be used. The entire
headline should read as one complete sentence.

Follow the headline with a blank line comment.

_Example:_

```
// join merges two input streams based on columns with equal values.
```

#### Description

Function descriptions are formatted the same as [package descriptions](#description) above,
with the exception that every function description should include documentation
for the function's parameters.

Every function description should include at least one code example.

#### Parameters

Documentation for function parameters is required, but it can appear anywhere in
the function description.

The parameters should be organized into a Markdown-formatted unordered list, and
should be immediately preceded by a Markdown H2 header, `## Parameters`.
Each list item should start with the name of the parameter as it appears in the
function signature, followed by a brief one-line description. Each list item
should read as a complete sentence, and should be properly punctuated.

The parameters in the list should be ordered exactly as they are in the function
signature. No parameter should be omitted.

While the top-level list item for each parameter is limited to a short, one-line
description, the parameter list may include extra Markdown content after each list item.
Much like the distinction between a headline and a full description outlined
earlier in this document, it can be helpful to think of the top-level list item
as a condensed description fit for short-form docs, and the notes that come after
each list item as a more detailed explanation for use in long-form docs.

Only top-level list items are required in the parameter list. Extra notes
for each parameter are optional, and may be omitted.

See the [`lists`](https://spec.commonmark.org/0.29/#lists) section in the CommonMark spec for specifics on supported formatting in the parameter list.

_Example:_

Here's what a parameter list might look like for the `aggregateWindow` function.
```
// ## Parameters
// - every is the duration of each window.
// - fn is the aggregate function to be used in the operation.
//
//        Acceptable arguments for `fn` are: "min", "max", "mean",
//        "sum", "count", "first", and "last"
//
// - offset is the offset of each window.
// - column is the column on which to operate.
//
//        If no argument is provided for `column`, it will
//        default to `_value`.
//
// - timeSrc is the time column from which time is copied for the aggregate record.
//
//        If no argument is provided for `timeSrc` it will
//        default to `_stop`.
//
// - timeDst is the column to which time is copied for the aggregate record.
//
//        If no argument is provided for `timeDst`, it will
//        default to `_time`.
//
// - createEmpty decides if windows with no data should be included in the final output.
//
//        If no argument is provided for `createEmpty`, it will
//        default to `true`.
//
// - tables is a stream of input tables.
```

#### Other information

There are other details about Flux functions that we will want to document, like
the full type signature of each function, and which of the parameters are required.
That information can be found in the function signature, and does not need be
included in the doc comment.

#### Example function documentation

Here's an example of a full doc comment for the `join` function.

```
// join merges two input streams based on columns with equal values.
//
// Null values are not considered equal when comparing column
// values. The resulting schema is the union of the input schemas. The resulting
// group key is the union of the input group keys.
//
// ## Parameters
// - `tables` is a stream of tables
// - `method` is the method to use when joining
//
//    Currently, this function only supports inner joins.
//    Performaing an inner join will require both input
//    tables to match their columns based on the `on` parameter.
//
// - `on` is a list of column names on which to join.
//
// ## Joining two tables
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
//
// ## Output schema of a joined table
//
// The column schema of the output stream is the union
// of the input schemas. It is also the same for the
// output group key. Columns are renamed using the pattern
// `<column>_<table>` to prevent ambiguity in joined tables.
//
// ```
// import "array"
//
// data_1 = array.from(
//     rows: [
//         {_time: 2021-06-01T01:00:00Z, _field: "meter", _value: 100},
//         {_time: 2021-06-01T02:00:00Z, _field: "meter", _value: 200},
//         {_time: 2021-06-01T03:00:00Z, _field: "meter", _value: 300},
//     ],
// ) |> group(columns: ["_time", "_field"])
//
// data_2 = array.from(
//     rows: [
//         {_time: 2021-06-01T01:00:00Z, _field: "meter", _value: 400},
//         {_time: 2021-06-01T02:00:00Z, _field: "meter", _value: 500},
//         {_time: 2021-06-01T03:00:00Z, _field: "meter", _value: 600},
//     ],
// ) |> group(columns: ["_time", "_field"])
//
// join(tables: {d1: data_1, d2: data_2}, on: ["_time"]) // group key should be [_time, _field_d1, _field_d2]
// ```
```

### Options

#### Headline

The first line of a doc comment for a flux option should start with the name of
the option, followed by a brief explanation of what the option is, or how it is
used.

Follow the headline with a blank line comment.

#### Description

This section contains a detailed explanation of the how the option can be used
in a Flux query, and should follow the same formatting rules as the 
[package descriptions](#description) above.

#### Example

Here's what a doc comment might look like for the `enabledProfilers` option in
the `profiler` package.
```
// enabledProfilers sets a list of profilers that should be enabled during execution.
//
// There are two profilers available: the query profiler and the operator profiler.
//
// - The query profiler measures time spent in various phases of query execution
// - The operator profiler measures time spent in each operator of the query
//
// ## Enabling the profilers
//
// Add the following lines to your flux query to see profiler results in the output.
//
// ```
// import "profiler"
// option profiler.enabledProfilers = ["query", "operator"]
// ```
option enabledProfilers = [""]
```

### Doc comments as Rust types

<!-- How is this goal related to the three goals above? -->
Another goal for this project is to parse doc comments into Rust types
that we can then serialize into JSON for broader consumption.

Below is a proposal for how to define those data structures.

```rust
// Doc represents a documentation for Flux source code.
enum Doc {
// Package represents documentation for an entire Flux package.
	Package(Box<PackageDoc>),
    // Value represents documentation for a value exposed from a package.
	Value(Box<ValueDoc>),
    // Builtin represents documentation for a builtin value exposed from a package.
	Builtin(Box<ValueDoc>),
    // Option represents documentation for a option value exposed from a package.
	Option(Box<ValueDoc>),
    // Function represents documentation for a function value exposed from a package.
	Function(Box<FunctionDoc>),
}

struct PackageDoc { 
    path: String,
	name: String,
	headline: String,
	description: Option<String>,
	members: HashMap<String, Doc>,
    link: String,
}

struct ValueDoc {
	name: String,
	headline: String,
	description: Option<String>,
	flux_type: String,
    link: String,
}

struct FunctionDoc {
	name: String,
	headline: String,
	description: String,
	parameters: Vec<ParameterDoc>,
	flux_type: String,
    link: String,
}

struct ParameterDoc {
	name: String,
	headline: String,
	description: Option<String>,
	required: bool,
}
```

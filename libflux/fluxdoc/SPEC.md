# Writing Flux Doc Comments

Flux will treat any uninterrupted series of line comments immediately preceding a function, package clause, builtin statement, or option declaration, as a doc comment. There should be no newlines separating the doc comment and the item it is documenting.

Flux doc comments are broken up into multiple sections, which are defined in this document. Each section should be separated from other sections by a single blank line comment. Not all of the sections are required, but they should come in the order in which they are presented in this document.

## Documenting functions

### Name and description

Doc comments for functions should begin with the function name, wrapped in backticks, followed by a brief description of what the function does and how it should be used.

_Example:_

```
// `join` merges two input streams into a single output stream based on columns
// with equal values. Null values are not considered equal when comparing column
// values. The resulting schema is the union of the input schemas. The resulting
// group key is the union of the input group keys.
```

### Details

Some function documentation may require further explanation. Any sentences or paragraphs between the first paragraph and the list of parameters are considered to be additional explantory notes, and will be treated as a separate value in the final data structure.

### Parameters

Parameters should be in the form of a markdown list. Each list item should start with the name of the formatter wrapped in backticks, followed by a brief description. Each list item should read as a complete sentence.

_Example:_

```
// `join` merges two input streams into a single output stream based on columns
// with equal values. Null values are not considered equal when comparing column
// values. The resulting schema is the union of the input schemas. The resulting
// group key is the union of the input group keys.
// 
// - `tables` is a stream of tables
// - `method` is the method to use when joining (defaults to 'inner')
// - `on` is a list of strings representing the names of columns on which to join
```

### Code Examples

Any markdown-formatted code block in this section will be treated as a code example. Examples should be executable flux code that readers can copy and paste into the data explorer and run on their own. Any sample data used in the examples should come from a call to `array.from()`.

At some point in the future, we may allow users to run these examples interactively on the new docs site.

Examples should not contain html- or markdown-formatted tables.

Examples may be preceded by an optional heading and brief explanation.

The heading should be in bold text (using markdown syntax) and on its own line.

An explanation should come in the form of markdown formatted text. It should come between the heading and the code block, and should be separated from both of them by a blank line comment.

The examples section may include any number of code examples.

_Example:_

```
// *Joining two tables*
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

### Full Example

For an example of what a complete doc comment looks like for a flux function, see the doc comment for `join()` in the `stdlib/universe` package.

## Documenting Packages

### Import path

Doc comments for packages should start with the full import path for the package, wrapped in double quotes.

_Example:_

```
// "experimental/geo"
```

### Description

Include a brief description of the package and and what it does.

_Example:_

```
// "math"
//
// The Flux math package provides basic constants and mathematical functions.
```

## Documenting Flux Options

### Name and description

The first line of a doc comment for a flux option should start with the name of the option wrapped in backticks, followed by a brief description of what the option is/does.

### Details

Additional notes explaining how users can use the option in their code.

### Code Example

A markdown-formatted code block that users can copy and paste into the InfluxDB data explorer and run to see the option in action.

## Documenting Builtin Values

Some Flux packages come with a set of predefined constants/variables that can be imported by users. To document one of these values, include a comment on the line above it that includes the full name that users will write when invoking that value (i.e., `<package-name>.<variable-name>`), follwed by `=`, followed by the actual value the variable represents.

_Example:_

```
// math.pi = 3.14159265358979323846264338327950288419716939937510582097494459
builtin pi : float

// math.e = 2.71828182845904523536028747135266249775724709369995957496696763
builtin e : float

// math.phi = 1.61803398874989484820458683436563811772030917980576286213544862
builtin phi : float

// math.sqrt2 = 1.41421356237309504880168872420969807856967187537694807317667974
builtin sqrt2 : float

// math.sqrte = 1.64872127070012814684865078781416357165377610071014801157507931
builtin sqrte : float
```

## Docs as Rust Data Structures

The overarching goal for this project is to be able to parse the new docs into rust structures, which we will then serialize into JSON for broader consumption.

Here's a proprosal for what those rust data structures could look like:

```rust
enum Doc {
	Function(Box<FunctionDoc>),
	Package(Box<PackageDoc>),
	Value(Box<ValueDoc>),
}

struct FunctionDoc {
	name: String,
	description: String,
	parameters: Vec<Parameter>,
	return_type: String,
	examples: Vec<CodeExample>,
	notes: String,
}

struct Parameter {
	name: String,
	description: String,
	flux_type: String,
	required: bool,
}

struct CodeExample {
	heading: Option<String>,
	explanation: Option<String>,
	code: String,
}

struct DocPackage {
	name: String,
	path: String,
	description: String,
	members: Vec<Doc>,
}
```

## Docs as JSON

The above data structures will be serialized to JSON data. Here is a rough outline of what that JSON data will look like:

```
"packages": [
	"math": {
		"functions": [
			"abs": {
				"description": "`abs` returns x as a positive value.",
				"notes": "",
				"parameters": [
					"x": {
						"type": "float",
						"required": true,
						"description": "`x` is the only argument for this function"
					}
				]
			},
			*...More functions from the math package*
		],
		"options": [],
		"values":
]
```

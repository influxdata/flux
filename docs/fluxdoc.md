# fluxdoc formatting

`fluxdoc` uses comments in `.flux` package files in `/stdlib` to generate and
output standard library documentation in JSON format.
The generated JSON is used to build the public-facing Flux standard library
documentation and ensure documentation is up-to-date and featur-complete with
each new Flux release.

## Syntax and structure
Each `.flux` package file in `/stdlib` should include comments using the
following syntax and structure.
Flux comment lines begin with `//`.

### Use Markdown
Inline Flux documentation uses Markdown or, more specifically,
[CommonMark)](https://spec.commonmark.org/).
For consistency, use the the following Markdown conventions:

- `#` header syntax
- `-` character for lists
- Fenced codeblocks

### Package summary documentation
Add package summary documentation before the package statement in the `.flux` package file.
Package summary documentation consists of a **headline**, **description**, 
**examples**, and **metadata** to be used by clients when consuming the JSON `fluxdoc` output.

- **headline**: _First paragraph_ of the package documentation that describes
  what the package does. Must begin with `Package <pkg-name>`.
- **description** _(Optional)_: All paragraphs between the first paragraph and
  optional metadata. Provides additional details about the package.
- **examples** _(Optional)_: _See [Package and function examples](#package-and-function-examples)._
- **metadata** _(Optional)_: Metadata that provides helpful information about
  the package. _See [Package metadata](#package-metadata)_.

#### Package metadata
Package metadata are string key-value pairs separated by `:`.
Each key-value pair must be on a single line.

- **introduced**: Flux version the package was added _(Strongly encouraged)_.
- **deprecated**: Flux version the package was deprecated.
- **tags**: Comma-separated list of tags used to associate related documentation
  and categorize packages. _See [Metadata tags](#metadata-tags)._
- **contributors**: Contributor GitHub usernames or other contact information.

When adding a new export to a Flux package the `introduced: NEXT` metadata can be added to the value's docs.
The release process will automatically replace the `NEXT` with the version of the Flux release.

```js
// Package examplePkg provides functions that do x and y.
//
// Package description with additional details not provided in the headline.
//
// ## Examples
// ```
// import "examplePkg"
//
// examplePkg.foo()
// ```
//
// introduced: 0.123.0
// tags: tag1,tag2,tag3
// contributors: [@someuser](https://github.com/someuser) (GitHub)
package examplePkg
```

### Function documentation
Add function documentation before the function definition in the `.flux` package file.
Function documentation consists of a **headline**, **description**,
**parameters**, **examples**, and **metadata** to be used by clients when
consuming the JSON `fluxdoc` output.

- **headline**: _First paragraph_ of the function documentation that describes
  what the function does. Must begin with the _function name_ (case sensitive).
- **description** _(Optional)_: All paragraphs between the first paragraph and
  parameters. Provides additional details about the function.
- **parameters**: _See [Function parameters](#function-parameters)._
- **examples**: _(Optional)_: _See [Package and function examples](#package-and-function-examples)._
- **metadata** _(Optional)_: Metadata that provides helpful information about
  the package. _See [Function metadata](#function-metadata)._

#### Function parameters
Identify the beginning of the parameter list with the `## Parameters` header.
List parameters as a markdown unordered list.
Each list item must follow these conventions:

- Use the `-` character when formatting the list.
- Begin each item with the parameter name (case-sensitive) followed by a colon (`:`).
- Provide a description of the parameter after the parameter name and colon.
      
##### Parameter description guidelines
- The first paragraph of the parameter description is used as the short description.
- The first paragraph and all subsequent content are used as the long description.
- Parameter descriptions can contain any valid markdown.
  If there are multiple paragraphs, lists, or other elements that need to be
  included in the description, indent them them under the parameter list item to
  nest them as part of the parameter description.
- Avoid starting parameter descriptions with an article (the, a, an).
  For example: 
  
    ```md
    <!-- Not recommended -->
    - param: The value to multiply.
    - param: A value to multiply.
    
    <!-- Recommended -->
    - param: Value to multiply.
    ```
- If a parameter has a default value, specify the default in the description
  with "Default is `defaultValue`."

#### Package and function examples
Identify the beginning of the examples list with the `## Examples` header.
Identify each example with a descriptive title formatted as an h3 header
(`### Example descriptive title`).

##### Example guidelines
- In example titles, use imperative voice and avoid gerunds (verbs ending in
  "ing" and used as a noun). For example, use "Filter by tag" instead of
  "Filtering by tag."
- Use fenced code blocks to identify the example code.
- If an example uses a syntax other than Flux, include the language identifier
  with the fenced codeblock. If the Flux example should not be executed, use
  `no_run` as the language identifier.
- Examples should pass the `flux fmt` formatting check.

##### Example execution, input, and output
Each example, if possible, should be able to be run as a standalone script.
This allows Flux to actually execute the examples to both test the validity of 
the example and provide actual example input and output.
Use the `array`, `csv`, or `sampledata` packages to include data as part of the
examples.

Use the following conventions to control example execution, input, and output.

- `# ` omits the line from the rendered example, but keeps the line during example execution.
- `< ` at the beginning of a line appends a `yield` to the end of the line to specify input.
  The `input` yield is parsed into Markdown tables and included with the JSON documentation output.
- `> ` at the beginning of a line appends a `yield` to the ind of the line to specify output.
  The `output` yield is parsed into Markdown tables and included with the JSON documentation output.
- `#<` omits the line from the rendered example and marks the line as the input data.
- `#>` omits the line from the rendered example and marks the line as the output data.
- To skip example execution, use the `no_run` language identifier on the code block.

```js
// ## Examples
// 
// ### Filter by tag value
// ```
// # import "sampledata"
//
// < sampledata.float()
// >    |> filter(fn: (r) => r.tag == "t1")
// ```
``` 

#### Function metadata
Function metadata are string key-value pairs separated by `:`.
Each key-value pair must be on a single line.

- **introduced**: Flux version the function was added _(if different than the package)_.
- **deprecated**: Flux version the function was deprecated _(if different than the package)_.
- **tags**: Comma-separated list of tags used to associate related documentation
  and categorize functions. _See [Metadata tags](#metadata-tags)._

### Metadata tags
While tags are somewhat arbitrary, some are used to categorize functions.
Use the following tags to categorize functions based on their usage and functionality:

- **aggregates**: Add to aggregate functions (functions that aggregate all rows
  in a table into a single row).
- **date/time**: Add to date/time-related functions.
- **dynamic queries**: Add to functions that convert streams of tables into
  another composite or basic type.
- **filters**: Add to functions that filter data.
- **geotemporal**: Add to geotemporal-related functions.
- **GIS**: Add to GIS-related functions.
- **inputs**: Add to functions the retrieve data from a data source.
- **metadata**: Add to functions that return metadata from input tables or a
  data source.
- **notification endpoints**: Add to notification endpoint functions.
- **outputs**: Add to functions to output data to a data source.
- **sample data**: Add to functions the provide sample data.
- **selectors**: Add to selector functions (functions that select rows from each
  input table).
- **single notification**: Add to functions that send a single notification.
- **tests**: Add to functions that perform tests.
- **transformations**: Add to transformations (functions that take a stream of
  tables as input and output a stream of tables).
- **type-conversions**: Add to functions that change data types.

## Full package documentation example
```js
// Package pkgName provides functions that do x and y.
//
// ## Examples
// ```
// import "pkgName"
//
// option pkgName.foo == "bar'
// ```
//
// introduced: 0.140.0
// contributors: [@username](https://github.com/username/) (GitHub)
package pkgName

// myFn multiplies `x` by `y`.
//
// ## Parameters
// - x: Left operand.
// - y: Right operand.
//
// ## Examples
// 
// ### Multiply x and y
// ```no_run
// import "pkgName"
// 
// pkgName.myFn(x: 2, y: 4)
// // Returns 8
// ```
//
myFn = (x, y) => x * y

// anotherFn drops columns from input tables.
//
// ## Parameters
// - columns: List of columns to drop.
// - tables: Input data. Default is piped-forward data.
//
// ## Examples
//
// ### Drop specific columns
// ```
// import "sampledata"
// import "pkgName"
//
// < sampledata.float()
// >     |> pkgName.anotherFn(columns: [tag])
// ```
// 
// introduced: 0.141.0
// tags: transformations
anotherFn = (columns, tables=<-) => tables |> drop(columns: columns)
```

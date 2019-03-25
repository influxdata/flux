# Flux Specification

The following document specifies the Flux language and query execution.

This document is a living document and does not represent the current implementation of Flux.
Any section that is not currently implemented is commented with a IMPL#XXX where XXX is an issue number tracking discussion and progress towards implementation.

## Language

The Flux language is centered on querying and manipulating time series data.

### Notation
The syntax of the language is specified using Extended Backus-Naur Form (EBNF):

    Production  = production_name "=" [ Expression ] "." .
    Expression  = Alternative { "|" Alternative } .
    Alternative = Term { Term } .
    Term        = production_name | token [ "…" token ] | Group | Option | Repetition .
    Group       = "(" Expression ")" .
    Option      = "[" Expression "]" .
    Repetition  = "{" Expression "}" .

Productions are expressions constructed from terms and the following operators, in increasing precedence:

    |   alternation
    ()  grouping
    []  option (0 or 1 times)
    {}  repetition (0 to n times)

Lower-case production names are used to identify lexical tokens.
Non-terminals are in CamelCase.
Lexical tokens are enclosed in double quotes "" or back quotes \`\`.

### Representation

Source code is encoded in UTF-8.
The text need not be canonicalized.

#### Characters

This document will use the term _character_ to refer to a Unicode code point.

The following terms are used to denote specific Unicode character classes:

    newline        = /* the Unicode code point U+000A */ .
    unicode_char   = /* an arbitrary Unicode code point except newline */ .
    unicode_letter = /* a Unicode code point classified as "Letter" */ .
    unicode_digit  = /* a Unicode code point classified as "Number, decimal digit" */ .

In The Unicode Standard 8.0, Section 4.5 "General Category" defines a set of character categories.
Flux treats all characters in any of the Letter categories Lu, Ll, Lt, Lm, or Lo as Unicode letters, and those in the Number category Nd as Unicode digits.

#### Letters and digits

The underscore character _ (U+005F) is considered a letter.

    letter        = unicode_letter | "_" .
    decimal_digit = "0" … "9" .

### Lexical Elements

#### Comments

Comment serve as documentation.
Comments begin with the character sequence `//` and stop at the end of the line.

Comments cannot start inside string or regexp literals.
Comments act like newlines.

#### Tokens

Flux is built up from tokens.
There are several classes of tokens: _identifiers_, _keywords_, _operators_, and _literals_.
_White space_, formed from spaces, horizontal tabs, carriage returns, and newlines, is ignored except as it separates tokens that would otherwise combine into a single token.
While breaking the input into tokens, the next token is the longest sequence of characters that form a valid token.

#### Identifiers

Identifiers name entities within a program.
An identifier is a sequence of one or more letters and digits.
An identifier must start with a letter.

    identifier = letter { letter | unicode_digit } .

Examples:

    a
    _x
    longIdentifierName
    αβ

#### Keywords

The following keywords are reserved and may not be used as identifiers:

    and    import  not  return   option   test
    empty  in      or   package  builtin

[IMPL#256](https://github.com/influxdata/platform/issues/256) Add in and empty operator support   

#### Operators

The following character sequences represent operators:

    +   ==   !=   (   )   =>
    -   <    !~   [   ]
    *   >    =~   {   }
    /   <=   =    ,   :
    %   >=   <-   .   |>

#### Numeric literals

Numeric literals may be integers or floating point values.
Literals have arbitrary precision and will be coerced to a specific type when used.

The following coercion rules apply to numeric literals:

* an integer literal can be coerced to an "int", "uint", or "float" type,
* an float literal can be coerced to a "float" type,
* an error will occur if the coerced type cannot represent the literal value.


[IMPL#255](https://github.com/influxdata/platform/issues/255) Allow numeric literal coercion.

##### Integer literals

An integer literal is a sequence of digits representing an integer value.
Only decimal integers are supported.

    int_lit     = "0" | decimal_lit .
    decimal_lit = ( "1" … "9" ) { decimal_digit } .

Examples:

    0
    42
    317316873

##### Floating-point literals

A floating-point literal is a decimal representation of a floating-point value.
It has an integer part, a decimal point, and a fractional part.
The integer and fractional part comprise decimal digits.
One of the integer part or the fractional part may be elided.

    float_lit = decimals "." [ decimals ]
              | "." decimals .
    decimals  = decimal_digit { decimal_digit } .

Examples:

    0.
    72.40
    072.40  // == 72.40
    2.71828
    .26

[IMPL#254](https://github.com/influxdata/platform/issues/254) Parse float literals

#### Duration literals

A duration literal is a representation of a length of time.
It has an integer part and a duration unit part.
Multiple durations may be specified together and the resulting duration is the sum of each smaller part.
When several durations are specified together, larger units must appear before smaller ones, and there can be no repeated units.

    duration_lit  = { int_lit duration_unit } .
    duration_unit = "y" | "mo" | "w" | "d" | "h" | "m" | "s" | "ms" | "us" | "µs" | "ns" .

| Units    | Meaning                                 |
| -----    | -------                                 |
| y        | year (12 months)                        |
| mo       | month                                   |
| w        | week (7 days)                           |
| d        | day                                     |
| h        | hour (60 minutes)                       |
| m        | minute (60 seconds)                     |
| s        | second                                  |
| ms       | milliseconds (1 thousandth of a second) |
| us or µs | microseconds (1 millionth of a second)  |
| ns       | nanoseconds (1 billionth of a second)   |

Durations represent a length of time.
Lengths of time are dependent on specific instants in time they occur and as such, durations do not represent a fixed amount of time.
No amount of seconds is equal to a day, as days vary in their number of seconds.
No amount of days is equal to a month, as months vary in their number of days.
A duration consists of three basic time units: seconds, days and months.

Durations can be combined via addition and subtraction.
Durations can be multiplied by an integer value.
These operations are performed on each time unit independently.

Examples:

    1s
    10d
    1h15m // 1 hour and 15 minutes
    5w
    1mo5d // 1 month and 5 days

Durations can be added to date times to produce a new date time.

Addition and subtraction of durations to date times do not commute and are left associative.
Addition and subtraction of durations to date times applies months, days and seconds in that order.
When months are added to a date times and the resulting date is past the end of the month, the day is rolled back to the last day of the month.


Examples:

    2018-01-01T00:00:00Z + 1d       // 2018-01-02T00:00:00Z
    2018-01-01T00:00:00Z + 1mo      // 2018-02-01T00:00:00Z
    2018-01-01T00:00:00Z + 2mo      // 2018-03-01T00:00:00Z
    2018-01-31T00:00:00Z + 2mo      // 2018-03-31T00:00:00Z
    2018-02-28T00:00:00Z + 2mo      // 2018-04-28T00:00:00Z
    2018-01-31T00:00:00Z + 1mo      // 2018-02-28T00:00:00Z, February 31th is rolled back to the last day of the month, February 28th in 2018.

    // Addition and subtraction of durations to date times does not commute
    2018-02-28T00:00:00Z + 1mo + 1d // 2018-03-29T00:00:00Z
    2018-02-28T00:00:00Z + 1d + 1mo // 2018-04-01T00:00:00Z
    2018-01-01T00:00:00Z + 2mo - 1d // 2018-02-28T00:00:00Z
    2018-01-01T00:00:00Z - 1d + 3mo // 2018-03-31T00:00:00Z

    // Addition and subtraction of durations to date times applies months, days and seconds in that order.
    2018-01-28T00:00:00Z + 1mo + 2d // 2018-03-02T00:00:00Z
    2018-01-28T00:00:00Z + 1mo2d    // 2018-03-02T00:00:00Z
    2018-01-28T00:00:00Z + 2d + 1mo // 2018-02-28T00:00:00Z, explicit left associative add of 2d first changes the result
    2018-02-01T00:00:00Z + 2mo2d    // 2018-04-03T00:00:00Z
    2018-01-01T00:00:00Z + 1mo30d   // 2018-03-02T00:00:00Z, Months are applied first to get February 1st, then days are added resulting in March 2 in 2018.
    2018-01-31T00:00:00Z + 1mo1d    // 2018-03-01T00:00:00Z, Months are applied first to get February 28th, then days are added resulting in March 1 in 2018.

[IMPL#657](https://github.com/influxdata/platform/issues/657) Implement Duration vectors

#### Date and time literals

A date and time literal represents a specific moment in time.
It has a date part, a time part and a time offset part.
The format follows the RFC 3339 specification.
The time is optional, when it is omitted the time is assumed to be midnight for the default location.
The time_offset is optional, when it is omitted the location option is used to determine the offset.

    date_time_lit     = date [ "T" time ] .
    date              = year_lit "-" month "-" day .
    year              = decimal_digit decimal_digit decimal_digit decimal_digit .
    month             = decimal_digit decimal_digit .
    day               = decimal_digit decimal_digit .
    time              = hour ":" minute ":" second [ fractional_second ] [ time_offset ] .
    hour              = decimal_digit decimal_digit .
    minute            = decimal_digit decimal_digit .
    second            = decimal_digit decimal_digit .
    fractional_second = "."  { decimal_digit } .
    time_offset       = "Z" | ("+" | "-" ) hour ":" minute .

Examples:

    1952-01-25T12:35:51Z
    2018-08-15T13:36:23-07:00
    2009-10-15T09:00:00       // October 15th 2009 at 9 AM in the default location
    2018-01-01                // midnight on January 1st 2018 in the default location

[IMPL#152](https://github.com/influxdata/flux/issues/152) Implement shorthand time literals

#### String literals

A string literal represents a sequence of characters enclosed in double quotes.
Within the quotes any character may appear except an unescaped double quote.
String literals support several escape sequences.

    \n   U+000A line feed or newline
    \r   U+000D carriage return
    \t   U+0009 horizontal tab
    \"   U+0022 double quote
    \\   U+005C backslash
    \{   U+007B open curly bracket
    \}   U+007D close curly bracket

Additionally any byte value may be specified via a hex encoding using `\x` as the prefix.


    string_lit       = `"` { unicode_value | byte_value | StringExpression | newline } `"` .
    byte_value       = `\` "x" hex_digit hex_digit .
    hex_digit        = "0" … "9" | "A" … "F" | "a" … "f" .
    unicode_value    = unicode_char | escaped_char .
    escaped_char     = `\` ( "n" | "r" | "t" | `\` | `"` ) .
    StringExpression = "{" Expression "}" .

TODO(nathanielc): With string interpolation string_lit is not longer a lexical token as part of a literal, but an entire expression in and of itself.


[IMPL#252](https://github.com/influxdata/platform/issues/252) Parse string literals


Examples:

    "abc"
    "string with double \" quote"
    "string with backslash \\"
    "日本語"
    "\xe6\x97\xa5\xe6\x9c\xac\xe8\xaa\x9e" // the explicit UTF-8 encoding of the previous line

String literals are also interpolated for embedded expressions to be evaluated as strings.
Embedded expressions are enclosed in curly brackets "{}".
The expressions are evaluated in the scope containing the string literal.
The result of an expression is formatted as a string and replaces the string content between the brackets.
All types are formatted as strings according to their literal representation.
A function "printf" exists to allow more precise control over formatting of various types.
To include the literal curly brackets within a string they must be escaped.


[IMPL#248](https://github.com/influxdata/platform/issues/248) Add printf function

Interpolation example:

    n = 42
    "the answer is {n}" // the answer is 42
    "the answer is not {n+1}" // the answer is not 43
    "openinng curly bracket \{" // openinng curly bracket {
    "closing curly bracket \}" // closing curly bracket }

[IMPL#251](https://github.com/influxdata/platform/issues/251) Add string interpolation support


#### Regular expression literals

A regular expression literal represents a regular expression pattern, enclosed in forward slashes.
Within the forward slashes, any unicode character may appear except for an unescaped forward slash.
The `\x` hex byte value representation from string literals may also be present.

Regular expression literals support only the following escape sequences:

    \/   U+002f forward slash
    \\   U+005c backslash


    regexp_lit         = "/" regexp_char { regexp_char } "/" .
    regexp_char        = unicode_char | byte_value | regexp_escape_char .
    regexp_escape_char = `\` (`/` | `\`)

Examples:

    /.*/
    /http:\/\/localhost:9999/
    /^\xe6\x97\xa5\xe6\x9c\xac\xe8\xaa\x9e(ZZ)?$/
    /^日本語(ZZ)?$/ // the above two lines are equivalent
    /\\xZZ/ // this becomes the literal pattern "\xZZ"
    /a\/b\\c\d/ // escape sequences and character class shortcuts are supported
    /(?:)/ // the empty regular expression

The regular expression syntax is defined by [RE2](https://github.com/google/re2/wiki/Syntax).

### Variables

A variable represents a storage location for a single value.
Variables are immutable.
Once a variable is given a value, it holds that value for the remainder of its lifetime.

### Options

An option represents a storage location for any value of a specified type.
Options are mutable.
An option can hold different values during its lifetime.

Below is a list of some built-in options that are currently implemented in the Flux language:

* now
* task
* location

##### now

The `now` option is a function that returns a time value to be used as a proxy for the current system time.

    // Query should execute as if the below time is the current system time
    option now = () => 2006-01-02T15:04:05-07:00

##### task

The `task` option is used by a scheduler to schedule the execution of a Flux query.

    option task = {
        name: "foo",        // name is required
        every: 1h,          // task should be run at this interval
        delay: 10m,         // delay scheduling this task by this duration
        cron: "0 2 * * *",  // cron is a more sophisticated way to schedule. every and cron are mutually exclusive
        retry: 5,           // number of times to retry a failed query
    }

##### location

The `location` option is used to set the default time zone of all times in the script.
The location maps the UTC offset in use at that location for a given time.
The default value is set using the time zone of the running process.

    option location = fixedZone(offset:-5h) // set timezone to be 5 hours west of UTC
    option location = loadLocation(name:"America/Denver") // set location to be America/Denver

[IMPL#660](https://github.com/influxdata/platform/issues/660) Implement Location option

### Types

A type defines a set of values and operations on those values.
Types are never explicitly declared as part of the syntax.
Types are always inferred from the usage of the value.
Type inference follows a Hindley-Milner style inference system.

#### Boolean types

A _boolean type_ represents a truth value, corresponding to the preassigned variables `true` and `false`.
The boolean type name is `bool`.

#### Numeric types

A _numeric type_ represents sets of integer or floating-point values.

The following numeric types exist:

    uint    the set of all unsigned 64-bit integers
    int     the set of all signed 64-bit integers
    float   the set of all IEEE-754 64-bit floating-point numbers

#### Time types

A _time type_ represents a single point in time with nanosecond precision.
The time type name is `time`.


#### Duration types

A _duration type_ represents a length of time with nanosecond precision.
The duration type name is `duration`.

Durations can be added to times to produce a new time.

Examples:

    2018-07-01T00:00:00Z + 1mo // 2018-08-01T00:00:00Z
    2018-07-01T00:00:00Z + 2y  // 2020-07-01T00:00:00Z
    2018-07-01T00:00:00Z + 5h  // 2018-07-01T05:00:00Z

#### String types

A _string type_ represents a possibly empty sequence of characters.
Strings are immutable: once created they cannot be modified.
The string type name is `string`.

The length of a string is its size in bytes, not the number of characters, since a single character may be multiple bytes.

#### Regular expression types

A _regular expression type_ represents the set of all patterns for regular expressions.
The regular expression type name is `regexp`.

#### Array types

An _array type_ represents a sequence of values of any other type.
All values in the array must be of the same type.
The length of an array is the number of elements in the array.

#### Object types

An _object type_ represents a set of unordered key and value pairs.
The key must always be a string.
The value may be any other type, and need not be the same as other values within the object.

#### Function types

A _function type_ represents a set of all functions with the same argument and result types.


[IMPL#249](https://github.com/influxdata/platform/issues/249) Specify type inference rules

#### Generator types

A _generator type_ represents a value that produces an unknown number of other values.
The generated values may be of any other type but must all be the same type.

[IMPL#658](https://github.com/influxdata/platform/query/issues/658) Implement Generators types

#### Polymorphism

Flux types can be polymorphic, meaning that a type may take on many different types.
Flux supports let-polymorphism and structural polymorphism.

Let-polymorphism is the concept that each time an identifier is referenced is may take on a different type.
For example:

    add = (a,b) => a + b
    add(a:1,b:2) // 3
    add(a:1.5,b:2.0) // 3.5

The identifiers `a` and `b` in the body of the `add` function are used as both `int` and `float` types.
This is let-polymorphism, each different use of an identifier may have a different type.

Structural polymorphism is the concept that structures (objects in Flux) can be used by the same function even if the structures themselves are different.
For example:

    john = {name:"John", lastName:"Smith"}
    jane = {name:"Jane", age:44}

    // John and Jane are objects with different types.
    // We can still define a function that can operate on both objects safely.

    // name returns the name of a person
    name = (person) => person.name

    name(person:john) // John
    name(person:jane) // Jane

    device = {id: 125325, lat: 15.6163, lon: 62.6623}

    name(person:device) // Type error, "device" does not have a property name.

This is structural polymorphism, objects of differing types can be used as the same type so long as they both contain the necessary properties. The necessary properties are determined by the use of the object.

This form of polymorphism means that these checks are performed during type inference and not during runtime. Type errors are found and reported before runtime.

### Blocks

A _block_ is a possibly empty sequence of statements within matching brace brackets.

    Block         = "{" StatementList "} .
    StatementList = { Statement } .

In addition to explicit blocks in the source code, there are implicit blocks:

1. The _universe block_ encompasses all Flux source text.
2. Each package has a _package block_ containing all Flux source text for that package.
3. Each file has a _file block_ containing all Flux source text in that file.
4. Each function literal has its own _function block_ even if not explicitly declared.

Blocks nest and influence scoping.

### Assignment and scope

An assignment binds an identifier to a variable, option, or function.
Every identifier in a program must be assigned.

Flux is lexically scoped using blocks:

1. The scope of a preassigned identifier is in the universe block.
2. The scope of an identifier denoting a variable, option, or function at the top level (outside any function) is the package block.
3. The scope of the name of an imported package is the file block of the file containing the import declaration.
4. The scope of an identifier denoting a function argument is the function body.
5. The scope of an identifier assigned inside a function is the innermost containing block.

Note that the package clause is not an assignment.
The package name does not appear in any scope.
Its purpose is to identify the files belonging to the same package and to specify the default package name for import declarations.

[IMPL#247](https://github.com/influxdata/platform/issues/247) Add package/namespace support

#### Variable assignment

    VariableAssignment = identifier "=" Expression

A variable assignment creates a variable bound to an identifier and gives it a type and value.
A variable keeps the same type and value for the remainder of its lifetime.
An identifier assigned to a variable in a block cannot be reassigned in the same block.
An identifier can be reassigned or shadowed in an inner block.

Examples:

    n = 1
    m = 2
    x = 5.4
    f = () => {
        n = "a"
        m = "b"
        return a + b
    }

#### Option assignment

    OptionAssignment = "option" [ identifier "." ] identifier "=" Expression

An option assignment creates an option bound to an identifier and gives it a type and a value.
Options may only be assigned in a package block.
Once declared, an option may not be redeclared in the same package block.
An option declared in one package may be reassigned a new value in another.
An option keeps the same type for the remainder of its lifetime.

Examples:

    // alert package
    option severity = ["low", "moderate", "high"]

    // foo package
    import "alert"

    option alert.severity = ["low", "critical"]  // qualified option

    option n = 2

    f = (a, b) => a + b + n

    x = f(a:1, b:1) // x = 4

### Expressions

An expression specifies the computation of a value by applying the operators and functions to operands.

#### Operands and primary expressions

Operands denote the elementary values in an expression.

Primary expressions are the operands for unary and binary expressions. 
A primary expressions may be a literal, an identifier denoting a variable, or a parenthesized expression.

    PrimaryExpression = identifier | Literal | "(" Expression ")" .

#### Logical Operators

Flux provides the logical operators `and` and `or`.  
Flux's logical operators observe the short-circuiting behavior seen in other programming languages, meaning that the right-hand side (RHS) operand is conditionally evaluated depending on the result of evaluating the left-hand side (LHS) operand.

When the operator is `and`:
- If the LHS operand evaluates to `false`, a value of `false` is produced and the RHS operand is not evaluated.

When the operator is `or`:
- If the LHS operand evaluates to `true`, a value of `true` is produced and the RHS operand is not evaluated.

#### Literals

Literals construct a value.

    Literal = int_lit
            | float_lit
            | string_lit
            | regex_lit
            | duration_lit
            | pipe_receive_lit
            | ObjectLiteral
            | ArrayLiteral
            | FunctionLiteral .

##### Object literals

Object literals construct a value with the object type.

    ObjectLiteral = "{" PropertyList "}" .
    PropertyList  = [ Property { "," Property } ] .
    Property      = identifier [ ":" Expression ]
                  | string_lit ":" Expression .

##### Array literals

Array literals construct a value with the array type.

    ArrayLiteral   = "[" ExpressionList "]" .
    ExpressionList = [ Expression { "," Expression } ] .

##### Function literals

A function literal defines a new function with a body and parameters.
The function body may be a block or a single expression.
The function body must have a return statement if it is an explicit block, otherwise the expression is the return value.

    FunctionLiteral    = FunctionParameters "=>" FunctionBody .
    FunctionParameters = "(" [ ParameterList [ "," ] ] ")" .
    ParameterList      = Parameter { "," Parameter } .
    Parameter          = identifier [ "=" Expression ] .
    FunctionBody       = Expression | Block .

Examples:

    () => 1 // function returns the value 1
    (a, b) => a + b // function returns the sum of a and b
    (x=1, y=1) => x * y // function with default values
    (a, b, c) => { // function with a block body
        d = a + b
        return d / c
    }

All function literals are anonymous.
A function may be given a name using a variable assignment.

    add = (a,b) => a + b
    mul = (a,b) => a * b

Function literals are _closures_: they may refer to variables defined is a surrounding block.
Those variables are shared between the function literal and the surrounding block.

#### Call expressions

A call expressions invokes a function with the provided arguments.
Arguments must be specified using the argument name, positional arguments not supported.
Argument order does not matter.
When an argument has a default value, it is not required to be specified.

    CallExpression = "(" PropertyList ")" .

Examples:

    f(a:1, b:9.6)
    float(v:1)

#### Pipe expressions

A pipe expression is a call expression with an implicit piped argument.
Pipe expressions simplify creating long nested call chains.

Pipe expressions pass the result of the left hand expression as the _pipe argument_ to the right hand call expression.
Function literals specify which if any argument is the pipe argument using the _pipe literal_ as the argument's default value.
It is an error to use a pipe expression if the function does not declare a pipe argument.

    pipe_receive_lit = "<-" .

Examples:

    foo = () => // function body elided
    bar = (x=<-) => // function body elided
    baz = (y=<-) => // function body elided
    foo() |> bar() |> baz() // equivalent to baz(x:bar(y:foo()))


#### Index expressions

Index expressions access a value from an array based on a numeric index.

    IndexExpression = "[" Expression "]" .

#### Member expressions

Member expressions access a property of an object.
The property being accessed must be either an identifer or a string literal.
In either case the literal value is the name of the property being accessed, the identifer is not evaluated.
It is not possible to access an object's property using an arbitrary expression.

    MemberExpression        = DotExpression  | MemberBracketExpression
    DotExpression           = "." identifer
    MemberBracketExpression = "[" string_lit "]" .

#### Operators

Operators combine operands into expressions.
The precedence of the operators is given in the table below. Operators with a lower number have higher precedence.

|Precedence| Operator |        Description        |
|----------|----------|---------------------------|
|     1    |  `a()`   |       Function call       |
|          |  `a[]`   |  Member or index access   |
|          |   `.`    |       Member access       |
|     2    | `*` `/`  |Multiplication and division|
|     3    | `+` `-`  | Addition and subtraction  |
|     4    |`==` `!=` |   Comparison operators    |
|          | `<` `<=` |                           |
|          | `>` `>=` |                           |
|          |`=~` `!~` |                           |
|     5    |  `not`   | Unary logical expression  |
|     6    |`and` `or`|    Logical AND and OR     |

The operator precedence is encoded directly into the grammar as the following.

    Expression               = LogicalExpression .
    LogicalExpression        = UnaryLogicalExpression
                             | LogicalExpression LogicalOperator UnaryLogicalExpression .
    LogicalOperator          = "and" | "or" .
    UnaryLogicalExpression   = ComparisonExpression
                             | UnaryLogicalOperator UnaryLogicalExpression .
    UnaryLogicalOperator     = "not" .
    ComparisonExpression     = MultiplicativeExpression
                             | ComparisonExpression ComparisonOperator MultiplicativeExpression .
    ComparisonOperator       = "==" | "!=" | "<" | "<=" | ">" | ">=" | "=~" | "!~" .
    AdditiveExpression       = MultiplicativeExpression
                             | AdditiveExpression AdditiveOperator MultiplicativeExpression .
    AdditiveOperator         = "+" | "-" .
    MultiplicativeExpression = PipeExpression
                             | MultiplicativeExpression MultiplicativeOperator PipeExpression .
    MultiplicativeOperator   = "*" | "/" .
    PipeExpression           = PostfixExpression
                             | PipeExpression PipeOperator UnaryExpression .
    PipeOperator             = "|>" .
    UnaryExpression          = PostfixExpression
                             | PrefixOperator UnaryExpression .
    PrefixOperator           = "+" | "-" .
    PostfixExpression        = PrimaryExpression
                             | PostfixExpression PostfixOperator .
    PostfixOperator          = MemberExpression
                             | CallExpression
                             | IndexExpression .
                             
### Packages

Flux source is organized into packages.
A package consists of one or more source files.
Each source file is parsed individually and composed into a single package.

    File = [ PackageClause ] [ ImportList ] StatementList .
    ImportList = { ImportDeclaration } .

#### Package clause

    PackageClause = "package" identifier .

A package clause defines the name for the current package.
Package names must be valid Flux identifiers.
The package clause must be at the begining of any Flux source file.
All files in the same package must declare the same package name.
When a file does not declare a package clause, all identifiers in that file will belong to the special _main_ package.

[IMPL#247](https://github.com/influxdata/platform/issues/247) Add package/namespace support

##### package main

The _main_ package is special for a few reasons:

1. It defines the entrypoint of a Flux program
2. It cannot be imported
3. All statements are marked as producing side effects

### Statements

A statement controls execution.

    Statement = OptionAssignment
              | BuiltinStatement
              | VariableAssignment
              | ReturnStatement
              | ExpressionStatement .


#### Import declaration

    ImportDeclaration = "import" [identifier] string_lit

Associated with every package is a package name and an import path.
The import statement takes a package's import path and brings all of the identifiers defined in that package into the current scope under a namespace.
The import statment defines the namespace through which to access the imported identifiers.
By default the identifier of this namespace is the package name unless otherwise specified.
For example, given a variable `x` declared in package `foo`, importing `foo` and referencing `x` would look like this:

```
import "import/path/to/package/foo"

foo.x
```

Or this:

```
import bar "import/path/to/package/foo"

bar.x
```

A package's import path is always absolute.
A package may reassign a new value to an option identifier declared in one of its imported packages.
A package cannot access nor modify the identifiers belonging to the imported packages of its imported packages.
Every statement contained in an imported package is evaluated.

#### Return statements

A terminating statement prevents execution of all statements that appear after it in the same block.
A return statement is a terminating statement.

    ReturnStatement = "return" Expression .

#### Expression statements

An expression statement is an expression where the computed value is discarded.

    ExpressionStatement = Expression .

Examples:

    1 + 1
    f()
    a

#### Named types

A named type can be created using a type assignment statement.
A named type is equivalent to the type it describes and may be used interchangeably.

    TypeAssignement   = "type" identifier "=" TypeExpression
    TypeExpression    = identifier
                      | TypeParameter
                      | ObjectType
                      | ArrayType
                      | GeneratorType
                      | FunctionType .
    TypeParameter     = "'" identifier .
    ObjectType        = "{" PropertyTypeList [";" ObjectUpperBound ] "}" .
    ObjectUpperBound  = "any" | PropertyTypeList .
    PropertyTypeList  = PropertyType [ "," PropertyType ] .
    PropertyType      = identifier ":" TypeExpression
                      | string_lit ":" TypeExpression .
    ArrayType         = "[]" TypeExpression .
    GeneratorType     = "[...]" TypeExpression .
    FunctionType      = ParameterTypeList "->" TypeExpression
    ParameterTypeList = "(" [ ParameterType { "," ParameterType } ] ")" .
    ParameterType     = identifier ":" [ pipe_receive_lit ] TypeExpression .

Named types are a separate namespace from values.
It is possible for a value and a type to have the same identifier.
The following named types are built-in.

    bool     // boolean
    int      // integer
    uint     // unsigned integer
    float    // floating point number
    duration // duration of time
    time     // time
    string   // utf-8 encoded string
    regexp   // regular expression
    type     // a type that itself describes a type


When an object's upper bound is not specified, it is assumed to be equal to its lower bound.

Parameters to function types define whether the parameter is a pipe forward parameter and whether the parameter has a default value.
The `<-` indicates the parameter is the pipe forward parameter.

Examples:
 
    // alias the bool type
    type boolean = bool

    // define a person as an object type
    type person = {
        name: string,
        age: int,
    }

    // Define addition on ints
    type intAdd = (a: int, b: int) -> int

    // Define polymorphic addition 
    type add = (a: 'a, b: 'a) -> 'a

    // Define funcion with pipe parameter
    type bar = (foo: <-string) -> string

    // Define object type with an empty lower bound and an explicit upper bound
    type address = {
        ;
        street: string,
        city: string,
        state: string,
        country: string,
        province: string,
        zip: int,
    }

### Side Effects

Side effects can occur in two ways.

1. By reassigning builtin options
2. By calling a function that produces side effects

A function produces side effects when it is explicitly declared to have side effects or when it calls a function that itself produces side effects.

## Package initialization

Packages are initialized in the following order:

1. All imported packages are initialized and assigned to their package identifier. 
2. All option declarations are evaluated and assigned regardless of order. An option cannot have a dependency on another option assigned in the same package block.
3. All variable declarations are evaluated and assigned regardless of order. A variable cannot have a direct or indirect dependency on itself.
4. Any package side effects are evaluated.

A package will only be initialized once across all file blocks and across all packages blocks regardless of how many times it is imported.

Initializing imported packages must be deterministic.
Specifically after all imported packages are initialized, each option must be assigned the same value.
Packages imported in the same file block are initialized in declaration order.
Packages imported across different file blocks have no known order.
When a set of imports modify the same option, they must be ordered by placing them in the same file block.

## Built-ins

Flux contains many preassigned values.
These preassigned values are defined in the source files for the various built-in packages.

### System built-ins

When a built-in value is not expressible in Flux its value may be defined by the hosting environment.
All such values must have a corresponding builtin statement to declare the existence and type of the built-in value.

    BuiltinStatement = "builtin" identifer ":" TypeExpression

Example

    builtin from : (bucket: string, bucketID: string) -> stream

### Time constants

#### Days of the week

Days of the week are represented as integers in the range `[0-6]`.
The following builtin values are defined:

```
Sunday    = 0
Monday    = 1
Tuesday   = 2
Wednesday = 3
Thursday  = 4
Friday    = 5
Saturday  = 6
```


[IMPL#153](https://github.com/influxdata/flux/issues/153) Add Days of the Week constants

### Months of the year

Months are represented as integers in the range `[1-12]`.
The following builtin values are defined:

```
January   = 1
February  = 2
March     = 3
April     = 4
May       = 5
June      = 6
July      = 7
August    = 8
September = 9
October   = 10
November  = 11
December  = 12
```

[IMPL#154](https://github.com/influxdata/flux/issues/154) Add Months of the Year constants

### Time and date functions

These are builtin functions that all take a single `time` argument and return an integer.

* `second` int
    Second returns the second of the minute for the provided time in the range `[0-59]`.
* `minute` int
    Minute returns the minute of the hour for the provided time in the range `[0-59]`.
* `hour` int
    Hour returns the hour of the day for the provided time in the range `[0-59]`.
* `weekDay` int
    WeekDay returns the day of the week for the provided time in the range `[0-6]`.
* `monthDay` int
    MonthDay returns the day of the month for the provided time in the range `[1-31]`.
* `yearDay` int
    YearDay returns the day of the year for the provided time in the range `[1-366]`.
* `month` int
    Month returns the month of the year for the provided time in the range `[1-12]`.

[IMPL#155](https://github.com/influxdata/flux/issues/155) Implement Time and date functions 

### System Time

The builtin function `systemTime` returns the current system time.
All calls to `systemTime` within a single evaluation of a Flux script return the same time.

### Intervals

Intervals is a function that produces a set of time intervals over a range of time.
An interval is an object with `start` and `stop` properties that correspond to the inclusive start and exclusive stop times of the time interval.
The return value of `intervals` is another function that accepts `start` and `stop` time parameters and returns an interval generator.
The generator is then used to produce the set of intervals.
The set of intervals will include all intervals that intersect with the initial range of time.
The `intervals` function is designed to be used with the `intervals` parameter of the `window` function.

By default the end boundary of an interval will align with the Unix epoch (zero time) modified by the offset of the `location` option.

An interval is a built-in named type:

    type interval = {
        start: time,
        stop: time,
    }
    
Intervals has the following parameters:

| Name   | Type                         | Description                                                                                                                                                                                                             |
| ----   | ----                         | -----------                                                                                                                                                                                                             |
| every  | duration                     | Every is the duration between starts of each of the intervals. Defaults to the value of the `period` duration.                                                                                                          |
| period | duration                     | Period is the length of each interval. It can be negative, indicating the start and stop boundaries are reversed. Defaults to the value of the `every` duration.                                                        |
| offset | duration                     | Offset is the duration by which to shift the window boundaries. It can be negative, indicating that the offset goes backwards in time. Defaults to 0, which will align window end boundaries with the `every` duration. |
| filter | (interval: interval) -> bool | Filter accepts an interval object and returns a boolean value. Defaults to include all intervals.                                                                                                                       |

The Nth interval start date is the initial start date plus the offset plus an Nth multiple of the every parameter.
Each interval stop date is equal to the interval start date plus the period duration.
When filtering intervals each potential interval is passed to the filter function, when the function returns false, that interval is excluded from the set of intervals.

The intervals function has the following signature:

    (start: time, stop: time) -> (start: time, stop: time) -> [...]interval

Examples:

    intervals(every:1h)                        // 1 hour intervals
    intervals(every:1h, period:2h)             // 2 hour long intervals every 1 hour
    intervals(every:1h, period:2h, offset:30m) // 2 hour long intervals every 1 hour starting at 30m past the hour
    intervals(every:1w, offset:1d)             // 1 week intervals starting on Monday (by default weeks start on Sunday)
    intervals(every:1d, period:-1h)            // the hour from 11PM - 12AM every night
    intervals(every:1mo, period:-1d)           // the last day of each month

Examples using a predicate:

    // 1 day intervals excluding weekends
    intervals(
        every:1d,
        filter: (interval) => !(weekday(time: interval.start) in [Sunday, Saturday]),
    )
    // Work hours from 9AM - 5PM on work days.
    intervals(
        every:1d,
        period:8h,
        offset:9h,
        filter:(interval) => !(weekday(time: interval.start) in [Sunday, Saturday]),
    )

Examples using known start and stop dates:

    // Every hour for six hours on Sep 5th.
    intervals(every:1h)(start:2018-09-05T00:00:00-07:00, stop: 2018-09-05T06:00:00-07:00)
    // [2018-09-05T00:00:00-07:00, 2018-09-05T01:00:00-07:00)
    // [2018-09-05T01:00:00-07:00, 2018-09-05T02:00:00-07:00)
    // [2018-09-05T02:00:00-07:00, 2018-09-05T03:00:00-07:00)
    // [2018-09-05T03:00:00-07:00, 2018-09-05T04:00:00-07:00)
    // [2018-09-05T04:00:00-07:00, 2018-09-05T05:00:00-07:00)
    // [2018-09-05T05:00:00-07:00, 2018-09-05T06:00:00-07:00)

    // Every hour for six hours with 1h30m periods on Sep 5th
    intervals(every:1h, period:1h30m)(start:2018-09-05T00:00:00-07:00, stop: 2018-09-05T06:00:00-07:00)
    // [2018-09-05T00:00:00-07:00, 2018-09-05T01:30:00-07:00)
    // [2018-09-05T01:00:00-07:00, 2018-09-05T02:30:00-07:00)
    // [2018-09-05T02:00:00-07:00, 2018-09-05T03:30:00-07:00)
    // [2018-09-05T03:00:00-07:00, 2018-09-05T04:30:00-07:00)
    // [2018-09-05T04:00:00-07:00, 2018-09-05T05:30:00-07:00)
    // [2018-09-05T05:00:00-07:00, 2018-09-05T06:30:00-07:00)

    // Every hour for six hours using the previous hour on Sep 5th
    intervals(every:1h, period:-1h)(start:2018-09-05T12:00:00-07:00, stop: 2018-09-05T18:00:00-07:00)
    // [2018-09-05T11:00:00-07:00, 2018-09-05T12:00:00-07:00)
    // [2018-09-05T12:00:00-07:00, 2018-09-05T13:00:00-07:00)
    // [2018-09-05T13:00:00-07:00, 2018-09-05T14:00:00-07:00)
    // [2018-09-05T14:00:00-07:00, 2018-09-05T15:00:00-07:00)
    // [2018-09-05T15:00:00-07:00, 2018-09-05T16:00:00-07:00)
    // [2018-09-05T16:00:00-07:00, 2018-09-05T17:00:00-07:00)
    // [2018-09-05T17:00:00-07:00, 2018-09-05T18:00:00-07:00)

    // Every month for 4 months starting on Jan 1st
    intervals(every:1mo)(start:2018-01-01, stop: 2018-05-01)
    // [2018-01-01, 2018-02-01)
    // [2018-02-01, 2018-03-01)
    // [2018-03-01, 2018-04-01)
    // [2018-04-01, 2018-05-01)

    // Every month for 4 months starting on Jan 15th
    intervals(every:1mo)(start:2018-01-15, stop: 2018-05-15)
    // [2018-01-15, 2018-02-15)
    // [2018-02-15, 2018-03-15)
    // [2018-03-15, 2018-04-15)
    // [2018-04-15, 2018-05-15)


[IMPL#659](https://github.com/influxdata/platform/query/issues/659) Implement intervals function


### Builtin Intervals

The following builtin intervals exist:

    // 1 second intervals
    seconds = intervals(every:1s)
    // 1 minute intervals
    minutes = intervals(every:1m)
    // 1 hour intervals
    hours = intervals(every:1h)
    // 1 day intervals
    days = intervals(every:1d)
    // 1 day intervals excluding Sundays and Saturdays
    weekdays = intervals(every:1d, filter: (interval) => weekday(time:interval.start) not in [Sunday, Saturday])
    // 1 day intervals including only Sundays and Saturdays
    weekdends = intervals(every:1d, filter: (interval) => weekday(time:interval.start) in [Sunday, Saturday])
    // 1 week intervals
    weeks = intervals(every:1w)
    // 1 month interval
    months = intervals(every:1mo)
    // 3 month intervals
    quarters = intervals(every:3mo)
    // 1 year intervals
    years = intervals(every:1y)


### FixedZone

FixedZone creates a location based on a fixed time offset from UTC.


FixedZone has the following parameters:

| Name   | Type     | Description                                                                                                                    |
| ----   | ----     | -----------                                                                                                                    |
| offset | duration | Offset is the offset from UTC for the time zone. Offset must be less than 24h. Defaults to 0, which produces the UTC location. |

Examples:

    fixedZone(offset:-5h) // time zone 5 hours west of UTC
    fixedZone(offset:4h30m) // time zone 4 and a half hours east of UTC


[IMPL#156](https://github.com/influxdata/flux/issues/156) Implement FixedZone function

#### LoadLocation

LoadLoacation loads a locations from a time zone database.

LoadLocation has the following parameters:

| Name | Type   | Description                                                                                                                  |
| ---- | ----   | -----------                                                                                                                  |
| name | string | Name is the name of the location to load. The names correspond to names in the [IANA tzdb](https://www.iana.org/time-zones). |

Examples:

    loadLocation(name:"America/Denver")
    loadLocation(name:"America/Chicago")
    loadLocation(name:"Africa/Tunis")

[IMPL#157](https://github.com/influxdata/flux/issues/157) Implement LoadLoacation function

## Data model

Flux employs a basic data model built from basic data types.
The data model consists of tables, records, columns and streams.

### Record

A record is a tuple of named values and is represented using an object type.

### Column

A column has a label and a data type.

The available data types for a column are:

    bool     a boolean value, true or false.
    uint     an unsigned 64-bit integer
    int      a signed 64-bit integer
    float    an IEEE-754 64-bit floating-point number
    string   a sequence of unicode characters
    bytes    a sequence of byte values
    time     a nanosecond precision instant in time
    duration a nanosecond precision duration of time


### Table

A table is set of records, with a common set of columns and a group key.

The group key is a list of columns.
A table's group key denotes which subset of the entire dataset is assigned to the table.
As such, all records within a table will have the same values for each column that is part of the group key.
These common values are referred to as the group key value, and can be represented as a set of key value pairs.

A tables schema consists of its group key, and its column's labels and types.

[IMPL#463](https://github.com/influxdata/flux/issues/463) Specify the primitive types that make up stream and table types

### Stream of tables

A stream represents a potentially unbounded set of tables.
A stream is grouped into individual tables using the group key.
Within a stream each table's group key value is unique.

[IMPL#463](https://github.com/influxdata/flux/issues/463) Specify the primitive types that make up stream and table types

### Missing values

A record may be missing a value for a specific column.
Missing values are represented with a special _null_ value.
The _null_ value can be of any data type.

[IMPL#300](https://github.com/influxdata/platform/issues/300) Design how nulls behave

### Transformations

Transformations define a change to a stream.
Transformations may consume an input stream and always produce a new output stream.
The order of group keys for the output stream will have a stable output order based on the input stream.
The specific ordering may change between releases and it would not be considered a breaking change.

Most transformations output one table for every table they receive from the input stream.
Transformations that modify the group keys or values will need to regroup the tables in the output stream.
A transformation produces side effects when it is constructed from a function that produces side effects.

Transformations are represented using function types.

### Built-in transformations

The following functions are preassigned in the universe block.
These functions each define a transformation.

#### From

From produces a stream of tables from the specified bucket.
Each unique series is contained within its own table.
Each record in the table represents a single point in the series.

The tables schema will include the following columns:

* `_time`
    the time of the record
* `_value`
    the value of the record
* `_start`
    the inclusive lower time bound of all records
* `_stop`
    the exclusive upper time bound of all records

Additionally any tags on the series will be added as columns.
The default group key for the tables is every column except `_time` and `_value`.


From has the following properties:

| Name     | Type   | Description                                                       |
| ----     | ----   | -----------                                                       |
| bucket   | string | Bucket is the name of the bucket to query.                        |
| bucketID | string | BucketID is the string encoding of the ID of the bucket to query. |

Example:

    from(bucket:"telegraf/autogen")
    from(bucketID:"0261d8287f4d6000")

#### Buckets

Buckets is a type of data source that retrieves a list of buckets that the caller is authorized to access.  
It takes no input parameters and produces an output table with the following columns: 

| Name            | Type     | Description                                                 |
| ----            | ----     | -----------                                                 |
| name            | string   | The name of the bucket.                                     |
| id              | string   | The internal ID of the bucket.                              |
| organization    | string   | The organization this bucket belongs to.                    |
| organizationID  | string   | The internal ID of the organization.                        |
| retentionPolicy | string   | The name of the retention policy, if present.               |
| retentionPeriod | duration | The duration of time for which data is held in this bucket. |

Example: 

    buckets() |> filter(fn: (r) => r.organization == "my-org") 

#### Yield

Yield indicates that the stream received by the yield operation should be delivered as a result of the query.
A query may have multiple results, each identified by the name provided to yield.

Yield outputs the input stream unmodified.

Yield has the following properties:

| Name | Type   | Description                             |
| ---- | ----   | -----------                             |
| name | string | Unique name to give to yielded results. |

Example:
`from(bucket: "telegraf/autogen") |> range(start: -5m) |> yield(name:"1")`

**Note:** The `yield` function produces side effects.

#### Fill

Fill will scan a stream for null values and replace them with a non-null value.  

The output stream will be the same as the input stream, with all null values in the column replaced.  

Fill has the following properties: 

| Name        | Type                                 | Description                                                                                   |
| ----        | ----                                 | -----------                                                                                   |
| column      | string                               | The column to fill. Defaults to `"_value"`                                                                          |
| value       | bool, int, uint, float, string, time | The constant value to use in place of nulls. The type must match the type of the valueColumn. |
| usePrevious | bool                                 | If set, then assign the value set in the previous non-null row. Cannot be used with `value`.  |

#### AssertEquals

AssertEquals is a function that will test whether two streams have identical data.  It also outputs the data from the tested stream unchanged, so that this function can be used to perform in-line tests in a query.

AssertEquals has the following properties:

| Name | Type   | Description                                                             |
| ---- | ----   | -----------                                                             |
| name | string | Unique name given to this assertion.                                    |
| got  | stream | The stream you are testing. May be piped-forward from another function. |
| want | stream | A copy of the expected stream.                                          |

Example:

```
option now = () => 2015-01-01T00:00:00Z
want = from(bucket: "backup-telegraf/autogen") |> range(start: -5m)
// in-line assertion
from(bucket: "telegraf/autogen") |> range(start: -5m) |> assertEquals(want: want)

// equivalent:
got = from(bucket: "telegraf/autogen") |> range(start: -5m)
assertEquals(got: got, want: want)
```

#### Diff

Diff is a function that will produce a diff between two table streams.

It will match tables from each stream that have the same group key. For each matched table, it will produce a diff. Any rows that were added or removed will be added to the table as a row.
An additional string column with the name `_diff` will be created which will contain a `"-"` if the row was present in the `got` table and not in the `want` table or `"+"` if the opposite is true.

The `diff` function is guaranteed to emit at least one row if the tables are different and no rows if the tables are the same. The exact diff that is produced may change.

Diff has the following properties:

| Name | Type   | Description                                                             |
| ---- | ----   | -----------                                                             |
| got  | stream | The stream you are testing. May be piped-forward from another function. |
| want | stream | A copy of the expected stream.                                          |

```
option now = () => 2015-01-01T00:00:00Z
want = from(bucket: "backup-telegraf/autogen") |> range(start: -5m)
// in-line assertion
from(bucket: "telegraf/autogen") |> range(start: -5m) |> assertEquals(want: want)

// equivalent:
got = from(bucket: "telegraf/autogen") |> range(start: -5m)
diff(got: got, want: want)
```

#### Aggregate operations

Aggregate operations output a table for every input table they receive.
A list of columns to aggregate must be provided to the operation.
The aggregate function is applied to each column in isolation.

Any output table will have the following properties:

* It always contains a single record.
* It will have the same group key as the input table.
* It will contain a column for each provided aggregate column.
    The column label will be the same as the input table.
    The type of the column depends on the specific aggregate operation.
    The value of the column will be null if the input table is empty or the input column has only null values.
* It will not have a _time column

All aggregate operations have the following properties:

| Name    | Type     | Description                                       |
| ----    | ----     | -----------                                       |
| columns | []string | Columns specifies a list of columns to aggregate. |

##### AggregateWindow

AggregateWindow is a function that simplifies aggregating data over fixed windows of time.
AggregateWindow windows the data, performs an aggregate operation, and then undoes the windowing to produce
an output table for every input table.

AggregateWindow has the following properties:


| Name        | Type                                            | Description                                                                                                           |
| ----        | ----                                            | -----------                                                                                                           |
| every       | duration                                        | Every specifies the window size to aggregate.                                                                         |
| fn          | (tables: <-stream, columns: []string) -> stream | Fn specifies the aggregate operation to perform. Any of the functions in this Aggregate section can be provided.      |
| columns     | []string                                        | Columns specifies a list of columns to aggregate. Defaults to ["_value"].                                             |
| timeSrc     | string                                          | TimeSrc is the name of a column from the group key to use as the source for the aggregated time. Defaults to "_stop". |
| timeDst     | string                                          | TimeDst is the name of a new column in which the aggregated time is placed. Defaults to "_time".                      |
| createEmpty | bool                                            | CreateEmpty, if true, will create empty windows and fill them with a null aggregate value.  Defaults to true.         |
Example:

```
// Compute the mean over 1m intervals for the last 1h.
from(bucket: "telegraf/autogen")
  |> range(start: -1h)
  |> filter(fn: (r) => r._measurement == "cpu" and r._field == "usage_system")
  |> aggregateWindow(every: 1m, fn:mean)
```

##### Covariance

Covariance is an aggregate operation.
Covariance computes the covariance between two columns.

Covariance has the following properties:

| Name     | Type     | Description                                                                                 |
| ----     | ----     | -----------                                                                                 |
| columns  | []string | Columns specifies a list of columns to aggregate. Defaults to `["_value"]`.                 |
| pearsonr | bool     | Pearsonr indicates whether the result should be normalized to be the Pearson R coefficient. |
| valueDst | string   | ValueDst is the column into which the result will be placed. Defaults to `_value`.          |

Additionally exactly two columns must be provided to the `columns` property.

Example:
`from(bucket: "telegraf/autogen") |> range(start:-5m) |> covariance(columns: ["x", "y"])`

#### Cov

Cov computes the covariance between two streams by first joining the streams and then performing the covariance operation per joined table.

Cov has the following properties:

| Name     | Type     | Description                                                                                 |
| ----     | ----     | -----------                                                                                 |
| x        | stream   | X is one of the input streams.                                                              |
| y        | stream   | Y is one of the input streams.                                                              |
| on       | []string | On is the list of columns on which to join.                                                 |
| pearsonr | bool     | Pearsonr indicates whether the result should be normalized to be the Pearson R coefficient. |
| valueDst | string   | ValueDst is the column into which the result will be placed.  Defaults to `_value`.         |

Example:
    
    cpu = from(bucket: "telegraf/autogen") |> range(start:-5m) |> filter(fn:(r) => r._measurement == "cpu")
    mem = from(bucket: "telegraf/autogen") |> range(start:-5m) |> filter(fn:(r) => r._measurement == "mem")
    cov(x: cpu, y: mem)

#### Pearsonr

Pearsonr computes the Pearson R correlation coefficient bewteen two streams.
It is defined in terms of the `cov` function:

    pearsonr = (x,y,on) => cov(x:x, y:y, on:on, pearsonr:true)

##### Count

Count is an aggregate operation.
For each aggregated column, it outputs the number of records as an integer. It will count both null and non-null records.

Count has the following property:

| Name    | Type     | Description                                                                 |
| ----    | ----     | -----------                                                                 |
| columns | []string | Columns specifies a list of columns to aggregate. Defaults to `["_value"]`. |

Example:
```
from(bucket: "telegraf/autogen") |> range(start: -5m) |> count()
```


##### Integral

Integral is an aggregate operation.
For each aggregate column, it outputs the area under the curve of records.
The curve is defined as a function where the domain is the record times and the range is the record values.
This function will return an error if values in the time column are null or not sorted in
ascending order.
Null values in aggregate columns are skipped.

Integral has the following properties:

| Name       | Type     | Description                                                                           |
| ----       | ----     | -----------                                                                           |
| columns    | []string | Columns specifies a list of columns to aggregate. Defaults to `["_value"]`.           |
| unit       | duration | Unit is the time duration to use when computing the integral.  Defaults to `1s`       |
| timeColumn | string   | TimeColumn is the name of the column containing the time value.  Defaults to `_time`. |

Example:

```
from(bucket: "telegraf/autogen")
    |> range(start: -5m)
    |> filter(fn: (r) => r._measurement == "cpu" and r._field == "usage_system")
    |> integral(unit:10s)
```

##### Mean

Mean is an aggregate operation.
For each aggregated column, it outputs the mean of the non null records as a float.

Mean has the following property:

| Name    | Type     | Description                                                                 |
| ----    | ----     | -----------                                                                 |
| columns | []string | Columns specifies a list of columns to aggregate. Defaults to `["_value"]`. |

Example:
```
from(bucket:"telegraf/autogen")
    |> filter(fn: (r) => r._measurement == "mem" AND
            r._field == "used_percent")
    |> range(start:-12h)
    |> window(every:10m)
    |> mean()
```

##### Median (aggregate)

Median is defined as:

    median = (method="estimate_tdigest", compression=0.0, tables=<-) =>
    	tables
    		|> percentile(percentile:0.5, method:method, compression:compression)

Is it simply a `percentile` with the `percentile` paramter always set to `0.5`.
It therefore shares all the same properties as the percentile function.

Example:
```
// Determine median cpu system usage:
from(bucket: "telegraf/autogen")
	|> range(start: -5m)
	|> filter(fn: (r) => r._measurement == "cpu" and r._field == "usage_system")
	|> median()
```

##### Percentile (aggregate)

Percentile is both an aggregate operation and a selector operation depending on selected options.
In the aggregate methods, it outputs the value that represents the specified percentile of the non null record as a float.

Percentile has the following properties:

| Name        | Type     | Description                                                                                                                                                                                   |
| ----        | ----     | -----------                                                                                                                                                                                   |
| columns     | []string | Columns specifies a list of columns to aggregate. Defaults to `["_value"]`                                                                                                                    |
| percentile  | float    | Percentile is a value between 0 and 1 indicating the desired percentile.                                                                                                                      |
| method      | string   | Method must be one of: estimate_tdigest, exact_mean, or exact_selector.                                                                                                                       |
| compression | float    | Compression indicates how many centroids to use when compressing the dataset. A larger number produces a more accurate result at the cost of increased memory requirements. Defaults to 1000. |


The method parameter must be one of:

* `estimate_tdigest`: an aggregate result that uses a tdigest data structure to compute an accurate percentile estimate on large data sources. 
* `exact_mean`: an aggregate result that takes the average of the two points closest to the percentile value. 
* `exact_selector`: see Percentile (selector) 

Example:
```
// Determine 99th percentile cpu system usage:
from(bucket: "telegraf/autogen")
	|> range(start: -5m)
	|> filter(fn: (r) => r._measurement == "cpu" and r._field == "usage_system")
	|> percentile(percentile: 0.99, method: "estimate_tdigest", compression: 1000.0)
```

##### Skew

Skew is an aggregate operation.
For each aggregated column, it outputs the skew of the non null record as a float.

Skew has the following parameter:

| Name    | Type     | Description                                                                 |
| ----    | ----     | -----------                                                                 |
| columns | []string | Columns specifies a list of columns to aggregate. Defaults to `["_value"]`. |

Example:

```
from(bucket: "telegraf/autogen")
    |> range(start: -5m)
    |> filter(fn: (r) => r._measurement == "cpu" and r._field == "usage_system")
    |> skew()
```

##### Spread

Spread is an aggregate operation.
For each aggregated column, it outputs the difference between the min and max values.
The type of the output column depends on the type of input column: for input columns with type `uint` or `int`, the output is an `int`; for `float` input columns the output is a `float`.
All other input types are invalid.

Spread has the following parameter:

| Name    | Type     | Description                                                                 |
| ----    | ----     | -----------                                                                 |
| columns | []string | Columns specifies a list of columns to aggregate. Defaults to `["_value"]`. |

Example:
```
from(bucket: "telegraf/autogen")
    |> range(start: -5m)
    |> filter(fn: (r) => r._measurement == "cpu" and r._field == "usage_system")
    |> spread()
```
##### Stddev

Stddev is an aggregate operation.
For each aggregated column, it outputs the standard deviation of the non null record as a float.

Stddev has the following parameter:

| Name    | Type     | Description                                                                 |
| ----    | ----     | -----------                                                                 |
| columns | []string | Columns specifies a list of columns to aggregate. Defaults to `["_value"]`. |

Example:

```
from(bucket: "telegraf/autogen")
    |> range(start: -5m)
    |> filter(fn: (r) => r._measurement == "cpu" and r._field == "usage_system")
    |> stddev()
```

##### Sum

Stddev is an aggregate operation.
For each aggregated column, it outputs the sum of the non null record.
The output column type is the same as the input column type.

Sum has the following parameter:

| Name    | Type     | Description                                                                 |
| ----    | ----     | -----------                                                                 |
| columns | []string | Columns specifies a list of columns to aggregate. Defaults to `["_value"]`. |

Example:
```
from(bucket: "telegraf/autogen")
    |> range(start: -5m)
    |> filter(fn: (r) => r._measurement == "cpu" and r._field == "usage_system")
    |> sum()
```

#### Multiple aggregates

Multiple aggregates can be applied to the same table using the `aggregate` function.

[IMPL#139](https://github.com/influxdata/platform/issues/139) Add aggregate function

#### Selector operations

Selector operations output a table for every input table they receive.
A single column on which to operate must be provided to the operation.

Any output table will have the following properties:

* It will have the same group key as the input table.
* It will contain the same columns as the input table.
* It will have a column `_time` which represents the time of the selected record.
    This can be set as the value of any time column on the input table.
    By default the `_stop` time column is used.

All selector operations have the following properties:

* `column` string
    column specifies a which column to use when selecting.

##### First

First is a selector operation.
First selects the first non null record from the input table.

Example:
`from(bucket:"telegraf/autogen") |> first()`

##### Last

Last is a selector operation.
Last selects the last non null record from the input table.

Example:
`from(bucket: "telegraf/autogen") |> last()`

##### Max

Max is a selector operation.
Max selects the maximum record from the input table.

Example:
```
from(bucket:"telegraf/autogen")
    |> range(start:-12h)
    |> filter(fn: (r) => r._measurement == "cpu" AND r._field == "usage_system")
    |> max()
```

##### Min

Min is a selector operation.
Min selects the minimum record from the input table.

Example: 

```
from(bucket:"telegraf/autogen")
    |> range(start:-12h)
    |> filter(fn: (r) => r._measurement == "cpu" AND r._field == "usage_system")
    |> min()
```

##### Percentile (selector)

Percentile is both an aggregate operation and a selector operation depending on selected options.
In the aggregate methods, it outputs the value that represents the specified percentile of the non null record as a float.

Percentile has the following properties:

| Name       | Type   | Description                                                                                       |
| ----       | ----   | -----------                                                                                       |
| column     | string | Column indicates which column will be used for the percentile computation. Defaults to `"_value"` |
| percentile | float  | Percentile is a value between 0 and 1 indicating the desired percentile.                                        |
| method     | string | Method must be one of: estimate_tdigest, exact_mean, exact_selector.                              |

The method parameter must be one of:

* `estimate_tdigest`: See Percentile (Aggregate).
* `exact_mean`: See Percentile (Aggregate).
* `exact_selector`: a selector result that returns the data point for which at least `percentile` points are less than.

Example:
```
// Determine 99th percentile cpu system usage:
from(bucket: "telegraf/autogen")
	|> range(start: -5m)
	|> filter(fn: (r) => r._measurement == "cpu" and r._field == "usage_system")
	|> percentile(percentile: 0.99, method: "exact_selector")
```

##### Median (selector)

Median is defined as:

    median = (method="estimate_tdigest", compression=0.0, tables=<-) =>
    	tables
    		|> percentile(percentile:0.5, method:method, compression:compression)

Is it simply a `percentile` with the `percentile` paramter always set to `0.5`.
It therefore shares all the same properties as the percentile function.

Example:
```
// Determine median cpu system usage:
from(bucket: "telegraf/autogen")
	|> range(start: -5m)
	|> filter(fn: (r) => r._measurement == "cpu" and r._field == "usage_system")
	|> median(method: "exact_selector")
```



##### Sample

Sample is a selector operation.
Sample selects a subset of the records from the input table.

The following properties define how the sample is selected.

| Name | Type | Description                                                                                                                                                                  |
| ---- | ---- | -----------                                                                                                                                                                  |
| n    | int  | Sample every Nth element. Must be a positive integer.                                                                                                                        |
| pos  | int  | Pos is the offset from start of results to begin sampling. The `pos` must be less than `n`. If `pos` is less than 0, a random offset is used. Default is -1 (random offset). |

Example:

```
from(bucket:"telegraf/autogen")
    |> filter(fn: (r) => r._measurement == "cpu" AND
               r._field == "usage_system")
    |> range(start:-1d)
    |> sample(n: 5, pos: 1)
```


#### Filter

Filter applies a predicate function to each input record, output tables contain only records which matched the predicate.
One output table is produced for each input table.
The output tables will have the same schema as their corresponding input tables.

Filter has the following properties:

| Name | Type                | Description                                                                                        |
| ---- | ----                | -----------                                                                                        |
| fn   | (r: record) -> bool | Fn is a predicate function. Records which evaluate to true, will be included in the output tables. |

Example:

```
from(bucket:"telegraf/autogen")
    |> range(start:-12h)
    |> filter(fn: (r) => r._measurement == "cpu" AND
                r._field == "usage_system" AND
                r.service == "app-server")
```

#### Highest/Lowest

There are six highest/lowest functions that compute the top or bottom N records from all tables in a stream based on a specific aggregation method.

* highestMax - computes the top N records from all tables using the maximum of each table.
* highestAverage  - computes the top N records from all tables using the average of each table.
* highestCurrent - computes the top N records from all tables using the last value of each table.
* lowestMin - computes the bottom N records from all tables using the minimum of each table.
* lowestAverage - computes the bottom N records from all tables using the average of each table.
* lowestCurrent - computes the bottom N records from all tables using the last value of each table.

All of the highest/lowest functions take the following parameters:

| Name         | Type     | Description                                                                        |
| ----         | ----     | -----------                                                                        |
| n            | int      | N is the number of records to select.                                              |
| columns      | []string | Columns is the list of columns to use when aggregating.  Defaults to `["_value"]`. |
| groupColumns | []string | GroupColumns are the columns on which to group to perform the aggregation.         |

#### Histogram

Histogram approximates the cumulative distribution function of a dataset by counting data frequencies for a list of bins.
A bin is defined by an upper bound where all data points that are less than or equal to the bound are counted in the bin.
The bin counts are cumulative.

Each input table is converted into a single output table representing a single histogram.
The output table will have a the same group key as the input table.
The columns not part of the group key will be removed and an upper bound column and a count column will be added.

Histogram has the following properties:

| Name             | Type    | Description                                                                                                                                                                         |
| ----             | ----    | -----------                                                                                                                                                                         |
| column           | string  | Column is the name of a column containing the input data values. The column type must be float.  Defaults to `_value`.                                                              |
| upperBoundColumn | string  | UpperBoundColumn is the name of the column in which to store the histogram upper bounds. Defaults to `le`.                                                                          |
| countColumn      | string  | CountColumn is the name of the column in which to store the histogram counts. Defaults to `_value`.                                                                                 |
| bins             | []float | Bins is a list of upper bounds to use when computing the histogram frequencies. Each element in the array should contain a float value that represents the maximum value for a bin. |
| normalize        | bool    | Normalize when true will convert the counts into frequencies values between 0 and 1. Normalized histograms cannot be aggregated by summing their counts. Defaults to `false`.       |


Example:

    histogram(bins:linearBins(start:0.0,width:10.0,count:10))  // compute the histogram of the data using 10 bins from 0,10,20,...,100

#### HistogramQuantile

HistogramQuantile approximates a quantile given an histogram that approximates the cumulative distribution of the dataset.
Each input table represents a single histogram.
The histogram tables must have two columns, a count column and an upper bound column.
The count is the number of values that are less than or equal to the upper bound value.
The table can have any number of records, each representing an entry in the histogram.
The counts must be monotonically increasing when sorted by upper bound.
If any values in the count column or upper bound column are null, an error will be returned.

Linear interpolation between the two closest bounds is used to compute the quantile.
If the either of the bounds used in interpolation are infinite, then the other finite bound is used and no interpolation is performed.

The output table will have a the same group key as the input table.
The columns not part of the group key will be removed and a single value column of type float will be added.
The count and upper bound columns must not be part of the group key.
The value column represents the value of the desired quantile from the histogram.

HistogramQuantile has the following properties:

| Name             | Type   | Description                                                                                                                                    |
| ----             | ----   | -----------                                                                                                                                    |
| quantile         | float  | Quantile is a value between 0 and 1 indicating the desired quantile to compute.                                                                |
| countColumn      | string | CountColumn is the name of the column containing the histogram counts. The count column type must be float. Defaults to `_value`.              |
| upperBoundColumn | string | UpperBoundColumn is the name of the column containing the histogram upper bounds. The upper bound column type must be float. Defaults to `le`. |
| valueColumn      | string | ValueColumn is the name of the output column which will contain the computed quantile. Defaults to `_value`.                                   |
| minValue         | float  | MinValue is the assumed minumum value of the dataset. Default to 0.                                                                            |

When the quantile falls below the lowest upper bound, interpolation is performed between minValue and the lowest upper bound. When minValue is equal to negative infinity, the lowest upper bound is used.

Example:

    histogramQuantile(quantile:0.9)  // compute the 90th quantile using histogram data.

##### LinearBins

LinearBins produces a list of linearly separated floats.

LinearBins has the following properties:

| Name      | Type  | Description                                                                                      |
| ----      | ----  | -----------                                                                                      |
| start     | float | Start is the first value in the returned list.                                                   |
| width     | float | Width is the distance between subsequent bin values.                                             |
| count     | int   | Count is the number of bins to create.                                                           |
| inifinity | bool  | Infinity when true adds an additional bin with a value of positive infinity. Defaults to `true`. |

##### LogarithmicBins

LogarithmicBins produces a list of exponentially separated floats.

LogarithmicBins has the following properties:

| Name      | Type  | Description                                                                                      |
| ----      | ----  | -----------                                                                                      |
| start     | float | Start is the first value in the returned bin list.                                               |
| factor    | float | Factor is the multiplier applied to each subsequent bin.                                         |
| count     | int   | Count is the number of bins to create.                                                           |
| inifinity | bool  | Infinity when true adds an additional bin with a value of positive infinity. Defaults to `true`. |

#### Limit

Limit caps the number of records in output tables to a fixed size `n`.
One output table is produced for each input table.
Each output table will contain the first `n` records after the first `offset` records of the input table.
If the input table has less than `offset + n` records, all records except the first `offset` ones will be output.

Limit has the following properties:

| Name   | Type | Description                                                                              |
| ----   | ---- | -----------                                                                              |
| n      | int  | N is the maximum number of records per table to output.                                  |
| offset | int  | Offest is the number of records to skip per table before limiting to `n`. Defaults to 0. |

Example:

```
from(bucket: "telegraf/autogen")
    |> range(start: -1h)
    |> limit(n: 10, offset: 1)
```

#### Map

Map applies a function to each record of the input tables.
The modified records are assigned to new tables based on the group key of the input table.
The output tables are the result of applying the map function to each record on the input tables.

When the output record contains a different value for the group key the record is regroup into the appropriate table.
When the output record drops a column that was part of the group key that column is removed from the group key.

Map has the following properties:

| Name     | Type                  | Description                                                                                               |
| ----     | ----                  | -----------                                                                                               |
| fn       | (r: record) -> record | Function to apply to each record.  The return value must be an object.                                    |
| mergeKey | bool                  | MergeKey indicates if the record returned from fn should be merged with the group key.  Defaults to true. |


When merging, all columns on the group key will be added to the record giving precedence to any columns already present on the record.
When not merging, only columns defined on the returned record will be present on the output records.


[IMPL#816](https://github.com/influxdata/flux/issues/816) Remove mergeKey parameter from map

Example:
```
from(bucket:"telegraf/autogen")
    |> filter(fn: (r) => r._measurement == "cpu" AND
                r._field == "usage_system" AND
                r.service == "app-server")
    |> range(start:-12h)
    // Square the value
    |> map(fn: (r) => r._value * r._value)
```
Example (creating a new table):
```
from(bucket:"telegraf/autogen")
    |> filter(fn: (r) => r._measurement == "cpu" AND
                r._field == "usage_system" AND
                r.service == "app-server")
    |> range(start:-12h)
    // create a new table by copying each row into a new format
    |> map(fn: (r) => ({_time: r._time, app_server: r._service}))
```

#### Reduce
Reduce aggregates records in each table according to the reducer `fn`.  The output for each table will be the group key of the table, plus columns corresponding to each field in the reducer object.  

If the reducer record contains a column with the same name as a group key column, then the group key column's value is overwritten and the the resulting record is regrouped into the appropriate table.

Reduce has the following properties:

| Name     | Type                  | Description                                                                                               |
| ----     | ----                  | -----------                                                                                               |
| fn       | (r: record, accumulator: 'a) -> 'a | Function to apply to each record with a reducer object of type 'a.  |
| identity | 'a                  | an initial value to use when creating a reducer. May be used more than once in asynchronous processing use cases.|


Example (compute the sum of the value column):
```
from(bucket:"telegraf/autogen")
    |> filter(fn: (r) => r._measurement == "cpu" AND
                r._field == "usage_system" AND
                r.service == "app-server")
    |> range(start:-12h)
    |> reduce(fn: (r, accumulator) =>
            ({sum: r._value + accumulator.sum}), identity: {sum: 0.0}))
```

Example (compute the sum and count in a single reducer):
```
from(bucket:"telegraf/autogen")
    |> filter(fn: (r) => r._measurement == "cpu" AND
                r._field == "usage_system" AND
                r.service == "app-server")
    |> range(start:-12h)
    |> reduce(fn: (r, accumulator) =>
            ({sum: r._value + accumulator.sum, count: accumulator.count + 1.0}), identity: {sum: 0.0, count:0.0}))
```

Example (compute the product of all values):
```
from(bucket:"telegraf/autogen")
    |> filter(fn: (r) => r._measurement == "cpu" AND
                r._field == "usage_system" AND
                r.service == "app-server")
    |> range(start:-12h)
    |> reduce(fn: (r, accumulator) =>
            ({prod: r._value * accumulator.prod}), identity: {prod: 1.0}))
```



#### Range

Range filters records based on provided time bounds.
Each input tables records are filtered to contain only records that exist within the time bounds.
Records with a null value for their time will be filtered.
Each input table's group key value is modified to fit within the time bounds.
Tables where all records exists outside the time bounds are filtered entirely.


[IMPL#244](https://github.com/influxdata/platform/issues/244) Update range to default to aligned window ranges.

Range has the following properties:

| Name        | Type   | Description                                                                                                             |
| ----        | ----   | -----------                                                                                                             |
| start       | time   | Start specifies the oldest time to be included in the results.                                                          |
| stop        | time   | Stop specifies the exclusive newest time to be included in the results. Defaults to the value of the `now` option time. |
| timeColumn  | string | Name of the time column to use. Defaults to `_time`.                                                                    |
| startColumn | string | StartColumn is the name of the column containing the start time. Defaults to `_start`.                                  |
| stopColumn  | string | StopColumn is the name of the column containing the stop time. Defaults to `_stop`.                                     |

Example:
```
from(bucket:"telegraf/autogen")
    |> range(start:-12h, stop: -15m)
    |> filter(fn: (r) => r._measurement == "cpu" AND
               r._field == "usage_system")
```
Example:
```
from(bucket:"telegraf/autogen")
    |> range(start:2018-05-22T23:30:00Z, stop: 2018-05-23T00:00:00Z)
    |> filter(fn: (r) => r._measurement == "cpu" AND
               r._field == "usage_system")
```

#### Rename 

Rename renames specified columns in a table.
There are two variants: one which takes a map of old column names to new column names,
and one which takes a mapping function.
If a column is renamed and is part of the group key, the column name in the group key will be updated.
If a specified column is not present in a table an error will be thrown.

Rename has the following properties: 

| Name    | Type                       | Description                                                                                    |
| ----    | ----                       | -----------                                                                                    |
| columns | object                     | Columns is a map of old column names to new names. Cannot be used with `fn`.                   |
| fn      | (column: string) -> string | Fn defines a function mapping between old and new column names. Cannot be used with `columns`. |

Example usage:

Rename a single column:

```
from(bucket: "telegraf/autogen")
    |> range(start: -5m)
    |> rename(columns: {host: "server"})
```

Rename all columns using `fn` parameter:

```
from(bucket: "telegraf/autogen")
    |> range(start: -5m)
    |> rename(fn: (column) => column + "_new")
```

#### Drop 

Drop excludes specified columns from a table. Columns to exclude can be specified either through a list, or a predicate function.
When a dropped column is part of the group key it will also be dropped from the key.
If a specified column is not present in a table an error will be thrown.

Drop has the following properties:

| Name    | Type                     | Description                                                                                           |
| ----    | ----                     | -----------                                                                                           |
| columns | []string                 | Columns is an array of column to exclude from the resulting table. Cannot be used with `fn`.          |
| fn      | (column: string) -> bool | Fn is a predicate function, columns that evaluate to true are dropped. Cannot be used with `columns`. |

Example Usage:

Drop a list of columns:

```
from(bucket: "telegraf/autogen")
	|> range(start: -5m)
	|> drop(columns: ["host", "_measurement"])
```

Drop columns matching a predicate:

```
from(bucket: "telegraf/autogen")
    |> range(start: -5m)
    |> drop(fn: (column) => column =~ /usage*/)
```

#### Keep 

Keep is the inverse of drop. It returns a table containing only columns that are specified,
ignoring all others.
Only columns in the group key that are also specified in `keep` will be kept in the resulting group key.
If a specified column is not present in a table an error will be thrown.

Keep has the following properties:

| Name    | Type                     | Description                                                                                        |
| ----    | ----                     | -----------                                                                                        |
| columns | []string                 | Columns is an array of column to exclude from the resulting table. Cannot be used with `fn`.       |
| fn      | (column: string) -> bool | Fn is a predicate function, columns that evaluate to true are kept. Cannot be used with `columns`. |

Example Usage:

Keep a list of columns:

```
from(bucket: "telegraf/autogen")
    |> range(start: -5m)
    |> keep(columns: ["_time", "_value"])
```

Keep all columns matching a predicate:

```
from(bucket: "telegraf/autogen")
    |> range(start: -5m)
    |> keep(fn: (column) => column =~ /inodes*/) 
```

#### Duplicate 

Duplicate duplicates a specified column in a table.
If the specified column is not present in a table an error will be thrown.
If the specified column is part of the group key, it will be duplicated, but it will not be part of the group key of the output table.

Duplicate has the following properties:

| Name   | Type   | Description                                                     |
| ----   | ----   | -----------                                                     |
| column | string | Column is the name of the column to duplicate.                  |
| as     | string | As is the name that should be assigned to the duplicate column. |

Example usage:

Duplicate column `server` under the name `host`:

```
from(bucket: "telegraf/autogen")
	|> range(start:-5m)
	|> filter(fn: (r) => r._measurement == "cpu")
	|> duplicate(column: "host", as: "server")
```

#### Set

Set assigns a static value to each record.
The key may modify and existing column or it may add a new column to the tables.
If the column that is modified is part of the group key, then the output tables will be regrouped as needed.

Set has the following properties:

| Name  | Type   | Description                             |
| ----  | ----   | -----------                             |
| key   | string | Key is the label for the column to set. |
| value | string | Value is the string value to set.       |

Example:

```
from(bucket: "telegraf/autogen") |> set(key: "mykey", value: "myvalue")
```

#### Sort

Sorts orders the records within each table.
One output table is produced for each input table.
The output tables will have the same schema as their corresponding input tables.
When sorting, nulls will always be first. When `desc: false` is set, then nulls are less than every other value. When `desc: true`, nulls are greater than every value.

Sort has the following properties:

| Name    | Type     | Description                                                                               |
| ----    | ----     | -----------                                                                               |
| columns | []string | Columns is the sort order to use; precedence from left to right. Default is `["_value"]`. |
| desc    | bool     | Desc indicates results should be sorted in descending order. Default is `false`.          |

Example:

```
from(bucket:"telegraf/autogen")
    |> filter(fn: (r) => r._measurement == "system" AND
               r._field == "uptime")
    |> range(start:-12h)
    |> sort(columns:["region", "host", "value"])
```

#### Group

Group groups records based on their values for specific columns.
It produces tables with new group keys based on the provided properties.

Group has the following properties:

| Name    | Type     | Description                                                                |
| ----    | ----     | -----------                                                                |
| columns | []string | Columns is a list used to calculate the new group key. Defaults to `[]`.   |
| mode    | string   | The grouping mode, can be one of `"by"` or `"except"`. Defaults to `"by"`. |

When using `"by"` mode, the specified `columns` are the new group key.
When using `"except"` mode, the new group key is the difference between the columns of the table under exam and `columns`.

__Examples__

_By_

```
from(bucket: "telegraf/autogen") 
    |> range(start: -30m) 
    |> group(columns: ["host", "_measurement"])
```

Or:

```
...
    |> group(columns: ["host", "_measurement"], mode: "by")
```

Records are grouped by the `"host"` and `"_measurement"` columns.  
The resulting group key is `["host", "_measurement"]`, so a new table for every different `["host", "_measurement"]`
value is created.  
Every table in the result contains every record for some `["host", "_measurement"]` value.  
Every record in some resulting table has the same value for the columns `"host"` and `"_measurement"`.

_Except_

```
from(bucket: "telegraf/autogen")
    |> range(start: -30m)
    |> group(columns: ["_time"], mode: "except")
```

Records are grouped by the set of all columns in the table, excluding `"_time"`.  
For example, if the table has columns `["_time", "host", "_measurement", "_field", "_value"]` then the group key would be
`["host", "_measurement", "_field", "_value"]`.

_Single-table grouping_

```
from(bucket: "telegraf/autogen")
    |> range(start: -30m)
    |> group()
```

Records are grouped into a single table.  
The group key of the resulting table is empty.

#### Columns

Columns lists the column labels of input tables.
For each input table, it outputs a table with the same group key columns, plus a new column containing the labels of the input table's columns.
Each row in an output table contains the group key value and the label of one column of the input table.
So, each output table has the same number of rows as the number of columns of the input table.

Columns has the following properties:

| Name   | Type   | Description                                                                               |
| ----   | ----   | -----------                                                                               |
| column | string | Column is the name of the output column to store the column labels. Defaults to `_value`. |

Example:

```
from(bucket: "telegraf/autogen")
    |> range(start: -30m)
    |> columns(column: "labels")
```

Getting every possible column label in a single table as output:

```
from(bucket: "telegraf/autogen")
    |> range(start: -30m)
    |> columns()
    |> keep(columns: ["_value"])
    |> group()
    |> distinct()
```

#### Keys

Keys outputs the group key of input tables.
For each input table, it outputs a table with the same group key columns, plus a `_value` column containing the labels of the input table's group key.
Each row in an output table contains the group key value and the label of one column in the group key of the input table.
So, each output table has the same number of rows as the size of the group key of the input table.

Keys has the following properties:

| Name   | Type   | Description                                                                                  |
| ----   | ----   | -----------                                                                                  |
| column | string | Column is the name of the output column to store the group key labels. Defaults to `_value`. |

Example:

```
from(bucket: "telegraf/autogen")
    |> range(start: -30m)
    |> keys(column: "keys")
```

Getting every possible key in a single table as output:

```
from(bucket: "telegraf/autogen")
    |> range(start: -30m)
    |> keys()
    |> keep(columns: ["_value"])
    |> group()
    |> distinct()
```

#### KeyValues

KeyValues outputs a table with the input table's group key, plus two columns  `_key` and `_value` that correspond to unique (column, value) pairs from the input table.

KeyValues has the following properties:

| Name       | Type                         | Description                                                                                      |
| ----       | ----                         | -----------                                                                                      |
| keyColumns | []string                     | KeyColumns is a list of columns from which values are extracted.                                 |
| fn         | (schema: schema) -> []string | Fn is a schema function that may by used instead of `keyColumns` to identify the set of columns. |

Additional requirements:

* Only one of `keyColumns` or `fn` may be used in a single call.
* All columns indicated must be of the same type.
* Each input table must have all of the columns listed by the `keyColumns` parameter.

```
from(bucket: "telegraf/autogen")
    |> range(start: -30m)
    |> filter(fn: (r) => r._measurement == "cpu")
    |> keyValues(keyColumns: ["usage_idle", "usage_user"])
```

```
from(bucket: "telegraf/autogen")
    |> range(start: -30m)
    |> filter(fn: (r) => r._measurement == "cpu")
    |> keyValues(fn: (schema) => schema.columns |> filter(fn:(v) =>  v.label =~ /usage_.*/))
```

```
filterColumns = (fn) => (schema) => schema.columns |> filter(fn:(v) => fn(column:v))

from(bucket: "telegraf/autogen")
    |> range(start: -30m)
    |> filter(fn: (r) => r._measurement == "cpu")
    |> keyValues(fn: filterColumns(fn: (column) => column.label =~ /usage_.*/))
```

Examples:

Given the following input table with group key `["_measurement"]`:

    | _time | _measurement | _value | tagA |
    | ----- | ------------ | ------ | ---- |
    | 00001 | "m1"         | 1      | "a"  |
    | 00002 | "m1"         | 2      | "b"  |
    | 00003 | "m1"         | 3      | "c"  |
    | 00004 | "m1"         | 4      | "b"  |

`keyValues(keyColumns: ["tagA"])` produces the following table with group key ["_measurement"]:

    | _measurement | _key   | _value |
    | ------------ | ------ | ------ |
    | "m1"         | "tagA" | "a"    |
    | "m1"         | "tagA" | "b"    |
    | "m1"         | "tagA" | "c"    |

`keyColumns(keyColumns: ["tagB"])` produces the following error message:

    received table with columns [_time, _measurement, _value, tagA] not having key columns [tagB]


#### Window

Window groups records based on a time value.
New columns are added to uniquely identify each window and those columns are added to the group key of the output tables.

A single input record will be placed into zero or more output tables, depending on the specific windowing function.

By default the start boundary of a window will align with the Unix epoch (zero time) modified by the offset of the `location` option.

Window has the following properties:

| Name        | Type                                       | Description                                                                                                                                                                                                                                   |
| ----        | ----                                       | -----------                                                                                                                                                                                                                                   |
| every       | duration                                   | Every is the duration of time between windows. Defaults to `period`'s value. One of `every`, `period` or `intervals` must be provided.                                                                                                         |
| period      | duration                                   | Period is the duration of the window. Period is the length of each interval. It can be negative, indicating the start and stop boundaries are reversed. Defaults to `every`'s value. One of `every`, `period` or `intervals` must be provided. |
| offset      | duration                                   | Offset is the duration by which to shift the window boundaries. It can be negative, indicating that the offset goes backwards in time. Defaults to 0, which will align window end boundaries with the `every` duration.                                |
| intervals   | (start: time, stop: time) -> [...]interval | Intervals is a set of intervals to be used as the windows. One of `every`, `period` or `intervals` must be provided. When `intervals` is provided, `every`, `period`, and `offset` must be zero.                                              |
| timeColumn  | string                                     | TimeColumn is the name of the time column to use.  Defaults to `_time`.                                                                                                                                                                       |
| startColumn | string                                     | StartColumn is the name of the column containing the window start time. Defaults to `_start`.                                                                                                                                                 |
| stopColumn  | string                                     | StopColumn is the name of the column containing the window stop time. Defaults to `_stop`.                                                                                                                                                    |
| createEmpty | bool                                       | CreateEmpty specifies whether empty tables should be created. Defaults to `false`.

Example: 
```
from(bucket:"telegraf/autogen")
    |> range(start:-12h)
    |> window(every:10m)
    |> max()
```

```
window(every:1h) // window the data into 1 hour intervals
window(intervals: intervals(every:1d, period:8h, offset:9h)) // window the data into 8 hour intervals starting at 9AM every day.
```

#### Pivot

Pivot collects values stored vertically (column-wise) in a table and aligns them horizontally (row-wise) into logical sets.  

Pivot has the following properties:

| Name        | Type     | Description                                                                                    |
| ----        | ----     | -----------                                                                                    |
| rowKey      | []string | RowKey is the list of columns used to uniquely identify a row for the output.                  |
| columnKey   | []string | ColumnKey is the list of columns used to pivot values onto each row identified by the rowKey.  |
| valueColumn | string   | ValueColumn identifies the single column that contains the value to be moved around the pivot. |

The group key of the resulting table will be the same as the input tables, excluding the columns found in the `columnKey` and `valueColumn`.
This is because these columns are not part of the resulting output table. 

Any columns in the original table that are not referenced in the `rowKey` or the original table's group key will be dropped.

Every input row should have a 1:1 mapping to a particular row, column pair in the output table, determined by its values for the `rowKey` and `columnKey`.
In the case where more than one value is identified for the same row, column pair in the output, the last value
encountered in the set of table rows is taken as the result.

The output is constructed as follows:
 - The set of columns for the new table is the `rowKey` unioned with the group key, but excluding the columns indicated by the `columnKey` and the `valueColumn`.
 A new column is added to the set of columns for each unique value identified in the input by the `columnKey` parameter.
 The label of a new column is the concatenation of the values at `columnKey` (if the value is null, `"null"` is used) using `_` as a separator.
 - A new row is created for each unique value identified in the input by the `rowKey` parameter.
 - For each new row, values for group key columns stay the same, while values for new columns are determined from the input tables by the value in `valueColumn` at the row identified by the `rowKey` values and the new column's label.
 If no value is found, the value is set to null.
 
Example 1, align fields within each measurement that have the same timestamp:

 ```
  from(bucket:"test")
      |> range(start: 1970-01-01T00:00:00.000000000Z)
      |> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
 ```
 
Input:

|              _time             | _value | _measurement | _field |
|:------------------------------:|:------:|:------------:|:------:|
| 1970-01-01T00:00:00.000000001Z |   1.0  |     "m1"     |  "f1"  |
| 1970-01-01T00:00:00.000000001Z |   2.0  |     "m1"     |  "f2"  |
| 1970-01-01T00:00:00.000000001Z |  null  |     "m1"     |  "f3"  |
| 1970-01-01T00:00:00.000000001Z |   3.0  |     "m1"     |  null  |
| 1970-01-01T00:00:00.000000002Z |   4.0  |     "m1"     |  "f1"  |
| 1970-01-01T00:00:00.000000002Z |   5.0  |     "m1"     |  "f2"  |
|              null              |   6.0  |     "m1"     |  "f2"  |
| 1970-01-01T00:00:00.000000002Z |  null  |     "m1"     |  "f3"  |
| 1970-01-01T00:00:00.000000003Z |  null  |     "m1"     |  "f1"  |
| 1970-01-01T00:00:00.000000003Z |   7.0  |     "m1"     |  null  |
| 1970-01-01T00:00:00.000000004Z |   8.0  |     "m1"     |  "f3"  |

Output:

|              _time             | _measurement |  f1  |  f2  |  f3  | null |
|:------------------------------:|:------------:|:----:|:----:|:----:|:----:|
| 1970-01-01T00:00:00.000000001Z |     "m1"     |  1.0 |  2.0 | null |  3.0 |
| 1970-01-01T00:00:00.000000002Z |     "m1"     |  4.0 |  5.0 | null | null |
|               null             |     "m1"     | null |  6.0 | null | null |
| 1970-01-01T00:00:00.000000003Z |     "m1"     | null | null | null |  7.0 |
| 1970-01-01T00:00:00.000000004Z |     "m1"     | null | null |  8.0 | null |

Example 2, align fields and measurements that have the same timestamp.  
Note the effect of:
 - having null values in some `columnKey` value;
 - having more values for the same `rowKey` and `columnKey` value (the 11th row overrides the 10th, and so does the 15th with the 14th).

```
  from(bucket:"test")
      |> range(start: 1970-01-01T00:00:00.000000000Z)
      |> pivot(rowKey:["_time"], columnKey: ["_measurement", _field"], valueColumn: "_value")
 ```

Input:

|              _time             | _value | _measurement | _field |
|:------------------------------:|:------:|:------------:|:------:|
| 1970-01-01T00:00:00.000000001Z |   1.0  |     "m1"     |  "f1"  |
| 1970-01-01T00:00:00.000000001Z |   2.0  |     "m1"     |  "f2"  |
| 1970-01-01T00:00:00.000000001Z |   3.0  |     null     |  "f3"  |
| 1970-01-01T00:00:00.000000001Z |   4.0  |     null     |  null  |
| 1970-01-01T00:00:00.000000002Z |   5.0  |     "m1"     |  "f1"  |
| 1970-01-01T00:00:00.000000002Z |   6.0  |     "m1"     |  "f2"  |
| 1970-01-01T00:00:00.000000002Z |   7.0  |     "m1"     |  "f3"  |
| 1970-01-01T00:00:00.000000002Z |   8.0  |     null     |  null  |
|              null              |   9.0  |     "m1"     |  "f3"  |
| 1970-01-01T00:00:00.000000003Z |  10.0  |     "m1"     |  null  |
| 1970-01-01T00:00:00.000000003Z |  11.0  |     "m1"     |  null  |
| 1970-01-01T00:00:00.000000003Z |  12.0  |     "m1"     |  "f3"  |
| 1970-01-01T00:00:00.000000003Z |  13.0  |     null     |  null  |
|              null              |  14.0  |     "m1"     |  null  |
|              null              |  15.0  |     "m1"     |  null  |

Output:

|              _time             | m1_f1 | m1_f2 |  null_f3  | null_null | m1_f3 | m1_null |
|:------------------------------:|:-----:|:-----:|:---------:|:---------:|:-----:|:-------:|
| 1970-01-01T00:00:00.000000001Z |  1.0  |  2.0  |    3.0    |    4.0    |  null |  null   |
| 1970-01-01T00:00:00.000000002Z |  5.0  |  6.0  |   null    |    8.0    |  7.0  |  null   |
|              null              |  null |  null |   null    |    null   |  9.0  |  15.0   |
| 1970-01-01T00:00:00.000000003Z |  null |  null |   null    |   13.0    |  12.0 |  11.0   |

#### Join

Join merges two or more input streams, whose values are equal on a set of common columns, into a single output stream.
Null values are not considered equal when comparing column values.
The resulting schema is the union of the input schemas, and the resulting group key is the union of the input group keys.

Join has the following properties:

| Name   | Type     | Description                                                                         |
| ----   | ----     | -----------                                                                         |
| tables | object   | Tables is the map of streams to be joined.                                          |
| on     | []string | On is the list of columns on which to join.                                         |
| method | string   | Method must be one of: inner, cross, left, right, or full. Defaults to `"inner"`  . |

Both `tables` and `on` are required parameters.
The `on` parameter and the `cross` method are mutually exclusive.
Join currently only supports two input streams.

[IMPL#83](https://github.com/influxdata/flux/issues/83) Add support for joining more than 2 streams  
[IMPL#84](https://github.com/influxdata/flux/issues/84) Add support for different join types  

Example:

Given the following two streams of data:

* SF_Temperature

    | _time | _field | _value |
    | ----- | ------ | ------ |
    | 0001  | "temp" | 70     |
    | 0002  | "temp" | 75     |
    | 0003  | "temp" | 72     |

* NY_Temperature

    | _time | _field | _value |
    | ----- | ------ | ------ |
    | 0001  | "temp" | 55     |
    | 0002  | "temp" | 56     |
    | 0003  | "temp" | 55     |

And the following join query:

    join(tables: {sf: SF_Temperature, ny: NY_Temperature}, on: ["_time", "_field"])

The output will be:

| _time | _field | _value_ny | _value_sf |
| ----- | ------ |---------- | --------- |
| 0001  | "temp" | 55        | 70        |
| 0002  | "temp" | 56        | 75        |
| 0003  | "temp" | 55        | 72        |


##### output schema

The column schema of the output stream is the union of the input schemas, and the same goes for the output group key.
Columns that must be renamed due to ambiguity (i.e. columns that occur in more than one input stream) are renamed
according to the template `<column>_<table>`.

Example:

* SF_Temperature
* Group Key for table `{ _field }`

    | _time | _field | _value |
    | ----- | ------ | ------ |
    | 0001  | "temp" | 70     |
    | 0002  | "temp" | 75     |
    | 0003  | "temp" | 72     |

* NY_Temperature
* Group Key for all tables `{ _time, _field }`

    | _time | _field | _value |
    | ----- | ------ | ------ |
    | 0001  | "temp" | 55     |

    | _time | _field | _value |
    | ----- | ------ | ------ |
    | 0002  | "temp" | 56     |

    | _time | _field | _value |
    | ----- | ------ | ------ |
    | 0003  | "temp" | 55     |

`join(tables: {sf: SF_Temperature, ny: NY_Temperature}, on: ["_time"])` produces:

* Group Key for all tables `{ _time, _field_ny, _field_sf }`

    | _time | _field_ny | _field_sf | _value_ny | _value_sf |
    | ----- | --------- | --------- |---------- | --------- |
    | 0001  | "temp"    | "temp"    | 55        | 70        |

    | _time | _field_ny | _field_sf | _value_ny | _value_sf |
    | ----- | --------- | --------- |---------- | --------- |
    | 0002  | "temp"    | "temp"    | 56        | 75        |

    | _time | _field_ny | _field_sf | _value_ny | _value_sf |
    | ----- | --------- | --------- |---------- | --------- |
    | 0003  | "temp"    | "temp"    | 55        | 72        |

#### Union

Union concatenates two or more input streams into a single output stream.  In tables that have identical
schema and group keys, contents of the tables will be concatenated in the output stream.  The output schemas of 
the Union operation shall be the union of all input schemas.

Union does not preserve the sort order of the rows within tables. A sort operation may be added if a specific sort order is needed.

Union has the following properties:

| Name   | Type     | Description                                                                         |
| ----   | ----     | -----------                                                                         |
| tables | []stream | Tables specifies the streams to union together. There must be at least two streams. |

For example, given this stream, `SF_Weather` with group key `"_field"` on both tables:

   |  _time |  _field |  _value |
   | ----- | ------ | ------ |
   | 0001  | "temp" | 70 |
   | 0002  | "temp" | 75 |

   | _time | _field | _value |
   | ----- | ------ | ------ |
   | 0001  | "humidity" | 81 |
   | 0002  | "humidity" | 82 |

And this stream, `NY_Weather`, also with group key `"_field"` on both tables:

   | _time | _field | _value |
   | ----- | ------ | ------ |
   | 0001  | "temp" | 55 |
   | 0002  | "temp" | 56 |

   | _time | _field | _value |
   | ----- | ------ | ------ |
   | 0001  | "pressure" | 29.82 |
   | 0002  | "pressure" | 30.01 |

`union(tables: [SF_Weather, NY_Weather])` produces this stream (whose tables are grouped by `"_field"`):

   | _time | _field | _value |
   | ----- | ------ | ------ |
   | 0001  | "temp" | 70 |
   | 0002  | "temp" | 75 |
   | 0001  | "temp" | 55 |
   | 0002  | "temp" | 56 |

   | _time | _field | _value |
   | ----- | ------ | ------ |
   | 0001  | "humidity" | 81 |
   | 0002  | "humidity" | 82 |

   | _time | _field | _value |
   | ----- | ------ | ------ |
   | 0001  | "pressure" | 29.82 |
   | 0002  | "pressure" | 30.01 |

#### Unique

Unique returns a table with unique values in a specified column.
In the case there are multiple rows taking on the same value in the provided column, the first row is kept and the remaining rows are discarded.

Unique has the following properties:

| Name   | Type   | Description                                                 |
| ----   | ----   | -----------                                                 |
| column | string | Column that is to have unique values. Defaults to `_value`. |

#### Cumulative sum

Cumulative sum computes a running sum for non null records in the table.
The output table schema will be the same as the input table.

Cumulative sum has the following properties:

| Name    | Type     | Description                                                                  |
| ----    | ----     | -----------                                                                  |
| columns | []string | Columns is a list of columns on which to operate.  Defaults to `["_value"]`. |

Example:

```
from(bucket: "telegraf/autogen")
    |> range(start: -5m)
    |> filter(fn: (r) => r._measurement == "disk" and r._field == "used_percent")
    |> cumulativeSum(columns: ["_value"])
```

#### Derivative

Derivative computes the time based difference between subsequent non-null records.
This function will return an error if values in the time column are null or not sorted in
ascending order.
If there are multiple rows with the same time value, only the first row will be used to
compute the derivative.

Derivative has the following properties:

| Name        | Type     | Description                                                                                                                                                                                       |
| ----        | ----     | -----------                                                                                                                                                                                       |
| unit        | duration | Unit is the time duration to use for the result.  Defaults to `1s`.                                                                                                                               |
| nonNegative | bool     | NonNegative indicates if the derivative is allowed to be negative. If a value is encountered which is less than the previous value, then the derivative will be null for that row.                     |
| columns     | []string | Columns is a list of columns on which to compute the derivative Defaults to `["_value"]`.                                                                                                         |
| timeColumn  | string   | TimeColumn is the column name for the time values.  Defaults to `_time`.                                                                                                                          |

```
from(bucket: "telegraf/autogen")
    |> range(start: -5m)
    |> filter(fn: (r) => r._measurement == "disk" and r._field == "used_percent")
    |> derivative(nonNegative: true, columns: ["used_percent"])
```

#### Difference

Difference computes the difference between subsequent records.  
Every user-specified column of numeric type is subtracted while others are kept intact.

Difference has the following properties:

| Name        | Type     | Description                                                                                                                                                 |
| ----        | ----     | -----------                                                                                                                                                 |
| nonNegative | bool     | NonNegative indicates if the difference is allowed to be negative. If a value is encountered which is less than the previous value then the result is null. |
| columns     | []string | Columns is a list of columns on which to compute the difference. Defaults to `["_value"]`.                                                                  |

Rules for subtracting values for numeric types:

 - the difference between two non-null values is their algebraic difference; or null, if the result is negative and `nonNegative: true`;
 - null minus some value is always null;
 - some value `v` minus null is `v` minus the last non-null value seen before `v`; or null if `v` is the first non-null value seen.

Example of difference:

| _time |   A  |   B  |   C  | tag |
|:-----:|:----:|:----:|:----:|:---:|
|  0001 | null |   1  |   2  |  tv |
|  0002 |   6  |   2  | null |  tv |
|  0003 |   4  |   2  |   4  |  tv |
|  0004 |  10  |  10  |   2  |  tv |
|  0005 | null | null |   1  |  tv |

Result (`nonNegative: false`):

| _time |   A  |   B  |   C  | tag |
|:-----:|:----:|:----:|:----:|:---:|
|  0002 | null |   1  | null |  tv |
|  0003 |  -2  |   0  |   2  |  tv |
|  0004 |   6  |   8  |  -2  |  tv |
|  0005 | null | null |  -1  |  tv |

Result (`nonNegative: true`):

| _time |   A  |   B  |   C  | tag |
|:-----:|:----:|:----:|:----:|:---:|
|  0002 | null |   1  | null |  tv |
|  0003 | null |   0  |   2  |  tv |
|  0004 |   6  |   8  | null |  tv |
|  0005 | null | null | null |  tv |

Example of script:

```
from(bucket: "telegraf/autogen")
    |> range(start: -5m)
    |> filter(fn: (r) => r._measurement == "cpu" and r._field == "usage_user")
    |> difference()
```

#### Increase

Increase returns the total non-negative difference between values in a table.
A main usage case is tracking changes in counter values which may wrap over time when they hit
a threshold or are reset. In the case of a wrap/reset,
we can assume that the absolute delta between two points will be at least their non-negative difference.

```
increase = (tables=<-, columns=["_value"]) =>
    tables
        |> difference(nonNegative: true, columns:columns)
        |> cumulativeSum()
```

Example:

Given the following input table.

    | _time | _value |
    | ----- | ------ |
    | 00001 | 1      |
    | 00002 | 5      |
    | 00003 | 3      |
    | 00004 | 4      |

`increase()` produces the following table.

    | _time | _value |
    | ----- | ------ |
    | 00002 | 4      |
    | 00003 | 7      |
    | 00004 | 8      |

#### Distinct

Distinct produces the unique values for a given column. Null is considered its own distinct value if it is present.

Distinct has the following properties:

| Name   | Type   | Description                                                                  |
| ----   | ----   | -----------                                                                  |
| column | string | Column is the column on which to track unique values.  Defaults to `_value`. |

Example:

```
from(bucket: "telegraf/autogen")
	|> range(start: -5m)
	|> filter(fn: (r) => r._measurement == "cpu")
	|> distinct(column: "host")
```


#### TimeShift

TimeShift adds a fixed duration to time columns.
The output table schema is the same as the input table.
If the time is null, the time will continue to be null.

TimeShift has the following properties:

| Name     | Type     | Description                                                                                        |
| ----     | ----     | -----------                                                                                        |
| duration | duration | Duration is the amount to add to each time value.  May be a negative duration.                     |
| columns  | []string | Columns is list of all columns that should be shifted. Defaults to `["_start", "_stop", "_time"]`. |

Example:

```
from(bucket: "telegraf/autogen")
	|> range(start: -5m)
	|> timeShift(duration: 1000h)
```

#### StateCount

StateCount computes the number of consecutive records in a given state.
The state is defined via a user-defined predicate. For each consecutive point for
which the predicate evaluates as true, the state count will be incremented.
When a point evaluates as false, the state count is reset.

The state count will be added as an additional column to each record. If the
expression evaluates as false, the value will be -1. If the expression
generates an error during evaluation, the point is discarded, and does not
affect the state count.

StateCount has the following parameters:

| Name   | Type                | Description                                                                                  |
| ----   | ----                | -----------                                                                                  |
| fn     | (r: record) -> bool | Fn is a function that returns true when the record is in the desired state.                  |
| column | string              | Column is the name of the column to use to output the state count. Defaults to `stateCount`. |

Example:

```
from(bucket: "telegraf/autogen")
    |> range(start: 2018-05-22T19:53:26Z)
    |> stateCount(fn:(r) => r._value > 80)
```

#### StateDuration

StateDuration computes the duration of a given state.
The state is defined via a user-defined predicate. For each consecutive point for
which the predicate evaluates as true, the state duration will be
incremented by the duration between points. When a point evaluates as false,
the state duration is reset.

The state duration will be added as an additional column to each record.
If the expression evaluates as false, the value will be -1. If the expression
generates an error during evaluation, the point is discarded, and does not
affect the state duration.

Note that as the first point in the given state has no previous point, its
state duration will be 0.

The duration is represented as an integer in the units specified.

StateDuration requires sorted and not-null timestamps. So, if one of this requirements
is not met, it returns an error.

StateDuration has the following parameters:

| Name       | Type                | Description                                                                                     |
| ----       | ----                | -----------                                                                                     |
| fn         | (r: record) -> bool | Fn is a function that returns true when the record is in the desired state.                     |
| column     | string              | Column is the name of the column to use to output the state value. Defaults to `stateDuration`. |
| timeColumn | string              | TimeColumn is the name of the column used to extract timestamps. Defaults to `_time`.           |
| unit       | duration            | Unit is the dimension of the output value. Defaults to `1s`.                                    |

Example:

```
from(bucket: "telegraf/autogen")
    |> range(start: 2018-05-22T19:53:26Z)
    |> stateDuration(fn:(r) => r._value > 80)
```

#### To

The To operation takes data from a stream and writes it to a bucket.
To has the following properties:

| Name       | Type                  | Description                                                                                                                                                                                                                        |
| ----       | ----                  | -----------                                                                                                                                                                                                                        |
| bucket     | string                | Bucket is the bucket name into which data will be written.                                                                                                                                                                         |
| bucketID   | string                | BucketID is the bucket ID into which data will be written.                                                                                                                                                                         |
| org        | string                | Org is the organization name of the bucket.                                                                                                                                                                                        |
| orgID      | string                | OrgID is the organization ID of the bucket.                                                                                                                                                                                        |
| host       | string                | Host is the location of a remote host to write to. Defaults to `""`.                                                                                                                                                               |
| token      | string                | Token is the authorization token to use when writing to a remote host. Defaults to `""`.                                                                                                                                           |
| timeColumn | string                | TimeColumn is the name of the time column of the output.  Defaults to `"_time"`.                                                                                                                                                   |
| tagColumns | []string              | TagColumns is a list of columns to be used as tags in the output. Defaults to all columns of type string, excluding all value columns and the `_field` column if present.                                                          |
| fieldFn    | (r: record) -> record | Function that takes a record from the input table and returns an object. For each record from the input table `fieldFn` returns on object that maps output field key to output value. Default: `(r) => ({ [r._field]: r._value })` |

TODO(nathanielc): The fieldFn is not valid and needs to change. It uses dynamic object keys which is not allowed.

Either `bucket` or `bucketID` is required.
Both are mutually exclusive.
Similarly `org` and `orgID` are mutually exclusive and only required when writing to a remote host.
Both `host` and `token` are optional parameters, however if `host` is specified, `token` is required.


For example, given the following table:

| _time | _start | _stop | _measurement | _field | _value |
| ----- | ------ | ----- | ------------ | ------ | ------ |
| 0005  | 0000   | 0009  | "a"          | "temp" | 100.1  |
| 0006  | 0000   | 0009  | "a"          | "temp" | 99.3   |
| 0007  | 0000   | 0009  | "a"          | "temp" | 99.9   |

The default `to` operation `to(bucket:"my-bucket", org:"my-org")` is equivalent to writing the above data using the following line protocol:

```
_measurement=a temp=100.1 0005
_measurement=a temp=99.3 0006
_measurement=a temp=99.9 0007
```

For an example overriding `to`'s default settings, given the following table:

| _time | _start | _stop | tag1 | tag2 | hum | temp |
| ----- | ------ | ----- | ---- | ---- | ---- | ---- |
| 0005  | 0000   | 0009  | "a"  | "b"  | 55.3 | 100.1  |
| 0006  | 0000   | 0009  | "a"  | "b"  | 55.4 | 99.3   |
| 0007  | 0000   | 0009  | "a"  | "b"  | 55.5 | 99.9   |

The operation `to(bucket:"my-bucket", org:"my-org", tagColumns:["tag1"], fieldFn: (r) => return {"hum": r.hum, "temp": r.temp})` is equivalent to writing the above data using the following line protocol:

```
_tag1=a hum=55.3,temp=100.1 0005
_tag1=a hum=55.4,temp=99.3 0006
_tag1=a hum=55.5,temp=99.9 0007
```

**Note:** The `to` function produces side effects.

#### Top/Bottom

Top and Bottom sort a table and limits the table to only n records.

Top and Bottom have the following parameters:

| Name    | Type     | Description                                     |
| ----    | ----     | -----------                                     |
| n       | int      | N is the number of records to keep.             |
| columns | []string | Columns provides the sort order for the tables. |

Example:

    from(bucket:"telegraf/autogen")
        |> range(start: -5m)
        |> filter(fn:(r) => r._measurement == "net" and r._field == "bytes_sent")
        |> top(n:10, columns:["_value"])

#### Contains 

Tests whether a value is a member of a set.  

Contains has the following parameters: 

| Name    | Type                                          | Description                  |
| ----    | ----                                          | -----------                  |
| value   | bool, int, uint, float, string, time          | The value to search for.     |
| set     | array of bool, int, uint, float, string, time | The set of values to search. |

Example: 
    `contains(value:1, set:[1,2,3])` will return `true`.   

#### Type conversion operations

##### toBool

Convert a value to a bool.

Example: `from(bucket: "telegraf") |> filter(fn:(r) => r._measurement == "mem" and r._field == "used") |> toBool()`

The function `toBool` is defined as `toBool = (tables=<-) => tables |> map(fn:(r) => bool(v:r._value))`.
If you need to convert other columns use the `map` function directly with the `bool` function.

##### toInt

Convert a value to a int.

Example: `from(bucket: "telegraf") |> filter(fn:(r) => r._measurement == "mem" and r._field == "used") |> toInt()`

The function `toInt` is defined as `toInt = (tables=<-) => tables |> map(fn:(r) => int(v:r._value))`.
If you need to convert other columns use the `map` function directly with the `int` function.

##### toFloat

Convert a value to a float.

Example: `from(bucket: "telegraf") |> filter(fn:(r) => r._measurement == "mem" and r._field == "used") |> toFloat()`

The function `toFloat` is defined as `toFloat = (tables=<-) => tables |> map(fn:(r) => float(v:r._value))`.
If you need to convert other columns use the `map` function directly with the `float` function.

##### toDuration

Convert a value to a duration.

Example: `from(bucket: "telegraf") |> filter(fn:(r) => r._measurement == "mem" and r._field == "used") |> toDuration()`

The function `toDuration` is defined as `toDuration = (tables=<-) => tables |> map(fn:(r) => duration(v:r._value))`.
If you need to convert other columns use the `map` function directly with the `duration` function.

##### toString

Convert a value to a string.

Example: `from(bucket: "telegraf") |> filter(fn:(r) => r._measurement == "mem" and r._field == "used") |> toString()`

The function `toString` is defined as `toString = (tables=<-) => tables |> map(fn:(r) => string(v:r._value))`.
If you need to convert other columns use the `map` function directly with the `string` function.

##### toTime

Convert a value to a time.

Example: `from(bucket: "telegraf") |> filter(fn:(r) => r._measurement == "mem" and r._field == "used") |> toTime()`

The function `toTime` is defined as `toTime = (tables=<-) => tables |> map(fn:(r) => time(v:r._value))`.
If you need to convert other columns use the `map` function directly with the `time` function.

##### toUInt

Convert a value to a uint.

Example: `from(bucket: "telegraf") |> filter(fn:(r) => r._measurement == "mem" and r._field == "used") |> toUInt()`

The function `toUInt` is defined as `toUInt = (tables=<-) => tables |> map(fn:(r) => uint(v:r._value))`.
If you need to convert other columns use the `map` function directly with the `uint` function.


[IMPL#242](https://github.com/influxdata/platform/issues/242) Update specification around type conversion functions.

#### String operations

##### trim

Remove leading and trailing characters specified in cutset from a string.

Example: `trim(v: ".abc.", cutset: ".")` returns the string `abc`.

##### trimSpace

Remove leading and trailing spaces from a string.

Example: `trimSpace(v: "  abc  ")` returns the string `abc`.

##### title

Convert a string to title case.

Example: `title(v: "a flux of foxes")` returns the string `A Flux Of Foxes`.

##### toUpper

Convert a string to upper case.

Example: `toUpper(v: "koala")` returns the string `KOALA`.

##### toLower

Convert a string to lower case.

Example: `toLower(v: "KOALA")` returns the string `koala`.

### Composite data types

A composite data type is a collection of primitive data types that together have a higher meaning.

### Triggers

A trigger is associated with a table and contains logic for when it should fire.
When a trigger fires its table is materialized.
Materializing a table makes it available for any down stream operations to consume.
Once a table is materialized it can no longer be modified.

Triggers can fire based on these inputs:

| Input                   | Description                                                                                       |
| -----                   | -----------                                                                                       |
| Current processing time | The current processing time is the system time when the trigger is being evaluated.               |
| Watermark time          | The watermark time is a time where it is expected that no data will arrive that is older than it. |
| Record count            | The number of records currently in the table.                                                     |
| Group key value         | The group key value of the table.                                                                 |

Additionally triggers can be _finished_, which means that they will never fire again.
Once a trigger is finished, its associated table is deleted.

Currently all tables use an _after watermark_ trigger which fires only once the watermark has exceeded the `_stop` value of the table and then is immediately finished.

Data sources are responsible for informing about updates to the watermark.

[IMPL#240](https://github.com/influxdata/platform/issues/240) Make trigger support not dependent on specific columns

### Execution model

A query specification defines what data and operations to perform.
The execution model reserves the right to perform those operations as efficiently as possible.
The execution model may rewrite the query in anyway it sees fit while maintaining correctness.

## Request and Response Formats

Included with the specification of the language and execution model, is a specification of how to submit queries and read their responses over HTTP.

### Request format

To submit a query for execution, make an HTTP POST request to the `/v1/query` endpoint.

The POST request may either submit parameters as the POST body or a subset of the parameters as URL query parameters.
The following parameters are supported:

| Parameter | Description                                                                                                                                       |
| --------- | -----------                                                                                                                                       |
| query     | Query is Flux text describing the query to run.  Only one of `query` or `spec` may be specified. This parameter may be passed as a URL parameter. |
| spec      | Spec is a query specification. Only one of `query` or `spec` may be specified.                                                                    |
| dialect   | Dialect is an object defining the options to use when encoding the response.                                                                      |


When using the POST body to submit the query the `Content-Type` HTTP header must contain the name of the request encoding being used.

Supported request content types:

* `application/json` - Use a JSON encoding of parameters.



Multiple response content types will be supported.
The desired response content type is specified using the `Accept` HTTP header on the request.
Each response content type will have its own dialect options.

Supported response encodings:

* `test/csv` - Corresponds with the MIME type specified in RFC 4180.
    Details on the encoding format are specified below.

If no `Accept` header is present it is assumed that `text/csv` was specified.
The HTTP header `Content-Type` of the response will specify the encoding of the response.

#### Examples requests

Make a request using a query string and URL query parameters:

```
POST /v1/query?query=%20from%28db%3A%22mydatabse%22%29%20%7C%3E%20last%28%29 HTTP/1.1
```

Make a request using a query string and the POST body as JSON:

```
POST /v1/query


{
    "query": "from(bucket:\"mydatabase/autogen\") |> last()"
}
```

Make a request using a query specification and the POST body as JSON:

```
POST /v1/query


{
    "spec": {
      "operations": [
        {
          "kind": "from",
          "id": "from0",
          "spec": {
            "db": "mydatabase"
          }
        },
        {
          "kind": "last",
          "id": "last1",
          "spec": {
            "column": ""
          }
        }
      ],
      "edges": [
        {
          "parent": "from0",
          "child": "last1"
        }
      ],
      "resources": {
        "priority": "high",
        "concurrency_quota": 0,
        "memory_bytes_quota": 0
      }
    }
}
```
Make a request using a query string and the POST body as JSON.
Dialect options are specified for the `text/csv` format.
See below for details on specific dialect options.

```
POST /v1/query


{
    "query": "from(bucket:\"mydatabase/autogen\") |> last()",
    "dialect" : {
        "header": true,
        "annotations": ["datatype"]
    }
}
```


### Response format

#### CSV

The result of a query is any number of named streams.
As a stream consists of multiple tables each table is encoded as CSV textual data.
CSV data should be encoded using UTF-8, and should be in Unicode Normal Form C as defined in [UAX15](https://www.w3.org/TR/2015/REC-tabular-data-model-20151217/#bib-UAX15).
Line endings must be CRLF as defined by the `text/csv` MIME type in RFC 4180

Each table may have the following rows:

* annotation rows - a set of rows describing properties about the columns of the table.
* header row - a single row that defines the column labels.
* record rows, a set of rows containing the record data, one record per row.

In addition to the columns on the tables themselves three additional columns may be added to the CSV table.

* annotation - Contains the name of an annotation.
    This column is optional, if it exists it is always the first column.
    The only valid values for the column are the list of supported annotations or an empty value.
* result - Contains the name of the result as specified by the query.
* table - Contains a unique ID for each table within a result.

Columns support the following annotations:

* datatype - a description of the type of data contained within the column.
* group - a boolean flag indicating if the column is part of the table's group key.
* default - a default value to be used for rows whose string value is the empty string.
* null - a value indicating that data is missing.

##### Multiple tables

Multiple tables may be encoded into the same file or data stream.
The table column indicates the table a row belongs to.
All rows for a table must be contiguous.

It is possible that multiple tables in the same result do not share a common table scheme.
It is also possible that a table has no records.
In such cases an empty row delimits a new table boundary and new annotations and header rows follow.
The empty row acts like a delimiter between two independent CSV files that have been concatenated together.

In the case were a table has no rows the `default` annotation is used to provide the values of the group key.

##### Multiple results

Multiple results may be encoded into the same file or data stream.
An empty row always delimits separate results within the same file.
The empty row acts like a delimiter between two independent CSV files that have been concatenated together.

##### Annotations

Annotations rows are prefixed with a comment marker.
The first column contains the name of the annotation being defined.
The subsequent columns contain the value of the annotation for the respective columns.

The `datatype` annotation specifies the data types of the remaining columns.
The possible data types are:

| Datatype     | Flux type | Description                                                                          |
| --------     | --------- | -----------                                                                          |
| boolean      | bool      | a truth value, one of "true" or "false"                                              |
| unsignedLong | uint      | an unsigned 64-bit integer                                                           |
| long         | int       | a signed 64-bit integer                                                              |
| double       | float     | a IEEE-754 64-bit floating-point number                                              |
| string       | string    | a UTF-8 encoded string                                                               |
| base64Binary | bytes     | a base64 encoded sequence of bytes as defined in RFC 4648                            |
| dateTime     | time      | an instant in time, may be followed with a colon `:` and a description of the format |
| duration     | duration  | a length of time represented as an unsigned 64-bit integer number of nanoseconds     |

The `group` annotation specifies if the column is part of the table's group key.
Possible values are `true` or `false`.

The `default` annotation specifies a default value, if it exists, for each column.

In order to fully encode a table with its group key the `datatype`, `group` and `default` annotations must be used.

The `null` annotation specifies the string value that indicates a missing value.
When the `null` annotation is not specified the empty string value is the `null` value for the column.
It is not possible to encode/decode a non-null string value that is the same as the `null` annotation value for columns of type string.

When the `default` annotation value of a column is the same as the `null` annotation value of a column, it is interpreted as the column's default value is null.

##### Errors

When an error occurs during execution a table will be returned with the first column label as `error` and the second column label as `reference`.
The error's properties are contained in the second row of the table.
The `error` column contains the error message and the `reference` column contains a unique reference code that can be used to get further information about the problem.

When an error occurs before any results are materialized then the HTTP status code will indicate an error and the error details will be encoded in the csv table.
When an error occurs after some results have already been sent to the client the error will be encoded as the next table and the rest of the results will be discarded.
In such a case the HTTP status code cannot be changed and will remain as 200 OK.

Example error encoding without annotations:

```
error,reference
Failed to parse query,897
```

##### Dialect options

The CSV response format support the following dialect options:


| Option        | Description                                                                                                                                             |
| ------        | -----------                                                                                                                                             |
| header        | Header is a boolean value, if true the header row is included, otherwise its is omitted. Defaults to true.                                              |
| delimiter     | Delimiter is a character to use as the delimiting value between columns.  Defaults to ",".                                                              |
| quoteChar     | QuoteChar is a character to use to quote values containing the delimiter. Defaults to `"`.                                                              |
| annotations   | Annotations is a list of annotations that should be encoded. If the list is empty the annotation column is omitted entirely. Defaults to an empty list. |
| commentPrefix | CommentPrefix is a string prefix to add to comment rows. Defaults to "#". Annotations are always comment rows.                                          |


##### Examples

For context the following example tables encode fictitious data in response to this query:

    from(bucket:"mydb/autogen")
        |> range(start:2018-05-08T20:50:00Z, stop:2018-05-08T20:51:00Z)
        |> group(columns:["_start","_stop", "region", "host"])
        |> mean()
        |> group(columns:["_start","_stop", "region"])
        |> yield(name:"mean")


Example encoding with of a single table with no annotations:

```
result,table,_start,_stop,_time,region,host,_value
mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:00Z,east,A,15.43
mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:20Z,east,B,59.25
mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:40Z,east,C,52.62
```


Example encoding with two tables in the same result with no annotations:

```
result,table,_start,_stop,_time,region,host,_value
mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:00Z,east,A,15.43
mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:20Z,east,B,59.25
mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:40Z,east,C,52.62
mean,1,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:00Z,west,A,62.73
mean,1,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:20Z,west,B,12.83
mean,1,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:40Z,west,C,51.62
```

Example encoding with two tables in the same result with no annotations and no header row:

```
mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:00Z,east,A,15.43
mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:20Z,east,B,59.25
mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:40Z,east,C,52.62
mean,1,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:00Z,west,A,62.73
mean,1,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:20Z,west,B,12.83
mean,1,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:40Z,west,C,51.62
```

Example encoding with two tables in the same result with the datatype annotation:

```
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,double
,result,table,_start,_stop,_time,region,host,_value
,mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:00Z,east,A,15.43
,mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:20Z,east,B,59.25
,mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:40Z,east,C,52.62
,mean,1,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:00Z,west,A,62.73
,mean,1,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:20Z,west,B,12.83
,mean,1,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:40Z,west,C,51.62
```

Example encoding with two tables in the same result with the datatype and group annotations:

```
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,double
#group,false,false,true,true,false,true,false,false
,result,table,_start,_stop,_time,region,host,_value
,mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:00Z,east,A,15.43
,mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:20Z,east,B,59.25
,mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:40Z,east,C,52.62
,mean,1,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:00Z,west,A,62.73
,mean,1,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:20Z,west,B,12.83
,mean,1,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:40Z,west,C,51.62
```

Example encoding with two tables with differing schemas in the same result with the datatype and group annotations:

```
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,double
#group,false,false,true,true,false,true,false,false
,result,table,_start,_stop,_time,region,host,_value
,mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:00Z,east,A,15.43
,mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:20Z,east,B,59.25
,mean,0,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:40Z,east,C,52.62

#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,double
#group,false,false,true,true,false,true,false,false
,result,table,_start,_stop,_time,location,device,min,max
,mean,1,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:00Z,USA,5825,62.73,68.42
,mean,1,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:20Z,USA,2175,12.83,56.12
,mean,1,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:40Z,USA,6913,51.62,54.25
```

Example error encoding with the datatype annotation:

```
#datatype,string,long
,error,reference
,Failed to parse query,897
```

Example error encoding with after a valid table has already been encoded.

```
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,double
,result,table,_start,_stop,_time,region,host,_value
,mean,1,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:00Z,west,A,62.73
,mean,1,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:20Z,west,B,12.83
,mean,1,2018-05-08T20:50:00Z,2018-05-08T20:51:00Z,2018-05-08T20:50:40Z,west,C,51.62

#datatype,string,long
,error,reference
,query terminated: reached maximum allowed memory limits,576
```

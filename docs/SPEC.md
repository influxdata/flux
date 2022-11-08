# Flux Specification

The following document specifies the Flux language and query execution.

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

[The Unicode Standard 8.0](https://www.unicode.org/versions/Unicode8.0.0/), Section 4.5 "General Category" defines a set of character categories.
Flux treats all characters in any of the Letter categories Lu, Ll, Lt, Lm, or Lo as Unicode letters, and those in the Number category Nd as Unicode digits.

#### Letters and digits

The underscore character _ (U+005F) is considered a letter.

    letter        = unicode_letter | "_" .
    decimal_digit = "0" … "9" .

### Lexical Elements

#### Comments

Comments serve as documentation.
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

    and import  option   if
    or  package builtin  then
    not return  testcase else exists

#### Operators

The following character sequences represent operators:

    +   ==   !=   (   )   =>
    -   <    !~   [   ]   ^
    *   >    =~   {   }   ?
    /   <=   =    ,   :   "
    %   >=   <-   .   |>  @

##### Integer literals

An integer literal is a sequence of digits representing an integer value.
Only decimal integers are supported.

    int_lit     = "0" | decimal_lit .
    decimal_lit = ( "1" … "9" ) { decimal_digit } .

Examples:

    0
    42
    317316873

Errors:

    0123 // invalid leading 0

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

#### Duration literals


A duration literal is a representation of a length of time.
It has an integer part and a duration unit part.
Multiple durations may be specified together and the resulting duration is the sum of each smaller part.
When several durations are specified together, larger units must appear before smaller ones, and there can be no repeated units.

    duration_lit  = { duration_magnitude duration_unit } .
    duration_magnitude = decimal_digit { decimal_digit }
    duration_unit = "y" | "mo" | "w" | "d" | "h" | "m" | "s" | "ms" | "us" | "µs" | "ns" .

| Units    | Meaning                                 | Base |
| -----    | -------                                 | ---- |
| y        | year (12 months)                        | mo   |
| mo       | month                                   | mo   |
| w        | week (7 days)                           | ns   |
| d        | day                                     | ns   |
| h        | hour (60 minutes)                       | ns   |
| m        | minute (60 seconds)                     | ns   |
| s        | second                                  | ns   |
| ms       | milliseconds (1 thousandth of a second) | ns   |
| us or µs | microseconds (1 millionth of a second)  | ns   |
| ns       | nanoseconds (1 billionth of a second)   | ns   |

Durations represent a length of time.
Lengths of time are dependent on specific instants in time they occur and as such, durations do not represent a fixed amount of time.
A duration is composed of a tuple of months and nanoseconds along with whether the duration is positive or negative.
Each duration unit corresponds to one of these two base units.
It is possible to compose a duration of multiple base units.

Durations cannot be combined by addition and subtraction.
This is because all of the values in the tuple must have the same sign and it is not possible to guarantee that is true when addition and subtraction are allowed.
Durations can be multiplied by any integer value.
The unary negative operator is the equivalent of multiplying the duration by -1.
These operations are performed on each time unit independently.

Examples:

    1s
    10d
    1h15m  // 1 hour and 15 minutes
    5w
    1mo5d  // 1 month and 5 days
    -1mo5d // negative 1 month and 5 days

Durations can be added to date times to produce a new date time.
Addition and subtraction of durations to date times applies months and nanoseconds in that order.
When months are added to a date times and the resulting date is past the end of the month, the day is rolled back to the last day of the month.
Of note is that addition and subtraction of durations to date times does not commute.


Examples:

    import "date"

    date.add(d: 1d,  to: 2018-01-01T00:00:00Z) // 2018-01-02T00:00:00Z
    date.add(d: 1mo, to: 2018-01-01T00:00:00Z) // 2018-02-01T00:00:00Z
    date.add(d: 2mo, to: 2018-01-01T00:00:00Z) // 2018-03-01T00:00:00Z
    date.add(d: 2mo, to: 2018-01-31T00:00:00Z) // 2018-03-31T00:00:00Z
    date.add(d: 2mo, to: 2018-02-28T00:00:00Z) // 2018-04-28T00:00:00Z
    date.add(d: 1mo, to: 2018-01-31T00:00:00Z) // 2018-02-28T00:00:00Z, February 31th is rolled back to the last day of the month, February 28th in 2018.

    // Addition and subtraction of durations to date times does not commute
    date.add(d: 1d, to: date.add(d: 1mo, to: 2018-02-28T00:00:00Z))   // 2018-03-29T00:00:00Z
    date.add(d: 1mo, to: date.add(d: 1d, to: 2018-02-28T00:00:00Z))   // 2018-04-01T00:00:00Z
    date.sub(d: 1d, from: date.add(d: 2mo, to: 2018-01-01T00:00:00Z)) // 2018-02-28T00:00:00Z
    date.add(d: 3mo, to: date.sub(d: 1d, from: 2018-01-01T00:00:00Z)) // 2018-03-31T00:00:00Z
    date.add(d: 1mo, to: date.add(d: 1mo, to: 2018-01-31T00:00:00Z))  // 2018-03-28T00:00:00Z
    date.add(d: 2mo, to: 2018-01-31T00:00:00Z)                        // 2018-03-31T00:00:00Z

    // Addition and subtraction of durations to date times applies months and nanoseconds in that order.
    date.add(d: 2d, to: date.add(d: 1mo, to: 2018-01-28T00:00:00Z))  // 2018-03-02T00:00:00Z
    date.add(d: 1mo2d, to: 2018-01-28T00:00:00Z)                     // 2018-03-02T00:00:00Z
    date.add(d: 1mo, to: date.add(d: 2d, to: 2018-01-28T00:00:00Z))  // 2018-02-28T00:00:00Z, explicit add of 2d first changes the result
    date.add(d: 2mo2d, to: 2018-02-01T00:00:00Z)                     // 2018-04-03T00:00:00Z
    date.add(d: 1mo30d, to: 2018-01-01T00:00:00Z)                    // 2018-03-03T00:00:00Z, Months are applied first to get February 1st, then days are added resulting in March 3 in 2018.
    date.add(d: 1mo1d, to: 2018-01-31T00:00:00Z)                     // 2018-03-01T00:00:00Z, Months are applied first to get February 28th, then days are added resulting in March 1 in 2018.

    // Multiplication and addition of durations to date times
    date.add(d: date.scale(d:1mo, n:1), to: 2018-01-01T00:00:00Z)  // 2018-02-01T00:00:00Z
    date.add(d: date.scale(d:1mo, n:2), to: 2018-01-01T00:00:00Z)  // 2018-03-01T00:00:00Z
    date.add(d: date.scale(d:1mo, n:3), to: 2018-01-01T00:00:00Z)  // 2018-04-01T00:00:00Z
    date.add(d: date.scale(d:1mo, n:1), to: 2018-01-31T00:00:00Z)  // 2018-02-28T00:00:00Z
    date.add(d: date.scale(d:1mo, n:2), to: 2018-01-31T00:00:00Z)  // 2018-03-31T00:00:00Z
    date.add(d: date.scale(d:1mo, n:3), to: 2018-01-31T00:00:00Z)  // 2018-04-30T00:00:00Z

#### Date and time literals

A date and time literal represents a specific moment in time.
It has a date part, a time part and a time offset part.
The format follows the RFC 3339 specification.
The time is optional. When it is omitted the time is assumed to be midnight UTC.

    date_time_lit     = date [ "T" time ] .
    date              = year_lit "-" month "-" day .
    year              = decimal_digit decimal_digit decimal_digit decimal_digit .
    month             = decimal_digit decimal_digit .
    day               = decimal_digit decimal_digit .
    time              = hour ":" minute ":" second [ fractional_second ] time_offset .
    hour              = decimal_digit decimal_digit .
    minute            = decimal_digit decimal_digit .
    second            = decimal_digit decimal_digit .
    fractional_second = "."  { decimal_digit } .
    time_offset       = "Z" | ("+" | "-" ) hour ":" minute .

Examples:

    1952-01-25T12:35:51Z
    2018-08-15T13:36:23-07:00
    2018-01-01                // midnight on January 1st 2018 UTC

#### String literals

A string literal represents a sequence of characters enclosed in double quotes.
Within the quotes any character may appear except an unescaped double quote.
String literals support several escape sequences.

    \n   U+000A line feed or newline
    \r   U+000D carriage return
    \t   U+0009 horizontal tab
    \"   U+0022 double quote
    \\   U+005C backslash
    \${  U+0024 U+007B dollar sign and opening curly bracket

Additionally any byte value may be specified via a hex encoding using `\x` as the prefix.
The hex encoding of values must result in a valid UTF-8 sequence.


    string_lit       = `"` { unicode_value | byte_value | StringExpression | newline } `"` .
    byte_value       = `\` "x" hex_digit hex_digit .
    hex_digit        = "0" … "9" | "A" … "F" | "a" … "f" .
    unicode_value    = unicode_char | escaped_char .
    escaped_char     = `\` ( "n" | "r" | "t" | `\` | `"` ) .
    StringExpression = "${" Expression "}" .


Examples:

    "abc"
    "string with double \" quote"
    "string with backslash \\"
    "日本語"
    "\xe6\x97\xa5\xe6\x9c\xac\xe8\xaa\x9e" // the explicit UTF-8 encoding of the previous line

String literals are also interpolated for embedded expressions to be evaluated as strings.
Embedded expressions are enclosed within the literals `${` and `}` respectively.
The expressions are evaluated in the scope containing the string literal.
The result of an expression is formatted as a string and replaces the string content between the brackets.
All types are formatted as strings according to their literal representation.
To include the literal `${` within a string it must be escaped.

Interpolation example:

    n = 42
    "the answer is ${n}" // the answer is 42
    "the answer is not ${n+1}" // the answer is not 43
    "dollar sign opening curly bracket \${" // dollar sign opening curly bracket ${

String interpolation expressions must satisfy the Stringable constraint.

Arbitrary Expression Interpolation example:

    n = duration(v: "1m")
    "the answer is ${n}" // the answer is 1m
    t0 = time(v: "2016-06-13T17:43:50.1004002Z")
    "the answer is ${t0}" // the answer is 2016-06-13T17:43:50.1004002Z

#### Regular expression literals

A regular expression literal represents a regular expression pattern, enclosed in forward slashes.
Within the forward slashes, any unicode character may appear except for an unescaped forward slash.
The `\x` hex byte value representation from string literals may also be present.
The hex encoding of values must result in a valid UTF-8 sequence.

In addition to standard escape sequences, regular expression literals also support the following escape sequences:

    \/   U+002f forward slash

    regexp_lit         = "/" regexp_char { regexp_char } "/" .
    regexp_char        = unicode_char | byte_value | regexp_escape_char .
    regexp_escape_char = `\/`


Examples:

    /.*/
    /http:\/\/localhost:8086/
    /^\xe6\x97\xa5\xe6\x9c\xac\xe8\xaa\x9e(ZZ)?$/
    /^日本語(ZZ)?$/ // the above two lines are equivalent
    /a\/b\s\w/ // escape sequences and character class shortcuts are supported
    /(?:)/ // the empty regular expression

The regular expression syntax is defined by [RE2](https://github.com/google/re2/wiki/Syntax).

#### Label literals

```flux

.mylabel
._value
."with spaces"
```

A label literal represents a "label" used to refer to specific record fields. They have two variants, where the `.` can
be followed by either an identifier or a string literal (allowing labels with characters that are not allowed in identifiers to be specified).

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

Options are not closed, meaning new options may be defined and consumed within packages and scripts.
Changing the value of an option for a package changes the value for all references to that option from any other package.

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

    import "timezone"

    option location = timezone.fixed(offset:-5h) // set timezone to be 5 hours west of UTC
    option location = timezone.location(name:"America/Denver") // set location to be America/Denver

### Types

A type defines a set of values and operations on those values.
Types are never explicitly declared as part of the syntax except as part of a [builtin statement](#built-ins).
Types are always inferred from the usage of the value.
Type inference follows a Hindley-Milner style inference system.

#### Union types

A union type defines a set of types.
In the rest of this section a union type will be specified as follows:

    T = t1 | t2 | ... | tn

where `t1`, `t2`, ..., and `tn` are types.
In the example above a value of type `T` is either of type `t1`, type `t2`, ..., or type `tn`.

#### Basic types

These are the types from which all other Flux data types are constructed.

##### Null type

The _null type_ represents a missing or unknown value.
The _null type_ name is `null`.
There is only one value that comprises the _null type_ and that is the _null_ value.

A type `t` is nullable if it can be expressed as follows:

    t = {s} | null

where `{s}` defines a set of values.

##### Boolean types

A _boolean type_ represents a truth value, corresponding to the preassigned variables `true` and `false`.
The boolean type name is `bool`.
The boolean type is nullable and can be formally specified as follows:

    bool = {true, false} | null

##### Numeric types

A _numeric type_ represents sets of integer or floating-point values.

The following numeric types exist:

    uint    = {the set of all unsigned 64-bit integers} | null
    int     = {the set of all signed 64-bit integers} | null
    float   = {the set of all IEEE-754 64-bit floating-point numbers} | null

Note all numeric types are nullable.

##### Time types

A _time type_ represents a single point in time with nanosecond precision.
The time type name is `time`.
The time type is nullable.

##### Duration types

A _duration type_ represents a length of time with nanosecond precision.
The duration type name is `duration`.
The duration type is nullable.

#### Binary types

A _bytes type_ represents a sequence of byte values.
The bytes type name is `bytes`.

##### String types

A _string type_ represents a possibly empty sequence of characters.
Strings are immutable: once created they cannot be modified.
The string type name is `string`.
The string type is nullable.
Note that an empty string is distinct from a _null_ value.

##### Label types (Upcoming/Feature flagged)

A _label type_ represents the name of a record field.
String literals may be treated as a label type instead of a `string` when used in a context that
expects a label type.

```
"a" // Can be treated as Label("a")
"xyz" // Can be treated as  Label("xyz")
```

In effect, this allows functions accepting a record and a label to refer to specific properties of
the record.

```
// "mycolumn" is treated as Label("mycolumn") for when passed to `mean`
mean(column: "mycolumn") // Calculates the mean of `mycolumn`
```

##### Regular expression types

A _regular expression type_ represents the set of all patterns for regular expressions.
The regular expression type name is `regexp`.
The regular expression type is **not** nullable.

#### Composite types

These are types constructed from basic types.
Composite types are not nullable.

##### Array types

An _array type_ represents a sequence of values of any other type.
All values in the array must be of the same type.
The length of an array is the number of elements in the array.

##### Record types

An _record type_ represents a set of unordered key and value pairs.
The key can be a string or a [type variable](<#Type variables>).
The value may be any other type, and need not be the same as other values within the record.

Type inference will determine the properties that are present on a record.
If type inference determines all the properties on a record it is said to be bounded.
Not all keys may be known on the type of a record in which case the record is said to be unbounded.
An unbounded record may contain any property in addition to the properties it is known to contain.

##### Dictionary types

A _dictionary type_ is a collection that associates keys to values.
Keys must be comparable and of the same type.
Values must also be of the same type.

##### Function types

A _function type_ represents a set of all functions with the same argument and result types.
Functions arguments are always named. (There are no positional arguments.)
Therefore implementing a function type requires that the arguments be named the same.

##### Stream types

A _stream type_ represents an unbounded collection of values.
The values must be records and those records may only hold int, uint, float, string, time or bool types.

#### Polymorphism

Flux functions can be polymorphic, meaning they can be applied to arguments of different types.
Flux supports parametric, record, and ad hoc polymorphism.

##### Type variables

Polymorphic types are represented via "type variables" which are specified with a single uppercase
letter (`A`, `B`, etc).

##### Parametric Polymorphism

Parametric polymorphism is the notion that a function can be applied uniformly to arguments of any type.
The identity function is one such example.

    f = (x) => x
    f(x: 1)
    f(x: 1.1)
    f(x: "1")
    f(x: true)
    f(x: f)

##### Record Polymorphism

Record polymorphism is the notion that a function can be applied to different types of records.

    john = {name:"John", lastName:"Smith"}
    jane = {name:"Jane", age:44}

    // John and Jane are records with different types.
    // We can still define a function that can operate on both records safely.

    // name returns the name of a person
    name = (person) => person.name

    name(person:john) // John
    name(person:jane) // Jane

    device = {id: 125325, lat: 15.6163, lon: 62.6623}

    name(person:device) // Type error, "device" does not have a property name.

Records of differing types can be passed to the same function so long as they contain the necessary properties.
The necessary properties are determined by the use of the record.

##### Ad hoc Polymorphism

Ad hoc polymorphism is the notion that a function can be applied to arguments of different types, with different behavior depending on the type.

    add = (a, b) => a + b

    // Integer addition
    add(a: 1, b: 1)

    // String concatenation
    add(a: "str", b: "ing")

    // Addition not defined for boolean data types
    add(a: true, b: false)

#### Type constraints

Type constraints are a type system concept used to implement static ad hoc polymorphism.
For example, `add = (a, b) => a + b` is a function that is defined only for `Addable` types.
If one were to pass a record to `add` like so:

    add(a: {}, b: {})

the result would be a compile-time type error because records are not addable.
Like types, constraints are never explicitly declared but rather inferred from the context.

##### Addable Constraint

Addable types are those the binary arithmetic operator `+` accepts.
Int, Uint, Float, and String types are Addable.

##### Subtractable Constraint

Subtractable types are those the binary arithmetic operator `-` accepts.
Int, Uint, and Float types are Subtractable.

##### Divisible Constraint

Divisible types are those the binary arithmetic operator `\` accepts.
Int, Uint, and Float types are Divisible.

##### Numeric Constraint

Int, Uint, and Float types are Numeric.

##### Comparable Constraint

Comparable types are those the binary comparison operators `<`, `<=`, `>`, or `>=` accept.
Int, Uint, Float, String, Duration, and Time types are Comparable.

##### Equatable Constraint

Equatable types are those that can be compared for equality using the `==` or `!=` operators.
Bool, Int, Uint, Float, String, Duration, Time, Bytes, Array, and Record types are Equatable.

##### Nullable Constraint

Nullable types are those that can be null.
Bool, Int, Uint, Float, String, Duration, and Time types are Nullable.

##### Record Constraint

Records are the only types that fall under this constraint.

##### Negatable Constraint

Negatable types are those the unary arithmetic operator `-` accepts.
Int, Uint, Float, and Duration types are Negatable.

##### Timeable Constraint

Duration and Time types are Timeable.

##### Stringable Constraint

Stringable types can be evaluated and expressed in string interpolation.
String, Int, Uint, Float, Bool, Time, and Duration types are Stringable.

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
            | regexp_lit
            | duration_lit
            | pipe_receive_lit
            | RecordLiteral
            | ArrayLiteral
            | DictLiteral
            | FunctionLiteral .

##### Record literals

Record literals construct a value with the record type.

    RecordLiteral  = "{" RecordBody "}" .
    RecordBody     = WithProperties | PropertyList .
    WithProperties = identifier "with"  PropertyList .
    PropertyList   = [ Property { "," Property } ] .
    Property       = identifier [ ":" Expression ]
                   | string_lit ":" Expression .

Examples:

    {a: 1, b: 2, c: 3}
    {a, b, c}
    {o with x: 5, y: 5}
    {o with a, b}

##### Array literals

Array literals construct a value with the array type.

    ArrayLiteral   = "[" ExpressionList "]" .
    ExpressionList = [ Expression { "," Expression } ] .

##### Dictionary literals

Dictionary literals construct a value with the dict type.

    DictLiteral     = EmptyDict | "[" AssociativeList "]" .
    EmptyDict       = "[" ":" "]" .
    AssociativeList = Association { "," AssociativeList } .
    Association     = Expression ":" Expression .

The keys can be arbitrary expressions.
The type system will enforce that all keys are of the same type.

Examples:

    a = "a"
    b = [:] // empty dictionary
    c = [a: 1, "b": 2] // dictionary mapping string values to integers
    d = [a: 1, 2: 3] // type error: cannot mix string and integer keys

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


Function arguments are named. There are no positional arguments.
Values implementing a function type must use the same argument names.

    apply = (f, x) => f(x: x)

    apply(f: (x) => x + 1, x: 2) // 3
    apply(f: (a) => a + 1, x: 2) // error, function must use the same argument name `x`.
    apply(f: (x, a=3) => a + x, x: 2) // 5, extra default arguments are allowed


#### Call expressions

A call expressions invokes a function with the provided arguments.
Arguments must be specified using the argument name, positional arguments are not supported.
Argument order does not matter.
When an argument has a default value, it is not required to be specified.

    CallExpression = "(" PropertyList ")" .

Call expressions support a short notation in case the name of the argument matches the parameter name.
This notation can be used only when every argument matches its parameter.

Examples:

```
add = (a,b) => a + b

a = 1
b = 2

add(a, b)
// is the same as
add(a: a, b: b)
// both FAIL: cannot mix short and long notation.
add(a: a, b)
add(a, b: b)
```

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

Member expressions access a property of a record.
They are specified via an expression of the form `rec.k` or `rec["k"]`.
The property being accessed must be either an identifier or a string literal.
In either case the literal value is the name of the property being accessed, the identifier is not evaluated.
It is not possible to access an record's property using an arbitrary expression.

If `rec` contains an entry with property `k`, both `rec.k` and `rec["k"]` return the value associated with `k`.
If `rec` is bounded and does *not* contain a property `k`, both `rec.k` and `rec["k"]` report a type checking error.
If `rec` is unbounded and does *not* contain a property `k`, both `rec.k` and `rec["k"]` return _null_.

    MemberExpression        = DotExpression  | MemberBracketExpression
    DotExpression           = "." identifier
    MemberBracketExpression = "[" string_lit "]" .

#### Conditional Expressions

Conditional expressions evaluate a boolean-valued condition and if the result is _true_,
the expression following the `then` keyword is evaluated and returned.
Otherwise the expression following the `else` keyword is evaluated and returned.
In either case, the branch not taken is not evaluated;
only side effects associated with the branch that is taken will occur.

    ConditionalExpression   = "if" Expression "then" Expression "else" Expression .

Example:

    color = if code == 0 then "green" else if code == 1 then "yellow" else "red"

Note according to the above definition, if a condition evaluates to a _null_ or unknown value, the _else_ branch is evaluated.

#### Operators

Operators combine operands into expressions.
The precedence of the operators is given in the table below. Operators with a lower number have higher precedence.

| Precedence | Operator       | Description               |
| ---------- | -------------- | ------------------------- |
| 1          | `a()`          | Function call             |
|            | `a[]`          | Member or index access    |
|            | `.`            | Member access             |
| 2          | `\|>`          | Pipe forward              |
| 3          | `() => 1`      | FunctionLiteral           |
| 4          | `^`            | Exponentiation            |
| 5          | `*` `/` `%`    | Multiplication, division, |
|            |                | and modulo                |
| 6          | `+` `-`        | Addition and subtraction  |
| 7          | `==` `!=`      | Comparison operators      |
|            | `<` `<=`       |                           |
|            | `>` `>=`       |                           |
|            | `=~` `!~`      |                           |
| 8          | `not`          | Unary logical operator    |
|            | `exists`       | Null check operator       |
| 9          | `and`          | Logical AND               |
| 10         | `or`           | Logical OR                |
| 11         | `if/then/else` | Conditional               |

The operator precedence is encoded directly into the grammar as the following.

    Expression               = ConditionalExpression .
    ConditionalExpression    = LogicalExpression
                             | "if" Expression "then" Expression "else" Expression .
    LogicalExpression        = UnaryLogicalExpression
                             | LogicalExpression LogicalOperator UnaryLogicalExpression .
    LogicalOperator          = "and" | "or" .
    UnaryLogicalExpression   = ComparisonExpression
                             | UnaryLogicalOperator UnaryLogicalExpression .
    UnaryLogicalOperator     = "not" | "exists" .
    ComparisonExpression     = AdditiveExpression
                             | ComparisonExpression ComparisonOperator AdditiveExpression .
    ComparisonOperator       = "==" | "!=" | "<" | "<=" | ">" | ">=" | "=~" | "!~" .
    AdditiveExpression       = MultiplicativeExpression
                             | AdditiveExpression AdditiveOperator MultiplicativeExpression .
    AdditiveOperator         = "+" | "-" .
    MultiplicativeExpression = ExponentExpression
                             | ExponentExpression MultiplicativeOperator MultiplicativeExpression .
    MultiplicativeOperator   = "*" | "/" | "%" .
    ExponentExpression       = PipeExpression
                             | ExponentExpression ExponentOperator PipeExpression .
    ExponentOperator         = "^" .
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

Dividing by zero or using the mod operator with a divisor of zero for integer and unsigned integer types will result in an error.
Floating point divide by zero produces positive or negative infinity according to the [IEEE-754](https://en.wikipedia.org/wiki/IEEE_754) floating point specification.

### Packages

Flux source is organized into packages.
A package consists of one or more source files.
Each source file is parsed individually and composed into a single package.

    File = [ PackageClause ] [ ImportList ] StatementList .
    ImportList = { ImportDeclaration } .

#### Package clause

    PackageClause = [ Attributes ] "package" identifier .

A package clause defines the name for the current package.
Package names must be valid Flux identifiers.
The package clause must be at the begining of any Flux source file.
All files in the same package must declare the same package name.
When a file does not declare a package clause, all identifiers in that file will belong to the special _main_ package.



##### package main

The _main_ package is special for a few reasons:

1. It defines the entrypoint of a Flux program
2. It cannot be imported
3. All statements are marked as producing side effects

#### Import declaration

    ImportDeclaration = [ Attributes ] "import" [identifier] string_lit


Associated with every package is a package name and an import path.
The import statement takes a package's import path and brings all of the identifiers defined in that package into the current scope under a namespace.
The import statement defines the namespace through which to access the imported identifiers.
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

### Attributes

Attributes define a set of properties on source code elements.

    Attributes             = { Attribute } .
    Attribute              = "@" identifier AttributeParameters .
    AttributeParameters    = "(" [ AttributeParameterList [ "," ] ] ")" .
    AttributeParameterList = AttributeParameter { "," AttributeParameter } .
    AttributeParameter     = PrimaryExpression

The full set of defined attributes and their meaning is left to the runtime to specify.

Example

    @edition("2022.1")
    package main


### Statements

A statement controls execution.

    Statement      = [ Attributes ] StatementInner .
    StatementInner = OptionAssignment
                   | BuiltinStatement
                   | VariableAssignment
                   | ReturnStatement
                   | ExpressionStatement
                   | TestcaseStatement .


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

#### Testcase statements

>NOTE: Testcase statements only work within the context of a Flux developement environment. We expect to expand their use in the future.

A statement that defines a test case.

    TestcaseStatement = "testcase" identifier [ TestcaseExtention ] Block .
    TestcaseExtention = "extends" string_lit

Test cases are defined as a set of statements with special scoping rules.
Each test case statement in a file is considered to be its own main package.
In effect all statements in package scope and all statements contained within the test case statment are flattened into a single main package and executed.
Use the `testing` package from the standard library to control the pass failure of the test case.

Test extention augments an existing test case with more statements or attributes.
A special function call `super()` must be made inside the body of a test case extention, in its place all statements from the parent test case will be executed.


Examples:

A basic test case for addition

    import "testing"

    testcase addition {
        testing.assertEqualValues(got: 1 + 1, want: 2)
    }

An example of test case extention to validate a feature does not regress existing behavior.

    @feature({vectorization: true})
    testcase vector_addition extends "basics_test.addition" {
        super()
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

    BuiltinStatement = "builtin" identifier ":" TypeExpression .
    TypeExpression   = MonoType ["where" Constraints] .

    MonoType     = Tvar | BasicType | ArrayType | StreamType | VectorType | RecordType | FunctionType .
    Tvar         = "A" … "Z" .
    BasicType    = "int" | "uint" | "float" | "string" | "bool" | "time" | "duration" | "bytes" | "regexp" .
    ArrayType    = "[" MonoType "]" .
    StreamType    = "stream" "[" MonoType "]" .
    VectorType    = "vector" "[" MonoType "]" .
    RecordType   = ( "{" [RecordTypeProperties] "}" ) | ( "{" Tvar "with" RecordTypeProperties "}" ) .
    FunctionType = "(" [FunctionTypeParameters] ")" "=>" MonoType .

    RecordTypeProperties = RecordTypeProperty { "," RecordTypeProperty } .
    RecordTypeProperty   = Label ":" MonoType .
    Label = identifier | string_lit

    FunctionTypeParameters = FunctionTypeParameter { "," FunctionTypeParameter } .
    FunctionTypeParameter = [ "<-" | "?" ] identifier ":" MonoType .

    Constraints = Constraint { "," Constraint } .
    Constraint  = Tvar ":" Kinds .
    Kinds       = identifier { "+" identifier } .

Example

    builtin filter : (<-tables: stream[T], fn: (r: T) => bool) => stream[T]

### Type expressions

Type expressions as defined above describe the type of values.
Every value have has a known type, its type can be described using type expressions.
The available types are defined in the [Types](#Types) section.

Here are a set of examples showing various values and their corresponding type using type expressions.

    // A floating point value
    1.5 // float

    // A string value
    "this is a string"  // string

    // A boolean value
    false // bool

    // A record with three properties all of which are integers
    {x: 1, y: 2, z: 4} // {x: int, y: int, z: int}

    // A record with three properties with different types
    {x: 1, y: true, z: 5.6} // {x: int, y: bool, z: float}

    // A function that takes an integer and returns an integer
    (x) => x + 1 // (x: int) => int

    // A function that takes two values of the same type where that type must be Addable.
    (a, b) => a + b // (a:A, b: A) => A where A: Addable

    // A function that takes two values of any type and returns a record containing those values
    (n, m) => {x: n, y: m} // (n: A, m: B) => {x: A, y: B}

    // A function that takes a record and returns a new records with additional properties
    (r) => ({r with z: 0}) // (r: A) => {A with z: int} where A: Record

    // A function that takes a record with a name property and returns it
    (r) => r.name // (r: {A with name: B}) => B where A: Record

    // A function that takes a record that has at least the status property which is an int and the function returns a boolean value.
    (r) => r.status == 400 // (r: {A with status: int}) => bool where A: Record


## Data model

Flux employs a basic data model built from basic data types.
The data model consists of tables, records, columns and streams.

### Record

A record is a tuple of named values and is represented using an record type.

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


### Stream of tables

A stream represents a potentially unbounded set of tables.
A stream is grouped into individual tables using the group key.
Within a stream each table's group key value is unique.

A stream is represented using the stream type `stream[A] where A: Record`.
The group key is not explicity modeled in the Flux type system.


### Missing values (null)

`null` is a predeclared identifier representing a missing or unknown value.
`null` is the only value comprising the _null type_.

Any non-boolean operator that operates on basic types, returns _null_ when at least one of its operands is _null_.
This can be explained intuitively with the following table and by thinking of a null value as an unknown value.

| Expression       | Evaluates To | Because                                                                             |
| ---------------- | ------------ | ----------------------------------------------------------------------------------- |
| _null_ + 5       | _null_       | Adding 5 to an unknown value is still unknown                                       |
| _null_ * 5       | _null_       | Multiplying an unknown value by 5 is still unknown                                  |
| _null_ == 5      | _null_       | We don't know if an unknown value is equal to 5                                     |
| _null_ < 5       | _null_       | We don't know if an unknown value is less than 5                                    |
| _null_ == _null_ | _null_       | We don't know if something unknown is equal to something else that is also unknown  |

In other words, operating on something unknown produces something that is still unknown.
The only place where this is not the case is in boolean logic.

Because boolean types are nullable, Flux implements ternary logic as a way of handling boolean operators with _null_ operands.
Again, by interpreting a _null_ operand as an unknown value, we have the following definitions:

* not _null_ = _null_
* _null_ or false = _null_
* _null_ or true = true
* _null_ or _null_ = _null_
* _null_ and false = false
* _null_ and true = _null_
* _null_ and _null_ = _null_

Note according to the definitions above, it is not possible to check whether or not an expression is _null_ using the `==` and `!=` operators as these operators will return _null_ if any of their operands are _null_.
In order to perform such a check, Flux provides a built-in `exists` operator defined as follows:

* `exists x` returns false if `x` is _null_
* `exists x` returns true if `x` is not _null_

The exists operator can also be applied to records as follows:

* `exists rec.x` returns false if `x` is not a property of `rec`
* `exists rec.x` returns true if `x` is a property of `rec`

### Transformations

Transformations define a change to a stream.
Transformations may consume an input stream and always produce a new output stream.
The order of group keys for the output stream will have a stable output order based on the input stream.
The specific ordering may change between releases and it would not be considered a breaking change.

Most transformations output one table for every table they receive from the input stream.
Transformations that modify the group keys or values will need to regroup the tables in the output stream.
A transformation produces side effects when it is constructed from a function that produces side effects.

Transformations are represented using function types.

Some transformations, for instance `map` and `filter`, are represented using higher-order functions (functions that accepts other functions).
When specifying the function passed in, _make sure that you use the same names for its parameters_.


`filter`, for instance, accepts argument `fn` which is of type `(r: A) => bool where A: Record`.
An invocation of `filter` must take a function with one argument named `r`:

```
from(bucket: "db")
    |> filter(fn: (r) => ...)
```

This script would fail:

```
from(bucket: "db")
    |> filter(fn: (v) => ...)

// FAILS!: 'v' != 'r'.
```

The reason is simple: Flux does not support positional arguments, so parameter names matter.
The transformation (in our example, `filter`) must know the name of the parameter in the given function in order to invoke it properly.
The process happens the other way around, actually:
our `filter` implementation supposes to invoke a function in this way:

```
fn(r: <the-record>)
```

So, you have to:

```
...
    |> filter(fn: (r) => ...)
...
```


### Standard Library

Flux provides a standard library of functions. Find documentation here https://docs.influxdata.com/flux/latest/stdlib/

#### Experimental namespace

Within the standard library there is an `experimental` package namespace.
Packages within this namesapce are subject to breaking changes without notice.
See the package documentation for more details https://docs.influxdata.com/flux/latest/stdlib/experimental/

### Composite data types

A composite data type is a collection of primitive data types that together have a higher meaning.

### Execution model

A query specification defines what data and operations to perform.
The execution model reserves the right to perform those operations as efficiently as possible.
The execution model may rewrite the query in anyway it sees fit while maintaining correctness.

## Modules

>NOTE: Modules are not fully implemented yet, follow https://github.com/influxdata/flux/issues/4296 for details.

A module is a collection of packages that can be imported.
A module has a module path, version and a collection of packages with their source code.

### Module path

The module path is the import path of the top level package within the module.
Additionally major versions of a module of two or greater must add a final element to the path of the form `v#` where `#` is the major version number.
For modules at version zero or one the path must not contain the major version as it is not necessary.
A change from `v0` to `v1` may include a breaking change but once `v1` is published any future breaking changes will be a new major version.

Example

    foo/bar    // module path of foo/bar for version zero or one
    foo/bar/v2 // module path of foo/bar for major version two

### Module versions

All modules are versioned using a [semantic version number](https://semver.org/) prefixed with a `v`, i.e. `vMAJOR.MINOR.PATCH`.
Once a module version has been published it cannot be modified.
A new version of the module must be published.

### Module registry

A modules is hosted and stored on a registry.
A module path is unique to the registry that hosts the module.
Module paths need not be unique across different registries.

#### Registry attribute

The `registry` attribute defines the available registries and must precede a package clause.
The registry attribute expects two arguments, the first is the name of the registry and the second is the `$base` URL of the registry API endpoint.
See the [Registry API](#registry-api) for more details.

The runtime may define default registries, a registry attribute will override any default registry.
The standard library will never contain a top level package named `modules` or any name containing a `.`.
This makes it possible for the runtime to use `modules` or any name containing a `.` (i.e. a DNS name) as a default registry name.

Example:

    @registry("modules", "http://localhost/modules")
    @registry("example.com", "https://example.com/api/modules")
    package main

### Importing modules

Flux modules are imported using an import declaration.
The import path may contain specifiers about which registry and which versions of a module should be imported.

An import path follows this grammar:

    ImportPath       = ModulePath [ "/" PackagePath ] [ Version ] .
    ModulePath       = RegistryName [ "/" ModuleName ] [ MajorVersion ] .
    RegistryName     = PathElement
    ModuleName       = PathElement
    MajorVersion     = "/v" int_lit .
    PackagePath      = PathElement { "/" PathElement } .
    Version          = PreVersion | MinVersion
    PreVersion       = "@pre"
    MinVersion       = "@v" int_lit "." int_lit "." int_lit "." .
    PathElement      = ascii_letter { ascii_letter } .
    ascii_letter     = /* alpha numeric and underscore ASCII characters */

Per the grammar a module path may have up to three path elements:

* Registry name
* Module name
* Module major version

A package path may have an arbitrary depth and must not begin with a major version.

When resolving an import path the first path element of the module path is compared against the defined registry names.
If a match is found the import path is understood to be relative to the registry.
If no match is found the import path is understood to be an import from the standard library.
It is an error to specify a version on imports from the standard library.
The standard library version is implicit to the runtime.

When no version information is provided the latest version of the module is used.
A _minimum_ version may be specified.
An import may also specify the version `pre` which is the most recent pre-release version of the module.


Examples

The following examples use default registry of `modules`.

    import "foo"                         // imports package `foo` from the standard library
    import "modules/foo"                 // imports the latest version 0.x or 1.x version of the `foo` module from the `modules` registry
    import "modules/foo/a/b/c"           // imports package a/b/c from the latest 0.x or 1.x version
    import "modules/foo@v1.5.6"          // imports at least version 1.5.6
    import "modules/foo/v2"              // imports the latest 2.x version
    import "modules/foo/v2/a"            // imports package `a` from the latest 2.x version
    import "modules/foo/v2/a@v2.3.0"     // imports package `a` from at least version 2.3.0 of the `foo` module
    import "modules/foo/v2/a/b/c"        // imports package `a/b/c` from the latest 2.x version
    import "modules/foo/v2/a/b/c@v2.3.0" // imports package `a/b/c` from at least version 2.3.0
    import "modules/foo/a/b/c@pre"       // imports package `a/b/c` from the latest pre-release 0.x or 1.x version
    import "modules/foo@pre"             // imports the latest pre-release 0.x or 1.x version
    import "modules/foo/v2@pre"          // imports the latest pre-release 2.x version

### Version resolution

When multiple modules both depend on a specific version of another module the maximum version of the minimum versions is used.
Major versions of a module are considered different modules (they have different module paths), therefore multiple major versions of a module may be imported into the same Flux script.

When multiple import declarations exist for the same module at most one import declaration must specify version information.

Example

```
// a.flux
package a

import "foo@v1.1.0"
```

```
// b.flux
package b

import "foo@v1.2.0"
```

```
// main.flux
@registry("modules", "http://localhost/modules")
package main

import "modules/a"
import "modules/b"
```

Package `main` depends on module `foo` via both of the modules `a` and `b`.
However `a` and `b` specify different versions of `foo`.
The possible versions of `foo` include `1.1.0` and `1.2.0`.
Flux will pick the maximum version of these possible versions, so version `1.2.0` of `foo` is used.
This is sound because module `a` has specified that it needs at least version `1.1.0` of `foo` and that constraint is satisfied.

### Registry API

Modules can be published and downloaded over an HTTP API from a registry.
Modules are immutable, once a version is published it cannot be modified, a new version must be published instead.

The HTTP API will have the routes listed in the following table where `$base` is the anchor point of the API, `$module` is a module path without the registry name, and `$version` is a semantic version of the form `vMAJOR.MINOR.PATCH`.

| Method | Path                          | Description                                                                                                                                                |
| ------ | ----                          | -----------                                                                                                                                                |
| GET    | $base/$module/@v/list         | Returns a list of known versions of the given module in plain text, one per line.                                                                          |
| GET    | $base/$module/@v/$version.zip | Returns a zip file of the contents of the module at a specific version.                                                                                    |
| GET    | $base/$module/@latest         | Returns the highest released version, or if no released versions exist the highest pre-release version of the given module in plain text on a single line. |
| GET    | $base/$module/@pre            | Returns the highest pre-released version of the given module in plain text on a single line.                                                               |
| POST   | $base/$module/@v/$version     | Publish a new version of the module where the POST body contains multipart formdata for the contents of the module.                                        |

The POST endpoint expects the module's contents to be encoded using multipart form data as defined in [RFC 2046](https://rfc-editor.org/rfc/rfc2046.html).
Each file within the module must be uploaded using `module` as the file key and the relative path to the module root as the filename.
The filename must end in `.flux` and must follow the rules of import paths for allowed characters.
A maximum POST body size of 10MB will be read, any larger body will result in an error.

As an example, to download the zip file for a module `foo` at version `v0.5.6`, for an API endpoint anchored at `https://example.com/flux/modules/` use this URL `https://example.com/flux/modules/foo/@v/v1.5.6.zip`.
Or for the module `foo/v2` at version `v2.3.4` the URL is `https://example.com/flux/modules/foo/v2/@v/v2.3.4.zip`.

Examples

The following examples use a $base of `/flux/modules/`

    GET /flux/modules/foo/@v/list          # Return a list of versions for the module foo
    GET /flux/modules/bar/@v/v1.3.4.zip    # Return a zip file of the bar module at version 1.3.4
    GET /flux/modules/bar/v2/@v/v2.3.4.zip # Return a zip file of the bar module at version 2.3.4
    GET /flux/modules/bar/@latest          # Return the latest 0.x or 1.x version of bar
    GET /flux/modules/bar/@pre             # Return the latest  0.x or 1.x pre-release version of bar
    GET /flux/modules/bar/v2/@latest       # Return the latest 2.x release version of bar
    GET /flux/modules/bar/v2/@pre          # Return the latest 2.x pre-release version of bar


## Versions and editions

Flux follows a [semantic versioning scheme](https://semver.org/) such that breaking changes are clearly communicated as part of the version information.
*Editions* allow for the introduction of breaking changes via an opt-in mechanism.

Flux editions are a set of features that are enabled in Flux.
If the edition is not enabled then the features are not enabled.
These features would otherwise be breaking changes to Flux.
An edition is explicitly opt-in.
The pattern of editions allows users to migrate to new Flux versions without risk that their scripts will break.

An edition is separate from a version.
A version of Flux represents a single point in the commit history of the Flux source code.
A user can only use a single version of Flux for a given script.
Editions represent a set of features that are enabled for a given Flux version.
So long as the Flux version supports all the features of the edition it can be enabled.
A user can upgrade to a newer Flux version without being required to upgrade to the newest edition of Flux.

New Flux features that require a breaking change to the syntax or semantics of Flux must always be part of a new edition.
This means that a script can regularly update to the newest Flux version without risk of breaking because any breaking changes are explicitly opt-in.
With editions the Flux community gets both a pattern where users can always be running the latest version of Flux and the ability to introduce new useful but otherwise breaking changes to Flux.

We anticipate that there will be at most one new editions of Flux a year. A slow cadence of new editions means users have ample time to migrate to a new edition if desired.
Even being a few years behind on editions should only mean a few migration steps in order to have access to features introduced in the newest edition.

### First Edition

The first edition of Flux is 2022.1.
This first edition is the only minimum required edition of Flux.

### Future Editions

Editions will be named after the year in which they are created, with an added sequence number if more than one edition needs to be created in a single year.
For example the first edition is `2022.1` because it is created in the year `2022` and is the first edition created in that year.

### Editions are optional

A Flux script may specify directly the edition it requires to function.
Additionally the Flux runtime will allow for the edition to be specified out of band of the script, thus allowing for deployments of Flux to have control over the edition.

#### Edition attribute

The edition of the current script is specified as an `edition` attribute on the package with a single parameter, the name of the edition.
It is an error for multiple files within a package or module to specify differing editions.

Examples:

Specify the edition with an explicit package clause

    @edition("2022.1")
    package math

    add = (x,y) => x + y


### Flux Editions and Modules

Each Flux module may specify its own edition, therefore a Flux script on an earlier edition may import and consume a module that uses a newer edition of Flux.
Naturally if a module exposes a new edition feature via its API, consumers of that module will be required to use at least that edition in order to directly consume the module.


### Migrating to a new edition

When a new edition is created a migration process will be provided to ease the migration from an older edition to a new edition.

### Editions and Experimental

Editions do not change the contract of the experimental package namespace.
Experimental packages are still subject to breaking changes without notice.
Most new features do not require a breaking change to Flux syntax or semantics.
As such it will remain common for new packages to be introduced as experimental packages.
When an API is stabilized, it can be promoted out of experimental without the need to create a new edition.

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

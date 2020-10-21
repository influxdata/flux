use super::*;
use std::str;

// This gives us a colorful diff.
#[cfg(test)]
use pretty_assertions::assert_eq;

fn assert_unchanged(script: &str) {
    let output = format(script.to_string()).unwrap();
    assert_eq!(script, output);
}

fn assert_format(script: &str, expected: &str) {
    let output = format(script.to_string()).unwrap();
    assert_eq!(expected, output);
}

#[test]
fn binary_op() {
    assert_unchanged("1 + 1 - 2");
    assert_format("1 +  1 - 2", "1 + 1 - 2");
    assert_unchanged("1 * 1 / 2");
    assert_unchanged("2 ^ 4");
    assert_unchanged("1 * (1 / 2)");
}

#[test]
fn funcs() {
    assert_format(
        r#"(r) =>
    (r.user ==     "user1")"#,
        "(r) => r.user == \"user1\"",
    );
    assert_unchanged(r#"(r) => r.user == "user1""#);
    assert_unchanged(r#"add = (a, b) => a + b"#); // decl
    assert_unchanged("add(a: 1, b: 2)"); // call
    assert_unchanged(r#"foo = (arg=[]) => 1"#); // nil value as default
    assert_unchanged(r#"foo = (arg=[1, 2]) => 1"#); // none nil value as default

    // record expressions
    assert_unchanged(r#"(r) => ({r with _value: r._value + 1})"#);
    assert_unchanged(
        r#"(r) => ({r with
    _value: r._value + 1,
})"#,
    );

    //
    // pipe expressions
    //

    // multiline based on pipe depth
    assert_unchanged(
        r#"(tables) => tables
    |> a()
    |> b()"#,
    );
    // single line
    assert_unchanged(r#"(tables) => tables |> a()"#);
    // multiline based on initial conditions
    assert_unchanged(
        r#"(tables) => tables
    |> a()"#,
    );
}

#[test]
fn object() {
    assert_unchanged("{a: 1, b: {c: 11, d: 12}}");
    assert_unchanged("{foo with a: 1, b: {c: 11, d: 12}}"); // with
    assert_unchanged("{a, b, c}"); // implicit key object literal
    assert_unchanged(r#"{"a": 1, "b": 2}"#); // object with string literal keys
    assert_unchanged(r#"{"a": 1, b: 2}"#); // object with mixed keys
    assert_unchanged("{\n    a: 1,\n    b: 2,\n    c: 3,\n    d: 4,\n}"); // multiline object based on property count
    assert_unchanged("{\n    a: 1,\n    b: 2,\n}"); // multiline object based on initial conditions
    assert_unchanged("{x with\n    a: 1,\n    b: 2,\n}"); // multiline object using "with"
}

#[test]
fn member() {
    assert_unchanged("object.property"); // member ident
    assert_unchanged(r#"object["property"]"#); // member string literal
}

#[test]
fn array() {
    assert_unchanged(
        r#"a = [1, 2, 3]

a[i]"#,
    );
}

#[test]
fn conditional() {
    assert_unchanged("if a then b else c");
    assert_unchanged(r#"if not a or b and c then 2 / (3 * 2) else obj.a(par: "foo")"#);
}

#[test]
fn return_expr() {
    assert_unchanged("return 42");
}

#[test]
fn option() {
    assert_unchanged("option foo = {a: 1}");
    assert_unchanged(r#"option alert.state = "Warning""#); // qualified
}

#[test]
fn vars() {
    assert_unchanged("0.1"); // float
    assert_unchanged("100000000.0"); // integer float
    assert_unchanged("365d"); // duration
    assert_unchanged("1d1m1s"); // duration_multiple
    assert_unchanged("2018-05-22T19:53:00Z"); // time
    assert_unchanged("2018-05-22T19:53:01+07:00"); // zone
    assert_unchanged("2018-05-22T19:53:23.09012Z"); // nano sec
    assert_unchanged("2018-05-22T19:53:01.09012-07:00"); // nano with zone
    assert_unchanged(r#"/^\w+@[a-zA-Z_]+?\.[a-zA-Z]{2,3}$/"#); // regexp
    assert_unchanged(r#"/^http:\/\/\w+\.com$/"#); // regexp_escape
}

#[test]
fn block() {
    assert_unchanged(
        r#"foo = () => {
    foo(f: 1)
    1 + 1
}"#,
    );
    assert_format(
        r#"foo = 1
foo
builtin bar : int
builtin rab : int
// comment
builtin baz : int"#,
        r#"foo = 1

foo

builtin bar : int
builtin rab : int

// comment
builtin baz : int"#,
    );
}

#[test]
fn str_lit() {
    assert_unchanged(r#""foo""#);
    assert_unchanged(
        r#""this is
a string
with multiple lines""#,
    ); // multi lines
    assert_unchanged(r#""foo \\ \" \r\n""#); // with escape
    assert_unchanged(r#""\xe6\x97\xa5\xe6\x9c\xac\xe8\xaa\x9e""#); // with byte
}

#[test]
fn package_import() {
    assert_unchanged(
        r#"package foo
"#,
    ); // package
    assert_unchanged(
        r#"import "path/foo"
import bar "path/bar""#,
    ); // imports
    assert_unchanged(
        r#"import foo "path/foo"

foo.from(bucket: "testdb")
    |> range(start: 2018-05-20T19:53:26Z)"#,
    ); // no_package
    assert_unchanged(
        r#"package foo


from(bucket: "testdb")
    |> range(start: 2018-05-20T19:53:26Z)"#,
    ); // no_imports
    assert_unchanged(
        r#"package foo


import "path/foo"
import bar "path/bar"

from(bucket: "testdb")
    |> range(start: 2018-05-20T19:53:26Z)"#,
    ); // package import
}

#[test]
fn simple() {
    assert_unchanged(
        r#"from(bucket: "testdb")
    |> range(start: 2018-05-20T19:53:26Z)
    |> filter(fn: (r) => r.name =~ /.*0/)
    |> group(by: ["_measurement", "_start"])
    |> map(fn: (r) => ({_time: r._time, io_time: r._value}))"#,
    );
}

#[test]
fn medium() {
    assert_unchanged(
        r#"from(bucket: "testdb")
    |> range(start: 2018-05-20T19:53:26Z)
    |> filter(fn: (r) => r.name =~ /.*0/)
    |> group(by: ["_measurement", "_start"])
    |> map(fn: (r) => ({_time: r1._time, io_time: r._value}))"#,
    )
}

#[test]
fn complex() {
    assert_unchanged(
        r#"left = from(bucket: "test")
    |> range(start: 2018-05-22T19:53:00Z, stop: 2018-05-22T19:55:00Z)
    |> drop(columns: ["_start", "_stop"])
    |> filter(fn: (r) => r.user == "user1")
    |> group(by: ["user"])
right = from(bucket: "test")
    |> range(start: 2018-05-22T19:53:00Z, stop: 2018-05-22T19:55:00Z)
    |> drop(columns: ["_start", "_stop"])
    |> filter(fn: (r) => r.user == "user2")
    |> group(by: ["_measurement"])

join(tables: {left: left, right: right}, on: ["_time", "_measurement"])"#,
    );
}

#[test]
fn option_complete() {
    assert_unchanged(
        r#"option task = {
    name: "foo",
    every: 1h,
    delay: 10m,
    cron: "02***",
    retry: 5,
}

from(bucket: "test")
    |> range(start: 2018-05-22T19:53:26Z)
    |> window(every: task.every)
    |> group(by: ["_field", "host"])
    |> sum()
    |> to(bucket: "test", tagColumns: ["host", "_field"])"#,
    )
}

#[test]
fn functions_complete() {
    assert_unchanged(
        r#"foo = () => from(bucket: "testdb")
bar = (x=<-) => x
    |> filter(fn: (r) => r.name =~ /.*0/)
baz = (y=<-) => y
    |> map(fn: (r) => ({_time: r._time, io_time: r._value}))

foo()
    |> bar()
    |> baz()"#,
    )
}

#[test]
fn multi_indent() {
    assert_unchanged(
        r#"_sortLimit = (n, desc, columns=["_value"], tables=<-) => tables
    |> sort(columns: columns, desc: desc)
    |> limit(n: n)
_highestOrLowest = (n, _sortLimit, reducer, columns=["_value"], by, tables=<-) => tables
    |> group(by: by)
    |> reducer()
    |> group(none: true)
    |> _sortLimit(n: n, columns: columns)
highestAverage = (n, columns=["_value"], by, tables=<-) => tables
    |> _highestOrLowest(
        n: n,
        columns: columns,
        by: by,
        reducer: (tables=<-) => tables
            |> mean(columns: [columns[0]]),
        _sortLimit: top,
    )"#,
    )
}

#[test]
fn comments() {
    assert_unchanged("// attach to id\nid");
    assert_unchanged("// attach to int\n1");
    assert_unchanged("// attach to float\n1.1");
    assert_unchanged("// attach to string\n\"hello\"");
    assert_unchanged("// attach to regex\n/hello/");
    assert_unchanged("// attach to time\n2020-02-28T00:00:00Z");
    assert_unchanged("// attach to duration\n2m");
    assert_unchanged("// attach to bool\ntrue");
    assert_unchanged("// attach to open paren\n(1 + 1)");
    assert_unchanged("(1 + 1\n    // attach to close paren\n    )");
    assert_unchanged("1 * \n    // attach to open paren\n    (1 + 1)");
    assert_unchanged("1 * (1 + 1\n    // attach to close paren\n    )");
    assert_unchanged(
        "from
    //comment
    (bucket: bucket)",
    );
    assert_unchanged(
        "from(
    //comment
    bucket: bucket)",
    );
    assert_unchanged(
        "from(
    bucket
        //comment
        : bucket,
)",
    );
    assert_unchanged(
        "from(
    bucket: 
        //comment
        bucket,
)",
    );
    assert_unchanged(
        "from(bucket: bucket
    //comment
    )",
    );
    assert_unchanged(
        "from(
    //comment
    bucket)",
    );
    assert_unchanged(
        "from(
    bucket
        //comment
        ,
        _option,
    )",
    );
    assert_unchanged(
        "from(
    bucket,
    //comment
    _option,
)",
    );
    assert_unchanged(
        "from(
    bucket,
    _option,
//comment
)",
    );
    assert_format(
        "from(bucket, _option//comment1
,//comment2
)",
        "from(bucket, _option
    //comment1
    //comment2
)",
    );

    /* Expressions. */
    assert_unchanged("1 \n    //comment\n    <= 1");
    assert_unchanged("1 \n    //comment\n    + 1");
    assert_unchanged("1 \n    //comment\n    * 1");
    assert_unchanged("from()\n    //comment\n    |> to()");
    assert_unchanged("//comment\n+1");
    assert_format("1 * //comment\n-1", "1 * (\n    //comment\n    -1)");
    assert_unchanged("i = \n    //comment\n    not true");
    assert_unchanged("//comment\nexists 1");
    assert_unchanged("a \n    //comment\n    =~ /foo/");
    assert_unchanged("a \n    //comment\n    !~ /foo/");
    assert_unchanged("a \n    //comment\n    and b");
    assert_unchanged("a \n    //comment\n    or b");

    assert_unchanged("a\n    //comment\n     = 1");
    assert_unchanged("//comment\noption a = 1");
    assert_unchanged("option a\n    //comment\n     = 1");
    assert_unchanged("option a\n    //comment\n    .b = 1");
    assert_unchanged("option a.\n    //comment\n    b = 1");
    assert_unchanged("option a.b\n    //comment\n     = 1");

    assert_unchanged("f = \n    //comment\n    (a) => a()");
    assert_unchanged("f = (\n    //comment\n    a) => a()");
    assert_unchanged("f = (\n    //comment\n    a, b) => a()");
    assert_unchanged("f = (a\n    //comment\n    , b) => a()");
    assert_unchanged("f = (a\n    //comment\n    =1, b=2) => a()");
    assert_unchanged("f = (a=\n    //comment\n    1, b=2) => a()");
    assert_unchanged("f = (a=1\n    //comment\n    , b=2) => a()");
    assert_unchanged("f = (a=1, \n    //comment\n    b=2) => a()");
    assert_unchanged("f = (a=1, b\n    //comment\n    =2) => a()");
    assert_unchanged("f = (a=1, b=\n    //comment\n    2) => a()");
    assert_format(
        "f = (a=1, b=2//comment\n,) =>\n    (a())",
        "f = (a=1, b=2\n    //comment\n    ) => a()",
    );
    assert_unchanged("f = (a=1, b=2\n    //comment\n    ) => a()");
    assert_format(
        "f = (a=1, b=2,//comment\n) =>\n    (a())",
        "f = (a=1, b=2\n    //comment\n    ) => a()",
    );
    assert_format(
        "f = (a=1, b=2//comment1\n,//comment2\n) =>\n    (a())",
        "f = (a=1, b=2\n    //comment1\n    //comment2\n    ) => a()",
    );
    assert_unchanged("f = (a=1, b=2) \n    //comment\n    => a()");
    assert_format(
        "f = (x=1, y=2) =>\n    //comment\n(a())",
        "f = (x=1, y=2) => \n    //comment\n    (a())",
    );
    assert_unchanged("f = (a=1, b=2) => \n    //comment\n    a()");

    assert_unchanged("//comment\ntest a = 1");
    assert_unchanged("test \n    //comment\n    a = 1");
    assert_unchanged("test a\n    //comment\n     = 1");
    assert_unchanged("test a = \n    //comment\n    1");

    assert_unchanged("//comment\nreturn x");
    assert_unchanged("return \n    //comment\n    x");

    assert_unchanged("//comment\nif 1 then 2 else 3");
    assert_unchanged("if \n    //comment\n    1 then 2 else 3");
    assert_unchanged("if 1\n    //comment\n     then 2 else 3");
    assert_unchanged("if 1 then \n    //comment\n    2 else 3");
    assert_unchanged("if 1 then 2\n    //comment\n     else 3");
    assert_unchanged("if 1 then 2 else \n    //comment\n    3");

    assert_unchanged("//comment\nfoo[\"bar\"]");
    assert_unchanged("foo\n    //comment\n    [\"bar\"]");
    assert_unchanged("foo[\n    //comment\n    \"bar\"]");
    assert_unchanged("foo[\"bar\"\n    //comment\n    ]");

    assert_unchanged("a = \n    //comment\n    [1, 2, 3]");
    assert_unchanged("a = [\n    //comment\n    1, 2, 3]");
    assert_unchanged("a = [1\n    //comment\n    , 2, 3]");
    assert_unchanged("a = [1, \n    //comment\n    2, 3]");
    assert_unchanged("a = [1, \n    //comment1\n    2\n    //comment2\n    , 3]");
    assert_unchanged("a = [1, 2, 3\n    //comment\n    ]");

    assert_unchanged("a = b\n    //comment\n    [1]");
    assert_unchanged("a = b[\n    //comment\n    1]");
    assert_unchanged("a = b[1\n    //comment\n    ]");

    assert_unchanged(
        "//comment
{_time: r._time, io_time: r._value}",
    );
    assert_unchanged(
        "{
    //comment
    _time: r._time,
    io_time: r._value,
}",
    );
    assert_unchanged(
        "{
    _time
        //comment
        : r._time,
    io_time: r._value,
}",
    );
    assert_unchanged(
        "{
    _time: 
        //comment
        r._time,
    io_time: r._value,
}",
    );
    assert_unchanged(
        "{
    _time: r
        //comment
        ._time,
    io_time: r._value,
}",
    );
    assert_unchanged(
        "{
    _time: r.
        //comment
        _time,
    io_time: r._value,
}",
    );
    assert_unchanged(
        "{
    _time: r
        //comment
        [\"_time\"],
    io_time: r._value,
}",
    );
    assert_unchanged(
        "{
    _time: r._time
        //comment
        ,
        io_time: r._value,
    }",
    );
    assert_unchanged(
        "{
    _time: r._time,
    //comment
    io_time: r._value,
}",
    );
    assert_unchanged(
        "{
    _time: r._time,
    io_time
        //comment
        : r._value,
}",
    );
    assert_unchanged(
        "{
    _time: r._time,
    io_time: 
        //comment
        r._value,
}",
    );
    assert_unchanged(
        "{
    _time: r._time,
    io_time: r
        //comment
        ._value,
}",
    );
    assert_unchanged(
        "{
    _time: r._time,
    io_time: r.
        //comment
        _value,
}",
    );
    assert_unchanged(
        "{
    _time: r._time,
    io_time: r._value,
//comment
}",
    );
    assert_format(
        "{_time: r._time, io_time: r._value
    //comment
    ,}",
        "{
    _time: r._time,
    io_time: r._value
        //comment
        ,
    }",
    );
    assert_format(
        "{_time: r._time, io_time: r._value,
    //comment
    }",
        "{
    _time: r._time,
    io_time: r._value,
//comment
}",
    );

    assert_unchanged("//comment\nimport \"foo\"");
    assert_unchanged("import \n    //comment\n    \"foo\"");
    assert_unchanged("import \n    //comment\n    foo \"foo\"");

    assert_unchanged("//comment\npackage foo\n");
    assert_unchanged("package \n    //comment\n    foo\n");

    assert_unchanged("{\n    //comment\n    foo with\n    a: 1,\n    b: 2,\n}");
    assert_unchanged("{foo\n    //comment\n     with\n    a: 1,\n    b: 2,\n}");
    assert_unchanged("{foo with\n    //comment\n    a: 1,\n    b: 2,\n}");

    assert_unchanged("fn = (tables=\n    //comment\n    <-) => tables");
    assert_unchanged("fn = (tables=<-) => \n    //comment\n    tables");
    assert_unchanged("fn = (tables=<-) => \n    //comment\n    (tables)");

    // Comments around braces needs some work.
    assert_unchanged("fn = (a) => \n    //comment\n    {\n    return a\n}");
    assert_unchanged("fn = (a) => {\n    return a\n// ending\n}");

    assert_format(
        r#"    // hi
// there
{_time: r._time, io_time: r._value, // this is the end
}

// minimal
foo = (arg=[1, 2]) => (1)

// left
left = from(bucket: "test")
    |> range(start: 2018-05-22T19:53:00Z
    // i write too many comments
    , stop: 2018-05-22T19:55:00Z)
    // and put them in strange places
    |>  drop

        // this hurts my eyes
(columns: ["_start", "_stop"])
        // just terrible
    |> filter(fn: (r) =>
        (r.user 

        // (don't fire me, this is intentional)
        == "user1"))
    |> group(by
    // strange place for a comment
: ["user"])

right = from(bucket: "test")
    |> range(start: 2018-05-22T19:53:00Z,
            // please stop
            stop: 2018-05-22T19:55:00Z)
    |> drop( // spare me the pain
// this hurts
columns: ["_start", "_stop"// what
])
    |> filter(
        // just why
        fn: (r) =>
        // user 2 is the best user
        (r.user == "user2"))
    |> group(by: //just stop
["_measurement"])

join(tables: {left: left, right: right}, on: ["_time", "_measurement"])

from(bucket, _option // friends
,// stick together
)

i = // definitely
not true
// a
// list
// of
// comments

j

// not lost
"#,
        r#"// hi
// there
{
    _time: r._time,
    io_time: r._value,
// this is the end
}

// minimal
foo = (arg=[1, 2]) => 1

// left
left = from(bucket: "test")
    |> range(
        start: 2018-05-22T19:53:00Z
            // i write too many comments
            ,
            stop: 2018-05-22T19:55:00Z,
        )
    // and put them in strange places
    |> drop
        // this hurts my eyes
        (columns: ["_start", "_stop"])
    // just terrible
    |> filter(
        fn: (r) => r.user 
            // (don't fire me, this is intentional)
            == "user1",
    )
    |> group(
        by
            // strange place for a comment
            : ["user"],
    )
right = from(bucket: "test")
    |> range(
        start: 2018-05-22T19:53:00Z,
        // please stop
        stop: 2018-05-22T19:55:00Z,
    )
    |> drop(
        // spare me the pain
        // this hurts
        columns: ["_start", "_stop"
            // what
            ],
    )
    |> filter(
        // just why
        fn: (r) => 
            // user 2 is the best user
            (r.user == "user2"),
    )
    |> group(
        by: 
            //just stop
            ["_measurement"],
    )

join(tables: {left: left, right: right}, on: ["_time", "_measurement"])
from(bucket, _option
    // friends
    // stick together
)

i = 
    // definitely
    not true

// a
// list
// of
// comments
j
// not lost
"#,
    );
}

#[test]
fn builtin() {
    assert_unchanged("builtin foo : [int]");
    assert_unchanged("builtin foo : A");
    assert_unchanged("builtin foo : (A: int, B: int) => int");
    assert_unchanged("builtin foo : {A: int, B: int} where A: Addable, B: Divisible");
    assert_unchanged(
        "builtin foo : int

x = 1",
    );
    assert_unchanged("// comment\nbuiltin foo : int");
    assert_unchanged("builtin \n    // comment\n    foo : int");
    assert_unchanged("builtin foo\n    // comment\n    : int");
    assert_unchanged("builtin foo : \n    // comment\n    int");
}

#[test]
fn parens() {
    // test that parens are preserved when needed and removed if uneccessary
    assert_format("(1 * 1)", "1 * 1");
    assert_unchanged("// comment\n(1 * 1)");
    assert_unchanged("(1 * 1\n    // comment\n    )");
    assert_unchanged("1 + (1 * 1)");
    assert_format("(1 + (1 * 1))", "1 + (1 * 1)");
    assert_format("((1 + 1) + 1)", "1 + 1 + 1");
    assert_format("(1 + (1 + 1))", "1 + (1 + 1)");
    assert_unchanged("() => ({_value: 1})");
    assert_unchanged("() => \n    // comment\n    ({_value: 1})");
}

#[test]
fn type_expressions() {
    assert_unchanged(r#"builtin foo : (a: int, b: string) => int"#);
    assert_unchanged(
        r#"builtin foo : (
    a: int,
    b: string,
) => int"#,
    );
    assert_unchanged(r#"builtin foo : {a: int, b: string}"#);
    assert_unchanged(
        r#"builtin foo : {
    a: int,
    b: string,
}"#,
    );
    assert_unchanged(
        r#"builtin foo : {X with
    a: int,
    b: string,
}"#,
    );
    assert_format(
        r#"builtin foo : {a: A, b: B, c: C, d: D, e: E} where A: Numeric, B: Numeric, C: Numeric, D: Numeric, E: Numeric"#,
        r#"builtin foo : {
    a: A,
    b: B,
    c: C,
    d: D,
    e: E,
} where
    A: Numeric,
    B: Numeric,
    C: Numeric,
    D: Numeric,
    E: Numeric"#,
    );
    assert_unchanged(
        r#"builtin foo : (
    a: int,
    b: string,
    c: A,
    d: [int],
    e: [[B]],
    fn: () => int,
) => {x with a: int, b: string} where
    A: Timeable,
    B: Record"#,
    );
}

#[test]
fn temp_indent() {
    // The formatter uses a temporary indent when it finds a comment where
    // the line would normally be on a single line

    assert_unchanged(
        r#"a + 
    // comment
    b"#,
    );
    assert_unchanged(
        r#"call(
    a: 1,
    b: 2,
    // c is special
    c: "special",
)"#,
    );
}

use std::str;

// This gives us a colorful diff.
use pretty_assertions::assert_eq;

use expect_test::{expect, Expect};

use super::*;

#[track_caller]
fn assert_unchanged(script: &str) {
    let _ = env_logger::try_init();
    let output = format(script).unwrap();
    assert_eq!(
        script, output,
        "\n EXPECTED: \n {} \n OUTPUT: \n {} \n",
        script, output
    );
}

#[track_caller]
fn assert_format(script: &str, expected: &str) {
    let _ = env_logger::try_init();
    let output = format(script).unwrap();
    assert_eq!(
        expected, output,
        "\n EXPECTED: \n {} \n OUTPUT: \n {} \n",
        expected, output
    );
}

#[track_caller]
fn expect_format(script: &str, expect: Expect) {
    expect.assert_eq(&format(script).unwrap());
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

    //
    // pipe expressions
    //

    // multiline based on pipe depth
    assert_unchanged(
        r#"(tables) =>
    tables
        |> a()
        |> b()"#,
    );
    // single line
    assert_unchanged(r#"(tables) => tables |> a()"#);
    // multiline based on initial conditions
    assert_unchanged(
        r#"(tables) =>
    tables
        |> a()"#,
    );
}

#[test]
fn call_expr() {
    // call function
    assert_unchanged("a()");
    assert_format("(a)()", "a()");
    // call anonymous function
    assert_unchanged("((a) => a)()");
    // pipe anonymous function
    assert_unchanged("(() => 1) |> f()");
}
#[test]
fn function_expr() {
    assert_unchanged("() => 1 and () => 2");
}

#[test]
fn record() {
    assert_unchanged("{a: 1, b: {c: 11, d: 12}}");
    assert_unchanged("{foo with a: 1, b: {c: 11, d: 12}}"); // with
    assert_unchanged("{a, b, c}"); // implicit key object literal
    assert_unchanged(r#"{"a": 1, "b": 2}"#); // object with string literal keys
    assert_unchanged(r#"{"a": 1, b: 2}"#); // object with mixed keys
    assert_unchanged(
        "[
    {a: 1, b: 2},
    {a: 111, b: 2},
    {a: 1, b: 222},
    {a: 1, b: -892},
    {a: 111, b: 11},
]",
    );
    assert_format(
        "[
    {
        a: 1,
        b: 2,
    },
    {
        a: 111,
        b: 2,
    },
    {
        a: 1,
        b: 222,
    },
    {
        a: 1,
        b: -892,
    },
    {
        a: 111,
        b: 22,
    }
]",
        "[
    {a: 1, b: 2},
    {a: 111, b: 2},
    {a: 1, b: 222},
    {a: 1, b: -892},
    {a: 111, b: 22},
]",
    );
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
fn dict_object() {
    // TODO assert_unchanged(r#"["a": 0, "b": 1]"#);
    assert_unchanged(
        r#"[
    "a": 0,
    //comment
    "b": 1,
]"#,
    );
    assert_unchanged(r#"["a": 0, "b": 1, "c": 2]"#);
    assert_unchanged(r#"[:]"#);
}

#[test]
fn dict_type() {
    assert_unchanged("builtin dict : [string:int]");
    assert_unchanged("builtin dict : [string:string]");
}

#[test]
fn conditional() {
    assert_unchanged("if a then b else c");
    assert_unchanged(r#"if not a or b and c then 2 / (3 * 2) else obj.a(par: "foo")"#);
    assert_unchanged(
        "if x then
    y
else
    z",
    );
    assert_unchanged(
        "if x then
    {a: 1, b: 2}
else
    {a: 2, b: 1}",
    );
    assert_unchanged(
        "if a == b then
    r.x
else if a == c then
    r.y
else
    r.z",
    );
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
fn str_lit1() {
    assert_unchanged(r#""foo""#);
    assert_unchanged(
        r#""this is
a string
with multiple lines""#,
    ); // multi lines
       // StringExpression format textPart with escape sequences
}

#[test]
fn str_lit2() {
    assert_format(
        r#"qux = "{
    \"@foo\": \"bar\",
    \"baz\": ${string(v:json.encode(v:rab))}
    }""#,
        r#"qux = "{
    \"@foo\": \"bar\",
    \"baz\": ${string(v: json.encode(v: rab))}
    }""#,
    );
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
        r#"left =
    from(bucket: "test")
        |> range(start: 2018-05-22T19:53:00Z, stop: 2018-05-22T19:55:00Z)
        |> drop(columns: ["_start", "_stop"])
        |> filter(fn: (r) => r.user == "user1")
        |> group(by: ["user"])
right =
    from(bucket: "test")
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
bar = (x=<-) =>
    x
        |> filter(fn: (r) => r.name =~ /.*0/)
baz = (y=<-) =>
    y
        |> map(fn: (r) => ({_time: r._time, io_time: r._value}))

foo()
    |> bar()
    |> baz()"#,
    )
}

#[test]
fn multi_indent() {
    assert_unchanged(
        r#"_sortLimit = (n, desc, columns=["_value"], tables=<-) =>
    tables
        |> sort(columns: columns, desc: desc)
        |> limit(n: n)

_highestOrLowest = (
    n,
    _sortLimit,
    reducer,
    columns=["_value"],
    by,
    tables=<-,
) =>
    tables
        |> group(by: by)
        |> reducer()
        |> group(none: true)
        |> _sortLimit(n: n, columns: columns)

highestAverage = (n, columns=["_value"], by, tables=<-) =>
    tables
        |> _highestOrLowest(
            n: n,
            columns: columns,
            by: by,
            reducer: (tables=<-) =>
                tables
                    |> mean(columns: [columns[0]]),
            _sortLimit: top,
        )"#,
    )
}

#[test]
fn comments1() {
    assert_unchanged("// attach to id\nid");
    assert_unchanged("// attach to int\n1");
    assert_unchanged("// attach to float\n1.1");
    assert_unchanged("// attach to string\n\"hello\"");
    assert_unchanged("// attach to regex\n/hello/");
    assert_unchanged("// attach to time\n2020-02-28T00:00:00Z");
    assert_unchanged("// attach to duration\n2m");
    assert_unchanged("// attach to bool\ntrue");
    assert_unchanged("// attach to open paren\n(1 + 1)");
    assert_unchanged(
        r#"(1 + 1
    // attach to close paren
    )"#,
    );
    assert_unchanged("1 * // attach to open paren\n    (1 + 1)");
    assert_unchanged("1 * (1 + 1\n        // attach to close paren\n        )");
    assert_unchanged(
        "from
    //comment
    (bucket: bucket)",
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
        "from(bucket: bucket//comment
)",
    );
    assert_unchanged(
        "from(
    //comment
    bucket,
)",
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

    /* Expressions. */
    assert_unchanged("1\n    //comment\n    <=\n    1");
    assert_unchanged("1\n    //comment\n    +\n    1");
    assert_unchanged("1\n    //comment\n    *\n    1");
    assert_unchanged("from()\n    //comment\n    |> to()");
    assert_unchanged("//comment\n+1");
    assert_format("1 * //comment\n-1", "1 * (//comment\n    -1)");
}

#[test]
fn comments2() {
    assert_unchanged("i =\n    //comment\n    not true");
    assert_unchanged("//comment\nexists 1");
    assert_unchanged("a\n    //comment\n    =~\n    /foo/");
    assert_unchanged("a\n    //comment\n    !~\n    /foo/");
    assert_unchanged("a\n    //comment\n    and\n    b");
    assert_unchanged("a\n    //comment\n    or\n    b");

    assert_unchanged("a\n    //comment\n     = 1");
    assert_unchanged("//comment\noption a = 1");
    assert_unchanged("option a\n    //comment\n     = 1");
    assert_unchanged("option a\n    //comment\n    .b = 1");
    assert_unchanged("option a.//comment\nb = 1");
    assert_unchanged("option a.b\n    //comment\n     = 1");

    assert_unchanged("f =\n    //comment\n    (a) => a()");
    assert_unchanged("f = (\n    //comment\n    a,\n    b,\n) =>\n    a()");
    assert_unchanged(
        r"f = (
    a//comment
    ,
    b,
) =>
    a()",
    );
    assert_unchanged(
        r"f = (
    a//comment
    =1,
    b=2,
) =>
    a()",
    );
    assert_format(
        "f = (a=1, b=2//comment\n,) =>\n    (a())",
        r"f = (
    a=1,
    b=2//comment
    ,
) =>
    a()",
    );
    assert_unchanged(
        "f = (
    a=1,
    b=2//comment
    ,
) =>
    a()",
    );
    assert_format(
        "f = (a=1, b=2,//comment\n) =>\n    (a())",
        r"f = (
    a=1,
    b=2,
//comment
) =>
    a()",
    );
    assert_unchanged(
        r"f = (
    a=1,
    b=2,
)
    //comment
    =>
    a()",
    );
    assert_format(
        "f = (x=1, y=2) =>\n    //comment\n(a())",
        "f = (x=1, y=2) =>\n    //comment\n    (a())",
    );
    assert_unchanged("f = (a=1, b=2) =>\n    //comment\n    a()");

    assert_unchanged("//comment\ntest a = 1");
    assert_unchanged("test //comment\na = 1");
    assert_unchanged("test a\n    //comment\n     = 1");
    assert_unchanged("test a =\n    //comment\n    1");

    assert_unchanged("//comment\nreturn x");
    assert_unchanged("return\n    //comment\n    x");

    assert_unchanged("//comment\nif 1 then 2 else 3");
    assert_unchanged("if //comment\n    1\nthen\n    2\nelse\n    3");
    assert_unchanged("if 1//comment\nthen\n    2\nelse\n    3");
    assert_unchanged("if 1 then\n    //comment\n    2\nelse\n    3");
    assert_unchanged("if 1 then\n    2\n//comment\nelse\n    3");
    assert_unchanged("if 1 then\n    2\nelse\n    //comment\n    3");

    assert_unchanged("//comment\nfoo[\"bar\"]");
    assert_unchanged("foo//comment\n[\"bar\"]");
    assert_unchanged("foo[//comment\n\"bar\"]");
    assert_unchanged("foo[\"bar\"\n    //comment\n    ]");

    assert_unchanged("a =\n    //comment\n    [1, 2, 3]");
    assert_unchanged("a = [\n    //comment\n    1,\n    2,\n    3,\n]");
    assert_unchanged("a = [\n    1//comment\n    ,\n    2,\n    3,\n]");
    assert_unchanged("a = [\n    1,\n    //comment\n    2,\n    3,\n]");
    assert_unchanged(
        r"a = [
    1,
    2,
    3,//comment
]",
    );
    assert_unchanged("a =\n    b//comment\n    [1]");
    assert_unchanged("a = b[//comment\n    1]");
    assert_unchanged("a = b[1//comment\n]");

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
    _time:
        r
            //comment
            ._time,
    io_time: r._value,
}",
    );
    assert_unchanged(
        "{
    _time:
        r.//comment
        _time,
    io_time: r._value,
}",
    );
    assert_unchanged(
        "{
    _time:
        r//comment
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
    io_time:
        r
            //comment
            ._value,
}",
    );
    assert_unchanged(
        "{
    _time: r._time,
    io_time:
        r.//comment
        _value,
}",
    );

    assert_unchanged("//comment\nimport \"foo\"");
    assert_unchanged("import //comment\n\"foo\"");
    assert_unchanged("import //comment\nfoo \"foo\"");

    assert_unchanged("//comment\npackage foo\n");
    assert_unchanged("package //comment\nfoo\n");

    assert_unchanged(
        r"{//comment\n
    foo with
    a: 1,
    b: 2,
}",
    );
    assert_unchanged(
        r"{foo//comment
    with
    a: 1,
    b: 2,
}",
    );
    assert_unchanged("{foo with //comment\n    a: 1,\n    b: 2,\n}");

    assert_unchanged("fn = (tables=<-) =>\n    //comment\n    tables");
    assert_unchanged("fn = (tables=<-) =>\n    //comment\n    (tables)");

    assert_unchanged(
        r"fn = (a) =>
//comment
{
    return a
}",
    );
    // Comments around braces needs some work.
    assert_unchanged(
        r"fn = (a) => {
    return a// ending

}",
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
    assert_unchanged("builtin\n    // comment\n    foo : int");
    assert_unchanged("builtin foo\n    // comment\n    : int");
    // assert_unchanged("builtin foo : \n    // comment\n    int");
}

#[test]
fn parens() {
    // test parens are preserved when comments are present
    assert_unchanged("// comment\n(1 * 1)");
    assert_unchanged("(1 * 1\n    // comment\n    )");
    assert_unchanged("() => ({_value: 1})");
    assert_unchanged("() =>\n    // comment\n    ({_value: 1})");

    // test parens are maintained according to operator precedence rules
    assert_format("(2 ^ 2)", "2 ^ 2");
    assert_unchanged("2 * 3 ^ 2");
    assert_unchanged("(2 * 3) ^ 2");
    assert_unchanged("4 / 2 ^ 2");
    assert_unchanged("(4 / 2) ^ 2");
    assert_unchanged("4 % 2 ^ 2");
    assert_unchanged("(4 % 2) ^ 2");
    assert_unchanged("1 + 2 * 3");
    assert_unchanged("(1 + 2) * 3");
    assert_unchanged("1 - 2 * 3");
    assert_unchanged("(1 - 2) * 3");
    assert_format("(1 + (2 * 3))", "1 + 2 * 3");
    assert_format("((1 + 2) + 3)", "1 + 2 + 3");
    assert_format("(1 + (2 + 3))", "1 + (2 + 3)");
    assert_unchanged("1 + 2 < 4");
    assert_format("(1 + 2) < 4", "1 + 2 < 4");
    assert_format("(1 + 2) <= 4", "1 + 2 <= 4");
    assert_format("(1 + 2) > 4", "1 + 2 > 4");
    assert_format("(1 + 2) >= 4", "1 + 2 >= 4");
    assert_format("((1 == 2) and (exists r.a))", "1 == 2 and exists r.a");
    assert_format(
        "((1 == 2) and (not exists r.a))",
        "1 == 2 and not exists r.a",
    );
    assert_format("((1 == 2) and (exists r.a))", "1 == 2 and exists r.a");

    assert_unchanged("a and b or c");
    assert_format("(a and b) or c", "a and b or c");
    assert_unchanged("a and (b or c)");
    assert_unchanged("a and (b or c) or d");
    assert_unchanged("a and b or c");
    assert_format("((a) and ((b or c) or d))", "a and (b or c or d)");

    assert_unchanged("(a() |> b()).c");
    assert_format("(a() |> b()) ^ 3", "a() |> b() ^ 3");
    assert_format("1 ^ (a() |> b())", "1 ^ a() |> b()");
    assert_unchanged("(1 ^ a()) |> b()");
    assert_unchanged(r#"qux = (r) => "foo: ${r._rab} is: " + (if r.bar then "bar" else "baz")"#);
}

#[test]
fn type_expressions() {
    assert_unchanged(r#"builtin foo : (a: int, b: string) => int"#);
    assert_unchanged(r#"builtin foo : {a: int, b: string}"#);
    assert_format(
        r#"builtin foo : {a: A, b: B, c: C, d: D, e: E} where A: Numeric, B: Numeric, C: Numeric, D: Numeric, E: Numeric"#,
        r#"builtin foo : {
        a: A,
        b: B,
        c: C,
        d: D,
        e: E,
    }
    where
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
    ) => {x with a: int, b: string}
    where
    A: Timeable,
    B: Record"#,
    );
}

#[test]
fn testcase() {
    assert_unchanged(
        r#"testcase a {
    assert.equal(want: 4, got: 2 + 2)
}"#,
    );
    assert_unchanged(
        r#"testcase a extends "other_test" {
    assert.equal(want: 4, got: 2 + 2)
}"#,
    );
    assert_format(
        r#"testcase a { assert.equal(want: 4, got: 2 + 2) }"#,
        r#"testcase a {
    assert.equal(want: 4, got: 2 + 2)
}"#,
    );
    assert_format(
        r#"testcase a extends "other_test" { assert.equal(want: 4, got: 2 + 2) }"#,
        r#"testcase a extends "other_test" {
    assert.equal(want: 4, got: 2 + 2)
}"#,
    );
    assert_format(
        r#"// comment
testcase x
{
    x = 1
    }"#,
        r#"// comment
testcase x {
    x = 1
}"#,
    );
}

#[test]
fn testcase_indentation() {
    assert_unchanged(
        r#"testcase basic {
    inData =
        "
#datatype,string,long,dateTime:RFC3339,long,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,host
,,0,2018-05-22T19:53:26Z,100,load1,system,host.local
"
    outData =
        "
#datatype,string,long,double
#group,false,false,false
#default,_result,,
,result,table,newValue
,,0,100.0
"

    got =
        csv.from(csv: inData)
            |> testing.load()
            |> range(start: 2018-05-22T19:53:26Z)
            |> map(fn: (r) => ({newValue: float(v: r._value)}))
    want = csv.from(csv: outData)

    testing.diff(want: want, got: got) |> yield()
}"#,
    );
}

#[test]
fn temp_indent() {
    // The formatter uses a temporary indent when it finds a comment where
    // the line would normally be on a single line

    assert_unchanged(
        r#"a + // comment
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
#[test]
fn else_indentation() {
    assert_unchanged(
        "tables
    |> map(
        fn: (r) =>
            ({r with level_value:
                    if r._level == levelCrit then
                        4
                    else if r._level == levelWarn then
                        3
                    else if r._level == levelInfo then
                        2
                    else if r._level == levelOK then
                        1
                    else
                        0,
                foo: bar,
            }),
    )",
    );
    assert_unchanged(
        "tables
    |> map(
        fn: (r) =>
            ({r with level_value:
                    if r._level == levelCrit then
                        4
                    else if r._level == levelWarn then
                        3
                    else if r._level == levelInfo then
                        2
                    else if r._level == levelOK then
                        1
                    else
                        0,
            }),
    )",
    );
    assert_unchanged(
        "if x then
    y
else if g then
    7
else if x then
    9
else if z then
    42
else if g then
    7
else if x then
    9
else if z then
    42
else
    z",
    );
}

#[test]
fn line_separation() {
    assert_unchanged(
        r#"inData =
    "
#datatype,string,long,string,string,string,string,double,dateTime:RFC3339
#group,false,false,true,true,true,true,false,false
#default,_result,,,,,,,
,result,table,_field,_measurement,cpu,host,_value,_time
,,0,usage_guest,cpu,cpu-total,ip-192-168-1-16.ec2.internal,0,2020-10-09T22:18:00Z
,,0,usage_guest,cpu,cpu-total,ip-192-168-1-16.ec2.internal,0,2020-10-09T22:19:00Z
,,0,usage_guest,cpu,cpu-total,ip-192-168-1-16.ec2.internal,0,2020-10-09T22:19:44.191958Z
,,1,usage_idle,cpu,cpu-total,ip-192-168-1-16.ec2.internal,94.62634341438049,2020-10-09T22:18:00Z
,,1,usage_idle,cpu,cpu-total,ip-192-168-1-16.ec2.internal,92.28242486302014,2020-10-09T22:19:00Z
,,1,usage_idle,cpu,cpu-total,ip-192-168-1-16.ec2.internal,91.15346397579125,2020-10-09T22:19:44.191958Z
,,2,usage_system,cpu,cpu-total,ip-192-168-1-16.ec2.internal,2.0994751312170705,2020-10-09T22:18:00Z
,,2,usage_system,cpu,cpu-total,ip-192-168-1-16.ec2.internal,2.5586762674700636,2020-10-09T22:19:00Z
,,2,usage_system,cpu,cpu-total,ip-192-168-1-16.ec2.internal,2.6547010580713986,2020-10-09T22:19:44.191958Z

#datatype,string,long,string,string,string,string,string,string,string,double,dateTime:RFC3339
#group,false,false,true,true,true,true,true,true,true,false,false
#default,_result,,,,,,,,,,
,result,table,_field,_measurement,device,fstype,host,mode,path,_value,_time
,,3,inodes_free,disk,disk1s1,apfs,ip-192-168-1-16.ec2.internal,rw,/System/Volumes/Data,4878333294,2020-10-09T22:18:00Z
,,3,inodes_free,disk,disk1s1,apfs,ip-192-168-1-16.ec2.internal,rw,/System/Volumes/Data,4878333286,2020-10-09T22:19:00Z
,,3,inodes_free,disk,disk1s1,apfs,ip-192-168-1-16.ec2.internal,rw,/System/Volumes/Data,4878333253.4,2020-10-09T22:19:44.191958Z
,,4,inodes_total,disk,disk1s1,apfs,ip-192-168-1-16.ec2.internal,rw,/System/Volumes/Data,4882452840,2020-10-09T22:18:00Z
,,4,inodes_total,disk,disk1s1,apfs,ip-192-168-1-16.ec2.internal,rw,/System/Volumes/Data,4882452840,2020-10-09T22:19:00Z
,,4,inodes_total,disk,disk1s1,apfs,ip-192-168-1-16.ec2.internal,rw,/System/Volumes/Data,4882452840,2020-10-09T22:19:44.191958Z
,,5,inodes_used,disk,disk1s1,apfs,ip-192-168-1-16.ec2.internal,rw,/System/Volumes/Data,4119546,2020-10-09T22:18:00Z
,,5,inodes_used,disk,disk1s1,apfs,ip-192-168-1-16.ec2.internal,rw,/System/Volumes/Data,4119554,2020-10-09T22:19:00Z
,,5,inodes_used,disk,disk1s1,apfs,ip-192-168-1-16.ec2.internal,rw,/System/Volumes/Data,4119586.6,2020-10-09T22:19:44.191958Z"

outData =
    "
#group,false,false,true,false,false,false,false,false,false,false,false,false
#datatype,string,long,string,string,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,double,double,double,double
#default,want,,,,,,,,,,,
,result,table,host,_measurement,_start,_stop,_time,cpu,inodes_free,usage_guest,usage_idle,usage_system
,,0,ip-192-168-1-16.ec2.internal,cpu,2020-10-01T00:00:00Z,2030-01-01T00:00:00Z,2020-10-09T22:20:00Z,cpu-total,4878333253.4,0,91.15346397579125,2.6547010580713986""#,
    );
    assert_unchanged(
        "test1 = 1
test2 = 2",
    );
    assert_unchanged(
        r#"fn =
    if nfields == 0 then
        (r) => true
    else
        (r) => contains(value: r._field, set: fields)

return
    tables
        |> filter(fn)
        |> v1.fieldsAsCols()
        |> _mask(columns: ["_measurement", "_start", "_stop"])"#,
    );
    assert_unchanged(
        "a = () => {
    test1 = 1

    test2 = 2
}",
    );
    assert_format(
        "test1 = 1



test2 = 2",
        "test1 = 1

test2 = 2",
    );
    assert_format(
        "test1 = 1
test2 = 2",
        "test1 = 1
test2 = 2",
    );
}
#[test]
fn preserve_multiline_test() {
    // ensure functions given preserve their structure
    //assert_unchanged("test _convariance_missing_column_2 = () =>
    //({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: covariance_missing_column_2})");

    assert_unchanged(
        "test _covariance_missing_column_2 = () =>
    ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: covariance_missing_column_2})",
    );

    assert_unchanged(
        r#"event = (
        url,
        username,
        password,
        action="EventsRouter",
        methods="add_event",
        type="rpc",
        tid=1,
        summary="",
        device="",
        component="",
        severity,
        eventClass="",
        eventClassKey="",
        collector="",
        message="",
    ) =>
    {
        body = json.encode(v: payload)

        return http.post(headers: headers, url: url, data: body)
    }"#,
    );

    //Checks that a method with <= 4 params does not get reformatted
    assert_unchanged(
        r#"event = (url, message="") => {
    body = json.encode(v: payload)

    return http.post(headers: headers, url: url, data: body)
}"#,
    );
}

#[test]
fn preserve_multiline_test_2() {
    //Checks that a method with >4 params gets expanded correctly
    expect_format(
        r#"selectWindow = (column="_value", fn, as, every, defaultValue, tables=<-) => {
    _column = column
    _as = as

    return tables
        |> aggregateWindow(every: every, fn: fn, column: _column, createEmpty: true)
        |> fill(column: _column, value: defaultValue)
        |> rename(fn: (column) => if column == _column then _as else column)}"#,
        expect![[r#"selectWindow = (
        column="_value",
        fn,
        as,
        every,
        defaultValue,
        tables=<-,
    ) =>
    {
        _column = column
        _as = as

        return
            tables
                |> aggregateWindow(every: every, fn: fn, column: _column, createEmpty: true)
                |> fill(column: _column, value: defaultValue)
                |> rename(fn: (column) => if column == _column then _as else column)
    }"#]],
    );
}

#[test]
fn tab_literals() {
    assert_unchanged(
        "// This is a comment with a literal tabstop character
//	<- that is a tab
a",
    );
    assert_unchanged(r#"a = "a string literal with a tabstop '	'""#);
}

#[test]
fn invalid_syntax() {
    let script = r#"builtin diff : (
        <-got: [A],
        want: [A],
        ?verbose: bool,
        ?epsilon: float,
        ?nansEqual: bool,
        aoeustahoesih
    ) => [{A with _diff: string}:]"#;

    assert!(format(script).is_err());
}

#[test]
fn consistent_multiline_formatting_1() {
    // builtins
    assert_format(
        r#"builtin getGrid : (
    region: T,?minSize: int,?maxSize: int,?level: int,?maxLevel: int,units: {distance: string},) => {level: int, set: [string]} where
    T: Record"#,
        r#"builtin getGrid : (
        region: T,
        ?minSize: int,
        ?maxSize: int,
        ?level: int,
        ?maxLevel: int,
        units: {distance: string},
    ) => {level: int, set: [string]}
    where
    T: Record"#,
    );

    assert_format(
        r#"builtin fakeFunc : (v: string, w: bool, x: string,
    y: int, z: string) => string"#,
        r#"builtin fakeFunc : (
        v: string,
        w: bool,
        x: string,
        y: int,
        z: string,
    ) => string"#,
    );

    // function expressions and call expressions
    assert_unchanged(
        r#"ST_Contains = (region, geometry, units=units) => stContains(region: region, geometry: geometry, units: units)"#,
    );

    assert_format(
        r#"ST_Contains = (region, geometry, units=units, more, stuff=stuff, is, better=nice) => stContains(region: region, geometry: geometry, units: units, more:more, stuff:stuff, is:is,better:better)"#,
        r#"ST_Contains = (
    region,
    geometry,
    units=units,
    more,
    stuff=stuff,
    is,
    better=nice,
) =>
    stContains(
        region: region,
        geometry: geometry,
        units: units,
        more: more,
        stuff: stuff,
        is: is,
        better: better,
    )"#,
    );

    assert_format(
        r#"asTracks = (tables=<-, groupBy=["id", "tid", "foo", "bar", "baz"], orderBy=["_time"]) =>
    tables
        |> group(columns: groupBy)
        |> sort(columns: orderBy)
"#,
        r#"asTracks = (
    tables=<-,
    groupBy=[
        "id",
        "tid",
        "foo",
        "bar",
        "baz",
    ],
    orderBy=["_time"],
) =>
    tables
        |> group(columns: groupBy)
        |> sort(columns: orderBy)"#,
    );

    // record expressions
    assert_unchanged(r#"option units = {distance: "km"}"#);

    assert_format(
        r#"option units = {distance: "km", length: "cm", speed: "km/h", volume: "ml", currency: "GBP", country: "United Kingdom"}"#,
        r#"option units = {
    distance: "km",
    length: "cm",
    speed: "km/h",
    volume: "ml",
    currency: "GBP",
    country: "United Kingdom",
}"#,
    );

    // array expressions
    assert_unchanged(
        r#"import "array"

array.from(
    rows: [
        {_value: "one", _time: 2021-08-30T00:00:00Z},
        {_value: "one", _time: 2021-08-30T00:00:00Z},
        {_value: "one", _time: 2021-08-30T00:00:00Z},
        {_value: "one", _time: 2021-08-30T00:00:00Z},
        {_value: "one", _time: 2021-08-30T00:00:00Z},
        {_value: "one", _time: 2021-08-30T00:00:00Z},
        {_value: "one", _time: 2021-08-30T00:00:00Z},
    ],
    foo: false,
)"#,
    );

    assert_format(
        r#"import "array"

array.from(rows: [{_value: "one", _time: 2021-08-30T00:00:00Z}, {_value: "one", _time: 2021-08-30T00:00:00Z}, {_value: "one", _time: 2021-08-30T00:00:00Z},
    {_value: "one", _time: 2021-08-30T00:00:00Z},
    {_value: "one", _time: 2021-08-30T00:00:00Z},
    {_value: "one", _time: 2021-08-30T00:00:00Z},
    {_value: "one", _time: 2021-08-30T00:00:00Z},
], foo: false)"#,
        r#"import "array"

array.from(
    rows: [
        {_value: "one", _time: 2021-08-30T00:00:00Z},
        {_value: "one", _time: 2021-08-30T00:00:00Z},
        {_value: "one", _time: 2021-08-30T00:00:00Z},
        {_value: "one", _time: 2021-08-30T00:00:00Z},
        {_value: "one", _time: 2021-08-30T00:00:00Z},
        {_value: "one", _time: 2021-08-30T00:00:00Z},
        {_value: "one", _time: 2021-08-30T00:00:00Z},
    ],
    foo: false,
)"#,
    );

    assert_format(
        r#"import "array"

array.from(rows: [
    {_value: "one", _time: 2021-08-30T00:00:00Z},
    {_value: "one", _time: 2021-08-30T00:00:00Z},
    {_value: "one", _time: 2021-08-30T00:00:00Z},
    {_value: "one", _time: 2021-08-30T00:00:00Z},
    {_value: "one", _time: 2021-08-30T00:00:00Z},
    {_value: "one", _time: 2021-08-30T00:00:00Z},
    {_value: "one", _time: 2021-08-30T00:00:00Z},
], foo: false)"#,
        r#"import "array"

array.from(
    rows: [
        {_value: "one", _time: 2021-08-30T00:00:00Z},
        {_value: "one", _time: 2021-08-30T00:00:00Z},
        {_value: "one", _time: 2021-08-30T00:00:00Z},
        {_value: "one", _time: 2021-08-30T00:00:00Z},
        {_value: "one", _time: 2021-08-30T00:00:00Z},
        {_value: "one", _time: 2021-08-30T00:00:00Z},
        {_value: "one", _time: 2021-08-30T00:00:00Z},
    ],
    foo: false,
)"#,
    );
}

#[test]
fn consistent_multiline_formatting_2() {
    // check that formatter output does not change
    assert_unchanged(
        r#"array.from(
    rows: [
        {_value: "one", _time: 2021-08-30T00:00:00Z},
        {_value: "one", _time: 2021-08-30T00:00:00Z},
        {_value: "one", _time: 2021-08-30T00:00:00Z},
        {_value: "one", _time: 2021-08-30T00:00:00Z},
        {_value: "one", _time: 2021-08-30T00:00:00Z},
        {_value: "one", _time: 2021-08-30T00:00:00Z},
        {_value: "one", _time: 2021-08-30T00:00:00Z},
    ],
    foo: false,
)"#,
    );

    assert_unchanged(
        r#"// diff = |Xi - MEDiXi| = math.abs(xi-med(xi))
diff =
    join(tables: {data: data, med: med}, on: ["_time"], method: "inner")
        |> map(fn: (r) => ({r with _value: math.abs(x: r._value_data - r._value_med)}))
        |> drop(columns: ["_start", "_stop", "_value_med", "_value_data"])"#,
    );

    assert_format(
        r#"// diff = |Xi - MEDiXi| = math.abs(xi-med(xi))
diff =
    join(tables: { data: [_some: data, _more: data],med: med,foo:foo, bar:bar,multi: multi },on: ["_time"],method: "inner" )
        |> map(fn: (r) => ({r with _value: math.abs(x: r._value_data - r._value_med)}))
        |> drop(columns: ["_start", "_stop", "_value_med", "_value_data"])"#,
        r#"// diff = |Xi - MEDiXi| = math.abs(xi-med(xi))
diff =
    join(
        tables: {
            data: [_some: data, _more: data],
            med: med,
            foo: foo,
            bar: bar,
            multi: multi,
        },
        on: ["_time"],
        method: "inner",
    )
        |> map(fn: (r) => ({r with _value: math.abs(x: r._value_data - r._value_med)}))
        |> drop(columns: ["_start", "_stop", "_value_med", "_value_data"])"#,
    );

    assert_unchanged(
        r#"t_mad = (table=<-) =>
    table
        |> range(start: 2020-04-27T00:00:00Z, stop: 2020-05-01T00:00:00Z)
        |> anomalydetection.mad(threshold: 3.0)

test _mad = () => ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_mad})"#,
    );

    // comments
    assert_format(
        r#"// diff = |Xi - MEDiXi| = math.abs(xi-med(xi))
diff = join(tables: { data: [_some: data, _more: data],med: med,foo:foo, bar:bar,multi: multi },on: ["_time"],method: "inner" )
    |> map(fn: (r) => ({r with _value: math.abs(x: r._value_data - r._value_med)}))
    |> drop(columns: ["_start", "_stop", "_value_med", "_value_data"])"#,
        r#"// diff = |Xi - MEDiXi| = math.abs(xi-med(xi))
diff =
    join(
        tables: {
            data: [_some: data, _more: data],
            med: med,
            foo: foo,
            bar: bar,
            multi: multi,
        },
        on: ["_time"],
        method: "inner",
    )
        |> map(fn: (r) => ({r with _value: math.abs(x: r._value_data - r._value_med)}))
        |> drop(columns: ["_start", "_stop", "_value_med", "_value_data"])"#,
    );

    // make sure that properly formatted text does not change
    assert_unchanged(
        r#"// diff = |Xi - MEDiXi| = math.abs(xi-med(xi))
diff =
    join(
        tables: {
            data: [_some: data, _more: data],
            med: med,
            foo: foo,
            bar: bar,
            multi: multi,
        },
        on: ["_time"],
        method: "inner",
    )
        |> map(fn: (r) => ({r with _value: math.abs(x: r._value_data - r._value_med)}))
        |> drop(columns: ["_start", "_stop", "_value_med", "_value_data"])"#,
    );

    assert_unchanged(
        r#"test _join_panic = () =>
    // to trigger the panic, switch the testing.loadStorage() csv from `passData` to `failData`
    ({input: testing.loadStorage(csv: passData), want: testing.loadMem(csv: outData), fn: t_join_panic})"#,
    );

    assert_format(
        r#"[
"a": 0,
    //comment
    "b": 1,
    ]"#,
        r#"[
    "a": 0,
    //comment
    "b": 1,
]"#,
    );

    assert_unchanged(
        r#"check = {
    _check_id: "rate-check",
    _check_name: "Rate Check",
    // tickscript?
    _type: "custom",
    tags: {},
}"#,
    );

    assert_format(
        r#"[
"a": 0,
    //comment
    "b": 1,
    ]"#,
        r#"[
    "a": 0,
    //comment
    "b": 1,
]"#,
    );

    assert_format(
        r#"[
"a": 0,
    //multilinecomment
    // multiline comment
    "b": 1,
    ]"#,
        r#"[
    "a": 0,
    //multilinecomment
    // multiline comment
    "b": 1,
]"#,
    );

    // string literals
    assert_unchanged(
        r#"want =
    csv.from(
        csv:
            "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string
#group,false,false,true,true,false,false,true,true,true
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,0,2018-05-22T19:53:00Z,2018-05-22T19:58:00Z,2018-05-22T19:58:00Z,,load1,system,host.local
,,0,2018-05-22T19:53:00Z,2018-05-22T19:58:00Z,2018-05-22T19:55:00Z,1.775,load1,system,host.local
",
    )"#,
    );
    assert_unchanged(
        r#"data =
    "
#datatype,string,long,string,string,string,dateTime:RFC3339,boolean
#group,false,false,false,false,false,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,0,m0,f0,tagvalue,2018-12-19T22:13:30Z,false"

check = {
    _check_id: "rate-check",
    _check_name: "Rate Check",
    // tickscript?
    _type: "custom",
    tags: {},
}"#,
    );
}

#[test]
fn consistent_multiline_formatting_3() {
    assert_unchanged(
        r#"string_lit = "
    this
    is a
    multiline
    string
"

// make sure that multiline string works
t_dict = (table=<-) =>
    table
        |> range(start: 2018-05-22T19:53:26Z)
        |> drop(columns: ["_start", "_stop"])
        |> map(
            fn: (r) => {
                //and doesn't do weird stuff
                return {r with error_code: error_code}
            },
        )"#,
    );
}

#[test]
#[ignore]
// this test is broken because comments aren't working right.
fn _ignored_tests() {
    // builtins
    assert_unchanged(
        r#"foo = {
    b: 2,
    c: () => {
        // line 1
        a = 0
        return "a
multi
line
string"
    },
}"#,
    );
    assert_unchanged(
        "from(bucket, _option
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
    assert_unchanged("f = (\n    //comment\n    a) => a()");
    assert_unchanged("f = (a=\n    //comment\n    1, b=2) => a()");
    assert_unchanged("f = (a=1\n    //comment\n    , b=2) => a()");
    assert_unchanged("f = (a=1, \n    //comment\n    b=2) => a()");
    assert_unchanged("f = (a=1, b\n    //comment\n    =2) => a()");
    assert_unchanged("f = (a=1, b=\n    //comment\n    2) => a()");
    assert_format(
        "f = (a=1, b=2//comment1\n,//comment2\n) =>\n    (a())",
        "f = (a=1, b=2\n    //comment1\n    //comment2\n    ) => a()",
    );
    assert_unchanged("a = [\n    1,\n    //comment1\n    2\n    //comment2\n    ,\n    3,\n]");
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
    assert_unchanged(
        "{
    _time: r._time,
    io_time: r._value,
//comment
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
    assert_unchanged("fn = (tables=\n    //comment\n    <-) => tables");
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
        columns: [
            "_start",
            "_stop",
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
fn dont_alter_string_literal() {
    let script = r#"sendSlackMessage(
    text: "*Earthquake Alert*
M *${string(v: r.mag)}*",
)"#;
    assert_unchanged(script);
}

#[test]
fn builtin_are_unchanged() {
    let script = r#"builtin covariance : (<-tables: [A], ?pearsonr: bool, ?valueDst: string, columns: [string]) => [B]
    where
    A: Record,
    B: Record
builtin cumulativeSum : (<-tables: [A], ?columns: [string]) => [B] where A: Record, B: Record"#;
    assert_unchanged(script);
}

#[test]
fn astutil_test_format_with_comments() {
    let src = r#"    // hi
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

    // not lost"#;

    expect![[r#"
        // hi
        // there
        {
            _time: r._time,
            io_time: r._value,
                // this is the end

        }

        // minimal
        foo = (arg=[1, 2]) => 1

        // left
        left =
            from(bucket: "test")
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
                    fn: (r) =>
                        r.user
                            // (don't fire me, this is intentional)
                            ==
                            "user1",
                )
                |> group(
                    by
                        // strange place for a comment
                        : ["user"],
                )

        right =
            from(bucket: "test")
                |> range(
                    start: 2018-05-22T19:53:00Z,
                    // please stop
                    stop: 2018-05-22T19:55:00Z,
                )
                |> drop(
                    // spare me the pain
                    // this hurts
                    columns:
                        [
                            "_start",
                            "_stop",// what
                        ],
                )
                |> filter(
                    // just why
                    fn:
                        (r) =>
                            // user 2 is the best user
                            (r.user == "user2"),
                )
                |> group(
                    by:
                        //just stop
                        ["_measurement"],
                )

        join(tables: {left: left, right: right}, on: ["_time", "_measurement"])

        from(
            bucket,
            _option
                // friends
                ,
        // stick together
        )

        i =
            // definitely
            not true

        // a
        // list
        // of
        // comments
        j// not lost"#]]
    .assert_eq(&format(src).unwrap());
}

#[test]
fn format_long_single_line_pipe_expression() {
    let src = r#"from(bucket: "foo") |> range(start: -1d, stop: now()) |> filter(fn: (r) => r._field == "usage_user") |> aggregateWindow(every: 1m, fn: mean) |> yield()"#;
    expect_format(
        src,
        expect![[r#"
        from(bucket: "foo")
            |> range(start: -1d, stop: now())
            |> filter(fn: (r) => r._field == "usage_user")
            |> aggregateWindow(every: 1m, fn: mean)
            |> yield()"#]],
    );
}

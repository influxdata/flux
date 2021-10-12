use crate::semantic::convert_source;
use crate::semantic::formatter::format;
use pretty_assertions::assert_eq;

#[test]
fn literals() {
    let sem_pkg = convert_source(
        r#"
            a = "Hello, World!"
            b = 12
            c = 18.5
            d = 12h
            e = 2019-10-31T00:00:00Z
            f = /server[01]/
        "#,
    )
    .unwrap();

    let got = format(&sem_pkg).unwrap();

    let want = r#"package main
a = "Hello, World!":string
b = 12:int
c = 18.5:float
d = 043200000000000false:duration
e = 2019-10-31T00:00:00Z:datetime
f = /server[01]/:regex"#;

    assert_eq!(want, got);
}

#[test]
fn array_lit() {
    let sem_pkg = convert_source(
        r#"
a = [1, 2, 3]
b = [1.1, 2.2, 3.3]
c = ["1", "2", "3"]
d = [1s, 2m, 3h]
e = [2019-10-31T00:00:00Z]
f = [/a/, /b/, /c/]
g = [{a:0, b:0.0}, {a:1, b:1.1}]
        "#,
    )
    .unwrap();

    let got = format(&sem_pkg).unwrap();

    let want = r#"package main
a = [1:int2:int3:int]
b = [1.1:float2.2:float3.3:float]
c = ["1":string"2":string"3":string]
d = [01000000000false:duration0120000000000false:duration010800000000000false:duration]
e = [2019-10-31T00:00:00Z:datetime]
f = [/a/:regex/b/:regex/c/:regex]
g = [{a: 0:int, b: 0.0:float}{a: 1:int, b: 1.1:float}]"#;

    assert_eq!(want, got);
}

#[test]
fn dictionary_literals() {
    let sem_pkg = convert_source(
        r#"
m = ["a": 0, "b": 1, "c": 2]
m = [1970-01-01T00:00:00Z: 0, 1970-01-01T01:00:00Z: 1]
        "#,
    )
    .unwrap();

    let got = format(&sem_pkg).unwrap();

    let want = r#"package main
m = ["a":string: 0:int"b":string: 1:int"c":string: 2:int]
m = [1970-01-01T00:00:00Z:datetime: 0:int1970-01-01T01:00:00Z:datetime: 1:int]"#;

    assert_eq!(want, got);
}

#[test]
fn identifier_expressions() {
    let sem_pkg = convert_source(
        r#"
n = 1.0
f = n + 3.0
        "#,
    )
    .unwrap();

    let got = format(&sem_pkg).unwrap();

    let want = r#"package main
n = 1.0:float
f = n:float +:float 3.0:float"#;

    assert_eq!(want, got);
}

#[test]
fn function_default_arguments() {
    let sem_pkg = convert_source(
        r#"
            f = (a, b=1) => a + b
            x = f(a:2)
            y = f(a: x, b: f(a:x))
        "#,
    )
    .unwrap();

    let got = format(&sem_pkg).unwrap();

    let want = r#"package main
f = (a, b=1:int) =>
x = f:(a:int, ?b:int) => int(a: 2:int)
y = f:(a:int, ?b:int) => int(a: x:int, b: f:(a:int, ?b:int) => int(a: x:int))"#;

    assert_eq!(want, got);
}

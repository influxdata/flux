use crate::semantic::convert_source;
use crate::semantic::formatter::format;
use expect_test::{expect, Expect};

fn check(actual: &str, expect: Expect) {

    let sem_pkg = convert_source(actual).unwrap();
    let actual = format(&sem_pkg).unwrap();
    
    expect.assert_eq(&actual);
}

#[test]
fn literals() {
    let script = r#"
            a = "Hello, World!"
            b = 12
            c = 18.5
            d = 12h
            e = 2019-10-31T00:00:00Z
            f = /server[01]/
            "#;

check(script, expect![[r#"
    package main
    a = "Hello, World!":string
    b = 12:int
    c = 18.5:float
    d = 043200000000000false:duration
    e = 2019-10-31T00:00:00Z:time
    f = /server[01]/:regex"#]])
}

#[test]
fn array_lit() {
    let script = r#"
            a = [1, 2, 3]
            b = [1.1, 2.2, 3.3]
            c = ["1", "2", "3"]
            d = [1s, 2m, 3h]
            e = [2019-10-31T00:00:00Z]
            f = [/a/, /b/, /c/]
            g = [{a:0, b:0.0}, {a:1, b:1.1}]
            "#;

check(script, expect![[r#"
    package main
    a = [1:int, 2:int, 3:int]:[int]
    b = [1.1:float, 2.2:float, 3.3:float]:[float]
    c = ["1":string, "2":string, "3":string]:[string]
    d = [01000000000false:duration, 0120000000000false:duration, 010800000000000false:duration]:[duration]
    e = [2019-10-31T00:00:00Z:time]:[time]
    f = [/a/:regex, /b/:regex, /c/:regex]:[regexp]
    g = [{a: 0:int, b: 0.0:float}, {a: 1:int, b: 1.1:float}]:[{a:int, b:float}]"#]])
}

#[test]
fn dictionary_literals() {
    let script = r#"
            m = ["a": 0, "b": 1, "c": 2]
            m = [1970-01-01T00:00:00Z: 0, 1970-01-01T01:00:00Z: 1]
            "#;

check(script, expect![[r#"
    package main
    m = ["a":string: 0:int, "b":string: 1:int, "c":string: 2:int]
    m = [1970-01-01T00:00:00Z:time: 0:int, 1970-01-01T01:00:00Z:time: 1:int]"#]])
}

#[test]
fn identifier_expressions() {
    let script = r#"
            n = 1.0
            f = n + 3.0
            "#;

check(script, expect![[r#"
    package main
    n = 1.0:float
    f = n:float +:float 3.0:float"#]])
}

#[test]
fn function_default_arguments() {
    let script = r#"
            f = (a, b=1) => a + b
            x = f(a:2)
            y = f(a: x, b: f(a:x))
            "#;

check(script, expect![[r#"
    package main
    f = (a, b=1:int) =>
    x = f:(a:int, ?b:int) => int(a: 2:int)
    y = f:(a:int, ?b:int) => int(a: x:int, b: f:(a:int, ?b:int) => int(a: x:int))"#]])
}
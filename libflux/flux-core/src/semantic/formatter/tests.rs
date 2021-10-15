use crate::semantic::env::Environment;
use crate::semantic::formatter::format;
use crate::semantic::types::{Function, MonoType, PolyTypeMap, SemanticMap, Tvar};
use crate::semantic::Analyzer;
use expect_test::{expect, Expect};

fn check(actual: &str, expect: Expect) {
    let mut analyzer = Analyzer::new(Environment::default(), PolyTypeMap::new());
    let (_, mut sem_pkg) = analyzer
        .analyze_source("main".to_string(), "main.flux".to_string(), actual)
        .unwrap();
    let actual = format(&sem_pkg).unwrap();

    expect.assert_eq(&actual);
}

#[test]
fn literals() {
    let script = r#"
            a = "Hello, World!"
            b = 12
            c = 18.5
            d = -1y2mo3w4d5h6m7s8ms9us10ns
            e = 2019-10-31T00:00:00Z
            f = /server[01]/
            "#;

    check(
        script,
        expect![[r#"
            package main
            a = "Hello, World!"
            b = 12:int
            c = 18.5:float
            d = -1y2mo3w4d5h6m7s8ms9us10ns:duration
            e = 2019-10-31T00:00:00Z:time
            f = /server[01]/:regexp"#]],
    )
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

    check(
        script,
        expect![[r#"
            package main
            a = [1:int, 2:int, 3:int]:[int]
            b = [1.1:float, 2.2:float, 3.3:float]:[float]
            c = ["1", "2", "3"]:[string]
            d = [1s:duration, 2m:duration, 3h:duration]:[duration]
            e = [2019-10-31T00:00:00Z:time]:[time]
            f = [/a/:regexp, /b/:regexp, /c/:regexp]:[regexp]
            g = [{a: 0:int, b: 0.0:float}, {a: 1:int, b: 1.1:float}]:[{a:int, b:float}]"#]],
    )
}

#[test]
fn dictionary_literals() {
    let script = r#"
            a = ["a": 0, "b": 1, "c": 2]
            b = [1970-01-01T00:00:00Z: 0, 1970-01-01T01:00:00Z: 1]
            "#;

    check(
        script,
        expect![[r#"
            package main
            a = ["a": 0:int, "b": 1:int, "c": 2:int]:[string:int]
            b = [1970-01-01T00:00:00Z:time: 0:int, 1970-01-01T01:00:00Z:time: 1:int]:[time:int]"#]],
    )
}

#[test]
fn identifier_expressions() {
    let script = r#"
            n = 1.0
            f = n + 3.0
            "#;

    check(
        script,
        expect![[r#"
            package main
            n = 1.0:float
            f = n:float +:float 3.0:float"#]],
    )
}

#[test]
fn function_default_arguments() {
    let script = r#"
            f = (a, b=1) => a + b
            x = f(a:2)
            y = f(a: x, b: f(a:x))
            "#;

    check(
        script,
        expect![[r#"
            package main
            f = (a, b=1:int) =>{
            return a:int +:int b:int
            }
            x = f:(a:int, ?b:int) => int(a: 2:int)
            y = f:(a:int, ?b:int) => int(a: x:int, b: f:(a:int, ?b:int) => int(a: x:int))"#]],
    )
}

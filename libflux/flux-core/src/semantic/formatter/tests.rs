use expect_test::{expect, Expect};

use crate::semantic::{
    formatter::format,
    import::Packages,
    types::{Function, MonoType, PolyTypeMap, SemanticMap, Tvar},
    Analyzer,
};

fn check(actual: &str, expect: Expect) {
    let mut analyzer = Analyzer::new_with_defaults(Default::default(), Packages::new());
    let (_, mut sem_pkg) = analyzer
        .analyze_source("main".to_string(), "main.flux".to_string(), actual)
        .unwrap_or_else(|err| panic!("{}", err));
    let actual = format(&sem_pkg).unwrap_or_else(|err| panic!("{}", err));

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
            b = 12
            c = 18.5
            d = -1y2mo3w4d5h6m7s8ms9us10ns:duration
            e = 2019-10-31T00:00:00Z
            f = /server[01]/"#]],
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
            a = [1, 2, 3]:[int]
            b = [1.1, 2.2, 3.3]:[float]
            c = ["1", "2", "3"]:[string]
            d = [1s, 2m, 3h]:[duration]
            e = [2019-10-31T00:00:00Z]:[time]
            f = [/a/, /b/, /c/]:[regexp]
            g = [{a: 0, b: 0.0}:{a:int, b:float}, {a: 1, b: 1.1}:{a:int, b:float}]:[{a:int, b:float}]"#]],
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
            a = ["a": 0, "b": 1, "c": 2]:[string:int]
            b = [1970-01-01T00:00:00Z: 0, 1970-01-01T01:00:00Z: 1]:[time:int]"#]],
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
            n = 1.0
            f = n:float +:float 3.0"#]],
    )
}

#[test]
fn format_function_expression() {
    let script = r#"
            (a) => a
            f = (a, b=1) => a + b
            x = f(a:2)
            y = f(a: x, b: f(a:x))
            g = (t=<-) => t
            "#;

    check(
        script,
        expect![[r#"
            package main
            (a) => {
                return a:A
            }:(a:A) => A
            f = (a, b=1) => {
                return a:C +:C b:C
            }:(a:C, ?b:C) => C
            x = f:(a:int, ?b:int) => int(a: 2):int
            y = f:(a:int, ?b:int) => int(a: x:int, b: f:(a:int, ?b:int) => int(a: x:int):int):int
            g = (t) => {
                return t:t12
            }:(<-t:t12) => t12"#]],
    )
}

#[test]
fn format_conditional_expression() {
    let script = r#"
            if 1 == 2 then 5 else 3
            ans = if 100 > 0 then "yes" else "no"
            "#;

    check(
        script,
        expect![[r#"
            package main
            (if 1 ==:bool 2 then 5 else 3):int
            ans = (if 100 >:bool 0 then "yes" else "no"):string"#]],
    )
}

#[test]
fn format_index_expression() {
    let script = r#"
            [1, 2, 3][1]
            "#;

    check(
        script,
        expect![[r#"
            package main
            [1, 2, 3]:[int][1]:int"#]],
    )
}

#[test]
fn format_unary_expression() {
    let script = r#"
            -1d
            x = -1
            y = +1
            "#;

    check(
        script,
        expect![[r#"
            package main
            -1d:duration
            x = -1:int
            y = +1:int"#]],
    )
}

#[test]
fn format_object_expression() {
    let script = r#"
            {a: 1, b: "2"}
            "#;

    check(
        script,
        expect![[r#"
            package main
            {a: 1, b: "2"}:{a:int, b:string}"#]],
    )
}

#[test]
fn format_member_expression() {
    let script = r#"
            o = {temp: 30.0, loc: "FL"}
            t = o.temp
            "#;

    check(
        script,
        expect![[r#"
            package main
            o = {temp: 30.0, loc: "FL"}:{temp:float, loc:string}
            t = o:{temp:float, loc:string}.temp:float"#]],
    )
}

#[test]
fn format_call_expression() {
    let script = r#"
            (() => 2)()
            "#;

    check(
        script,
        expect![[r#"
            package main
            (() => {
                return 2
            }:() => int)():int"#]],
    )
}

#[test]
fn format_option_statement() {
    let script = r#"
            option now = () => 2019-05-22T00:00:00Z
            "#;

    check(
        script,
        expect![[r#"
            package main
            option now = () => {
                return 2019-05-22T00:00:00Z
            }:() => time"#]],
    )
}

#[test]
fn format_test_statement() {
    let script = r#"
            test foo = {}
            "#;

    check(
        script,
        expect![[r#"
            package main
            test foo = {}:{}"#]],
    )
}

#[test]
fn format_block_statement() {
    let script = r#"
            (r) => {
                v = if r < 0 then -r else r
                return v * v
            }
            "#;

    check(
        script,
        expect![[r#"
            package main
            (r) => {
                v = (if r:A <:bool 0 then -r:A:A else r:A):A
                return v:A *:A v:A
            }:(r:A) => A"#]],
    )
}

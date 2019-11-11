//! This is the main test module for type inference.
//!
//! This module defines two macros:
//!
//! * `test_infer!`
//! * `test_infer_err!`
//!
//! `test_infer!` types a flux program and asserts that the inferred types
//! are equivalent to what was expected.
//!
//! `test_infer_err!` asserts that a flux program does not type check.
//!
//! Note macros are a great tool for emulating table driven tests. With
//! macros we can define a single pattern that each test case follows. We
//! may also define certain inputs as optional. This allows us the ability
//! to introduce new test inputs while maintaining backwards compatibility
//! with previously specified tests.
//!
//! With macros we can also provide custom test output that is much easier
//! to inspect in the case of a failed assertion than it is with
//! `assert_eq`, as the types retured from type inference can be
//! arbitrarily complex.
//!
use std::collections::HashMap;

use crate::semantic::analyze::analyze_with;
use crate::semantic::env::Environment;
use crate::semantic::fresh::Fresher;
use crate::semantic::nodes;
use crate::semantic::parser::parse;
use crate::semantic::sub::Substitutable;
use crate::semantic::types::{MaxTvar, MonoType, PolyType};

use crate::ast;
use crate::parser::parse_string;

use colored::*;

fn parse_program(src: &str) -> ast::Package {
    let file = parse_string("program", src);

    ast::Package {
        base: file.base.clone(),
        path: "path".to_string(),
        package: "main".to_string(),
        files: vec![file],
    }
}

fn validate(t: PolyType) -> Result<PolyType, String> {
    if let MonoType::Var(_) = t.expr {
        return Err(format!(
            "polymorphic values not allowed in type environment: {}",
            t
        ));
    }
    if !t.free_vars().is_empty() {
        return Err(format!(
            "free variables not allowed in type environment: {}",
            t
        ));
    }
    return Ok(t);
}

fn parse_env_map(env: HashMap<&str, &str>) -> HashMap<String, PolyType> {
    env.into_iter()
        .map(|(name, expr)| {
            let init = parse(expr).unwrap();
            let poly = validate(init).unwrap();
            return (name.to_string(), poly);
        })
        .collect()
}

fn infer_types(
    src: &str,
    env: HashMap<&str, &str>,
    want: Option<HashMap<&str, &str>>,
) -> Result<Environment, nodes::Error> {
    // Parse polytype expressions in initial environment.
    let env = parse_env_map(env);

    // Compute the maximum type variable in the environment
    // and initialize a fresher with this type variable.
    let max = env.max_tvar();
    let mut f = Fresher::from(max.0 + 1);

    let pkg = parse_program(src);

    let got = match nodes::infer_pkg_types(
        &mut analyze_with(pkg, &mut f).unwrap(),
        Environment::new(env.into()),
        &mut f,
    ) {
        Ok((env, _)) => env.values,
        Err(e) => return Err(e),
    };

    // Parse polytype expressions in expected environment.
    // Only perform this step if a map of wanted types exists.
    if let Some(env) = want {
        let want = parse_env_map(env);
        if want != got {
            panic!(
                "\n\n{}\n\n{}\n{}\n{}\n{}\n",
                "unexpected types:".red().bold(),
                "want:".green().bold(),
                want.iter().fold(String::new(), |acc, (name, poly)| acc
                    + &format!("\t{}: {}\n", name, poly)),
                "got:".red().bold(),
                got.iter().fold(String::new(), |acc, (name, poly)| acc
                    + &format!("\t{}: {}\n", name, poly)),
            );
        }
    }
    return Ok(got.into());
}

/// The test_infer! macro generates test cases for type inference.
///
/// A test case consists of:
///
/// 1. An optional type environment (representing a prelude)
/// 2. The flux program to be type checked
/// 3. The expected types of the top-level identifiers of said program
///
/// # Example
///
/// ```
/// #[test]
/// fn instantiation() {
///    test_infer! {
///         env: map![
///             "f" => "forall [t0] where t0: Addable (a: t0, b: t0) -> t0",
///         ],
///         src: "x = f",
///         exp: map![
///             "x" => "forall [t0] where t0: Addable (a: t0, b: t0) -> t0",
///         ],
///     }
/// }
/// ```
///
macro_rules! test_infer {
    ( env: $env:expr, src: $src:expr, exp: $exp:expr $(,)? ) => {{
        if let Err(e) = infer_types($src, $env, Some($exp)) {
            panic!(format!("{}", e));
        }
    }};
    ( src: $src:expr, exp: $exp:expr $(,)? ) => {{
        let env = HashMap::new();
        test_infer!(env: env, src: $src, exp: $exp);
    }};
}

/// The test_infer_err! macro generates test cases that don't type check.
///
/// These test cases consist of:
///
/// 1. An optional type environment (representing a prelude)
/// 2. A flux program that will not type check
///
/// # Example
///
/// ```
/// #[test]
/// fn undeclared_variable() {
///     test_infer_err! {
///         src: "x = f",
///     }
/// }
/// ```
///
macro_rules! test_infer_err {
    ( env: $env:expr, src: $src:expr $(,)? ) => {{
        if let Ok(env) = infer_types($src, $env, None) {
            panic!(
                "\n\n{}\n\n{}\n",
                "expected type error but instead inferred the following types:"
                    .red()
                    .bold(),
                env.values
                    .iter()
                    .fold(String::new(), |acc, (name, poly)| acc
                        + &format!("\t{}: {}\n", name, poly))
            )
        }
    }};
    ( src: $src:expr $(,)? ) => {{
        let env = HashMap::new();
        test_infer_err!(env: env, src: $src);
    }};
}

macro_rules! map {
    ($( $key: expr => $val: expr ),*$(,)?) => {{
         let mut map = ::std::collections::HashMap::new();
         $( map.insert($key, $val); )*
         map
    }}
}

#[test]
fn instantiation_0() {
    test_infer! {
        env: map![
            "f" => "forall [t0] where t0: Addable (a: t0, b: t0) -> t0",
        ],
        src: "x = f",
        exp: map![
            "x" => "forall [t0] where t0: Addable (a: t0, b: t0) -> t0",
        ],
    }
}
#[test]
fn instantiation_1() {
    test_infer! {
        env: map![
            "f" => "forall [t0] where t0: Addable (a: t0, b: t0) -> t0",
        ],
        src: r#"
            a = f
            x = a
        "#,
        exp: map![
            "a" => "forall [t0] where t0: Addable (a: t0, b: t0) -> t0",
            "x" => "forall [t0] where t0: Addable (a: t0, b: t0) -> t0",
        ],
    }
}
#[test]
fn literals() {
    test_infer! {
        src: r#"
            a = "Hello, World!"
            b = 12
            c = 18.5
            d = 12h
            e = 2019-10-31T00:00:00Z
            f = /server[01]/
        "#,
        exp: map![
            "a" => "forall [] string",
            "b" => "forall [] int",
            "c" => "forall [] float",
            "d" => "forall [] duration",
            "e" => "forall [] time",
            "f" => "forall [] regexp",
        ],
    }
}

// TODO(jsternberg): Use a function expression to test
// variables that are type variables.

#[test]
fn string_interpolation() {
    // test_infer! {
    //     env: map![
    //         "name" => "forall [] t0",
    //     ],
    //     src: src,
    //     exp: map![
    //         "name" => "forall [] string",
    //         "message" => "forall [] string",
    //     ],
    // }
    test_infer! {
        env: map![
            "name" => "forall [] string",
        ],
        src: r#"
            message = "Hello, ${name}!"
        "#,
        exp: map![
            "message" => "forall [] string",
        ],
    }
    test_infer_err! {
        env: map![
            "name" => "forall [] bool",
        ],
        src: r#"
            "Hello, ${name}!"
        "#,
    }
    test_infer_err! {
        env: map![
            "name" => "forall [] int",
        ],
        src: r#"
            "Hello, ${name}!"
        "#,
    }
    test_infer_err! {
        env: map![
            "name" => "forall [] uint",
        ],
        src: r#"
            "Hello, ${name}!"
        "#,
    }
    test_infer_err! {
        env: map![
            "name" => "forall [] float",
        ],
        src: r#"
            "Hello, ${name}!"
        "#,
    }
    test_infer_err! {
        env: map![
            "name" => "forall [] duration",
        ],
        src: r#"
            "Hello, ${name}!"
        "#,
    }
    test_infer_err! {
        env: map![
            "name" => "forall [] time",
        ],
        src: r#"
            "Hello, ${name}!"
        "#,
    }
    test_infer_err! {
        env: map![
            "name" => "forall [] regexp",
        ],
        src: r#"
            "Hello, ${name}!"
        "#,
    }
    test_infer_err! {
        env: map![
            "name" => "forall [] [int]",
        ],
        src: r#"
            "Hello, ${name}!"
        "#,
    }
    test_infer_err! {
        env: map![
            "name" => "forall [] {a: int | b: float}",
        ],
        src: r#"
            "Hello, ${name}!"
        "#,
    }
    test_infer_err! {
        env: map![
            "name" => "forall [t0] (x: t0) -> t0",
        ],
        src: r#"
            "Hello, ${name}!"
        "#,
    }
}
#[test]
fn array_lit() {
    test_infer! {
        src: "a = []",
        exp: map![
            "a" => "forall [t0] [t0]",
        ],
    }
    test_infer! {
        src: "a = [1, 2, 3]",
        exp: map![
            "a" => "forall [] [int]",
        ],
    }
    test_infer! {
        src: "a = [1.1, 2.2, 3.3]",
        exp: map![
            "a" => "forall [] [float]",
        ],
    }
    test_infer! {
        src: r#"
            a = ["1", "2", "3"]
        "#,
        exp: map![
            "a" => "forall [] [string]",
        ],
    }
    test_infer! {
        src: "a = [1s, 2m, 3h]",
        exp: map![
            "a" => "forall [] [duration]",
        ],
    }
    test_infer! {
        src: "a = [2019-10-31T00:00:00Z]",
        exp: map![
            "a" => "forall [] [time]",
        ],
    }
    test_infer! {
        src: "a = [/a/, /b/, /c/]",
        exp: map![
            "a" => "forall [] [regexp]",
        ],
    }
    test_infer! {
        src: "a = [{a:0, b:0.0}, {a:1, b:1.1}]",
        exp: map![
            "a" => "forall [] [{a: int | b: float}]",
        ],
    }
    test_infer_err! {
        src: "a = [1, 1.1]",
    }
    // test_infer! {
    //     env: map![
    //         "a" => "forall [] t0",
    //         "b" => "forall [] t1",
    //     ],
    //     src: r#"
    //            c = [a, b]
    //        "#,
    //     exp: map![
    //         "a" => "forall [] t3",
    //         "b" => "forall [] t3",
    //         "c" => "forall [] [t3]",
    //     ],
    // }
    // test_infer! {
    //     env: map![
    //         "a" => "forall [] t0",
    //         "b" => "forall [] int",
    //     ],
    //     src: r#"
    //            c = [a, b]
    //        "#,
    //     exp: map![
    //         "a" => "forall [] int",
    //         "c" => "forall [] [int]",
    //     ],
    // }
}
#[test]
fn array_expr() {
    let src = "b = [a]";

    test_infer! {
        env: map![
            "a" => "forall [] int",
        ],
        src: src,
        exp: map![
            "b" => "forall [] [int]",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] uint",
        ],
        src: src,
        exp: map![
            "b" => "forall [] [uint]",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] float",
        ],
        src: src,
        exp: map![
            "b" => "forall [] [float]",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] string",
        ],
        src: src,
        exp: map![
            "b" => "forall [] [string]",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] duration",
        ],
        src: src,
        exp: map![
            "b" => "forall [] [duration]",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] time",
        ],
        src: src,
        exp: map![
            "b" => "forall [] [time]",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] regexp",
        ],
        src: src,
        exp: map![
            "b" => "forall [] [regexp]",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] {a: int | b: float}",
        ],
        src: src,
        exp: map![
            "b" => "forall [] [{a: int | b: float}]",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] {a: string | b: (x: int) -> int}",
        ],
        src: src,
        exp: map![
            "b" => "forall [] [{a: string | b: (x: int) -> int}]",
        ],
    }
}
#[test]
fn binary_expr_addition() {
    // test_infer! {
    //     env: map![
    //         "a" => "forall [] t0",
    //         "b" => "forall [] t1",
    //     ],
    //     src: r#"
    //            c = a + b
    //        "#,
    //     exp: map![
    //         "a" => "forall [] t1",
    //         "b" => "forall [] t1",
    //         "c" => "forall [] t1",
    //     ],
    // }
    test_infer! {
        env: map![
            "a" => "forall [] int",
            "b" => "forall [] int",
        ],
        src: r#"
            c = a + b
        "#,
        exp: map![
            "c" => "forall [] int",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] uint",
            "b" => "forall [] uint",
        ],
        src: r#"
            c = a + b
        "#,
        exp: map![
            "c" => "forall [] uint",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] float",
            "b" => "forall [] float",
        ],
        src: r#"
            c = a + b
        "#,
        exp: map![
            "c" => "forall [] float",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] string",
            "b" => "forall [] string",
        ],
        src: r#"
            c = a + b
        "#,
        exp: map![
            "c" => "forall [] string",
        ],
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] bool",
            "b" => "forall [] bool",
        ],
        src: r#"
            a + b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] bool",
            "b" => "forall [] bool",
        ],
        src: r#"
            a + b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] duration",
            "b" => "forall [] duration",
        ],
        src: r#"
            a + b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] time",
            "b" => "forall [] time",
        ],
        src: r#"
            a + b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] regexp",
            "b" => "forall [] regexp",
        ],
        src: r#"
            a + b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] {a: int | b: float}",
            "b" => "forall [] {a: int | b: float}",
        ],
        src: r#"
            a + b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] [int]",
            "b" => "forall [] [int]",
        ],
        src: r#"
            a + b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] float",
            "b" => "forall [] int",
        ],
        src: r#"
            c = a + b
        "#,
    }
    // test_infer! {
    //     env: map![
    //         "a" => "forall [] t0",
    //         "b" => "forall [] int",
    //     ],
    //     src: r#"
    //            c = a + b
    //        "#,
    //     exp: map![
    //         "a" => "forall [] int",
    //         "c" => "forall [] int",
    //     ],
    // }
}
#[test]
fn binary_expr_subtraction() {
    // test_infer! {
    //     env: map![
    //         "a" => "forall [] t0",
    //         "b" => "forall [] t1",
    //     ],
    //     src: r#"
    //            c = a - b
    //        "#,
    //     exp: map![
    //         "a" => "forall [] t1",
    //         "b" => "forall [] t1",
    //         "c" => "forall [] t1",
    //     ],
    // }
    test_infer! {
        env: map![
            "a" => "forall [] int",
            "b" => "forall [] int",
        ],
        src: r#"
            c = a - b
        "#,
        exp: map![
            "c" => "forall [] int",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] float",
            "b" => "forall [] float",
        ],
        src: r#"
            c = a - b
        "#,
        exp: map![
            "c" => "forall [] float",
        ],
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] bool",
            "b" => "forall [] bool",
        ],
        src: r#"
            a - b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] uint",
            "b" => "forall [] uint",
        ],
        src: r#"
            a - b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] string",
            "b" => "forall [] string",
        ],
        src: r#"
            a - b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] duration",
            "b" => "forall [] duration",
        ],
        src: r#"
            a - b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] time",
            "b" => "forall [] time",
        ],
        src: r#"
            a - b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] regexp",
            "b" => "forall [] regexp",
        ],
        src: r#"
            a - b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] {a: int | b: float}",
            "b" => "forall [] {a: int | b: float}",
        ],
        src: r#"
            a - b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] [int]",
            "b" => "forall [] [int]",
        ],
        src: r#"
            a - b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] float",
            "b" => "forall [] int",
        ],
        src: r#"
            a - b
        "#,
    }
    // test_infer! {
    //     env: map![
    //         "a" => "forall [] t0",
    //         "b" => "forall [] int",
    //     ],
    //     src: r#"
    //            c = a - b
    //        "#,
    //     exp: map![
    //         "a" => "forall [] int",
    //         "c" => "forall [] int",
    //     ],
    // }
}
#[test]
fn binary_expr_multiplication() {
    // test_infer! {
    //     env: map![
    //         "a" => "forall [] t0",
    //         "b" => "forall [] t1",
    //     ],
    //     src: r#"
    //            c = a * b
    //        "#,
    //     exp: map![
    //         "a" => "forall [] t1",
    //         "b" => "forall [] t1",
    //         "c" => "forall [] t1",
    //     ],
    // }
    test_infer! {
        env: map![
            "a" => "forall [] int",
            "b" => "forall [] int",
        ],
        src: r#"
            c = a * b
        "#,
        exp: map![
            "c" => "forall [] int",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] uint",
            "b" => "forall [] uint",
        ],
        src: r#"
            c = a * b
        "#,
        exp: map![
            "c" => "forall [] uint",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] float",
            "b" => "forall [] float",
        ],
        src: r#"
            c = a * b
        "#,
        exp: map![
            "c" => "forall [] float",
        ],
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] bool",
            "b" => "forall [] bool",
        ],
        src: r#"
            a * b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] string",
            "b" => "forall [] string",
        ],
        src: r#"
            a * b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] duration",
            "b" => "forall [] duration",
        ],
        src: r#"
            a * b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] time",
            "b" => "forall [] time",
        ],
        src: r#"
            a * b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] regexp",
            "b" => "forall [] regexp",
        ],
        src: r#"
            a * b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] {a: int | b: float}",
            "b" => "forall [] {a: int | b: float}",
        ],
        src: r#"
            a * b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] [int]",
            "b" => "forall [] [int]",
        ],
        src: r#"
            a * b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] float",
            "b" => "forall [] int",
        ],
        src: r#"
            a * b
        "#,
    }
    // test_infer! {
    //     env: map![
    //         "a" => "forall [] t0",
    //         "b" => "forall [] int",
    //     ],
    //     src: r#"
    //            c = a * b
    //        "#,
    //     exp: map![
    //         "a" => "forall [] int",
    //         "c" => "forall [] int",
    //     ],
    // }
}
#[test]
fn binary_expr_division() {
    // test_infer! {
    //     env: map![
    //         "a" => "forall [] t0",
    //         "b" => "forall [] t1",
    //     ],
    //     src: r#"
    //            c = a / b
    //        "#,
    //     exp: map![
    //         "a" => "forall [] t1",
    //         "b" => "forall [] t1",
    //         "c" => "forall [] t1",
    //     ],
    // }
    test_infer! {
        env: map![
            "a" => "forall [] int",
            "b" => "forall [] int",
        ],
        src: r#"
            c = a / b
        "#,
        exp: map![
            "c" => "forall [] int",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] uint",
            "b" => "forall [] uint",
        ],
        src: r#"
            c = a / b
        "#,
        exp: map![
            "c" => "forall [] uint",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] float",
            "b" => "forall [] float",
        ],
        src: r#"
            c = a / b
        "#,
        exp: map![
            "c" => "forall [] float",
        ],
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] bool",
            "b" => "forall [] bool",
        ],
        src: r#"
            a / b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] string",
            "b" => "forall [] string",
        ],
        src: r#"
            a / b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] duration",
            "b" => "forall [] duration",
        ],
        src: r#"
            a / b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] time",
            "b" => "forall [] time",
        ],
        src: r#"
            a / b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] regexp",
            "b" => "forall [] regexp",
        ],
        src: r#"
            a / b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] {a: int | b: float}",
            "b" => "forall [] {a: int | b: float}",
        ],
        src: r#"
            a / b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] [int]",
            "b" => "forall [] [int]",
        ],
        src: r#"
            a / b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] float",
            "b" => "forall [] int",
        ],
        src: r#"
            a / b
        "#,
    }
    // test_infer! {
    //     env: map![
    //         "a" => "forall [] t0",
    //         "b" => "forall [] int",
    //     ],
    //     src: r#"
    //            c = a / b
    //        "#,
    //     exp: map![
    //         "a" => "forall [] int",
    //         "c" => "forall [] int",
    //     ],
    // }
}
#[test]
fn binary_expr_power() {
    // test_infer! {
    //     env: map![
    //         "a" => "forall [] t0",
    //         "b" => "forall [] t1",
    //     ],
    //     src: r#"
    //            c = a ^ b
    //        "#,
    //     exp: map![
    //         "a" => "forall [] t1",
    //         "b" => "forall [] t1",
    //         "c" => "forall [] t1",
    //     ],
    // }
    test_infer! {
        env: map![
            "a" => "forall [] int",
            "b" => "forall [] int",
        ],
        src: r#"
            c = a ^ b
        "#,
        exp: map![
            "c" => "forall [] int",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] uint",
            "b" => "forall [] uint",
        ],
        src: r#"
            c = a ^ b
        "#,
        exp: map![
            "c" => "forall [] uint",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] float",
            "b" => "forall [] float",
        ],
        src: r#"
            c = a ^ b
        "#,
        exp: map![
            "c" => "forall [] float",
        ],
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] bool",
            "b" => "forall [] bool",
        ],
        src: r#"
            a ^ b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] string",
            "b" => "forall [] string",
        ],
        src: r#"
            a ^ b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] duration",
            "b" => "forall [] duration",
        ],
        src: r#"
            a ^ b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] time",
            "b" => "forall [] time",
        ],
        src: r#"
            a ^ b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] regexp",
            "b" => "forall [] regexp",
        ],
        src: r#"
            a ^ b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] {a: int | b: float}",
            "b" => "forall [] {a: int | b: float}",
        ],
        src: r#"
            a ^ b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] [int]",
            "b" => "forall [] [int]",
        ],
        src: r#"
            a ^ b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] float",
            "b" => "forall [] int",
        ],
        src: r#"
            a ^ b
        "#,
    }
    // test_infer! {
    //     env: map![
    //         "a" => "forall [] t0",
    //         "b" => "forall [] int",
    //     ],
    //     src: r#"
    //            c = a ^ b
    //        "#,
    //     exp: map![
    //         "a" => "forall [] int",
    //         "c" => "forall [] int",
    //     ],
    // }
}
#[test]
fn binary_expr_modulo() {
    // test_infer! {
    //     env: map![
    //         "a" => "forall [] t0",
    //         "b" => "forall [] t1",
    //     ],
    //     src: r#"
    //            c = a % b
    //        "#,
    //     exp: map![
    //         "a" => "forall [] t1",
    //         "b" => "forall [] t1",
    //         "c" => "forall [] t1",
    //     ],
    // }
    test_infer! {
        env: map![
            "a" => "forall [] int",
            "b" => "forall [] int",
        ],
        src: r#"
            c = a % b
        "#,
        exp: map![
            "c" => "forall [] int",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] uint",
            "b" => "forall [] uint",
        ],
        src: r#"
            c = a % b
        "#,
        exp: map![
            "c" => "forall [] uint",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] float",
            "b" => "forall [] float",
        ],
        src: r#"
            c = a % b
        "#,
        exp: map![
            "c" => "forall [] float",
        ],
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] bool",
            "b" => "forall [] bool",
        ],
        src: r#"
            a % b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] string",
            "b" => "forall [] string",
        ],
        src: r#"
            a % b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] duration",
            "b" => "forall [] duration",
        ],
        src: r#"
            a % b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] time",
            "b" => "forall [] time",
        ],
        src: r#"
            a % b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] regexp",
            "b" => "forall [] regexp",
        ],
        src: r#"
            a % b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] {a: int | b: float}",
            "b" => "forall [] {a: int | b: float}",
        ],
        src: r#"
            a % b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] [int]",
            "b" => "forall [] [int]",
        ],
        src: r#"
            a % b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] float",
            "b" => "forall [] int",
        ],
        src: r#"
            a % b
        "#,
    }
    // test_infer! {
    //     env: map![
    //         "a" => "forall [] t0",
    //         "b" => "forall [] int",
    //     ],
    //     src: r#"
    //            c = a % b
    //        "#,
    //     exp: map![
    //         "a" => "forall [] int",
    //         "b" => "forall [] int",
    //         "c" => "forall [] int",
    //     ],
    // }
}
#[test]
fn binary_expr_comparison() {
    for op in vec!["==", "!="] {
        let src = format!("c = a {} b", op);

        test_infer! {
            env: map![
                "a" => "forall [] bool",
                "b" => "forall [] bool",
            ],
            src: &src,
            exp: map![
                "c" => "forall [] bool",
            ],
        }
        test_infer! {
            env: map![
                "a" => "forall [] int",
                "b" => "forall [] int",
            ],
            src: &src,
            exp: map![
                "c" => "forall [] bool",
            ],
        }
        test_infer! {
            env: map![
                "a" => "forall [] uint",
                "b" => "forall [] uint",
            ],
            src: &src,
            exp: map![
                "c" => "forall [] bool",
            ],
        }
        test_infer! {
            env: map![
                "a" => "forall [] float",
                "b" => "forall [] float",
            ],
            src: &src,
            exp: map![
                "c" => "forall [] bool",
            ],
        }
        test_infer! {
            env: map![
                "a" => "forall [] string",
                "b" => "forall [] string",
            ],
            src: &src,
            exp: map![
                "c" => "forall [] bool",
            ],
        }
        test_infer! {
            env: map![
                "a" => "forall [] duration",
                "b" => "forall [] duration",
            ],
            src: &src,
            exp: map![
                "c" => "forall [] bool",
            ],
        }
        test_infer! {
            env: map![
                "a" => "forall [] time",
                "b" => "forall [] time",
            ],
            src: &src,
            exp: map![
                "c" => "forall [] bool",
            ],
        }
        test_infer_err! {
            env: map![
                "a" => "forall [] regexp",
                "b" => "forall [] regexp",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "forall [] {a: int | b: float}",
                "b" => "forall [] {a: int | b: float}",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "forall [] [int]",
                "b" => "forall [] [int]",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "forall [] float",
                "b" => "forall [] int",
            ],
            src: &src,
        }
    }
    for op in vec![">=", "<=", ">", "<"] {
        let src = format!("c = a {} b", op);
        // test_infer! {
        //     env: map![
        //         "a" => "forall [] t0",
        //         "b" => "forall [] t1",
        //     ],
        //     src: &src,
        //     exp: map![
        //         "c" => "forall [] bool",
        //     ],
        // }
        test_infer! {
            env: map![
                "a" => "forall [] int",
                "b" => "forall [] int",
            ],
            src: &src,
            exp: map![
                "c" => "forall [] bool",
            ],
        }
        test_infer! {
            env: map![
                "a" => "forall [] uint",
                "b" => "forall [] uint",
            ],
            src: &src,
            exp: map![
                "c" => "forall [] bool",
            ],
        }
        test_infer! {
            env: map![
                "a" => "forall [] float",
                "b" => "forall [] float",
            ],
            src: &src,
            exp: map![
                "c" => "forall [] bool",
            ],
        }
        test_infer! {
            env: map![
                "a" => "forall [] string",
                "b" => "forall [] string",
            ],
            src: &src,
            exp: map![
                "c" => "forall [] bool",
            ],
        }
        test_infer! {
            env: map![
                "a" => "forall [] duration",
                "b" => "forall [] duration",
            ],
            src: &src,
            exp: map![
                "c" => "forall [] bool",
            ],
        }
        test_infer! {
            env: map![
                "a" => "forall [] time",
                "b" => "forall [] time",
            ],
            src: &src,
            exp: map![
                "c" => "forall [] bool",
            ],
        }
        test_infer_err! {
            env: map![
                "a" => "forall [] bool",
                "b" => "forall [] bool",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "forall [] regexp",
                "b" => "forall [] regexp",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "forall [] {a: int | b: float}",
                "b" => "forall [] {a: int | b: float}",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "forall [] [int]",
                "b" => "forall [] [int]",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "forall [] float",
                "b" => "forall [] int",
            ],
            src: &src,
        }
        // test_infer! {
        //     env: map![
        //         "a" => "forall [] t0",
        //         "b" => "forall [] int",
        //     ],
        //     src: &src,
        //     exp: map![
        //         "c" => "forall [] bool",
        //     ],
        // }
    }
}
#[test]
fn binary_expr_regex_op() {
    for op in vec!["=~", "!~"] {
        let src = format!("c = a {} b", op);
        // test_infer! {
        //     env: map![
        //         "a" => "forall [] t0",
        //         "b" => "forall [] t1",
        //     ],
        //     src: &src,
        //     exp: map![
        //         "a" => "forall [] string",
        //         "b" => "forall [] regexp",
        //         "c" => "forall [] bool",
        //     ],
        // }
        test_infer! {
            env: map![
                "a" => "forall [] string",
                "b" => "forall [] regexp",
            ],
            src: &src,
            exp: map![
                "c" => "forall [] bool",
            ],
        }
        test_infer_err! {
            env: map![
                "a" => "forall [] float",
                "b" => "forall [] regexp",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "forall [] string",
                "b" => "forall [] float",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "forall [] regexp",
                "b" => "forall [] string",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "forall [] float",
                "b" => "forall [] int",
            ],
            src: &src,
        }
        // test_infer! {
        //     env: map![
        //         "a" => "forall [] t0",
        //         "b" => "forall [] regexp",
        //     ],
        //     src: &src,
        //     exp: map![
        //         "a" => "forall [] string",
        //         "c" => "forall [] bool",
        //     ],
        // }
        // test_infer! {
        //     env: map![
        //         "a" => "forall [] string",
        //         "b" => "forall [] t0",
        //     ],
        //     src: &src,
        //     exp: map![
        //         "b" => "forall [] regexp",
        //         "c" => "forall [] bool",
        //     ],
        // }
    }
}
#[test]
fn conditional_expr() {
    test_infer! {
        env: map![
            "a" => "forall [] bool",
            "b" => "forall [] bool",
            "c" => "forall [] bool",
        ],
        src: r#"
            d = if a then b else c
        "#,
        exp: map![
            "d" => "forall [] bool",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] bool",
            "b" => "forall [] int",
            "c" => "forall [] int",
        ],
        src: r#"
            d = if a then b else c
        "#,
        exp: map![
            "d" => "forall [] int",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] bool",
            "b" => "forall [] uint",
            "c" => "forall [] uint",
        ],
        src: r#"
            d = if a then b else c
        "#,
        exp: map![
            "d" => "forall [] uint",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] bool",
            "b" => "forall [] float",
            "c" => "forall [] float",
        ],
        src: r#"
            d = if a then b else c
        "#,
        exp: map![
            "d" => "forall [] float",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] bool",
            "b" => "forall [] string",
            "c" => "forall [] string",
        ],
        src: r#"
            d = if a then b else c
        "#,
        exp: map![
            "d" => "forall [] string",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] bool",
            "b" => "forall [] duration",
            "c" => "forall [] duration",
        ],
        src: r#"
            d = if a then b else c
        "#,
        exp: map![
            "d" => "forall [] duration",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] bool",
            "b" => "forall [] time",
            "c" => "forall [] time",
        ],
        src: r#"
            d = if a then b else c
        "#,
        exp: map![
            "d" => "forall [] time",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] bool",
            "b" => "forall [] regexp",
            "c" => "forall [] regexp",
        ],
        src: r#"
            d = if a then b else c
        "#,
        exp: map![
            "d" => "forall [] regexp",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] bool",
            "b" => "forall [] [int]",
            "c" => "forall [] [int]",
        ],
        src: r#"
            d = if a then b else c
        "#,
        exp: map![
            "d" => "forall [] [int]",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] bool",
            "b" => "forall [] {a: int | b: regexp}",
            "c" => "forall [] {a: int | b: regexp}",
        ],
        src: r#"
            d = if a then b else c
        "#,
        exp: map![
            "d" => "forall [] {a: int | b: regexp}",
        ],
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] int",
            "b" => "forall [] int",
            "c" => "forall [] int",
        ],
        src: r#"
            d = if a then b else c
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] uint",
            "b" => "forall [] int",
            "c" => "forall [] int",
        ],
        src: r#"
            d = if a then b else c
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] float",
            "b" => "forall [] int",
            "c" => "forall [] int",
        ],
        src: r#"
            d = if a then b else c
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] string",
            "b" => "forall [] int",
            "c" => "forall [] int",
        ],
        src: r#"
            d = if a then b else c
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] duration",
            "b" => "forall [] int",
            "c" => "forall [] int",
        ],
        src: r#"
            d = if a then b else c
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] time",
            "b" => "forall [] int",
            "c" => "forall [] int",
        ],
        src: r#"
            d = if a then b else c
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] regexp",
            "b" => "forall [] int",
            "c" => "forall [] int",
        ],
        src: r#"
            d = if a then b else c
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] [int]",
            "b" => "forall [] int",
            "c" => "forall [] int",
        ],
        src: r#"
            d = if a then b else c
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] {a: int | b: [float]}",
            "b" => "forall [] int",
            "c" => "forall [] int",
        ],
        src: r#"
            d = if a then b else c
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] bool",
            "b" => "forall [] int",
            "c" => "forall [] float",
        ],
        src: r#"
            d = if a then b else c
        "#,
    }
    // test_infer! {
    //     env: map![
    //         "a" => "forall [] t0",
    //         "b" => "forall [] t1",
    //         "c" => "forall [] t2",
    //     ],
    //     src: r#"
    //            d = if a then b else c
    //        "#,
    //     exp: map![
    //         "a" => "forall [] bool",
    //         "b" => "forall [] t3",
    //         "c" => "forall [] t3",
    //         "d" => "forall [] t3",
    //     ],
    // }
    // test_infer! {
    //     env: map![
    //         "a" => "forall [] t0",
    //         "b" => "forall [] float",
    //         "c" => "forall [] float",
    //     ],
    //     src: r#"
    //            d = if a then b else c
    //        "#,
    //     exp: map![
    //         "a" => "forall [] bool",
    //         "d" => "forall [] float",
    //     ],
    // }
    // test_infer! {
    //     env: map![
    //         "a" => "forall [] t0",
    //         "b" => "forall [] t1",
    //         "c" => "forall [] float",
    //     ],
    //     src: r#"
    //            d = if a then b else c
    //        "#,
    //     exp: map![
    //         "a" => "forall [] bool",
    //         "b" => "forall [] float",
    //         "d" => "forall [] float",
    //     ],
    // }
}
#[test]
fn logical_expr() {
    for op in vec!["and", "or"] {
        let src = format!("c = a {} b", op);
        test_infer! {
            env: map![
                "a" => "forall [] bool",
                "b" => "forall [] bool",
            ],
            src: &src,
            exp: map![
                "c" => "forall [] bool",
            ],
        }
        test_infer_err! {
            env: map![
                "a" => "forall [] int",
                "b" => "forall [] int",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "forall [] uint",
                "b" => "forall [] uint",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "forall [] float",
                "b" => "forall [] float",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "forall [] string",
                "b" => "forall [] string",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "forall [] duration",
                "b" => "forall [] duration",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "forall [] time",
                "b" => "forall [] time",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "forall [] regexp",
                "b" => "forall [] regexp",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "forall [] [int]",
                "b" => "forall [] [int]",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "forall [] {a: bool}",
                "b" => "forall [] {a: bool}",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "forall [] int",
                "b" => "forall [] bool",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "forall [] bool",
                "b" => "forall [] int",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "forall [] int",
                "b" => "forall [] float",
            ],
            src: &src,
        }
        // test_infer! {
        //     env: map![
        //         "a" => "forall [] bool",
        //         "b" => "forall [] t1",
        //     ],
        //     src: &src,
        //     exp: map![
        //         "b" => "forall [] bool",
        //         "c" => "forall [] bool",
        //     ],
        // }
        // test_infer! {
        //     env: map![
        //         "a" => "forall [] t0",
        //         "b" => "forall [] bool",
        //     ],
        //     src: &src,
        //     exp: map![
        //         "a" => "forall [] bool",
        //         "c" => "forall [] bool",
        //     ],
        // }
    }
}
#[test]
fn index_expr() {
    let src = "c = a[b]";

    test_infer! {
        env: map![
            "a" => "forall [] [bool]",
            "b" => "forall [] int",
        ],
        src: src,
        exp: map![
            "c" => "forall [] bool",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] [int]",
            "b" => "forall [] int",
        ],
        src: src,
        exp: map![
            "c" => "forall [] int",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] [uint]",
            "b" => "forall [] int",
        ],
        src: src,
        exp: map![
            "c" => "forall [] uint",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] [float]",
            "b" => "forall [] int",
        ],
        src: src,
        exp: map![
            "c" => "forall [] float",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] [string]",
            "b" => "forall [] int",
        ],
        src: src,
        exp: map![
            "c" => "forall [] string",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] [duration]",
            "b" => "forall [] int",
        ],
        src: src,
        exp: map![
            "c" => "forall [] duration",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] [time]",
            "b" => "forall [] int",
        ],
        src: src,
        exp: map![
            "c" => "forall [] time",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] [regexp]",
            "b" => "forall [] int",
        ],
        src: src,
        exp: map![
            "c" => "forall [] regexp",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] [[int]]",
            "b" => "forall [] int",
        ],
        src: src,
        exp: map![
            "c" => "forall [] [int]",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] [{a: regexp}]",
            "b" => "forall [] int",
        ],
        src: src,
        exp: map![
            "c" => "forall [] {a: regexp}",
        ],
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] int",
            "b" => "forall [] int",
        ],
        src: src,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] uint",
            "b" => "forall [] int",
        ],
        src: src,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] float",
            "b" => "forall [] int",
        ],
        src: src,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] string",
            "b" => "forall [] int",
        ],
        src: src,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] duration",
            "b" => "forall [] int",
        ],
        src: src,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] time",
            "b" => "forall [] int",
        ],
        src: src,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] regexp",
            "b" => "forall [] int",
        ],
        src: src,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] {}",
            "b" => "forall [] int",
        ],
        src: src,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] [int]",
            "b" => "forall [] bool",
        ],
        src: src,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] [int]",
            "b" => "forall [] uint",
        ],
        src: src,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] [int]",
            "b" => "forall [] float",
        ],
        src: src,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] [int]",
            "b" => "forall [] string",
        ],
        src: src,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] [int]",
            "b" => "forall [] duration",
        ],
        src: src,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] [int]",
            "b" => "forall [] time",
        ],
        src: src,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] [int]",
            "b" => "forall [] regexp",
        ],
        src: src,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] [int]",
            "b" => "forall [] [int]",
        ],
        src: src,
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] [int]",
            "b" => "forall [] {}",
        ],
        src: src,
    }
    // test_infer! {
    //     env: map![
    //         "a" => "forall [] t0",
    //         "b" => "forall [] t1",
    //     ],
    //     src: src,
    //     exp: map![
    //         "a" => "forall [] [t2]",
    //         "b" => "forall [] int",
    //         "c" => "forall [] t2",
    //     ],
    // }
    // test_infer! {
    //     env: map![
    //         "a" => "forall [t0] [t0]",
    //         "b" => "forall [] t1",
    //     ],
    //     src: src,
    //     exp: map![
    //         "a" => "forall [t0] [t0]",
    //         "b" => "forall [] int",
    //         "c" => "forall [] t2",
    //     ],
    // }
    // test_infer! {
    //     env: map![
    //         "a" => "forall [] [int]",
    //         "b" => "forall [] t1",
    //     ],
    //     src: src,
    //     exp: map![
    //         "b" => "forall [] int",
    //         "c" => "forall [] int",
    //     ],
    // }
}
#[test]
fn unary_add() {
    test_infer! {
        env: map![
            "a" => "forall [] int",
        ],
        src: "b = +a",
        exp: map![
            "b" => "forall [] int",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] uint",
        ],
        src: "b = +a",
        exp: map![
            "b" => "forall [] uint",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] float",
        ],
        src: "b = +a",
        exp: map![
            "b" => "forall [] float",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] string",
        ],
        src: "b = +a",
        exp: map![
            "b" => "forall [] string",
        ],
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] bool",
        ],
        src: "+a",
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] duration",
        ],
        src: "+a",
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] time",
        ],
        src: "+a",
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] regexp",
        ],
        src: "+a",
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] [int]",
        ],
        src: "+a",
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] {}",
        ],
        src: "+a",
    }
}
#[test]
fn unary_sub() {
    test_infer! {
        env: map![
            "a" => "forall [] int",
        ],
        src: "b = -a",
        exp: map![
            "b" => "forall [] int",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] float",
        ],
        src: "b = -a",
        exp: map![
            "b" => "forall [] float",
        ],
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] bool",
        ],
        src: "-a",
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] uint",
        ],
        src: "-a",
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] string",
        ],
        src: "-a",
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] duration",
        ],
        src: "-a",
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] time",
        ],
        src: "-a",
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] regexp",
        ],
        src: "-a",
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] [int]",
        ],
        src: "-a",
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] {}",
        ],
        src: "-a",
    }
}
#[test]
fn exists() {
    test_infer! {
        env: map![
            "a" => "forall [] bool",
        ],
        src: "b = exists a",
        exp: map![
            "b" => "forall [] bool",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] int",
        ],
        src: "b = exists a",
        exp: map![
            "b" => "forall [] bool",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] uint",
        ],
        src: "b = exists a",
        exp: map![
            "b" => "forall [] bool",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] float",
        ],
        src: "b = exists a",
        exp: map![
            "b" => "forall [] bool",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] string",
        ],
        src: "b = exists a",
        exp: map![
            "b" => "forall [] bool",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] duration",
        ],
        src: "b = exists a",
        exp: map![
            "b" => "forall [] bool",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] time",
        ],
        src: "b = exists a",
        exp: map![
            "b" => "forall [] bool",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] regexp",
        ],
        src: "b = exists a",
        exp: map![
            "b" => "forall [] bool",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] [int]",
        ],
        src: "b = exists a",
        exp: map![
            "b" => "forall [] bool",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] {}",
        ],
        src: "b = exists a",
        exp: map![
            "b" => "forall [] bool",
        ],
    }
}
#[test]
fn logical_not() {
    test_infer! {
        env: map![
            "a" => "forall [] bool",
        ],
        src: "b = not a",
        exp: map![
            "b" => "forall [] bool",
        ],
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] int",
        ],
        src: "not a",
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] uint",
        ],
        src: "not a",
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] float",
        ],
        src: "not a",
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] string",
        ],
        src: "not a",
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] duration",
        ],
        src: "not a",
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] time",
        ],
        src: "not a",
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] regexp",
        ],
        src: "not a",
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] [int]",
        ],
        src: "not a",
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] {}",
        ],
        src: "not a",
    }
}
#[test]
fn undeclared_variable() {
    test_infer_err! {
        src: "x = f",
    }
}
#[test]
fn member_expression() {
    test_infer! {
        env: map![
            "r" => "forall [] {a: int | b: float | c: string}",
        ],
        src: r#"
            a = r.a
            b = r.b
            c = r.c
        "#,
        exp: map![
            "a" => "forall [] int",
            "b" => "forall [] float",
            "c" => "forall [] string",
        ],
    }
}
#[test]
fn non_existent_property() {
    test_infer_err! {
        env: map![
            "r" => "forall [] {a: int | b: float | c: string}",
        ],
        src: "r.d",
    }
}
#[test]
fn derived_record_literal() {
    test_infer! {
        env: map![
            "r" => "forall [] {a: int | b: float | c: string}",
        ],
        src: r#"
            o = {x: r.a, y: r.b, z: r.c}
        "#,
        exp: map![
            "o" => "forall [] {x: int | y: float | z: string}",
        ],
    }
}
#[test]
fn extend_record_literal() {
    test_infer! {
        env: map![
            "r" => "forall [] {a: int | b: float | c: string}",
        ],
        src: r#"
            o = {r with x: r.a}
        "#,
        exp: map![
            "o" => "forall [] {x: int | a: int | b: float | c: string}",
        ],
    }
}
#[test]
fn extend_generic_record() {
    test_infer! {
        env: map![
            "r" => "forall [t0] {a: int | b: float | t0}",
        ],
        src: r#"
            o = {r with x: r.a}
        "#,
        exp: map![
            "o" => "forall [t0] {x: int | a: int | b: float | t0}",
        ],
    }
}
#[test]
fn record_with_scoped_labels() {
    test_infer! {
        env: map![
            "r" => "forall [t0] {a: int | b: float | t0}",
            "x" => "forall [] int",
            "y" => "forall [] float",
        ],
        src: r#"
            u = {r with a: x}
            v = {r with a: y}
            w = {r with b: x}
        "#,
        exp: map![
            "u" => "forall [t0] {a: int   | a: int | b: float | t0}",
            "v" => "forall [t0] {a: float | a: int | b: float | t0}",
            "w" => "forall [t0] {b: int   | a: int | b: float | t0}",
        ],
    }
}

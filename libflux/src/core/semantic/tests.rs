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

use crate::semantic::bootstrap::build_polytype;
use crate::semantic::convert::convert_with;
use crate::semantic::env::Environment;
use crate::semantic::fresh::Fresher;
use crate::semantic::import::Importer;
use crate::semantic::nodes;
use crate::semantic::parser::parse;
use crate::semantic::sub::Substitutable;
use crate::semantic::types::{MaxTvar, MonoType, PolyType, PolyTypeMap, SemanticMap};

use crate::ast;
use crate::parser::parse_string;

use colored::*;

fn parse_program(src: &str) -> ast::Package {
    let file = parse_string("", src);

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

fn parse_map(m: HashMap<&str, &str>) -> PolyTypeMap {
    m.into_iter()
        .map(|(name, expr)| {
            let init = parse(expr).expect(format!("failed to parse {}", name).as_str());
            let poly = validate(init).expect(format!("failed to validate {}", name).as_str());
            return (name.to_string(), poly);
        })
        .collect()
}

impl Importer for HashMap<&str, PolyType> {
    fn import(&self, name: &str) -> Option<PolyType> {
        match self.get(name) {
            Some(pty) => Some(pty.clone()),
            None => None,
        }
    }
}

fn infer_types(
    src: &str,
    env: HashMap<&str, &str>,
    imp: HashMap<&str, HashMap<&str, &str>>,
    want: Option<HashMap<&str, &str>>,
) -> Result<Environment, nodes::Error> {
    // Parse polytype expressions in external packages.
    let imports: SemanticMap<&str, SemanticMap<String, PolyType>> = imp
        .into_iter()
        .map(|(path, pkg)| (path, parse_map(pkg)))
        .collect();

    // Compute the maximum type variable and init fresher
    let mut max = imports.max_tvar();
    let mut f = Fresher::from(max.0 + 1);

    // Instantiate package importer using generic objects
    let importer: HashMap<&str, PolyType> = imports
        .into_iter()
        .map(|(path, types)| (path, build_polytype(types, &mut f).unwrap()))
        .collect();

    for (_, t) in &importer {
        max = if t.max_tvar() > max {
            t.max_tvar()
        } else {
            max
        };
    }

    // Parse polytype expressions in initial environment.
    let env = parse_map(env);

    // Compute the maximum type variable and init fresher
    let max = if env.max_tvar() > max {
        env.max_tvar()
    } else {
        max
    };

    let mut f = Fresher::from(max.0 + 1);

    let pkg = parse_program(src);

    let got = match nodes::infer_pkg_types(
        &mut convert_with(pkg, &mut f).expect("analysis failed"),
        Environment::new(env.into()),
        &mut f,
        &importer,
        &None,
    ) {
        Ok((env, _)) => env.values,
        Err(e) => return Err(e),
    };

    // Parse polytype expressions in expected environment.
    // Only perform this step if a map of wanted types exists.
    if let Some(env) = want {
        let want = parse_map(env);
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
/// 2. Optional package imports (for any import statements)
/// 3. The flux program to be type checked
/// 4. The expected types of the top-level identifiers of said program
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
/// ```
/// #[test]
/// fn with_imports() {
///     test_infer! {
///         imp: map![
///             "path/to/foo" => package![
///                 "f" => "forall [t0] (x: t0) -> t0",
///             ],
///         ],
///         src: r#"
///             import foo "path/to/foo"
///
///             f = foo.f
///         "#,
///         exp: map![
///             "f" => "forall [t0] (x: t0) -> t0",
///         ],
///     }
/// }
/// ```
///
macro_rules! test_infer {
    ($(env: $env:expr,)? $(imp: $imp:expr,)? src: $src:expr, exp: $exp:expr $(,)? ) => {{
        #[allow(unused_mut, unused_assignments)]
        let mut env = HashMap::new();
        $(
            env = $env;
        )?
        #[allow(unused_mut, unused_assignments)]
        let mut imp = HashMap::new();
        $(
            imp = $imp;
        )?
        if let Err(e) = infer_types($src, env, imp, Some($exp)) {
            panic!(format!("{}", e));
        }
    }};
}

/// The test_infer_err! macro generates test cases that don't type check.
///
/// These test cases consist of:
///
/// 1. An optional type environment (representing a prelude)
/// 2. Optional package imports (for any import statements)
/// 3. A flux program that will not type check
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
    ( $(imp: $imp:expr,)? $(env: $env:expr,)? src: $src:expr $(,)? ) => {{
        #[allow(unused_mut, unused_assignments)]
        let mut imp = HashMap::new();
        $(
            imp = $imp;
        )?
        #[allow(unused_mut, unused_assignments)]
        let mut env = HashMap::new();
        $(
            env = $env;
        )?
        if let Ok(env) = infer_types($src, env, imp, None) {
            panic!(
                "\n\n{}\n\n{}\n",
                "expected type error but instead inferred the: following types:"
                    .red()
                    .bold(),
                env.values
                    .iter()
                    .fold(String::new(), |acc, (name, poly)| acc
                        + &format!("\t{}: {}\n", name, poly))
            )
        }
    }};
}

/// The test_error_msg! macro generates test cases for checking the error
/// messages produced by the type checker.
///
/// # Example
///
/// ```
/// #[test]
/// test_error_msg! {
///     src: r#"
///         1 + "1"
///     "#,
///     err: "type error @2:17-2:20: int != string",
/// }
/// ```
///
macro_rules! test_error_msg {
    ( src: $src:expr $(,)?, err: $err:expr $(,)? ) => {{
        match infer_types($src, HashMap::new(), HashMap::new(), None) {
            Err(e) => {
                if e.to_string() != $err {
                    panic!("\n\nexpected error:\n\t{}\n\ngot error:\n\t{}\n\n", $err, e)
                }
            }
            Ok(_) => panic!("expected error, instead program passed type checking"),
        }
    }};
}

macro_rules! map {
    ($( $key: expr => $val: expr ),*$(,)?) => {{
         let mut map = HashMap::new();
         $( map.insert($key, $val); )*
         map
    }}
}

macro_rules! package {
    ($( $key: expr => $val: expr ),*$(,)?) => {{
         let mut map = HashMap::new();
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
fn imports() {
    test_infer! {
        imp: map![
            "path/to/foo" => package![
                "a" => "forall [] int",
                "b" => "forall [] string",
            ],
            "path/to/bar" => package![
                "a" => "forall [] int",
                "b" => "forall [] {c: int | d: float}",
            ],
        ],
        src: r#"
            import foo "path/to/foo"
            import bar "path/to/bar"

            a = foo.a
            b = foo.b
            c = bar.a
            d = bar.b
        "#,
        exp: map![
            "a" => "forall [] int",
            "b" => "forall [] string",
            "c" => "forall [] int",
            "d" => "forall [] {c: int | d: float}",
        ],
    }
    test_infer! {
        imp: map![
            "path/to/foo" => package![
                "f" => "forall [t0] (x: t0) -> t0",
            ],
        ],
        src: r#"
            import foo "path/to/foo"

            f = foo.f
        "#,
        exp: map![
            "f" => "forall [t0] (x: t0) -> t0",
        ],
    }
    test_infer! {
        imp: map![
            "path/to/foo" => package![
                "f" => "forall [t0] (x: t0) -> t0",
            ],
        ],
        src: r#"
            import "path/to/foo"

            f = foo.f
        "#,
        exp: map![
            "f" => "forall [t0] (x: t0) -> t0",
        ],
    }
    test_infer! {
        imp: map![
            "path/to/foo" => package![
                "f" => "forall [t0] where t0: Addable + Divisible (x: t0) -> t0",
            ],
        ],
        src: r#"
            import foo "path/to/foo"

            f = foo.f
        "#,
        exp: map![
            "f" => "forall [t0] where t0: Addable + Divisible (x: t0) -> t0",
        ],
    }
    test_infer_err! {
        imp: map![
            "path/to/foo" => package![
                "a" => "forall [] bool",
                "b" => "forall [] time",
            ],
        ],
        src: r#"
            import foo "path/to/foo"

            foo.a + foo.b
        "#,
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
#[test]
fn string_interpolation() {
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
            "name" => "forall [] bytes",
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
        env: map![
            "bs" => "forall [] bytes",
        ],
        src: "a = [bs, bs, bs]",
        exp: map![
            "a" => "forall [] [bytes]",
        ],
    }
    test_infer! {
        env: map![
            "f" => "forall [] () -> bytes",
        ],
        src: "a = [f(), f(), f()]",
        exp: map![
            "a" => "forall [] [bytes]",
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
            "a" => "forall [] bytes",
        ],
        src: src,
        exp: map![
            "b" => "forall [] [bytes]",
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
}
#[test]
fn binary_expr_subtraction() {
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
    test_infer! {
        env: map![
            "a" => "forall [] uint",
            "b" => "forall [] uint",
        ],
        src: r#"
            c = a - b
        "#,
        exp: map![
            "c" => "forall [] uint",
        ],
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
}
#[test]
fn binary_expr_multiplication() {
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
}
#[test]
fn binary_expr_division() {
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
}
#[test]
fn binary_expr_power() {
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
}
#[test]
fn binary_expr_modulo() {
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
        test_infer! {
            env: map![
                "a" => "forall [] {a: int | b: float}",
                "b" => "forall [] {a: int | b: float}",
            ],
            src: &src,
            exp: map![
                "c" => "forall [] bool",
            ],
        }
        test_infer_err! {
            env: map![
                "a" => "forall [] {a: int | b: float | c: regexp}",
                "b" => "forall [] {a: int | b: float | c: regexp}",
            ],
            src: &src,
        }
        test_infer! {
            env: map![
                "a" => "forall [] [int]",
                "b" => "forall [] [int]",
            ],
            src: &src,
            exp: map![
                "c" => "forall [] bool",
            ],
        }
        test_infer_err! {
            env: map![
                "a" => "forall [] [regexp]",
                "b" => "forall [] [regexp]",
            ],
            src: &src,
        }
        // TODO(algow): re-introduce equality constraints for binary comparison operators
        // https://github.com/influxdata/flux/issues/2466
        test_infer! {
            env: map![
                "a" => "forall [] float",
                "b" => "forall [] int",
            ],
            src: &src,
            exp: map![
                "c" => "forall [] bool",
            ],
        }
    }
    for op in vec![">=", "<=", ">", "<"] {
        let src = format!("c = a {} b", op);

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
        // TODO(algow): re-introduce equality constraints for binary comparison operators
        // https://github.com/influxdata/flux/issues/2466
        test_infer! {
            env: map![
                "a" => "forall [] float",
                "b" => "forall [] int",
            ],
            src: &src,
            exp: map![
                "c" => "forall [] bool",
            ],
        }
    }
}
#[test]
fn binary_expr_regex_op() {
    for op in vec!["=~", "!~"] {
        let src = format!("c = a {} b", op);

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
            "a" => "forall [] float",
        ],
        src: "b = +a",
        exp: map![
            "b" => "forall [] float",
        ],
    }
    test_infer! {
        env: map![
            "a" => "forall [] duration",
        ],
        src: "b = +a",
        exp: map![
            "b" => "forall [] duration",
        ],
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] bool",
        ],
        src: "+a",
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
    test_infer_err! {
        env: map![
            "a" => "forall [] string",
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
    test_infer! {
        env: map![
            "a" => "forall [] duration",
        ],
        src: "b = -a",
        exp: map![
            "b" => "forall [] duration",
        ],
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] bool",
        ],
        src: "-a",
    }
    test_infer! {
        env: map![
            "a" => "forall [] uint",
        ],
        src: "b = -a",
        exp: map![
            "b" => "forall [] uint",
        ],
    }
    test_infer_err! {
        env: map![
            "a" => "forall [] string",
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

#[test]
fn pseudo_complete_query() {
    // TODO(algow): re-introduce equality constraints for binary comparison operators
    // https://github.com/influxdata/flux/issues/2466
    test_infer! {
        env: map![
            "from"   => "forall [t0, t1] (bucket: string) -> [{field: string | value: t1 | t0}]",
            "range"  => "forall [t0] (<-tables: [t0], start: duration) -> [t0]",
            "filter" => "forall [t0] (<-tables: [t0], fn: (r: t0) -> bool) -> [t0]",
            "map"    => "forall [t0,t1] (<-tables : [t0], fn: (r: t0) -> t1) -> [t1]",
            "int"    => "forall [t0] (v: t0) -> int",
        ],
        src: r#"
            out = from(bucket:"foo")
                |> range(start: 1d)
                |> filter(fn: (r) => r.host == "serverA" and r.measurement == "mem")
                |> map(fn: (r) => ({r with value: int(v: r.value)}))
        "#,
        exp: map![
            "out" => "forall [t0,t1,t2,t3] where t2: Equatable, t3: Equatable [{value: int | host: t2 | measurement: t3 | field: string | value: t1 | t0}]",
        ],
    }
}

#[test]
fn identity_function() {
    test_infer! {
        src: "f = (x) => x",
        exp: map![
            "f" => "forall [t0] (x: t0) -> t0",
        ],
    }
}

#[test]
fn call_expr() {
    // missing parameter
    test_infer_err! {
        src: r#"
            plusOne = (x) => x + 1.0
            plusOne()
        "#,
    }
    // missing pipe
    test_infer_err! {
        src: r#"
            add = (a=<-,b) => a + b
            add(b:2)
        "#,
    }
    // function does not take a pipe argument
    test_infer_err! {
        src: r#"
            f = () => 0
            g = () => 1 |> f()
        "#,
    }
    // function requires a pipe argument
    test_infer_err! {
        src: r#"
            f = (x=<-) => x
            g = () => f()
        "#,
    }
    test_infer! {
        src: r#"
            f = (x) => 0 |> x()
            f(x: (v=<-) => v)
            f(x: (w=<-) => w)
        "#,
        exp: map![
            "f" => "forall [t2] (x:(<-:int) -> t2) -> t2",
        ]
    }
    // pipe args have different names
    test_infer_err! {
        src: r#"
            f = (arg=(x=<-) => x, w) => w |> arg()
            f(arg: (v=<-) => v, w: 0)
        "#,
    }
    // Seems like it might fail because of pipe arg mismatch,
    // but it's okay.
    test_infer! {
        src: r#"
            f = (x, y) => x(arg: y)
            f(x: (arg=<-) => arg, y: 0)
        "#,
        exp: map![
            "f" => "forall [t2, t4] (x:(arg:t2) -> t4, y:t2) -> t4",
        ]
    }
    test_infer! {
        src: r#"
            f = (arg=(x=<-) => x) => 0 |> arg()
            g = () => f(arg: (x) => 5 + x)
        "#,
        exp: map![
            "f" => "forall [] (?arg:(<-x:int) -> int) -> int",
            "g" => "forall [] () -> int",
        ]
    }
}

#[test]
fn polymorphic_instantiation() {
    test_infer! {
        src: r#"
            f = (x) => x

            a = f(x: 100)
            b = f(x: 0.1)
            c = f(x: "0")
            d = f(x: 10h)
            e = f(x: 2019-10-31T00:00:00Z)
            g = f(x: /*/)
            h = f(x: [0])
            i = f(x: [])
            j = f(x: {a:0, b:0.1})
        "#,
        exp: map![
            "f" => "forall [t0] (x: t0) -> t0",

            "a" => "forall [] int",
            "b" => "forall [] float",
            "c" => "forall [] string",
            "d" => "forall [] duration",
            "e" => "forall [] time",
            "g" => "forall [] regexp",
            "h" => "forall [] [int]",
            "i" => "forall [t0] [t0]",
            "j" => "forall [] {a: int | b: float}",
        ],
    }
}
#[test]
fn constrain_tvars() {
    test_infer! {
        src: r#"
            f = (x) => x + 1
            a = f(x: 100)
        "#,
        exp: map![
            "f" => "forall [] (x: int) -> int",
            "a" => "forall [] int",
        ],
    }
    test_infer_err! {
        src: r#"
            f = (x) => x + 1
            f(x: "string")
        "#,
    }
    test_infer! {
        src: r#"
            f = (x) => "x = ${x}"
            a = f(x: "10")
        "#,
        exp: map![
            "f" => "forall [] (x: string) -> string",
            "a" => "forall [] string",
        ],
    }
    test_infer_err! {
        src: r#"
            f = (x) => "x = ${x}"
            f(x: 10)
        "#,
    }
}
#[test]
fn constrained_generics_addable() {
    test_infer! {
        src: r#"
            f = (a, b) => a + b
            a = f(a: 100, b: 200)
            b = f(a: 0.1, b: 0.2)
            c = f(a: "0", b: "1")
        "#,
        exp: map![
            "f" => "forall [t0] where t0: Addable (a: t0, b: t0) -> t0",
            "a" => "forall [] int",
            "b" => "forall [] float",
            "c" => "forall [] string",
        ],
    }
    test_infer_err! {
        src: r#"
            f = (a, b) => a + b
            f(a: 100, b: 0.1)
        "#,
    }
    test_infer_err! {
        src: r#"
            f = (a, b) => a + b
            f(a: 10d, b: 10d)
        "#,
    }
    test_infer_err! {
        src: r#"
            f = (a, b) => a + b
            f(a: 2019-10-31T00:00:00Z, b: 2019-10-31T00:00:00Z)
        "#,
    }
    test_infer_err! {
        src: r#"
            f = (a, b) => a + b
            f(a: /*/, b: /*/)
        "#,
    }
    test_infer_err! {
        src: r#"
            f = (a, b) => a + b
            f(a: [], b: [])
        "#,
    }
    test_infer_err! {
        src: r#"
            f = (a, b) => a + b
            f(a: [0], b: [1])
        "#,
    }
    test_infer_err! {
        src: r#"
            f = (a, b) => a + b
            f(a: {}, b: {})
        "#,
    }
}
#[test]
fn constrained_generics_subtractable() {
    test_infer! {
        src: r#"
            f = (a, b) => a - b
            a = f(a: 100, b: 200)
            b = f(a: 0.1, b: 0.2)
        "#,
        exp: map![
            "f" => "forall [t0] where t0: Subtractable (a: t0, b: t0) -> t0",
            "a" => "forall [] int",
            "b" => "forall [] float",
        ],
    }
    test_infer_err! {
        src: r#"
            f = (a, b) => a - b
            f(a: "string", b: "ing")
        "#,
    }
    test_infer_err! {
        src: r#"
            f = (a, b) => a - b
            f(a: 10d, b: 10d)
        "#,
    }
    test_infer_err! {
        src: r#"
            f = (a, b) => a - b
            f(a: 2019-10-31T00:00:00Z, b: 2019-10-31T00:00:00Z)
        "#,
    }
    test_infer_err! {
        src: r#"
            f = (a, b) => a - b
            f(a: /*/, b: /*/)
        "#,
    }
    test_infer_err! {
        src: r#"
            f = (a, b) => a - b
            f(a: [], b: [])
        "#,
    }
    test_infer_err! {
        src: r#"
            f = (a, b) => a - b
            f(a: {}, b: {})
        "#,
    }
}
#[test]
fn constrained_generics_divisible() {
    test_infer! {
        src: r#"
            f = (a, b) => a / b
            a = f(a: 100, b: 200)
            b = f(a: 0.1, b: 0.2)
        "#,
        exp: map![
            "f" => "forall [t0] where t0: Divisible (a: t0, b: t0) -> t0",
            "a" => "forall [] int",
            "b" => "forall [] float",
        ],
    }
    test_infer_err! {
        src: r#"
            f = (a, b) => a / b
            f(a: "string", b: "ing")
        "#,
    }
    test_infer_err! {
        src: r#"
            f = (a, b) => a / b
            f(a: 10d, b: 10d)
        "#,
    }
    test_infer_err! {
        src: r#"
            f = (a, b) => a / b
            f(a: 2019-10-31T00:00:00Z, b: 2019-10-31T00:00:00Z)
        "#,
    }
    test_infer_err! {
        src: r#"
            f = (a, b) => a / b
            f(a: /*/, b: /*/)
        "#,
    }
    test_infer_err! {
        src: r#"
            f = (a, b) => a / b
            f(a: [], b: [])
        "#,
    }
    test_infer_err! {
        src: r#"
            f = (a, b) => a / b
            f(a: {}, b: {})
        "#,
    }
}
#[test]
fn constrained_generics_comparable() {
    // TODO(algow): re-introduce equality constraints for binary comparison operators
    // https://github.com/influxdata/flux/issues/2466
    test_infer! {
        src: r#"
            f = (a, b) => a < b
            a = f(a: 100, b: 200)
            b = f(a: 0.1, b: 0.2)
            c = f(a: "0", b: "1")
            d = f(a: 10d, b: 20d)
            e = f(a: 2019-10-30T00:00:00Z, b: 2019-10-31T00:00:00Z)
        "#,
        exp: map![
            "f" => "forall [t0, t1] where t0: Comparable, t1: Comparable (a: t0, b: t1) -> bool",
            "a" => "forall [] bool",
            "b" => "forall [] bool",
            "c" => "forall [] bool",
            "d" => "forall [] bool",
            "e" => "forall [] bool",
        ],
    }
    test_infer_err! {
        env: map![
            "true" => "forall [] bool",
            "false" => "forall [] bool",
        ],
        src: r#"
            f = (a, b) => a < b
            f(a: true, b: false)
        "#,
    }
    test_infer_err! {
        src: r#"
            f = (a, b) => a < b
            f(a: /*/, b: /*/)
        "#,
    }
    test_infer_err! {
        src: r#"
            f = (a, b) => a < b
            f(a: [], b: [])
        "#,
    }
    test_infer_err! {
        src: r#"
            f = (a, b) => a < b
            f(a: {}, b: {})
        "#,
    }
}
#[test]
fn constrained_generics_equatable() {
    // TODO(algow): re-introduce equality constraints for binary comparison operators
    // https://github.com/influxdata/flux/issues/2466
    test_infer! {
        env: map![
            "true" => "forall [] bool",
            "false" => "forall [] bool",
        ],
        src: r#"
            f = (a, b) => a == b
            a = f(a: 100, b: 200)
            b = f(a: 0.1, b: 0.2)
            c = f(a: "0", b: "1")
            d = f(a: 10d, b: 20d)
            e = f(a: 2019-10-30T00:00:00Z, b: 2019-10-31T00:00:00Z)
            g = f(a: true, b: false)
        "#,
        exp: map![
            "f" => "forall [t0, t1] where t0: Equatable, t1: Equatable (a: t0, b: t1) -> bool",
            "a" => "forall [] bool",
            "b" => "forall [] bool",
            "c" => "forall [] bool",
            "d" => "forall [] bool",
            "e" => "forall [] bool",
            "g" => "forall [] bool",
        ],
    }
    test_infer_err! {
        src: r#"
            f = (a, b) => a < b
            f(a: /*/, b: /*/)
        "#,
    }
    test_infer_err! {
        src: r#"
            f = (a, b) => a < b
            f(a: [], b: [])
        "#,
    }
    test_infer_err! {
        src: r#"
            f = (a, b) => a < b
            f(a: {}, b: {})
        "#,
    }
}
#[test]
fn multiple_constraints() {
    // TODO(algow): re-introduce equality constraints for binary comparison operators
    // https://github.com/influxdata/flux/issues/2466
    test_infer! {
        src: r#"
            f = (a, b) => a <= b
            a = f(a: 100, b: 200)
            b = f(a: 0.1, b: 0.2)
            c = f(a: "0", b: "1")
            d = f(a: 10d, b: 20d)
            e = f(a: 2019-10-30T00:00:00Z, b: 2019-10-31T00:00:00Z)
        "#,
        exp: map![
            "f" => "forall [t0, t1] where t0: Comparable + Equatable, t1: Comparable + Equatable (a: t0, b: t1) -> bool",
            "a" => "forall [] bool",
            "b" => "forall [] bool",
            "c" => "forall [] bool",
            "d" => "forall [] bool",
            "e" => "forall [] bool",
        ],
    }
    test_infer_err! {
        env: map![
            "true" => "forall [] bool",
            "false" => "forall [] bool",
        ],
        src: r#"
            f = (a, b) => a < b
            f(a: true, b: false)
        "#,
    }
    test_infer_err! {
        src: r#"
            f = (a, b) => a < b
            f(a: /*/, b: /*/)
        "#,
    }
    test_infer_err! {
        src: r#"
            f = (a, b) => a < b
            f(a: [], b: [])
        "#,
    }
    test_infer_err! {
        src: r#"
            f = (a, b) => a < b
            f(a: {}, b: {})
        "#,
    }
}
#[test]
fn function_instantiation_and_generalization() {
    test_infer! {
        src: r#"
            r = ((o) => o)(o: 0)
            x = r
            s = ((o) => o)(o: {list: [0,1]})
            y = s.list
            t = ((o) => ({list: o}))(o: [0,1])
            z = t.list
            f = (x) => {
                y = x
                g = (a) => {
                    z = y
                    return z(b: a)
                }
                return g(a: [0])
            }
            a = f(x: (b) => b)
            b = f(x: (b) => b[0])
            c = f(x: (b) => ({b: b}))
        "#,
        exp: map![
            "r" => "forall [] int",
            "x" => "forall [] int",
            "s" => "forall [] {list: [int]}",
            "y" => "forall [] [int]",
            "t" => "forall [] {list: [int]}",
            "z" => "forall [] [int]",
            "f" => "forall [t0] (x: (b: [int]) -> t0) -> t0",
            "a" => "forall [] [int]",
            "b" => "forall [] int",
            "c" => "forall [] {b: [int]}",
        ],
    }
    test_infer_err! {
        src: r#"
            r = ((o) => o)(o: 0)
            x = r
            s = ((o) => o)(o: {list: [0,1]})
            y = s.list
            t = ((o) => ({list: o}))(o: [0,1])
            z = t.list
            f = (x) => {
                y = x
                g = (a) => {
                    z = y
                    return z(b: a)
                }
                return g(a: [0])
            }
            a = f(x: (b) => b)
            b = f(x: (b) => b[0])
            c = f(x: (b) => ({b: b}))
            d = f(x: (b) => 1 + b) // int != [int]
        "#,
    }
}
#[test]
fn function_default_arguments_1() {
    test_infer! {
        src: r#"
            f = (a, b=1) => a + b
            x = f(a:2)
            y = f(a: x, b: f(a:x))
        "#,
        exp: map![
            "f" => "forall [] (a: int, ?b: int) -> int",
            "x" => "forall [] int",
            "y" => "forall [] int",
        ],
    }
}
#[test]
fn function_default_arguments_2() {
    test_infer! {
        src: r#"
            f = (a, b, c=2.2, d=1) => ({r: a + c, s: b + d})
            w = f(a: 0.1, b: 4, c: 3.3, d: 3)
            x = f(a: 1.1, b: 2, c: 3.3)
            y = f(a: 2.2, b: 1, d: 3)
            z = f(a: 3.3, b: 3)
        "#,
        exp: map![
            "f" => "forall [] (a: float, b: int, ?c: float, ?d: int) -> {r: float | s: int}",
            "w" => "forall [] {r: float | s: int}",
            "x" => "forall [] {r: float | s: int}",
            "y" => "forall [] {r: float | s: int}",
            "z" => "forall [] {r: float | s: int}",
        ],
    }
}
#[test]
fn function_pipe_identity() {
    test_infer! {
        src: r#"
            f = (a=<-) => a
            x = f(a:2.2)
            y = 1 |> f()
        "#,
        exp: map![
            "f" => "forall [t0] (<-a: t0) -> t0",
            "x" => "forall [] float",
            "y" => "forall [] int",
        ],
    }
}
#[test]
fn function_default_arguments_and_pipes() {
    test_infer! {
        src: r#"
            f = (f, g, t=<-) => t |> f(a: g)
            x = (a, b=2, m=<-) => a + b + m
            z = (a, b=3.3, c=4.3, m=<-) => ({r: a.m, s: b + c + m})
            y = f(f: x, g: 100, t: 33)
            v = 2.2 |> f(f: z, g: {m: "4.5"})
        "#,
        exp: map![
            "f" => "forall [t0, t1, t2] (<-t: t1, f: (<-: t1, a: t0) -> t2, g: t0) -> t2",
            "x" => "forall [] (a: int, ?b: int, <-m: int) -> int",
            "z" => "forall [t0, t1] (a: {m: t0 | t1}, ?b: float, ?c: float, <-m: float) -> {r: t0 | s: float}",
            "y" => "forall [] int",
            "v" => "forall [] {r: string | s: float}",
        ],
    }
}
#[test]
fn test_error_messages() {
    test_error_msg! {
        src: r#"
            1 + "1"
        "#,
        // Location points to right expression expression
        err: "type error @2:17-2:20: expected int but found string",
    }
    test_error_msg! {
        src: r#"
            -"s"
        "#,
        // Location points to argument of unary expression
        err: "type error @2:14-2:17: string is not Negatable",
    }
    test_error_msg! {
        src: r#"
            1h + 2h
        "#,
        // Location points to entire binary expression
        err: "type error @2:13-2:20: duration is not Addable",
    }
    test_error_msg! {
        src: r#"
            bob = "Bob"
            joe = 0
            "Hey ${bob} it's me ${joe}!"
        "#,
        // Location points to second interpolated expression
        err: "type error @4:35-4:38: expected string but found int",
    }
    test_error_msg! {
        src: r#"
            if 0 then "a" else "b"
        "#,
        // Location points to if expression
        err: "type error @2:16-2:17: expected bool but found int",
    }
    test_error_msg! {
        src: r#"
            if exists 0 then 0 else "b"
        "#,
        // Location points to else expression
        err: "type error @2:37-2:40: expected int but found string",
    }
    test_error_msg! {
        src: r#"
            [1, "2"]
        "#,
        // Location points to second element of array
        err: "type error @2:17-2:20: expected int but found string",
    }
    test_error_msg! {
        src: r#"
            a = [1, 2, 3]
            a[1.1]
        "#,
        // Location points to expression representing the index
        err: "type error @3:15-3:18: expected int but found float",
    }
    test_error_msg! {
        src: r#"
            a = [1, 2, 3]
            a[1] + 1.1
        "#,
        // Location points to right expression
        err: "type error @3:20-3:23: expected int but found float",
    }
    test_error_msg! {
        src: r#"
            a = 1
            a[1]
        "#,
        // Location points to the identifier a
        err: "type error @3:13-3:14: expected [t2] but found int",
    }
    test_error_msg! {
        src: r#"
            a = [1, 2, 3]
            a.x
        "#,
        // Location points to the identifier a
        err: "type error @3:13-3:14: expected {x:t3 | t5} but found [int]",
    }
    test_error_msg! {
        src: r#"
            f = (x, y) => x - y
            f(x: "x", y: "y")
        "#,
        // Location points to entire call expression
        err: "type error @3:13-3:30: string is not Subtractable (argument x)",
    }
    test_error_msg! {
        src: r#"
            f = (r) => r.a
            f(r: {b: 1})
        "#,
        // Location points to entire call expression
        err: "type error @3:13-3:25: record is missing label a (argument r)",
    }
    test_error_msg! {
        src: r#"
            x = 1 + 1
            a
        "#,
        // Location points to the identifier a
        err: "error @3:13-3:14: undefined identifier a",
    }
    test_error_msg! {
        src: r#"
            match = (o) => o.name =~ /^a/
            fn = (r) => match(r)
        "#,
        // Location points to call expression `match(r)`
        err: "type error @3:25-3:33: found unexpected argument r",
    }
    test_error_msg! {
        src: r#"
            f = (a, b) => a + b
            f(a: 0, c: 1)
        "#,
        // Location points to call expression `f(a: 0, c: 1)`
        err: "type error @3:13-3:26: found unexpected argument c",
    }
    test_error_msg! {
        src: r#"
            f = (a, b) => a + b
            f(a: 0)
        "#,
        // Location points to call expression `f(a: 0)`
        err: "type error @3:13-3:20: missing required argument b",
    }
}

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

use crate::semantic::analyze::analyze;
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
///         env: &[
///             ("f", "forall [t0] where t0: Addable (a: t0, b: t0) -> t0"),
///         ],
///         src: "x = f",
///         exp: &[
///             ("x", "forall [t0] where t0: Addable (a: t0, b: t0) -> t0"),
///         ],
///     }
/// }
/// ```
///
macro_rules! test_infer {
    (
        env: $env:expr,
        src: $src:expr,
        exp: $exp:expr $(,)?
    ) => {{
        // Parse polytype expressions in initial environment
        let env: HashMap<String, PolyType> = $env
            .iter()
            .map(|(name, expr)| {
                let init = parse(expr).unwrap();
                let poly = validate(init).unwrap();
                return (name.to_string(), poly);
            })
            .collect();

        // Parse polytype expressions in expected environment
        let want: HashMap<String, PolyType> = $exp
            .iter()
            .map(|(name, expr)| {
                let init = parse(expr).unwrap();
                let poly = validate(init).unwrap();
                return (name.to_string(), poly);
            })
            .collect();

        // Compute the maximum type variable in the environment
        // and initialize a fresher with this type variable.
        let max = env.max_tvar();
        let mut f = Fresher::from(max.0 + 1);

        let pkg = parse_program($src);

        let got = match nodes::infer_pkg_types(
            &mut analyze(pkg, &mut f).unwrap(),
            Environment::new(env.into()),
            &mut f,
        ) {
            Err(err) => panic!(
                "\n\n{}: {}\n\n",
                "unexpected type error".red().bold(),
                err.to_string(),
            ),
            Ok((env, _)) => env.values,
        };

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
            )
        }
    }};
    ( src: $src:expr, exp: $exp:expr $(,)? ) => {{
        let env: Vec<(&str, &str)> = Vec::new();
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
    (
        env: $env:expr,
        src: $src:expr $(,)?
    ) => {{
        // Parse polytype expressions in initial environment
        let env: HashMap<String, PolyType> = $env
            .iter()
            .map(|(name, expr)| {
                let init = parse(expr).unwrap();
                let poly = validate(init).unwrap();
                return (name.to_string(), poly);
            })
            .collect();

        // Compute the maximum type variable in the environment
        // and initialize a fresher with this type variable.
        let max = env.max_tvar();
        let mut f = Fresher::from(max.0 + 1);

        let pkg = parse_program($src);

        if let Ok((env, _)) = nodes::infer_pkg_types(
            &mut analyze(pkg, &mut f).unwrap(),
            Environment::new(env.into()),
            &mut f,
        ) {
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
        };
    }};
    ( src: $src:expr $(,)? ) => {{
        let env: Vec<(&str, &str)> = Vec::new();
        test_infer_err!(env: env, src: $src);
    }};
}

#[test]
fn instantiation_0() {
    test_infer! {
        env: &[
            ("f", "forall [t0] where t0: Addable (a: t0, b: t0) -> t0"),
        ],
        src: "x = f",
        exp: &[
            ("x", "forall [t0] where t0: Addable (a: t0, b: t0) -> t0"),
        ],
    }
}
#[test]
fn instantiation_1() {
    test_infer! {
        env: &[
            ("f", "forall [t0] where t0: Addable (a: t0, b: t0) -> t0"),
        ],
        src: r#"
            a = f
            x = a
        "#,
        exp: &[
            ("a", "forall [t0] where t0: Addable (a: t0, b: t0) -> t0"),
            ("x", "forall [t0] where t0: Addable (a: t0, b: t0) -> t0"),
        ],
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
        env: &[
            ("r", "forall [] {a: int | b: float | c: string}"),
        ],
        src: r#"
            a = r.a
            b = r.b
            c = r.c
        "#,
        exp: &[
            ("a", "forall [] int"),
            ("b", "forall [] float"),
            ("c", "forall [] string"),
        ],
    }
}
#[test]
fn non_existent_property() {
    test_infer_err! {
        env: &[
            ("r", "forall [] {a: int | b: float | c: string}"),
        ],
        src: "r.d",
    }
}
#[test]
fn derived_record_literal() {
    test_infer! {
        env: &[
            ("r", "forall [] {a: int | b: float | c: string}"),
        ],
        src: r#"
            o = {x: r.a, y: r.b, z: r.c}
        "#,
        exp: &[
            ("o", "forall [] {x: int | y: float | z: string}")
        ],
    }
}
#[test]
fn extend_record_literal() {
    test_infer! {
        env: &[
            ("r", "forall [] {a: int | b: float | c: string}"),
        ],
        src: r#"
            o = {r with x: r.a}
        "#,
        exp: &[
            ("o", "forall [] {x: int | a: int | b: float | c: string}")
        ],
    }
}
#[test]
fn extend_generic_record() {
    test_infer! {
        env: &[
            ("r", "forall [t0] {a: int | b: float | t0}"),
        ],
        src: r#"
            o = {r with x: r.a}
        "#,
        exp: &[
            ("o", "forall [t0] {x: int | a: int | b: float | t0}")
        ],
    }
}
#[test]
fn record_with_scoped_labels() {
    test_infer! {
        env: &[
            ("r", "forall [t0] {a: int | b: float | t0}"),
            ("x", "forall [] int"),
            ("y", "forall [] float"),
        ],
        src: r#"
            u = {r with a: x}
            v = {r with a: y}
            w = {r with b: x}
        "#,
        exp: &[
            ("u", "forall [t0] {a: int   | a: int | b: float | t0}"),
            ("v", "forall [t0] {a: float | a: int | b: float | t0}"),
            ("w", "forall [t0] {b: int   | a: int | b: float | t0}"),
        ],
    }
}

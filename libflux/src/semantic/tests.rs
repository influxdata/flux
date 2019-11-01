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
use crate::semantic::types::PolyType;

use crate::ast;
use crate::parser::parse_string;

#[cfg(test)]
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
        $(env: $env:expr,)?
        src: $src:expr,
        exp: $exp:expr,
    ) => {
        let mut m = HashMap::new();
        let mut f = Fresher::new();

        $(
            for (name, expr) in $env {
                let poly = parse(expr).unwrap().normalize(&mut f);
                m.insert(name.to_string(), poly);
            }
        )?

        let pkg = parse_program($src);

        let types = match nodes::infer_pkg_types(
            &mut analyze(pkg, &mut f).unwrap(),
            Environment::new(m.into()),
            &mut f,
        ) {
            Err(err) => panic!("unexpected type error: {}", err.to_string()),
            Ok((e, _)) => e.values,
        };

        let got: HashMap<String, PolyType> = types
            .into_iter()
            .map(|(name, poly)| (name, poly.normalize(&mut Fresher::new())))
            .collect();

        let want: HashMap<String, PolyType> = $exp
            .iter()
            .map(|(name, expr)| {
                (
                    name.to_string(),
                    parse(expr).unwrap().normalize(&mut Fresher::new()),
                )
            })
            .collect();

        if want != got {
            panic!("\n\n{}\n\n{}\n{}\n{}\n{}\n",
                "unexpected types:".red().bold(),
                "want:".green().bold(),
                want.iter().fold(String::new(), |acc, (name, poly)| acc
                    + &format!("\t{}: {}\n", name, poly)),
                "got:".red().bold(),
                got.iter().fold(String::new(), |acc, (name, poly)| acc
                    + &format!("\t{}: {}\n", name, poly)),)
        }
    };
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
        $(env: $env:expr,)?
        src: $src:expr,
    ) => {
        let mut m = HashMap::new();
        let mut f = Fresher::new();

        $(
            for (name, expr) in $env {
                let poly = parse(expr).unwrap().normalize(&mut f);
                m.insert(name.to_string(), poly);
            }
        )?

        let pkg = parse_program($src);

        if let Ok((env, _)) = nodes::infer_pkg_types(
            &mut analyze(pkg, &mut f).unwrap(),
            Environment::new(m.into()),
            &mut f,
        ) {
            panic!("\n\n{}\n\n{}\n",
                "expected type error but instead inferred the following types:".red().bold(),
                env.values.iter().fold(String::new(), |acc, (name, poly)| acc
                    + &format!("\t{}: {}\n", name, poly)))
        };
    };
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

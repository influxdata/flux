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
use crate::semantic::types::{MaxTvar, MonoType, PolyType, PolyTypeMap, SemanticMap, TvarKinds};

use crate::ast;
use crate::ast::get_err_type_expression;
use crate::parser;
use crate::parser::parse_string;
use crate::semantic::convert::convert_polytype;

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

fn parse_map(m: HashMap<&str, &str>) -> PolyTypeMap {
    m.into_iter()
        .map(|(name, expr)| {
            let mut p = parser::Parser::new(expr);

            let typ_expr = p.parse_type_expression();
            let err = get_err_type_expression(typ_expr.clone());

            if err != "" {
                let msg = format!("TypeExpression parsing failed for {}. {:?}", name, err);
                panic!(msg)
            }
            let poly = convert_polytype(typ_expr, &mut Fresher::default());

            // let poly = parse(expr).expect(format!("failed to parse {}", name).as_str());
            return (name.to_string(), poly.unwrap());
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

    let mut env: Environment = env.into();
    env.readwrite = true;

    let mut f = Fresher::from(max.0 + 1);

    let pkg = parse_program(src);

    let got = match nodes::infer_pkg_types(
        &mut convert_with(pkg, &mut f).expect("analysis failed"),
        Environment::new(env),
        &mut f,
        &importer,
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
///             "f" => "where A: Addable (a: A, b: A) => A",
///         ],
///         src: "x = f",
///         exp: map![
///             "x" => "where A: Addable (a: A, b: A) => A",
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
///                 "f" => "(x: A) => A",
///             ],
///         ],
///         src: r#"
///             import foo "path/to/foo"
///
///             f = foo.f
///         "#,
///         exp: map![
///             "f" => "(x: A) => A",
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
            "f" => "(a: A, b: A) => A where A: Addable ",
        ],
        src: "x = f",
        exp: map![
            "x" => "(a: A, b: A) => A where A: Addable ",
        ],
    }
}
#[test]
fn instantiation_1() {
    test_infer! {
        env: map![
            "f" => "(a: A, b: A) => A where A: Addable ",
        ],
        src: r#"
            a = f
            x = a
        "#,
        exp: map![
            "a" => " (a: A, b: A) => A where A: Addable ",
            "x" => "(a: A, b: A) => A where A: Addable",
        ],
    }
}
#[test]
fn imports() {
    test_infer! {
        imp: map![
            "path/to/foo" => package![
                "a" => "int",
                "b" => "string",
            ],
            "path/to/bar" => package![
                "a" => "int",
                "b" => "{c: int , d: float}",
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
            "a" => "int",
            "b" => "string",
            "c" => "int",
            "d" => "{c: int , d: float}",
        ],
    }
    test_infer! {
        imp: map![
            "path/to/foo" => package![
                "f" => "(x: A) => A",
            ],
        ],
        src: r#"
            import foo "path/to/foo"

            f = foo.f
        "#,
        exp: map![
            "f" => "(x: A) => A",
        ],
    }
    test_infer! {
        imp: map![
            "path/to/foo" => package![
                "f" => "(x: A) => A",
            ],
        ],
        src: r#"
            import "path/to/foo"

            f = foo.f
        "#,
        exp: map![
            "f" => "(x: A) => A",
        ],
    }
    test_infer! {
        imp: map![
            "path/to/foo" => package![
                "f" => " (x: A) => A where A: Addable + Divisible",
            ],
        ],
        src: r#"
            import foo "path/to/foo"

            f = foo.f
        "#,
        exp: map![
            "f" => "(x: A) => A where A: Addable + Divisible ",
        ],
    }
    test_infer_err! {
        imp: map![
            "path/to/foo" => package![
                "a" => "bool",
                "b" => "time",
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
            "a" => "string",
            "b" => "int",
            "c" => "float",
            "d" => "duration",
            "e" => "time",
            "f" => "regexp",
        ],
    }
}
#[test]
fn string_interpolation() {
    test_infer! {
        env: map![
            "name" => "string",
        ],
        src: r#"
            message = "Hello, ${name}!"
        "#,
        exp: map![
            "message" => "string",
        ],
    }
    test_infer_err! {
        env: map![
            "name" => "bool",
        ],
        src: r#"
            "Hello, ${name}!"
        "#,
    }
    test_infer_err! {
        env: map![
            "name" => "int",
        ],
        src: r#"
            "Hello, ${name}!"
        "#,
    }
    test_infer_err! {
        env: map![
            "name" => "uint",
        ],
        src: r#"
            "Hello, ${name}!"
        "#,
    }
    test_infer_err! {
        env: map![
            "name" => "float",
        ],
        src: r#"
            "Hello, ${name}!"
        "#,
    }
    test_infer_err! {
        env: map![
            "name" => "duration",
        ],
        src: r#"
            "Hello, ${name}!"
        "#,
    }
    test_infer_err! {
        env: map![
            "name" => "time",
        ],
        src: r#"
            "Hello, ${name}!"
        "#,
    }
    test_infer_err! {
        env: map![
            "name" => "regexp",
        ],
        src: r#"
            "Hello, ${name}!"
        "#,
    }
    test_infer_err! {
        env: map![
            "name" => "bytes",
        ],
        src: r#"
            "Hello, ${name}!"
        "#,
    }
    test_infer_err! {
        env: map![
            "name" => "[int]",
        ],
        src: r#"
            "Hello, ${name}!"
        "#,
    }
    test_infer_err! {
        env: map![
            "name" => "{a: int , b: float}",
        ],
        src: r#"
            "Hello, ${name}!"
        "#,
    }
    test_infer_err! {
        env: map![
            "name" => "(x: A) => A",
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
            "a" => "[A]",
        ],
    }
    test_infer! {
        src: "a = [1, 2, 3]",
        exp: map![
            "a" => "[int]",
        ],
    }
    test_infer! {
        src: "a = [1.1, 2.2, 3.3]",
        exp: map![
            "a" => "[float]",
        ],
    }
    test_infer! {
        src: r#"
            a = ["1", "2", "3"]
        "#,
        exp: map![
            "a" => "[string]",
        ],
    }
    test_infer! {
        src: "a = [1s, 2m, 3h]",
        exp: map![
            "a" => "[duration]",
        ],
    }
    test_infer! {
        src: "a = [2019-10-31T00:00:00Z]",
        exp: map![
            "a" => "[time]",
        ],
    }
    test_infer! {
        src: "a = [/a/, /b/, /c/]",
        exp: map![
            "a" => "[regexp]",
        ],
    }
    test_infer! {
        env: map![
            "bs" => "bytes",
        ],
        src: "a = [bs, bs, bs]",
        exp: map![
            "a" => "[bytes]",
        ],
    }
    test_infer! {
        env: map![
            "f" => "() => bytes",
        ],
        src: "a = [f(), f(), f()]",
        exp: map![
            "a" => "[bytes]",
        ],
    }
    test_infer! {
        src: "a = [{a:0, b:0.0}, {a:1, b:1.1}]",
        exp: map![
            "a" => "[{a: int , b: float}]",
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
            "a" => "int",
        ],
        src: src,
        exp: map![
            "b" => "[int]",
        ],
    }
    test_infer! {
        env: map![
            "a" => "uint",
        ],
        src: src,
        exp: map![
            "b" => "[uint]",
        ],
    }
    test_infer! {
        env: map![
            "a" => "float",
        ],
        src: src,
        exp: map![
            "b" => "[float]",
        ],
    }
    test_infer! {
        env: map![
            "a" => "string",
        ],
        src: src,
        exp: map![
            "b" => "[string]",
        ],
    }
    test_infer! {
        env: map![
            "a" => "duration",
        ],
        src: src,
        exp: map![
            "b" => "[duration]",
        ],
    }
    test_infer! {
        env: map![
            "a" => "time",
        ],
        src: src,
        exp: map![
            "b" => "[time]",
        ],
    }
    test_infer! {
        env: map![
            "a" => "regexp",
        ],
        src: src,
        exp: map![
            "b" => "[regexp]",
        ],
    }
    test_infer! {
        env: map![
            "a" => "bytes",
        ],
        src: src,
        exp: map![
            "b" => "[bytes]",
        ],
    }
    test_infer! {
        env: map![
            "a" => "{a: int , b: float}",
        ],
        src: src,
        exp: map![
            "b" => "[{a: int , b: float}]",
        ],
    }
    test_infer! {
        env: map![
            "a" => "{a: string , b: (x: int) => int}",
        ],
        src: src,
        exp: map![
            "b" => "[{a: string , b: (x: int) => int}]",
        ],
    }
}
#[test]
fn binary_expr_addition() {
    test_infer! {
        env: map![
            "a" => "int",
            "b" => "int",
        ],
        src: r#"
            c = a + b
        "#,
        exp: map![
            "c" => "int",
        ],
    }
    test_infer! {
        env: map![
            "a" => "uint",
            "b" => "uint",
        ],
        src: r#"
            c = a + b
        "#,
        exp: map![
            "c" => "uint",
        ],
    }
    test_infer! {
        env: map![
            "a" => "float",
            "b" => "float",
        ],
        src: r#"
            c = a + b
        "#,
        exp: map![
            "c" => "float",
        ],
    }
    test_infer! {
        env: map![
            "a" => "string",
            "b" => "string",
        ],
        src: r#"
            c = a + b
        "#,
        exp: map![
            "c" => "string",
        ],
    }
    test_infer_err! {
        env: map![
            "a" => "bool",
            "b" => "bool",
        ],
        src: r#"
            a + b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "bool",
            "b" => "bool",
        ],
        src: r#"
            a + b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "duration",
            "b" => "duration",
        ],
        src: r#"
            a + b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "time",
            "b" => "time",
        ],
        src: r#"
            a + b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "regexp",
            "b" => "regexp",
        ],
        src: r#"
            a + b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "{a: int , b: float}",
            "b" => "{a: int , b: float}",
        ],
        src: r#"
            a + b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "[int]",
            "b" => "[int]",
        ],
        src: r#"
            a + b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "float",
            "b" => "int",
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
            "a" => "int",
            "b" => "int",
        ],
        src: r#"
            c = a - b
        "#,
        exp: map![
            "c" => "int",
        ],
    }
    test_infer! {
        env: map![
            "a" => "float",
            "b" => "float",
        ],
        src: r#"
            c = a - b
        "#,
        exp: map![
            "c" => "float",
        ],
    }
    test_infer_err! {
        env: map![
            "a" => "bool",
            "b" => "bool",
        ],
        src: r#"
            a - b
        "#,
    }
    test_infer! {
        env: map![
            "a" => "uint",
            "b" => "uint",
        ],
        src: r#"
            c = a - b
        "#,
        exp: map![
            "c" => "uint",
        ],
    }
    test_infer_err! {
        env: map![
            "a" => "string",
            "b" => "string",
        ],
        src: r#"
            a - b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "duration",
            "b" => "duration",
        ],
        src: r#"
            a - b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "time",
            "b" => "time",
        ],
        src: r#"
            a - b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "regexp",
            "b" => "regexp",
        ],
        src: r#"
            a - b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "{a: int , b: float}",
            "b" => "{a: int , b: float}",
        ],
        src: r#"
            a - b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "[int]",
            "b" => "[int]",
        ],
        src: r#"
            a - b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "float",
            "b" => "int",
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
            "a" => "int",
            "b" => "int",
        ],
        src: r#"
            c = a * b
        "#,
        exp: map![
            "c" => "int",
        ],
    }
    test_infer! {
        env: map![
            "a" => "uint",
            "b" => "uint",
        ],
        src: r#"
            c = a * b
        "#,
        exp: map![
            "c" => "uint",
        ],
    }
    test_infer! {
        env: map![
            "a" => "float",
            "b" => "float",
        ],
        src: r#"
            c = a * b
        "#,
        exp: map![
            "c" => "float",
        ],
    }
    test_infer_err! {
        env: map![
            "a" => "bool",
            "b" => "bool",
        ],
        src: r#"
            a * b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "string",
            "b" => "string",
        ],
        src: r#"
            a * b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "duration",
            "b" => "duration",
        ],
        src: r#"
            a * b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "time",
            "b" => "time",
        ],
        src: r#"
            a * b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "regexp",
            "b" => "regexp",
        ],
        src: r#"
            a * b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "{a: int , b: float}",
            "b" => "{a: int , b: float}",
        ],
        src: r#"
            a * b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "[int]",
            "b" => "[int]",
        ],
        src: r#"
            a * b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "float",
            "b" => "int",
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
            "a" => "int",
            "b" => "int",
        ],
        src: r#"
            c = a / b
        "#,
        exp: map![
            "c" => "int",
        ],
    }
    test_infer! {
        env: map![
            "a" => "uint",
            "b" => "uint",
        ],
        src: r#"
            c = a / b
        "#,
        exp: map![
            "c" => "uint",
        ],
    }
    test_infer! {
        env: map![
            "a" => "float",
            "b" => "float",
        ],
        src: r#"
            c = a / b
        "#,
        exp: map![
            "c" => "float",
        ],
    }
    test_infer_err! {
        env: map![
            "a" => "bool",
            "b" => "bool",
        ],
        src: r#"
            a / b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "string",
            "b" => "string",
        ],
        src: r#"
            a / b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "duration",
            "b" => "duration",
        ],
        src: r#"
            a / b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "time",
            "b" => "time",
        ],
        src: r#"
            a / b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "regexp",
            "b" => "regexp",
        ],
        src: r#"
            a / b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "{a: int , b: float}",
            "b" => "{a: int , b: float}",
        ],
        src: r#"
            a / b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "[int]",
            "b" => "[int]",
        ],
        src: r#"
            a / b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "float",
            "b" => "int",
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
            "a" => "int",
            "b" => "int",
        ],
        src: r#"
            c = a ^ b
        "#,
        exp: map![
            "c" => "int",
        ],
    }
    test_infer! {
        env: map![
            "a" => "uint",
            "b" => "uint",
        ],
        src: r#"
            c = a ^ b
        "#,
        exp: map![
            "c" => "uint",
        ],
    }
    test_infer! {
        env: map![
            "a" => "float",
            "b" => "float",
        ],
        src: r#"
            c = a ^ b
        "#,
        exp: map![
            "c" => "float",
        ],
    }
    test_infer_err! {
        env: map![
            "a" => "bool",
            "b" => "bool",
        ],
        src: r#"
            a ^ b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "string",
            "b" => "string",
        ],
        src: r#"
            a ^ b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "duration",
            "b" => "duration",
        ],
        src: r#"
            a ^ b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "time",
            "b" => "time",
        ],
        src: r#"
            a ^ b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "regexp",
            "b" => "regexp",
        ],
        src: r#"
            a ^ b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "{a: int , b: float}",
            "b" => "{a: int , b: float}",
        ],
        src: r#"
            a ^ b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "[int]",
            "b" => "[int]",
        ],
        src: r#"
            a ^ b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "float",
            "b" => "int",
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
            "a" => "int",
            "b" => "int",
        ],
        src: r#"
            c = a % b
        "#,
        exp: map![
            "c" => "int",
        ],
    }
    test_infer! {
        env: map![
            "a" => "uint",
            "b" => "uint",
        ],
        src: r#"
            c = a % b
        "#,
        exp: map![
            "c" => "uint",
        ],
    }
    test_infer! {
        env: map![
            "a" => "float",
            "b" => "float",
        ],
        src: r#"
            c = a % b
        "#,
        exp: map![
            "c" => "float",
        ],
    }
    test_infer_err! {
        env: map![
            "a" => "bool",
            "b" => "bool",
        ],
        src: r#"
            a % b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "string",
            "b" => "string",
        ],
        src: r#"
            a % b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "duration",
            "b" => "duration",
        ],
        src: r#"
            a % b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "time",
            "b" => "time",
        ],
        src: r#"
            a % b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "regexp",
            "b" => "regexp",
        ],
        src: r#"
            a % b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "{a: int , b: float}",
            "b" => "{a: int , b: float}",
        ],
        src: r#"
            a % b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "[int]",
            "b" => "[int]",
        ],
        src: r#"
            a % b
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "float",
            "b" => "int",
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
                "a" => "bool",
                "b" => "bool",
            ],
            src: &src,
            exp: map![
                "c" => "bool",
            ],
        }
        test_infer! {
            env: map![
                "a" => "int",
                "b" => "int",
            ],
            src: &src,
            exp: map![
                "c" => "bool",
            ],
        }
        test_infer! {
            env: map![
                "a" => "uint",
                "b" => "uint",
            ],
            src: &src,
            exp: map![
                "c" => "bool",
            ],
        }
        test_infer! {
            env: map![
                "a" => "float",
                "b" => "float",
            ],
            src: &src,
            exp: map![
                "c" => "bool",
            ],
        }
        test_infer! {
            env: map![
                "a" => "string",
                "b" => "string",
            ],
            src: &src,
            exp: map![
                "c" => "bool",
            ],
        }
        test_infer! {
            env: map![
                "a" => "duration",
                "b" => "duration",
            ],
            src: &src,
            exp: map![
                "c" => "bool",
            ],
        }
        test_infer! {
            env: map![
                "a" => "time",
                "b" => "time",
            ],
            src: &src,
            exp: map![
                "c" => "bool",
            ],
        }
        test_infer_err! {
            env: map![
                "a" => "regexp",
                "b" => "regexp",
            ],
            src: &src,
        }
        test_infer! {
            env: map![
                "a" => "{a: int , b: float}",
                "b" => "{a: int , b: float}",
            ],
            src: &src,
            exp: map![
                "c" => "bool",
            ],
        }
        test_infer_err! {
            env: map![
                "a" => "{a: int , b: float , c: regexp}",
                "b" => "{a: int , b: float , c: regexp}",
            ],
            src: &src,
        }
        test_infer! {
            env: map![
                "a" => "[int]",
                "b" => "[int]",
            ],
            src: &src,
            exp: map![
                "c" => "bool",
            ],
        }
        test_infer_err! {
            env: map![
                "a" => "[regexp]",
                "b" => "[regexp]",
            ],
            src: &src,
        }
        // TODO(algow): re-introduce equality constraints for binary comparison operators
        // https://github.com/influxdata/flux/issues/2466
        test_infer! {
            env: map![
                "a" => "float",
                "b" => "int",
            ],
            src: &src,
            exp: map![
                "c" => "bool",
            ],
        }
    }
    for op in vec![">=", "<=", ">", "<"] {
        let src = format!("c = a {} b", op);

        test_infer! {
            env: map![
                "a" => "int",
                "b" => "int",
            ],
            src: &src,
            exp: map![
                "c" => "bool",
            ],
        }
        test_infer! {
            env: map![
                "a" => "uint",
                "b" => "uint",
            ],
            src: &src,
            exp: map![
                "c" => "bool",
            ],
        }
        test_infer! {
            env: map![
                "a" => "float",
                "b" => "float",
            ],
            src: &src,
            exp: map![
                "c" => "bool",
            ],
        }
        test_infer! {
            env: map![
                "a" => "string",
                "b" => "string",
            ],
            src: &src,
            exp: map![
                "c" => "bool",
            ],
        }
        test_infer! {
            env: map![
                "a" => "duration",
                "b" => "duration",
            ],
            src: &src,
            exp: map![
                "c" => "bool",
            ],
        }
        test_infer! {
            env: map![
                "a" => "time",
                "b" => "time",
            ],
            src: &src,
            exp: map![
                "c" => "bool",
            ],
        }
        test_infer_err! {
            env: map![
                "a" => "bool",
                "b" => "bool",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "regexp",
                "b" => "regexp",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "{a: int , b: float}",
                "b" => "{a: int , b: float}",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "[int]",
                "b" => "[int]",
            ],
            src: &src,
        }
        // TODO(algow): re-introduce equality constraints for binary comparison operators
        // https://github.com/influxdata/flux/issues/2466
        test_infer! {
            env: map![
                "a" => "float",
                "b" => "int",
            ],
            src: &src,
            exp: map![
                "c" => "bool",
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
                "a" => "string",
                "b" => "regexp",
            ],
            src: &src,
            exp: map![
                "c" => "bool",
            ],
        }
        test_infer_err! {
            env: map![
                "a" => "float",
                "b" => "regexp",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "string",
                "b" => "float",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "regexp",
                "b" => "string",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "float",
                "b" => "int",
            ],
            src: &src,
        }
    }
}
#[test]
fn conditional_expr() {
    test_infer! {
        env: map![
            "a" => "bool",
            "b" => "bool",
            "c" => "bool",
        ],
        src: r#"
            d = if a then b else c
        "#,
        exp: map![
            "d" => "bool",
        ],
    }
    test_infer! {
        env: map![
            "a" => "bool",
            "b" => "int",
            "c" => "int",
        ],
        src: r#"
            d = if a then b else c
        "#,
        exp: map![
            "d" => "int",
        ],
    }
    test_infer! {
        env: map![
            "a" => "bool",
            "b" => "uint",
            "c" => "uint",
        ],
        src: r#"
            d = if a then b else c
        "#,
        exp: map![
            "d" => "uint",
        ],
    }
    test_infer! {
        env: map![
            "a" => "bool",
            "b" => "float",
            "c" => "float",
        ],
        src: r#"
            d = if a then b else c
        "#,
        exp: map![
            "d" => "float",
        ],
    }
    test_infer! {
        env: map![
            "a" => "bool",
            "b" => "string",
            "c" => "string",
        ],
        src: r#"
            d = if a then b else c
        "#,
        exp: map![
            "d" => "string",
        ],
    }
    test_infer! {
        env: map![
            "a" => "bool",
            "b" => "duration",
            "c" => "duration",
        ],
        src: r#"
            d = if a then b else c
        "#,
        exp: map![
            "d" => "duration",
        ],
    }
    test_infer! {
        env: map![
            "a" => "bool",
            "b" => "time",
            "c" => "time",
        ],
        src: r#"
            d = if a then b else c
        "#,
        exp: map![
            "d" => "time",
        ],
    }
    test_infer! {
        env: map![
            "a" => "bool",
            "b" => "regexp",
            "c" => "regexp",
        ],
        src: r#"
            d = if a then b else c
        "#,
        exp: map![
            "d" => "regexp",
        ],
    }
    test_infer! {
        env: map![
            "a" => "bool",
            "b" => "[int]",
            "c" => "[int]",
        ],
        src: r#"
            d = if a then b else c
        "#,
        exp: map![
            "d" => "[int]",
        ],
    }
    test_infer! {
        env: map![
            "a" => "bool",
            "b" => "{a: int , b: regexp}",
            "c" => "{a: int , b: regexp}",
        ],
        src: r#"
            d = if a then b else c
        "#,
        exp: map![
            "d" => "{a: int , b: regexp}",
        ],
    }
    test_infer_err! {
        env: map![
            "a" => "int",
            "b" => "int",
            "c" => "int",
        ],
        src: r#"
            d = if a then b else c
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "uint",
            "b" => "int",
            "c" => "int",
        ],
        src: r#"
            d = if a then b else c
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "float",
            "b" => "int",
            "c" => "int",
        ],
        src: r#"
            d = if a then b else c
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "string",
            "b" => "int",
            "c" => "int",
        ],
        src: r#"
            d = if a then b else c
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "duration",
            "b" => "int",
            "c" => "int",
        ],
        src: r#"
            d = if a then b else c
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "time",
            "b" => "int",
            "c" => "int",
        ],
        src: r#"
            d = if a then b else c
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "regexp",
            "b" => "int",
            "c" => "int",
        ],
        src: r#"
            d = if a then b else c
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "[int]",
            "b" => "int",
            "c" => "int",
        ],
        src: r#"
            d = if a then b else c
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "{a: int , b: [float]}",
            "b" => "int",
            "c" => "int",
        ],
        src: r#"
            d = if a then b else c
        "#,
    }
    test_infer_err! {
        env: map![
            "a" => "bool",
            "b" => "int",
            "c" => "float",
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
                "a" => "bool",
                "b" => "bool",
            ],
            src: &src,
            exp: map![
                "c" => "bool",
            ],
        }
        test_infer_err! {
            env: map![
                "a" => "int",
                "b" => "int",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "uint",
                "b" => "uint",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "float",
                "b" => "float",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "string",
                "b" => "string",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "duration",
                "b" => "duration",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "time",
                "b" => "time",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "regexp",
                "b" => "regexp",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "[int]",
                "b" => "[int]",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "{a: bool}",
                "b" => "{a: bool}",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "int",
                "b" => "bool",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "bool",
                "b" => "int",
            ],
            src: &src,
        }
        test_infer_err! {
            env: map![
                "a" => "int",
                "b" => "float",
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
            "a" => "[bool]",
            "b" => "int",
        ],
        src: src,
        exp: map![
            "c" => "bool",
        ],
    }
    test_infer! {
        env: map![
            "a" => "[int]",
            "b" => "int",
        ],
        src: src,
        exp: map![
            "c" => "int",
        ],
    }
    test_infer! {
        env: map![
            "a" => "[uint]",
            "b" => "int",
        ],
        src: src,
        exp: map![
            "c" => "uint",
        ],
    }
    test_infer! {
        env: map![
            "a" => "[float]",
            "b" => "int",
        ],
        src: src,
        exp: map![
            "c" => "float",
        ],
    }
    test_infer! {
        env: map![
            "a" => "[string]",
            "b" => "int",
        ],
        src: src,
        exp: map![
            "c" => "string",
        ],
    }
    test_infer! {
        env: map![
            "a" => "[duration]",
            "b" => "int",
        ],
        src: src,
        exp: map![
            "c" => "duration",
        ],
    }
    test_infer! {
        env: map![
            "a" => "[time]",
            "b" => "int",
        ],
        src: src,
        exp: map![
            "c" => "time",
        ],
    }
    test_infer! {
        env: map![
            "a" => "[regexp]",
            "b" => "int",
        ],
        src: src,
        exp: map![
            "c" => "regexp",
        ],
    }
    test_infer! {
        env: map![
            "a" => "[[int]]",
            "b" => "int",
        ],
        src: src,
        exp: map![
            "c" => "[int]",
        ],
    }
    test_infer! {
        env: map![
            "a" => "[{a: regexp}]",
            "b" => "int",
        ],
        src: src,
        exp: map![
            "c" => "{a: regexp}",
        ],
    }
    test_infer_err! {
        env: map![
            "a" => "int",
            "b" => "int",
        ],
        src: src,
    }
    test_infer_err! {
        env: map![
            "a" => "uint",
            "b" => "int",
        ],
        src: src,
    }
    test_infer_err! {
        env: map![
            "a" => "float",
            "b" => "int",
        ],
        src: src,
    }
    test_infer_err! {
        env: map![
            "a" => "string",
            "b" => "int",
        ],
        src: src,
    }
    test_infer_err! {
        env: map![
            "a" => "duration",
            "b" => "int",
        ],
        src: src,
    }
    test_infer_err! {
        env: map![
            "a" => "time",
            "b" => "int",
        ],
        src: src,
    }
    test_infer_err! {
        env: map![
            "a" => "regexp",
            "b" => "int",
        ],
        src: src,
    }
    test_infer_err! {
        env: map![
            "a" => "{}",
            "b" => "int",
        ],
        src: src,
    }
    test_infer_err! {
        env: map![
            "a" => "[int]",
            "b" => "bool",
        ],
        src: src,
    }
    test_infer_err! {
        env: map![
            "a" => "[int]",
            "b" => "uint",
        ],
        src: src,
    }
    test_infer_err! {
        env: map![
            "a" => "[int]",
            "b" => "float",
        ],
        src: src,
    }
    test_infer_err! {
        env: map![
            "a" => "[int]",
            "b" => "string",
        ],
        src: src,
    }
    test_infer_err! {
        env: map![
            "a" => "[int]",
            "b" => "duration",
        ],
        src: src,
    }
    test_infer_err! {
        env: map![
            "a" => "[int]",
            "b" => "time",
        ],
        src: src,
    }
    test_infer_err! {
        env: map![
            "a" => "[int]",
            "b" => "regexp",
        ],
        src: src,
    }
    test_infer_err! {
        env: map![
            "a" => "[int]",
            "b" => "[int]",
        ],
        src: src,
    }
    test_infer_err! {
        env: map![
            "a" => "[int]",
            "b" => "{}",
        ],
        src: src,
    }
}
#[test]
fn unary_add() {
    test_infer! {
        env: map![
            "a" => "int",
        ],
        src: "b = +a",
        exp: map![
            "b" => "int",
        ],
    }
    test_infer! {
        env: map![
            "a" => "float",
        ],
        src: "b = +a",
        exp: map![
            "b" => "float",
        ],
    }
    test_infer! {
        env: map![
            "a" => "duration",
        ],
        src: "b = +a",
        exp: map![
            "b" => "duration",
        ],
    }
    test_infer_err! {
        env: map![
            "a" => "bool",
        ],
        src: "+a",
    }
    test_infer! {
        env: map![
            "a" => "uint",
        ],
        src: "b = +a",
        exp: map![
            "b" => "uint",
        ],
    }
    test_infer_err! {
        env: map![
            "a" => "string",
        ],
        src: "+a",
    }
    test_infer_err! {
        env: map![
            "a" => "time",
        ],
        src: "+a",
    }
    test_infer_err! {
        env: map![
            "a" => "regexp",
        ],
        src: "+a",
    }
    test_infer_err! {
        env: map![
            "a" => "[int]",
        ],
        src: "+a",
    }
    test_infer_err! {
        env: map![
            "a" => "{}",
        ],
        src: "+a",
    }
}
#[test]
fn unary_sub() {
    test_infer! {
        env: map![
            "a" => "int",
        ],
        src: "b = -a",
        exp: map![
            "b" => "int",
        ],
    }
    test_infer! {
        env: map![
            "a" => "float",
        ],
        src: "b = -a",
        exp: map![
            "b" => "float",
        ],
    }
    test_infer! {
        env: map![
            "a" => "duration",
        ],
        src: "b = -a",
        exp: map![
            "b" => "duration",
        ],
    }
    test_infer_err! {
        env: map![
            "a" => "bool",
        ],
        src: "-a",
    }
    test_infer! {
        env: map![
            "a" => "uint",
        ],
        src: "b = -a",
        exp: map![
            "b" => "uint",
        ],
    }
    test_infer_err! {
        env: map![
            "a" => "string",
        ],
        src: "-a",
    }
    test_infer_err! {
        env: map![
            "a" => "time",
        ],
        src: "-a",
    }
    test_infer_err! {
        env: map![
            "a" => "regexp",
        ],
        src: "-a",
    }
    test_infer_err! {
        env: map![
            "a" => "[int]",
        ],
        src: "-a",
    }
    test_infer_err! {
        env: map![
            "a" => "{}",
        ],
        src: "-a",
    }
}
#[test]
fn exists() {
    test_infer! {
        env: map![
            "a" => "bool",
        ],
        src: "b = exists a",
        exp: map![
            "b" => "bool",
        ],
    }
    test_infer! {
        env: map![
            "a" => "int",
        ],
        src: "b = exists a",
        exp: map![
            "b" => "bool",
        ],
    }
    test_infer! {
        env: map![
            "a" => "uint",
        ],
        src: "b = exists a",
        exp: map![
            "b" => "bool",
        ],
    }
    test_infer! {
        env: map![
            "a" => "float",
        ],
        src: "b = exists a",
        exp: map![
            "b" => "bool",
        ],
    }
    test_infer! {
        env: map![
            "a" => "string",
        ],
        src: "b = exists a",
        exp: map![
            "b" => "bool",
        ],
    }
    test_infer! {
        env: map![
            "a" => "duration",
        ],
        src: "b = exists a",
        exp: map![
            "b" => "bool",
        ],
    }
    test_infer! {
        env: map![
            "a" => "time",
        ],
        src: "b = exists a",
        exp: map![
            "b" => "bool",
        ],
    }
    test_infer! {
        env: map![
            "a" => "regexp",
        ],
        src: "b = exists a",
        exp: map![
            "b" => "bool",
        ],
    }
    test_infer! {
        env: map![
            "a" => "[int]",
        ],
        src: "b = exists a",
        exp: map![
            "b" => "bool",
        ],
    }
    test_infer! {
        env: map![
            "a" => "{}",
        ],
        src: "b = exists a",
        exp: map![
            "b" => "bool",
        ],
    }
}
#[test]
fn logical_not() {
    test_infer! {
        env: map![
            "a" => "bool",
        ],
        src: "b = not a",
        exp: map![
            "b" => "bool",
        ],
    }
    test_infer_err! {
        env: map![
            "a" => "int",
        ],
        src: "not a",
    }
    test_infer_err! {
        env: map![
            "a" => "uint",
        ],
        src: "not a",
    }
    test_infer_err! {
        env: map![
            "a" => "float",
        ],
        src: "not a",
    }
    test_infer_err! {
        env: map![
            "a" => "string",
        ],
        src: "not a",
    }
    test_infer_err! {
        env: map![
            "a" => "duration",
        ],
        src: "not a",
    }
    test_infer_err! {
        env: map![
            "a" => "time",
        ],
        src: "not a",
    }
    test_infer_err! {
        env: map![
            "a" => "regexp",
        ],
        src: "not a",
    }
    test_infer_err! {
        env: map![
            "a" => "[int]",
        ],
        src: "not a",
    }
    test_infer_err! {
        env: map![
            "a" => "{}",
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
            "r" => "{a: int , b: float , c: string}",
        ],
        src: r#"
            a = r.a
            b = r.b
            c = r.c
        "#,
        exp: map![
            "a" => "int",
            "b" => "float",
            "c" => "string",
        ],
    }
}
#[test]
fn non_existent_property() {
    test_infer_err! {
        env: map![
            "r" => "{a: int , b: float , c: string}",
        ],
        src: "r.d",
    }
}
#[test]
fn derived_record_literal() {
    test_infer! {
        env: map![
            "r" => "{a: int , b: float , c: string}",
        ],
        src: r#"
            o = {x: r.a, y: r.b, z: r.c}
        "#,
        exp: map![
            "o" => "{x: int , y: float , z: string}",
        ],
    }
}
#[test]
fn extend_record_literal() {
    test_infer! {
        env: map![
            "r" => "{a: int , b: float , c: string}",
        ],
        src: r#"
            o = {r with x: r.a}
        "#,
        exp: map![
            "o" => "{x: int , a: int , b: float , c: string}",
        ],
    }
}
#[test]
fn extend_generic_record() {
    test_infer! {
        env: map![
            "r" => "{A with a: int , b: float}",
        ],
        src: r#"
            o = {r with x: r.a}
        "#,
        exp: map![
            "o" => "{A with x: int , a: int , b: float }",
        ],
    }
}
#[test]
fn record_with_scoped_labels() {
    test_infer! {
        env: map![
            "r" => "{A with a: int , b: float }",
            "x" => "int",
            "y" => "float",
        ],
        src: r#"
            u = {r with a: x}
            v = {r with a: y}
            w = {r with b: x}
        "#,
        exp: map![
            "u" => "{A with a: int   , a: int , b: float }",
            "v" => "{A with b: float , a: int , a: float }",
            "w" => "{A with b: float   , a: int , b: int }",
        ],
    }
}

#[test]
fn pseudo_complete_query() {
    // TODO(algow): re-introduce equality constraints for binary comparison operators
    // https://github.com/influxdata/flux/issues/2466
    test_infer! {
        env: map![
            "from"   => "(bucket: string) => [{A with field: string , value: B }]",
            "range"  => "(<-tables: [A], start: duration) => [A]",
            "filter" => "(<-tables: [A], fn: (r: A) => bool) => [A]",
            "map"    => "(<-tables : [A], fn: (r: A) => B) => [B]",
            "int"    => "(v: A) => int",
        ],
        src: r#"
            out = from(bucket:"foo")
                |> range(start: 1d)
                |> filter(fn: (r) => r.host == "serverA" and r.measurement == "mem")
                |> map(fn: (r) => ({r with value: int(v: r.value)}))

        "#,
        exp: map![
            "out" => "[{A with field: string,  value: B, value: int,  host: C, measurement: D  }] where C: Equatable, D: Equatable ",
        ],
    }
}

#[test]
fn identity_function() {
    test_infer! {
        src: "f = (x) => x",
        exp: map![
            "f" => "(x: A) => A",
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
            "f" => "(x:(<-:int) => C) => C",
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
            "f" => "(x:(arg:C) => E, y:C) => E",
        ]
    }
    test_infer! {
        src: r#"
            f = (arg=(x=<-) => x) => 0 |> arg()
            g = () => f(arg: (x) => 5 + x)
        "#,
        exp: map![
            "f" => "(?arg:(<-x:int) => int) => int",
            "g" => "() => int",
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
            "f" => "(x: A) => A",

            "a" => "int",
            "b" => "float",
            "c" => "string",
            "d" => "duration",
            "e" => "time",
            "g" => "regexp",
            "h" => "[int]",
            "i" => "[A]",
            "j" => "{a: int , b: float}",
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
            "f" => "(x: int) => int",
            "a" => "int",
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
            "f" => "(x: string) => string",
            "a" => "string",
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
            "f" => "(a: A, b: A) => A where A: Addable ",
            "a" => "int",
            "b" => "float",
            "c" => "string",
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
            "f" => "(a: A, b: A) => A where A: Subtractable ",
            "a" => "int",
            "b" => "float",
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
            "f" => "(a: A, b: A) => A where A: Divisible ",
            "a" => "int",
            "b" => "float",
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
            "f" => "(a: A, b: B) => bool where A: Comparable, B: Comparable ",
            "a" => "bool",
            "b" => "bool",
            "c" => "bool",
            "d" => "bool",
            "e" => "bool",
        ],
    }
    test_infer_err! {
        env: map![
            "true" => "bool",
            "false" => "bool",
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
            "true" => "bool",
            "false" => "bool",
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
            "f" => "(a: A, b: B) => bool where A: Equatable, B: Equatable ",
            "a" => "bool",
            "b" => "bool",
            "c" => "bool",
            "d" => "bool",
            "e" => "bool",
            "g" => "bool",
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
            "f" => "(a: A, b: B) => bool where A: Comparable + Equatable, B: Comparable + Equatable ",
            "a" => "bool",
            "b" => "bool",
            "c" => "bool",
            "d" => "bool",
            "e" => "bool",
        ],
    }
    test_infer_err! {
        env: map![
            "true" => "bool",
            "false" => "bool",
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
fn constrained_generics_timeable() {
    test_infer! {
        env: map![
            "a" => "(t: A) => A where A: Timeable ",
            "b" => "time",
            "c" => "duration",
        ],
        src: r#"
            d = a(t: b)
            e = a(t: c)
        "#,
        exp: map![
            "d" => "time",
            "e" => "duration",
        ],
    }

    test_infer_err! {
        env: map![
            "a" => "(t: A) => A where A: Timeable ",
            "b" => "string",
        ],
        src: r#"
            c = a(t: b)
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
            "r" => "int",
            "x" => "int",
            "s" => "{list: [int]}",
            "y" => "[int]",
            "t" => "{list: [int]}",
            "z" => "[int]",
            "f" => "(x: (b: [int]) => A) => A",
            "a" => "[int]",
            "b" => "int",
            "c" => "{b: [int]}",
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
            "f" => "(a: int, ?b: int) => int",
            "x" => "int",
            "y" => "int",
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
            "f" => "(a: float, b: int, ?c: float, ?d: int) => {r: float , s: int}",
            "w" => "{r: float , s: int}",
            "x" => "{r: float , s: int}",
            "y" => "{r: float , s: int}",
            "z" => "{r: float , s: int}",
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
            "f" => "(<-a: A) => A",
            "x" => "float",
            "y" => "int",
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
            "f" => "(<-t: B, f: (<-: B, a: A) => C, g: A) => C",
            "x" => "(a: int, ?b: int, <-m: int) => int",
            "z" => "(a: {B with m: A}, ?b: float, ?c: float, <-m: float) => {r: A , s: float}",
            "y" => "int",
            "v" => "{s: float, r: string}",
        ],
    }
}

#[test]
fn copy_bindings_from_other_env() {
    let mut env = Environment::empty(true);
    let mut f = Fresher::default();
    env.add(
        "a".to_string(),
        PolyType {
            vars: Vec::new(),
            cons: TvarKinds::new(),
            expr: MonoType::Bool,
        },
    );
    let mut sub_env = Environment::new(env.clone());
    sub_env.add(
        "b".to_string(),
        PolyType {
            vars: Vec::new(),
            cons: TvarKinds::new(),
            expr: MonoType::Var(f.fresh()),
        },
    );
    sub_env.copy_bindings_from(&env);
    assert_eq!(
        sub_env,
        Environment {
            parent: Some(env.clone().into()),
            readwrite: true,
            values: semantic_map!(
                "b".to_string() => PolyType {
                    vars: Vec::new(),
                    cons: TvarKinds::new(),
                    expr: MonoType::Var(f.fresh()),
                },
                "a".to_string() => PolyType {
                    vars: Vec::new(),
                    cons: TvarKinds::new(),
                    expr: MonoType::Bool,
                }
            )
        }
    );
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

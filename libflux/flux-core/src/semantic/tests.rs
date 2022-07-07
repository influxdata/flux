//! This is th main test module for type inference.
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

use colored::*;
use derive_more::Display;
use expect_test::expect;

use crate::{
    ast,
    map::HashMap,
    parser,
    semantic::{
        self,
        convert::convert_polytype,
        env::Environment,
        fresh::Fresher,
        import::Packages,
        nodes::Symbol,
        types::{BoundTvarKinds, MonoType, PolyType, PolyTypeHashMap, SemanticMap},
        Analyzer, AnalyzerConfig, Feature, PackageExports,
    },
};

mod vectorize;

fn parse_map(package: Option<&str>, m: HashMap<&str, &str>) -> PolyTypeHashMap<Symbol> {
    m.into_iter()
        .map(|(name, expr)| {
            let mut p = parser::Parser::new(expr);

            let typ_expr = p.parse_type_expression();

            if let Err(err) = ast::check::check(ast::walk::Node::TypeExpression(&typ_expr)) {
                panic!("TypeExpression parsing failed for {}. {}", name, err);
            }
            let poly = convert_polytype(&typ_expr, &Default::default())
                .unwrap_or_else(|err| panic!("{}", err));

            (
                match package {
                    None => Symbol::from(name),
                    Some(package) => Symbol::from(name).with_package(package),
                },
                poly,
            )
        })
        .collect()
}

#[derive(Debug, Display, PartialEq)]
enum Error {
    #[display(fmt = "{}", _0)]
    Semantic(semantic::FileErrors),
    #[display(
        fmt = "\n\n{}\n\n{}\n{}\n{}\n{}\n",
        r#""unexpected types:".red().bold()"#,
        r#""want:".green().bold()"#,
        r#"want.iter().fold(String::new(), |acc, (name, poly)| acc
            + &format!("\t{}: {}\n", name, poly))"#,
        r#""got:".red().bold()"#,
        r#"got.iter().fold(String::new(), |acc, (name, poly)| acc
                    + &format!("\t{}: {}\n", name, poly))"#
    )]
    TypeMismatch {
        want: SemanticMap<String, PolyType>,
        got: SemanticMap<String, PolyType>,
    },
}

impl Error {
    fn pretty(&self, source: &str) -> String {
        match self {
            Self::Semantic(err) => err.pretty(source),
            _ => self.to_string(),
        }
    }
    fn pretty_short(&self, source: &str) -> String {
        match self {
            Self::Semantic(err) => err.pretty_short(source),
            _ => self.to_string(),
        }
    }
}

impl std::error::Error for Error {}

fn infer_types(
    src: &str,
    env: HashMap<&str, &str>,
    imp: HashMap<&str, HashMap<&str, &str>>,
    want: Option<HashMap<&str, &str>>,
    config: AnalyzerConfig,
) -> Result<(PackageExports, semantic::nodes::Package), Error> {
    let _ = env_logger::try_init();
    // Parse polytype expressions in external packages.
    let imports: SemanticMap<&str, _> = imp
        .into_iter()
        .map(|(path, pkg)| (path, parse_map(Some(path), pkg)))
        .collect();

    // Instantiate package importer using generic objects
    let importer: Packages = imports
        .into_iter()
        .map(|(path, types)| (path.to_string(), PackageExports::try_from(types).unwrap()))
        .collect();

    // Parse polytype expressions in initial environment.
    let env = parse_map(None, env);

    let env = Environment::from(env);

    let mut analyzer = Analyzer::new(Environment::new(env), importer, config);
    let (env, pkg) = analyzer
        .analyze_source("main".into(), "".into(), src)
        .map_err(|err| Error::Semantic(err.error))?;

    // Parse polytype expressions in expected environment.
    // Only perform this step if a map of wanted types exists.
    if let Some(want_env) = want {
        let got = env
            .clone()
            .into_bindings()
            .map(|(k, v)| (k.to_string(), v))
            .collect();
        let want = parse_map(Some("main"), want_env)
            .into_iter_by(|l, r| l.name().cmp(r.name()))
            .map(|(k, v)| (k.to_string(), v))
            .collect();
        if want != got {
            return Err(Error::TypeMismatch { want, got });
        }
    }
    return Ok((env, pkg));
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
    ($(config: $config:expr,)? $(env: $env:expr,)? $(imp: $imp:expr,)? src: $src:expr, exp: $exp:expr $(,)? ) => {{
        #[allow(unused_mut, unused_assignments)]
        let mut env = HashMap::default();
        $(
            env = $env;
        )?
        #[allow(unused_mut, unused_assignments)]
        let mut imp = HashMap::default();
        $(
            imp = $imp;
        )?
        #[allow(unused_mut, unused_assignments)]
        let mut config = AnalyzerConfig::default();
        $(
            config = $config;
        )?
        if let Err(e) = infer_types($src, env, imp, Some($exp), config) {
            eprintln!("{:#?}", e);
            panic!("{}", e.pretty($src));
        }
    }}
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
        let mut imp = HashMap::default();
        $(
            imp = $imp;
        )?
        #[allow(unused_mut, unused_assignments)]
        let mut env = HashMap::default();
        $(
            env = $env;
        )?
        match infer_types($src, env, imp, None, AnalyzerConfig::default()) {
            Ok((env, _)) => {
                panic!(
                    "\n\n{}\n\n{}\n",
                    "expected type error but instead inferred the: following types:"
                        .red()
                        .bold(),
                    env.iter()
                        .fold(String::new(), |acc, (name, poly)| acc
                            + &format!("\t{}: {}\n", name, poly))
                )
            }
            Err(err @ Error::TypeMismatch {.. }) => {
                panic!("{}", err)
            }
            Err(Error::Semantic(error)) => {
                for err in error.diagnostics.errors {
                    if let semantic::ErrorKind::InvalidAST(_) = err.error {
                        panic!("{}", err);
                    }
                }
            }
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
    ( $(config: $config:expr,)?  $(test: $test: ident,)? $(imp: $imp:expr,)? $(env: $env:expr,)? src: $src:expr $(,)?, expect: $expect:expr $(,)? ) => {
        $(#[test] fn $test() )? {

        #[allow(unused_mut, unused_assignments)]
        let mut imp = HashMap::default();
        $(
            imp = $imp;
        )?
        #[allow(unused_mut, unused_assignments)]
        let mut env = HashMap::default();
        $(
            env = $env;
        )?
        #[allow(unused_mut, unused_assignments)]
        let mut config = AnalyzerConfig::default();
        $(
            config = $config;
        )?
        match infer_types(
            $src,
            env,
            imp,
            None,
            config,
        ) {
            Err(e) => {
                let got = e.pretty($src);
                $expect.assert_eq(&got);
            }
            Ok(_) => panic!("expected error, instead program passed type checking"),
        }
    }};

    ( $(test: $test: ident,)? $(imp: $imp:expr,)? $(env: $env:expr,)? src: $src:expr $(,)?, expect_short: $expect:expr $(,)? ) => {
        $(#[test] fn $test() )? {

        #[allow(unused_mut, unused_assignments)]
        let mut imp = HashMap::default();
        $(
            imp = $imp;
        )?
        #[allow(unused_mut, unused_assignments)]
        let mut env = HashMap::default();
        $(
            env = $env;
        )?
        match infer_types(
            $src,
            env,
            imp,
            None,
            AnalyzerConfig::default(),
        ) {
            Err(e) => {
                let got = e.pretty_short($src);
                $expect.assert_eq(&got);
            }
            Ok(_) => panic!("expected error, instead program passed type checking"),
        }
    }};

    ( $(test: $test: ident,)? $(imp: $imp:expr,)? $(env: $env:expr,)? src: $src:expr $(,)?, err: $err:expr $(,)? ) => {
        $(#[test] fn $test() )? {

        #[allow(unused_mut, unused_assignments)]
        let mut imp = HashMap::default();
        $(
            imp = $imp;
        )?
        #[allow(unused_mut, unused_assignments)]
        let mut env = HashMap::default();
        $(
            env = $env;
        )?
        match infer_types(
            $src,
            env,
            imp,
            None,
            AnalyzerConfig::default(),
        ) {
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
         let mut map = HashMap::default();
         $( map.insert($key, $val); )*
         map
    }}
}

macro_rules! package {
    ($( $key: expr => $val: expr ),*$(,)?) => {{
         let mut map = HashMap::default();
         $( map.insert($key, $val); )*
         map
    }}
}

mod labels;

#[test]
fn dictionary_literals() {
    test_infer! {
        src: r#"
            m = ["a": 0, "b": 1, "c": 2]
        "#,
        exp: map![
            "m" => "[string: int]",
        ],
    }
    test_infer! {
        env: map![
            "a" => "string",
            "b" => "string",
            "one" => "int",
            "two" => "int"
        ],
        src: r#"
            m = [a: one, b: two]
        "#,
        exp: map![
            "m" => "[string: int]",
        ],
    }
    test_infer! {
        src: r#"
            m = [1970-01-01T00:00:00Z: 0, 1970-01-01T01:00:00Z: 1]"#,
        exp: map![
            "m" => "[time: int]",
        ],
    }
    test_infer_err! {
        src: r#"
            m = ["1": 1.1, 2: 2.2]
        "#,
    }
    test_infer_err! {
        src: r#"
            m = ["a": "1", "b": 2]
        "#,
    }
    test_infer_err! {
        src: r#"
            m = [[]: 1]
        "#,
    }
}
#[test]
fn dictionary() {
    test_infer! {
        env: map![
            "fromList" => "(pairs: [{key: K, value: V}]) => [K: V] where K: Comparable",
            "get" => "(key: K, dict: [K: V], default: V) => V where K: Comparable",
            "insert" => "(key: K, value: V, dict: [K: V]) => [K: V] where K: Comparable",
            "remove" => "(key: K, dict: [K: V]) => [K: V] where K: Comparable",
        ],
        src: r#"
            d0 = fromList(pairs: [{key: "a0", value: 0}, {key: "a1", value: 1}])
            a0 = get(key: "a0", dict: d0, default: -1)
            d1 = insert(key: "a2", value: 2, dict: d0)
            d2 = remove(key: "a1", dict: d1)
        "#,
        exp: map![
            "d0" => "[string: int]",
            "a0" => "int",
            "d1" => "[string: int]",
            "d2" => "[string: int]",
        ],
    }
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
    test_infer! {
        env: map![
            "name" => "bool",
        ],
        src: r#"
            message = "Hello, ${name}!"
        "#,
        exp: map![
            "message" => "string"
        ],
    }
    test_infer! {
        env: map![
            "name" => "int",
        ],
        src: r#"
            message = "Hello, ${name}!"
        "#,
        exp: map![
            "message" => "string"
        ],

    }
    test_infer! {
        env: map![
            "name" => "uint",
        ],
        src: r#"
            message = "Hello, ${name}!"
        "#,
        exp: map![
            "message" => "string"
        ],
    }
    test_infer! {
        env: map![
            "name" => "float",
        ],
        src: r#"
            message = "Hello, ${name}!"
        "#,
        exp: map![
            "message" => "string"
        ],
    }
    test_infer! {
        env: map![
            "name" => "duration",
        ],
        src: r#"
            message = "Hello, ${name}!"
        "#,
        exp: map![
            "message" => "string"
        ],

    }
    test_infer! {
        env: map![
            "name" => "time",
        ],
        src: r#"
            message = "Hello, ${name}!"
        "#,
        exp: map![
            "message" => "string"
        ],
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
fn record_with_literal_fields() {
    test_infer! {
        env: map![
            "r" => r##"{ "with spaces": int, "#$%": string }"##,
        ],
        src: r##"
            o = {x: r["with spaces"], y: r["#$%"]}
        "##,
        exp: map![
            "o" => "{x: int , y: string}",
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
    // pipe args have different names however we infer `f` to have an anonymous pipe argument so
    // this passes
    test_infer! {
        src: r#"
            f = (arg=(x=<-) => x, w) => w |> arg()
            f(arg: (v=<-) => v, w: 0)
        "#,
        exp: map![
            "f" => "(w:A, ?arg:(<-:A) => B) => B",
        ]
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
}

#[test]
fn infer_pipe() {
    test_error_msg! {
        src: r#"
            f = (arg=(x=<-) => x) => 0 |> arg()
            g = () => f(arg: (x) => 5 + x)
        "#,
        expect: expect![[r#"
            error: missing pipe argument (argument arg)
              ┌─ main:3:30
              │
            3 │             g = () => f(arg: (x) => 5 + x)
              │                              ^^^^^^^^^^^^

            error: found unexpected argument x (argument arg)
              ┌─ main:3:30
              │
            3 │             g = () => f(arg: (x) => 5 + x)
              │                              ^^^^^^^^^^^^

        "#]],
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
            g = f(x: /.*/)
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
            "f" => "(x: G) => string where G: Stringable",
            "a" => "string",
        ],
    }
    test_infer_err! {
        src: r#"
            f = (x) => "x = ${x}"
            f(x: {a: 100, b: 10})
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
            f(a: /.*/, b: /.*/)
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
            f(a: /.*/, b: /.*/)
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
            f(a: /.*/, b: /.*/)
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
            f(a: /.*/, b: /.*/)
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
            f(a: /.*/, b: /.*/)
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
            f(a: /.*/, b: /.*/)
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
            "f" => "(a: A, ?b: A) => A where A: Addable",
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
            "f" => "(a: A, b: B, ?c: A, ?d: B) => {r: A, s: B} where A: Addable, B: Addable",
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
            "x" => "(a: D, ?b: D, <-m: D) => D where D: Addable",
            "z" => "(a: {B with m: A}, ?b: E, ?c: E, <-m: E) => {r: A , s: E} where E: Addable",
            "y" => "int",
            "v" => "{s: float, r: string}",
        ],
    }
}

#[test]
fn function_default_arguments_polymorphic() {
    test_infer! {
        src: r#"
            f = (x = "default") => x
            g = (x = 0) => x + x
        "#,
        exp: map![
            "f" => "(?x: A) => A",
            "g" => "(?x: B) => B where B: Addable",
        ],
    }
}

#[test]
fn issue_4051() {
    test_infer! {
        env: map![
            "r" => "A where A: Record",
        ],
        src: r#"
        f = (r) => {
            a = r.a
            b = r.b + 1
            c = r.c * 1.0
            return r
        }
        x = f(r:r)
        "#,
        exp: map![
            "f" => "(r:{S with a:X, b: int, c: float}) => {S with a:X, b: int, c: float}",
            "x" => "{S with a: X, b: int, c: float}",
        ],
    }
}

#[test]
fn copy_bindings_from_other_env() {
    let mut env = Environment::empty(true);
    let mut f = Fresher::default();
    let a = Symbol::from("a");
    env.add(
        a.clone(),
        PolyType {
            vars: Vec::new(),
            cons: BoundTvarKinds::new(),
            expr: MonoType::BOOL,
        },
    );
    let mut sub_env = Environment::new(env.clone());
    let b = Symbol::from("b");
    sub_env.add(
        b.clone(),
        PolyType {
            vars: Vec::new(),
            cons: BoundTvarKinds::new(),
            expr: MonoType::Var(f.fresh()),
        },
    );
    sub_env.copy_bindings_from(&env);
    assert_eq!(
        sub_env,
        Environment {
            external: None,
            parent: Some(env.clone().into()),
            readwrite: true,
            values: indexmap::indexmap!(
                b => PolyType {
                    vars: Vec::new(),
                    cons: BoundTvarKinds::new(),
                    expr: MonoType::Var(f.fresh()),
                },
                a => PolyType {
                    vars: Vec::new(),
                    cons: BoundTvarKinds::new(),
                    expr: MonoType::BOOL,
                }
            )
        }
    );
}

test_error_msg! {
    test: location_points_to_right_expression_error,
    src: r#"
            1 + "1"
        "#,
    // Location points to right expression expression
    expect: expect![[r#"
        error: expected int but found string
          ┌─ main:2:17
          │
        2 │             1 + "1"
          │                 ^^^

    "#]],
}

test_error_msg! {
    test: location_points_to_argument_of_unary_error,
    src: r#"
            -"s"
        "#,
    // Location points to argument of unary expression
    expect: expect![[r#"
        error: string is not Negatable
          ┌─ main:2:14
          │
        2 │             -"s"
          │              ^^^

    "#]],
}
test_error_msg! {
    test: location_points_to_entire_binary_error,
    src: r#"
            1h + 2h
        "#,
    // Location points to entire binary expression
    expect: expect![[r#"
        error: duration is not Addable
          ┌─ main:2:13
          │
        2 │             1h + 2h
          │             ^^^^^^^

    "#]],
}
test_error_msg! {
    test: location_points_to_second_interpolated_error,
    src: r#"
            bob = "Bob"
            joe = {a: 0, b: 0.1}
            "Hey ${bob} it's me ${joe}!"
        "#,
    // Location points to second interpolated expression
    expect: expect![[r#"
        error: {b: float, a: int} (record) is not Stringable
          ┌─ main:4:35
          │
        4 │             "Hey ${bob} it's me ${joe}!"
          │                                   ^^^

    "#]],
}
test_error_msg! {
    test: location_points_to_if_error,
    src: r#"
            if 0 then "a" else "b"
        "#,
    // Location points to if expression
    expect: expect![[r#"
        error: expected bool but found int
          ┌─ main:2:16
          │
        2 │             if 0 then "a" else "b"
          │                ^

    "#]],
}
test_error_msg! {
    test: location_points_to_else_error,
    src: r#"
            if exists 0 then 0 else "b"
        "#,
    // Location points to else expression
    expect: expect![[r#"
        error: expected int but found string
          ┌─ main:2:37
          │
        2 │             if exists 0 then 0 else "b"
          │                                     ^^^

    "#]],
}
test_error_msg! {
    test: location_points_to_second_element_error,
    src: r#"
            [1, "2"]
        "#,
    // Location points to second element of array
    expect: expect![[r#"
        error: expected int but found string
          ┌─ main:2:17
          │
        2 │             [1, "2"]
          │                 ^^^

    "#]],
}
test_error_msg! {
    test: location_points_to_index_error,
    src: r#"
            a = [1, 2, 3]
            a[1.1]
        "#,
    // Location points to expression representing the index
    expect: expect![[r#"
        error: expected int but found float
          ┌─ main:3:15
          │
        3 │             a[1.1]
          │               ^^^

    "#]],
}
test_error_msg! {
    test: location_points_to_right_error,
    src: r#"
            a = [1, 2, 3]
            a[1] + 1.1
        "#,
    // Location points to right expression
    expect: expect![[r#"
        error: expected int but found float
          ┌─ main:3:20
          │
        3 │             a[1] + 1.1
          │                    ^^^

    "#]],
}
test_error_msg! {
    test: location_points_to_identifier_a_error,
    src: r#"
            a = 1
            a[1]
        "#,
    // Location points to the identifier a
    expect: expect![[r#"
        error: expected [A] (array) but found int
          ┌─ main:3:13
          │
        3 │             a[1]
          │             ^

    "#]],
}
test_error_msg! {
    test: location_points_to_identifier_a_error_2,
    src: r#"
            a = [1, 2, 3]
            a.x
        "#,
    // Location points to the identifier a
    expect: expect![[r#"
        error: expected {A with x: B} (record) but found [int] (array)
          ┌─ main:3:13
          │
        3 │             a.x
          │             ^

    "#]],
}
test_error_msg! {
    test: location_points_to_entire_call_error,
    src: r#"
            f = (x, y) => x - y
            f(x: "x", y: "y")
        "#,
    // Location points to entire call expression
    expect: expect![[r#"
        error: string is not Subtractable (argument x)
          ┌─ main:3:18
          │
        3 │             f(x: "x", y: "y")
          │                  ^^^

    "#]],
}
test_error_msg! {
    test: location_points_to_entire_call_error_2,
    src: r#"
            f = (r) => r.a
            f(r: {b: 1})
        "#,
    // Location points to entire call expression
    expect: expect![[r#"
        error: record is missing label a (argument r)
          ┌─ main:3:18
          │
        3 │             f(r: {b: 1})
          │                  ^^^^^^

    "#]],
}
test_error_msg! {
    test: location_points_to_identifier_a_error_3,
    src: r#"
            x = 1 + 1
            a
        "#,
    // Location points to the identifier a
    expect: expect![[r#"
        error: undefined identifier a
          ┌─ main:3:13
          │
        3 │             a
          │             ^

    "#]],
}
test_error_msg! {
    test: location_points_to_call_error,
    src: r#"
            match = (o) => o.name =~ /^a/
            fn = (r) => match(r)
        "#,
    // Location points to call expression `match(r)`
    expect: expect![[r#"
            error: found unexpected argument r
              ┌─ main:3:31
              │
            3 │             fn = (r) => match(r)
              │                               ^

            error: missing required argument o
              ┌─ main:3:25
              │
            3 │             fn = (r) => match(r)
              │                         ^^^^^^^^

        "#]],
}
test_error_msg! {
    test: location_points_to_call_error_2,
    src: r#"
            f = (a, b) => a + b
            f(a: 0, c: 1)
        "#,
    // Location points to call expression `f(a: 0, c: 1)`
    expect: expect![[r#"
            error: found unexpected argument c
              ┌─ main:3:24
              │
            3 │             f(a: 0, c: 1)
              │                        ^

            error: missing required argument b
              ┌─ main:3:13
              │
            3 │             f(a: 0, c: 1)
              │             ^^^^^^^^^^^^^

        "#]],
}

test_error_msg! {
    test: location_points_to_call_error_3,
    src: r#"
            f = (a, b) => a + b
            f(a: 0)
        "#,
    // Location points to call expression `f(a: 0)`
    expect: expect![[r#"
        error: missing required argument b
          ┌─ main:3:13
          │
        3 │             f(a: 0)
          │             ^^^^^^^

    "#]],
}

#[test]
fn test_analyzer_returns_package_after_errors() {
    // Test that we can get a package result even if the source has errors.

    let mut analyzer = Analyzer::new(
        Environment::default(),
        Packages::default(),
        Default::default(),
    );
    match analyzer.analyze_source(
        "main".into(),
        "".into(),
        r#"
            x = () => 1
            y = x(
        "#,
    ) {
        Ok(_) => panic!("Unexpected success"),
        Err(err) => {
            let want = map![
                "x" => "() => int",
                "y" => "int",
            ];
            let got: SemanticMap<String, PolyType> = err
                .value
                .unwrap()
                .0
                .into_bindings()
                .map(|(k, v)| (k.to_string(), v))
                .collect();
            let want: SemanticMap<String, PolyType> = parse_map(Some("main"), want)
                .into_iter_by(|l, r| l.name().cmp(r.name()))
                .map(|(k, v)| (k.to_string(), v))
                .collect();
            assert_eq!(want, got);
        }
    }
}

#[test]
fn undefined_variable_has_the_same_type_across_multiple_uses() {
    test_error_msg! {
        src: r#"
            x = y
            z = x + 1
            z2 = x + "a"
        "#,
        expect: expect![[r#"
            error: undefined identifier y
              ┌─ main:2:17
              │
            2 │             x = y
              │                 ^

            error: expected int but found string
              ┌─ main:4:22
              │
            4 │             z2 = x + "a"
              │                      ^^^

        "#]],
    }
}

#[test]
#[ignore]
fn error_types_do_not_suppress_additional_actual_errors() {
    test_error_msg! {
        src: r#"
            x = y - 1
            z = x + ""
        "#,
        // Should get an int != string error here, but binary expressions are inferred as
        // `left <> right` and `left <> result` instead of being unified with a function type
        // `(left, right) => return <> (A, A) => A` (where `A` is a type variable)
        err: r#"error @2:17-2:18: undefined identifier y

              error @3:17-2:18: expected int but found string"#,
    }
}

#[test]
fn parse_and_inference_errors_are_reported_simultaneously() {
    test_error_msg! {
        src: r#"
            x = / 1
            z = y + 1
        "#,
        err: "error @2:17-2:18: invalid expression: invalid token for primary expression: DIV

error @3:17-3:18: undefined identifier y",
    }
}

#[test]
fn primitive_kind_errors() {
    test_error_msg! {
        env: map![
            "isType" => "(v: A, type: string) => bool where A: Basic",
        ],
        src: r#"
            isType(v: {}, type: "record")
            isType(v: [], type: "array")
        "#,
        expect: expect_test::expect![[r#"
            error: {} (record) is not Basic (argument v)
              ┌─ main:2:23
              │
            2 │             isType(v: {}, type: "record")
              │                       ^^

            error: [A] (array) is not Basic (argument v)
              ┌─ main:3:23
              │
            3 │             isType(v: [], type: "array")
              │                       ^^

        "#]]
    }
}

#[test]
fn primitive_kind_short_errors() {
    test_error_msg! {
        env: map![
            "isType" => "(v: A, type: string) => bool where A: Basic",
        ],
        src: r#"
            isType(v: {}, type: "record")
            isType(v: [], type: "array")
        "#,
        expect_short: expect_test::expect![[r#"
            main:2:23: error: {} (record) is not Basic (argument v)
            main:3:23: error: [A] (array) is not Basic (argument v)
        "#]]
    }
}

#[test]
fn invalid_mono_type() {
    test_error_msg! {
        src: r#"
            builtin x : abc
        "#,
        expect: expect_test::expect![[r#"
            error: invalid named type abc
              ┌─ main:2:25
              │
            2 │             builtin x : abc
              │                         ^^^

        "#]]
    }
}

#[test]
fn missing_return() {
    test_error_msg! {
        src: r#"
            () => { }
        "#,
        expect: expect_test::expect![[r#"
            error: missing return statement in block
              ┌─ main:2:19
              │
            2 │             () => { }
              │                   ^^^

        "#]]
    }
}

#[test]
fn symbol_resolution() {
    let imp = map![
        "types" => package![
            "isType" => "(v: A, type: string) => bool } where A: Basic",
        ],
    ];
    let src = r#"
            import "types"
            // Comment on x
            x = types.isType(v: 1, type: "int")

            // Comment on foo
            foo = () => (1)
            foo()

            types = { isType: (v, type) => 1 }
            y = types.isType(v: 1, type: "int")

            t = types
            z = t.isType(v: 1, type: "int")
        "#;
    let (package_exports, pkg) =
        infer_types(src, Default::default(), imp, None, Default::default())
            .unwrap_or_else(|err| panic!("{}", err));

    let mut member_expr_1 = None;
    let mut member_expr_2 = None;
    let mut member_expr_3 = None;
    let mut ident_expr = None;
    semantic::walk::walk(
        &mut |node| {
            if let semantic::walk::Node::MemberExpr(e) = node {
                if e.loc.start.line == 4 {
                    member_expr_1 = Some(e);
                }
                if e.loc.start.line == 11 {
                    member_expr_2 = Some(e);
                }
                if e.loc.start.line == 14 {
                    member_expr_3 = Some(e);
                }
            }
            if let semantic::walk::Node::IdentifierExpr(e) = node {
                if e.name == "foo" {
                    ident_expr = Some(e);
                }
            }
        },
        semantic::walk::Node::Package(&pkg),
    );
    assert_eq!(
        member_expr_1.expect("member expression").property,
        Symbol::from("isType").with_package("types").to_string()
    );
    assert_eq!(
        ident_expr.expect("ident expression").name,
        Symbol::from("foo").with_package("main").to_string()
    );
    assert_eq!(member_expr_2.expect("member expression").property, "isType");

    // Not currently detected as from the `types` package but could be with better analysis
    assert_eq!(member_expr_3.expect("member expression").property, "isType");

    assert_eq!(
        package_exports.get_entry("x").map(|e| &e.comments[..]),
        Some(&["// Comment on x\n".to_string()][..]),
    );

    assert_eq!(
        package_exports.get_entry("foo").map(|e| &e.comments[..]),
        Some(&["// Comment on foo\n".to_string()][..]),
    );
}

#[test]
fn multiple_errors_in_function_call() {
    test_error_msg! {
        env: map![
            "f" => "(a: float, b: int, c: string) => bool",
        ],
        src: r#"
            f(a: 1, b: "record", d: {})
        "#,
        expect: expect![[r#"
            error: found unexpected argument d
              ┌─ main:2:37
              │
            2 │             f(a: 1, b: "record", d: {})
              │                                     ^^

            error: expected float but found int (argument a)
              ┌─ main:2:18
              │
            2 │             f(a: 1, b: "record", d: {})
              │                  ^

            error: expected int but found string (argument b)
              ┌─ main:2:24
              │
            2 │             f(a: 1, b: "record", d: {})
              │                        ^^^^^^^^

            error: missing required argument c
              ┌─ main:2:13
              │
            2 │             f(a: 1, b: "record", d: {})
              │             ^^^^^^^^^^^^^^^^^^^^^^^^^^^

        "#]]
    }
}

#[test]
fn unused_variable() {
    test_error_msg! {
        config: AnalyzerConfig{
            features: vec![Feature::UnusedSymbolWarnings],
            ..AnalyzerConfig::default()
        },
        src: r#"
            f = () => {
                x = "" + 1
                return 1
            }
        "#,
        expect: expect_test::expect![[r#"
            warning: symbol x is never used
              ┌─ main:3:17
              │
            3 │                 x = "" + 1
              │                 ^

            error: expected string but found int
              ┌─ main:3:26
              │
            3 │                 x = "" + 1
              │                          ^

        "#]]
    }
}

#[test]
fn no_unused_variable_warning_for_function_parameter() {
    test_error_msg! {
        config: AnalyzerConfig{
            features: vec![Feature::UnusedSymbolWarnings],
            ..AnalyzerConfig::default()
        },
        src: r#"
            f = (x) => {
                return 1 + ""
            }
        "#,
        expect: expect_test::expect![[r#"
            error: expected int but found string
              ┌─ main:3:28
              │
            3 │                 return 1 + ""
              │                            ^^

        "#]]
    }
}

#[test]
fn unused_import() {
    test_error_msg! {
        config: AnalyzerConfig{
            features: vec![Feature::UnusedSymbolWarnings],
            ..AnalyzerConfig::default()
        },
        imp: map![
            "path/to/foo" => package![
                "f" => " (x: A) => A where A: Addable + Divisible",
            ],
        ],
        src: r#"
            import "path/to/foo"

            x = 1 + ""
        "#,
        expect: expect_test::expect![[r#"
            warning: symbol foo is never used
              ┌─ main:2:13
              │
            2 │             import "path/to/foo"
              │             ^^^^^^^^^^^^^^^^^^^^

            error: expected int but found string
              ┌─ main:4:21
              │
            4 │             x = 1 + ""
              │                     ^^

        "#]]
    }
}

#[test]
fn vec_type() {
    test_infer! {
        env: map![
            "vec" => "vector[int]",
        ],
        src: r#"
            builtin _vecFloat: (v: vector[T]) => vector[float]

            x = _vecFloat(v: vec)
        "#,
        exp: map![
            "_vecFloat" => "(v: vector[T]) => vector[float]",
            "x" => "vector[float]",
        ],
    }
}

#[test]
fn pipe_error() {
    test_error_msg! {
        config: AnalyzerConfig{
            features: vec![Feature::UnusedSymbolWarnings],
            ..AnalyzerConfig::default()
        },
        env: map![
            "findColumn" =>  "() => [A] where A: Record",
            "yield" => "(<-tables: stream[A]) => stream[A] where A: Record",
        ],
        src: r#"

            findColumn()
                |> yield()
        "#,
        expect: expect_test::expect![[r#"
            error: expected stream[A] but found [A] (array) (argument tables)
              ┌─ main:4:20
              │
            4 │                 |> yield()
              │                    ^^^^^^^

        "#]]
    }
}

#[test]
fn multiple_builtins() {
    test_infer! {
        src: r#"
            builtin x : (x: string) => string

            // @feature labelPolymorphism
            builtin x : (x: A) => string where A: Label
        "#,
        exp: map![
            "x" => "(x: string) => string",
        ],
    }

    test_infer! {
        config: AnalyzerConfig{
            features: vec![Feature::LabelPolymorphism],
            ..AnalyzerConfig::default()
        },
        src: r#"
            builtin x : (x: string) => string

            // @feature labelPolymorphism
            builtin x : (x: A) => string where A: Label
        "#,
        exp: map![
            "x" => "(x: A) => string where A: Label",
        ],
    }
}

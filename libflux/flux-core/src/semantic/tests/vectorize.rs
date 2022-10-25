use std::sync::Arc;

use super::*;

use crate::semantic::{
    nodes::{FunctionExpr, Package},
    walk::{walk, Node},
    AnalyzerConfig,
};

fn analyzer_config() -> AnalyzerConfig {
    AnalyzerConfig::default()
}

fn vectorize(src: &str) -> anyhow::Result<Package> {
    // packages/symbols which should be exposed to these tests can be defined here.
    let imp = map![
        "boolean" => package![
            "true" => "bool",
            "false" => "bool",
        ],
        "universe" => package![
            "float" => "(v: A) => float",
        ],
    ];
    let imports: SemanticMap<&str, _> = imp
        .into_iter()
        .map(|(path, pkg)| (path, parse_map(Some(path), pkg)))
        .collect();
    let importer: Packages = imports
        .into_iter()
        .map(|(path, types)| {
            (
                path.to_string(),
                Arc::new(PackageExports::try_from(types).unwrap()),
            )
        })
        .collect();
    let mut prelude = PackageExports::new();
    prelude.copy_bindings_from(importer.get("boolean").unwrap());
    prelude.copy_bindings_from(importer.get("universe").unwrap());
    let env = Environment::from(&prelude);

    let mut analyzer = Analyzer::new(env, importer, analyzer_config());
    let (_, pkg) = analyzer
        .analyze_source("main".into(), "".into(), src)
        .map_err(|err| err.error)?;

    Ok(pkg)
}

fn get_vectorized_function(pkg: &Package) -> &FunctionExpr {
    let mut function = None;
    walk(
        &mut |node| {
            if let Node::FunctionExpr(func) = node {
                if func.vectorized.is_some() {
                    function = func.vectorized.as_ref();
                }
            }
        },
        Node::Package(pkg),
    );
    function.expect("function")
}

#[test]
fn vectorize_field_access() -> anyhow::Result<()> {
    let pkg = vectorize(r#"(r) => ({ a: r.a, b: r.b })"#)?;

    let function = get_vectorized_function(&pkg);

    expect_test::expect![[r##"
        (r) => {
            return {a: r:{#F with b: v[#B], a: v[#D]}.a:v[#D], b: r:{#F with b: v[#B], a: v[#D]}.b:v[#B]}:{a: v[#D], b: v[#B]}
        }:(r: {#F with b: v[#B], a: v[#D]}) => {a: v[#D], b: v[#B]}"##]].assert_eq(&crate::semantic::formatter::format_node(
        Node::FunctionExpr(function),
    )?);

    Ok(())
}

#[test]
fn vectorize_with_construction() -> anyhow::Result<()> {
    let pkg = vectorize(r#"(r) => ({ r with b: r.a })"#)?;

    let function = get_vectorized_function(&pkg);

    expect_test::expect![[r##"
        (r) => {
            return {r:{#C with a: v[#B]} with b: r:{#C with a: v[#B]}.a:v[#B]}:{#C with b: v[#B], a: v[#B]}
        }:(r: {#C with a: v[#B]}) => {#C with b: v[#B], a: v[#B]}"##]]
        .assert_eq(&crate::semantic::formatter::format_node(
            Node::FunctionExpr(function),
        )?);

    Ok(())
}

#[test]
fn vectorize_even_when_another_function_fails_to_vectorize() -> anyhow::Result<()> {
    let pkg = vectorize(
        r#"
        map = (fn) => fn
        map(fn: (r) => ({r with x: r.a + r.b}))
    "#,
    )?;

    let function = get_vectorized_function(&pkg);

    expect_test::expect![[r##"
        (r) => {
            return {r:{#I with a: v[#G], b: v[#G]} with x: r:{#I with a: v[#G], b: v[#G]}.a:v[#G] +:v[#G] r:{#I with a: v[#G], b: v[#G]}.b:v[#G]}:{#I with x: v[#G], a: v[#G], b: v[#G]}
        }:(r: {#I with a: v[#G], b: v[#G]}) => {#I with x: v[#G], a: v[#G], b: v[#G]}"##]]
        .assert_eq(&crate::semantic::formatter::format_node(
            Node::FunctionExpr(function),
        )?);

    Ok(())
}

#[test]
fn vectorize_non_map_like_function() {
    let mut pkg = vectorize(r#"(s) => ({ x: s.a })"#).unwrap();

    let err = semantic::vectorize::vectorize(&analyzer_config(), &mut pkg).unwrap_err();

    expect_test::expect![[
        r#"error @1:1-1:20: can't vectorize function: Does not match the `map` signature"#
    ]]
    .assert_eq(&err.to_string());
}

#[test]
fn vectorize_addition_operator() -> anyhow::Result<()> {
    let pkg = vectorize(r#"(r) => ({ x: r.a + r.b })"#)?;

    let function = get_vectorized_function(&pkg);

    expect_test::expect![[r##"
        (r) => {
            return {x: r:{#F with a: v[#D], b: v[#D]}.a:v[#D] +:v[#D] r:{#F with a: v[#D], b: v[#D]}.b:v[#D]}:{x: v[#D]}
        }:(r: {#F with a: v[#D], b: v[#D]}) => {x: v[#D]}"##]].assert_eq(&crate::semantic::formatter::format_node(
        Node::FunctionExpr(function),
    )?);

    Ok(())
}

#[test]
fn vectorize_subtraction_operator() -> anyhow::Result<()> {
    let pkg = vectorize(r#"(r) => ({ x: r.a - r.b })"#).unwrap();

    let function = get_vectorized_function(&pkg);

    expect_test::expect![[r##"
        (r) => {
            return {x: r:{#F with a: v[#D], b: v[#D]}.a:v[#D] -:v[#D] r:{#F with a: v[#D], b: v[#D]}.b:v[#D]}:{x: v[#D]}
        }:(r: {#F with a: v[#D], b: v[#D]}) => {x: v[#D]}"##]].assert_eq(&crate::semantic::formatter::format_node(
        Node::FunctionExpr(function),
    )?);

    Ok(())
}

#[test]
fn vectorize_eq_operator() -> anyhow::Result<()> {
    let pkg = vectorize(r#"(r) => ({ x: r.a == r.b })"#).unwrap();

    let function = get_vectorized_function(&pkg);

    expect_test::expect![[r##"
        (r) => {
            return {x: r:{#F with a: v[#B], b: v[#D]}.a:v[#B] ==:v[bool] r:{#F with a: v[#B], b: v[#D]}.b:v[#D]}:{x: v[bool]}
        }:(r: {#F with a: v[#B], b: v[#D]}) => {x: v[bool]}"##]].assert_eq(&crate::semantic::formatter::format_node(
        Node::FunctionExpr(function),
    )?);

    Ok(())
}

#[test]
fn vectorize_neq_operator() -> anyhow::Result<()> {
    let pkg = vectorize(r#"(r) => ({ x: r.a != r.b })"#).unwrap();

    let function = get_vectorized_function(&pkg);

    expect_test::expect![[r##"
        (r) => {
            return {x: r:{#F with a: v[#B], b: v[#D]}.a:v[#B] !=:v[bool] r:{#F with a: v[#B], b: v[#D]}.b:v[#D]}:{x: v[bool]}
        }:(r: {#F with a: v[#B], b: v[#D]}) => {x: v[bool]}"##]].assert_eq(&crate::semantic::formatter::format_node(
        Node::FunctionExpr(function),
    )?);

    Ok(())
}

#[test]
fn vectorize_lt_operator() -> anyhow::Result<()> {
    let pkg = vectorize(r#"(r) => ({ x: r.a < r.b })"#).unwrap();

    let function = get_vectorized_function(&pkg);

    expect_test::expect![[r##"
        (r) => {
            return {x: r:{#F with a: v[#B], b: v[#D]}.a:v[#B] <:v[bool] r:{#F with a: v[#B], b: v[#D]}.b:v[#D]}:{x: v[bool]}
        }:(r: {#F with a: v[#B], b: v[#D]}) => {x: v[bool]}"##]].assert_eq(&crate::semantic::formatter::format_node(
        Node::FunctionExpr(function),
    )?);

    Ok(())
}

#[test]
fn vectorize_lte_operator() -> anyhow::Result<()> {
    let pkg = vectorize(r#"(r) => ({ x: r.a <= r.b })"#).unwrap();

    let function = get_vectorized_function(&pkg);

    expect_test::expect![[r##"
        (r) => {
            return {x: r:{#F with a: v[#B], b: v[#D]}.a:v[#B] <=:v[bool] r:{#F with a: v[#B], b: v[#D]}.b:v[#D]}:{x: v[bool]}
        }:(r: {#F with a: v[#B], b: v[#D]}) => {x: v[bool]}"##]].assert_eq(&crate::semantic::formatter::format_node(
        Node::FunctionExpr(function),
    )?);

    Ok(())
}

#[test]
fn vectorize_gt_operator() -> anyhow::Result<()> {
    let pkg = vectorize(r#"(r) => ({ x: r.a > r.b })"#).unwrap();

    let function = get_vectorized_function(&pkg);

    expect_test::expect![[r##"
        (r) => {
            return {x: r:{#F with a: v[#B], b: v[#D]}.a:v[#B] >:v[bool] r:{#F with a: v[#B], b: v[#D]}.b:v[#D]}:{x: v[bool]}
        }:(r: {#F with a: v[#B], b: v[#D]}) => {x: v[bool]}"##]].assert_eq(&crate::semantic::formatter::format_node(
        Node::FunctionExpr(function),
    )?);

    Ok(())
}

#[test]
fn vectorize_gte_operator() -> anyhow::Result<()> {
    let pkg = vectorize(r#"(r) => ({ x: r.a >= r.b })"#).unwrap();

    let function = get_vectorized_function(&pkg);

    expect_test::expect![[r##"
        (r) => {
            return {x: r:{#F with a: v[#B], b: v[#D]}.a:v[#B] >=:v[bool] r:{#F with a: v[#B], b: v[#D]}.b:v[#D]}:{x: v[bool]}
        }:(r: {#F with a: v[#B], b: v[#D]}) => {x: v[bool]}"##]].assert_eq(&crate::semantic::formatter::format_node(
        Node::FunctionExpr(function),
    )?);

    Ok(())
}

#[test]
fn vectorizing_non_vector_variables_are_not_implemented() {
    let mut pkg = vectorize(
        r#"
        var = 1
        f = (r) => ({ x: var })
    "#,
    )
    .unwrap();

    let err = semantic::vectorize::vectorize(&analyzer_config(), &mut pkg).unwrap_err();

    expect_test::expect![[
        r#"error @3:26-3:29: can't vectorize function: Unable to vectorize non-vector symbol `var`"#
    ]]
    .assert_eq(&err.to_string());
}

#[test]
fn vectorize_with_construction_using_literal_float() -> anyhow::Result<()> {
    let pkg = vectorize(r#"(r) => ({ r with a: 1.0 })"#)?;

    let function = get_vectorized_function(&pkg);

    expect_test::expect![[r##"
        (r) => {
            return {r:#A with a: ~~vecRepeat~~:float(v: 1.0):v[float]}:{#A with a: v[float]}
        }:(r: #A) => {#A with a: v[float]}"##]]
    .assert_eq(&crate::semantic::formatter::format_node(
        Node::FunctionExpr(function),
    )?);

    Ok(())
}

#[test]
fn vectorize_with_construction_using_const_folding() -> anyhow::Result<()> {
    let pkg = vectorize(r#"(r) => ({ r with a: 1.0 + 2.0 })"#).unwrap();

    let function = get_vectorized_function(&pkg);

    expect_test::expect![[r##"
        (r) => {
            return {r:#A with a: ~~vecRepeat~~:float(v: 1.0):v[float] +:v[float] ~~vecRepeat~~:float(v: 2.0):v[float]}:{#A with a: v[float]}
        }:(r: #A) => {#A with a: v[float]}"##]]
        .assert_eq(&crate::semantic::formatter::format_node(
            Node::FunctionExpr(function),
        )?);
    Ok(())
}

#[test]
fn vectorize_with_construction_using_literal_string() -> anyhow::Result<()> {
    let pkg = vectorize(r#"(r) => ({ r with a: "hello" })"#)?;

    let function = get_vectorized_function(&pkg);

    expect_test::expect![[r##"
        (r) => {
            return {r:#A with a: ~~vecRepeat~~:string(v: "hello"):v[string]}:{#A with a: v[string]}
        }:(r: #A) => {#A with a: v[string]}"##]]
    .assert_eq(&crate::semantic::formatter::format_node(
        Node::FunctionExpr(function),
    )?);

    Ok(())
}

#[test]
fn vectorize_with_construction_using_literal_int() -> anyhow::Result<()> {
    let pkg = vectorize(r#"(r) => ({ r with a: 1 })"#)?;

    let function = get_vectorized_function(&pkg);

    expect_test::expect![[r##"
        (r) => {
            return {r:#A with a: ~~vecRepeat~~:int(v: 1):v[int]}:{#A with a: v[int]}
        }:(r: #A) => {#A with a: v[int]}"##]]
    .assert_eq(&crate::semantic::formatter::format_node(
        Node::FunctionExpr(function),
    )?);

    Ok(())
}

#[test]
fn vectorize_with_construction_using_literal_bool() -> anyhow::Result<()> {
    let pkg = vectorize(r#"(r) => ({ r with a: false, b: true })"#)?;

    let function = get_vectorized_function(&pkg);

    expect_test::expect![[r##"
        (r) => {
            return {r:#A with a: ~~vecRepeat~~:bool(v: false):v[bool], b: ~~vecRepeat~~:bool(v: true):v[bool]}:{#A with a: v[bool], b: v[bool]}
        }:(r: #A) => {#A with a: v[bool], b: v[bool]}"##]]
    .assert_eq(&crate::semantic::formatter::format_node(
        Node::FunctionExpr(function),
    )?);

    Ok(())
}

#[test]
fn vectorize_with_construction_using_literal_duration_not_implemented() -> anyhow::Result<()> {
    let mut pkg = vectorize(r#"(r) => ({ r with a: 1h })"#)?;

    let err = semantic::vectorize::vectorize(&analyzer_config(), &mut pkg).unwrap_err();

    expect_test::expect![[
        r#"error @1:21-1:23: can't vectorize function: Unable to vectorize expression"#
    ]]
    .assert_eq(&err.to_string());
    Ok(())
}

#[test]
fn vectorize_with_construction_using_literal_time() -> anyhow::Result<()> {
    let pkg = vectorize(r#"(r) => ({ r with a: 2021-11-01 })"#)?;

    let function = get_vectorized_function(&pkg);

    expect_test::expect![[r##"
        (r) => {
            return {r:#A with a: ~~vecRepeat~~:time(v: 2021-11-01T00:00:00Z):v[time]}:{#A with a: v[time]}
        }:(r: #A) => {#A with a: v[time]}"##]]
        .assert_eq(&crate::semantic::formatter::format_node(
            Node::FunctionExpr(function),
        )?);

    Ok(())
}

#[test]
fn vectorize_with_conditional_expr() -> anyhow::Result<()> {
    let pkg = vectorize(r#"(r) => ({ r with a: if r.cond then 1 else 0 })"#).unwrap();

    let function = get_vectorized_function(&pkg);

    expect_test::expect![[r##"
        (r) => {
            return {r:{#C with cond: v[bool]} with a: (if r:{#C with cond: v[bool]}.cond:v[bool] then ~~vecRepeat~~:int(v: 1):v[int] else ~~vecRepeat~~:int(v: 0):v[int]):v[int]}:{#C with a: v[int], cond: v[bool]}
        }:(r: {#C with cond: v[bool]}) => {#C with a: v[int], cond: v[bool]}"##]]
        .assert_eq(&crate::semantic::formatter::format_node(
            Node::FunctionExpr(function),
        )?);
    Ok(())
}

#[test]
fn vectorize_with_float_calls() -> anyhow::Result<()> {
    let pkg = vectorize(r#"(r) => ({ r with a: float(v: r._value) })"#).unwrap();

    let function = get_vectorized_function(&pkg);

    expect_test::expect![[r##"
        (r) => {
            return {r:{#E with _value: v[#D]} with a: _vectorizedFloat:v[float](v: r:{#E with _value: v[#D]}._value:v[#D]):v[float]}:{#E with a: v[float], _value: v[#D]}
        }:(r: {#E with _value: v[#D]}) => {#E with a: v[float], _value: v[#D]}"##]].assert_eq(&crate::semantic::formatter::format_node(
        Node::FunctionExpr(function),
    )?);
    Ok(())
}

#[test]
fn vectorize_with_unary_add_sub() -> anyhow::Result<()> {
    let pkg = vectorize(r#"(r) => ({ r with add: +r._value, sub: -r._value })"#).unwrap();

    let function = get_vectorized_function(&pkg);

    expect_test::expect![[r##"
        (r) => {
            return {r:{#C with _value: v[#B]} with add: +r:{#C with _value: v[#B]}._value:v[#B]:v[#B], sub: -r:{#C with _value: v[#B]}._value:v[#B]:v[#B]}:{#C with add: v[#B], sub: v[#B], _value: v[#B]}
        }:(r: {#C with _value: v[#B]}) => {#C with add: v[#B], sub: v[#B], _value: v[#B]}"##]].assert_eq(&crate::semantic::formatter::format_node(
        Node::FunctionExpr(function),
    )?);
    Ok(())
}

#[test]
fn vectorize_with_unary_not() -> anyhow::Result<()> {
    let pkg = vectorize(r#"(r) => ({ r with a: not r._value })"#).unwrap();

    let function = get_vectorized_function(&pkg);

    expect_test::expect![[r##"
        (r) => {
            return {r:{#C with _value: v[bool]} with a: not r:{#C with _value: v[bool]}._value:v[bool]:v[bool]}:{#C with a: v[bool], _value: v[bool]}
        }:(r: {#C with _value: v[bool]}) => {#C with a: v[bool], _value: v[bool]}"##]].assert_eq(&crate::semantic::formatter::format_node(
        Node::FunctionExpr(function),
    )?);
    Ok(())
}

#[test]
fn vectorize_with_unary_exists() -> anyhow::Result<()> {
    let pkg = vectorize(r#"(r) => ({ r with a: exists r._value })"#).unwrap();

    let function = get_vectorized_function(&pkg);

    expect_test::expect![[r##"
        (r) => {
            return {r:{#C with _value: v[#B]} with a: exists r:{#C with _value: v[#B]}._value:v[#B]:v[bool]}:{#C with a: v[bool], _value: v[#B]}
        }:(r: {#C with _value: v[#B]}) => {#C with a: v[bool], _value: v[#B]}"##]].assert_eq(&crate::semantic::formatter::format_node(
        Node::FunctionExpr(function),
    )?);
    Ok(())
}

#[test]
fn vectorize_with_conditional_nested_logical() -> anyhow::Result<()> {
    let pkg = vectorize(r#"(r) => ({x: if r.a and r.b then r.x else r.y})"#).unwrap();

    let function = get_vectorized_function(&pkg);

    expect_test::expect![[r##"
        (r) => {
            return {x: (if r:{#O with a: v[bool], b: v[bool], x: v[#K], y: v[#K]}.a:v[bool] and:bool r:{#O with a: v[bool], b: v[bool], x: v[#K], y: v[#K]}.b:v[bool] then r:{#O with a: v[bool], b: v[bool], x: v[#K], y: v[#K]}.x:v[#K] else r:{#O with a: v[bool], b: v[bool], x: v[#K], y: v[#K]}.y:v[#K]):v[#G]}:{x: v[#G]}
        }:(r: {#O with a: v[bool], b: v[bool], x: v[#K], y: v[#K]}) => {x: v[#G]}"##]].assert_eq(&crate::semantic::formatter::format_node(
        Node::FunctionExpr(function),
    )?);
    Ok(())
}

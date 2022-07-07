use super::*;
use crate::semantic::{
    import::Packages,
    nodes::{FunctionExpr, Package},
    walk::{walk, Node},
    AnalyzerConfig, Feature,
};

fn analyzer_config() -> AnalyzerConfig {
    AnalyzerConfig {
        features: vec![Feature::VectorizedMap],
        ..AnalyzerConfig::default()
    }
}

fn vectorize(src: &str) -> anyhow::Result<Package> {
    let mut analyzer = Analyzer::new(Default::default(), Packages::default(), analyzer_config());
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
        Node::Package(&pkg),
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
#[ignore] // FIXME: "undefined identifier true"
fn vectorize_with_construction_using_literal_bool() -> anyhow::Result<()> {
    let pkg = vectorize(r#"(r) => ({ r with a: true })"#)?;

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

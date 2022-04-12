use super::*;
use crate::semantic::{
    import::Packages,
    nodes::{FunctionExpr, Package},
    walk::{walk, Node},
    AnalyzerConfig, Feature,
};

fn vectorize(src: &str) -> anyhow::Result<Package> {
    let mut analyzer = Analyzer::new(
        Default::default(),
        Packages::default(),
        AnalyzerConfig {
            features: vec![Feature::VectorizedMap],
            ..AnalyzerConfig::default()
        },
    );
    let (_, mut pkg) = analyzer
        .analyze_source("main".into(), "".into(), src)
        .map_err(|err| err.error)?;

    semantic::vectorize::vectorize(&analyzer.config, &mut pkg)?;
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
            return {a: r:{F with b:v[#B], a:v[#D]}.a:v[#D], b: r:{F with b:v[#B], a:v[#D]}.b:v[#B]}:{a:v[#D], b:v[#B]}
        }:(r:{F with b:v[#B], a:v[#D]}) => {a:v[#D], b:v[#B]}"##]].assert_eq(&crate::semantic::formatter::format_node(
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
            return {r:{C with a:v[#B]} with b: r:{C with a:v[#B]}.a:v[#B]}:{C with b:v[#B], a:v[#B]}
        }:(r:{C with a:v[#B]}) => {C with b:v[#B], a:v[#B]}"##]]
    .assert_eq(&crate::semantic::formatter::format_node(
        Node::FunctionExpr(function),
    )?);

    Ok(())
}

#[test]
fn vectorize_with_construction_and_addition() -> anyhow::Result<()> {
    let pkg = vectorize(
        r#"
        builtin map: (fn: A) => A
        map(fn: (r) => ({r with x: r.a + r.b}))
    "#,
    )?;

    let function = get_vectorized_function(&pkg);

    expect_test::expect![[r##"
        (r) => {
            return {r:{I with a:v[#G], b:v[#G]} with x: r:{I with a:v[#G], b:v[#G]}.a:v[#G] +:v[#G] r:{I with a:v[#G], b:v[#G]}.b:v[#G]}:{I with x:v[#G], a:v[#G], b:v[#G]}
        }:(r:{I with a:v[#G], b:v[#G]}) => {I with x:v[#G], a:v[#G], b:v[#G]}"##]]
    .assert_eq(&crate::semantic::formatter::format_node(
        Node::FunctionExpr(function),
    )?);

    Ok(())
}

#[test]
fn vectorize_non_map_like_function() {
    let err = vectorize(r#"(s) => ({ x: s.a })"#).unwrap_err();

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
            return {x: r:{F with a:v[#D], b:v[#D]}.a:v[#D] +:v[#D] r:{F with a:v[#D], b:v[#D]}.b:v[#D]}:{x:v[#D]}
        }:(r:{F with a:v[#D], b:v[#D]}) => {x:v[#D]}"##]].assert_eq(&crate::semantic::formatter::format_node(
        Node::FunctionExpr(function),
    )?);

    Ok(())
}

#[test]
fn vectorize_subtraction_operator() {
    let err = vectorize(r#"(r) => ({ x: r.a - r.b })"#).unwrap_err();

    expect_test::expect![[
        r#"error @1:14-1:23: can't vectorize function: Unable to vectorize non-addition operators"#
    ]]
    .assert_eq(&err.to_string());
}

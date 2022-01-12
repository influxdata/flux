use super::*;
use crate::semantic::{
    import::Packages,
    nodes::{FunctionExpr, Package},
    walk::{walk, Node},
};

fn vectorize(src: &str) -> anyhow::Result<Package> {
    let mut analyzer = Analyzer::new(Default::default(), Packages::default(), Default::default());
    let (_, mut pkg) = analyzer.analyze_source("main".into(), "".into(), src)?;

    semantic::nodes::vectorize(&mut pkg)?;
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

    expect_test::expect![[r#"
        (r) => {
            return {a: r:{F with b:v[B], a:v[D]}.a:v[D], b: r:{F with b:v[B], a:v[D]}.b:v[B]}:{a:v[D], b:v[B]}
        }:(r:{F with b:B, a:D}) => {a:D, b:B}"#]].assert_eq(&crate::semantic::formatter::format_node(
            Node::FunctionExpr(function),
    )?);

    Ok(())
}

#[test]
fn vectorize_with_construction() -> anyhow::Result<()> {
    let pkg = vectorize(r#"(r) => ({ r with b: r.a })"#)?;

    let function = get_vectorized_function(&pkg);

    expect_test::expect![[r#"
        (r) => {
            return {r:{C with a:v[B]} with b: r:{C with a:v[B]}.a:v[B]}:{C with b:v[B], a:v[B]}
        }:(r:{C with a:B}) => {C with b:B, a:B}"#]]
    .assert_eq(&crate::semantic::formatter::format_node(
        Node::FunctionExpr(function),
    )?);

    Ok(())
}

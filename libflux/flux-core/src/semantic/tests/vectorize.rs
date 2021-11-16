use super::*;
use crate::semantic::{
    nodes::{FunctionExpr, Package},
    walk::{walk, Node},
};

fn vectorize(src: &str) -> anyhow::Result<Package> {
    let pkg = parse_program(src);
    let mut analyzer = Analyzer::new(
        Environment::default(),
        HashMap::default(),
        Default::default(),
    );
    let (_, mut pkg) = analyzer.analyze_ast(pkg)?;

    semantic::nodes::vectorize(&mut pkg)?;
    Ok(pkg)
}

fn get_vectorized_function(pkg: &Package) -> &FunctionExpr {
    let mut function = None;
    walk(
        &mut |node| {
            if let Node::FunctionExpr(func) = node {
                function = func.vectorized.as_ref();
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
            return {a: r:{J with b:v[D], a:v[B]}.a:v[B], b: r:{J with b:v[D], a:v[B]}.b:v[D]}:{a:v[B], b:v[D]}
        }:(r:{J with b:D, a:B}) => {a:B, b:D}"#]].assert_eq(&crate::semantic::formatter::format_node(
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
            return {r:{G with a:v[B]} with b: r:{G with a:v[B]}.a:v[B]}:{G with b:v[B], a:v[B]}
        }:(r:{G with a:B}) => {G with b:B, a:B}"#]]
    .assert_eq(&crate::semantic::formatter::format_node(
        Node::FunctionExpr(function),
    )?);

    Ok(())
}

use flux::ast;
use flux::parser::parse_string;
use flux::semantic::analyze::{analyze_with, Result};
use flux::semantic::env::Environment;
use flux::semantic::nodes;
use flux::semantic::nodes::{infer_pkg_types, inject_pkg_types};
use libstd;

fn main() {
    // TODO: read source from stdin
    let src = "";
    analyze_source(src).unwrap();
}

fn analyze_source(source: &str) -> Result<nodes::Package> {
    let file = parse_string("", source);
    let errs = ast::check::check(ast::walk::Node::File(&file));
    if !errs.is_empty() {
        return Err(format!("got errors on parsing: {:?}", errs));
    }
    let ast_pkg = ast::Package {
        base: file.base.clone(),
        path: "".to_string(),
        package: "main".to_string(),
        files: vec![file],
    };
    let mut fresher = libstd::fresher();
    let mut sem_pkg = analyze_with(ast_pkg, &mut fresher)?;
    let env = Environment::new(libstd::prelude().unwrap());
    let imports = libstd::imports().unwrap();

    let (_, sub) = infer_pkg_types(&mut sem_pkg, env, &mut fresher, &imports, &None)?;
    Ok(inject_pkg_types(sem_pkg, &sub))
}

use flux::ast;
use flux::parser::parse_string;
use flux::semantic::convert::convert_with;
use flux::semantic::convert::Result as ConversionResult;
use flux::semantic::env::Environment;
use flux::semantic::fresh::Fresher;
//use flux::semantic::import::Importer;
use flux::semantic::nodes;
use flux::semantic::nodes::{infer_pkg_types, inject_pkg_types};
use libstd;
use std::io::{self, Read};

fn main() -> io::Result<()> {
    let mut src = String::new();
    io::stdin().read_to_string(&mut src)?;
    convert_source(&src).unwrap();
    Ok(())
}

pub fn convert_source(source: &str) -> ConversionResult<nodes::Package> {
    let file = parse_string("", source);
    let errs = ast::check::check(ast::walk::Node::File(&file));
    if !errs.is_empty() {
        return Err(format!("Parsing Error: {:?}", errs));
    }
    let ast_pkg: ast::Package = file.into();
    let mut f = Fresher::default();
    let mut sem_pkg = convert_with(ast_pkg, &mut f)?;
    let env = Environment::new(libstd::prelude().unwrap());
    let imports = libstd::imports().unwrap();

    let (_, sub) = infer_pkg_types(&mut sem_pkg, env, &mut f, &imports, &None)?;
    Ok(inject_pkg_types(sem_pkg, &sub))
}

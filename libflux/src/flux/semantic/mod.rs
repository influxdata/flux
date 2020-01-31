pub mod convert;

mod import;

mod infer;
mod sub;

pub mod bootstrap;
pub mod env;
pub mod fresh;
pub mod nodes;
pub mod parser;
pub mod types;
pub mod walk;

#[cfg(test)]
mod tests;

#[allow(unused, non_snake_case)]
pub mod flatbuffers;

pub mod builtins;

use crate::ast;
use crate::parser::parse_string;
use crate::semantic::convert::convert_with;
use crate::semantic::convert::Result as ConversionResult;
use crate::semantic::env::Environment;
use crate::semantic::fresh::Fresher;
use crate::semantic::import::Importer;
use crate::semantic::nodes::{infer_pkg_types, inject_pkg_types};

impl Importer for Option<()> {}

pub fn convert_source(source: &str) -> ConversionResult<nodes::Package> {
    let file = parse_string("", source);
    let errs = ast::check::check(ast::walk::Node::File(&file));
    if !errs.is_empty() {
        return Err(format!("got errors on parsing: {:?}", errs));
    }
    let ast_pkg: ast::Package = file.into();
    let mut f = Fresher::default();
    let mut sem_pkg = convert_with(ast_pkg, &mut f)?;
    // TODO(affo): add a stdlib Importer.
    let (_, sub) = infer_pkg_types(&mut sem_pkg, Environment::empty(), &mut f, &None, &None)?;
    Ok(inject_pkg_types(sem_pkg, &sub))
}

//! Semantic analysis.

pub mod convert;

mod import;

mod infer;

use anyhow::{bail, Result};

#[macro_use]
pub mod types;
pub mod bootstrap;
pub mod check;
pub mod doc;
pub mod env;
pub mod fresh;
pub mod nodes;
pub mod sub;
pub mod walk;

#[cfg(test)]
mod tests;

#[allow(unused, non_snake_case)]
pub mod flatbuffers;

use crate::ast;
use crate::parser::parse_string;
use crate::semantic::convert::convert_with;
use crate::semantic::env::Environment;
use crate::semantic::fresh::Fresher;
use crate::semantic::import::Importer;

impl Importer for Option<()> {}

/// Get a semantic package from the given source and fresher
/// The returned semantic package is not type-inferred.
fn get_sem_pkg_from_source(source: &str, fresher: &mut Fresher) -> Result<nodes::Package> {
    let file = parse_string("", source);
    let errs = ast::check::check(ast::walk::Node::File(&file));
    if !errs.is_empty() {
        bail!("got errors on parsing: {:?}", errs);
    }
    let ast_pkg: ast::Package = file.into();
    convert_with(ast_pkg, fresher)
}

/// Get a type-inferred semantic package from the given Flux source.
pub fn convert_source(source: &str) -> Result<nodes::Package> {
    let mut f = Fresher::default();
    let mut sem_pkg = get_sem_pkg_from_source(source, &mut f)?;
    // TODO(affo): add a stdlib Importer.
    let (_, sub) = nodes::infer_pkg_types(&mut sem_pkg, Environment::empty(false), &mut f, &None)?;
    Ok(nodes::inject_pkg_types(sem_pkg, &sub))
}

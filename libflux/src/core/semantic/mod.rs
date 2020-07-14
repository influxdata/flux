#![allow(missing_docs)]
pub mod convert;

mod import;

mod infer;

#[macro_use]
pub mod types;
pub mod bootstrap;
pub mod check;
pub mod env;
pub mod fresh;
pub mod nodes;
pub mod parser;
pub mod sub;
pub mod walk;

#[cfg(test)]
mod tests;

#[allow(unused, non_snake_case)]
pub mod flatbuffers;

pub mod builtins;

use crate::ast;
use crate::parser::parse_string;
use crate::semantic::convert::convert_with;
use crate::semantic::env::Environment;
use crate::semantic::fresh::Fresher;
// This needs to be public so libstd can access it.
// Once we merge libstd and flux this can be made private again.
pub use crate::semantic::import::Importer;
use crate::semantic::types::{MonoType, PolyType, TvarKinds};
use std::fmt;

#[derive(Debug)]
pub struct Error {
    msg: String,
}

impl fmt::Display for Error {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "{}", self.msg)
    }
}

impl From<nodes::Error> for Error {
    fn from(err: nodes::Error) -> Error {
        Error {
            msg: err.to_string(),
        }
    }
}

impl From<String> for Error {
    fn from(msg: String) -> Error {
        Error { msg }
    }
}

impl Importer for Option<()> {}

/// Get a semantic package from the given source and fresher
/// The returned semantic package is not type-inferred.
fn get_sem_pkg_from_source(source: &str, fresher: &mut Fresher) -> Result<nodes::Package, Error> {
    let file = parse_string("", source);
    let errs = ast::check::check(ast::walk::Node::File(&file));
    if !errs.is_empty() {
        return Err(Error {
            msg: format!("got errors on parsing: {:?}", errs),
        });
    }
    let ast_pkg: ast::Package = file.into();
    convert_with(ast_pkg, fresher).map_err(|err| err.into())
}

/// Get a type-inferred semantic package from the given Flux source.
pub fn convert_source(source: &str) -> Result<nodes::Package, Error> {
    let mut f = Fresher::default();
    let mut sem_pkg = get_sem_pkg_from_source(source, &mut f)?;
    // TODO(affo): add a stdlib Importer.
    let (_, sub) = nodes::infer_pkg_types(
        &mut sem_pkg,
        Environment::empty(false),
        &mut f,
        &None,
        &None,
    )?;
    Ok(nodes::inject_pkg_types(sem_pkg, &sub))
}

/// Given a Flux source and a variable name, find out the type of that variable in the Flux source code.
/// A type variable will be automatically generated and injected into the type environment that
/// will be used in semantic analysis. The Flux source code itself should not contain any definition
/// for that variable.
pub fn find_var_type(source: &str, var_name: &str) -> Result<MonoType, Error> {
    let mut f = Fresher::default();
    let mut env = Environment::empty(true);
    // We generate and inject a type variable for the designated variable in the Flux source.
    env.add(
        var_name.to_string(),
        PolyType {
            vars: Vec::new(),
            cons: TvarKinds::new(),
            expr: MonoType::Var(f.fresh()),
        },
    );
    let mut sem_pkg = get_sem_pkg_from_source(source, &mut f)?;
    let (env, _) = nodes::infer_pkg_types(&mut sem_pkg, env, &mut f, &None, &None)?;

    Ok(env.lookup(var_name).unwrap().expr.clone())
}

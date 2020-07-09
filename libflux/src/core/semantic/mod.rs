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

pub fn convert_source(source: &str) -> Result<nodes::Package, Error> {
    let file = parse_string("", source);
    let errs = ast::check::check(ast::walk::Node::File(&file));
    if !errs.is_empty() {
        return Err(Error {
            msg: format!("got errors on parsing: {:?}", errs),
        });
    }
    let ast_pkg: ast::Package = file.into();
    let mut f = Fresher::default();
    let mut sem_pkg = convert_with(ast_pkg, &mut f)?;
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

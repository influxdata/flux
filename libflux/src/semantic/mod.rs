mod analyze;
pub use analyze::analyze;

mod env;
mod fresh;
mod infer;

pub mod nodes;

mod sub;
pub mod types;
pub mod walk;

#[cfg(test)]
mod parser;

#[cfg(test)]
mod tests;

use crate::ast;
use crate::parser::parse_string;
use crate::semantic::analyze::analyze_with;
use crate::semantic::analyze::Result as AnalysisResult;
use crate::semantic::analyze::SemanticError;
use crate::semantic::env::Environment;
use crate::semantic::fresh::Fresher;
use crate::semantic::nodes::Error as InferError;
use crate::semantic::nodes::{infer_pkg_types, inject_pkg_types, Importer};

impl From<InferError> for SemanticError {
    fn from(err: InferError) -> SemanticError {
        err.msg.clone()
    }
}

pub fn analyze_source(source: &str) -> AnalysisResult<nodes::Package> {
    let file = parse_string("", source);
    let errs = ast::check::check(ast::walk::Node::File(&file));
    if errs.len() > 0 {
        return Err(format!("got errors on parsing: {:?}", errs));
    }
    let ast_pkg = ast::Package {
        base: file.base.clone(),
        path: "".to_string(),
        package: "main".to_string(),
        files: vec![file],
    };
    let mut f = Fresher::new();
    let mut sem_pkg = analyze_with(ast_pkg, &mut f)?;
    // TODO(affo): add a stdlib Importer.
    let (_, sub) = infer_pkg_types(&mut sem_pkg, Environment::empty(), &mut f, &Importer::new())?;
    Ok(inject_pkg_types(sem_pkg, &sub))
}

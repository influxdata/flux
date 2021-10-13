//! Semantic analysis.

pub mod convert;

mod fs;
mod infer;

#[macro_use]
pub mod types;

pub mod bootstrap;
pub mod check;
pub mod doc;
pub mod env;
pub mod fresh;
pub mod import;
pub mod nodes;
pub mod sub;
pub mod walk;

#[cfg(test)]
mod tests;

#[allow(unused, non_snake_case)]
pub mod flatbuffers;

use crate::{ast, parser};

use thiserror::Error;

/// Error represents any any error that can occur during any step of the type analysis process.
///
/// Users of Flux do not care to understand the various steps involved with type analysis
/// as such these errors do not add any context to the error messages.
///
/// However users of this library may care and therefore can use this Error enum to determine where in
/// the process an error occurred.
#[derive(Error, Debug, PartialEq)]
pub enum Error {
    /// Errors that occur because of bad syntax or in valid AST
    #[error("{0}")]
    InvalidAST(#[from] ast::check::Errors),
    /// Errors that occur converting AST to semantic graph
    #[error("{0}")]
    Convert(#[from] convert::Error),
    /// Errors that occur because of bad semantic graph
    #[error("{0}")]
    InvalidSemantic(#[from] check::Error),
    /// Errors that occur because of incompatible/incomplete types
    #[error("{0}")]
    Inference(#[from] nodes::Error),
}

/// Analyzer provides an API for analyzing Flux code.
pub struct Analyzer<I: import::Importer> {
    env: env::Environment,
    importer: I,
}

impl<I: import::Importer> Analyzer<I> {
    /// Create an analyzer with the given environment and importer.
    /// The environment represents any values in scope.
    pub fn new(env: env::Environment, importer: I) -> Self {
        Analyzer { env, importer }
    }
    /// Analyze Flux source code returning the semantic package and the package environment.
    pub fn analyze_source(
        &mut self,
        pkgpath: String,
        file_name: String,
        src: &str,
    ) -> Result<(env::Environment, nodes::Package), Error> {
        let ast_file = parser::parse_string(file_name, src);
        let ast_pkg = ast::Package {
            base: ast_file.base.clone(),
            path: pkgpath,
            package: ast_file.get_package().to_string(),
            files: vec![ast_file],
        };
        self.analyze_ast(ast_pkg)
    }
    /// Analyze Flux AST returning the semantic package and the package environment.
    pub fn analyze_ast(
        &mut self,
        ast_pkg: ast::Package,
    ) -> Result<(env::Environment, nodes::Package), Error> {
        self.analyze_ast_with_fresher(ast_pkg, &mut fresh::Fresher::default())
    }
    /// Analyze Flux AST returning the semantic package and the package environment.
    /// A custom fresher may be provided.
    pub fn analyze_ast_with_fresher(
        &mut self,
        // TODO(nathanielc): Change analyze steps to not move the ast::Package as it is a readonly
        // operation.
        ast_pkg: ast::Package,
        fresher: &mut fresh::Fresher,
    ) -> Result<(env::Environment, nodes::Package), Error> {
        ast::check::check(ast::walk::Node::Package(&ast_pkg))?;

        let mut sem_pkg = convert::convert_package(ast_pkg, fresher)?;
        check::check(&sem_pkg)?;

        // Clone the environment as the inferred package may mutate it.
        let env = self.env.clone();
        let (env, sub) = nodes::infer_package(&mut sem_pkg, env, fresher, &mut self.importer)?;
        Ok((env, nodes::inject_pkg_types(sem_pkg, &sub)))
    }

    /// Drop returns ownership of the environment and importer.
    pub fn drop(self) -> (env::Environment, I) {
        (self.env, self.importer)
    }
}

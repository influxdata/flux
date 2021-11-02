//! Semantic analysis.

pub mod convert;

mod fs;
mod infer;

#[macro_use]
pub mod types;

pub mod bootstrap;
pub mod check;
pub mod env;
pub mod formatter;
pub mod fresh;
pub mod import;
pub mod nodes;
pub mod sub;
pub mod walk;

#[cfg(test)]
mod tests;

#[allow(unused, non_snake_case)]
pub mod flatbuffers;

use crate::{ast, errors::Errors, parser};

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
    InvalidAST(#[from] ast::check::Error),
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
    config: AnalyzerConfig,
}

/// A set of configuration options for the behavior of an Analyzer.
#[derive(Default)]
pub struct AnalyzerConfig {
    /// If true no AST or Semantic checks are performed.
    /// Default is false.
    pub skip_checks: bool,
}

impl<I: import::Importer> Analyzer<I> {
    /// Create an analyzer with the given environment and importer.
    /// The environment represents any values in scope.
    pub fn new(env: env::Environment, importer: I, config: AnalyzerConfig) -> Self {
        Analyzer {
            env,
            importer,
            config,
        }
    }
    /// Create an analyzer with the given environment and importer using default configuration.
    /// The environment represents any values in scope.
    pub fn new_with_defaults(env: env::Environment, importer: I) -> Self {
        Analyzer::new(env, importer, AnalyzerConfig::default())
    }
    /// Analyze Flux source code returning the semantic package and the package environment.
    pub fn analyze_source(
        &mut self,
        pkgpath: String,
        file_name: String,
        src: &str,
    ) -> Result<(env::Environment, nodes::Package), Errors<Error>> {
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
    ) -> Result<(env::Environment, nodes::Package), Errors<Error>> {
        self.analyze_ast_with_substitution(ast_pkg, &mut sub::Substitution::default())
    }
    /// Analyze Flux AST returning the semantic package and the package environment.
    /// A custom fresher may be provided.
    pub fn analyze_ast_with_substitution(
        &mut self,
        // TODO(nathanielc): Change analyze steps to not move the ast::Package as it is a readonly
        // operation.
        ast_pkg: ast::Package,
        sub: &mut sub::Substitution,
    ) -> Result<(env::Environment, nodes::Package), Errors<Error>> {
        let mut errors = Errors::new();
        if !self.config.skip_checks {
            if let Err(err) = ast::check::check(ast::walk::Node::Package(&ast_pkg)) {
                errors.extend(err.into_iter().map(Error::from));
            }
        }

        let mut sem_pkg = match convert::convert_package(ast_pkg, sub) {
            Ok(sem_pkg) => sem_pkg,
            Err(err) => {
                errors.push(err.into());
                return Err(errors);
            }
        };
        if !self.config.skip_checks {
            if let Err(err) = check::check(&sem_pkg) {
                errors.push(err.into());
            }
        }

        // Clone the environment as the inferred package may mutate it.
        let env = self.env.clone();
        let env = match nodes::infer_package(&mut sem_pkg, env, sub, &mut self.importer) {
            Ok(env) => env,
            Err(err) => {
                errors.extend(err.into_iter().map(Error::from));
                return Err(errors);
            }
        };
        if errors.has_errors() {
            return Err(errors);
        }
        Ok((env, nodes::inject_pkg_types(sem_pkg, sub)))
    }

    /// Drop returns ownership of the environment and importer.
    pub fn drop(self) -> (env::Environment, I) {
        (self.env, self.importer)
    }
}

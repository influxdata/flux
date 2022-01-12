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

use std::{fmt, ops::Range};

use codespan_reporting::{
    diagnostic,
    files::Files,
    term::{
        self,
        termcolor::{self, WriteColor},
    },
};
use thiserror::Error;

use crate::{
    ast,
    errors::{AsDiagnostic, Errors, Located},
    parser,
    semantic::{
        infer::Constraints,
        nodes::Symbol,
        sub::Substitution,
        types::{Label, MonoType, PolyType, PolyTypeHashMap, Property, Record},
    },
};

/// Result type for multiple semantic errors
pub type Result<T, E = FileErrors> = std::result::Result<T, E>;
/// Error represents any error that can occur during any step of the type analysis process.
pub type Error = Located<ErrorKind>;

/// ErrorKind exposes details about where in the type analysis process an error occurred.
///
/// Users of Flux do not care to understand the various steps involved with type analysis
/// as such these errors do not add any context to the error messages.
///
/// However users of this library may care and therefore can use this enum to determine where in
/// the process an error occurred.
#[derive(Error, Debug, PartialEq)]
pub enum ErrorKind {
    /// Errors that occur because of bad syntax or in valid AST
    #[error("{0}")]
    InvalidAST(ast::check::ErrorKind),
    /// Errors that occur converting AST to semantic graph
    #[error("{0}")]
    Convert(convert::ErrorKind),
    /// Errors that occur because of bad semantic graph
    #[error("{0}")]
    InvalidSemantic(check::ErrorKind),
    /// Errors that occur because of incompatible/incomplete types
    #[error("{0}")]
    Inference(nodes::ErrorKind),
}
impl From<ast::check::Error> for Error {
    fn from(error: ast::check::Error) -> Self {
        Self {
            location: error.location,
            error: ErrorKind::InvalidAST(error.error),
        }
    }
}
impl From<convert::Error> for Error {
    fn from(error: convert::Error) -> Self {
        Self {
            location: error.location,
            error: ErrorKind::Convert(error.error),
        }
    }
}
impl From<check::Error> for Error {
    fn from(error: check::Error) -> Self {
        Self {
            location: error.location,
            error: ErrorKind::InvalidSemantic(error.error),
        }
    }
}
impl From<nodes::Error> for Error {
    fn from(error: nodes::Error) -> Self {
        Self {
            location: error.location,
            error: ErrorKind::Inference(error.error),
        }
    }
}
impl From<Errors<nodes::Error>> for Errors<Error> {
    fn from(error: Errors<nodes::Error>) -> Self {
        error.into_iter().map(Error::from).collect()
    }
}

/// An environment of values that are available outside of a package
#[derive(Debug, Clone, PartialEq)]
pub struct PackageExports {
    /// Values in the environment.
    values: types::SemanticMap<String, (Symbol, PolyType)>,

    /// The type representing this package
    typ: PolyType,
}

impl Default for PackageExports {
    fn default() -> Self {
        PackageExports {
            typ: PolyType {
                vars: Default::default(),
                cons: Default::default(),
                expr: MonoType::from(Record::Empty),
            },
            values: Default::default(),
        }
    }
}

impl TryFrom<PolyTypeHashMap<Symbol>> for PackageExports {
    type Error = Errors<Error>;
    fn try_from(values: PolyTypeHashMap<Symbol>) -> Result<Self, Errors<Error>> {
        Ok(PackageExports {
            typ: build_polytype(values.iter().map(|(k, v)| (k.clone(), v.clone())))?,
            values: values
                .into_iter()
                .map(|(symbol, typ)| (symbol.to_string(), (symbol, typ)))
                .collect(),
        })
    }
}

impl TryFrom<Vec<(Symbol, PolyType)>> for PackageExports {
    type Error = Errors<Error>;
    fn try_from(values: Vec<(Symbol, PolyType)>) -> Result<Self, Errors<Error>> {
        Ok(PackageExports {
            typ: build_polytype(values.iter().map(|(k, v)| (k.clone(), v.clone())))?,
            values: values
                .into_iter()
                .map(|(symbol, typ)| (symbol.to_string(), (symbol, typ)))
                .collect(),
        })
    }
}

impl PackageExports {
    /// Returns an empty environment
    pub fn new() -> Self {
        Self::default()
    }

    /// Returns the type representing this package
    pub fn typ(&self) -> PolyType {
        self.typ.clone()
    }

    /// Add a new variable binding to the current stack frame.
    pub fn add(&mut self, name: Symbol, t: PolyType) {
        self.values.insert(name.to_string(), (name, t));
        self.typ = build_polytype(self.values.values().cloned()).unwrap();
    }

    /// Check whether a `PolyType` `k` given by a
    /// string identifier is in the environment.
    pub fn lookup(&self, k: &str) -> Option<&PolyType> {
        self.values.get(k).map(|(_, typ)| typ)
    }

    /// Check whether a `Symbol` `k` identifier is in the environment.
    pub fn lookup_symbol(&self, k: &str) -> Option<&Symbol> {
        self.values.get(k).map(|(symbol, _)| symbol)
    }

    /// Copy all the variable bindings from another [`ExportEnvironment`] to the current environment.
    /// This does not change the current environment's `parent` or `readwrite` flag.
    pub fn copy_bindings_from(&mut self, other: &Self) {
        for (name, t) in other.values.iter() {
            self.values.insert(name.clone(), t.clone());
        }
        self.typ = build_polytype(self.values.values().cloned()).unwrap();
    }

    /// Returns an iterator over all values
    pub fn iter(&self) -> impl Iterator<Item = (&str, &PolyType)> + '_ {
        self.values.iter().map(|(k, (_, v))| (k.as_str(), v))
    }

    /// Returns how many values exist in the environment
    pub fn len(&self) -> usize {
        self.values.len()
    }

    /// Returns `true` if the environment contains no elements.
    pub fn is_empty(&self) -> bool {
        self.values.is_empty()
    }

    /// Returns an iterator over exported bindings in this package
    pub fn into_bindings(self) -> impl Iterator<Item = (Symbol, PolyType)> {
        self.values.into_iter().map(|(_, v)| v)
    }
}

/// Constructs a polytype, or more specifically a generic record type, from a hash map.
pub fn build_polytype(
    from: impl IntoIterator<Item = (Symbol, PolyType)>,
) -> Result<PolyType, Errors<Error>> {
    let mut sub = Substitution::default();
    let (r, cons) = build_record(from, &mut sub);
    infer::solve(&cons, &mut sub).map_err(Errors::<nodes::Error>::from)?;
    let typ = MonoType::record(r);
    Ok(infer::generalize(
        &env::Environment::empty(false),
        sub.cons(),
        typ,
    ))
}

fn build_record(
    from: impl IntoIterator<Item = (Symbol, PolyType)>,
    sub: &mut Substitution,
) -> (Record, Constraints) {
    let mut r = Record::Empty;
    let mut cons = Constraints::empty();

    for (name, poly) in from {
        let (ty, constraints) = infer::instantiate(
            poly.clone(),
            sub,
            ast::SourceLocation {
                file: None,
                start: ast::Position::default(),
                end: ast::Position::default(),
                source: None,
            },
        );
        r = Record::Extension {
            head: Property {
                k: Label::from(name.clone()),
                v: ty,
            },
            tail: MonoType::record(r),
        };
        cons += constraints;
    }
    (r, cons)
}

/// Error represents any any error that can occur during any step of the type analysis process.
#[derive(Error, Debug, PartialEq)]
pub struct FileErrors {
    /// The file that the errors occured in
    pub file: String,
    /// The source for this error, if one exists
    // TODO Change the API such that we must always provide the source?
    pub source: Option<String>,
    #[source]
    /// The errors the occurred in that file
    pub errors: Errors<Located<ErrorKind>>,
}

impl fmt::Display for FileErrors {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        // TODO Use codespan's formatting for errors
        self.errors.fmt(f)
    }
}

pub(crate) trait Source {
    fn codespan_range(&self, location: &ast::SourceLocation) -> Range<usize>;
}

impl Source for codespan_reporting::files::SimpleFile<&str, &str> {
    fn codespan_range(&self, location: &ast::SourceLocation) -> Range<usize> {
        (|| {
            let start = self
                .line_range((), location.start.line as usize - 1)
                .ok()?
                .start;
            let end = self
                .line_range((), location.end.line as usize - 1)
                .ok()?
                .start;
            Some(start + location.start.column as usize - 1..end + location.end.column as usize - 1)
        })()
        .unwrap_or_default()
    }
}

impl FileErrors {
    /// Prints the errors
    pub fn pretty(&self, source: &str) -> String {
        self.pretty_config(&term::Config::default(), source)
    }

    /// Prints the errors in their short form
    pub fn pretty_short(&self, source: &str) -> String {
        self.pretty_config(
            &term::Config {
                display_style: term::DisplayStyle::Short,
                ..term::Config::default()
            },
            source,
        )
    }

    /// Prints the errors to stdout
    pub fn print(&self) {
        match &self.source {
            Some(source) => {
                let mut stdout = termcolor::StandardStream::stdout(termcolor::ColorChoice::Auto);
                // Mirror println! by ignoring errors
                let _ = self.print_config(&term::Config::default(), source, &mut stdout);
            }
            None => println!("{}", self.errors),
        }
    }

    /// Prints the errors to a `String`
    pub fn pretty_config(&self, config: &term::Config, source: &str) -> String {
        let mut buffer = termcolor::Buffer::no_color();
        self.print_config(config, source, &mut buffer)
            .expect("Writing to a termcolor::Buffer can't fail");
        String::from_utf8(buffer.into_inner())
            .expect("We only write utf-8 when we don't use coloring")
    }

    fn print_config(
        &self,
        config: &term::Config,
        source: &str,
        writer: &mut dyn WriteColor,
    ) -> Result<(), codespan_reporting::files::Error> {
        let files = codespan_reporting::files::SimpleFile::new(&self.file[..], source);
        for err in &self.errors {
            err.pretty_fmt(config, &files, writer)?;
        }
        Ok(())
    }
}

impl Error {
    fn pretty_fmt(
        &self,
        config: &term::Config,
        files: &codespan_reporting::files::SimpleFile<&str, &str>,
        writer: &mut dyn WriteColor,
    ) -> Result<(), codespan_reporting::files::Error> {
        let diagnostic = self.as_diagnostic(files);

        term::emit(writer, config, files, &diagnostic)?;
        Ok(())
    }
}

impl AsDiagnostic for ErrorKind {
    fn as_diagnostic(&self, source: &dyn Source) -> diagnostic::Diagnostic<()> {
        match self {
            Self::InvalidAST(err) => err.as_diagnostic(source),
            Self::Convert(err) => err.as_diagnostic(source),
            Self::InvalidSemantic(err) => err.as_diagnostic(source),
            Self::Inference(err) => err.as_diagnostic(source),
        }
    }
}

/// Analyzer provides an API for analyzing Flux code.
#[derive(Default)]
pub struct Analyzer<'env, I: import::Importer> {
    env: env::Environment<'env>,
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

impl<'env, I: import::Importer> Analyzer<'env, I> {
    /// Create an analyzer with the given environment and importer.
    /// The environment represents any values in scope.
    pub fn new(env: env::Environment<'env>, importer: I, config: AnalyzerConfig) -> Self {
        Analyzer {
            env,
            importer,
            config,
        }
    }
    /// Create an analyzer with the given environment and importer using default configuration.
    /// The environment represents any values in scope.
    pub fn new_with_defaults(env: env::Environment<'env>, importer: I) -> Self {
        Analyzer::new(env, importer, AnalyzerConfig::default())
    }

    /// Analyze Flux source code returning the semantic package and the package environment.
    pub fn analyze_source(
        &mut self,
        pkgpath: String,
        file_name: String,
        src: &str,
    ) -> Result<(PackageExports, nodes::Package), FileErrors> {
        let ast_file = parser::parse_string(file_name, src);
        let ast_pkg = ast::Package {
            base: ast_file.base.clone(),
            path: pkgpath,
            package: ast_file.get_package().to_string(),
            files: vec![ast_file],
        };
        self.analyze_ast(ast_pkg).map_err(|mut err| {
            err.source = Some(src.into());
            err
        })
    }

    /// Analyze Flux AST returning the semantic package and the package environment.
    pub fn analyze_ast(
        &mut self,
        ast_pkg: ast::Package,
    ) -> Result<(PackageExports, nodes::Package), FileErrors> {
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
    ) -> Result<(PackageExports, nodes::Package), FileErrors> {
        let mut errors = Errors::new();
        if !self.config.skip_checks {
            if let Err(err) = ast::check::check(ast::walk::Node::Package(&ast_pkg)) {
                errors.extend(err.into_iter().map(Error::from));
            }
        }

        let file = ast_pkg.package.clone();
        let mut sem_pkg = match convert::convert_package(ast_pkg, &self.env, sub) {
            Ok(sem_pkg) => sem_pkg,
            Err(err) => {
                errors.extend(err.into_iter().map(Error::from));
                return Err(FileErrors {
                    file,
                    source: None,
                    errors,
                });
            }
        };
        if !self.config.skip_checks {
            if let Err(err) = check::check(&sem_pkg) {
                errors.push(err.into());
            }
        }

        self.env.enter_scope();
        let env = match nodes::infer_package(&mut sem_pkg, &mut self.env, sub, &mut self.importer) {
            Ok(()) => {
                let env = self.env.exit_scope();
                match PackageExports::try_from(env.values) {
                    Ok(env) => env,
                    Err(err) => {
                        errors.extend(err);
                        return Err(FileErrors {
                            file: sem_pkg.package,
                            source: None,
                            errors,
                        });
                    }
                }
            }
            Err(err) => {
                self.env.exit_scope();
                errors.extend(err.into_iter().map(Error::from));
                return Err(FileErrors {
                    file: sem_pkg.package,
                    source: None,
                    errors,
                });
            }
        };
        if errors.has_errors() {
            return Err(FileErrors {
                file: sem_pkg.package,
                source: None,
                errors,
            });
        }

        Ok((env, nodes::inject_pkg_types(sem_pkg, sub)))
    }

    /// Drop returns ownership of the environment and importer.
    pub fn drop(self) -> (env::Environment<'env>, I) {
        (self.env, self.importer)
    }
}

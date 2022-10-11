use crate::{
    errors::{located, Errors, SalvageResult},
    parser,
    semantic::{
        convert::Symbol,
        env::Environment,
        import::{Importer, Packages},
        nodes,
        types::PolyType,
        Analyzer, AnalyzerConfig, FileErrors, PackageExports,
    },
};

use super::*;

use std::{
    collections::{HashMap, HashSet},
    io,
    path::PathBuf,
    sync::{Arc, Mutex},
};

use thiserror::Error;

const INTERNAL_PRELUDE: [&str; 2] = ["internal/boolean", "internal/location"];

pub type Result<T, E = Error> = std::result::Result<T, E>;

#[derive(Error, Clone, Debug, Eq, PartialEq)]
pub enum Error {
    #[error("{0}")]
    FileError(#[from] Arc<FileErrors>),

    #[error("{0}")]
    Message(String),
}

/// Base trait for the flux database
pub trait FluxBase {
    #[doc(hidden)]
    fn clear_error(&self, package: &str);

    #[doc(hidden)]
    fn record_error(&self, package: String, error: Error);

    /// Returns the errors for all compiled packages
    fn package_errors(&self) -> Errors<Error>;

    /// Returns the file names that are part of `package`
    fn package_files(&self, package: &str) -> Result<Vec<String>>;

    /// Sets the source code for a file at `path`. (Alternative to loading the files dynamically
    /// from disk).
    fn set_source(&mut self, path: String, source: Arc<str>);

    /// Returns the source code for `path`, returning an error if it does not exist
    fn source(&self, path: String) -> Result<Arc<str>>;
}

/// Defines queries that drives flux compilation
#[salsa::query_group(FluxStorage)]
pub trait Flux: FluxBase {
    /// Source code for a particular flux file
    #[salsa::input]
    #[doc(hidden)]
    // Input queries generates both `<QUERY>` and set_<QUERY>` methods which can be called to set and later retrieve
    // values
    fn source_inner(&self, file_path: String) -> Arc<str>;

    /// Sets the AnalyzerConfig for the compilation
    #[salsa::input]
    fn analyzer_config(&self) -> AnalyzerConfig;

    /// Enables the prelude for all compiled packages
    ///
    /// Default: true
    #[salsa::input]
    fn use_prelude(&self) -> bool;

    /// Sets any precompiled packages that should be included in the compilation
    #[salsa::input]
    fn precompiled_packages(&self) -> Option<&'static Packages>;

    /// Returns the `ast::Package` for a given module path
    // Normal `dependency` query that may call recursively into other queries. If the recursive
    // queries change their outpot then this will be forced to run again, otherwise we always
    // return the cached value (hence the `Arc`, so we can clone it easily)
    fn ast_package(&self, package_path: String) -> Result<Arc<ast::Package>>;

    #[doc(hidden)]
    fn internal_prelude(&self) -> Result<Arc<PackageExports>>;

    /// Returns the `PackageExports` for the prelude
    fn prelude(&self) -> Result<Arc<PackageExports>>;

    /// Returns the `semantic::Package`
    // We need to query for the semantic package when compiling `import`s so it is possible for
    // users to write cycles. `salsa::cycle` adds a handler which tells salsa how to recover
    // (by default it assumes it is a bug and panics)
    #[salsa::cycle(recover_cycle2)]
    fn semantic_package(
        &self,
        package_path: String,
    ) -> SalvageResult<(Arc<PackageExports>, Arc<nodes::Package>), Error>;

    /// Returns the `PackageExports` for a given package path. Will consuled `precompiled_packages`
    /// if it is set.
    // Transparent queries are just plain functions, no special behavior
    #[salsa::transparent]
    fn package_exports(&self, package_path: String) -> SalvageResult<Arc<PackageExports>, Error>;

    // Wrapper around `semantic_package` which is called when resolving imports. Only returns
    // `PackageExports` since it also checks any precompiled data (if it exists).
    #[doc(hidden)]
    #[salsa::cycle(recover_cycle)]
    fn package_exports_import(
        &self,
        package_path: String,
    ) -> Result<Arc<PackageExports>, nodes::ErrorKind>;
}

/// Builder that configures a flux compiler database
#[derive(Default)]
pub struct DatabaseBuilder {
    filesystem_roots: Vec<PathBuf>,
}

impl DatabaseBuilder {
    /// Creates a new builder with the default values
    pub fn new() -> Self {
        Self::default()
    }

    /// Enables loading `.flux` files from `filesystem_roots`
    pub fn filesystem_roots(mut self, filesystem_root: Vec<PathBuf>) -> Self {
        self.filesystem_roots = filesystem_root;
        self
    }

    /// Builds the flux compiler database
    pub fn build(self) -> Database {
        let mut db = Database::default();

        db.filesystem_roots = self.filesystem_roots;

        db
    }
}

/// Storage for flux programs and their intermediates
#[salsa::database(FluxStorage)]
pub struct Database {
    storage: salsa::Storage<Self>,
    pub(crate) packages: Mutex<HashSet<String>>,
    package_errors: Mutex<HashMap<String, Error>>,
    filesystem_roots: Vec<PathBuf>,
}

impl Default for Database {
    fn default() -> Self {
        let mut db = Self {
            storage: Default::default(),
            packages: Default::default(),
            package_errors: Default::default(),
            filesystem_roots: Vec::new(),
        };
        db.set_analyzer_config(AnalyzerConfig::default());
        db.set_use_prelude(true);
        db.set_precompiled_packages(None);
        db
    }
}

impl salsa::Database for Database {}

impl FluxBase for Database {
    fn package_files(&self, package: &str) -> Result<Vec<String>> {
        let mut found_files = self.search_flux_files(package)?;

        let packages = self.packages.lock().unwrap();

        found_files.extend(
            packages
                .iter()
                .filter(|p| {
                    // Example: package: `internal/boolean` matches the file
                    // `internal/boolean/XXX.flux`
                    p.starts_with(package)
                        && p[package.len()..].starts_with('/')
                        && p[package.len() + 1..].split('/').count() == 1
                })
                .cloned(),
        );

        Ok(found_files)
    }

    fn clear_error(&self, package: &str) {
        self.package_errors.lock().unwrap().remove(package);
    }

    fn record_error(&self, package: String, error: Error) {
        self.package_errors.lock().unwrap().insert(package, error);
    }

    fn package_errors(&self) -> Errors<Error> {
        self.package_errors
            .lock()
            .unwrap()
            .values()
            .cloned()
            .collect::<Errors<_>>()
    }

    fn source(&self, path: String) -> Result<Arc<str>> {
        if !self.filesystem_roots.is_empty() {
            for filesystem_root in &self.filesystem_roots {
                let source = match std::fs::read_to_string(filesystem_root.join(&path)) {
                    Ok(source) => source,
                    Err(err) if err.kind() == io::ErrorKind::NotFound => continue,
                    Err(err) => {
                        return Err(Error::Message(format!(
                            "Unable to read `{}`: {}",
                            path, err
                        )))
                    }
                };
                self.packages.lock().unwrap().insert(path.clone());
                return Ok(Arc::from(source));
            }
        }
        Ok(self.source_inner(path))
    }

    fn set_source(&mut self, path: String, source: Arc<str>) {
        self.packages.lock().unwrap().insert(path.clone());

        self.set_source_inner(path, source)
    }
}

impl Database {
    fn search_flux_files(&self, package: &str) -> Result<Vec<String>> {
        let mut found_files = Vec::new();

        if !self.filesystem_roots.is_empty() {
            for filesystem_root in &self.filesystem_roots {
                let package_root = filesystem_root.join(package);
                for entry in std::fs::read_dir(&package_root).map_err(|err| {
                    Error::Message(format!("Unable to read directory `{}`: {}", package, err))
                })? {
                    let path = entry
                        .map_err(|err| {
                            Error::Message(format!("Unable to read path `{}`: {}", package, err))
                        })?
                        .path();
                    let path = path.strip_prefix(&filesystem_root).map_err(|err| {
                        Error::Message(format!(
                            "Unable to strip prefix `{}` of `{}`: {}",
                            filesystem_root.display(),
                            path.display(),
                            err
                        ))
                    })?;

                    if path.extension().and_then(|e| e.to_str()) == Some("flux")
                        && path
                            .file_stem()
                            .and_then(|f| f.to_str())
                            .map_or(true, |f| !f.ends_with("_test"))
                    {
                        let path = path.to_str().ok_or_else(|| {
                            Error::Message(format!("Invalid UTF-8 in path: {:?}", path))
                        })?;
                        found_files.push(path.to_string());
                    }
                }
            }
        }

        // It is possible that we find the same file twice if the roots contain duplicates
        found_files.sort();
        found_files.dedup();

        Ok(found_files)
    }
}

fn ast_package(db: &dyn Flux, path: String) -> Result<Arc<ast::Package>> {
    let files = db
        .package_files(&path)?
        .into_iter()
        .map(|file_path| {
            let source = db.source(file_path.clone())?;

            Ok(parser::parse_string(file_path, &source))
        })
        .collect::<Result<Vec<_>>>()?;

    if files.is_empty() {
        Err(Error::Message(format!(
            "No files exist for package `{}`",
            path
        )))
    } else {
        Ok(Arc::new(ast::Package {
            base: ast::BaseNode::default(),
            path,
            package: String::from(files[0].get_package()),
            files,
        }))
    }
}

fn internal_prelude(db: &dyn Flux) -> Result<Arc<PackageExports>> {
    let mut prelude_map = PackageExports::new();
    for name in INTERNAL_PRELUDE {
        // Infer each package in the prelude allowing the earlier packages to be used by later
        // packages within the prelude list.
        let (types, _sem_pkg) = db.semantic_package(name.into()).map_err(|err| err.error)?;

        prelude_map.copy_bindings_from(&types);
    }
    Ok(Arc::new(prelude_map))
}

fn prelude(db: &dyn Flux) -> Result<Arc<PackageExports>> {
    let mut prelude_map = PackageExports::new();
    for name in crate::semantic::bootstrap::PRELUDE {
        // Infer each package in the prelude allowing the earlier packages to be used by later
        // packages within the prelude list.
        let (types, _sem_pkg) = db.semantic_package(name.into()).map_err(|err| err.error)?;

        prelude_map.copy_bindings_from(&types);
    }
    Ok(Arc::new(prelude_map))
}

fn semantic_package(
    db: &dyn Flux,
    path: String,
) -> SalvageResult<(Arc<PackageExports>, Arc<nodes::Package>), Error> {
    // The previous standard library compiler happened to result in the prelude being incrementally
    // added to with later packages in the prelude depending on earlier ones. This was mostly
    // arbitrary and we should try to encode these dependencies more deliberately but these stages
    // of no/internal/full prelude seem to do the trick in getting things to work.
    let prelude = if !db.use_prelude() || INTERNAL_PRELUDE.contains(&&path[..]) {
        Default::default()
    } else if [
        "system",
        "date",
        "math",
        "strings",
        "regexp",
        "experimental/table",
    ]
    .contains(&&path[..])
        || crate::semantic::bootstrap::PRELUDE.contains(&&path[..])
    {
        db.internal_prelude()?
    } else {
        db.prelude()?
    };

    semantic_package_with_prelude(db, path, &prelude)
}

fn semantic_package_with_prelude(
    db: &dyn Flux,
    path: String,
    prelude: &PackageExports,
) -> SalvageResult<(Arc<PackageExports>, Arc<nodes::Package>), Error> {
    let file = db.ast_package(path.clone())?;

    let env = Environment::new(prelude.into());
    let mut importer = &*db;
    let mut analyzer = Analyzer::new(env, &mut importer, db.analyzer_config());
    let (exports, sem_pkg) = analyzer.analyze_ast(&file).map_err(|err| {
        err.map(|(exports, sem_pkg)| (Arc::new(exports), Arc::new(sem_pkg)))
            .map_err(Arc::new)
            .map_err(Error::from)
    })?;

    Ok((Arc::new(exports), Arc::new(sem_pkg)))
}

fn package_exports(db: &dyn Flux, path: String) -> SalvageResult<Arc<PackageExports>, Error> {
    if let Some(packages) = db.precompiled_packages() {
        if let Some(exports) = packages.get(&path) {
            return Ok(exports.clone());
        }
    }

    let (exports, _) = db
        .semantic_package(path)
        .map_err(|err| err.map(|(exports, _)| exports))?;
    Ok(exports)
}

fn package_exports_import(
    db: &dyn Flux,
    path: String,
) -> Result<Arc<PackageExports>, nodes::ErrorKind> {
    db.package_exports(path.clone())
        .map(|exports| {
            db.clear_error(&path);
            exports
        })
        .map_err(|err| {
            db.record_error(path.clone(), err.error);
            nodes::ErrorKind::InvalidImportPath(path)
        })
}

fn recover_cycle2<T>(db: &dyn Flux, cycle: &[String], name: &str) -> SalvageResult<T, Error> {
    let mut cycle: Vec<_> = cycle
        .iter()
        .filter(|k| k.starts_with("package_exports_import("))
        .map(|k| {
            k.trim_matches(|c: char| c != '"')
                .trim_matches('"')
                .trim_start_matches('@')
                .to_string()
        })
        .collect();
    cycle.pop();

    Err(Error::FileError(Arc::new(FileErrors {
        file: name.to_owned(),
        source: None,
        diagnostics: From::from(located(
            Default::default(),
            semantic::ErrorKind::Inference(nodes::ErrorKind::ImportCycle {
                package: name.into(),
                cycle,
            }),
        )),
        pretty_fmt: db
            .analyzer_config()
            .features
            .contains(&semantic::Feature::PrettyError),
    }))
    .into())
}

fn recover_cycle<T>(_db: &dyn Flux, cycle: &[String], name: &str) -> Result<T, nodes::ErrorKind> {
    // We get a list of strings like "semantic_package(\"b\")",
    let mut cycle: Vec<_> = cycle
        .iter()
        .filter(|k| k.starts_with("semantic_package("))
        .map(|k| {
            k.trim_matches(|c: char| c != '"')
                .trim_matches('"')
                .to_string()
        })
        .collect();
    cycle.pop();

    Err(nodes::ErrorKind::ImportCycle {
        package: name.into(),
        cycle,
    })
}

impl Importer for Database {
    fn import(&mut self, path: &str) -> Result<PolyType, nodes::ErrorKind> {
        self.package_exports_import(path.into())
            .map(|exports| exports.typ())
    }
    fn symbol(&mut self, path: &str, symbol_name: &str) -> Option<Symbol> {
        self.package_exports_import(path.into())
            .ok()
            .and_then(|exports| exports.lookup_symbol(symbol_name).cloned())
    }
}

impl Importer for &dyn Flux {
    fn import(&mut self, path: &str) -> Result<PolyType, nodes::ErrorKind> {
        self.package_exports_import(path.into())
            .map(|exports| exports.typ())
    }
    fn symbol(&mut self, path: &str, symbol_name: &str) -> Option<Symbol> {
        self.package_exports_import(path.into())
            .ok()
            .and_then(|exports| exports.lookup_symbol(symbol_name).cloned())
    }
}

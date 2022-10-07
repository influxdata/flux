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
    sync::{Arc, Mutex},
};

const INTERNAL_PRELUDE: [&str; 2] = ["internal/boolean", "internal/location"];

pub trait FluxBase {
    fn has_package(&self, package: &str) -> bool;
    fn clear_error(&self, package: &str);
    fn record_error(&self, package: String, error: Arc<FileErrors>);
    fn package_errors(&self) -> Errors<Arc<FileErrors>>;
    fn package_files(&self, package: &str) -> Vec<String>;
    fn set_source(&mut self, path: String, source: Arc<str>);
    fn source(&self, path: String) -> Arc<str>;
}

/// Defines queries that drives flux compilation
#[salsa::query_group(FluxStorage)]
pub trait Flux: FluxBase {
    /// Source code for a particular flux file
    #[salsa::input]
    #[doc(hidden)]
    fn source_inner(&self, path: String) -> Arc<str>;

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

    #[doc(hidden)]
    fn ast_package_inner(&self, path: String) -> Arc<ast::Package>;

    /// Returns the `ast::Package` for a given module path
    #[salsa::transparent]
    fn ast_package(&self, path: String) -> Option<Arc<ast::Package>>;

    #[doc(hidden)]
    fn internal_prelude(&self) -> Result<Arc<PackageExports>, Arc<FileErrors>>;

    #[doc(hidden)]
    fn prelude_inner(&self) -> Result<Arc<PackageExports>, Arc<FileErrors>>;

    /// Returns the `PackageExports` for the prelude
    #[salsa::transparent]
    fn prelude(&self) -> Result<Arc<PackageExports>, Arc<FileErrors>>;

    /// Returns the `semantic::Package`
    #[salsa::cycle(recover_cycle2)]
    fn semantic_package(
        &self,
        path: String,
    ) -> SalvageResult<(Arc<PackageExports>, Arc<nodes::Package>), Arc<FileErrors>>;

    /// Returns the `PackageExports` for a given package path. Will consuled `precompiled_packages`
    /// if it is set.
    #[salsa::transparent]
    fn package_exports(&self, path: String) -> SalvageResult<Arc<PackageExports>, Arc<FileErrors>>;

    #[doc(hidden)]
    #[salsa::cycle(recover_cycle)]
    fn semantic_package_cycle(&self, path: String)
        -> Result<Arc<PackageExports>, nodes::ErrorKind>;
}

/// Storage for flux programs and their intermediates
#[salsa::database(FluxStorage)]
pub struct Database {
    storage: salsa::Storage<Self>,
    pub(crate) packages: Mutex<HashSet<String>>,
    package_errors: Mutex<HashMap<String, Arc<FileErrors>>>,
}

impl Default for Database {
    fn default() -> Self {
        let mut db = Self {
            storage: Default::default(),
            packages: Default::default(),
            package_errors: Default::default(),
        };
        db.set_analyzer_config(AnalyzerConfig::default());
        db.set_use_prelude(true);
        db.set_precompiled_packages(None);
        db
    }
}

impl salsa::Database for Database {}

impl FluxBase for Database {
    fn has_package(&self, package: &str) -> bool {
        self.packages.lock().unwrap().contains(package)
    }

    fn package_files(&self, package: &str) -> Vec<String> {
        let packages = self.packages.lock().unwrap();
        let found_packages = packages
            .iter()
            .filter(|p| {
                // Example: package: `internal/boolean` matches the file
                // `internal/boolean/XXX.flux`
                p.starts_with(package)
                    && p[package.len()..].starts_with('/')
                    && p[package.len() + 1..].split('/').count() == 1
            })
            .cloned()
            .collect::<Vec<_>>();

        assert!(
            !packages.is_empty(),
            "Did not find any package files for `{}`",
            package,
        );

        found_packages
    }

    fn clear_error(&self, package: &str) {
        self.package_errors.lock().unwrap().remove(package);
    }

    fn record_error(&self, package: String, error: Arc<FileErrors>) {
        self.package_errors.lock().unwrap().insert(package, error);
    }

    fn package_errors(&self) -> Errors<Arc<FileErrors>> {
        self.package_errors
            .lock()
            .unwrap()
            .values()
            .cloned()
            .collect::<Errors<_>>()
    }

    fn source(&self, path: String) -> Arc<str> {
        self.source_inner(path)
    }

    fn set_source(&mut self, path: String, source: Arc<str>) {
        self.packages.lock().unwrap().insert(path.clone());

        self.set_source_inner(path, source)
    }
}

fn ast_package_inner(db: &dyn Flux, path: String) -> Arc<ast::Package> {
    let files = db
        .package_files(&path)
        .into_iter()
        .map(|file_path| {
            let source = db.source(file_path.clone());

            parser::parse_string(file_path, &source)
        })
        .collect::<Vec<_>>();

    Arc::new(ast::Package {
        base: ast::BaseNode::default(),
        path,
        package: String::from(files[0].get_package()),
        files,
    })
}

fn ast_package(db: &dyn Flux, path: String) -> Option<Arc<ast::Package>> {
    if db.has_package(&path) {
        Some(db.ast_package_inner(path))
    } else {
        None
    }
}

fn internal_prelude(db: &dyn Flux) -> Result<Arc<PackageExports>, Arc<FileErrors>> {
    let mut prelude_map = PackageExports::new();
    for name in INTERNAL_PRELUDE {
        // Infer each package in the prelude allowing the earlier packages to be used by later
        // packages within the prelude list.
        let (types, _sem_pkg) = db.semantic_package(name.into()).map_err(|err| err.error)?;

        prelude_map.copy_bindings_from(&types);
    }
    Ok(Arc::new(prelude_map))
}

fn prelude_inner(db: &dyn Flux) -> Result<Arc<PackageExports>, Arc<FileErrors>> {
    let mut prelude_map = PackageExports::new();
    for name in crate::semantic::bootstrap::PRELUDE {
        // Infer each package in the prelude allowing the earlier packages to be used by later
        // packages within the prelude list.
        let (types, _sem_pkg) = db.semantic_package(name.into()).map_err(|err| err.error)?;

        prelude_map.copy_bindings_from(&types);
    }
    Ok(Arc::new(prelude_map))
}

fn prelude(db: &dyn Flux) -> Result<Arc<PackageExports>, Arc<FileErrors>> {
    db.prelude_inner()
}

fn semantic_package(
    db: &dyn Flux,
    path: String,
) -> SalvageResult<(Arc<PackageExports>, Arc<nodes::Package>), Arc<FileErrors>> {
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
) -> SalvageResult<(Arc<PackageExports>, Arc<nodes::Package>), Arc<FileErrors>> {
    let file = db.ast_package_inner(path);

    let env = Environment::new(prelude.into());
    let mut importer = &*db;
    let mut analyzer = Analyzer::new(env, &mut importer, db.analyzer_config());
    let (exports, sem_pkg) = analyzer.analyze_ast(&file).map_err(|err| {
        err.map(|(exports, sem_pkg)| (Arc::new(exports), Arc::new(sem_pkg)))
            .map_err(Arc::new)
    })?;

    Ok((Arc::new(exports), Arc::new(sem_pkg)))
}

fn package_exports(
    db: &dyn Flux,
    path: String,
) -> SalvageResult<Arc<PackageExports>, Arc<FileErrors>> {
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

fn semantic_package_cycle(
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

fn recover_cycle2<T>(
    db: &dyn Flux,
    cycle: &[String],
    name: &str,
) -> SalvageResult<T, Arc<FileErrors>> {
    let mut cycle: Vec<_> = cycle
        .iter()
        .filter(|k| k.starts_with("semantic_package_cycle("))
        .map(|k| {
            k.trim_matches(|c: char| c != '"')
                .trim_matches('"')
                .trim_start_matches('@')
                .to_string()
        })
        .collect();
    cycle.pop();

    Err(Arc::new(FileErrors {
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
    })
    .into())
}
fn recover_cycle<T>(_db: &dyn Flux, cycle: &[String], name: &str) -> Result<T, nodes::ErrorKind> {
    // We get a list of strings like "semantic_package_inner(\"b\")",
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
        self.semantic_package_cycle(path.into())
            .map(|exports| exports.typ())
    }
    fn symbol(&mut self, path: &str, symbol_name: &str) -> Option<Symbol> {
        self.semantic_package_cycle(path.into())
            .ok()
            .and_then(|exports| exports.lookup_symbol(symbol_name).cloned())
    }
}

impl Importer for &dyn Flux {
    fn import(&mut self, path: &str) -> Result<PolyType, nodes::ErrorKind> {
        self.semantic_package_cycle(path.into())
            .map(|exports| exports.typ())
    }
    fn symbol(&mut self, path: &str, symbol_name: &str) -> Option<Symbol> {
        self.semantic_package_cycle(path.into())
            .ok()
            .and_then(|exports| exports.lookup_symbol(symbol_name).cloned())
    }
}

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
    fmt,
    io::{self, Read},
    path::PathBuf,
    sync::{
        atomic::{self, AtomicUsize},
        Arc, Mutex, RwLock,
    },
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

/// Interface for retrieving external flux modules
pub trait Fluxmod: fmt::Debug + std::panic::RefUnwindSafe {
    fn get_module(&self, module: &str) -> Option<Vec<(String, Arc<str>)>>;
}

#[derive(Debug)]
struct HttpFluxmod {
    base_url: String,
    token: String,
    // The `RwLock` implements `UnwindSafe`, allowing this to be stored in salsa (which uses
    // panics for some errors)
    agent: RwLock<ureq::Agent>,
}

impl HttpFluxmod {
    fn new(base_url: String, token: String) -> Self {
        HttpFluxmod {
            base_url,
            token,
            agent: RwLock::new(ureq::agent()),
        }
    }

    #[cfg(all(test, feature = "integration_test"))]
    fn publish(
        &self,
        module: &str,
        files: Vec<(String, Arc<str>)>,
        version: &semver::Version,
    ) -> Result<()> {
        let mut multipart = multipart::client::lazy::Multipart::new();
        for (name, contents) in &files {
            let name =
                percent_encoding::utf8_percent_encode(name, &percent_encoding::NON_ALPHANUMERIC)
                    .to_string();
            multipart.add_stream("module", contents.as_bytes(), Some(name), None);
        }
        let body = multipart
            .prepare()
            .map_err(|err| Error::Message(format!("Unable to publish module: {}", err)))?;

        let agent = self.agent.read().unwrap();
        let response = agent
            .post(&format!("{}/{}/@v/v{}.zip", self.base_url, module, version))
            .set("Authorization", &format!("Token {}", self.token))
            .set(
                "Content-Type",
                &format!("multipart/form-data; boundary={}", body.boundary()),
            )
            .send(body)
            .map_err(|err| Self::ureq_error("Unable to publish module", err))?;

        if !(200..300).contains(&response.status()) {
            return Err(Error::Message(format!(
                "Unable to publish module: {} {}",
                response.status(),
                response
                    .into_string()
                    .map_err(|err| Error::Message(err.to_string()))?
            )));
        }

        Ok(())
    }

    fn ureq_error(msg: &str, err: ureq::Error) -> Error {
        match err {
            ureq::Error::Status(status, response) => Error::Message(format!(
                "{}: {} {}",
                msg,
                status,
                match response
                    .into_string()
                    .map_err(|err| Error::Message(err.to_string()))
                {
                    Ok(text) => text,
                    Err(err) => return err,
                }
            )),
            _ => Error::Message(format!("{}: {}", msg, err)),
        }
    }

    fn latest_version(&self, module: &str) -> Result<String> {
        let agent = self.agent.read().unwrap();
        let response = agent
            .get(&format!("{}/{}/@latest", self.base_url, module))
            .set("Authorization", &format!("Token {}", self.token))
            .call()
            .map_err(|err| Self::ureq_error("Unable to retrieve the latest version", err))?;
        if response.status() != 200 {
            return Err(Error::Message(
                response
                    .into_string()
                    .map_err(|err| Error::Message(err.to_string()))?,
            ));
        }

        let version = response
            .into_string()
            .map_err(|err| Error::Message(err.to_string()))?
            .trim()
            .to_owned();

        Ok(version)
    }

    fn download_module(&self, module: &str) -> Result<Vec<(String, Arc<str>)>> {
        log::debug!("Searching fluxmod for `{}`", module);
        if module.contains("/") {
            return Err(Error::Message(format!("Invalid module name `{}`", module)));
        }

        let version = self.latest_version(module)?;
        log::debug!("Found latest version `{}` for `{}`", version, module);

        let agent = self.agent.read().unwrap();
        let response = agent
            .get(&format!("{}/{}/@v/{}.zip", self.base_url, module, version))
            .set("Authorization", &format!("Token {}", self.token))
            .call()
            .map_err(|err| Self::ureq_error("Unable to download flux module", err))?;

        if response.status() != 200 {
            return Err(Error::Message(
                response
                    .into_string()
                    .map_err(|err| Error::Message(err.to_string()))?,
            ));
        }

        let mut module_files = Vec::new();
        let mut bytes = Vec::new();
        // TODO Limit the size
        response
            .into_reader()
            .read_to_end(&mut bytes)
            .map_err(|err| Error::Message(err.to_string()))?;
        let mut archive = zip::ZipArchive::new(io::Cursor::new(&bytes[..]))
            .map_err(|err| Error::Message(err.to_string()))?;
        for i in 0..archive.len() {
            let mut file = archive
                .by_index(i)
                .map_err(|err| Error::Message(err.to_string()))?;

            let mut source = String::new();
            file.read_to_string(&mut source)
                .map_err(|err| Error::Message(err.to_string()))?;

            module_files.push(([module, file.name()].join("/"), Arc::from(source)));
        }

        Ok(module_files)
    }
}

impl Fluxmod for HttpFluxmod {
    fn get_module(&self, module: &str) -> Option<Vec<(String, Arc<str>)>> {
        self.download_module(module)
            .map_err(|err| eprintln!("{}", err))
            .ok()
    }
}

type MockFluxmod = HashMap<String, Vec<(String, Arc<str>)>>;

impl Fluxmod for MockFluxmod {
    fn get_module(&self, path: &str) -> Option<Vec<(String, Arc<str>)>> {
        let module = path.split('/').next()?;
        self.get(module).map(|v| {
            v.iter()
                .map(|(file, v)| ([path, file].join("/"), v.clone()))
                .collect()
        })
    }
}

// Technically exposed through the `FluxBase` trait but should only be used internally
#[doc(hidden)]
pub struct DepthGuard<'a> {
    flux: &'a dyn FluxBase,
}

impl Drop for DepthGuard<'_> {
    fn drop(&mut self) {
        self.flux.exit_scope();
    }
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

    #[doc(hidden)]
    fn enter_scope(&self) -> Result<DepthGuard<'_>>;
    #[doc(hidden)]
    fn exit_scope(&self);
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

    /// Defines the fluxmod interface for fetching external modules
    #[salsa::input]
    fn fluxmod(&self) -> Option<Arc<dyn Fluxmod>>;

    /// Enables the prelude for all compiled packages
    ///
    /// Default: true
    #[salsa::input]
    fn use_prelude(&self) -> bool;

    /// Sets any precompiled packages that should be included in the compilation
    #[salsa::input]
    fn precompiled_packages(&self) -> Option<&'static Packages>;

    /// Defines the fluxmod interface for fetching external modules
    fn get_flux_module(&self, module: String) -> Option<Arc<Vec<(String, Arc<str>)>>>;

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
    base_url: Option<String>,
    token: Option<String>,
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

    /// Enables fluxmod lookups for the database
    pub fn enable_fluxmod(mut self, base_url: String, token: String) -> Self {
        log::debug!("Enabling fluxmod");
        self.base_url = Some(base_url);
        self.token = Some(token);
        self
    }

    /// Builds the flux compiler database
    pub fn build(self) -> Database {
        let mut db = Database {
            filesystem_roots: self.filesystem_roots,
            ..Default::default()
        };

        if let (Some(base_url), Some(token)) = (self.base_url, self.token) {
            db.set_fluxmod(Some(Arc::new(HttpFluxmod::new(base_url, token))));
        }

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
    package_depth: AtomicUsize,
}

impl Default for Database {
    fn default() -> Self {
        let mut db = Self {
            storage: Default::default(),
            packages: Default::default(),
            package_errors: Default::default(),
            filesystem_roots: Vec::new(),
            package_depth: AtomicUsize::default(),
        };
        db.set_analyzer_config(AnalyzerConfig::default());
        db.set_use_prelude(true);
        db.set_precompiled_packages(None);
        db.set_fluxmod(None);
        db
    }
}

impl salsa::Database for Database {}

fn is_part_of_package(package: &str, path: &str) -> bool {
    // Example: package: `internal/boolean` matches the file
    // `internal/boolean/XXX.flux`
    path.starts_with(package)
        && path[package.len()..].starts_with('/')
        && path[package.len() + 1..].split('/').count() == 1
}

impl FluxBase for Database {
    fn package_files(&self, package: &str) -> Result<Vec<String>> {
        let mut found_files = self.search_flux_files(package)?;

        let packages = self.packages.lock().unwrap();

        found_files.extend(
            packages
                .iter()
                .filter(|path| is_part_of_package(package, path))
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
                self.packages.lock().unwrap().insert(path);
                return Ok(Arc::from(source));
            }
        }
        if self.fluxmod().is_some() {
            let module = path.split('/').next().unwrap();
            // TODO Only reach out to fluxmod if `module` points to a registry
            if let Some(modules) = self.get_flux_module(module.to_owned()) {
                return if let Some(source) = modules
                    .iter()
                    .find(|(k, _)| *k == path)
                    .map(|(_, source)| source)
                {
                    Ok(source.clone())
                } else {
                    Err(Error::Message(format!(
                        "No files exist for package `{}`",
                        path
                    )))
                };
            }
        }

        Ok(self.source_inner(path))
    }

    fn set_source(&mut self, path: String, source: Arc<str>) {
        self.packages.lock().unwrap().insert(path.clone());

        self.set_source_inner(path, source)
    }

    fn enter_scope(&self) -> Result<DepthGuard<'_>> {
        // This is maybe a bit low but should be fine for most practical uses and is low enough
        // that the depth test errors instead of overflowing the stack. Could be raised if there is
        // a need for it (though that requires some tweaking to get the test working).
        const MAX_MODULE_DEPTH: usize = 60;
        let depth = self.package_depth.fetch_add(1, atomic::Ordering::Acquire);
        if depth < MAX_MODULE_DEPTH {
            Ok(DepthGuard { flux: self })
        } else {
            Err(Error::Message(format!(
                "Module imports exceeded the maximum depth of `{}`",
                MAX_MODULE_DEPTH
            )))
        }
    }
    fn exit_scope(&self) {
        self.package_depth.fetch_sub(1, atomic::Ordering::Release);
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

        if self.fluxmod().is_some() {
            let module = package.split('/').next().unwrap();
            // TODO Only reach out to fluxmod if `module` points to a registry
            if let Some(modules) = self.get_flux_module(module.to_owned()) {
                found_files.extend(
                    modules
                        .iter()
                        .map(|(k, _)| k.clone())
                        .filter(|path| is_part_of_package(package, path)),
                );
            }
        }

        // It is possible that we find the same file twice if the roots contain duplicates
        found_files.sort();
        found_files.dedup();

        Ok(found_files)
    }
}

fn get_flux_module(db: &dyn Flux, module: String) -> Option<Arc<Vec<(String, Arc<str>)>>> {
    if let Some(fluxmod) = db.fluxmod() {
        fluxmod.get_module(&module).map(Arc::new)
    } else {
        None
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
        let types = db.package_exports(name.into()).map_err(|err| err.error)?;

        prelude_map.copy_bindings_from(&types);
    }
    Ok(Arc::new(prelude_map))
}

fn semantic_package(
    db: &dyn Flux,
    path: String,
) -> SalvageResult<(Arc<PackageExports>, Arc<nodes::Package>), Error> {
    let _guard = db.enter_scope()?;
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
    let file = db.ast_package(path)?;

    let env = Environment::new(prelude.into());
    let mut importer = db;
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

#[cfg(test)]
mod tests {
    use super::*;

    #[cfg(feature = "integration_test")]
    fn setup_module(fluxmod: &HttpFluxmod, modules: MockFluxmod) {
        let version = fluxmod
            .latest_version("mymodule")
            .unwrap_or_else(|err| panic!("{}", err));

        let mut version = semver::Version::parse(version.trim_start_matches("v")).unwrap();
        version.patch += 1;

        for (name, module) in modules {
            fluxmod
                .publish(&name, module, &version)
                .unwrap_or_else(|err| panic!("{}", err));
        }
    }

    #[cfg(not(feature = "integration_test"))]
    fn test_db(modules: MockFluxmod) -> Database {
        let mut db = Database::default();
        db.set_use_prelude(false);
        db.set_fluxmod(Some(Arc::new(modules)));
        db
    }

    #[cfg(feature = "integration_test")]
    fn test_db(modules: MockFluxmod) -> Database {
        let fluxmod = HttpFluxmod::new(
            std::env::var("FLUXMOD_BASE_URL").unwrap_or_else(|err| panic!("{}", err)),
            std::env::var("FLUXMOD_TOKEN").unwrap_or_else(|err| panic!("{}", err)),
        );
        setup_module(&fluxmod, modules);

        let mut db = Database::default();
        db.set_use_prelude(false);
        db.set_fluxmod(Some(Arc::new(fluxmod)));
        db
    }

    #[test]
    fn fluxmod() {
        let _ = env_logger::try_init();

        let mut db = test_db(
            [(
                "mymodule".into(),
                vec![("pack.flux".into(), "x = 1".into())],
            )]
            .into_iter()
            .collect(),
        );

        db.set_source(
            "main/main.flux".into(),
            r#"
        import "mymodule"
        y = mymodule.x + 1
        "#
            .into(),
        );

        match db.semantic_package("main".into()) {
            Ok(_) => (),
            Err(err) => {
                let mut errors = db.package_errors();
                errors.push(err.error);
                panic!("{}", errors);
            }
        }
    }

    #[test]
    fn fluxmod_nested() {
        let _ = env_logger::try_init();

        let mut db = test_db(
            [(
                "mymodulenested".into(),
                vec![
                    ("nested/nested.flux".into(), "y = 1".into()),
                    ("nested/nestedagain/nestedagain.flux".into(), "z = 3".into()),
                ],
            )]
            .into_iter()
            .collect(),
        );

        db.set_source(
            "main/main.flux".into(),
            r#"
        import "mymodulenested/nested"
        import "mymodulenested/nested/nestedagain"
        y = nested.y + nestedagain.z
        "#
            .into(),
        );

        match db.semantic_package("main".into()) {
            Ok(_) => (),
            Err(err) => {
                let mut errors = db.package_errors();
                errors.push(err.error);
                panic!("{}", errors);
            }
        }
    }

    #[test]
    fn fluxmod_recursive_dependencies() {
        let _ = env_logger::try_init();

        let mut db = test_db(
            [
                (
                    "recursive_mymodule".into(),
                    vec![(
                        "pack.flux".into(),
                        Arc::from(
                            r#"
                    import "recursive_mymodule2"
                    x = 1 + recursive_mymodule2.y
                    "#,
                        ),
                    )],
                ),
                (
                    "recursive_mymodule2".into(),
                    vec![("main.flux".into(), Arc::from("y = 3"))],
                ),
            ]
            .into_iter()
            .collect(),
        );

        db.set_source(
            "main/main.flux".into(),
            r#"
        import "recursive_mymodule"
        y = recursive_mymodule.x + 1
        "#
            .into(),
        );

        match db.semantic_package("main".into()) {
            Ok(_) => (),
            Err(err) => {
                let mut errors = db.package_errors();
                errors.push(err.error);
                panic!("{}", errors);
            }
        }
    }

    #[test]
    fn fluxmod_recursive_dependencies_2() {
        let _ = env_logger::try_init();

        let mut db = test_db(
            [
                (
                    "recursive2_mymodule".into(),
                    vec![(
                        "pack.flux".into(),
                        Arc::from(
                            r#"
                    import "recursive2_mymodule2"
                    x = 1 + recursive2_mymodule2.y
                    "#,
                        ),
                    )],
                ),
                (
                    "recursive2_mymodule2".into(),
                    vec![("main.flux".into(), Arc::from("y = 3"))],
                ),
            ]
            .into_iter()
            .collect(),
        );

        db.set_source(
            "main/main.flux".into(),
            r#"
        import "recursive2_mymodule"
        import "recursive2_mymodule2"
        y = recursive2_mymodule.x + recursive2_mymodule2.y + 3
        "#
            .into(),
        );

        match db.semantic_package("main".into()) {
            Ok(_) => (),
            Err(err) => {
                let mut errors = db.package_errors();
                errors.push(err.error);
                panic!("{}", errors);
            }
        }
    }

    #[test]
    fn fluxmod_cyclic_dependency() {
        let _ = env_logger::try_init();

        let mut db = test_db(
            [
                (
                    "cycle".into(),
                    vec![(
                        "pack.flux".into(),
                        Arc::from(
                            r#"
                    import "cycle2"
                    x = 1 + cycle2.y
                    "#,
                        ),
                    )],
                ),
                (
                    "cycle2".into(),
                    vec![(
                        "main.flux".into(),
                        Arc::from(
                            r#"
                    import "cycle"
                    y = cycle.x + 3
                    "#,
                        ),
                    )],
                ),
            ]
            .into_iter()
            .collect(),
        );

        db.set_source(
            "main/main.flux".into(),
            r#"
        import "cycle"
        y = cycle.x
        "#
            .into(),
        );

        match db.semantic_package("main".into()) {
            Ok(_) => panic!("Expected cycle error"),
            Err(err) => {
                let mut errors = db.package_errors();
                errors.push(err.error);

                // TODO This should ideally just be one error
                expect_test::expect![[r#"
                    error @0:0-0:0: package "cycle" depends on itself: cycle -> cycle -> cycle

                    error @0:0-0:0: package "cycle2" depends on itself: cycle2 -> cycle -> cycle2

                    error main/main.flux@2:9-2:23: package "cycle" depends on itself: cycle -> cycle -> cycle"#]].assert_eq(&errors.to_string());
            }
        }
    }

    #[test]
    fn fluxmod_deep_dependency_error() {
        let _ = env_logger::try_init();

        let mut db = Database::default();
        db.set_use_prelude(false);

        // Returns a module "deepN" which imports "deep{N+1}" etc
        #[derive(Default, Debug)]
        struct DeepFluxmod;
        impl Fluxmod for DeepFluxmod {
            fn get_module(&self, module: &str) -> Option<Vec<(String, Arc<str>)>> {
                eprintln!("{}", module);
                if module.starts_with("deep") {
                    let i = module.trim_start_matches("deep").parse::<i32>().unwrap() + 1;
                    Some(vec![(
                        format!("{}/file.flux", module),
                        Arc::from(format!(
                            r#"
                            import "deep{i}"
                            x = deep{i}.x
                        "#,
                            i = i,
                        )),
                    )])
                } else {
                    None
                }
            }
        }
        db.set_fluxmod(Some(Arc::new(DeepFluxmod::default())));

        db.set_source(
            "main/main.flux".into(),
            r#"
        import "deep0"
        y = deep0.x
        "#
            .into(),
        );

        match db.semantic_package("main".into()) {
            Ok(_) => panic!("Expected cycle error"),
            Err(err) => {
                let mut errors = db.package_errors();
                errors.push(err.error);

                // TODO This should ideally just be one error
                expect_test::expect![[r#"
                    error deep40/file.flux@2:29-2:44: invalid import path deep41

                    error deep19/file.flux@2:29-2:44: invalid import path deep20

                    error deep4/file.flux@2:29-2:43: invalid import path deep5

                    error deep52/file.flux@2:29-2:44: invalid import path deep53

                    error deep47/file.flux@2:29-2:44: invalid import path deep48

                    error deep38/file.flux@2:29-2:44: invalid import path deep39

                    error deep54/file.flux@2:29-2:44: invalid import path deep55

                    error deep43/file.flux@2:29-2:44: invalid import path deep44

                    error deep39/file.flux@2:29-2:44: invalid import path deep40

                    error deep51/file.flux@2:29-2:44: invalid import path deep52

                    error deep1/file.flux@2:29-2:43: invalid import path deep2

                    error deep58/file.flux@2:29-2:44: invalid import path deep59

                    error deep57/file.flux@2:29-2:44: invalid import path deep58

                    error deep31/file.flux@2:29-2:44: invalid import path deep32

                    error deep35/file.flux@2:29-2:44: invalid import path deep36

                    error deep22/file.flux@2:29-2:44: invalid import path deep23

                    error deep41/file.flux@2:29-2:44: invalid import path deep42

                    error deep2/file.flux@2:29-2:43: invalid import path deep3

                    error deep45/file.flux@2:29-2:44: invalid import path deep46

                    error deep34/file.flux@2:29-2:44: invalid import path deep35

                    error deep9/file.flux@2:29-2:44: invalid import path deep10

                    error deep26/file.flux@2:29-2:44: invalid import path deep27

                    error deep21/file.flux@2:29-2:44: invalid import path deep22

                    error deep10/file.flux@2:29-2:44: invalid import path deep11

                    error deep48/file.flux@2:29-2:44: invalid import path deep49

                    error deep13/file.flux@2:29-2:44: invalid import path deep14

                    error deep49/file.flux@2:29-2:44: invalid import path deep50

                    error deep42/file.flux@2:29-2:44: invalid import path deep43

                    error deep36/file.flux@2:29-2:44: invalid import path deep37

                    error deep17/file.flux@2:29-2:44: invalid import path deep18

                    error deep44/file.flux@2:29-2:44: invalid import path deep45

                    error deep16/file.flux@2:29-2:44: invalid import path deep17

                    error deep53/file.flux@2:29-2:44: invalid import path deep54

                    error deep55/file.flux@2:29-2:44: invalid import path deep56

                    error deep0/file.flux@2:29-2:43: invalid import path deep1

                    error deep5/file.flux@2:29-2:43: invalid import path deep6

                    error deep20/file.flux@2:29-2:44: invalid import path deep21

                    error deep29/file.flux@2:29-2:44: invalid import path deep30

                    error deep25/file.flux@2:29-2:44: invalid import path deep26

                    error deep12/file.flux@2:29-2:44: invalid import path deep13

                    error deep50/file.flux@2:29-2:44: invalid import path deep51

                    error deep3/file.flux@2:29-2:43: invalid import path deep4

                    error deep6/file.flux@2:29-2:43: invalid import path deep7

                    error deep28/file.flux@2:29-2:44: invalid import path deep29

                    error deep8/file.flux@2:29-2:43: invalid import path deep9

                    error deep15/file.flux@2:29-2:44: invalid import path deep16

                    error deep56/file.flux@2:29-2:44: invalid import path deep57

                    error deep37/file.flux@2:29-2:44: invalid import path deep38

                    Module imports exceeded the maximum depth of `60`

                    error deep27/file.flux@2:29-2:44: invalid import path deep28

                    error deep14/file.flux@2:29-2:44: invalid import path deep15

                    error deep18/file.flux@2:29-2:44: invalid import path deep19

                    error deep7/file.flux@2:29-2:43: invalid import path deep8

                    error deep23/file.flux@2:29-2:44: invalid import path deep24

                    error deep46/file.flux@2:29-2:44: invalid import path deep47

                    error deep30/file.flux@2:29-2:44: invalid import path deep31

                    error deep11/file.flux@2:29-2:44: invalid import path deep12

                    error deep32/file.flux@2:29-2:44: invalid import path deep33

                    error deep33/file.flux@2:29-2:44: invalid import path deep34

                    error deep24/file.flux@2:29-2:44: invalid import path deep25

                    error main/main.flux@2:9-2:23: invalid import path deep0"#]]
                .assert_eq(&errors.to_string());
            }
        }
    }
}

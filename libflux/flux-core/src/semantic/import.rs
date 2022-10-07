//! Module import defines the abstractions for importing Flux package types from various sources.
use std::sync::Arc;

use crate::semantic::{
    nodes::{ErrorKind, Symbol},
    types::{PolyType, SemanticMap},
    PackageExports,
};

/// Importer defines an API for resolving Flux import paths to their corresponding types.
pub trait Importer {
    /// Import resolves an absolute import path and returns the type for the corresponding Flux
    /// package or None if no such package exists.
    fn import(&mut self, path: &str) -> Result<PolyType, ErrorKind> {
        Err(ErrorKind::InvalidImportPath(path.to_owned()))
    }

    /// Returns the `Symbol` for a specific exported item from a package
    fn symbol(&mut self, _package_path: &str, _symbol_name: &str) -> Option<Symbol> {
        None
    }
}

impl<T> Importer for &'_ mut T
where
    T: ?Sized + Importer,
{
    fn import(&mut self, path: &str) -> Result<PolyType, ErrorKind> {
        T::import(self, path)
    }
    fn symbol(&mut self, package_path: &str, symbol_name: &str) -> Option<Symbol> {
        T::symbol(self, package_path, symbol_name)
    }
}

/// In memory storage for packages
pub type Packages = SemanticMap<String, Arc<PackageExports>>;

impl Importer for Packages {
    fn import(&mut self, path: &str) -> Result<PolyType, ErrorKind> {
        self.get(path)
            .map(|exports| exports.typ())
            .ok_or_else(|| ErrorKind::InvalidImportPath(path.to_owned()))
    }
    fn symbol(&mut self, package_path: &str, symbol_name: &str) -> Option<Symbol> {
        self.get(package_path)
            .and_then(|exports| exports.lookup_symbol(symbol_name))
            .cloned()
    }
}

impl Importer for &'_ Packages {
    fn import(&mut self, path: &str) -> Result<PolyType, ErrorKind> {
        self.get(path)
            .map(|exports| exports.typ())
            .ok_or_else(|| ErrorKind::InvalidImportPath(path.to_owned()))
    }
    fn symbol(&mut self, package_path: &str, symbol_name: &str) -> Option<Symbol> {
        self.get(package_path)
            .and_then(|exports| exports.lookup_symbol(symbol_name))
            .cloned()
    }
}

//! Module import defines the abstractions for importing Flux package types from various sources.

use crate::semantic::{
    nodes::Symbol,
    types::{PolyType, SemanticMap},
    PackageExports,
};

/// Importer defines an API for resolving Flux import paths to their corresponding types.
pub trait Importer {
    /// Import resolves an absolute import path and returns the type for the corresponding Flux
    /// package or None if no such package exists.
    fn import(&mut self, _path: &str) -> Option<PolyType> {
        None
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
    fn import(&mut self, path: &str) -> Option<PolyType> {
        T::import(self, path)
    }
    fn symbol(&mut self, package_path: &str, symbol_name: &str) -> Option<Symbol> {
        T::symbol(self, package_path, symbol_name)
    }
}

/// In memory storage for packages
pub type Packages = SemanticMap<String, PackageExports>;

impl Importer for Packages {
    fn import(&mut self, path: &str) -> Option<PolyType> {
        self.get(path).map(|exports| exports.typ())
    }
    fn symbol(&mut self, package_path: &str, symbol_name: &str) -> Option<Symbol> {
        self.get(package_path)
            .and_then(|exports| exports.lookup_symbol(symbol_name))
            .cloned()
    }
}

impl Importer for &'_ Packages {
    fn import(&mut self, path: &str) -> Option<PolyType> {
        self.get(path).map(|exports| exports.typ())
    }
    fn symbol(&mut self, package_path: &str, symbol_name: &str) -> Option<Symbol> {
        self.get(package_path)
            .and_then(|exports| exports.lookup_symbol(symbol_name))
            .cloned()
    }
}

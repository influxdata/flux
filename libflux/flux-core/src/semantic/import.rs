//! Module import defines the abstractions for importing Flux package types from various sources.

use crate::semantic::{
    types::{PolyType, PolyTypeMap, SemanticMap},
    PackageExports,
};

/// Importer defines an API for resolving Flux import paths to their corresponding types.
pub trait Importer {
    /// Import resolves an absolute import path and returns the type for the corresponding Flux
    /// package or None if no such package exists.
    fn import(&mut self, _path: &str) -> Option<PolyType> {
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
}

/// In memory storage for packages
pub type Packages = SemanticMap<String, PackageExports>;

impl Importer for Packages {
    fn import(&mut self, path: &str) -> Option<PolyType> {
        self.get(path).map(|exports| exports.typ())
    }
}

impl Importer for PolyTypeMap {
    fn import(&mut self, name: &str) -> Option<PolyType> {
        self.get(name).cloned()
    }
}

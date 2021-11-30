//! Module import defines the abstractions for importing Flux package types from various sources.

use crate::semantic::{
    types::{PolyType, PolyTypeMap},
    ExternalEnvironment,
};

/// Importer defines an API for resolving Flux import paths to their corresponding types.
pub trait Importer {
    /// Import resolves an absolute import path and returns the type for the corresponding Flux
    /// package or None if no such package exists.
    fn import(&mut self, _path: &str) -> Option<PolyType> {
        None
    }
}
impl Importer for ExternalEnvironment {
    fn import(&mut self, path: &str) -> Option<PolyType> {
        self.lookup(path).cloned()
    }
}
impl Importer for PolyTypeMap {
    fn import(&mut self, name: &str) -> Option<PolyType> {
        self.get(name).cloned()
    }
}

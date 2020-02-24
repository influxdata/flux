use crate::semantic::types::PolyType;
use std::collections::HashMap;

pub trait Importer {
    fn import(&self, _name: &str) -> Option<PolyType> {
        None
    }
}

impl<S: std::hash::BuildHasher> Importer for HashMap<String, PolyType, S> {
    fn import(&self, name: &str) -> Option<PolyType> {
        match self.get(name) {
            Some(pty) => Some(pty.clone()),
            None => None,
        }
    }
}

impl Importer for Box<dyn Importer> {
    fn import(&self, name: &str) -> Option<PolyType> {
        self.as_ref().import(name)
    }
}

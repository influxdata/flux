use crate::semantic::types::PolyType;
use std::collections::HashMap;

pub trait Importer {
    fn import(&self, _name: &str) -> Option<PolyType> {
        None
    }
}

impl Importer for HashMap<String, PolyType> {
    fn import(&self, name: &str) -> Option<PolyType> {
        match self.get(name) {
            Some(pty) => Some(pty.clone()),
            None => None,
        }
    }
}

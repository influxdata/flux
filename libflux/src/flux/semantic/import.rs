use crate::semantic::types::PolyType;
use std::collections::HashMap;

pub trait Importer {
    fn import(&self, _name: &str) -> Option<&PolyType> {
        None
    }
}

impl Importer for HashMap<String, PolyType> {
    fn import(&self, name: &str) -> Option<&PolyType> {
        self.get(name)
    }
}

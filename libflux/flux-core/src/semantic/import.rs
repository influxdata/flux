use crate::semantic::types::{PolyType, PolyTypeMap};

pub trait Importer {
    fn import(&self, _name: &str) -> Option<PolyType> {
        None
    }
}

impl Importer for PolyTypeMap {
    fn import(&self, name: &str) -> Option<PolyType> {
        self.get(name).cloned()
    }
}

impl Importer for Box<dyn Importer> {
    fn import(&self, name: &str) -> Option<PolyType> {
        self.as_ref().import(name)
    }
}

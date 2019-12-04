use crate::semantic::types::PolyType;

pub trait Importer {
    fn import(&self, _name: &str) -> Option<&PolyType> {
        None
    }
}

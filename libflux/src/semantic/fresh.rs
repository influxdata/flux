use crate::semantic::types::Tvar;

// Fresher returns incrementing type variables
pub struct Fresher(u64);

// Create a tvar fresher from a u64
impl From<u64> for Fresher {
    fn from(u: u64) -> Fresher {
        Fresher(u)
    }
}

impl Fresher {
    pub fn new() -> Fresher {
        Fresher(0)
    }

    pub fn fresh(&mut self) -> Tvar {
        let u = self.0;
        self.0 += 1;
        Tvar(u)
    }
}

use crate::semantic::types::{MonoType, Tvar};
use std::{collections::HashMap, fmt};

// A substitution defines a function that takes a monotype as input
// and returns a monotype as output. The output type is interpreted
// as being equivalent to the input type.
//
// Substitutions are idempotent. Given a substitution s and an input
// type x, we have s(s(x)) = s(x).
//
#[derive(Debug, PartialEq)]
pub struct Subst(HashMap<Tvar, MonoType>);

impl fmt::Display for Subst {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        f.write_str("substitution:\n")?;
        for (k, v) in &self.0 {
            write!(f, "\t{}: {}\n", k, v)?;
        }
        Ok(())
    }
}

impl Subst {
    pub fn empty() -> Subst {
        Subst(HashMap::new())
    }

    pub fn init(values: HashMap<Tvar, MonoType>) -> Subst {
        Subst(values)
    }

    pub fn lookup(&self, tv: Tvar) -> Option<&MonoType> {
        self.0.get(&tv)
    }

    pub fn merge(self, with: Subst) -> Subst {
        let applied: HashMap<Tvar, MonoType> = self
            .0
            .into_iter()
            .map(|(k, v)| (k, v.apply(&with)))
            .collect();
        Subst(applied.into_iter().chain(with.0.into_iter()).collect())
    }
}

// A type is substitutable if a substitution can be applied to it.
pub trait Substitutable {
    fn apply(self, sub: &Subst) -> Self;
}

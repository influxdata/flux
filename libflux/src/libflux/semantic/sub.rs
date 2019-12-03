use crate::semantic::types::{MonoType, Tvar};
use std::collections::HashMap;

// A substitution defines a function that takes a monotype as input
// and returns a monotype as output. The output type is interpreted
// as being equivalent to the input type.
//
// Substitutions are idempotent. Given a substitution s and an input
// type x, we have s(s(x)) = s(x).
//
#[derive(Debug, PartialEq)]
pub struct Substitution(HashMap<Tvar, MonoType>);

// Derive a substitution from a hash map.
impl From<HashMap<Tvar, MonoType>> for Substitution {
    fn from(values: HashMap<Tvar, MonoType>) -> Substitution {
        Substitution(values)
    }
}

// Derive a hash map from a substitution.
impl From<Substitution> for HashMap<Tvar, MonoType> {
    fn from(sub: Substitution) -> HashMap<Tvar, MonoType> {
        sub.0
    }
}

impl Substitution {
    pub fn empty() -> Substitution {
        Substitution(HashMap::new())
    }

    pub fn apply(&self, tv: Tvar) -> MonoType {
        match self.0.get(&tv) {
            Some(t) => t.clone(),
            None => MonoType::Var(tv),
        }
    }

    pub fn merge(self, with: Substitution) -> Substitution {
        let applied: HashMap<Tvar, MonoType> = self
            .0
            .into_iter()
            .map(|(k, v)| (k, v.apply(&with)))
            .collect();
        Substitution(applied.into_iter().chain(with.0.into_iter()).collect())
    }
}

// A type is substitutable if a substitution can be applied to it.
pub trait Substitutable {
    fn apply(self, sub: &Substitution) -> Self;
    fn free_vars(&self) -> Vec<Tvar>;
}

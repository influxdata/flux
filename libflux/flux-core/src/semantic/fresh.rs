//! "Fresh" type variable identifiers.

use crate::semantic::types::{
    Array, Function, MonoType, MonoTypeVecMap, PolyType, Property, Record, SemanticMap, Tvar,
    TvarMap,
};
use std::collections::BTreeMap;
use std::hash::Hash;

/// A struct used for incrementing type variable identifiers.
#[derive(Default)]
pub struct Fresher(pub u64);

/// Creates a type variable [`Fresher`] from a `u64`.
impl From<u64> for Fresher {
    fn from(u: u64) -> Fresher {
        Fresher(u)
    }
}

impl Fresher {
    /// Takes a `Fresher` and returns an incremented [`Tvar`].
    pub fn fresh(&mut self) -> Tvar {
        let u = self.0;
        self.0 += 1;
        Tvar(u)
    }
}

/// Trait for implementing `fresh` for various types.
pub trait Fresh {
    #[allow(missing_docs)]
    fn fresh(self, f: &mut Fresher, sub: &mut TvarMap) -> Self;
}

impl<T: Fresh> Fresh for Vec<T> {
    fn fresh(self, f: &mut Fresher, sub: &mut TvarMap) -> Self {
        self.into_iter().map(|t| t.fresh(f, sub)).collect::<Self>()
    }
}

impl<T: Fresh> Fresh for Option<T> {
    fn fresh(self, f: &mut Fresher, sub: &mut TvarMap) -> Self {
        self.map(|t| t.fresh(f, sub))
    }
}

#[allow(clippy::implicit_hasher)]
impl<T: Fresh> Fresh for SemanticMap<String, T> {
    fn fresh(self, f: &mut Fresher, sub: &mut TvarMap) -> Self {
        self.into_iter()
            .collect::<BTreeMap<String, T>>()
            .into_iter()
            .map(|(s, t)| (s, t.fresh(f, sub)))
            .collect::<Self>()
    }
}

#[allow(clippy::implicit_hasher)]
impl<T: Hash + Ord + Eq + Fresh, S> Fresh for SemanticMap<T, S> {
    fn fresh(self, f: &mut Fresher, sub: &mut TvarMap) -> Self {
        self.into_iter()
            .collect::<BTreeMap<T, S>>()
            .into_iter()
            .map(|(t, s)| (t.fresh(f, sub), s))
            .collect::<Self>()
    }
}

impl<T: Fresh> Fresh for Box<T> {
    fn fresh(self, f: &mut Fresher, sub: &mut TvarMap) -> Self {
        Box::new((*self).fresh(f, sub))
    }
}

impl Fresh for PolyType {
    fn fresh(self, f: &mut Fresher, sub: &mut TvarMap) -> Self {
        let expr = self.expr.fresh(f, sub);
        let vars = self.vars.fresh(f, sub);
        let cons = self.cons.fresh(f, sub);
        PolyType { vars, cons, expr }
    }
}

impl Fresh for MonoType {
    fn fresh(self, f: &mut Fresher, sub: &mut TvarMap) -> Self {
        match self {
            MonoType::Var(tvr) => MonoType::Var(tvr.fresh(f, sub)),
            MonoType::Arr(arr) => MonoType::arr(arr.fresh(f, sub)),
            MonoType::Record(obj) => MonoType::record(obj.fresh(f, sub)),
            MonoType::Fun(fun) => MonoType::fun(fun.fresh(f, sub)),
            _ => self,
        }
    }
}

impl Fresh for Tvar {
    fn fresh(self, f: &mut Fresher, sub: &mut TvarMap) -> Self {
        *sub.entry(self).or_insert_with(|| f.fresh())
    }
}

impl Fresh for Array {
    fn fresh(self, f: &mut Fresher, sub: &mut TvarMap) -> Self {
        Array(self.0.fresh(f, sub))
    }
}

impl Fresh for Record {
    fn fresh(mut self, f: &mut Fresher, sub: &mut TvarMap) -> Self {
        let mut props = MonoTypeVecMap::new();
        let mut extends = false;
        let mut tv = Tvar(0);
        loop {
            match self {
                Record::Empty => {
                    break;
                }
                Record::Extension {
                    head,
                    tail: MonoType::Record(b),
                } => {
                    props.entry(head.k).or_insert_with(Vec::new).push(head.v);
                    self = *b;
                }
                Record::Extension {
                    head,
                    tail: MonoType::Var(t),
                } => {
                    extends = true;
                    tv = t;
                    props.entry(head.k).or_insert_with(Vec::new).push(head.v);
                    break;
                }
                _ => {
                    break;
                }
            }
        }
        // If record extends a tvar, freshen it.
        // Otherwise record must extend empty record.
        let mut r: MonoType = if extends {
            MonoType::Var(tv.fresh(f, sub))
        } else {
            MonoType::from(Record::Empty)
        };
        // Freshen record properties in deterministic order
        props = props.fresh(f, sub);
        // Construct new record from the fresh properties
        for (label, types) in props {
            for ty in types {
                let extension = Record::Extension {
                    head: Property {
                        k: label.clone(),
                        v: ty,
                    },
                    tail: r,
                };
                r = MonoType::from(extension);
            }
        }
        match r {
            MonoType::Record(b) => *b,
            _ => Record::Empty,
        }
    }
}

impl Fresh for Property {
    fn fresh(self, f: &mut Fresher, sub: &mut TvarMap) -> Self {
        Property {
            k: self.k,
            v: self.v.fresh(f, sub),
        }
    }
}

impl Fresh for Function {
    fn fresh(self, f: &mut Fresher, sub: &mut TvarMap) -> Self {
        Function {
            req: self.req.fresh(f, sub),
            opt: self.opt.fresh(f, sub),
            pipe: self.pipe.fresh(f, sub),
            retn: self.retn.fresh(f, sub),
        }
    }
}

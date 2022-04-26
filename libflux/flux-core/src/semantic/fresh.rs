//! "Fresh" type variable identifiers.

use std::cell::{Cell, RefCell};

use crate::semantic::{
    sub::{Substitutable, Substituter},
    types::{MonoType, Tvar, TvarMap},
};

/// A struct used for incrementing type variable identifiers.
#[derive(Default)]
pub struct Fresher {
    fresher: Cell<u64>,
    sub: RefCell<TvarMap>,
}

impl Fresher {
    /// Takes a `Fresher` and returns an incremented [`Tvar`].
    pub fn fresh(&mut self) -> Tvar {
        let u = self.fresher.get();
        self.fresher.set(u + 1);
        Tvar(u)
    }
}

impl From<u64> for Fresher {
    fn from(id: u64) -> Self {
        Fresher {
            fresher: Cell::new(id),
            sub: Default::default(),
        }
    }
}

impl Substituter for Fresher {
    fn try_apply(&mut self, var: Tvar) -> Option<MonoType> {
        let fresher = &self.fresher;
        Some(MonoType::Var(
            *self.sub.borrow_mut().entry(var).or_insert_with(|| {
                let u = fresher.get();
                fresher.set(u + 1);
                Tvar(u)
            }),
        ))
    }

    fn try_apply_bound(&mut self, var: Tvar) -> Option<MonoType> {
        let fresher = &self.fresher;
        Some(MonoType::BoundVar(
            *self.sub.borrow_mut().entry(var).or_insert_with(|| {
                let u = fresher.get();
                fresher.set(u + 1);
                Tvar(u)
            }),
        ))
    }

    fn visit_type(&mut self, typ: &MonoType) -> Option<MonoType> {
        use crate::semantic::types::{MonoTypeVecMap, Property, Record};
        match typ {
            MonoType::Var(var) => self
                .try_apply(*var)
                .map(|typ| typ.walk(self).unwrap_or(typ)),
            MonoType::BoundVar(var) => self
                .try_apply_bound(*var)
                .map(|typ| typ.walk(self).unwrap_or(typ)),
            MonoType::Record(record) => {
                let mut props = MonoTypeVecMap::new();

                let mut fields = record.fields();
                for field in &mut fields {
                    props
                        .entry(field.k.clone())
                        .or_insert_with(Vec::new)
                        .push(field.v.clone());
                }
                let tail = fields.tail();
                // If record extends a tvar, freshen it.
                // Otherwise record must extend empty record.
                let mut r: MonoType = if let Some(tail) = tail {
                    tail.visit(self).unwrap_or_else(|| tail.clone())
                } else {
                    MonoType::from(Record::Empty)
                };
                // Freshen record properties in deterministic order
                props = props.visit(self).unwrap_or(props);
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
                // TODO Should optimize when no variables needs freshening
                Some(MonoType::from(match r {
                    MonoType::Record(b) => (*b).clone(),
                    _ => Record::Empty,
                }))
            }
            _ => typ.walk(self),
        }
    }
}

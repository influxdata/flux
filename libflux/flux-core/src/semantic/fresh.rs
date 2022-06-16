//! "Fresh" type variable identifiers.

use crate::semantic::{
    sub::{Substitutable, Substituter},
    types::{BoundTvar, MonoType, Tvar, TvarMap},
};

/// A struct used for incrementing type variable identifiers.
#[derive(Default)]
pub struct Fresher {
    fresher: u64,
    sub: TvarMap,
}

impl Fresher {
    /// Takes a `Fresher` and returns an incremented [`Tvar`].
    pub fn fresh(&mut self) -> Tvar {
        let u = self.fresher;
        self.fresher += 1;
        Tvar(u)
    }

    fn fresh_var(&mut self, var: Tvar) -> Tvar {
        let fresher = &mut self.fresher;
        *self.sub.entry(var).or_insert_with(|| {
            let u = *fresher;
            *fresher += 1;
            Tvar(u)
        })
    }
}

impl From<u64> for Fresher {
    fn from(id: u64) -> Self {
        Fresher {
            fresher: id,
            sub: Default::default(),
        }
    }
}

impl Substituter for Fresher {
    fn try_apply(&mut self, var: Tvar) -> Option<MonoType> {
        Some(MonoType::Var(self.fresh_var(var)))
    }

    fn try_apply_bound(&mut self, var: BoundTvar) -> Option<MonoType> {
        Some(MonoType::BoundVar(BoundTvar(self.fresh_var(Tvar(var.0)).0)))
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

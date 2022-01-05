//! "Fresh" type variable identifiers.

use std::{collections::BTreeMap, hash::Hash};

use crate::semantic::{
    nodes::Symbol,
    sub::{merge, merge3, merge4, merge_collect},
    types::{
        Collection, Dictionary, Function, Kind, Label, MonoType, MonoTypeVecMap, PolyType,
        Property, Record, RecordLabel, SemanticMap, Tvar, TvarMap,
    },
};

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
    fn fresh(self, f: &mut Fresher, sub: &mut TvarMap) -> Self
    where
        Self: Sized,
    {
        self.fresh_ref(f, sub).unwrap_or(self)
    }
    #[allow(missing_docs)]
    fn fresh_ref(&self, f: &mut Fresher, sub: &mut TvarMap) -> Option<Self>
    where
        Self: Sized;
}

impl Fresh for RecordLabel {
    fn fresh_ref(&self, _: &mut Fresher, _: &mut TvarMap) -> Option<Self> {
        None
    }
}

impl Fresh for Label {
    fn fresh_ref(&self, _: &mut Fresher, _: &mut TvarMap) -> Option<Self> {
        None
    }
}

impl Fresh for Symbol {
    fn fresh_ref(&self, _: &mut Fresher, _: &mut TvarMap) -> Option<Self> {
        None
    }
}

impl Fresh for String {
    fn fresh_ref(&self, _: &mut Fresher, _: &mut TvarMap) -> Option<Self> {
        None
    }
}

impl Fresh for Kind {
    fn fresh_ref(&self, _: &mut Fresher, _: &mut TvarMap) -> Option<Self> {
        None
    }
}

impl<T: Fresh + Clone> Fresh for Vec<T> {
    fn fresh(self, f: &mut Fresher, sub: &mut TvarMap) -> Self {
        self.into_iter().map(|t| t.fresh(f, sub)).collect::<Self>()
    }
    fn fresh_ref(&self, f: &mut Fresher, sub: &mut TvarMap) -> Option<Self> {
        merge_collect(&mut (), self, |_, v| v.fresh_ref(f, sub), |_, v| v.clone())
    }
}

impl<T: Fresh> Fresh for Option<T> {
    fn fresh(self, f: &mut Fresher, sub: &mut TvarMap) -> Self {
        self.map(|t| t.fresh(f, sub))
    }
    fn fresh_ref(&self, f: &mut Fresher, sub: &mut TvarMap) -> Option<Self> {
        match self {
            None => None,
            Some(t) => t.fresh_ref(f, sub).map(Some),
        }
    }
}

#[allow(clippy::implicit_hasher)]
impl<T: Hash + Ord + Eq + Fresh + Clone, S: Clone + Fresh> Fresh for SemanticMap<T, S> {
    fn fresh(self, f: &mut Fresher, sub: &mut TvarMap) -> Self {
        self.into_iter()
            .collect::<BTreeMap<T, S>>()
            .into_iter()
            .map(|(t, s)| (t.fresh(f, sub), s.fresh(f, sub)))
            .collect::<Self>()
    }
    fn fresh_ref(&self, f: &mut Fresher, sub: &mut TvarMap) -> Option<Self> {
        merge_collect(
            &mut (),
            self,
            |_, (k, v)| merge(k, k.fresh_ref(f, sub), v, v.fresh_ref(f, sub)),
            |_, (k, v)| (k.clone(), v.clone()),
        )
    }
}

impl<T: Fresh> Fresh for Box<T> {
    fn fresh(self, f: &mut Fresher, sub: &mut TvarMap) -> Self {
        Box::new((*self).fresh(f, sub))
    }
    fn fresh_ref(&self, f: &mut Fresher, sub: &mut TvarMap) -> Option<Self> {
        T::fresh_ref(self, f, sub).map(Box::new)
    }
}

impl Fresh for PolyType {
    fn fresh(self, f: &mut Fresher, sub: &mut TvarMap) -> Self {
        let expr = self.expr.fresh(f, sub);
        let vars = self.vars.fresh(f, sub);
        let cons = self.cons.fresh(f, sub);
        PolyType { vars, cons, expr }
    }
    fn fresh_ref(&self, f: &mut Fresher, sub: &mut TvarMap) -> Option<Self> {
        let PolyType { vars, cons, expr } = self;
        merge3(
            expr,
            expr.fresh_ref(f, sub),
            vars,
            vars.fresh_ref(f, sub),
            cons,
            cons.fresh_ref(f, sub),
        )
        .map(|(expr, vars, cons)| PolyType { expr, vars, cons })
    }
}

impl Fresh for MonoType {
    fn fresh_ref(&self, f: &mut Fresher, sub: &mut TvarMap) -> Option<Self> {
        match self {
            MonoType::Error | MonoType::Builtin(_) | MonoType::Label(_) => None,
            MonoType::BoundVar(tvr) => tvr.fresh_ref(f, sub).map(MonoType::BoundVar),
            MonoType::Var(tvr) => tvr.fresh_ref(f, sub).map(MonoType::Var),
            MonoType::Collection(app) => app.fresh_ref(f, sub).map(MonoType::app),
            MonoType::Record(obj) => obj.fresh_ref(f, sub).map(MonoType::record),
            MonoType::Fun(fun) => fun.fresh_ref(f, sub).map(MonoType::fun),
            MonoType::Dict(dict) => dict.fresh_ref(f, sub).map(MonoType::dict),
        }
    }
}

impl Fresh for Tvar {
    fn fresh(self, f: &mut Fresher, sub: &mut TvarMap) -> Self {
        *sub.entry(self).or_insert_with(|| f.fresh())
    }
    fn fresh_ref(&self, f: &mut Fresher, sub: &mut TvarMap) -> Option<Self> {
        Some(*sub.entry(*self).or_insert_with(|| f.fresh()))
    }
}

impl Fresh for Collection {
    fn fresh_ref(&self, f: &mut Fresher, sub: &mut TvarMap) -> Option<Self> {
        self.arg.fresh_ref(f, sub).map(|arg| Collection {
            collection: self.collection,
            arg,
        })
    }
}

impl Fresh for Dictionary {
    fn fresh_ref(&self, f: &mut Fresher, sub: &mut TvarMap) -> Option<Self> {
        merge(
            &self.key,
            self.key.fresh_ref(f, sub),
            &self.val,
            self.val.fresh_ref(f, sub),
        )
        .map(|(key, val)| Self { key, val })
    }
}

impl Fresh for Record {
    fn fresh(self, f: &mut Fresher, sub: &mut TvarMap) -> Self {
        self.fresh_ref(f, sub).unwrap_or(self)
    }
    fn fresh_ref(&self, f: &mut Fresher, sub: &mut TvarMap) -> Option<Self> {
        let mut props = MonoTypeVecMap::new();

        let mut fields = self.fields();
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
            tail.fresh_ref(f, sub).unwrap_or_else(|| tail.clone())
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
        // TODO Should optimize when no variables needs freshening
        Some(match r {
            MonoType::Record(b) => (*b).clone(),
            _ => Record::Empty,
        })
    }
}

impl<T> Fresh for Property<T>
where
    T: Clone,
{
    fn fresh(self, f: &mut Fresher, sub: &mut TvarMap) -> Self {
        Property {
            k: self.k,
            v: self.v.fresh(f, sub),
        }
    }
    fn fresh_ref(&self, f: &mut Fresher, sub: &mut TvarMap) -> Option<Self> {
        self.v.fresh_ref(f, sub).map(|v| Property {
            k: self.k.clone(),
            v,
        })
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
    fn fresh_ref(&self, f: &mut Fresher, sub: &mut TvarMap) -> Option<Self> {
        let Function {
            req,
            opt,
            pipe,
            retn,
        } = self;
        merge4(
            req,
            req.fresh_ref(f, sub),
            opt,
            opt.fresh_ref(f, sub),
            pipe,
            pipe.fresh_ref(f, sub),
            retn,
            retn.fresh_ref(f, sub),
        )
        .map(|(req, opt, pipe, retn)| Function {
            req,
            opt,
            pipe,
            retn,
        })
    }
}

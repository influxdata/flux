//! "Fresh" type variable identifiers.

use std::{collections::BTreeMap, hash::Hash};

use crate::semantic::{
    nodes::Symbol,
    sub::{merge3, merge4, merge_collect},
    types::{
        Array, Function, Label, MonoType, MonoTypeVecMap, PolyType, Property, Record, SemanticMap,
        Tvar, TvarMap,
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
impl<T: Fresh + Clone> Fresh for SemanticMap<Label, T> {
    fn fresh(self, f: &mut Fresher, sub: &mut TvarMap) -> Self {
        self.into_iter()
            .collect::<BTreeMap<_, _>>()
            .into_iter()
            .map(|(s, t)| (s, t.fresh(f, sub)))
            .collect::<Self>()
    }
    fn fresh_ref(&self, f: &mut Fresher, sub: &mut TvarMap) -> Option<Self> {
        merge_collect(
            &mut (),
            self,
            |_, (k, v)| v.fresh_ref(f, sub).map(|v| (k.clone(), v)),
            |_, (k, v)| (k.clone(), v.clone()),
        )
    }
}

#[allow(clippy::implicit_hasher)]
impl<T: Fresh + Clone> Fresh for SemanticMap<Symbol, T> {
    fn fresh(self, f: &mut Fresher, sub: &mut TvarMap) -> Self {
        self.into_iter()
            .collect::<BTreeMap<Symbol, T>>()
            .into_iter()
            .map(|(s, t)| (s, t.fresh(f, sub)))
            .collect::<Self>()
    }
    fn fresh_ref(&self, f: &mut Fresher, sub: &mut TvarMap) -> Option<Self> {
        merge_collect(
            &mut (),
            self,
            |_, (k, v)| v.fresh_ref(f, sub).map(|v| (k.clone(), v)),
            |_, (k, v)| (k.clone(), v.clone()),
        )
    }
}

#[allow(clippy::implicit_hasher)]
impl<T: Fresh + Clone> Fresh for SemanticMap<String, T> {
    fn fresh(self, f: &mut Fresher, sub: &mut TvarMap) -> Self {
        self.into_iter()
            .collect::<BTreeMap<String, T>>()
            .into_iter()
            .map(|(s, t)| (s, t.fresh(f, sub)))
            .collect::<Self>()
    }
    fn fresh_ref(&self, f: &mut Fresher, sub: &mut TvarMap) -> Option<Self> {
        merge_collect(
            &mut (),
            self,
            |_, (k, v)| v.fresh_ref(f, sub).map(|v| (k.clone(), v)),
            |_, (k, v)| (k.clone(), v.clone()),
        )
    }
}

#[allow(clippy::implicit_hasher)]
impl<T: Hash + Ord + Eq + Fresh + Clone, S: Clone> Fresh for SemanticMap<T, S> {
    fn fresh(self, f: &mut Fresher, sub: &mut TvarMap) -> Self {
        self.into_iter()
            .collect::<BTreeMap<T, S>>()
            .into_iter()
            .map(|(t, s)| (t.fresh(f, sub), s))
            .collect::<Self>()
    }
    fn fresh_ref(&self, f: &mut Fresher, sub: &mut TvarMap) -> Option<Self> {
        merge_collect(
            &mut (),
            self,
            |_, (k, v)| k.fresh_ref(f, sub).map(|k| (k, v.clone())),
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
            MonoType::Var(tvr) => tvr.fresh_ref(f, sub).map(MonoType::Var),
            MonoType::Arr(arr) => arr.fresh_ref(f, sub).map(MonoType::arr),
            MonoType::Record(obj) => obj.fresh_ref(f, sub).map(MonoType::record),
            MonoType::Fun(fun) => fun.fresh_ref(f, sub).map(MonoType::fun),
            _ => None,
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

impl Fresh for Array {
    fn fresh(self, f: &mut Fresher, sub: &mut TvarMap) -> Self {
        Array(self.0.fresh(f, sub))
    }
    fn fresh_ref(&self, f: &mut Fresher, sub: &mut TvarMap) -> Option<Self> {
        self.0.fresh_ref(f, sub).map(Array)
    }
}

impl Fresh for Record {
    fn fresh(self, f: &mut Fresher, sub: &mut TvarMap) -> Self {
        self.fresh_ref(f, sub).unwrap_or(self)
    }
    fn fresh_ref(&self, f: &mut Fresher, sub: &mut TvarMap) -> Option<Self> {
        let mut props = MonoTypeVecMap::new();
        let mut extends = false;
        let mut tv = Tvar(0);
        let mut cur = self;
        loop {
            match cur {
                Record::Empty => {
                    break;
                }
                Record::Extension {
                    head,
                    tail: MonoType::Record(b),
                } => {
                    props
                        .entry(head.k.clone())
                        .or_insert_with(Vec::new)
                        .push(head.v.clone());
                    cur = b;
                }
                Record::Extension {
                    head,
                    tail: MonoType::Var(t),
                } => {
                    extends = true;
                    tv = *t;
                    props
                        .entry(head.k.clone())
                        .or_insert_with(Vec::new)
                        .push(head.v.clone());
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

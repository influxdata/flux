use crate::semantic::types::{Array, Function, MonoType, PolyType, Property, Row, Tvar};
use std::collections::{BTreeMap, HashMap};
use std::hash::Hash;

// Fresher returns incrementing type variables
pub struct Fresher(pub u64);

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

pub trait Fresh {
    fn fresh(self, f: &mut Fresher, sub: &mut HashMap<Tvar, Tvar>) -> Self;
}

impl<T: Fresh> Fresh for Vec<T> {
    fn fresh(self, f: &mut Fresher, sub: &mut HashMap<Tvar, Tvar>) -> Self {
        self.into_iter().map(|t| t.fresh(f, sub)).collect::<Self>()
    }
}

impl<T: Fresh> Fresh for Option<T> {
    fn fresh(self, f: &mut Fresher, sub: &mut HashMap<Tvar, Tvar>) -> Self {
        match self {
            None => None,
            Some(t) => Some(t.fresh(f, sub)),
        }
    }
}

impl<T: Fresh> Fresh for HashMap<String, T> {
    fn fresh(self, f: &mut Fresher, sub: &mut HashMap<Tvar, Tvar>) -> Self {
        self.into_iter()
            .collect::<BTreeMap<String, T>>()
            .into_iter()
            .map(|(s, t)| (s, t.fresh(f, sub)))
            .collect::<Self>()
    }
}

impl<T: Hash + Ord + Eq + Fresh, S> Fresh for HashMap<T, S> {
    fn fresh(self, f: &mut Fresher, sub: &mut HashMap<Tvar, Tvar>) -> Self {
        self.into_iter()
            .collect::<BTreeMap<T, S>>()
            .into_iter()
            .map(|(t, s)| (t.fresh(f, sub), s))
            .collect::<Self>()
    }
}

impl<T: Fresh> Fresh for Box<T> {
    fn fresh(self, f: &mut Fresher, sub: &mut HashMap<Tvar, Tvar>) -> Self {
        Box::new((*self).fresh(f, sub))
    }
}

impl Fresh for PolyType {
    fn fresh(self, f: &mut Fresher, sub: &mut HashMap<Tvar, Tvar>) -> Self {
        let expr = self.expr.fresh(f, sub);
        let vars = self.vars.fresh(f, sub);
        let cons = self.cons.fresh(f, sub);
        PolyType { vars, cons, expr }
    }
}

impl Fresh for MonoType {
    fn fresh(self, f: &mut Fresher, sub: &mut HashMap<Tvar, Tvar>) -> Self {
        match self {
            MonoType::Var(tvr) => MonoType::Var(tvr.fresh(f, sub)),
            MonoType::Arr(arr) => MonoType::Arr(arr.fresh(f, sub)),
            MonoType::Row(obj) => MonoType::Row(obj.fresh(f, sub)),
            MonoType::Fun(fun) => MonoType::Fun(fun.fresh(f, sub)),
            _ => self,
        }
    }
}

impl Fresh for Tvar {
    fn fresh(self, f: &mut Fresher, sub: &mut HashMap<Tvar, Tvar>) -> Self {
        *sub.entry(self).or_insert(f.fresh())
    }
}

impl Fresh for Array {
    fn fresh(self, f: &mut Fresher, sub: &mut HashMap<Tvar, Tvar>) -> Self {
        Array(self.0.fresh(f, sub))
    }
}

impl Fresh for Row {
    fn fresh(self, f: &mut Fresher, sub: &mut HashMap<Tvar, Tvar>) -> Self {
        match self {
            Row::Empty => Row::Empty,
            Row::Extension { head, tail } => Row::Extension {
                head: head.fresh(f, sub),
                tail: tail.fresh(f, sub),
            },
        }
    }
}

impl Fresh for Property {
    fn fresh(self, f: &mut Fresher, sub: &mut HashMap<Tvar, Tvar>) -> Self {
        Property {
            k: self.k,
            v: self.v.fresh(f, sub),
        }
    }
}

impl Fresh for Function {
    fn fresh(self, f: &mut Fresher, sub: &mut HashMap<Tvar, Tvar>) -> Self {
        Function {
            req: self.req.fresh(f, sub),
            opt: self.opt.fresh(f, sub),
            pipe: self.pipe.fresh(f, sub),
            retn: self.retn.fresh(f, sub),
        }
    }
}

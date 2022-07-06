//! Substitutions during type inference.
use std::{borrow::Cow, cell::RefCell, collections::BTreeMap, fmt, iter::FusedIterator};

use crate::semantic::{
    fresh::Fresher,
    types::{union, BoundTvar, Error, MonoType, PolyType, SubstitutionMap, Tvar, TvarKinds},
};

use ena::unify::UnifyKey;

/// A substitution defines a function that takes a monotype as input
/// and returns a monotype as output. The output type is interpreted
/// as being equivalent to the input type.
///
/// Substitutions are idempotent. Given a substitution *s* and an input
/// type *x*, we have *s*(*s*(*x*)) = *s*(*x*).
#[derive(Default)]
pub struct Substitution {
    table: RefCell<UnificationTable>,
    // TODO Add `snapshot`/`rollback_to` for `TvarKinds` (like `ena::UnificationTable`) so that
    // modifications can be reverted. Then replace `temporary_generalize` with
    // `snapshot(); generalize(); rollback_to()`
    cons: RefCell<TvarKinds>,
}

impl fmt::Debug for Substitution {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        let mut roots = BTreeMap::new();

        let mut table = self.table.borrow_mut();

        #[derive(Debug)]
        struct Root<T> {
            variables: Vec<Tvar>,
            #[allow(dead_code)]
            value: T,
        }
        for i in 0..table.len() as u32 {
            let i = Tvar::from_index(i);
            let root = table.find(i);
            let root_node = roots.entry(root).or_insert_with(|| Root {
                variables: Vec::new(),
                value: table.probe_value(root),
            });
            if i != root {
                root_node.variables.push(i);
            }
        }

        f.debug_struct("Substitution")
            .field("table", &roots)
            .field("cons", &*self.cons.borrow())
            .finish()
    }
}

/// An implementation of a
/// (Disjoint-set](https://en.wikipedia.org/wiki/Disjoint-set_data_structure) which is used to
/// track which type variables are them same (unified) and which type they have unified to (if any)
type UnificationTable = ena::unify::InPlaceUnificationTable<Tvar>;

impl From<SubstitutionMap> for Substitution {
    /// Derive a substitution from a hash map.
    fn from(values: SubstitutionMap) -> Substitution {
        let mut sub = Substitution::default();
        for (var, typ) in values {
            // Create any variables referenced in the input map
            while var.0 >= sub.table.borrow().len() as u64 {
                sub.fresh();
            }
            sub.union_type(var, typ).unwrap();
        }
        sub
    }
}

impl Substitution {
    /// Return a new empty substitution.
    pub fn new() -> Substitution {
        Substitution::default()
    }

    /// Return a new empty substitution.
    pub fn empty() -> Substitution {
        Substitution::default()
    }

    /// Returns true if no variables has been created by this substitution
    pub fn is_empty(&self) -> bool {
        self.len() == 0
    }

    /// Returns how many variables have been created by this substitution
    pub fn len(&self) -> usize {
        self.table.borrow().len()
    }

    /// Takes a `Substitution` and returns an incremented [`Tvar`].
    pub fn fresh(&self) -> Tvar {
        self.table.borrow_mut().new_key(None)
    }

    /// Prepares `count` type variables for testing
    pub(crate) fn mk_fresh(&self, count: usize) {
        let mut sub = self.table.borrow_mut();
        for _ in 0..count {
            sub.new_key(None);
        }
    }

    pub(crate) fn cons(&mut self) -> &mut TvarKinds {
        self.cons.get_mut()
    }

    /// Apply a substitution to a type variable.
    pub fn apply(&self, tv: Tvar) -> MonoType {
        self.try_apply(tv).unwrap_or(MonoType::Var(tv))
    }

    /// Apply a substitution to a type variable, returning None if there is no substitution for the
    /// variable.
    pub fn try_apply(&self, tv: Tvar) -> Option<MonoType> {
        let mut sub = self.table.borrow_mut();
        match sub.probe_value(tv) {
            Some(typ) => Some(typ),
            None => {
                // If `tv` hasn't been unified with a type we still want to see if it has been
                // unified with any other variables. If it has and it isn't the root we replace
                // `tv` with its root so that `exp.apply(sub).to_string() == actual.apply(sub)`
                // may be equal if they to contain different type variables that has been unified
                // with each other (simplifies debugging even if it isn't strictly necessary for
                // inference itself)
                let root = sub.find(tv);
                if root == tv {
                    None
                } else {
                    Some(MonoType::Var(root))
                }
            }
        }
    }

    /// Returns the "root variable" which is the variable that uniquely identifies a group of
    /// variables that were unified
    pub fn root(&self, tv: Tvar) -> Tvar {
        self.table.borrow_mut().find(tv)
    }

    /// Unifies as a `Tvar` and a `MonoType`, recording the result in the substitution for later
    /// lookup
    pub fn union_type(&mut self, var: Tvar, typ: MonoType) -> Result<(), Error> {
        match typ {
            MonoType::Var(r) => self.union(var, r),
            _ => {
                self.table.borrow_mut().union_value(var, Some(typ.clone()));

                if let Some(kinds) = self.cons().remove(&var) {
                    for kind in &kinds {
                        // The monotype that is being unified with the
                        // tvar must be constrained with the same kinds
                        // as that of the tvar.
                        typ.clone().constrain(*kind, self)?;
                    }
                    if matches!(typ, MonoType::BoundVar(_)) {
                        self.cons().insert(var, kinds);
                    }
                }
            }
        }
        Ok(())
    }

    /// Unifies two `Tvar`s, recording the result in the substitution for later.
    pub fn union(&self, l: Tvar, r: Tvar) {
        self.table.borrow_mut().union(l, r);

        let mut cons = self.cons.borrow_mut();
        // Kind constraints for both type variables
        let kinds = union(
            cons.remove(&l).unwrap_or_default(),
            cons.remove(&r).unwrap_or_default(),
        );
        if !kinds.is_empty() {
            let root = self.root(l);
            cons.insert(root, kinds);
        }
    }
}

/// A type is `Substitutable` if a substitution can be applied to it.
pub trait Substitutable {
    /// Apply a substitution to a type variable.
    fn apply(self, sub: &mut (impl ?Sized + Substituter)) -> Self
    where
        Self: Sized,
    {
        self.visit(sub).unwrap_or(self)
    }

    /// Apply a substitution to a type variable.
    fn apply_mut(&mut self, sub: &mut (impl ?Sized + Substituter))
    where
        Self: Sized,
    {
        if let Some(new) = self.visit(sub) {
            *self = new;
        }
    }

    /// Apply a substitution to a type variable.
    fn apply_cow(&self, sub: &mut (impl ?Sized + Substituter)) -> Cow<'_, Self>
    where
        Self: Clone + Sized,
    {
        match self.visit(sub) {
            Some(t) => Cow::Owned(t),
            None => Cow::Borrowed(self),
        }
    }

    /// Apply a substitution to a type variable. Should return `None` if there was nothing to apply
    /// which allows for optimizations.
    fn visit(&self, sub: &mut (impl ?Sized + Substituter)) -> Option<Self>
    where
        Self: Sized,
    {
        self.walk(sub)
    }

    /// Apply a substitution to a type variable. Should return `None` if there was nothing to apply
    /// which allows for optimizations.
    fn walk(&self, sub: &mut (impl ?Sized + Substituter)) -> Option<Self>
    where
        Self: Sized;

    /// Get all free type variables in a type.
    fn free_vars(&self, sub: &mut Substitution) -> Vec<Tvar>
    where
        Self: Sized,
    {
        let mut vars = Vec::new();
        self.extend_free_vars(&mut vars, sub);
        vars
    }

    /// Get all free type variables in a type.
    fn extend_free_vars(&self, vars: &mut Vec<Tvar>, sub: &mut Substitution)
    where
        Self: Sized,
    {
        struct FreeVars<'a> {
            vars: &'a mut Vec<Tvar>,
            sub: &'a mut Substitution,
        }

        impl Substituter for FreeVars<'_> {
            fn try_apply(&mut self, var: Tvar) -> Option<MonoType> {
                match self.sub.try_apply(var) {
                    Some(typ) => typ.visit(self),
                    None => {
                        if let Err(i) = self.vars.binary_search(&var) {
                            self.vars.insert(i, var);
                        }
                        None
                    }
                }
            }
        }

        self.visit(&mut FreeVars { vars, sub });
    }

    /// Returns `Self` but with "fresh" type variables
    fn fresh(&self, fresher: &mut Fresher) -> Self
    where
        Self: Sized + Clone,
    {
        self.visit(fresher).unwrap_or_else(|| self.clone())
    }
}

impl Substitutable for String {
    fn walk(&self, _sub: &mut (impl ?Sized + Substituter)) -> Option<Self> {
        None
    }
}

impl<T> Substitutable for Box<T>
where
    T: Substitutable,
{
    fn walk(&self, sub: &mut (impl ?Sized + Substituter)) -> Option<Self> {
        T::visit(self, sub).map(Box::new)
    }
}

impl<T> Substitutable for Vec<T>
where
    T: Substitutable + Clone,
{
    fn walk(&self, sub: &mut (impl ?Sized + Substituter)) -> Option<Self> {
        merge_collect(&mut (), self, |_, v| v.visit(sub), |_, v| v.clone())
    }
}

/// Objects from which variable substitutions can be looked up.
pub trait Substituter {
    /// Apply a substitution to a type variable, returning None if there is no substitution for the
    /// variable.
    fn try_apply(&mut self, var: Tvar) -> Option<MonoType>;
    /// Apply a substitution to a bound type variable, returning None if there is no substitution for the
    /// variable.
    fn try_apply_bound(&mut self, var: BoundTvar) -> Option<MonoType> {
        let _ = var;
        None
    }

    /// Apply a substitution to a polytype, returning None if there is no substitution for the
    /// type.
    fn visit_poly_type(&mut self, typ: &PolyType) -> Option<PolyType> {
        typ.walk(self)
    }

    /// Apply a substitution to a type, returning None if there is no substitution for the
    /// type.
    fn visit_type(&mut self, typ: &MonoType) -> Option<MonoType> {
        match *typ {
            MonoType::Var(var) => self.try_apply(var).map(|typ| typ.walk(self).unwrap_or(typ)),
            MonoType::BoundVar(var) => self
                .try_apply_bound(var)
                .map(|typ| typ.walk(self).unwrap_or(typ)),
            _ => typ.walk(self),
        }
    }
}

impl<F> Substituter for F
where
    F: ?Sized + FnMut(Tvar) -> Option<MonoType>,
{
    fn try_apply(&mut self, var: Tvar) -> Option<MonoType> {
        self(var)
    }
}

impl Substituter for SubstitutionMap {
    fn try_apply(&mut self, var: Tvar) -> Option<MonoType> {
        self.get(&var).cloned()
    }
}

impl Substituter for Substitution {
    fn try_apply(&mut self, var: Tvar) -> Option<MonoType> {
        Substitution::try_apply(self, var)
    }
}

pub(crate) struct BindVars<'a> {
    sub: &'a mut dyn Substituter,
    unbound_vars: SubstitutionMap,
}

impl<'a> BindVars<'a> {
    pub fn new(sub: &'a mut dyn Substituter) -> Self {
        Self {
            sub,
            unbound_vars: Default::default(),
        }
    }
}

impl Substituter for BindVars<'_> {
    fn try_apply(&mut self, var: Tvar) -> Option<MonoType> {
        Some(if let Some(typ) = self.sub.try_apply(var) {
            typ
        } else {
            let new_var = BoundTvar(self.unbound_vars.len() as u64);
            self.unbound_vars
                .entry(var)
                .or_insert_with(|| MonoType::BoundVar(new_var))
                .clone()
        })
    }
}

pub(crate) fn apply4<A, B, C, D>(
    a: &A,
    b: &B,
    c: &C,
    d: &D,
    sub: &mut (impl ?Sized + Substituter),
) -> Option<(A, B, C, D)>
where
    A: Substitutable + Clone,
    B: Substitutable + Clone,
    C: Substitutable + Clone,
    D: Substitutable + Clone,
{
    merge4(
        a,
        a.visit(sub),
        b,
        b.visit(sub),
        c,
        c.visit(sub),
        d,
        d.visit(sub),
    )
}

pub(crate) fn apply3<A, B, C>(
    a: &A,
    b: &B,
    c: &C,
    sub: &mut (impl ?Sized + Substituter),
) -> Option<(A, B, C)>
where
    A: Substitutable + Clone,
    B: Substitutable + Clone,
    C: Substitutable + Clone,
{
    merge3(a, a.visit(sub), b, b.visit(sub), c, c.visit(sub))
}

pub(crate) fn apply2<A, B>(a: &A, b: &B, sub: &mut (impl ?Sized + Substituter)) -> Option<(A, B)>
where
    A: Substitutable + Clone,
    B: Substitutable + Clone,
{
    merge(a, a.visit(sub), b, b.visit(sub))
}

#[allow(clippy::too_many_arguments, clippy::type_complexity)]
pub(crate) fn merge4<A: ?Sized, B: ?Sized, C: ?Sized, D: ?Sized>(
    a_original: &A,
    a: Option<A::Owned>,
    b_original: &B,
    b: Option<B::Owned>,
    c_original: &C,
    c: Option<C::Owned>,
    d_original: &D,
    d: Option<D::Owned>,
) -> Option<(A::Owned, B::Owned, C::Owned, D::Owned)>
where
    A: ToOwned,
    B: ToOwned,
    C: ToOwned,
    D: ToOwned,
{
    let a_b_c = merge3(a_original, a, b_original, b, c_original, c);
    merge_fn(
        &(a_original, b_original, c_original),
        |_| {
            (
                a_original.to_owned(),
                b_original.to_owned(),
                c_original.to_owned(),
            )
        },
        a_b_c,
        d_original,
        D::to_owned,
        d,
    )
    .map(|((a, b, c), d)| (a, b, c, d))
}

pub(crate) fn merge3<A: ?Sized, B: ?Sized, C: ?Sized>(
    a_original: &A,
    a: Option<A::Owned>,
    b_original: &B,
    b: Option<B::Owned>,
    c_original: &C,
    c: Option<C::Owned>,
) -> Option<(A::Owned, B::Owned, C::Owned)>
where
    A: ToOwned,
    B: ToOwned,
    C: ToOwned,
{
    let a_b = merge(a_original, a, b_original, b);
    merge_fn(
        &(a_original, b_original),
        |_| (a_original.to_owned(), b_original.to_owned()),
        a_b,
        c_original,
        C::to_owned,
        c,
    )
    .map(|((a, b), c)| (a, b, c))
}

/// Merges two values using `f` if either or both them is `Some(..)`.
/// If both are `None`, `None` is returned.
pub(crate) fn merge<A: ?Sized, B: ?Sized>(
    a_original: &A,
    a: Option<A::Owned>,
    b_original: &B,
    b: Option<B::Owned>,
) -> Option<(A::Owned, B::Owned)>
where
    A: ToOwned,
    B: ToOwned,
{
    merge_fn(a_original, A::to_owned, a, b_original, B::to_owned, b)
}

fn merge_fn<'a, 'b, G, H, A: ?Sized, B: ?Sized, A1, B1>(
    a_original: &'a A,
    g: G,
    a: Option<A1>,
    b_original: &'b B,
    h: H,
    b: Option<B1>,
) -> Option<(A1, B1)>
where
    G: FnOnce(&'a A) -> A1,
    H: FnOnce(&'b B) -> B1,
{
    match (a, b) {
        (Some(a), Some(b)) => Some((a, b)),
        (Some(a), None) => Some((a, h(b_original))),
        (None, Some(b)) => Some((g(a_original), b)),
        (None, None) => None,
    }
}

struct MergeIter<'s, I, F, G, T, S>
where
    S: ?Sized,
{
    types: I,
    clone_types_iter: I,
    action: F,
    converter: G,
    clone_types: usize,
    next: Option<T>,
    pub state: &'s mut S,
}

impl<'s, I, F, G, U, S> Iterator for MergeIter<'s, I, F, G, U, S>
where
    S: ?Sized,
    I: Iterator,
    F: FnMut(&mut S, I::Item) -> Option<U>,
    G: FnMut(&mut S, I::Item) -> U,
{
    type Item = U;
    fn next(&mut self) -> Option<Self::Item> {
        if self.clone_types > 0 {
            self.clone_types -= 1;
            let converter = &mut self.converter;
            let state = &mut self.state;
            self.clone_types_iter.next().map(|e| converter(state, e))
        } else if let Some(typ) = self.next.take() {
            self.clone_types_iter.next();
            Some(typ)
        } else {
            let action = &mut self.action;
            let state = &mut self.state;

            if let Some((i, typ)) = self
                .types
                .by_ref()
                .enumerate()
                .find_map(|(i, typ)| action(state, typ).map(|typ| (i, typ)))
            {
                self.clone_types = i;
                self.next = Some(typ);
            } else {
                self.clone_types = usize::max_value();
            }
            self.next()
        }
    }

    fn size_hint(&self) -> (usize, Option<usize>) {
        self.clone_types_iter.size_hint()
    }
}

impl<I, F, G, U, S> ExactSizeIterator for MergeIter<'_, I, F, G, U, S>
where
    S: ?Sized,
    I: ExactSizeIterator,
    F: FnMut(&mut S, I::Item) -> Option<U>,
    G: FnMut(&mut S, I::Item) -> U,
{
    fn len(&self) -> usize {
        self.clone_types_iter.len()
    }
}

pub(crate) fn merge_collect<I, F, G, U, S, R>(
    state: &mut S,
    types: I,
    action: F,
    converter: G,
) -> Option<R>
where
    S: ?Sized,
    I: IntoIterator,
    I::IntoIter: FusedIterator + Clone,
    F: FnMut(&mut S, I::Item) -> Option<U>,
    G: FnMut(&mut S, I::Item) -> U,
    R: FromIterator<U>,
{
    merge_iter(state, types, action, converter).map(|iter| iter.collect())
}

fn merge_iter<I, F, G, U, S>(
    state: &mut S,
    types: I,
    mut action: F,
    converter: G,
) -> Option<MergeIter<'_, I::IntoIter, F, G, U, S>>
where
    S: ?Sized,
    I: IntoIterator,
    I::IntoIter: FusedIterator + Clone,
    F: FnMut(&mut S, I::Item) -> Option<U>,
    G: FnMut(&mut S, I::Item) -> U,
{
    let mut types = types.into_iter();
    let clone_types_iter = types.clone();
    if let Some((i, typ)) = types
        .by_ref()
        .enumerate()
        .find_map(|(i, typ)| action(state, typ).map(|typ| (i, typ)))
    {
        Some(MergeIter {
            clone_types_iter,
            types,
            action,
            converter,
            clone_types: i,
            next: Some(typ),
            state,
        })
    } else {
        None
    }
}

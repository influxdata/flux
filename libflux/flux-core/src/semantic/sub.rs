//! Substitutions during type inference.
use std::{cell::RefCell, iter::FusedIterator};

use crate::semantic::types::{MonoType, SubstitutionMap, Tvar};

/// A substitution defines a function that takes a monotype as input
/// and returns a monotype as output. The output type is interpreted
/// as being equivalent to the input type.
///
/// Substitutions are idempotent. Given a substitution *s* and an input
/// type *x*, we have *s*(*s*(*x*)) = *s*(*x*).
#[derive(Clone, Debug)]
pub struct Substitution(RefCell<UnificationTable>);

type UnificationTable = ena::unify::InPlaceUnificationTable<Tvar>;

impl From<SubstitutionMap> for Substitution {
    /// Derive a substitution from a hash map.
    fn from(values: SubstitutionMap) -> Substitution {
        let sub = Substitution(RefCell::new(UnificationTable::new()));
        for (var, typ) in values {
            // Create any variables referenced in the input map
            while var.0 >= sub.0.borrow().len() as u64 {
                sub.fresh();
            }
            sub.union_type(var, typ);
        }
        sub
    }
}

impl Default for Substitution {
    fn default() -> Self {
        Self::empty()
    }
}

impl Substitution {
    /// Return a new empty substitution.
    pub fn empty() -> Substitution {
        Substitution(RefCell::new(UnificationTable::new()))
    }

    /// Takes a `Substitution` and returns an incremented [`Tvar`].
    pub fn fresh(&self) -> Tvar {
        self.0.borrow_mut().new_key(None)
    }

    /// Prepares `count` type variables for testing
    #[cfg(test)]
    pub(crate) fn mk_fresh(&self, count: usize) {
        let mut sub = self.0.borrow_mut();
        for _ in 0..count {
            sub.new_key(None);
        }
    }

    /// Returns `true` if the `Substitution` is empty
    pub fn is_empty(&self) -> bool {
        self.0.borrow().len() == 0 // TODO This is not the same with ena
    }

    /// Apply a substitution to a type variable.
    pub fn apply(&self, tv: Tvar) -> MonoType {
        self.try_apply(tv).unwrap_or(MonoType::Var(tv))
    }

    /// Apply a substitution to a type variable, returning None if there is no substitution for the
    /// variable.
    pub fn try_apply(&self, tv: Tvar) -> Option<MonoType> {
        let mut sub = self.0.borrow_mut();
        match sub.probe_value(tv) {
            Some(typ) => Some(typ),
            None => {
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
        self.0.borrow_mut().find(tv)
    }

    /// Unifies as a `Tvar` and a `MonoType`, recording the result in the substitution for later
    /// lookup
    pub fn union_type(&self, l: Tvar, r: MonoType) {
        match r {
            MonoType::Var(r) => self.union(l, r),
            _ => self.0.borrow_mut().union_value(l, Some(r)),
        }
    }

    /// Unifies two `Tvar`s, recording the result in the substitution for later.
    pub fn union(&self, l: Tvar, r: Tvar) {
        self.0.borrow_mut().union(l, r);
    }

    /// Merge two substitutions.
    pub fn merge(self, _with: Substitution) -> Substitution {
        todo!("remove")
        // let applied: SubstitutionMap = self.0.apply(&with);
        // Substitution(applied.into_iter().chain(with.0.into_iter()).collect())
    }
}

/// A type is `Substitutable` if a substitution can be applied to it.
pub trait Substitutable {
    /// Apply a substitution to a type variable.
    fn apply(self, sub: &dyn Substituter) -> Self
    where
        Self: Sized,
    {
        self.apply_ref(sub).unwrap_or(self)
    }

    /// Apply a substitution to a type variable.
    fn apply_mut(&mut self, sub: &dyn Substituter)
    where
        Self: Sized,
    {
        if let Some(new) = self.apply_ref(sub) {
            *self = new;
        }
    }
    /// Apply a substitution to a type variable. Should return `None` if there was nothing to apply
    /// which allows for optimizations.
    fn apply_ref(&self, sub: &dyn Substituter) -> Option<Self>
    where
        Self: Sized;
    /// Get all free type variables in a type.
    fn free_vars(&self) -> Vec<Tvar>;
}

/// Objects from which variable substitutions can be looked up.
pub trait Substituter {
    /// Apply a substitution to a type variable, returning None if there is no substitution for the
    /// variable.
    fn try_apply(&self, var: Tvar) -> Option<MonoType>;
}

impl<F> Substituter for F
where
    F: ?Sized + Fn(Tvar) -> Option<MonoType>,
{
    fn try_apply(&self, var: Tvar) -> Option<MonoType> {
        self(var)
    }
}

impl Substituter for SubstitutionMap {
    fn try_apply(&self, var: Tvar) -> Option<MonoType> {
        self.get(&var).cloned()
    }
}

impl Substituter for Substitution {
    fn try_apply(&self, var: Tvar) -> Option<MonoType> {
        Substitution::try_apply(self, var)
    }
}

pub(crate) fn apply4<A, B, C, D>(
    a: &A,
    b: &B,
    c: &C,
    d: &D,
    sub: &dyn Substituter,
) -> Option<(A, B, C, D)>
where
    A: Substitutable + Clone,
    B: Substitutable + Clone,
    C: Substitutable + Clone,
    D: Substitutable + Clone,
{
    merge4(
        a,
        a.apply_ref(sub),
        b,
        b.apply_ref(sub),
        c,
        c.apply_ref(sub),
        d,
        d.apply_ref(sub),
    )
}

pub(crate) fn apply2<A, B>(a: &A, b: &B, sub: &dyn Substituter) -> Option<(A, B)>
where
    A: Substitutable + Clone,
    B: Substitutable + Clone,
{
    merge(a, a.apply_ref(sub), b, b.apply_ref(sub))
}

#[allow(clippy::too_many_arguments, clippy::type_complexity)]
fn merge4<A: ?Sized, B: ?Sized, C: ?Sized, D: ?Sized>(
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

fn merge3<A: ?Sized, B: ?Sized, C: ?Sized>(
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
fn merge<A: ?Sized, B: ?Sized>(
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
                self.next()
            } else {
                self.clone_types = usize::max_value();
                self.next()
            }
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
    R: std::iter::FromIterator<U>,
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

//! Substitutions during type inference.
use std::iter::FusedIterator;

use crate::semantic::types::{MonoType, SubstitutionMap, Tvar};

/// A substitution defines a function that takes a monotype as input
/// and returns a monotype as output. The output type is interpreted
/// as being equivalent to the input type.
///
/// Substitutions are idempotent. Given a substitution *s* and an input
/// type *x*, we have *s*(*s*(*x*)) = *s*(*x*).
#[derive(Debug, PartialEq)]
pub struct Substitution(SubstitutionMap);

impl From<SubstitutionMap> for Substitution {
    /// Derive a substitution from a hash map.
    fn from(values: SubstitutionMap) -> Substitution {
        Substitution(values)
    }
}

// The `allow` attribute below is a side effect of the orphan impl rule as
// well as the implicit_hasher lint. For more info, see
// https://github.com/rust-lang/rfcs/issues/1856
#[allow(clippy::implicit_hasher)]
impl From<Substitution> for SubstitutionMap {
    /// Derive a hash map from a substitution.
    fn from(sub: Substitution) -> SubstitutionMap {
        sub.0
    }
}

impl Substitution {
    /// Return a new empty substitution.
    pub fn empty() -> Substitution {
        Substitution(SubstitutionMap::new())
    }

    /// Apply a substitution to a type variable.
    pub fn apply(&self, tv: Tvar) -> MonoType {
        self.try_apply(tv).unwrap_or(MonoType::Var(tv))
    }

    /// Apply a substitution to a type variable, returning None if there is no substitution for the
    /// variable.
    pub fn try_apply(&self, tv: Tvar) -> Option<MonoType> {
        self.0.get(&tv).cloned()
    }

    /// Merge two substitutions.
    pub fn merge(self, with: Substitution) -> Substitution {
        let applied: SubstitutionMap = self.0.apply(&with);
        Substitution(applied.into_iter().chain(with.0.into_iter()).collect())
    }
}

/// A type is `Substitutable` if a substitution can be applied to it.
pub trait Substitutable {
    /// Apply a substitution to a type variable.
    fn apply(self, sub: &Substitution) -> Self
    where
        Self: Sized,
    {
        self.apply_ref(sub).unwrap_or(self)
    }
    /// Apply a substitution to a type variable. Should return `None` if there was nothing to apply
    /// which allows for optimizations.
    fn apply_ref(&self, sub: &Substitution) -> Option<Self>
    where
        Self: Sized;
    /// Get all free type variables in a type.
    fn free_vars(&self) -> Vec<Tvar>;
}

pub(crate) fn apply4<A, B, C, D>(
    a: &A,
    b: &B,
    c: &C,
    d: &D,
    sub: &Substitution,
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
        |a, b, c, d| (a, b, c, d),
    )
}

pub(crate) fn apply2<A, B>(a: &A, b: &B, sub: &Substitution) -> Option<(A, B)>
where
    A: Substitutable + Clone,
    B: Substitutable + Clone,
{
    merge(a, a.apply_ref(sub), b, b.apply_ref(sub), |a, b| (a, b))
}

#[allow(clippy::too_many_arguments)]
fn merge4<F, A: ?Sized, B: ?Sized, C: ?Sized, D: ?Sized, R>(
    a_original: &A,
    a: Option<A::Owned>,
    b_original: &B,
    b: Option<B::Owned>,
    c_original: &C,
    c: Option<C::Owned>,
    d_original: &D,
    d: Option<D::Owned>,
    action: F,
) -> Option<R>
where
    A: ToOwned,
    B: ToOwned,
    C: ToOwned,
    D: ToOwned,
    F: FnOnce(A::Owned, B::Owned, C::Owned, D::Owned) -> R,
{
    let a_b_c = merge3(a_original, a, b_original, b, c_original, c, |a, b, c| {
        (a, b, c)
    });
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
        |(a, b, c), d| action(a, b, c, d),
    )
}

fn merge3<F, A: ?Sized, B: ?Sized, C: ?Sized, R>(
    a_original: &A,
    a: Option<A::Owned>,
    b_original: &B,
    b: Option<B::Owned>,
    c_original: &C,
    c: Option<C::Owned>,
    f: F,
) -> Option<R>
where
    A: ToOwned,
    B: ToOwned,
    C: ToOwned,
    F: FnOnce(A::Owned, B::Owned, C::Owned) -> R,
{
    let a_b = merge(a_original, a, b_original, b, |a, b| (a, b));
    merge_fn(
        &(a_original, b_original),
        |_| (a_original.to_owned(), b_original.to_owned()),
        a_b,
        c_original,
        C::to_owned,
        c,
        |(a, b), c| f(a, b, c),
    )
}

/// Merges two values using `f` if either or both them is `Some(..)`.
/// If both are `None`, `None` is returned.
fn merge<F, A: ?Sized, B: ?Sized, R>(
    a_original: &A,
    a: Option<A::Owned>,
    b_original: &B,
    b: Option<B::Owned>,
    f: F,
) -> Option<R>
where
    A: ToOwned,
    B: ToOwned,
    F: FnOnce(A::Owned, B::Owned) -> R,
{
    merge_fn(a_original, A::to_owned, a, b_original, B::to_owned, b, f)
}

fn merge_fn<'a, 'b, F, G, H, A: ?Sized, B: ?Sized, A1, B1, R>(
    a_original: &'a A,
    g: G,
    a: Option<A1>,
    b_original: &'b B,
    h: H,
    b: Option<B1>,
    merger: F,
) -> Option<R>
where
    F: FnOnce(A1, B1) -> R,
    G: FnOnce(&'a A) -> A1,
    H: FnOnce(&'b B) -> B1,
{
    match (a, b) {
        (Some(a), Some(b)) => Some(merger(a, b)),
        (Some(a), None) => Some(merger(a, h(b_original))),
        (None, Some(b)) => Some(merger(g(a_original), b)),
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

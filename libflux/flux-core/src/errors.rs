/// An error type which can represent multiple errors.
use std::{
    any::Any,
    error::Error as StdError,
    fmt,
    ops::{Index, IndexMut},
    slice, vec,
};

use codespan_reporting::diagnostic;
use derive_more::Display;

use crate::{
    ast,
    semantic::{
        sub::{Substitutable, Substituter},
        types::Tvar,
    },
};

#[derive(Clone, Debug, Eq, PartialEq, Hash)]
pub struct Errors<T> {
    errors: Vec<T>,
}

impl<T> Default for Errors<T> {
    fn default() -> Self {
        Errors::new()
    }
}

impl<T> Errors<T> {
    /// Creates a new, empty `Errors` instance.
    pub fn new() -> Errors<T> {
        Errors::from(Vec::new())
    }

    /// Returns true if `self` contains any errors
    pub fn has_errors(&self) -> bool {
        !self.is_empty()
    }

    /// The number of errors in the error list
    pub fn len(&self) -> usize {
        self.errors.len()
    }

    pub fn is_empty(&self) -> bool {
        self.errors.is_empty()
    }

    /// Adds an error to `self`
    pub fn push(&mut self, t: T) {
        self.errors.push(t);
    }

    /// Pops and error off the error list
    pub fn pop(&mut self) -> Option<T> {
        self.errors.pop()
    }

    pub fn iter(&self) -> slice::Iter<T> {
        self.errors.iter()
    }

    pub fn drain(
        &mut self,
        range: impl std::ops::RangeBounds<usize>,
    ) -> impl Iterator<Item = T> + '_ {
        self.errors.drain(range)
    }
}

impl<T> Index<usize> for Errors<T> {
    type Output = T;
    fn index(&self, index: usize) -> &T {
        &self.errors[index]
    }
}

impl<T> IndexMut<usize> for Errors<T> {
    fn index_mut(&mut self, index: usize) -> &mut T {
        &mut self.errors[index]
    }
}

impl<T> Extend<T> for Errors<T> {
    fn extend<Iter: IntoIterator<Item = T>>(&mut self, iter: Iter) {
        self.errors.extend(iter);
    }
}

impl<T> From<T> for Errors<T> {
    fn from(err: T) -> Errors<T> {
        Errors { errors: vec![err] }
    }
}

impl<T> From<Vec<T>> for Errors<T> {
    fn from(errors: Vec<T>) -> Errors<T> {
        Errors { errors }
    }
}

impl<T> FromIterator<T> for Errors<T> {
    fn from_iter<Iter: IntoIterator<Item = T>>(iter: Iter) -> Errors<T> {
        Errors {
            errors: iter.into_iter().collect(),
        }
    }
}

impl<T> From<Errors<T>> for Vec<T> {
    fn from(errors: Errors<T>) -> Vec<T> {
        errors.errors
    }
}

impl<T> IntoIterator for Errors<T> {
    type Item = T;

    type IntoIter = vec::IntoIter<T>;

    fn into_iter(self) -> vec::IntoIter<T> {
        self.errors.into_iter()
    }
}

impl<'a, T> IntoIterator for &'a Errors<T> {
    type Item = &'a T;

    type IntoIter = slice::Iter<'a, T>;

    fn into_iter(self) -> slice::Iter<'a, T> {
        self.errors.iter()
    }
}

impl<'a, T> IntoIterator for &'a mut Errors<T> {
    type Item = &'a mut T;

    type IntoIter = slice::IterMut<'a, T>;

    fn into_iter(self) -> slice::IterMut<'a, T> {
        self.errors.iter_mut()
    }
}

impl<T: fmt::Display> fmt::Display for Errors<T> {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        for (i, error) in self.errors.iter().enumerate() {
            write!(f, "{}", error)?;
            // Errors are assumed to not have a newline at the end so we add one to keep errors on
            // separate lines and one to space them out
            if i + 1 != self.errors.len() {
                writeln!(f)?;
                writeln!(f)?;
            }
        }
        Ok(())
    }
}

impl<T: fmt::Display + fmt::Debug + Any> StdError for Errors<T> {
    fn description(&self) -> &str {
        "Errors"
    }
}

/// An error with an attached location
#[derive(Debug, Display, PartialEq)]
#[display(fmt = "error {}: {}", location, error)]
pub struct Located<E> {
    /// The location where the error occured
    pub location: ast::SourceLocation,
    /// The error itself
    pub error: E,
}

impl<T: StdError> StdError for Located<T> {
    fn source(&self) -> Option<&(dyn StdError + 'static)> {
        self.error.source()
    }
}

/// Constructs a located error
pub fn located<E>(location: ast::SourceLocation, error: E) -> Located<E> {
    Located { location, error }
}

impl<E> Substitutable for Located<E>
where
    E: Substitutable,
{
    fn apply_ref(&self, sub: &dyn Substituter) -> Option<Self> {
        self.error.apply_ref(sub).map(|error| Located {
            location: self.location.clone(),
            error,
        })
    }
    fn free_vars(&self) -> Vec<Tvar> {
        self.error.free_vars()
    }
}

pub(crate) trait AsDiagnostic {
    fn as_diagnostic(&self, source: &dyn crate::semantic::Source) -> diagnostic::Diagnostic<()>;
}

impl<E> AsDiagnostic for Located<E>
where
    E: AsDiagnostic,
{
    fn as_diagnostic(&self, source: &dyn crate::semantic::Source) -> diagnostic::Diagnostic<()> {
        self.error
            .as_diagnostic(source)
            .with_labels(vec![diagnostic::Label::primary(
                (),
                source.codespan_range(&self.location),
            )])
    }
}

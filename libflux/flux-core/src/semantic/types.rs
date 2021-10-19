//! Semantic representations of types.

use crate::semantic::fresh::{Fresh, Fresher};
use crate::semantic::sub::{
    apply2, apply4, merge_collect, Substitutable, Substituter, Substitution,
};
use derive_more::Display;
use std::fmt::Write;

use std::{
    cmp,
    collections::{BTreeMap, BTreeSet, HashMap},
    fmt,
};

/// For use in generics where the specific type of map is not mentioned.
pub type SemanticMap<K, V> = BTreeMap<K, V>;
#[allow(missing_docs)]
pub type SemanticMapIter<'a, K, V> = std::collections::btree_map::Iter<'a, K, V>;

/// A type scheme that quantifies the free variables of a monotype.
#[derive(Debug, Clone)]
pub struct PolyType {
    /// List of the free variables within the monotypes.
    pub vars: Vec<Tvar>,
    /// The list of kind constraints on any of the free variables.
    pub cons: TvarKinds,
    /// The underlying monotype.
    pub expr: MonoType,
}

/// Map of identifier to a polytype that preserves a sorted order when iterating.
pub type PolyTypeMap = SemanticMap<String, PolyType>;
/// Nested map of polytypes that preserves a sorted order when iterating
pub type PolyTypeMapMap = SemanticMap<String, SemanticMap<String, PolyType>>;

/// Alias the maplit literal construction macro so we can specify the type here.
#[macro_export]
macro_rules! semantic_map {
    ( $($x:tt)* ) => ( maplit::btreemap!( $($x)* ) );
}

impl fmt::Display for PolyType {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        if self.cons.is_empty() {
            self.expr.fmt(f)
        } else {
            write!(
                f,
                "{} where {}",
                self.expr,
                PolyType::display_constraints(&self.cons),
            )
        }
    }
}

impl PartialEq for PolyType {
    fn eq(&self, poly: &Self) -> bool {
        let a = self.max_tvar();
        let b = poly.max_tvar();

        let max = if a > b { a } else { b };
        let max = max.map(|t| t.0).unwrap_or_default();

        let mut f = Fresher::from(max + 1);
        let mut g = Fresher::from(max + 1);

        let mut a = self.clone().fresh(&mut f, &mut TvarMap::new());
        let mut b = poly.clone().fresh(&mut g, &mut TvarMap::new());

        a.vars.sort();
        b.vars.sort();

        for kinds in a.cons.values_mut() {
            kinds.sort();
        }
        for kinds in b.cons.values_mut() {
            kinds.sort();
        }

        a.vars == b.vars && a.cons == b.cons && a.expr == b.expr
    }
}

impl Substitutable for PolyType {
    fn apply_ref(&self, sub: &dyn Substituter) -> Option<Self> {
        // `vars` defines new distinct variables for `expr` so any substitutions applied on a
        // variable named the same must not be applied in `expr`
        self.expr
            .apply_ref(&|var| {
                if self.vars.contains(&var) {
                    None
                } else {
                    sub.try_apply(var)
                }
            })
            .map(|expr| PolyType {
                vars: self.vars.clone(),
                cons: self.cons.clone(),
                expr,
            })
    }
    fn free_vars(&self) -> Vec<Tvar> {
        minus(&self.vars, self.expr.free_vars())
    }
}

impl MaxTvar for [Tvar] {
    fn max_tvar(&self) -> Option<Tvar> {
        self.iter().max().cloned()
    }
}

impl MaxTvar for [Option<Tvar>] {
    fn max_tvar(&self) -> Option<Tvar> {
        self.iter().max().and_then(|t| *t)
    }
}

impl MaxTvar for PolyType {
    fn max_tvar(&self) -> Option<Tvar> {
        [self.vars.max_tvar(), self.expr.max_tvar()].max_tvar()
    }
}

impl PolyType {
    fn display_constraints(cons: &TvarKinds) -> String {
        cons.iter()
            // A BTree produces a sorted iterator for
            // deterministic display output
            .collect::<BTreeMap<_, _>>()
            .iter()
            .map(|(&&tv, &kinds)| format!("{}: {}", tv, PolyType::display_kinds(kinds)))
            .collect::<Vec<_>>()
            .join(", ")
    }
    fn display_kinds(kinds: &[Kind]) -> String {
        kinds
            .iter()
            // Sort kinds with BTree
            .collect::<BTreeSet<_>>()
            .iter()
            .map(|x| x.to_string())
            .collect::<Vec<_>>()
            .join(" + ")
    }
    /// Produces a `PolyType` where the type variables have been normalized to start at 0
    /// (i.e. A), instead of whatever type variables are present in the orginal.
    ///
    /// Useful for pretty printing the type in error messages.
    pub fn normal(&self) -> PolyType {
        self.clone()
            .fresh(&mut Fresher::from(0), &mut TvarMap::new())
    }
}

/// Helper function that concatenates two vectors into a single vector while removing duplicates.
pub(crate) fn union<T: PartialEq>(mut vars: Vec<T>, mut with: Vec<T>) -> Vec<T> {
    with.retain(|tv| !vars.contains(tv));
    vars.append(&mut with);
    vars
}

/// Helper function that removes all elements in `vars` from `from`.
pub(crate) fn minus<T: PartialEq>(vars: &[T], mut from: Vec<T>) -> Vec<T> {
    from.retain(|tv| !vars.contains(tv));
    from
}

/// Errors that can be returned during type inference.
/// (Note that these error messages are read by end users.
/// This should be kept in mind when returning one of these errors.)
#[derive(Debug, PartialEq)]
#[allow(missing_docs)]
pub enum Error {
    CannotUnify {
        exp: MonoType,
        act: MonoType,
    },
    CannotConstrain {
        exp: Kind,
        act: MonoType,
    },
    OccursCheck(Tvar, MonoType),
    MissingLabel(String),
    ExtraLabel(String),
    CannotUnifyLabel {
        lab: String,
        exp: MonoType,
        act: MonoType,
    },
    MissingArgument(String),
    ExtraArgument(String),
    CannotUnifyArgument(String, Box<Error>),
    CannotUnifyReturn {
        exp: MonoType,
        act: MonoType,
    },
    MissingPipeArgument,
    MultiplePipeArguments {
        exp: String,
        act: String,
    },
}

impl fmt::Display for Error {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        let mut fresh = Fresher::from(0);
        match self {
            Error::CannotUnify { exp, act } => write!(
                f,
                "expected {} but found {}",
                exp.clone().fresh(&mut fresh, &mut TvarMap::new()),
                act.clone().fresh(&mut fresh, &mut TvarMap::new()),
            ),
            Error::CannotConstrain { exp, act } => write!(
                f,
                "{} is not {}",
                act.clone().fresh(&mut fresh, &mut TvarMap::new()),
                exp,
            ),
            Error::OccursCheck(tv, ty) => {
                write!(f, "recursive types not supported {} != {}", tv, ty)
            }
            Error::MissingLabel(a) => write!(f, "record is missing label {}", a),
            Error::ExtraLabel(a) => write!(f, "found unexpected label {}", a),
            Error::CannotUnifyLabel { lab, exp, act } => write!(
                f,
                "expected {} but found {} for label {}",
                exp.clone().fresh(&mut fresh, &mut TvarMap::new()),
                act.clone().fresh(&mut fresh, &mut TvarMap::new()),
                lab
            ),
            Error::MissingArgument(x) => write!(f, "missing required argument {}", x),
            Error::ExtraArgument(x) => write!(f, "found unexpected argument {}", x),
            Error::CannotUnifyArgument(x, e) => write!(f, "{} (argument {})", e, x),
            Error::CannotUnifyReturn { exp, act } => write!(
                f,
                "expected {} but found {} for return type",
                exp.clone().fresh(&mut fresh, &mut TvarMap::new()),
                act.clone().fresh(&mut fresh, &mut TvarMap::new())
            ),
            Error::MissingPipeArgument => write!(f, "missing pipe argument"),
            Error::MultiplePipeArguments { exp, act } => {
                write!(f, "expected pipe argument {} but found {}", exp, act)
            }
        }
    }
}

/// Represents a constraint on a type variable to a specific kind (*i.e.*, a type class).
#[derive(Debug, Display, Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Hash)]
#[allow(missing_docs)]
// Kinds are ordered by name so that polytypes are displayed deterministically
pub enum Kind {
    Addable,
    Comparable,
    Divisible,
    Equatable,
    Negatable,
    Nullable,
    Numeric,
    Record,
    Stringable,
    Subtractable,
    Timeable,
}

/// Pointer type used in `MonoType`
pub type Ptr<T> = Box<T>;

/// Represents a Flux type. The type may be unknown, represented as a type variable,
/// or may be a known concrete type.
#[derive(Debug, Display, Clone, PartialEq, Serialize)]
#[allow(missing_docs)]
pub enum MonoType {
    #[display(fmt = "bool")]
    Bool,
    #[display(fmt = "int")]
    Int,
    #[display(fmt = "uint")]
    Uint,
    #[display(fmt = "float")]
    Float,
    #[display(fmt = "string")]
    String,
    #[display(fmt = "duration")]
    Duration,
    #[display(fmt = "time")]
    Time,
    #[display(fmt = "regexp")]
    Regexp,
    #[display(fmt = "bytes")]
    Bytes,
    #[display(fmt = "{}", _0)]
    Var(Tvar),
    #[display(fmt = "{}", _0)]
    Arr(Ptr<Array>),
    #[display(fmt = "{}", _0)]
    Dict(Ptr<Dictionary>),
    #[display(fmt = "{}", _0)]
    Record(Ptr<Record>),
    #[display(fmt = "{}", _0)]
    Fun(Ptr<Function>),
    #[display(fmt = "{}", _0)]
    Vector(Ptr<Vector>),
}

/// An ordered map of string identifiers to monotypes.
pub type MonoTypeMap = SemanticMap<String, MonoType>;
#[allow(missing_docs)]
pub type MonoTypeVecMap = SemanticMap<String, Vec<MonoType>>;
#[allow(missing_docs)]
type RefMonoTypeVecMap<'a> = HashMap<&'a String, Vec<&'a MonoType>>;

impl Substitutable for MonoType {
    fn apply_ref(&self, sub: &dyn Substituter) -> Option<Self> {
        match self {
            MonoType::Bool
            | MonoType::Int
            | MonoType::Uint
            | MonoType::Float
            | MonoType::String
            | MonoType::Duration
            | MonoType::Time
            | MonoType::Regexp
            | MonoType::Bytes => None,
            MonoType::Var(tvr) => sub.try_apply(*tvr).map(|new| {
                if MonoType::Var(*tvr) == new {
                    new
                } else {
                    new.apply(sub)
                }
            }),
            MonoType::Arr(arr) => arr.apply_ref(sub).map(MonoType::arr),
            MonoType::Vector(vector) => vector.apply_ref(sub).map(MonoType::vector),
            MonoType::Dict(dict) => dict.apply_ref(sub).map(MonoType::dict),
            MonoType::Record(obj) => obj.apply_ref(sub).map(MonoType::record),
            MonoType::Fun(fun) => fun.apply_ref(sub).map(MonoType::fun),
        }
    }
    fn free_vars(&self) -> Vec<Tvar> {
        match self {
            MonoType::Bool
            | MonoType::Int
            | MonoType::Uint
            | MonoType::Float
            | MonoType::String
            | MonoType::Duration
            | MonoType::Time
            | MonoType::Regexp
            | MonoType::Bytes => Vec::new(),
            MonoType::Var(tvr) => vec![*tvr],
            MonoType::Arr(arr) => arr.free_vars(),
            MonoType::Vector(vector) => vector.free_vars(),
            MonoType::Dict(dict) => dict.free_vars(),
            MonoType::Record(obj) => obj.free_vars(),
            MonoType::Fun(fun) => fun.free_vars(),
        }
    }
}

impl MaxTvar for MonoType {
    fn max_tvar(&self) -> Option<Tvar> {
        match self {
            MonoType::Bool
            | MonoType::Int
            | MonoType::Uint
            | MonoType::Float
            | MonoType::String
            | MonoType::Duration
            | MonoType::Time
            | MonoType::Regexp
            | MonoType::Bytes => None,
            MonoType::Var(tvr) => tvr.max_tvar(),
            MonoType::Arr(arr) => arr.max_tvar(),
            MonoType::Vector(vector) => vector.max_tvar(),
            MonoType::Dict(dict) => dict.max_tvar(),
            MonoType::Record(obj) => obj.max_tvar(),
            MonoType::Fun(fun) => fun.max_tvar(),
        }
    }
}

impl From<Tvar> for MonoType {
    fn from(a: Tvar) -> MonoType {
        MonoType::Var(a)
    }
}

impl From<Array> for MonoType {
    fn from(a: Array) -> MonoType {
        MonoType::Arr(Ptr::new(a))
    }
}

impl From<Dictionary> for MonoType {
    fn from(d: Dictionary) -> MonoType {
        MonoType::Dict(Ptr::new(d))
    }
}

impl From<Record> for MonoType {
    fn from(r: Record) -> MonoType {
        MonoType::Record(Ptr::new(r))
    }
}

impl From<Function> for MonoType {
    fn from(f: Function) -> MonoType {
        MonoType::Fun(Ptr::new(f))
    }
}

impl MonoType {
    /// Creates an array type
    pub fn arr(a: impl Into<Ptr<Array>>) -> Self {
        Self::Arr(a.into())
    }

    /// Creates a vector type
    pub fn vector(v: impl Into<Ptr<Vector>>) -> Self {
        Self::Vector(v.into())
    }

    /// Creates a dictionary type
    pub fn dict(d: impl Into<Ptr<Dictionary>>) -> Self {
        Self::Dict(d.into())
    }

    /// Creates a function type
    pub fn fun(f: impl Into<Ptr<Function>>) -> Self {
        Self::Fun(f.into())
    }

    /// Creates a record type
    pub fn record(r: impl Into<Ptr<Record>>) -> Self {
        Self::Record(r.into())
    }

    /// Performs unification on the type with another type.
    /// If successful, results in a solution to the unification problem,
    /// in the form of a substitution. If there is no solution to the
    /// unification problem then unification fails and an error is reported.
    pub fn unify(
        self, // self represents the expected type
        actual: Self,
        cons: &mut TvarKinds,
        sub: &mut Substitution,
    ) -> Result<(), Error> {
        eprintln!("Unify {} <=> {}", self, actual);
        match (self, actual) {
            (MonoType::Bool, MonoType::Bool)
            | (MonoType::Int, MonoType::Int)
            | (MonoType::Uint, MonoType::Uint)
            | (MonoType::Float, MonoType::Float)
            | (MonoType::String, MonoType::String)
            | (MonoType::Duration, MonoType::Duration)
            | (MonoType::Time, MonoType::Time)
            | (MonoType::Regexp, MonoType::Regexp)
            | (MonoType::Bytes, MonoType::Bytes) => Ok(()),
            (MonoType::Var(tv), MonoType::Var(tv2)) => {
                match (sub.try_apply(tv), sub.try_apply(tv2)) {
                    (Some(self_), Some(actual)) => self_.unify(actual, cons, sub),
                    (Some(self_), None) => self_.unify(MonoType::Var(tv2), cons, sub),
                    (None, Some(actual)) => MonoType::Var(tv).unify(actual, cons, sub),
                    (None, None) => tv.unify(MonoType::Var(tv2), cons, sub),
                }
            }
            (MonoType::Var(tv), t) => match sub.try_apply(tv) {
                Some(typ) => typ.unify(t, cons, sub),
                None => tv.unify(t, cons, sub),
            },
            (t, MonoType::Var(tv)) => match sub.try_apply(tv) {
                Some(typ) => t.unify(typ, cons, sub),
                None => tv.unify(t, cons, sub),
            },
            (MonoType::Arr(t), MonoType::Arr(s)) => t.unify(*s, cons, sub),
            (MonoType::Vector(t), MonoType::Vector(s)) => t.unify(*s, cons, sub),
            (MonoType::Dict(t), MonoType::Dict(s)) => t.unify(*s, cons, sub),
            (MonoType::Record(t), MonoType::Record(s)) => t.unify(*s, cons, sub),
            (MonoType::Fun(t), MonoType::Fun(s)) => t.unify(*s, cons, sub),
            (exp, act) => Err(Error::CannotUnify { exp, act }),
        }
    }

    /// Validates that the current type meets the constraints of the specified kind.
    pub fn constrain(self, with: Kind, cons: &mut TvarKinds) -> Result<(), Error> {
        match self {
            MonoType::Bool => match with {
                Kind::Equatable | Kind::Nullable | Kind::Stringable => Ok(()),
                _ => Err(Error::CannotConstrain {
                    act: self,
                    exp: with,
                }),
            },
            MonoType::Int => match with {
                Kind::Addable
                | Kind::Subtractable
                | Kind::Divisible
                | Kind::Numeric
                | Kind::Comparable
                | Kind::Equatable
                | Kind::Nullable
                | Kind::Stringable
                | Kind::Negatable => Ok(()),
                _ => Err(Error::CannotConstrain {
                    act: self,
                    exp: with,
                }),
            },
            MonoType::Uint => match with {
                Kind::Addable
                | Kind::Subtractable
                | Kind::Divisible
                | Kind::Numeric
                | Kind::Comparable
                | Kind::Equatable
                | Kind::Nullable
                | Kind::Stringable
                | Kind::Negatable => Ok(()),
                _ => Err(Error::CannotConstrain {
                    act: self,
                    exp: with,
                }),
            },
            MonoType::Float => match with {
                Kind::Addable
                | Kind::Subtractable
                | Kind::Divisible
                | Kind::Numeric
                | Kind::Comparable
                | Kind::Equatable
                | Kind::Nullable
                | Kind::Stringable
                | Kind::Negatable => Ok(()),
                _ => Err(Error::CannotConstrain {
                    act: self,
                    exp: with,
                }),
            },
            MonoType::String => match with {
                Kind::Addable
                | Kind::Comparable
                | Kind::Equatable
                | Kind::Nullable
                | Kind::Stringable => Ok(()),
                _ => Err(Error::CannotConstrain {
                    act: self,
                    exp: with,
                }),
            },
            MonoType::Duration => match with {
                Kind::Comparable
                | Kind::Equatable
                | Kind::Nullable
                | Kind::Negatable
                | Kind::Stringable
                | Kind::Timeable => Ok(()),
                _ => Err(Error::CannotConstrain {
                    act: self,
                    exp: with,
                }),
            },
            MonoType::Time => match with {
                Kind::Comparable
                | Kind::Equatable
                | Kind::Nullable
                | Kind::Timeable
                | Kind::Stringable => Ok(()),
                _ => Err(Error::CannotConstrain {
                    act: self,
                    exp: with,
                }),
            },
            MonoType::Regexp => Err(Error::CannotConstrain {
                act: self,
                exp: with,
            }),
            MonoType::Bytes => match with {
                Kind::Equatable => Ok(()),
                _ => Err(Error::CannotConstrain {
                    act: self,
                    exp: with,
                }),
            },
            MonoType::Var(tvr) => {
                tvr.constrain(with, cons);
                Ok(())
            }
            MonoType::Arr(arr) => arr.constrain(with, cons),
            MonoType::Vector(vector) => vector.constrain(with, cons),
            MonoType::Dict(dict) => dict.constrain(with, cons),
            MonoType::Record(obj) => obj.constrain(with, cons),
            MonoType::Fun(fun) => fun.constrain(with, cons),
        }
    }

    fn contains(&self, tv: Tvar) -> bool {
        match self {
            MonoType::Bool
            | MonoType::Int
            | MonoType::Uint
            | MonoType::Float
            | MonoType::String
            | MonoType::Duration
            | MonoType::Time
            | MonoType::Regexp
            | MonoType::Bytes => false,
            MonoType::Var(tvr) => tv == *tvr,
            MonoType::Arr(arr) => arr.contains(tv),
            MonoType::Vector(vector) => vector.contains(tv),
            MonoType::Dict(dict) => dict.contains(tv),
            MonoType::Record(row) => row.contains(tv),
            MonoType::Fun(fun) => fun.contains(tv),
        }
    }
}

/// `Tvar` stands for *type variable*.
/// A type variable holds an unknown type, before type inference.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, PartialOrd, Ord, Serialize)]
pub struct Tvar(pub u64); // TODO u32 to match ena?

impl ena::unify::UnifyKey for Tvar {
    type Value = Option<MonoType>;
    fn index(&self) -> u32 {
        self.0 as u32
    }
    fn from_index(u: u32) -> Self {
        Self(From::from(u))
    }
    fn tag() -> &'static str {
        "Tvar"
    }
}
impl ena::unify::UnifyValue for MonoType {
    type Error = ena::unify::NoError;
    fn unify_values(_value1: &Self, _value2: &Self) -> Result<Self, Self::Error> {
        unreachable!("We should never unify two values with each other within the substitution. If we reach this we did not resolve the variable before unifying")
    }
}

/// A map from type variables to their constraining kinds.
pub type TvarKinds = SemanticMap<Tvar, Vec<Kind>>;
#[allow(missing_docs)]
pub type TvarMap = SemanticMap<Tvar, Tvar>;
#[allow(missing_docs)]
pub type SubstitutionMap = SemanticMap<Tvar, MonoType>;

impl fmt::Display for Tvar {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match self.0 {
            0 => write!(f, "A"),
            1 => write!(f, "B"),
            2 => write!(f, "C"),
            3 => write!(f, "D"),
            4 => write!(f, "E"),
            5 => write!(f, "F"),
            6 => write!(f, "G"),
            7 => write!(f, "H"),
            8 => write!(f, "I"),
            9 => write!(f, "J"),
            _ => write!(f, "t{}", self.0),
        }
    }
}

impl MaxTvar for Tvar {
    fn max_tvar(&self) -> Option<Tvar> {
        Some(*self)
    }
}

impl Tvar {
    fn unify(
        self,
        with: MonoType,
        cons: &mut TvarKinds,
        sub: &mut Substitution,
    ) -> Result<(), Error> {
        match with {
            MonoType::Var(tv) => {
                if self == tv {
                    // The empty substitution will always
                    // unify a type variable with itself.
                    Ok(())
                } else {
                    // Unify two distinct type variables.
                    // This will update the kind constraints
                    // associated with these type variables.
                    self.unify_with_tvar(tv, cons, sub)
                }
            }
            _ => {
                let with = with.apply(sub);
                eprintln!("Contains {} .. {}", self, with);
                if with.contains(self) {
                    eprintln!("Invalid");
                    // Invalid recursive type
                    Err(Error::OccursCheck(self, with))
                } else {
                    // Unify a type variable with a monotype.
                    // The monotype must satisify any
                    // constraints placed on the type variable.
                    self.unify_with_type(with, cons, sub)
                }
            }
        }
    }

    fn unify_with_tvar(
        self,
        tv: Tvar,
        cons: &mut TvarKinds,
        sub: &mut Substitution,
    ) -> Result<(), Error> {
        sub.union(self, tv);
        // Kind constraints for both type variables
        let kinds = union(
            cons.remove(&self).unwrap_or_default(),
            cons.remove(&tv).unwrap_or_default(),
        );
        if !kinds.is_empty() {
            let root = sub.root(self);
            cons.insert(root, kinds);
        }
        Ok(())
    }

    fn unify_with_type(
        self,
        t: MonoType,
        cons: &mut TvarKinds,
        sub: &mut Substitution,
    ) -> Result<(), Error> {
        sub.union_type(self, t.clone());
        match cons.remove(&self) {
            None => Ok(()),
            Some(kinds) => {
                for kind in kinds {
                    // The monotype that is being unified with the
                    // tvar must be constrained with the same kinds
                    // as that of the tvar.
                    t.clone().constrain(kind, cons)?;
                }
                Ok(())
            }
        }
    }

    fn constrain(self, with: Kind, cons: &mut TvarKinds) {
        match cons.get_mut(&self) {
            Some(kinds) => {
                if !kinds.contains(&with) {
                    kinds.push(with);
                }
            }
            None => {
                cons.insert(self, vec![with]);
            }
        }
    }
}

/// A homogeneous list type.
#[derive(Debug, Display, Clone, PartialEq, Serialize)]
#[display(fmt = "[{}]", _0)]
pub struct Array(pub MonoType);

impl Substitutable for Array {
    fn apply_ref(&self, sub: &dyn Substituter) -> Option<Self> {
        self.0.apply_ref(sub).map(Array)
    }
    fn free_vars(&self) -> Vec<Tvar> {
        self.0.free_vars()
    }
}

impl MaxTvar for Array {
    fn max_tvar(&self) -> Option<Tvar> {
        self.0.max_tvar()
    }
}

impl Array {
    // self represents the expected type.
    fn unify(self, with: Self, cons: &mut TvarKinds, f: &mut Substitution) -> Result<(), Error> {
        self.0.unify(with.0, cons, f)
    }

    fn constrain(self, with: Kind, cons: &mut TvarKinds) -> Result<(), Error> {
        match with {
            Kind::Equatable => self.0.constrain(with, cons),
            _ => Err(Error::CannotConstrain {
                act: MonoType::arr(self),
                exp: with,
            }),
        }
    }

    fn contains(&self, tv: Tvar) -> bool {
        self.0.contains(tv)
    }
}

/// monotype vector used by vectorization transformation
#[derive(Debug, Display, Clone, PartialEq, Serialize)]
#[display(fmt = "v[{}]", _0)]
pub struct Vector(pub MonoType);

impl Substitutable for Vector {
    fn apply_ref(&self, sub: &dyn Substituter) -> Option<Self> {
        self.0.apply_ref(sub).map(Vector)
    }
    fn free_vars(&self) -> Vec<Tvar> {
        self.0.free_vars()
    }
}

impl MaxTvar for Vector {
    fn max_tvar(&self) -> Option<Tvar> {
        self.0.max_tvar()
    }
}

impl Vector {
    // self represents the expected type.
    fn unify(self, with: Self, cons: &mut TvarKinds, f: &mut Substitution) -> Result<(), Error> {
        self.0.unify(with.0, cons, f)
    }

    fn constrain(self, with: Kind, cons: &mut TvarKinds) -> Result<(), Error> {
        self.0.constrain(with, cons)
    }

    fn contains(&self, tv: Tvar) -> bool {
        self.0.contains(tv)
    }
}

/// A key-value data structure.
#[derive(Debug, Display, Clone, PartialEq, Serialize)]
#[display(fmt = "[{}:{}]", key, val)]
pub struct Dictionary {
    /// Type of key.
    pub key: MonoType,
    /// Type of value.
    pub val: MonoType,
}

impl Substitutable for Dictionary {
    fn apply_ref(&self, sub: &dyn Substituter) -> Option<Self> {
        apply2(&self.key, &self.val, sub).map(|(key, val)| Dictionary { key, val })
    }
    fn free_vars(&self) -> Vec<Tvar> {
        union(self.key.free_vars(), self.val.free_vars())
    }
}

impl MaxTvar for Dictionary {
    fn max_tvar(&self) -> Option<Tvar> {
        [self.key.max_tvar(), self.val.max_tvar()].max_tvar()
    }
}

impl Dictionary {
    fn unify(
        self,
        actual: Self,
        cons: &mut TvarKinds,
        sub: &mut Substitution,
    ) -> Result<(), Error> {
        self.key.unify(actual.key, cons, sub)?;
        apply_then_unify(self.val, actual.val, cons, sub)
    }
    fn constrain(self, with: Kind, _: &mut TvarKinds) -> Result<(), Error> {
        Err(Error::CannotConstrain {
            act: MonoType::dict(self),
            exp: with,
        })
    }
    fn contains(&self, tv: Tvar) -> bool {
        self.key.contains(tv) || self.val.contains(tv)
    }
}

/// An extensible record type.
///
/// A record is either `Empty`, meaning it has no properties,
/// or it is an extension of a record.
///
/// A record may extend what is referred to as a *record
/// variable*. A record variable is a type variable that
/// represents an unknown record type.
#[derive(Debug, Clone, Serialize)]
#[serde(tag = "type")]
pub enum Record {
    /// A record that has no properties.
    Empty,
    /// Extension of a record.
    Extension {
        /// The [`Property`] that extends the record type.
        head: Property,
        /// `tail` is the record variable.
        tail: MonoType,
    },
}

impl fmt::Display for Record {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        f.write_str("{")?;
        let mut s = String::new();
        let tvar = self.format(&mut s)?;
        if let Some(tv) = tvar {
            write!(f, "{} with ", tv)?;
        }
        if s.len() > 2 {
            // remove trailing ', ' delimiter
            s.truncate(s.len() - 2);
        }
        f.write_str(s.as_str())?;
        f.write_str("}")
    }
}

impl cmp::PartialEq for Record {
    fn eq(mut self: &Self, mut other: &Self) -> bool {
        let mut a = RefMonoTypeVecMap::new();
        let t = loop {
            match self {
                Record::Empty => break None,
                Record::Extension {
                    head,
                    tail: MonoType::Record(o),
                } => {
                    a.entry(&head.k).or_insert_with(Vec::new).push(&head.v);
                    self = o;
                }
                Record::Extension {
                    head,
                    tail: MonoType::Var(t),
                } => {
                    a.entry(&head.k).or_insert_with(Vec::new).push(&head.v);
                    break Some(t);
                }
                _ => return false,
            }
        };
        let mut b = RefMonoTypeVecMap::new();
        let v = loop {
            match other {
                Record::Empty => break None,
                Record::Extension {
                    head,
                    tail: MonoType::Record(o),
                } => {
                    b.entry(&head.k).or_insert_with(Vec::new).push(&head.v);
                    other = o;
                }
                Record::Extension {
                    head,
                    tail: MonoType::Var(t),
                } => {
                    b.entry(&head.k).or_insert_with(Vec::new).push(&head.v);
                    break Some(t);
                }
                _ => return false,
            }
        };
        t == v && a == b
    }
}

impl Substitutable for Record {
    fn apply_ref(&self, sub: &dyn Substituter) -> Option<Self> {
        match self {
            Record::Empty => None,
            Record::Extension { head, tail } => {
                apply2(head, tail, sub).map(|(head, tail)| Record::Extension { head, tail })
            }
        }
    }
    fn free_vars(&self) -> Vec<Tvar> {
        match self {
            Record::Empty => Vec::new(),
            Record::Extension { head, tail } => union(tail.free_vars(), head.v.free_vars()),
        }
    }
}

impl MaxTvar for Record {
    fn max_tvar(&self) -> Option<Tvar> {
        match self {
            Record::Empty => None,
            Record::Extension { head, tail } => [head.max_tvar(), tail.max_tvar()].max_tvar(),
        }
    }
}

#[allow(clippy::many_single_char_names)]
impl Record {
    /// Creates a new `Record`
    pub fn new<I>(props: I, tail: Option<MonoType>) -> Self
    where
        I: IntoIterator<Item = Property>,
        I::IntoIter: std::iter::DoubleEndedIterator,
    {
        let mut props = props.into_iter().rev();
        let ret = match tail {
            None => Record::Empty,
            Some(tail) => {
                let head = props
                    .next()
                    .expect("extensible records must have at least one field");

                Record::Extension { head, tail }
            }
        };

        props.fold(ret, |ret, head| Record::Extension {
            head,
            tail: MonoType::record(ret),
        })
    }

    // Below are the rules for record unification. In what follows monotypes
    // are denoted using lowercase letters, and type variables are denoted
    // by a lowercase letter preceded by an apostrophe `'`.
    //
    // `t = u` is read as:
    //
    //     type t unifies with type u
    //
    // `t = u => a = b` is read as:
    //
    //     if t unifies with u, then a must unify with b
    //
    // 1. Two empty records always unify, producing an empty substitution.
    // 2. {a: t | 'r} = {b: u | 'r} => error
    // 3. {a: t | 'r} = {a: u | 'r} => t = u
    // 4. {a: t |  r} = {a: u |  s} => t = u, r = s
    // 5. {a: t |  r} = {b: u |  s} => r = {b: u | 'v}, s = {a: t | 'v}
    //
    // Note rule 2. states that if two records extend the same type variable
    // they must have the same property name otherwise they cannot unify.
    //
    // self represents the expected type.
    //
    fn unify(
        self,
        actual: Self,
        cons: &mut TvarKinds,
        sub: &mut Substitution,
    ) -> Result<(), Error> {
        match (self.clone(), actual.clone()) {
            (Record::Empty, Record::Empty) => Ok(()),
            (
                Record::Extension {
                    head: Property { k: a, v: t },
                    tail: MonoType::Var(l),
                },
                Record::Extension {
                    head: Property { k: b, v: u },
                    tail: MonoType::Var(r),
                },
            ) if a == b && l == r => {
                t.clone()
                    .unify(u.clone(), cons, sub)
                    .map_err(|_| Error::CannotUnifyLabel {
                        lab: a,
                        exp: t,
                        act: u,
                    })
            }
            (
                Record::Extension {
                    head: Property { k: a, .. },
                    tail: MonoType::Var(l),
                },
                Record::Extension {
                    head: Property { k: b, .. },
                    tail: MonoType::Var(r),
                },
            ) if a != b && l == r => Err(Error::CannotUnify {
                exp: MonoType::from(self),
                act: MonoType::from(actual),
            }),
            (
                Record::Extension {
                    head: Property { k: a, v: t },
                    tail: l,
                },
                Record::Extension {
                    head: Property { k: b, v: u },
                    tail: r,
                },
            ) if a == b => {
                t.unify(u, cons, sub)?;
                apply_then_unify(l, r, cons, sub)
            }
            (
                Record::Extension {
                    head: Property { k: a, v: t },
                    tail: l,
                },
                Record::Extension {
                    head: Property { k: b, v: u },
                    tail: r,
                },
            ) if a != b => {
                let var = sub.fresh();
                eprintln!("{} <=> {}", self, actual);
                eprintln!("{}", var);
                let exp = MonoType::from(Record::Extension {
                    head: Property { k: a, v: t },
                    tail: MonoType::Var(var),
                });
                let act = MonoType::from(Record::Extension {
                    head: Property { k: b, v: u },
                    tail: MonoType::Var(var),
                });
                eprintln!("{}: {}", Tvar(0), sub.apply(Tvar(0)));
                eprintln!("1 {} <=> {}", l, act);
                l.unify(act, cons, sub)?;
                eprintln!("2 {} <=> {}", exp, r);
                apply_then_unify(exp, r, cons, sub)
            }
            // If we are expecting {a: u | r} but find {}, label `a` is missing.
            (
                Record::Extension {
                    head: Property { k: a, .. },
                    ..
                },
                Record::Empty,
            ) => Err(Error::MissingLabel(a)),
            // If we are expecting {} but find {a: u | r}, label `a` is extra.
            (
                Record::Empty,
                Record::Extension {
                    head: Property { k: a, .. },
                    ..
                },
            ) => Err(Error::ExtraLabel(a)),
            _ => Err(Error::CannotUnify {
                exp: MonoType::from(self),
                act: MonoType::from(actual),
            }),
        }
    }

    fn constrain(self, with: Kind, cons: &mut TvarKinds) -> Result<(), Error> {
        match with {
            Kind::Record => Ok(()),
            Kind::Equatable => match self {
                Record::Empty => Ok(()),
                Record::Extension { head, tail } => {
                    head.v.constrain(with, cons)?;
                    Ok(tail.constrain(with, cons)?)
                }
            },
            _ => Err(Error::CannotConstrain {
                act: MonoType::from(self),
                exp: with,
            }),
        }
    }

    fn contains(&self, tv: Tvar) -> bool {
        match self {
            Record::Empty => false,
            Record::Extension { head, tail } => head.v.contains(tv) || tail.contains(tv),
        }
    }

    fn format(&self, f: &mut String) -> Result<Option<Tvar>, fmt::Error> {
        match self {
            Record::Empty => Ok(None),
            Record::Extension { head, tail } => match tail {
                MonoType::Var(tv) => {
                    write!(f, "{}, ", head)?;
                    Ok(Some(*tv))
                }
                MonoType::Record(obj) => {
                    write!(f, "{}, ", head)?;
                    obj.format(f)
                }
                _ => Err(fmt::Error),
            },
        }
    }
}

// Unification requires that the current substitution be applied
// to both sides of a constraint before unifying.
//
// This helper function applies a substitution to a constraint
// before unifying the two types. Note the substitution produced
// from unification is merged with input substitution before it
// is returned.
//
// TODO Remove
fn apply_then_unify(
    exp: MonoType,
    act: MonoType,
    cons: &mut TvarKinds,
    sub: &mut Substitution,
) -> Result<(), Error> {
    exp.unify(act, cons, sub)?;
    Ok(())
}

/// A key-value pair representing a property type in a record.
#[derive(Debug, Display, Clone, PartialEq, Serialize)]
#[display(fmt = "{}:{}", k, v)]
#[allow(missing_docs)]
pub struct Property {
    pub k: String,
    pub v: MonoType,
}

impl Substitutable for Property {
    fn apply_ref(&self, sub: &dyn Substituter) -> Option<Self> {
        self.v.apply_ref(sub).map(|v| Property {
            k: self.k.clone(),
            v,
        })
    }
    fn free_vars(&self) -> Vec<Tvar> {
        self.v.free_vars()
    }
}

impl MaxTvar for Property {
    fn max_tvar(&self) -> Option<Tvar> {
        self.v.max_tvar()
    }
}

/// Represents a function type.
///
/// A function type is defined by a set of required arguments,
/// a set of optional arguments, an optional pipe argument, and
/// a required return type.
#[derive(Debug, Clone, PartialEq, Serialize)]
pub struct Function {
    /// Required arguments to a function.
    pub req: MonoTypeMap,
    /// Optional arguments to a function.
    pub opt: MonoTypeMap,
    /// An optional pipe argument.
    pub pipe: Option<Property>,
    /// Required return type.
    pub retn: MonoType,
}

impl fmt::Display for Function {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        let required = self
            .req
            .iter()
            // Sort args with BTree
            .collect::<BTreeMap<_, _>>()
            .iter()
            .map(|(&k, &v)| Property {
                k: k.clone(),
                v: v.clone(),
            })
            .collect::<Vec<_>>();

        let optional = self
            .opt
            .iter()
            // Sort args with BTree
            .collect::<BTreeMap<_, _>>()
            .iter()
            .map(|(&k, &v)| Property {
                k: String::from("?") + k,
                v: v.clone(),
            })
            .collect::<Vec<_>>();

        let pipe = match &self.pipe {
            Some(pipe) => {
                if pipe.k == "<-" {
                    vec![pipe.clone()]
                } else {
                    vec![Property {
                        k: String::from("<-") + &pipe.k,
                        v: pipe.v.clone(),
                    }]
                }
            }
            None => vec![],
        };

        write!(
            f,
            "({}) => {}",
            pipe.iter()
                .chain(required.iter().chain(optional.iter()))
                .map(|x| x.to_string())
                .collect::<Vec<_>>()
                .join(", "),
            self.retn
        )
    }
}

#[allow(clippy::implicit_hasher)]
impl<K: Ord + Clone, T: Substitutable + Clone> Substitutable for SemanticMap<K, T> {
    fn apply_ref(&self, sub: &dyn Substituter) -> Option<Self> {
        merge_collect(
            &mut (),
            self,
            |_, (k, v)| v.apply_ref(sub).map(|v| (k.clone(), v)),
            |_, (k, v)| (k.clone(), v.clone()),
        )
    }
    fn free_vars(&self) -> Vec<Tvar> {
        self.values()
            .fold(Vec::new(), |vars, t| union(vars, t.free_vars()))
    }
}

impl<T: Substitutable> Substitutable for Option<T> {
    fn apply_ref(&self, sub: &dyn Substituter) -> Option<Self> {
        match self {
            None => None,
            Some(t) => t.apply_ref(sub).map(Some),
        }
    }
    fn free_vars(&self) -> Vec<Tvar> {
        match self {
            Some(t) => t.free_vars(),
            None => Vec::new(),
        }
    }
}

impl Substitutable for Function {
    fn apply_ref(&self, sub: &dyn Substituter) -> Option<Self> {
        let Function {
            req,
            opt,
            pipe,
            retn,
        } = self;
        apply4(req, opt, pipe, retn, sub).map(|(req, opt, pipe, retn)| Function {
            req,
            opt,
            pipe,
            retn,
        })
    }
    fn free_vars(&self) -> Vec<Tvar> {
        union(
            self.req.free_vars(),
            union(
                self.opt.free_vars(),
                union(self.pipe.free_vars(), self.retn.free_vars()),
            ),
        )
    }
}

impl<U, T: MaxTvar> MaxTvar for SemanticMap<U, T> {
    fn max_tvar(&self) -> Option<Tvar> {
        self.iter()
            .map(|(_, t)| t.max_tvar())
            .fold(None, |max, tv| if tv > max { tv } else { max })
    }
}

impl<T: MaxTvar> MaxTvar for Option<T> {
    fn max_tvar(&self) -> Option<Tvar> {
        match self {
            None => None,
            Some(t) => t.max_tvar(),
        }
    }
}

impl MaxTvar for Function {
    fn max_tvar(&self) -> Option<Tvar> {
        [
            self.req.max_tvar(),
            self.opt.max_tvar(),
            self.pipe.max_tvar(),
            self.retn.max_tvar(),
        ]
        .max_tvar()
    }
}

impl Function {
    /// Given two function types f and g, the process for unifying their arguments is as follows:
    /// 1. If a required arg of f is not present in the arguments of g,
    ///    otherwise unify both argument types.
    /// 2. If an optional arg of f is not present in the arguments of g, continue,
    ///    otherwise unify both argument types (repeat for g).
    /// 3. Lastly unify pipe args. Note that pipe arguments are optional.
    ///    However if a pipe arg was used in a calling context, i.e it's an un-named pipe arg,
    ///    then the other type must specify a pipe arg too, otherwise unification fails.
    ///
    /// For pipe arguments, it becomes quite tricky. Take these statements:
    ///
    /// 1. f = (a=<-, b) => {...}
    /// 2. 0 |> f(b: 1)
    /// 3. f(a: 0, b: 1)
    /// 4. f = (d=<-, b, c=0) => {...}
    ///
    /// 2 and 3 are two equivalent ways of invoking 1, and they should both unify.
    /// `a` is the named pipe argument in 1. In 2, the pipe argument is unnamed.
    ///
    /// Unify 1 and 2: one of the required arguments of 1 will not be in its call,
    /// so, we should check for the pipe argument and succeed. If we do the other way around (unify
    /// 2 with 1), the unnamed pipe argument unifies with the other pipe argument.
    ///
    /// Unify 1 and 3: no problem, required arguments are satisfied. Take care that, if you unify
    /// 3 with 1, you will find `a` in 1's pipe argument.
    ///
    /// Unify 1 and 4: should fail because `d` != `a`.
    ///
    /// Unify 2 and 3: should fail because `a` is not in the arguments of 2.
    ///
    /// Unify 2 and 4: should succeed, the same as 1 and 2.
    ///
    /// Unify 3 and 4: should fail because `a` is not in the arguments of 4.
    ///
    /// self represents the expected type.
    fn unify(
        self,
        actual: Self,
        cons: &mut TvarKinds,
        sub: &mut Substitution,
    ) -> Result<(), Error> {
        // Some aliasing for coherence with the doc.
        let mut f = self;
        let mut g = actual;
        // Fix pipe arguments:
        // Make them required arguments with the correct name.
        match (f.pipe, g.pipe) {
            // Both functions have pipe arguments.
            (Some(fp), Some(gp)) => {
                if fp.k != "<-" && gp.k != "<-" && fp.k != gp.k {
                    // Both are named and the name differs, fail unification.
                    return Err(Error::MultiplePipeArguments {
                        exp: fp.k,
                        act: gp.k,
                    });
                } else {
                    // At least one is unnamed or they are both named with the same name.
                    // This means they should match. Enforce this condition by inserting
                    // the pipe argument into the required ones with the same key.
                    f.req.insert(fp.k.clone(), fp.v);
                    g.req.insert(fp.k, gp.v);
                }
            }
            // F has a pipe argument and g does not.
            (Some(fp), None) => {
                if fp.k == "<-" {
                    // The pipe argument is unnamed and g does not have one.
                    // Fail unification.
                    return Err(Error::MissingPipeArgument);
                } else {
                    // This is a named argument, simply put it into the required ones.
                    f.req.insert(fp.k, fp.v);
                }
            }
            // G has a pipe argument and f does not.
            (None, Some(gp)) => {
                if gp.k == "<-" {
                    // The pipe argument is unnamed and f does not have one.
                    // Fail unification.
                    return Err(Error::MissingPipeArgument);
                } else {
                    // This is a named argument, simply put it into the required ones.
                    g.req.insert(gp.k, gp.v);
                }
            }
            // Nothing to do.
            (None, None) => (),
        }
        // Now that f has not been consumed yet, check that every required argument in g is in f too.
        for (name, _) in g.req.iter() {
            if !f.req.contains_key(name) && !f.opt.contains_key(name) {
                return Err(Error::ExtraArgument(String::from(name)));
            }
        }
        // Unify f's required arguments.

        let g_opt = &mut g.opt;
        for (name, exp) in f.req.into_iter() {
            if let Some(act) = g.req.remove(&name).or_else(|| g_opt.remove(&name)) {
                // The required argument is in g's required arguments.
                apply_then_unify(exp, act, cons, sub)
                    .map_err(|e| Error::CannotUnifyArgument(name, Box::new(e)))?;
            } else {
                return Err(Error::MissingArgument(name));
            }
        }
        // Unify f's optional arguments.
        for (name, exp) in f.opt.into_iter() {
            if let Some(act) = g.req.remove(&name).or_else(|| g_opt.remove(&name)) {
                apply_then_unify(exp, act, cons, sub)
                    .map_err(|e| Error::CannotUnifyArgument(name, Box::new(e)))?;
            }
        }
        // Unify return types.
        match apply_then_unify(f.retn.clone(), g.retn.clone(), cons, sub) {
            Err(_) => Err(Error::CannotUnifyReturn {
                exp: f.retn,
                act: g.retn,
            }),
            Ok(sub) => Ok(sub),
        }
    }

    fn constrain(self, with: Kind, _: &mut TvarKinds) -> Result<(), Error> {
        Err(Error::CannotConstrain {
            act: MonoType::fun(self),
            exp: with,
        })
    }

    fn contains(&self, tv: Tvar) -> bool {
        if let Some(pipe) = &self.pipe {
            self.req.values().any(|t| t.contains(tv))
                || self.opt.values().any(|t| t.contains(tv))
                || pipe.v.contains(tv)
                || self.retn.contains(tv)
        } else {
            self.req.values().any(|t| t.contains(tv))
                || self.opt.values().any(|t| t.contains(tv))
                || self.retn.contains(tv)
        }
    }
}

/// Trait for returning the maximum type variable of a type.
pub trait MaxTvar {
    /// Return the maximum type variable of a type.
    fn max_tvar(&self) -> Option<Tvar>;
}

#[cfg(test)]
mod tests {
    use super::*;

    use std::collections::BTreeMap;

    use crate::ast::get_err_type_expression;
    use crate::parser;
    use crate::semantic::convert::{convert_monotype, convert_polytype};

    /// `polytype` is a utility method that returns a `PolyType` from a string.
    pub fn polytype(typ: &str) -> PolyType {
        let mut p = parser::Parser::new(typ);

        let typ_expr = p.parse_type_expression();
        let err = get_err_type_expression(typ_expr.clone());

        if err != "" {
            panic!("TypeExpression parsing failed for {}. {:?}", typ, err);
        }
        convert_polytype(typ_expr, &mut Substitution::default()).unwrap()
    }

    fn parse_type(
        expr: &str,
        tvars: &mut BTreeMap<String, Tvar>,
        sub: &mut Substitution,
    ) -> MonoType {
        let mut p = parser::Parser::new(expr);

        let typ_expr = p.parse_type_expression();
        let err = get_err_type_expression(typ_expr.clone());

        if err != "" {
            panic!("TypeExpression parsing failed. {:?}", err);
        }
        convert_monotype(typ_expr.monotype, tvars, sub).unwrap()
    }

    #[test]
    fn display_kind_addable() {
        assert!(Kind::Addable.to_string() == "Addable");
    }
    #[test]
    fn display_kind_subtractable() {
        assert!(Kind::Subtractable.to_string() == "Subtractable");
    }
    #[test]
    fn display_kind_divisible() {
        assert!(Kind::Divisible.to_string() == "Divisible");
    }
    #[test]
    fn display_kind_numeric() {
        assert!(Kind::Numeric.to_string() == "Numeric");
    }
    #[test]
    fn display_kind_comparable() {
        assert!(Kind::Comparable.to_string() == "Comparable");
    }
    #[test]
    fn display_kind_equatable() {
        assert!(Kind::Equatable.to_string() == "Equatable");
    }
    #[test]
    fn display_kind_nullable() {
        assert!(Kind::Nullable.to_string() == "Nullable");
    }
    #[test]
    fn display_kind_row() {
        assert!(Kind::Record.to_string() == "Record");
    }
    #[test]
    fn display_kind_stringable() {
        assert!(Kind::Stringable.to_string() == "Stringable");
    }

    #[test]
    fn display_type_bool() {
        assert_eq!("bool", MonoType::Bool.to_string());
    }
    #[test]
    fn display_type_int() {
        assert_eq!("int", MonoType::Int.to_string());
    }
    #[test]
    fn display_type_uint() {
        assert_eq!("uint", MonoType::Uint.to_string());
    }
    #[test]
    fn display_type_float() {
        assert_eq!("float", MonoType::Float.to_string());
    }
    #[test]
    fn display_type_string() {
        assert_eq!("string", MonoType::String.to_string());
    }
    #[test]
    fn display_type_duration() {
        assert_eq!("duration", MonoType::Duration.to_string());
    }
    #[test]
    fn display_type_time() {
        assert_eq!("time", MonoType::Time.to_string());
    }
    #[test]
    fn display_type_regexp() {
        assert_eq!("regexp", MonoType::Regexp.to_string());
    }
    #[test]
    fn display_type_bytes() {
        assert_eq!("bytes", MonoType::Bytes.to_string());
    }
    #[test]
    fn display_type_tvar() {
        assert_eq!("t10", MonoType::Var(Tvar(10)).to_string());
    }
    #[test]
    fn display_type_array() {
        assert_eq!("[int]", MonoType::from(Array(MonoType::Int)).to_string());
    }
    #[test]
    fn display_type_vector() {
        assert_eq!(
            "v[int]",
            MonoType::Vector(Box::new(Vector(MonoType::Int))).to_string()
        );
    }
    #[test]
    fn display_type_record() {
        assert_eq!(
            "{A with a:int, b:string}",
            Record::new(
                [
                    Property {
                        k: String::from("a"),
                        v: MonoType::Int,
                    },
                    Property {
                        k: String::from("b"),
                        v: MonoType::String,
                    }
                ],
                Some(MonoType::Var(Tvar(0))),
            )
            .to_string()
        );
        assert_eq!(
            "{a:int, b:string}",
            Record::new(
                [
                    Property {
                        k: String::from("a"),
                        v: MonoType::Int,
                    },
                    Property {
                        k: String::from("b"),
                        v: MonoType::String,
                    }
                ],
                Some(MonoType::from(Record::Empty)),
            )
            .to_string()
        );
    }
    #[test]
    fn display_type_function() {
        assert_eq!(
            "() => int",
            Function {
                req: MonoTypeMap::new(),
                opt: MonoTypeMap::new(),
                pipe: None,
                retn: MonoType::Int,
            }
            .to_string()
        );
        assert_eq!(
            "(<-:int) => int",
            Function {
                req: MonoTypeMap::new(),
                opt: MonoTypeMap::new(),
                pipe: Some(Property {
                    k: String::from("<-"),
                    v: MonoType::Int,
                }),
                retn: MonoType::Int,
            }
            .to_string()
        );
        assert_eq!(
            "(<-a:int) => int",
            Function {
                req: MonoTypeMap::new(),
                opt: MonoTypeMap::new(),
                pipe: Some(Property {
                    k: String::from("a"),
                    v: MonoType::Int,
                }),
                retn: MonoType::Int,
            }
            .to_string()
        );
        assert_eq!(
            "(<-:int, a:int, b:int) => int",
            Function {
                req: semantic_map! {
                    String::from("a") => MonoType::Int,
                    String::from("b") => MonoType::Int,
                },
                opt: MonoTypeMap::new(),
                pipe: Some(Property {
                    k: String::from("<-"),
                    v: MonoType::Int,
                }),
                retn: MonoType::Int,
            }
            .to_string()
        );
        assert_eq!(
            "(<-:int, ?a:int, ?b:int) => int",
            Function {
                req: MonoTypeMap::new(),
                opt: semantic_map! {
                    String::from("a") => MonoType::Int,
                    String::from("b") => MonoType::Int,
                },
                pipe: Some(Property {
                    k: String::from("<-"),
                    v: MonoType::Int,
                }),
                retn: MonoType::Int,
            }
            .to_string()
        );
        assert_eq!(
            "(<-:int, a:int, b:int, ?c:int, ?d:int) => int",
            Function {
                req: semantic_map! {
                    String::from("a") => MonoType::Int,
                    String::from("b") => MonoType::Int,
                },
                opt: semantic_map! {
                    String::from("c") => MonoType::Int,
                    String::from("d") => MonoType::Int,
                },
                pipe: Some(Property {
                    k: String::from("<-"),
                    v: MonoType::Int,
                }),
                retn: MonoType::Int,
            }
            .to_string()
        );
        assert_eq!(
            "(a:int, ?b:bool) => int",
            Function {
                req: semantic_map! {
                    String::from("a") => MonoType::Int,
                },
                opt: semantic_map! {
                    String::from("b") => MonoType::Bool,
                },
                pipe: None,
                retn: MonoType::Int,
            }
            .to_string()
        );
        assert_eq!(
            "(<-a:int, b:int, c:int, ?d:bool) => int",
            Function {
                req: semantic_map! {
                    String::from("b") => MonoType::Int,
                    String::from("c") => MonoType::Int,
                },
                opt: semantic_map! {
                    String::from("d") => MonoType::Bool,
                },
                pipe: Some(Property {
                    k: String::from("a"),
                    v: MonoType::Int,
                }),
                retn: MonoType::Int,
            }
            .to_string()
        );
    }

    #[test]
    fn display_polytype() {
        assert_eq!(
            "int",
            PolyType {
                vars: Vec::new(),
                cons: TvarKinds::new(),
                expr: MonoType::Int,
            }
            .to_string(),
        );
        assert_eq!(
            "(x:A) => A",
            PolyType {
                vars: vec![Tvar(0)],
                cons: TvarKinds::new(),
                expr: MonoType::from(Function {
                    req: semantic_map! {
                        String::from("x") => MonoType::Var(Tvar(0)),
                    },
                    opt: MonoTypeMap::new(),
                    pipe: None,
                    retn: MonoType::Var(Tvar(0)),
                }),
            }
            .to_string(),
        );
        assert_eq!(
            "(x:A, y:B) => {x:A, y:B}",
            PolyType {
                vars: vec![Tvar(0), Tvar(1)],
                cons: TvarKinds::new(),
                expr: MonoType::from(Function {
                    req: semantic_map! {
                        String::from("x") => MonoType::Var(Tvar(0)),
                        String::from("y") => MonoType::Var(Tvar(1)),
                    },
                    opt: MonoTypeMap::new(),
                    pipe: None,
                    retn: MonoType::from(Record::new(
                        [
                            Property {
                                k: String::from("x"),
                                v: MonoType::Var(Tvar(0)),
                            },
                            Property {
                                k: String::from("y"),
                                v: MonoType::Var(Tvar(1)),
                            }
                        ],
                        Some(MonoType::from(Record::Empty)),
                    )),
                }),
            }
            .to_string(),
        );
        assert_eq!(
            "(a:A, b:A) => A where A: Addable",
            PolyType {
                vars: vec![Tvar(0)],
                cons: semantic_map! {Tvar(0) => vec![Kind::Addable]},
                expr: MonoType::from(Function {
                    req: semantic_map! {
                        String::from("a") => MonoType::Var(Tvar(0)),
                        String::from("b") => MonoType::Var(Tvar(0)),
                    },
                    opt: MonoTypeMap::new(),
                    pipe: None,
                    retn: MonoType::Var(Tvar(0)),
                }),
            }
            .to_string(),
        );
        assert_eq!(
            "(x:A, y:B) => {x:A, y:B} where A: Addable, B: Divisible",
            PolyType {
                vars: vec![Tvar(0), Tvar(1)],
                cons: semantic_map! {
                    Tvar(0) => vec![Kind::Addable],
                    Tvar(1) => vec![Kind::Divisible],
                },
                expr: MonoType::from(Function {
                    req: semantic_map! {
                        String::from("x") => MonoType::Var(Tvar(0)),
                        String::from("y") => MonoType::Var(Tvar(1)),
                    },
                    opt: MonoTypeMap::new(),
                    pipe: None,
                    retn: MonoType::from(Record::new(
                        [
                            Property {
                                k: String::from("x"),
                                v: MonoType::Var(Tvar(0)),
                            },
                            Property {
                                k: String::from("y"),
                                v: MonoType::Var(Tvar(1)),
                            }
                        ],
                        Some(MonoType::from(Record::Empty)),
                    )),
                }),
            }
            .to_string(),
        );
        assert_eq!(
            "(x:A, y:B) => {x:A, y:B} where A: Comparable + Equatable, B: Addable + Divisible",
            PolyType {
                vars: vec![Tvar(0), Tvar(1)],
                cons: semantic_map! {
                    Tvar(0) => vec![Kind::Comparable, Kind::Equatable],
                    Tvar(1) => vec![Kind::Addable, Kind::Divisible],
                },
                expr: MonoType::from(Function {
                    req: semantic_map! {
                        String::from("x") => MonoType::Var(Tvar(0)),
                        String::from("y") => MonoType::Var(Tvar(1)),
                    },
                    opt: MonoTypeMap::new(),
                    pipe: None,
                    retn: MonoType::from(Record::new(
                        [
                            Property {
                                k: String::from("x"),
                                v: MonoType::Var(Tvar(0)),
                            },
                            Property {
                                k: String::from("y"),
                                v: MonoType::Var(Tvar(1)),
                            }
                        ],
                        Some(MonoType::from(Record::Empty))
                    )),
                }),
            }
            .to_string(),
        );
    }

    #[test]
    fn compare_records() {
        assert_eq!(
            // {A with a:int, b:string}
            MonoType::from(Record::new(
                [
                    Property {
                        k: String::from("a"),
                        v: MonoType::Int,
                    },
                    Property {
                        k: String::from("b"),
                        v: MonoType::String,
                    }
                ],
                Some(MonoType::Var(Tvar(0))),
            )),
            // {A with b:string, a:int}
            MonoType::from(Record::new(
                [
                    Property {
                        k: String::from("b"),
                        v: MonoType::String,
                    },
                    Property {
                        k: String::from("a"),
                        v: MonoType::Int,
                    }
                ],
                Some(MonoType::Var(Tvar(0))),
            )),
        );
        assert_eq!(
            // {A with a:int, b:string, b:int, c:float}
            MonoType::from(Record::new(
                [
                    Property {
                        k: String::from("a"),
                        v: MonoType::Int,
                    },
                    Property {
                        k: String::from("b"),
                        v: MonoType::String,
                    },
                    Property {
                        k: String::from("b"),
                        v: MonoType::Int,
                    },
                    Property {
                        k: String::from("c"),
                        v: MonoType::Float,
                    }
                ],
                Some(MonoType::Var(Tvar(0))),
            )),
            // {A with c:float, b:string, b:int, a:int}
            MonoType::from(Record::new(
                [
                    Property {
                        k: String::from("c"),
                        v: MonoType::Float,
                    },
                    Property {
                        k: String::from("b"),
                        v: MonoType::String,
                    },
                    Property {
                        k: String::from("b"),
                        v: MonoType::Int,
                    },
                    Property {
                        k: String::from("a"),
                        v: MonoType::Int,
                    }
                ],
                Some(MonoType::Var(Tvar(0))),
            ))
        );
        assert_ne!(
            // {A with a:int, b:string, b:int, c:float}
            MonoType::from(Record::new(
                [
                    Property {
                        k: String::from("a"),
                        v: MonoType::Int,
                    },
                    Property {
                        k: String::from("b"),
                        v: MonoType::String,
                    },
                    Property {
                        k: String::from("b"),
                        v: MonoType::Int,
                    },
                    Property {
                        k: String::from("c"),
                        v: MonoType::Float,
                    }
                ],
                Some(MonoType::Var(Tvar(0))),
            )),
            // {A with a:int, b:int, b:string, c:float}
            MonoType::from(Record::new(
                [
                    Property {
                        k: String::from("a"),
                        v: MonoType::Int,
                    },
                    Property {
                        k: String::from("b"),
                        v: MonoType::Int,
                    },
                    Property {
                        k: String::from("b"),
                        v: MonoType::String,
                    },
                    Property {
                        k: String::from("c"),
                        v: MonoType::Float,
                    }
                ],
                Some(MonoType::Var(Tvar(0))),
            ))
        );
        assert_ne!(
            // {a:int, b:string}
            MonoType::from(Record::new(
                [
                    Property {
                        k: String::from("a"),
                        v: MonoType::Int,
                    },
                    Property {
                        k: String::from("b"),
                        v: MonoType::String,
                    }
                ],
                Some(MonoType::from(Record::Empty)),
            )),
            // {b:int, a:int}
            MonoType::from(Record::new(
                [
                    Property {
                        k: String::from("b"),
                        v: MonoType::Int,
                    },
                    Property {
                        k: String::from("a"),
                        v: MonoType::Int,
                    }
                ],
                Some(MonoType::from(Record::Empty)),
            ))
        );
        assert_ne!(
            // {a:int}
            MonoType::from(Record::Extension {
                head: Property {
                    k: String::from("a"),
                    v: MonoType::Int,
                },
                tail: MonoType::from(Record::Empty),
            }),
            // {A with a:int}
            MonoType::from(Record::Extension {
                head: Property {
                    k: String::from("a"),
                    v: MonoType::Int,
                },
                tail: MonoType::Var(Tvar(0)),
            }),
        );
        assert_ne!(
            // {A with a:int}
            MonoType::from(Record::Extension {
                head: Property {
                    k: String::from("a"),
                    v: MonoType::Int,
                },
                tail: MonoType::Var(Tvar(0)),
            }),
            // {B with a:int}
            MonoType::from(Record::Extension {
                head: Property {
                    k: String::from("a"),
                    v: MonoType::Int,
                },
                tail: MonoType::Var(Tvar(1)),
            }),
        );
    }

    #[test]
    fn unify_ints() {
        MonoType::Int
            .unify(
                MonoType::Int,
                &mut TvarKinds::new(),
                &mut Substitution::default(),
            )
            .unwrap();
    }
    #[test]
    fn constrain_ints() {
        let allowable_cons = vec![
            Kind::Addable,
            Kind::Subtractable,
            Kind::Divisible,
            Kind::Numeric,
            Kind::Comparable,
            Kind::Equatable,
            Kind::Nullable,
            Kind::Stringable,
        ];
        for c in allowable_cons {
            MonoType::Int.constrain(c, &mut TvarKinds::new()).unwrap();
        }

        let sub = MonoType::Int
            .constrain(Kind::Record, &mut TvarKinds::new())
            .map(|_| ());
        assert_eq!(
            Err(Error::CannotConstrain {
                act: MonoType::Int,
                exp: Kind::Record
            }),
            sub
        );
    }
    #[test]
    fn constrain_rows() {
        Record::Empty
            .constrain(Kind::Record, &mut TvarKinds::new())
            .unwrap();

        let unallowable_cons = vec![
            Kind::Addable,
            Kind::Subtractable,
            Kind::Divisible,
            Kind::Numeric,
            Kind::Comparable,
            Kind::Nullable,
        ];
        for c in unallowable_cons {
            let sub = Record::Empty
                .constrain(c, &mut TvarKinds::new())
                .map(|_| ());
            assert_eq!(
                Err(Error::CannotConstrain {
                    act: MonoType::from(Record::Empty),
                    exp: c
                }),
                sub
            );
        }
    }
    #[test]
    fn constrain_vectors() {
        // kind constraints allowed for Vector(MonoType::Int)
        let allowable_cons_int = vec![
            Kind::Addable,
            Kind::Subtractable,
            Kind::Divisible,
            Kind::Numeric,
            Kind::Comparable,
            Kind::Equatable,
            Kind::Nullable,
            Kind::Stringable,
        ];

        for c in allowable_cons_int {
            let vector_int = MonoType::Vector(Box::new(Vector(MonoType::Int)));
            vector_int.constrain(c, &mut TvarKinds::new()).unwrap();
        }

        // kind constraints not allowed for Vector(MonoType::String)
        let unallowable_cons_string = vec![Kind::Subtractable, Kind::Divisible, Kind::Numeric];
        for c in unallowable_cons_string {
            let vector_string = MonoType::Vector(Box::new(Vector(MonoType::String)));
            let sub = vector_string
                .constrain(c, &mut TvarKinds::new())
                .map(|_| ());
            assert_eq!(
                Err(Error::CannotConstrain {
                    act: MonoType::String,
                    exp: c
                }),
                sub
            );
        }

        // kind constraints not allowed for Vector(MonoType::Time)
        let unallowable_cons_time = vec![Kind::Subtractable, Kind::Divisible, Kind::Numeric];
        for c in unallowable_cons_time {
            let vector_time = MonoType::Vector(Box::new(Vector(MonoType::Time)));
            let sub = vector_time.constrain(c, &mut TvarKinds::new()).map(|_| ());
            assert_eq!(
                Err(Error::CannotConstrain {
                    act: MonoType::Time,
                    exp: c
                }),
                sub
            );
        }

        // kind constraints allowed for Vector(MonoType::Time)
        let allowable_cons_time = vec![
            Kind::Comparable,
            Kind::Equatable,
            Kind::Nullable,
            Kind::Stringable,
            Kind::Timeable,
        ];

        for c in allowable_cons_time {
            let vector_time = MonoType::Vector(Box::new(Vector(MonoType::Time)));
            vector_time.constrain(c, &mut TvarKinds::new()).unwrap();
        }
    }
    #[test]
    fn unify_error() {
        let err = MonoType::Int
            .unify(
                MonoType::String,
                &mut TvarKinds::new(),
                &mut Substitution::default(),
            )
            .unwrap_err();
        assert_eq!(
            err.to_string(),
            String::from("expected int but found string"),
        );
    }
    #[test]
    fn unify_tvars() {
        let mut sub = Substitution::default();
        sub.mk_fresh(2);
        MonoType::Var(Tvar(0))
            .unify(MonoType::Var(Tvar(1)), &mut TvarKinds::new(), &mut sub)
            .unwrap();
        assert_eq!(sub.apply(Tvar(0)), sub.apply(Tvar(1)));
    }
    #[test]
    fn unify_constrained_tvars() {
        let mut cons = semantic_map! {Tvar(0) => vec![Kind::Addable, Kind::Divisible]};
        let mut sub = Substitution::default();
        sub.mk_fresh(2);
        MonoType::Var(Tvar(0))
            .unify(MonoType::Var(Tvar(1)), &mut cons, &mut sub)
            .unwrap();
        assert_eq!(sub.apply(Tvar(0)), MonoType::Var(Tvar(1)));
        assert_eq!(
            cons,
            semantic_map! {Tvar(1) => vec![Kind::Addable, Kind::Divisible]},
        );
    }
    #[test]
    fn cannot_unify_functions() {
        // g-required and g-optional arguments do not contain a f-required argument (and viceversa).
        let f = polytype("(a: A, b: A, ?c: B) => A where A: Addable, B: Divisible ");
        let g = polytype("(d: C, ?e: C) => C where C: Addable ");
        if let (
            PolyType {
                vars: _,
                cons: f_cons,
                expr: MonoType::Fun(f),
            },
            PolyType {
                vars: _,
                cons: g_cons,
                expr: MonoType::Fun(g),
            },
        ) = (f, g)
        {
            // this extends the first map with the second by generating a new one.
            let mut cons = f_cons.into_iter().chain(g_cons).collect();
            let mut sub = Substitution::default();
            sub.mk_fresh(2);
            let res = f.clone().unify(*g.clone(), &mut cons, &mut sub);
            assert!(res.is_err());
            let res = g.clone().unify(*f.clone(), &mut cons, &mut sub);
            assert!(res.is_err());
        } else {
            panic!("the monotypes under examination are not functions");
        }
        // f has a pipe argument, but g does not (and viceversa).
        let f = polytype("(<-pip:A, a: B) => A where A: Addable, B: Divisible ");
        let g = polytype("(a: C) => C where C: Addable ");
        if let (
            PolyType {
                vars: _,
                cons: f_cons,
                expr: MonoType::Fun(f),
            },
            PolyType {
                vars: _,
                cons: g_cons,
                expr: MonoType::Fun(g),
            },
        ) = (f, g)
        {
            let mut cons = f_cons.into_iter().chain(g_cons).collect();
            let mut sub = Substitution::default();
            sub.mk_fresh(2);
            let res = f.clone().unify(*g.clone(), &mut cons, &mut sub);
            assert!(res.is_err());
            let res = g.clone().unify(*f.clone(), &mut cons, &mut sub);
            assert!(res.is_err());
        } else {
            panic!("the monotypes under examination are not functions");
        }
    }
    #[test]
    fn unify_function_with_function_call() {
        let fn_type = polytype("(a: A, b: A, ?c: B) => A where A: Addable, B: Divisible ");
        // (a: int, b: int) => int
        let call_type = Function {
            // all arguments are required in a function call.
            req: semantic_map! {
                "a".to_string() => MonoType::Int,
                "b".to_string() => MonoType::Int,
            },
            opt: semantic_map! {},
            pipe: None,
            retn: MonoType::Int,
        };
        if let PolyType {
            vars: _,
            mut cons,
            expr: MonoType::Fun(f),
        } = fn_type
        {
            let mut sub = Substitution::default();
            sub.mk_fresh(2);
            f.unify(call_type, &mut cons, &mut sub).unwrap();
            assert_eq!(sub.apply(Tvar(0)), MonoType::Int);
            // the constraint on A gets removed.
            assert_eq!(cons, semantic_map! {Tvar(1) => vec![Kind::Divisible]});
        } else {
            panic!("the monotype under examination is not a function");
        }
    }
    #[test]
    fn unify_higher_order_functions() {
        let f = polytype(
            "(a: A, b: A, ?c: (a: A) => B) => (d:  string) => A where A: Addable, B: Divisible ",
        );
        let g = polytype("(a: int, b: int, c: (a: int) => float) => (d: string) => int");
        if let (
            PolyType {
                vars: _,
                cons: f_cons,
                expr: MonoType::Fun(f),
            },
            PolyType {
                vars: _,
                cons: g_cons,
                expr: MonoType::Fun(g),
            },
        ) = (f, g)
        {
            // this extends the first map with the second by generating a new one.
            let mut cons = f_cons.into_iter().chain(g_cons).collect();
            let mut sub = Substitution::default();
            sub.mk_fresh(2);
            f.unify(*g, &mut cons, &mut sub).unwrap();
            assert_eq!(sub.apply(Tvar(0)), MonoType::Int);
            assert_eq!(sub.apply(Tvar(1)), MonoType::Float);
            // we know everything about tvars, there is no constraint.
            assert_eq!(cons, semantic_map! {});
        } else {
            panic!("the monotypes under examination are not functions");
        }
    }

    #[allow(unused)]
    macro_rules! assert_unify {
        ($expected: expr, $actual: expr $(,)?) => {{
            let mut sub = Substitution::default();
            let mut tvars = BTreeMap::new();
            parse_type($expected, &mut tvars, &mut sub)
                .unify(
                    parse_type($actual, &mut tvars, &mut sub),
                    &mut Default::default(),
                    &mut sub,
                )
                .unwrap_or_else(|err| panic!("{}", err));
        }};
    }

    macro_rules! assert_unify_err {
        ($expected: expr, $actual: expr $(, $pat: pat)? $(,)?) => {{
            let mut sub = Substitution::default();
            let mut tvars = BTreeMap::new();
            let result = parse_type($expected, &mut tvars, &mut sub).unify(
                parse_type($actual, &mut tvars, &mut sub),
                &mut Default::default(),
                &mut sub,
            );
            match result {
                $(
                    Err($pat) => (),
                    Err(err) => panic!("Unexpected error: {}", err),
                )?
                #[allow(unreachable_patterns)]
                Err(_) => (),
                Ok(_) => panic!("Unexpected "),
            }
        }};
    }

    #[test]
    fn unify_records() {
        assert_unify_err!(
            "(fn:(r:A) => A) => [A]",
            "(fn:(r:{C with _value_data:float}) => {C with _value:float, _value_data:float}) => D",
        );
    }

    #[test]
    fn kind_order_is_lexical() {
        macro_rules! complete_list {
            ($($path: path),* $(,)?) => { {
                if false {
                    // Verifies that the list contains all variants
                    #[allow(unreachable_code)]
                    match panic!() {
                        $(
                            $path => (),
                        )*
                    }
                }
                [$($path),*]
            } }
        }
        let mut kinds = complete_list![
            Kind::Addable,
            Kind::Subtractable,
            Kind::Divisible,
            Kind::Numeric,
            Kind::Comparable,
            Kind::Equatable,
            Kind::Nullable,
            Kind::Record,
            Kind::Negatable,
            Kind::Timeable,
            Kind::Stringable,
        ];
        let mut str_kinds: Vec<_> = kinds.iter().map(|k| k.to_string()).collect();
        str_kinds.sort();

        kinds.sort();

        assert_eq!(
            kinds.iter().map(|k| k.to_string()).collect::<Vec<_>>(),
            str_kinds,
            "Expected that `Kind`s were specified in lexical order"
        );
    }
}

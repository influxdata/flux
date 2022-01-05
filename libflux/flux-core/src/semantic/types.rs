//! Semantic representations of types.

use std::{
    cell::Cell,
    cmp,
    collections::{BTreeMap, BTreeSet},
    fmt::{self, Write},
    hash::Hash,
    str::FromStr,
};

use codespan_reporting::diagnostic;
use derive_more::Display;
use serde::ser::{Serialize, Serializer};

use crate::{
    errors::{Errors, Located},
    map::HashMap,
    semantic::{
        fresh::{Fresh, Fresher},
        nodes::Symbol,
        sub::{apply2, apply3, apply4, merge_collect, Substitutable, Substituter, Substitution},
    },
};

/// For use in generics where the specific type of map is not mentioned.
pub type SemanticMap<K, V> = BTreeMap<K, V>;
#[allow(missing_docs)]
pub type SemanticMapIter<'a, K, V> = std::collections::btree_map::Iter<'a, K, V>;

struct Unifier<'a, E = Error> {
    sub: &'a mut Substitution,
    delayed_records: Vec<(Record, Record)>,
    errors: Errors<E>,
}

impl<'a, E> Unifier<'a, E> {
    fn new(sub: &'a mut Substitution) -> Self {
        Unifier {
            sub,
            delayed_records: Vec::new(),
            errors: Errors::new(),
        }
    }

    fn finish<T>(mut self, value: T) -> Result<T, Errors<E>>
    where
        E: From<Error>,
    {
        if !self.delayed_records.is_empty() {
            let mut sub_unifier = Unifier::new(self.sub);
            while let Some((expected, actual)) = self.delayed_records.pop() {
                let expected = expected.apply(sub_unifier.sub);
                let actual = actual.apply(sub_unifier.sub);
                expected.unify_now(&actual, &mut sub_unifier);
            }

            self.errors
                .extend(sub_unifier.errors.into_iter().map(E::from));
        }

        if self.errors.has_errors() {
            Err(self.errors)
        } else {
            Ok(value)
        }
    }
}

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
pub type PolyTypeMap<T = String> = SemanticMap<T, PolyType>;
/// Nested map of polytypes that preserves a sorted order when iterating
pub type PolyTypeMapMap = SemanticMap<String, PolyTypeMap>;

/// Map of identifier to a polytype.
pub type PolyTypeHashMap<T = String> = HashMap<T, PolyType>;

/// Alias the maplit literal construction macro so we can specify the type here.
#[macro_export]
macro_rules! semantic_map {
    ( $($x:tt)* ) => ( maplit::btreemap!( $($x)* ) );
}

impl fmt::Display for PolyType {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "{}", self.expr)?;
        if !self.cons.is_empty() {
            write!(f, " where {}", PolyType::display_constraints(&self.cons),)?;
        }
        Ok(())
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
    fn visit(&self, sub: &dyn Substituter) -> Option<Self> {
        sub.visit_poly_type(self)
    }

    fn walk(&self, sub: &dyn Substituter) -> Option<Self> {
        // `vars` defines new distinct variables for `expr` so any substitutions applied on a
        // variable named the same must not be applied in `expr`
        self.expr.visit(sub).map(|expr| PolyType {
            vars: self.vars.clone(),
            cons: self.cons.clone(),
            expr,
        })
    }
}

impl PolyType {
    pub(crate) fn error() -> Self {
        PolyType {
            vars: Vec::new(),
            cons: BTreeMap::new(),
            expr: MonoType::Error,
        }
    }

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

/// Errors that can be returned during type inference.
/// (Note that these error messages are read by end users.
/// This should be kept in mind when returning one of these errors.)
#[derive(Clone, Debug, PartialEq)]
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
        cause: Box<Error>,
    },
    MissingArgument(String),
    ExtraArgument(String),
    CannotUnifyArgument(String, Box<Error>),
    CannotUnifyReturn {
        exp: MonoType,
        act: MonoType,
        cause: Box<Error>,
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
                "expected {exp}{exp_info} but found {act}{act_info}",
                exp = exp.clone().fresh(&mut fresh, &mut TvarMap::new()),
                exp_info = exp.type_info(),
                act = act.clone().fresh(&mut fresh, &mut TvarMap::new()),
                act_info = act.type_info(),
            ),
            Error::CannotConstrain { exp, act } => write!(
                f,
                "{act}{act_info} is not {exp}",
                act = act.clone().fresh(&mut fresh, &mut TvarMap::new()),
                act_info = act.type_info(),
                exp = exp,
            ),
            Error::OccursCheck(tv, ty) => {
                write!(f, "recursive types not supported {} != {}", tv, ty)
            }
            Error::MissingLabel(a) => write!(f, "record is missing label {}", a),
            Error::ExtraLabel(a) => write!(f, "found unexpected label {}", a),
            Error::CannotUnifyLabel {
                lab,
                exp,
                act,
                cause,
            } => write!(
                f,
                "expected {exp}{exp_info} but found {act}{act_info} for label {lab} caused by {cause}",
                exp = exp.clone().fresh(&mut fresh, &mut TvarMap::new()),
                exp_info = exp.type_info(),
                act = act.clone().fresh(&mut fresh, &mut TvarMap::new()),
                act_info = act.type_info(),
                lab = lab,
                cause = cause
            ),
            Error::MissingArgument(x) => write!(f, "missing required argument {}", x),
            Error::ExtraArgument(x) => write!(f, "found unexpected argument {}", x),
            Error::CannotUnifyArgument(x, e) => write!(f, "{} (argument {})", e, x),
            Error::CannotUnifyReturn { exp, act, cause } => write!(
                f,
                "expected {exp}{exp_info} but found {act}{act_info} for return type caused by {cause}",
                exp = exp.clone().fresh(&mut fresh, &mut TvarMap::new()),
                exp_info = exp.type_info(),
                act = act.clone().fresh(&mut fresh, &mut TvarMap::new()),
                act_info = act.type_info(),
                cause = cause
            ),
            Error::MissingPipeArgument => write!(f, "missing pipe argument"),
            Error::MultiplePipeArguments { exp, act } => {
                write!(f, "expected pipe argument {} but found {}", exp, act)
            }
        }
    }
}

impl Substitutable for Error {
    fn walk(&self, sub: &dyn Substituter) -> Option<Self> {
        match self {
            Error::CannotUnify { exp, act } => {
                apply2(exp, act, sub).map(|(exp, act)| Error::CannotUnify { exp, act })
            }
            Error::CannotConstrain { exp, act } => act
                .visit(sub)
                .map(|act| Error::CannotConstrain { exp: *exp, act }),
            Error::OccursCheck(tv, ty) => ty.visit(sub).map(|ty| Error::OccursCheck(*tv, ty)),
            Error::CannotUnifyLabel {
                lab,
                exp,
                act,
                cause,
            } => apply3(exp, act, cause, sub).map(|(exp, act, cause)| Error::CannotUnifyLabel {
                lab: lab.clone(),
                exp,
                act,
                cause,
            }),
            Error::CannotUnifyArgument(x, e) => e
                .visit(sub)
                .map(|e| Error::CannotUnifyArgument(x.clone(), e)),
            Error::CannotUnifyReturn { exp, act, cause } => apply3(exp, act, cause, sub)
                .map(|(exp, act, cause)| Error::CannotUnifyReturn { exp, act, cause }),
            Error::MissingLabel(_)
            | Error::ExtraLabel(_)
            | Error::MissingArgument(_)
            | Error::ExtraArgument(_)
            | Error::MissingPipeArgument
            | Error::MultiplePipeArguments { .. } => None,
        }
    }
}

impl Error {
    pub(crate) fn as_diagnostic(&self) -> diagnostic::Diagnostic<()> {
        diagnostic::Diagnostic::error().with_message(self.to_string())
    }
}

/// Represents a constraint on a type variable to a specific kind (*i.e.*, a type class).
#[derive(Debug, Display, Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Hash)]
#[allow(missing_docs)]
// Kinds are ordered by name so that polytypes are displayed deterministically
pub enum Kind {
    Addable,
    Basic,
    Comparable,
    Divisible,
    Equatable,
    Label,
    Negatable,
    Nullable,
    Numeric,
    Record,
    Stringable,
    Subtractable,
    Timeable,
}

impl FromStr for Kind {
    type Err = ();

    fn from_str(name: &str) -> Result<Self, Self::Err> {
        Ok(match name {
            "Addable" => Kind::Addable,
            "Subtractable" => Kind::Subtractable,
            "Divisible" => Kind::Divisible,
            "Numeric" => Kind::Numeric,
            "Comparable" => Kind::Comparable,
            "Equatable" => Kind::Equatable,
            "Label" => Kind::Label,
            "Nullable" => Kind::Nullable,
            "Negatable" => Kind::Negatable,
            "Timeable" => Kind::Timeable,
            "Record" => Kind::Record,
            "Basic" => Kind::Basic,
            "Stringable" => Kind::Stringable,
            _ => return Err(()),
        })
    }
}

/// Pointer type used in `MonoType`
pub type Ptr<T> = std::sync::Arc<T>;

impl<T: Substitutable> Substitutable for Ptr<T> {
    fn walk(&self, sub: &dyn Substituter) -> Option<Self> {
        T::visit(self, sub).map(Ptr::new)
    }
}

/// An ordered map of string identifiers to monotypes.

/// Represents a Flux primitive primitive type such as int or string.
#[derive(Debug, Display, Clone, Copy, PartialEq, Serialize)]
#[allow(missing_docs)]
pub enum BuiltinType {
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
}

/// Represents a Flux type. The type may be unknown, represented as a type variable,
/// or may be a known concrete type.
#[derive(Debug, Display, Clone, PartialEq)]
#[allow(missing_docs)]
pub enum MonoType {
    #[display(fmt = "<error>")]
    Error,

    #[display(fmt = "{}", _0)]
    Builtin(BuiltinType),

    #[display(fmt = "\"{}\"", _0)]
    Label(Label),
    #[display(fmt = "#{}", _0)]
    Var(Tvar),

    /// A type variable that is bound to to a `PolyType` that this variable is contained in.
    #[display(fmt = "{}", _0)]
    BoundVar(Tvar),

    #[display(fmt = "{}", _0)]
    Collection(Ptr<Collection>),

    #[display(fmt = "{}", _0)]
    Dict(Ptr<Dictionary>),

    #[display(fmt = "{}", _0)]
    Record(Ptr<Record>),

    #[display(fmt = "{}", _0)]
    Fun(Ptr<Function>),
}

impl Serialize for MonoType {
    fn serialize<S>(&self, serializer: S) -> Result<<S as Serializer>::Ok, <S as Serializer>::Error>
    where
        S: Serializer,
    {
        // For backwards compatibility (and readability) we flatten the builtin variants
        #[derive(Serialize)]
        enum MonoTypeSer<'a> {
            Error,
            Bool,
            Int,
            Uint,
            Float,
            String,
            Duration,
            Time,
            Regexp,
            Bytes,
            Var(Tvar),
            Label(&'a Label),
            Arr(&'a MonoType),
            Dict(&'a Ptr<Dictionary>),
            Record(&'a Ptr<Record>),
            Fun(&'a Ptr<Function>),
            Vector(&'a MonoType),
            Stream(&'a MonoType),
        }

        match self {
            Self::Error => MonoTypeSer::Error,
            Self::Builtin(p) => match p {
                BuiltinType::Bool => MonoTypeSer::Bool,
                BuiltinType::Int => MonoTypeSer::Int,
                BuiltinType::Uint => MonoTypeSer::Uint,
                BuiltinType::Float => MonoTypeSer::Float,
                BuiltinType::String => MonoTypeSer::String,
                BuiltinType::Duration => MonoTypeSer::Duration,
                BuiltinType::Time => MonoTypeSer::Time,
                BuiltinType::Regexp => MonoTypeSer::Regexp,
                BuiltinType::Bytes => MonoTypeSer::Bytes,
            },
            // When serializing we tend to expect that all variables are already bound so treat
            // them the same here
            Self::BoundVar(v) | Self::Var(v) => MonoTypeSer::Var(*v),
            Self::Collection(p) => match p.collection {
                CollectionType::Array => MonoTypeSer::Arr(&p.arg),
                CollectionType::Vector => MonoTypeSer::Vector(&p.arg),
                CollectionType::Stream => MonoTypeSer::Stream(&p.arg),
            },
            Self::Label(p) => MonoTypeSer::Label(p),
            Self::Dict(p) => MonoTypeSer::Dict(p),
            Self::Record(p) => MonoTypeSer::Record(p),
            Self::Fun(p) => MonoTypeSer::Fun(p),
        }
        .serialize(serializer)
    }
}

#[derive(Debug, Clone, PartialEq)]
#[allow(missing_docs)]
pub struct Collection {
    pub collection: CollectionType,
    pub arg: MonoType,
}

impl fmt::Display for Collection {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match self.collection {
            CollectionType::Array => {
                write!(f, "[{}]", self.arg)
            }
            CollectionType::Vector => {
                write!(f, "v[{}]", self.arg)
            }
            CollectionType::Stream => {
                write!(f, "stream[{}]", self.arg)
            }
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq)]
#[allow(missing_docs)]
pub enum CollectionType {
    Array,
    Vector,
    Stream,
}

/// An ordered map of string identifiers to monotypes.
pub type MonoTypeMap<K = String, V = MonoType> = SemanticMap<K, V>;
#[allow(missing_docs)]
pub type MonoTypeVecMap<T = String> = SemanticMap<T, Vec<MonoType>>;
#[allow(missing_docs)]
type RefMonoTypeVecMap<'a, T = String> = HashMap<&'a T, Vec<&'a MonoType>>;

impl BuiltinType {
    fn unify(self, actual: Self, unifier: &mut Unifier<'_>) {
        if self != actual {
            unifier.errors.push(Error::CannotUnify {
                exp: self.into(),
                act: actual.into(),
            });
        }
    }

    fn constrain(self, with: Kind) -> Result<(), Error> {
        match self {
            BuiltinType::Bool => match with {
                Kind::Equatable | Kind::Nullable | Kind::Basic | Kind::Stringable => Ok(()),
                _ => Err(Error::CannotConstrain {
                    act: self.into(),
                    exp: with,
                }),
            },
            BuiltinType::Int => match with {
                Kind::Addable
                | Kind::Subtractable
                | Kind::Divisible
                | Kind::Numeric
                | Kind::Comparable
                | Kind::Equatable
                | Kind::Nullable
                | Kind::Basic
                | Kind::Stringable
                | Kind::Negatable => Ok(()),
                _ => Err(Error::CannotConstrain {
                    act: self.into(),
                    exp: with,
                }),
            },
            BuiltinType::Uint => match with {
                Kind::Addable
                | Kind::Subtractable
                | Kind::Divisible
                | Kind::Numeric
                | Kind::Comparable
                | Kind::Equatable
                | Kind::Nullable
                | Kind::Basic
                | Kind::Stringable
                | Kind::Negatable => Ok(()),
                _ => Err(Error::CannotConstrain {
                    act: self.into(),
                    exp: with,
                }),
            },
            BuiltinType::Float => match with {
                Kind::Addable
                | Kind::Subtractable
                | Kind::Divisible
                | Kind::Numeric
                | Kind::Comparable
                | Kind::Equatable
                | Kind::Nullable
                | Kind::Basic
                | Kind::Stringable
                | Kind::Negatable => Ok(()),
                _ => Err(Error::CannotConstrain {
                    act: self.into(),
                    exp: with,
                }),
            },
            BuiltinType::String => match with {
                Kind::Addable
                | Kind::Comparable
                | Kind::Equatable
                | Kind::Label
                | Kind::Nullable
                | Kind::Basic
                | Kind::Stringable => Ok(()),
                _ => Err(Error::CannotConstrain {
                    act: self.into(),
                    exp: with,
                }),
            },
            BuiltinType::Duration => match with {
                Kind::Comparable
                | Kind::Equatable
                | Kind::Nullable
                | Kind::Basic
                | Kind::Negatable
                | Kind::Stringable
                | Kind::Timeable => Ok(()),
                _ => Err(Error::CannotConstrain {
                    act: self.into(),
                    exp: with,
                }),
            },
            BuiltinType::Time => match with {
                Kind::Comparable
                | Kind::Equatable
                | Kind::Nullable
                | Kind::Basic
                | Kind::Timeable
                | Kind::Stringable => Ok(()),
                _ => Err(Error::CannotConstrain {
                    act: self.into(),
                    exp: with,
                }),
            },
            BuiltinType::Regexp => match with {
                Kind::Basic => Ok(()),
                _ => Err(Error::CannotConstrain {
                    act: self.into(),
                    exp: with,
                }),
            },
            BuiltinType::Bytes => match with {
                Kind::Equatable | Kind::Basic => Ok(()),
                _ => Err(Error::CannotConstrain {
                    act: self.into(),
                    exp: with,
                }),
            },
        }
    }
}

impl Substitutable for MonoType {
    fn visit(&self, sub: &dyn Substituter) -> Option<Self> {
        match sub.visit_type(self) {
            Some(typ) => Some(typ.walk(sub).unwrap_or(typ)),
            None => self.walk(sub),
        }
    }

    fn walk(&self, sub: &dyn Substituter) -> Option<Self> {
        match self {
            MonoType::Error
            | MonoType::Builtin(_)
            | MonoType::Label(_)
            | MonoType::BoundVar(_)
            | MonoType::Var(_) => None,
            MonoType::Collection(app) => app.visit(sub).map(MonoType::app),
            MonoType::Dict(dict) => dict.visit(sub).map(MonoType::dict),
            MonoType::Record(obj) => obj.visit(sub).map(MonoType::record),
            MonoType::Fun(fun) => fun.visit(sub).map(MonoType::fun),
        }
    }
}

impl From<Tvar> for MonoType {
    fn from(a: Tvar) -> MonoType {
        MonoType::Var(a)
    }
}

impl From<BuiltinType> for MonoType {
    fn from(t: BuiltinType) -> MonoType {
        MonoType::Builtin(t)
    }
}

impl From<Collection> for MonoType {
    fn from(a: Collection) -> MonoType {
        MonoType::Collection(Ptr::new(a))
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

#[allow(missing_docs)]
impl MonoType {
    pub const INT: MonoType = MonoType::Builtin(BuiltinType::Int);
    pub const UINT: MonoType = MonoType::Builtin(BuiltinType::Uint);
    pub const FLOAT: MonoType = MonoType::Builtin(BuiltinType::Float);
    pub const BOOL: MonoType = MonoType::Builtin(BuiltinType::Bool);
    pub const STRING: MonoType = MonoType::Builtin(BuiltinType::String);
    pub const TIME: MonoType = MonoType::Builtin(BuiltinType::Time);
    pub const REGEXP: MonoType = MonoType::Builtin(BuiltinType::Regexp);
    pub const BYTES: MonoType = MonoType::Builtin(BuiltinType::Bytes);
    pub const DURATION: MonoType = MonoType::Builtin(BuiltinType::Duration);
}

impl MonoType {
    /// Creates an application type
    pub fn app(a: impl Into<Ptr<Collection>>) -> Self {
        Self::Collection(a.into())
    }

    /// Creates an array type
    pub fn arr(arg: MonoType) -> Self {
        Self::app(Collection {
            collection: CollectionType::Array,
            arg,
        })
    }

    /// Creates a vector type
    pub fn vector(arg: MonoType) -> Self {
        Self::app(Collection {
            collection: CollectionType::Vector,
            arg,
        })
    }

    /// Creates a stream type
    pub fn stream(arg: MonoType) -> Self {
        Self::app(Collection {
            collection: CollectionType::Stream,
            arg,
        })
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
    pub fn try_unify(
        &self, // self represents the expected type
        actual: &Self,
        sub: &mut Substitution,
    ) -> Result<MonoType, Errors<Error>> {
        let mut unifier = Unifier::new(sub);

        let typ = self.unify(actual, &mut unifier);

        unifier.finish(typ)
    }

    fn unify(
        &self, // self represents the expected type
        actual: &Self,
        unifier: &mut Unifier<'_>,
    ) -> MonoType {
        log::debug!("Unify {} <=> {}", self, actual);
        match (self, actual) {
            // An error has already occurred so assume everything is ok here so that we do not
            // create additional, spurious errors
            (MonoType::Error, _) | (_, MonoType::Error) => (),

            (MonoType::Builtin(exp), MonoType::Builtin(act)) => exp.unify(*act, unifier),

            (MonoType::Label(l), MonoType::Label(r)) if l == r => {}
            (MonoType::Label(_), MonoType::Label(_))
            | (MonoType::Builtin(BuiltinType::String), MonoType::Label(_))
            | (MonoType::Label(_), MonoType::Builtin(BuiltinType::String)) => {
                return MonoType::STRING
            }

            (MonoType::Var(tv), MonoType::Var(tv2)) => {
                match (unifier.sub.try_apply(*tv), unifier.sub.try_apply(*tv2)) {
                    (Some(self_), Some(actual)) => {
                        self_.unify(&actual, unifier);
                    }
                    (Some(self_), None) => {
                        self_.unify(&MonoType::Var(*tv2), unifier);
                    }
                    (None, Some(actual)) => {
                        MonoType::Var(*tv).unify(&actual, unifier);
                    }
                    (None, None) => tv.unify(&MonoType::Var(*tv2), unifier),
                }
            }
            (MonoType::Var(tv), t) => match unifier.sub.try_apply(*tv) {
                Some(typ) => {
                    typ.unify(t, unifier);
                }
                None => tv.unify(t, unifier),
            },
            (t, MonoType::Var(tv)) => match unifier.sub.try_apply(*tv) {
                Some(typ) => {
                    t.unify(&typ, unifier);
                }
                None => tv.unify(t, unifier),
            },

            (MonoType::Collection(t), MonoType::Collection(s)) => t.unify(s, unifier),

            (MonoType::Dict(t), MonoType::Dict(s)) => t.unify(s, unifier),

            (MonoType::Record(t), MonoType::Record(s)) => t.unify(s, unifier),

            (MonoType::Fun(t), MonoType::Fun(s)) => t.unify(s, unifier),

            (exp, act) => {
                unifier.errors.push(Error::CannotUnify {
                    exp: exp.clone(),
                    act: act.clone(),
                });
            }
        }
        self.clone()
    }

    /// Validates that the current type meets the constraints of the specified kind.
    pub fn constrain(&self, with: Kind, cons: &mut TvarKinds) -> Result<(), Error> {
        match self {
            MonoType::Error => Ok(()),
            MonoType::Builtin(typ) => typ.constrain(with),
            // TODO Should constrain bound vars as well, but we can't just store it in `cons` as
            // they would override constraints of free variables
            MonoType::BoundVar(_) => Ok(()),
            MonoType::Label(_) => BuiltinType::String.constrain(with),
            MonoType::Var(tvr) => {
                tvr.constrain(with, cons);
                Ok(())
            }
            MonoType::Collection(app) => app.constrain(with, cons),
            MonoType::Dict(dict) => dict.constrain(with, cons),
            MonoType::Record(obj) => obj.constrain(with, cons),
            MonoType::Fun(fun) => fun.constrain(with, cons),
        }
    }

    fn contains(&self, tv: Tvar) -> bool {
        match self {
            MonoType::Error | MonoType::Builtin(_) | MonoType::Label(_) | MonoType::BoundVar(_) => {
                false
            }
            MonoType::Var(tvr) => tv == *tvr,
            MonoType::Collection(app) => app.contains(tv),
            MonoType::Dict(dict) => dict.contains(tv),
            MonoType::Record(row) => row.contains(tv),
            MonoType::Fun(fun) => fun.contains(tv),
        }
    }

    /// Returns the type of `param` if `self` is a function type
    pub fn parameter(&self, param: &str) -> Option<&MonoType> {
        match self {
            MonoType::Fun(f) => f.req.get(param).or_else(|| f.opt.get(param)).or_else(|| {
                f.pipe
                    .as_ref()
                    .and_then(|pipe| if pipe.k == param { Some(&pipe.v) } else { None })
            }),
            _ => None,
        }
    }

    /// Returns the type of `field` if `self` is a record type
    pub fn field(&self, field: &str) -> Option<&Property> {
        match self {
            MonoType::Record(r) => r.fields().find(|p| p.k == field),
            _ => None,
        }
    }

    /// Returns an iterator over the fields in the record (or an empty iterator of the type is not
    /// a record)
    pub fn fields(&self) -> impl Iterator<Item = &Property> {
        match self {
            MonoType::Record(r) => r.fields(),
            _ => {
                const RECORD: Record = Record::Empty;
                RECORD.fields()
            }
        }
    }

    fn type_info(&self) -> &str {
        match self {
            MonoType::Fun(_) => " (function)",
            MonoType::Dict(_) => " (dictionary)",
            MonoType::Record(_) => " (record)",
            MonoType::Collection(app) => match app.collection {
                CollectionType::Array => " (array)",
                CollectionType::Vector => " (vector)",
                CollectionType::Stream => "",
            },
            _ => "",
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
    fn unify_values(value1: &Self, value2: &Self) -> Result<Self, ena::unify::NoError> {
        unreachable!("We should never unify two values with each other within the substitution. If we reach this we did not resolve the variable before unifying {} <=> {}", value1, value2)
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

impl Tvar {
    fn unify(self, with: &MonoType, unifier: &mut Unifier<'_>) {
        match *with {
            MonoType::Var(tv) => {
                if self == tv {
                    // The empty substitution will always
                    // unify a type variable with itself.
                } else {
                    // Unify two distinct type variables.
                    // This will update the kind constraints
                    // associated with these type variables.
                    self.unify_with_tvar(tv, unifier);
                }
            }
            _ => {
                let with = with.apply_cow(unifier.sub);
                if with.contains(self) {
                    // Invalid recursive type
                    unifier
                        .errors
                        .push(Error::OccursCheck(self, with.into_owned()));
                } else {
                    // Unify a type variable with a monotype.
                    // The monotype must satisify any
                    // constraints placed on the type variable.
                    self.unify_with_type(with.into_owned(), unifier)
                }
            }
        }
    }

    fn unify_with_tvar(self, tv: Tvar, unifier: &mut Unifier<'_>) {
        unifier.sub.union(self, tv);
    }

    fn unify_with_type(self, t: MonoType, unifier: &mut Unifier<'_>) {
        if let Err(err) = unifier.sub.union_type(self, t) {
            unifier.errors.push(err);
        }
    }

    fn constrain(&self, with: Kind, cons: &mut TvarKinds) {
        match cons.get_mut(self) {
            Some(kinds) => {
                if !kinds.contains(&with) {
                    kinds.push(with);
                }
            }
            None => {
                cons.insert(*self, vec![with]);
            }
        }
    }
}

impl Substitutable for Collection {
    fn walk(&self, sub: &dyn Substituter) -> Option<Self> {
        self.arg.visit(sub).map(|arg| Collection {
            collection: self.collection,
            arg,
        })
    }
}

impl Collection {
    // self represents the expected type.
    fn unify(&self, with: &Self, unifier: &mut Unifier<'_>) {
        if self.collection != with.collection {
            unifier.errors.push(Error::CannotUnify {
                exp: MonoType::from(self.clone()),
                act: MonoType::from(with.clone()),
            });
        }
        self.arg.unify(&with.arg, unifier);
    }

    fn constrain(&self, with: Kind, cons: &mut TvarKinds) -> Result<(), Error> {
        match self.collection {
            CollectionType::Array | CollectionType::Stream => match with {
                Kind::Equatable => self.arg.constrain(with, cons),
                _ => Err(Error::CannotConstrain {
                    act: MonoType::app(self.clone()),
                    exp: with,
                }),
            },
            CollectionType::Vector => self.arg.constrain(with, cons),
        }
    }

    fn contains(&self, tv: Tvar) -> bool {
        self.arg.contains(tv)
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
    fn walk(&self, sub: &dyn Substituter) -> Option<Self> {
        apply2(&self.key, &self.val, sub).map(|(key, val)| Dictionary { key, val })
    }
}

impl Dictionary {
    fn unify(&self, actual: &Self, unifier: &mut Unifier<'_>) {
        self.key.unify(&actual.key, unifier);
        self.val.unify(&actual.val, unifier);
    }

    fn constrain(&self, with: Kind, _: &mut TvarKinds) -> Result<(), Error> {
        Err(Error::CannotConstrain {
            act: MonoType::dict(self.clone()),
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
#[derive(Clone, Serialize)]
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

impl fmt::Debug for Record {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        let (props, tail) = collect_record(self);
        f.debug_struct("Record")
            .field("fields", &props.into_iter().collect::<BTreeMap<_, _>>())
            .field("tail", &tail)
            .finish()
    }
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

fn collect_record(record: &Record) -> (RefMonoTypeVecMap<'_, RecordLabel>, Option<&MonoType>) {
    let mut a = RefMonoTypeVecMap::new();

    let mut fields = record.fields();
    for field in &mut fields {
        a.entry(&field.k).or_insert_with(Vec::new).push(&field.v);
    }
    (a, fields.tail())
}

impl cmp::PartialEq for Record {
    fn eq(&self, other: &Self) -> bool {
        let (a, t) = collect_record(self);
        let (b, v) = collect_record(other);
        t == v && a == b
    }
}

impl Substitutable for Record {
    fn walk(&self, sub: &dyn Substituter) -> Option<Self> {
        match self {
            Record::Empty => None,
            Record::Extension { head, tail } => {
                apply2(head, tail, sub).map(|(head, tail)| Record::Extension { head, tail })
            }
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
    fn unify(&self, actual: &Self, unifier: &mut Unifier<'_>) {
        let has_variable_label = |r: &Record| {
            r.fields().any(|prop| match prop.k {
                RecordLabel::Variable(v) => unifier.sub.try_apply(v).is_none(),
                RecordLabel::Concrete(_) => false,
            })
        };
        if has_variable_label(self) || has_variable_label(actual) {
            unifier.delayed_records.push((self.clone(), actual.clone()));
            return;
        }
        self.unify_now(actual, unifier)
    }

    fn unify_now(&self, actual: &Self, unifier: &mut Unifier<'_>) {
        match (self, actual) {
            (Record::Empty, Record::Empty) => (),
            (
                Record::Extension {
                    head: Property { k: a, v: t },
                    tail: MonoType::Var(l),
                },
                Record::Extension {
                    head: Property { k: b, v: u },
                    tail: MonoType::Var(r),
                },
            ) if a == b && l == r => unify_in_context(t, u, unifier, |e| Error::CannotUnifyLabel {
                lab: a.to_string(),
                exp: t.clone(),
                act: u.clone(),
                cause: Box::new(e),
            }),
            (
                Record::Extension {
                    head: Property { k: a, .. },
                    tail: MonoType::Var(l),
                },
                Record::Extension {
                    head: Property { k: b, .. },
                    tail: MonoType::Var(r),
                },
            ) if a != b && l == r => {
                unifier.errors.push(Error::CannotUnify {
                    exp: MonoType::from(self.clone()),
                    act: MonoType::from(actual.clone()),
                });
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
            ) if a == b => {
                t.unify(u, unifier);
                l.unify(r, unifier);
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
                match (a, b) {
                    (RecordLabel::Variable(a), RecordLabel::Concrete(b)) => {
                        if unifier.sub.try_apply(*a).is_none() {
                            a.unify(&MonoType::Label(b.clone()), unifier)
                        }
                    }
                    (RecordLabel::Concrete(a), RecordLabel::Variable(b)) => {
                        if unifier.sub.try_apply(*b).is_none() {
                            b.unify(&MonoType::Label(a.clone()), unifier)
                        }
                    }
                    _ => (),
                }
                let var = unifier.sub.fresh();
                let exp = MonoType::from(Record::Extension {
                    head: Property {
                        k: a.clone(),
                        v: t.clone(),
                    },
                    tail: MonoType::Var(var),
                });
                let act = MonoType::from(Record::Extension {
                    head: Property {
                        k: b.clone(),
                        v: u.clone(),
                    },
                    tail: MonoType::Var(var),
                });
                l.unify(&act, unifier);
                exp.unify(r, unifier);
            }
            // If we are expecting {a: u | r} but find {}, label `a` is missing.
            (
                Record::Extension {
                    head: Property { k: a, .. },
                    ..
                },
                Record::Empty,
            ) => {
                unifier.errors.push(Error::MissingLabel(a.to_string()));
            }
            // If we are expecting {} but find {a: u | r}, label `a` is extra.
            (
                Record::Empty,
                Record::Extension {
                    head: Property { k: a, .. },
                    ..
                },
            ) => {
                unifier.errors.push(Error::ExtraLabel(a.to_string()));
            }
            _ => {
                unifier.errors.push(Error::CannotUnify {
                    exp: MonoType::from(self.clone()),
                    act: MonoType::from(actual.clone()),
                });
            }
        }
    }

    fn constrain(&self, with: Kind, cons: &mut TvarKinds) -> Result<(), Error> {
        match with {
            Kind::Record => Ok(()),
            Kind::Equatable => {
                let mut fields = self.fields();
                for head in &mut fields {
                    head.v.constrain(with, cons)?;
                }
                match fields.tail() {
                    Some(t) => t.constrain(with, cons),
                    None => Ok(()),
                }
            }
            _ => Err(Error::CannotConstrain {
                act: MonoType::from(self.clone()),
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
        let mut fields = self.fields();
        for head in &mut fields {
            write!(f, "{}, ", head)?;
        }
        Ok(match fields.tail() {
            Some(MonoType::BoundVar(tv)) | Some(MonoType::Var(tv)) => Some(*tv),
            _ => None,
        })
    }

    /// Returns an iterator over the fields in the record
    pub fn fields(&self) -> FieldIter<'_> {
        FieldIter::Record(self)
    }
}

/// An iterator over a records fields
#[allow(missing_docs)]
pub enum FieldIter<'a> {
    Record(&'a Record),
    Tail(&'a MonoType),
}

impl<'a> FieldIter<'a> {
    /// Returns the tail of a `Record` once the iterator is exhausted or `None` if the record was
    /// bounded.
    pub fn tail(&self) -> Option<&'a MonoType> {
        match *self {
            FieldIter::Record(_) => None,
            FieldIter::Tail(tail) => Some(tail),
        }
    }
}

impl<'a> Iterator for FieldIter<'a> {
    type Item = &'a Property;

    fn next(&mut self) -> Option<Self::Item> {
        match self {
            FieldIter::Record(Record::Extension { head, tail }) => {
                match tail {
                    MonoType::Record(tail) => *self = FieldIter::Record(tail),
                    _ => *self = FieldIter::Tail(tail),
                }
                Some(head)
            }
            _ => None,
        }
    }
}

// Applies `context` to each error generated by unifying `exp` and `act`
fn unify_in_context<T>(
    exp: &MonoType,
    act: &T,
    unifier: &mut Unifier<'_, T::Error>,
    mut context: impl FnMut(Error) -> Error,
) where
    T: TypeLike,
{
    let mut sub_unifier = Unifier::new(unifier.sub);
    exp.unify(act.typ(), &mut sub_unifier);

    unifier.errors.extend(
        sub_unifier
            .errors
            .into_iter()
            .map(|e| act.error(context(e))),
    );

    unifier.delayed_records.extend(sub_unifier.delayed_records);
}

/// Labels in records that are allowed be variables
#[derive(Debug, Eq, PartialEq, Ord, PartialOrd, Hash, Clone, Serialize, derive_more::From)]
pub enum RecordLabel {
    /// A variable label
    Variable(Tvar),
    /// A concrete label
    Concrete(Label),
}

impl Substitutable for RecordLabel {
    fn walk(&self, sub: &dyn Substituter) -> Option<Self> {
        match self {
            Self::Variable(tvr) => sub.try_apply(*tvr).and_then(|new| match new {
                MonoType::Label(l) => Some(Self::Concrete(l)),
                MonoType::Var(l) => Some(Self::Variable(l)),
                _ => None,
            }),
            Self::Concrete(_) => None,
        }
    }
}

impl fmt::Display for RecordLabel {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match self {
            Self::Variable(v) => v.fmt(f),
            Self::Concrete(v) => v.fmt(f),
        }
    }
}

impl PartialEq<str> for RecordLabel {
    fn eq(&self, other: &str) -> bool {
        match self {
            Self::Variable(_) => false,
            Self::Concrete(l) => l == other,
        }
    }
}

impl PartialEq<&str> for RecordLabel {
    fn eq(&self, other: &&str) -> bool {
        *self == **other
    }
}

impl From<String> for RecordLabel {
    fn from(name: String) -> Self {
        Self::Concrete(Label::from(name))
    }
}

impl From<&str> for RecordLabel {
    fn from(name: &str) -> Self {
        Self::Concrete(Label::from(name))
    }
}

impl From<Symbol> for RecordLabel {
    fn from(name: Symbol) -> Self {
        Self::Concrete(Label::from(name))
    }
}

/// Wrapper around [`Symbol`] that ignores the package in comparisons. Allowing field lookups of
/// package exported labels to be done with local symbols
#[derive(Eq, Clone, Serialize)]
pub struct Label(Symbol);

impl std::hash::Hash for Label {
    fn hash<H: std::hash::Hasher>(&self, hasher: &mut H) {
        self.0.name().hash(hasher)
    }
}

impl std::ops::Deref for Label {
    type Target = str;
    fn deref(&self) -> &Self::Target {
        self.0.name()
    }
}

impl fmt::Debug for Label {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        f.debug_tuple("Label").field(&self.0.name()).finish()
    }
}

impl fmt::Display for Label {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        self.0.fmt(f)
    }
}

impl From<String> for Label {
    fn from(name: String) -> Self {
        Label(Symbol::from(name))
    }
}

impl From<Label> for String {
    fn from(name: Label) -> Self {
        name.to_string()
    }
}

impl From<Label> for Symbol {
    fn from(name: Label) -> Self {
        name.0
    }
}

impl From<&str> for Label {
    fn from(name: &str) -> Self {
        assert!(!name.contains('@'));
        Self(Symbol::from(name))
    }
}

impl From<Symbol> for Label {
    fn from(name: Symbol) -> Self {
        Self(if name.package().is_none() {
            name
        } else {
            Symbol::from(name.name())
        })
    }
}

impl PartialEq for Label {
    fn eq(&self, other: &Self) -> bool {
        self.0.name() == other.0.name()
    }
}

impl PartialEq<str> for Label {
    fn eq(&self, other: &str) -> bool {
        self.0.name() == other
    }
}

impl PartialEq<&str> for Label {
    fn eq(&self, other: &&str) -> bool {
        self.0.name() == *other
    }
}

impl PartialOrd for Label {
    fn partial_cmp(&self, other: &Self) -> Option<std::cmp::Ordering> {
        self.0.name().partial_cmp(other.0.name())
    }
}

impl Ord for Label {
    fn cmp(&self, other: &Self) -> std::cmp::Ordering {
        self.0.name().cmp(other.0.name())
    }
}

impl Label {
    /// Returns the inner [`Symbol`]
    pub fn as_symbol(&self) -> &Symbol {
        &self.0
    }
}

/// A key-value pair representing a property type in a record.
#[derive(Debug, Display, Clone, PartialEq, Serialize)]
#[display(fmt = "{}:{}", k, v)]
#[allow(missing_docs)]
pub struct Property<K = RecordLabel, V = MonoType> {
    pub k: K,
    pub v: V,
}

impl<K, V> Substitutable for Property<K, V>
where
    K: Substitutable + Clone,
    V: Substitutable + Clone,
{
    fn walk(&self, sub: &dyn Substituter) -> Option<Self> {
        self.v.visit(sub).map(|v| Property {
            k: self.k.clone(),
            v,
        })
    }
}

/// Represents a function type.
///
/// A function type is defined by a set of required arguments,
/// a set of optional arguments, an optional pipe argument, and
/// a required return type.
#[derive(Debug, Clone, PartialEq, Serialize)]
pub struct Function<T = MonoType> {
    /// Required arguments to a function.
    pub req: MonoTypeMap<String, T>,
    /// Optional arguments to a function.
    pub opt: MonoTypeMap<String, T>,
    /// An optional pipe argument.
    pub pipe: Option<Property<String, T>>,
    /// Required return type.
    pub retn: T,
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

impl<K: Eq + Hash + Clone> Substitutable for PolyTypeHashMap<K> {
    fn walk(&self, sub: &dyn Substituter) -> Option<Self> {
        merge_collect(
            &mut (),
            self.unordered_iter(),
            |_, (k, v)| v.visit(sub).map(|v| (k.clone(), v)),
            |_, (k, v)| (k.clone(), v.clone()),
        )
    }
}

impl<K: Ord + Clone, T: Substitutable + Clone> Substitutable for SemanticMap<K, T> {
    fn walk(&self, sub: &dyn Substituter) -> Option<Self> {
        merge_collect(
            &mut (),
            self,
            |_, (k, v)| v.visit(sub).map(|v| (k.clone(), v)),
            |_, (k, v)| (k.clone(), v.clone()),
        )
    }
}

impl<T: Substitutable> Substitutable for Option<T> {
    fn walk(&self, sub: &dyn Substituter) -> Option<Self> {
        match self {
            None => None,
            Some(t) => t.visit(sub).map(Some),
        }
    }
}

impl Substitutable for Function {
    fn walk(&self, sub: &dyn Substituter) -> Option<Self> {
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
}

impl Function {
    pub(crate) fn try_unify<T>(
        &self,
        actual: &Function<T>,
        sub: &mut Substitution,
    ) -> Result<(), Errors<T::Error>>
    where
        T: TypeLike + Clone,
        T::Error: From<Error>,
    {
        let mut unifier = Unifier::new(sub);

        self.unify(actual, &mut unifier);

        unifier.finish(())
    }

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
    fn unify<T>(&self, actual: &Function<T>, unifier: &mut Unifier<'_, T::Error>)
    where
        T: TypeLike + Clone,
    {
        // Some aliasing for coherence with the doc.
        let mut f = self.clone();
        let mut g = actual.clone();
        // Fix pipe arguments:
        // Make them required arguments with the correct name.
        match (f.pipe, g.pipe) {
            // Both functions have pipe arguments.
            (Some(fp), Some(gp)) => {
                if fp.k != "<-" && gp.k != "<-" && fp.k != gp.k {
                    // Both are named and the name differs, fail unification.
                    unifier
                        .errors
                        .push(gp.v.error(Error::MultiplePipeArguments {
                            exp: fp.k,
                            act: gp.k,
                        }));
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
                    unifier
                        .errors
                        .push(g.retn.error(Error::MissingPipeArgument));
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
                    unifier.errors.push(gp.v.error(Error::MissingPipeArgument));
                } else {
                    // This is a named argument, simply put it into the required ones.
                    g.req.insert(gp.k, gp.v);
                }
            }
            // Nothing to do.
            (None, None) => (),
        }
        // Now that f has not been consumed yet, check that every required argument in g is in f too.
        for (name, typ) in g.req.iter() {
            if !f.req.contains_key(name) && !f.opt.contains_key(name) {
                unifier
                    .errors
                    .push(typ.error(Error::ExtraArgument(String::from(name))));
            }
        }
        // Unify f's required arguments.

        let g_opt = &mut g.opt;
        for (name, exp) in f.req.into_iter() {
            if let Some(act) = g.req.remove(&name).or_else(|| g_opt.remove(&name)) {
                // The required argument is in g's required arguments.
                unify_in_context(&exp, &act, unifier, |e| {
                    Error::CannotUnifyArgument(name.clone(), Box::new(e))
                });
            } else {
                unifier
                    .errors
                    .push(g.retn.error(Error::MissingArgument(name)));
            }
        }
        // Unify f's optional arguments.
        for (name, exp) in f.opt.into_iter() {
            if let Some(act) = g.req.remove(&name).or_else(|| g_opt.remove(&name)) {
                unify_in_context(&exp, &act, unifier, |e| {
                    Error::CannotUnifyArgument(name.clone(), Box::new(e))
                });
            }
        }

        let f_retn = &f.retn;
        let g_retn = &g.retn;
        // Unify return types.
        unify_in_context(&f.retn, &g.retn, unifier, |cause| {
            Error::CannotUnifyReturn {
                exp: f_retn.clone(),
                act: g_retn.typ().clone(),
                cause: Box::new(cause),
            }
        })
    }

    fn constrain(&self, with: Kind, _: &mut TvarKinds) -> Result<(), Error> {
        Err(Error::CannotConstrain {
            act: MonoType::from(self.clone()),
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

pub(crate) trait TypeLike {
    type Error;
    fn typ(&self) -> &MonoType;
    fn into_type(self) -> MonoType;
    fn error(&self, error: Error) -> Self::Error;
}

impl TypeLike for MonoType {
    type Error = Error;
    fn typ(&self) -> &MonoType {
        self
    }
    fn into_type(self) -> MonoType {
        self
    }
    fn error(&self, error: Error) -> Error {
        error
    }
}

impl TypeLike for (MonoType, &'_ crate::ast::SourceLocation) {
    type Error = Located<Error>;
    fn typ(&self) -> &MonoType {
        &self.0
    }
    fn into_type(self) -> MonoType {
        self.0
    }
    fn error(&self, error: Error) -> Self::Error {
        Located {
            location: self.1.clone(),
            error,
        }
    }
}

/// Trait for returning the maximum type variable of a type.
pub trait MaxTvar {
    /// Return the maximum type variable of a type.
    fn max_tvar(&self) -> Option<Tvar>;
}

impl<T> MaxTvar for T
where
    T: Substitutable,
{
    fn max_tvar(&self) -> Option<Tvar> {
        #[derive(Default)]
        struct MaxTvars {
            max: Cell<Option<Tvar>>,
        }

        impl Substituter for MaxTvars {
            fn try_apply(&self, var: Tvar) -> Option<MonoType> {
                self.max.set(self.max.get().max(Some(var)));
                None
            }
        }

        let max = MaxTvars::default();
        self.visit(&max);
        max.max.into_inner()
    }
}

#[cfg(test)]
mod tests {
    use std::collections::BTreeMap;

    use super::*;
    use crate::{
        ast, parser,
        semantic::{
            convert::{convert_monotype, convert_polytype},
            infer,
        },
    };

    /// `polytype` is a utility method that returns a `PolyType` from a string.
    pub fn polytype(typ: &str) -> PolyType {
        let mut p = parser::Parser::new(typ);

        let typ_expr = p.parse_type_expression();
        if let Err(err) = ast::check::check(ast::walk::Node::TypeExpression(&typ_expr)) {
            panic!("TypeExpression parsing failed for {}. {:?}", typ, err);
        }
        convert_polytype(&typ_expr, &mut Substitution::default()).unwrap()
    }

    fn parse_type(
        expr: &str,
        tvars: &mut BTreeMap<String, Tvar>,
        sub: &mut Substitution,
    ) -> MonoType {
        let mut p = parser::Parser::new(expr);

        let typ_expr = p.parse_type_expression();
        if let Err(err) = ast::check::check(ast::walk::Node::TypeExpression(&typ_expr)) {
            panic!("TypeExpression parsing failed. {:?}", err);
        }
        convert_monotype(&typ_expr.monotype, tvars, sub).unwrap()
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
        assert_eq!("bool", MonoType::BOOL.to_string());
    }
    #[test]
    fn display_type_int() {
        assert_eq!("int", MonoType::INT.to_string());
    }
    #[test]
    fn display_type_uint() {
        assert_eq!("uint", MonoType::UINT.to_string());
    }
    #[test]
    fn display_type_float() {
        assert_eq!("float", MonoType::FLOAT.to_string());
    }
    #[test]
    fn display_type_string() {
        assert_eq!("string", MonoType::STRING.to_string());
    }
    #[test]
    fn display_type_duration() {
        assert_eq!("duration", MonoType::DURATION.to_string());
    }
    #[test]
    fn display_type_time() {
        assert_eq!("time", MonoType::TIME.to_string());
    }
    #[test]
    fn display_type_regexp() {
        assert_eq!("regexp", MonoType::REGEXP.to_string());
    }
    #[test]
    fn display_type_bytes() {
        assert_eq!("bytes", MonoType::BYTES.to_string());
    }
    #[test]
    fn display_type_tvar() {
        assert_eq!("t10", MonoType::BoundVar(Tvar(10)).to_string());
    }
    #[test]
    fn display_type_array() {
        assert_eq!("[int]", MonoType::arr(MonoType::INT).to_string());
    }
    #[test]
    fn display_type_vector() {
        assert_eq!("v[int]", MonoType::vector(MonoType::INT).to_string());
    }
    #[test]
    fn display_type_record() {
        assert_eq!(
            "{A with a:int, b:string}",
            Record::new(
                [
                    Property {
                        k: RecordLabel::from("a"),
                        v: MonoType::INT,
                    },
                    Property {
                        k: RecordLabel::from("b"),
                        v: MonoType::STRING,
                    }
                ],
                Some(MonoType::BoundVar(Tvar(0))),
            )
            .to_string()
        );
        assert_eq!(
            "{a:int, b:string}",
            Record::new(
                [
                    Property {
                        k: RecordLabel::from("a"),
                        v: MonoType::INT,
                    },
                    Property {
                        k: RecordLabel::from("b"),
                        v: MonoType::STRING,
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
                retn: MonoType::INT,
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
                    v: MonoType::INT,
                }),
                retn: MonoType::INT,
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
                    v: MonoType::INT,
                }),
                retn: MonoType::INT,
            }
            .to_string()
        );
        assert_eq!(
            "(<-:int, a:int, b:int) => int",
            Function {
                req: semantic_map! {
                    String::from("a") => MonoType::INT,
                    String::from("b") => MonoType::INT,
                },
                opt: MonoTypeMap::new(),
                pipe: Some(Property {
                    k: String::from("<-"),
                    v: MonoType::INT,
                }),
                retn: MonoType::INT,
            }
            .to_string()
        );
        assert_eq!(
            "(<-:int, ?a:int, ?b:int) => int",
            Function {
                req: MonoTypeMap::new(),
                opt: semantic_map! {
                    String::from("a") => MonoType::INT,
                    String::from("b") => MonoType::INT,
                },
                pipe: Some(Property {
                    k: String::from("<-"),
                    v: MonoType::INT,
                }),
                retn: MonoType::INT,
            }
            .to_string()
        );
        assert_eq!(
            "(<-:int, a:int, b:int, ?c:int, ?d:int) => int",
            Function {
                req: semantic_map! {
                    String::from("a") => MonoType::INT,
                    String::from("b") => MonoType::INT,
                },
                opt: semantic_map! {
                    String::from("c") => MonoType::INT,
                    String::from("d") => MonoType::INT,
                },
                pipe: Some(Property {
                    k: String::from("<-"),
                    v: MonoType::INT,
                }),
                retn: MonoType::INT,
            }
            .to_string()
        );
        assert_eq!(
            "(a:int, ?b:bool) => int",
            Function {
                req: semantic_map! {
                    String::from("a") => MonoType::INT,
                },
                opt: semantic_map! {
                    String::from("b") => MonoType::BOOL,
                },
                pipe: None,
                retn: MonoType::INT,
            }
            .to_string()
        );
        assert_eq!(
            "(<-a:int, b:int, c:int, ?d:bool) => int",
            Function {
                req: semantic_map! {
                    String::from("b") => MonoType::INT,
                    String::from("c") => MonoType::INT,
                },
                opt: semantic_map! {
                    String::from("d") => MonoType::BOOL,
                },
                pipe: Some(Property {
                    k: String::from("a"),
                    v: MonoType::INT,
                }),
                retn: MonoType::INT,
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
                expr: MonoType::INT,
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
                        String::from("x") => MonoType::BoundVar(Tvar(0)),
                    },
                    opt: MonoTypeMap::new(),
                    pipe: None,
                    retn: MonoType::BoundVar(Tvar(0)),
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
                        String::from("x") => MonoType::BoundVar(Tvar(0)),
                        String::from("y") => MonoType::BoundVar(Tvar(1)),
                    },
                    opt: MonoTypeMap::new(),
                    pipe: None,
                    retn: MonoType::from(Record::new(
                        [
                            Property {
                                k: RecordLabel::from("x"),
                                v: MonoType::BoundVar(Tvar(0)),
                            },
                            Property {
                                k: RecordLabel::from("y"),
                                v: MonoType::BoundVar(Tvar(1)),
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
                        String::from("a") => MonoType::BoundVar(Tvar(0)),
                        String::from("b") => MonoType::BoundVar(Tvar(0)),
                    },
                    opt: MonoTypeMap::new(),
                    pipe: None,
                    retn: MonoType::BoundVar(Tvar(0)),
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
                        String::from("x") => MonoType::BoundVar(Tvar(0)),
                        String::from("y") => MonoType::BoundVar(Tvar(1)),
                    },
                    opt: MonoTypeMap::new(),
                    pipe: None,
                    retn: MonoType::from(Record::new(
                        [
                            Property {
                                k: RecordLabel::from("x"),
                                v: MonoType::BoundVar(Tvar(0)),
                            },
                            Property {
                                k: RecordLabel::from("y"),
                                v: MonoType::BoundVar(Tvar(1)),
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
                        String::from("x") => MonoType::BoundVar(Tvar(0)),
                        String::from("y") => MonoType::BoundVar(Tvar(1)),
                    },
                    opt: MonoTypeMap::new(),
                    pipe: None,
                    retn: MonoType::from(Record::new(
                        [
                            Property {
                                k: RecordLabel::from("x"),
                                v: MonoType::BoundVar(Tvar(0)),
                            },
                            Property {
                                k: RecordLabel::from("y"),
                                v: MonoType::BoundVar(Tvar(1)),
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
                        k: RecordLabel::from("a"),
                        v: MonoType::INT,
                    },
                    Property {
                        k: RecordLabel::from("b"),
                        v: MonoType::STRING,
                    }
                ],
                Some(MonoType::BoundVar(Tvar(0))),
            )),
            // {A with b:string, a:int}
            MonoType::from(Record::new(
                [
                    Property {
                        k: RecordLabel::from("b"),
                        v: MonoType::STRING,
                    },
                    Property {
                        k: RecordLabel::from("a"),
                        v: MonoType::INT,
                    }
                ],
                Some(MonoType::BoundVar(Tvar(0))),
            )),
        );
        assert_eq!(
            // {A with a:int, b:string, b:int, c:float}
            MonoType::from(Record::new(
                [
                    Property {
                        k: RecordLabel::from("a"),
                        v: MonoType::INT,
                    },
                    Property {
                        k: RecordLabel::from("b"),
                        v: MonoType::STRING,
                    },
                    Property {
                        k: RecordLabel::from("b"),
                        v: MonoType::INT,
                    },
                    Property {
                        k: RecordLabel::from("c"),
                        v: MonoType::FLOAT,
                    }
                ],
                Some(MonoType::BoundVar(Tvar(0))),
            )),
            // {A with c:float, b:string, b:int, a:int}
            MonoType::from(Record::new(
                [
                    Property {
                        k: RecordLabel::from("c"),
                        v: MonoType::FLOAT,
                    },
                    Property {
                        k: RecordLabel::from("b"),
                        v: MonoType::STRING,
                    },
                    Property {
                        k: RecordLabel::from("b"),
                        v: MonoType::INT,
                    },
                    Property {
                        k: RecordLabel::from("a"),
                        v: MonoType::INT,
                    }
                ],
                Some(MonoType::BoundVar(Tvar(0))),
            ))
        );
        assert_ne!(
            // {A with a:int, b:string, b:int, c:float}
            MonoType::from(Record::new(
                [
                    Property {
                        k: RecordLabel::from("a"),
                        v: MonoType::INT,
                    },
                    Property {
                        k: RecordLabel::from("b"),
                        v: MonoType::STRING,
                    },
                    Property {
                        k: RecordLabel::from("b"),
                        v: MonoType::INT,
                    },
                    Property {
                        k: RecordLabel::from("c"),
                        v: MonoType::FLOAT,
                    }
                ],
                Some(MonoType::BoundVar(Tvar(0))),
            )),
            // {A with a:int, b:int, b:string, c:float}
            MonoType::from(Record::new(
                [
                    Property {
                        k: RecordLabel::from("a"),
                        v: MonoType::INT,
                    },
                    Property {
                        k: RecordLabel::from("b"),
                        v: MonoType::INT,
                    },
                    Property {
                        k: RecordLabel::from("b"),
                        v: MonoType::STRING,
                    },
                    Property {
                        k: RecordLabel::from("c"),
                        v: MonoType::FLOAT,
                    }
                ],
                Some(MonoType::BoundVar(Tvar(0))),
            ))
        );
        assert_ne!(
            // {a:int, b:string}
            MonoType::from(Record::new(
                [
                    Property {
                        k: RecordLabel::from("a"),
                        v: MonoType::INT,
                    },
                    Property {
                        k: RecordLabel::from("b"),
                        v: MonoType::STRING,
                    }
                ],
                Some(MonoType::from(Record::Empty)),
            )),
            // {b:int, a:int}
            MonoType::from(Record::new(
                [
                    Property {
                        k: RecordLabel::from("b"),
                        v: MonoType::INT,
                    },
                    Property {
                        k: RecordLabel::from("a"),
                        v: MonoType::INT,
                    }
                ],
                Some(MonoType::from(Record::Empty)),
            ))
        );
        assert_ne!(
            // {a:int}
            MonoType::from(Record::Extension {
                head: Property {
                    k: RecordLabel::from("a"),
                    v: MonoType::INT,
                },
                tail: MonoType::from(Record::Empty),
            }),
            // {A with a:int}
            MonoType::from(Record::Extension {
                head: Property {
                    k: RecordLabel::from("a"),
                    v: MonoType::INT,
                },
                tail: MonoType::BoundVar(Tvar(0)),
            }),
        );
        assert_ne!(
            // {A with a:int}
            MonoType::from(Record::Extension {
                head: Property {
                    k: RecordLabel::from("a"),
                    v: MonoType::INT,
                },
                tail: MonoType::BoundVar(Tvar(0)),
            }),
            // {B with a:int}
            MonoType::from(Record::Extension {
                head: Property {
                    k: RecordLabel::from("a"),
                    v: MonoType::INT,
                },
                tail: MonoType::BoundVar(Tvar(1)),
            }),
        );
    }

    #[test]
    fn unify_ints() {
        MonoType::INT
            .try_unify(&MonoType::INT, &mut Substitution::default())
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
            MonoType::INT.constrain(c, &mut TvarKinds::new()).unwrap();
        }

        let sub = MonoType::INT
            .constrain(Kind::Record, &mut TvarKinds::new())
            .map(|_| ());
        assert_eq!(
            Err(Error::CannotConstrain {
                act: MonoType::INT,
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
        // kind constraints allowed for Vector(MonoType::INT)
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
            let vector_int = MonoType::vector(MonoType::INT);
            vector_int.constrain(c, &mut TvarKinds::new()).unwrap();
        }

        // kind constraints not allowed for Vector(MonoType::STRING)
        let unallowable_cons_string = vec![Kind::Subtractable, Kind::Divisible, Kind::Numeric];
        for c in unallowable_cons_string {
            let vector_string = MonoType::vector(MonoType::STRING);
            let sub = vector_string
                .constrain(c, &mut TvarKinds::new())
                .map(|_| ());
            assert_eq!(
                Err(Error::CannotConstrain {
                    act: MonoType::STRING,
                    exp: c
                }),
                sub
            );
        }

        // kind constraints not allowed for Vector(MonoType::TIME)
        let unallowable_cons_time = vec![Kind::Subtractable, Kind::Divisible, Kind::Numeric];
        for c in unallowable_cons_time {
            let vector_time = MonoType::vector(MonoType::TIME);
            let sub = vector_time.constrain(c, &mut TvarKinds::new()).map(|_| ());
            assert_eq!(
                Err(Error::CannotConstrain {
                    act: MonoType::TIME,
                    exp: c
                }),
                sub
            );
        }

        // kind constraints allowed for Vector(MonoType::TIME)
        let allowable_cons_time = vec![
            Kind::Comparable,
            Kind::Equatable,
            Kind::Nullable,
            Kind::Stringable,
            Kind::Timeable,
        ];

        for c in allowable_cons_time {
            let vector_time = MonoType::vector(MonoType::TIME);
            vector_time.constrain(c, &mut TvarKinds::new()).unwrap();
        }
    }
    #[test]
    fn unify_error() {
        let err = MonoType::INT
            .try_unify(&MonoType::STRING, &mut Substitution::default())
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
            .try_unify(&MonoType::Var(Tvar(1)), &mut sub)
            .unwrap();
        assert_eq!(sub.apply(Tvar(0)), sub.apply(Tvar(1)));
    }
    #[test]
    fn unify_constrained_tvars() {
        let mut sub = Substitution::default();
        sub.cons()
            .extend(semantic_map! {Tvar(0) => vec![Kind::Addable, Kind::Divisible]});
        sub.mk_fresh(2);
        MonoType::Var(Tvar(0))
            .try_unify(&MonoType::Var(Tvar(1)), &mut sub)
            .unwrap();
        assert_eq!(sub.apply(Tvar(0)), MonoType::Var(Tvar(1)));
        assert_eq!(
            sub.cons(),
            &semantic_map! {Tvar(1) => vec![Kind::Addable, Kind::Divisible]},
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
            let mut sub = Substitution::default();
            sub.cons().extend(f_cons.into_iter().chain(g_cons));
            sub.mk_fresh(2);
            let res = f.clone().try_unify(&g, &mut sub);
            assert!(res.is_err());
            let res = g.clone().try_unify(&f, &mut sub);
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
            let mut sub = Substitution::default();
            sub.cons().extend(f_cons.into_iter().chain(g_cons));
            sub.mk_fresh(2);
            let res = f.clone().try_unify(&g, &mut sub);
            assert!(res.is_err());
            let res = g.try_unify(&f, &mut sub);
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
                "a".to_string() => MonoType::INT,
                "b".to_string() => MonoType::INT,
            },
            opt: semantic_map! {},
            pipe: None,
            retn: MonoType::INT,
        };

        let mut sub = Substitution::default();
        let (fn_type, cons) = infer::instantiate(fn_type, &mut sub, Default::default());
        infer::solve(&cons, &mut sub).unwrap();
        if let MonoType::Fun(f) = fn_type {
            sub.mk_fresh(2);
            f.try_unify(&call_type, &mut sub).unwrap();
            assert_eq!(sub.apply(Tvar(0)), MonoType::INT);
            // the constraint on A gets removed.
            assert_eq!(
                sub.cons(),
                &semantic_map! {Tvar(1) => vec![Kind::Divisible]}
            );
        } else {
            panic!("the monotype under examination is not a function");
        }
    }
    #[test]
    fn unify_higher_order_functions() {
        let mut sub = Substitution::default();

        let f = polytype(
            "(a: A, b: A, ?c: (a: A) => B) => (d:  string) => A where A: Addable, B: Divisible ",
        );
        let g = polytype("(a: int, b: int, c: (a: int) => float) => (d: string) => int");

        let (f, cons) = infer::instantiate(f, &mut sub, Default::default());
        infer::solve(&cons, &mut sub).unwrap();

        let (g, cons) = infer::instantiate(g, &mut sub, Default::default());
        infer::solve(&cons, &mut sub).unwrap();

        if let (MonoType::Fun(f), MonoType::Fun(g)) = (f, g) {
            sub.mk_fresh(2);
            f.try_unify(&g, &mut sub).unwrap();
            assert_eq!(sub.apply(Tvar(0)), MonoType::INT);
            assert_eq!(sub.apply(Tvar(1)), MonoType::FLOAT);
            // we know everything about tvars, there is no constraint.
            assert_eq!(sub.cons(), &semantic_map! {});
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
                .try_unify(&parse_type($actual, &mut tvars, &mut sub), &mut sub)
                .unwrap_or_else(|err| panic!("{}", err));
        }};
    }

    macro_rules! assert_unify_err {
        ($expected: expr, $actual: expr $(, $pat: pat)? $(,)?) => {{
            let mut sub = Substitution::default();
            let mut tvars = BTreeMap::new();
            let result = parse_type($expected, &mut tvars, &mut sub).try_unify(
                &parse_type($actual, &mut tvars, &mut sub),
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
            Kind::Label,
            Kind::Nullable,
            Kind::Record,
            Kind::Negatable,
            Kind::Timeable,
            Kind::Basic,
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

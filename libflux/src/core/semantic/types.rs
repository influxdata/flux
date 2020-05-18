use crate::semantic::fresh::{Fresh, Fresher};
use crate::semantic::sub::{Substitutable, Substitution};

use std::{
    cmp,
    collections::{BTreeMap, BTreeSet, HashMap},
    fmt,
};

// For use in generics where the specific type of map is not not mentioned.
pub type SemanticMap<K, V> = BTreeMap<K, V>;
pub type SemanticMapIter<'a, K, V> = std::collections::btree_map::Iter<'a, K, V>;

#[derive(Debug, Clone)]
pub struct PolyType {
    pub vars: Vec<Tvar>,
    pub cons: TvarKinds,
    pub expr: MonoType,
}

pub type PolyTypeMap = SemanticMap<String, PolyType>;
pub type PolyTypeMapMap = SemanticMap<String, SemanticMap<String, PolyType>>;

#[macro_export]
/// Alias the maplit literal construction macro so we can specify the type here.
macro_rules! semantic_map {
    ( $($x:tt)* ) => ( maplit::btreemap!( $($x)* ) );
}

impl fmt::Display for PolyType {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        let vars = &self
            .vars
            .iter()
            .map(|x| x.to_string())
            .collect::<Vec<_>>()
            .join(", ");
        if self.cons.is_empty() {
            write!(f, "forall [{}] {}", vars, self.expr)
        } else {
            write!(
                f,
                "forall [{}] where {} {}",
                vars,
                PolyType::display_constraints(&self.cons),
                self.expr
            )
        }
    }
}

impl PartialEq for PolyType {
    fn eq(&self, poly: &Self) -> bool {
        let a: Tvar = self.max_tvar();
        let b: Tvar = poly.max_tvar();

        let max = if a > b { a.0 } else { b.0 };

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
    fn apply(self, sub: &Substitution) -> Self {
        PolyType {
            vars: self.vars,
            cons: self.cons,
            expr: self.expr.apply(sub),
        }
    }
    fn free_vars(&self) -> Vec<Tvar> {
        minus(&self.vars, self.expr.free_vars())
    }
}

impl MaxTvar for Vec<Tvar> {
    fn max_tvar(&self) -> Tvar {
        self.iter()
            .fold(Tvar(0), |max, tv| if *tv > max { *tv } else { max })
    }
}

impl MaxTvar for PolyType {
    fn max_tvar(&self) -> Tvar {
        vec![self.vars.max_tvar(), self.expr.max_tvar()].max_tvar()
    }
}

impl PolyType {
    fn display_constraints(cons: &TvarKinds) -> String {
        cons.iter()
            // A BTree produces a sorted iterator for
            // deterministic display output
            .collect::<BTreeMap<_, _>>()
            .iter()
            .map(|(&&tv, &kinds)| format!("{}:{}", tv, PolyType::display_kinds(kinds)))
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
}

pub fn union<T: PartialEq>(mut vars: Vec<T>, mut with: Vec<T>) -> Vec<T> {
    with.retain(|tv| !vars.contains(tv));
    vars.append(&mut with);
    vars
}

pub fn minus<T: PartialEq>(vars: &[T], mut from: Vec<T>) -> Vec<T> {
    from.retain(|tv| !vars.contains(tv));
    from
}

#[derive(Debug, PartialEq)]
pub enum Error {
    CannotUnify(MonoType, MonoType),
    CannotConstrain(MonoType, Kind),
    OccursCheck(Tvar, MonoType),
    MissingLabel(String),
    ExtraLabel(String),
    CannotUnifyLabel(String, MonoType, MonoType),
    MissingArgument(String),
    ExtraArgument(String),
    CannotUnifyArgument(String, Box<Error>),
    CannotUnifyReturn(MonoType, MonoType),
    MissingPipeArgument,
    MultiplePipeArguments(String, String),
}

impl fmt::Display for Error {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match self {
            Error::CannotUnify(expected, found) => write!(f, "{} != {}", found, expected),
            Error::CannotConstrain(found, expected) => write!(f, "{} is not {}", found, expected),
            Error::OccursCheck(tv, ty) => write!(f, "{} != {} (recursive type)", tv, ty),
            Error::MissingLabel(a) => write!(f, "record is missing label {}", a),
            Error::ExtraLabel(a) => write!(f, "found unexpected label {}", a),
            Error::CannotUnifyLabel(a, t, with) => {
                write!(f, "{} != {} (record label {})", t, with, a)
            }
            Error::MissingArgument(x) => write!(f, "missing required argument {}", x),
            Error::ExtraArgument(x) => write!(f, "found unexpected argument {}", x),
            Error::CannotUnifyArgument(x, e) => write!(f, "{} (argument {})", e, x),
            Error::CannotUnifyReturn(expected, found) => {
                write!(f, "return type {} != {}", found, expected)
            }
            Error::MissingPipeArgument => write!(f, "missing pipe argument"),
            Error::MultiplePipeArguments(a, b) => {
                write!(f, "inconsistent pipe arguments {} != {}", a, b)
            }
        }
    }
}

// Kind represents a class or family of types
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub enum Kind {
    Addable,
    Subtractable,
    Divisible,
    Numeric,
    Comparable,
    Equatable,
    Nullable,
    Row,
    Negatable,
}

impl fmt::Display for Kind {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match self {
            Kind::Addable => f.write_str("Addable"),
            Kind::Subtractable => f.write_str("Subtractable"),
            Kind::Divisible => f.write_str("Divisible"),
            Kind::Numeric => f.write_str("Numeric"),
            Kind::Comparable => f.write_str("Comparable"),
            Kind::Equatable => f.write_str("Equatable"),
            Kind::Nullable => f.write_str("Nullable"),
            Kind::Row => f.write_str("Row"),
            Kind::Negatable => f.write_str("Negatable"),
        }
    }
}

// Kinds are ordered by name so that polytypes are displayed deterministically
impl cmp::Ord for Kind {
    fn cmp(&self, other: &Self) -> cmp::Ordering {
        self.to_string().cmp(&other.to_string())
    }
}

// Kinds are ordered by name so that polytypes are displayed deterministically
impl cmp::PartialOrd for Kind {
    fn partial_cmp(&self, other: &Self) -> Option<cmp::Ordering> {
        Some(self.cmp(other))
    }
}

// MonoType represents a specific named type
#[derive(Debug, Clone, PartialEq)]
pub enum MonoType {
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
    Arr(Box<Array>),
    Row(Box<Row>),
    Fun(Box<Function>),
}

pub type MonoTypeMap = SemanticMap<String, MonoType>;
pub type MonoTypeVecMap = SemanticMap<String, Vec<MonoType>>;
type RefMonoTypeVecMap<'a> = HashMap<&'a String, Vec<&'a MonoType>>;

impl fmt::Display for MonoType {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match self {
            MonoType::Bool => f.write_str("bool"),
            MonoType::Int => f.write_str("int"),
            MonoType::Uint => f.write_str("uint"),
            MonoType::Float => f.write_str("float"),
            MonoType::String => f.write_str("string"),
            MonoType::Duration => f.write_str("duration"),
            MonoType::Time => f.write_str("time"),
            MonoType::Regexp => f.write_str("regexp"),
            MonoType::Bytes => f.write_str("bytes"),
            MonoType::Var(var) => var.fmt(f),
            MonoType::Arr(arr) => arr.fmt(f),
            MonoType::Row(obj) => obj.fmt(f),
            MonoType::Fun(fun) => fun.fmt(f),
        }
    }
}

impl Substitutable for MonoType {
    fn apply(self, sub: &Substitution) -> Self {
        match self {
            MonoType::Bool
            | MonoType::Int
            | MonoType::Uint
            | MonoType::Float
            | MonoType::String
            | MonoType::Duration
            | MonoType::Time
            | MonoType::Regexp
            | MonoType::Bytes => self,
            MonoType::Var(tvr) => sub.apply(tvr),
            MonoType::Arr(arr) => MonoType::Arr(Box::new(arr.apply(sub))),
            MonoType::Row(obj) => MonoType::Row(Box::new(obj.apply(sub))),
            MonoType::Fun(fun) => MonoType::Fun(Box::new(fun.apply(sub))),
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
            MonoType::Row(obj) => obj.free_vars(),
            MonoType::Fun(fun) => fun.free_vars(),
        }
    }
}

impl MaxTvar for MonoType {
    fn max_tvar(&self) -> Tvar {
        match self {
            MonoType::Bool
            | MonoType::Int
            | MonoType::Uint
            | MonoType::Float
            | MonoType::String
            | MonoType::Duration
            | MonoType::Time
            | MonoType::Regexp
            | MonoType::Bytes => Tvar(0),
            MonoType::Var(tvr) => tvr.max_tvar(),
            MonoType::Arr(arr) => arr.max_tvar(),
            MonoType::Row(obj) => obj.max_tvar(),
            MonoType::Fun(fun) => fun.max_tvar(),
        }
    }
}

impl From<Row> for MonoType {
    fn from(r: Row) -> MonoType {
        MonoType::Row(Box::new(r))
    }
}

impl MonoType {
    // self represents the expected type
    pub fn unify(
        self,
        actual: Self,
        cons: &mut TvarKinds,
        f: &mut Fresher,
    ) -> Result<Substitution, Error> {
        match (self, actual) {
            (MonoType::Bool, MonoType::Bool)
            | (MonoType::Int, MonoType::Int)
            | (MonoType::Uint, MonoType::Uint)
            | (MonoType::Float, MonoType::Float)
            | (MonoType::String, MonoType::String)
            | (MonoType::Duration, MonoType::Duration)
            | (MonoType::Time, MonoType::Time)
            | (MonoType::Regexp, MonoType::Regexp)
            | (MonoType::Bytes, MonoType::Bytes) => Ok(Substitution::empty()),
            (MonoType::Var(tv), t) => tv.unify(t, cons),
            (t, MonoType::Var(tv)) => tv.unify(t, cons),
            (MonoType::Arr(t), MonoType::Arr(s)) => t.unify(*s, cons, f),
            (MonoType::Row(t), MonoType::Row(s)) => t.unify(*s, cons, f),
            (MonoType::Fun(t), MonoType::Fun(s)) => t.unify(*s, cons, f),
            (exp, act) => Err(Error::CannotUnify(exp, act)),
        }
    }

    pub fn constrain(self, with: Kind, cons: &mut TvarKinds) -> Result<Substitution, Error> {
        match self {
            MonoType::Bool => match with {
                Kind::Equatable | Kind::Nullable => Ok(Substitution::empty()),
                _ => Err(Error::CannotConstrain(self, with)),
            },
            MonoType::Int => match with {
                Kind::Addable
                | Kind::Subtractable
                | Kind::Divisible
                | Kind::Numeric
                | Kind::Comparable
                | Kind::Equatable
                | Kind::Nullable
                | Kind::Negatable => Ok(Substitution::empty()),
                _ => Err(Error::CannotConstrain(self, with)),
            },
            MonoType::Uint => match with {
                Kind::Addable
                | Kind::Subtractable
                | Kind::Divisible
                | Kind::Numeric
                | Kind::Comparable
                | Kind::Equatable
                | Kind::Nullable
                | Kind::Negatable => Ok(Substitution::empty()),
                _ => Err(Error::CannotConstrain(self, with)),
            },
            MonoType::Float => match with {
                Kind::Addable
                | Kind::Subtractable
                | Kind::Divisible
                | Kind::Numeric
                | Kind::Comparable
                | Kind::Equatable
                | Kind::Nullable
                | Kind::Negatable => Ok(Substitution::empty()),
                _ => Err(Error::CannotConstrain(self, with)),
            },
            MonoType::String => match with {
                Kind::Addable | Kind::Comparable | Kind::Equatable | Kind::Nullable => {
                    Ok(Substitution::empty())
                }
                _ => Err(Error::CannotConstrain(self, with)),
            },
            MonoType::Duration => match with {
                Kind::Comparable | Kind::Equatable | Kind::Nullable | Kind::Negatable => {
                    Ok(Substitution::empty())
                }
                _ => Err(Error::CannotConstrain(self, with)),
            },
            MonoType::Time => match with {
                Kind::Comparable | Kind::Equatable | Kind::Nullable => Ok(Substitution::empty()),
                _ => Err(Error::CannotConstrain(self, with)),
            },
            MonoType::Regexp => Err(Error::CannotConstrain(self, with)),
            MonoType::Bytes => match with {
                Kind::Equatable => Ok(Substitution::empty()),
                _ => Err(Error::CannotConstrain(self, with)),
            },
            MonoType::Var(tvr) => {
                tvr.constrain(with, cons);
                Ok(Substitution::empty())
            }
            MonoType::Arr(arr) => arr.constrain(with, cons),
            MonoType::Row(obj) => obj.constrain(with, cons),
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
            MonoType::Row(row) => row.contains(tv),
            MonoType::Fun(fun) => fun.contains(tv),
        }
    }
}

// Tvar stands for type variable.
// A type variable holds an unknown type.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, PartialOrd, Ord)]
pub struct Tvar(pub u64);

// TvarKinds is a map from type variables to their constraining kinds.
pub type TvarKinds = SemanticMap<Tvar, Vec<Kind>>;
pub type TvarMap = SemanticMap<Tvar, Tvar>;
pub type SubstitutionMap = SemanticMap<Tvar, MonoType>;

impl fmt::Display for Tvar {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "t{}", self.0)
    }
}

impl MaxTvar for Tvar {
    fn max_tvar(&self) -> Tvar {
        *self
    }
}

impl Tvar {
    fn unify(self, with: MonoType, cons: &mut TvarKinds) -> Result<Substitution, Error> {
        match with {
            MonoType::Var(tv) => {
                if self == tv {
                    // The empty substitution will always
                    // unify a type variable with itself.
                    Ok(Substitution::empty())
                } else {
                    // Unify two distinct type variables.
                    // This will update the kind constraints
                    // associated with these type variables.
                    self.unify_with_tvar(tv, cons)
                }
            }
            _ => {
                if with.contains(self) {
                    // Invalid recursive type
                    Err(Error::OccursCheck(self, with))
                } else {
                    // Unify a type variable with a monotype.
                    // The monotype must satisify any
                    // constraints placed on the type variable.
                    self.unify_with_type(with, cons)
                }
            }
        }
    }

    fn unify_with_tvar(self, tv: Tvar, cons: &mut TvarKinds) -> Result<Substitution, Error> {
        // Kind constraints for both type variables
        let kinds = union(
            cons.remove(&self).unwrap_or_default(),
            cons.remove(&tv).unwrap_or_default(),
        );
        if !kinds.is_empty() {
            cons.insert(tv, kinds);
        }
        Ok(Substitution::from(
            semantic_map! {self => MonoType::Var(tv)},
        ))
    }

    fn unify_with_type(self, t: MonoType, cons: &mut TvarKinds) -> Result<Substitution, Error> {
        let sub = Substitution::from(semantic_map! {self => t.clone()});
        match cons.remove(&self) {
            None => Ok(sub),
            Some(kinds) => Ok(sub.merge(kinds.into_iter().try_fold(
                Substitution::empty(),
                |sub, kind| {
                    // The monotype that is being unified with the
                    // tvar must be constrained with the same kinds
                    // as that of the tvar.
                    Ok(sub.merge(t.clone().constrain(kind, cons)?))
                },
            )?)),
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

// Array is a homogeneous list type
#[derive(Debug, Clone, PartialEq)]
pub struct Array(pub MonoType);

impl fmt::Display for Array {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "[{}]", self.0)
    }
}

impl Substitutable for Array {
    fn apply(self, sub: &Substitution) -> Self {
        Array(self.0.apply(sub))
    }
    fn free_vars(&self) -> Vec<Tvar> {
        self.0.free_vars()
    }
}

impl MaxTvar for Array {
    fn max_tvar(&self) -> Tvar {
        self.0.max_tvar()
    }
}

impl Array {
    // self represents the expected type.
    fn unify(
        self,
        with: Self,
        cons: &mut TvarKinds,
        f: &mut Fresher,
    ) -> Result<Substitution, Error> {
        self.0.unify(with.0, cons, f)
    }

    fn constrain(self, with: Kind, cons: &mut TvarKinds) -> Result<Substitution, Error> {
        match with {
            Kind::Equatable => self.0.constrain(with, cons),
            _ => Err(Error::CannotConstrain(MonoType::Arr(Box::new(self)), with)),
        }
    }

    fn contains(&self, tv: Tvar) -> bool {
        self.0.contains(tv)
    }
}

// Row is an extensible record type.
//
// A row is either Empty meaning it has no properties,
// or it is an extension of a row.
//
// A row may extend what is referred to as a row
// variable. A row variable is a type variable that
// represents an unknown record type.
//
#[derive(Debug, Clone)]
pub enum Row {
    Empty,
    Extension { head: Property, tail: MonoType },
}

impl fmt::Display for Row {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        f.write_str("{")?;
        self.format(f)?;
        f.write_str("}")
    }
}

impl cmp::PartialEq for Row {
    fn eq(mut self: &Self, mut other: &Self) -> bool {
        let mut a = RefMonoTypeVecMap::new();
        let t = loop {
            match self {
                Row::Empty => break None,
                Row::Extension {
                    head,
                    tail: MonoType::Row(o),
                } => {
                    a.entry(&head.k).or_insert_with(Vec::new).push(&head.v);
                    self = o;
                }
                Row::Extension {
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
                Row::Empty => break None,
                Row::Extension {
                    head,
                    tail: MonoType::Row(o),
                } => {
                    b.entry(&head.k).or_insert_with(Vec::new).push(&head.v);
                    other = o;
                }
                Row::Extension {
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

impl Substitutable for Row {
    fn apply(self, sub: &Substitution) -> Self {
        match self {
            Row::Empty => Row::Empty,
            Row::Extension { head, tail } => Row::Extension {
                head: head.apply(sub),
                tail: tail.apply(sub),
            },
        }
    }
    fn free_vars(&self) -> Vec<Tvar> {
        match self {
            Row::Empty => Vec::new(),
            Row::Extension { head, tail } => union(tail.free_vars(), head.v.free_vars()),
        }
    }
}

impl MaxTvar for Row {
    fn max_tvar(&self) -> Tvar {
        match self {
            Row::Empty => Tvar(0),
            Row::Extension { head, tail } => vec![head.max_tvar(), tail.max_tvar()].max_tvar(),
        }
    }
}

#[allow(clippy::many_single_char_names)]
impl Row {
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
        f: &mut Fresher,
    ) -> Result<Substitution, Error> {
        match (self.clone(), actual.clone()) {
            (Row::Empty, Row::Empty) => Ok(Substitution::empty()),
            (
                Row::Extension {
                    head: Property { k: a, v: t },
                    tail: MonoType::Var(l),
                },
                Row::Extension {
                    head: Property { k: b, v: u },
                    tail: MonoType::Var(r),
                },
            ) if a == b && l == r => match t.clone().unify(u.clone(), cons, f) {
                Err(_) => Err(Error::CannotUnifyLabel(a, t, u)),
                Ok(sub) => Ok(sub),
            },
            (
                Row::Extension {
                    head: Property { k: a, .. },
                    tail: MonoType::Var(l),
                },
                Row::Extension {
                    head: Property { k: b, .. },
                    tail: MonoType::Var(r),
                },
            ) if a != b && l == r => Err(Error::CannotUnify(
                MonoType::Row(Box::new(self)),
                MonoType::Row(Box::new(actual)),
            )),
            (
                Row::Extension {
                    head: Property { k: a, v: t },
                    tail: l,
                },
                Row::Extension {
                    head: Property { k: b, v: u },
                    tail: r,
                },
            ) if a == b => {
                let sub = t.unify(u, cons, f)?;
                apply_then_unify(l, r, sub, cons, f)
            }
            (
                Row::Extension {
                    head: Property { k: a, v: t },
                    tail: l,
                },
                Row::Extension {
                    head: Property { k: b, v: u },
                    tail: r,
                },
            ) if a != b => {
                let var = f.fresh();
                let exp = MonoType::from(Row::Extension {
                    head: Property { k: a, v: t },
                    tail: MonoType::Var(var),
                });
                let act = MonoType::from(Row::Extension {
                    head: Property { k: b, v: u },
                    tail: MonoType::Var(var),
                });
                let sub = l.unify(act, cons, f)?;
                apply_then_unify(exp, r, sub, cons, f)
            }
            // If we are expecting {a: u | r} but find {}, label `a` is missing.
            (
                Row::Extension {
                    head: Property { k: a, .. },
                    ..
                },
                Row::Empty,
            ) => Err(Error::MissingLabel(a)),
            // If we are expecting {} but find {a: u | r}, label `a` is extra.
            (
                Row::Empty,
                Row::Extension {
                    head: Property { k: a, .. },
                    ..
                },
            ) => Err(Error::ExtraLabel(a)),
            _ => Err(Error::CannotUnify(
                MonoType::Row(Box::new(self)),
                MonoType::Row(Box::new(actual)),
            )),
        }
    }

    fn constrain(self, with: Kind, cons: &mut TvarKinds) -> Result<Substitution, Error> {
        match with {
            Kind::Row => Ok(Substitution::empty()),
            Kind::Equatable => match self {
                Row::Empty => Ok(Substitution::empty()),
                Row::Extension { head, tail } => {
                    let sub = head.v.constrain(with, cons)?;
                    Ok(sub.merge(tail.constrain(with, cons)?))
                }
            },
            _ => Err(Error::CannotConstrain(MonoType::Row(Box::new(self)), with)),
        }
    }

    fn contains(&self, tv: Tvar) -> bool {
        match self {
            Row::Empty => false,
            Row::Extension { head, tail } => head.v.contains(tv) && tail.contains(tv),
        }
    }

    fn format(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match self {
            Row::Empty => f.write_str("{}"),
            Row::Extension { head, tail } => match tail {
                MonoType::Var(_) => write!(f, "{} | {}", head, tail),
                MonoType::Row(obj) => {
                    write!(f, "{} | ", head)?;
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
fn apply_then_unify(
    exp: MonoType,
    act: MonoType,
    sub: Substitution,
    cons: &mut TvarKinds,
    f: &mut Fresher,
) -> Result<Substitution, Error> {
    let s = exp.apply(&sub).unify(act.apply(&sub), cons, f)?;
    Ok(sub.merge(s))
}

// A key value pair representing a property type in a record
#[derive(Debug, Clone, PartialEq)]
pub struct Property {
    pub k: String,
    pub v: MonoType,
}

impl fmt::Display for Property {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "{}:{}", self.k, self.v)
    }
}

impl Substitutable for Property {
    fn apply(self, sub: &Substitution) -> Self {
        Property {
            k: self.k,
            v: self.v.apply(sub),
        }
    }
    fn free_vars(&self) -> Vec<Tvar> {
        self.v.free_vars()
    }
}

impl MaxTvar for Property {
    fn max_tvar(&self) -> Tvar {
        self.v.max_tvar()
    }
}

// Function represents a function type.
//
// A function type is defined by as set of required arguments,
// a set of optional arguments, an optional pipe argument, and
// a required return type.
//
#[derive(Debug, Clone, PartialEq)]
pub struct Function {
    pub req: MonoTypeMap,
    pub opt: MonoTypeMap,
    pub pipe: Option<Property>,
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
                k: String::from("?") + &k,
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
            "({}) -> {}",
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
impl<T: Substitutable> Substitutable for SemanticMap<String, T> {
    fn apply(self, sub: &Substitution) -> Self {
        self.into_iter().map(|(k, v)| (k, v.apply(sub))).collect()
    }
    fn free_vars(&self) -> Vec<Tvar> {
        self.values()
            .fold(Vec::new(), |vars, t| union(vars, t.free_vars()))
    }
}

impl<T: Substitutable> Substitutable for Option<T> {
    fn apply(self, sub: &Substitution) -> Self {
        match self {
            Some(t) => Some(t.apply(sub)),
            None => None,
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
    fn apply(self, sub: &Substitution) -> Self {
        Function {
            req: self.req.apply(sub),
            opt: self.opt.apply(sub),
            pipe: self.pipe.apply(sub),
            retn: self.retn.apply(sub),
        }
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
    fn max_tvar(&self) -> Tvar {
        self.iter()
            .map(|(_, t)| t.max_tvar())
            .fold(Tvar(0), |max, tv| if tv > max { tv } else { max })
    }
}

impl<T: MaxTvar> MaxTvar for Option<T> {
    fn max_tvar(&self) -> Tvar {
        match self {
            None => Tvar(0),
            Some(t) => t.max_tvar(),
        }
    }
}

impl MaxTvar for Function {
    fn max_tvar(&self) -> Tvar {
        vec![
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
    /// 1. f = (a=<-, b) -> {...}
    /// 2. 0 |> f(b: 1)
    /// 3. f(a: 0, b: 1)
    /// 4. f = (d=<-, b, c=0) -> {...}
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
        fresh: &mut Fresher,
    ) -> Result<Substitution, Error> {
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
                    return Err(Error::MultiplePipeArguments(fp.k, gp.k));
                } else {
                    // At least one is unnamed or they are both named with the same name.
                    // This means they should match. Enforce this condition by inserting
                    // the pipe argument into the required ones with the same key.
                    f.req.insert(fp.k.clone(), fp.v);
                    g.req.insert(fp.k.clone(), gp.v);
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
        let mut sub = Substitution::empty();
        // Unify f's required arguments.
        for (name, exp) in f.req.into_iter() {
            if let Some(act) = g.req.remove(&name) {
                // The required argument is in g's required arguments.
                sub = match apply_then_unify(exp.clone(), act.clone(), sub, cons, fresh) {
                    Err(e) => Err(Error::CannotUnifyArgument(name, Box::new(e))),
                    Ok(sub) => Ok(sub),
                }?;
            } else if let Some(act) = g.opt.remove(&name) {
                // The required argument is in g's optional arguments.
                sub = match apply_then_unify(exp.clone(), act.clone(), sub, cons, fresh) {
                    Err(e) => Err(Error::CannotUnifyArgument(name, Box::new(e))),
                    Ok(sub) => Ok(sub),
                }?;
            } else {
                return Err(Error::MissingArgument(name));
            }
        }
        // Unify f's optional arguments.
        for (name, exp) in f.opt.into_iter() {
            if let Some(act) = g.req.remove(&name) {
                sub = match apply_then_unify(exp.clone(), act.clone(), sub, cons, fresh) {
                    Err(e) => Err(Error::CannotUnifyArgument(name, Box::new(e))),
                    Ok(sub) => Ok(sub),
                }?;
            } else if let Some(act) = g.opt.remove(&name) {
                sub = match apply_then_unify(exp.clone(), act.clone(), sub, cons, fresh) {
                    Err(e) => Err(Error::CannotUnifyArgument(name, Box::new(e))),
                    Ok(sub) => Ok(sub),
                }?;
            }
        }
        // Unify return types.
        match apply_then_unify(f.retn.clone(), g.retn.clone(), sub, cons, fresh) {
            Err(_) => Err(Error::CannotUnifyReturn(f.retn, g.retn)),
            Ok(sub) => Ok(sub),
        }
    }

    fn constrain(self, with: Kind, _: &mut TvarKinds) -> Result<Substitution, Error> {
        Err(Error::CannotConstrain(MonoType::Fun(Box::new(self)), with))
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

pub trait MaxTvar {
    fn max_tvar(&self) -> Tvar;
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::semantic::parser::parse;

    /// Polytype is an util method that returns a PolyType from a string.
    pub fn polytype(typ: &str) -> PolyType {
        parse(typ).unwrap()
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
        assert!(Kind::Row.to_string() == "Row");
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
        assert_eq!(
            "[int]",
            MonoType::Arr(Box::new(Array(MonoType::Int))).to_string()
        );
    }
    #[test]
    fn display_type_row() {
        assert_eq!(
            "{a:int | b:string | t0}",
            Row::Extension {
                head: Property {
                    k: String::from("a"),
                    v: MonoType::Int,
                },
                tail: MonoType::Row(Box::new(Row::Extension {
                    head: Property {
                        k: String::from("b"),
                        v: MonoType::String,
                    },
                    tail: MonoType::Var(Tvar(0)),
                })),
            }
            .to_string()
        );
        assert_eq!(
            "{a:int | b:string | {}}",
            Row::Extension {
                head: Property {
                    k: String::from("a"),
                    v: MonoType::Int,
                },
                tail: MonoType::Row(Box::new(Row::Extension {
                    head: Property {
                        k: String::from("b"),
                        v: MonoType::String,
                    },
                    tail: MonoType::Row(Box::new(Row::Empty)),
                })),
            }
            .to_string()
        );
    }
    #[test]
    fn display_type_function() {
        assert_eq!(
            "() -> int",
            Function {
                req: MonoTypeMap::new(),
                opt: MonoTypeMap::new(),
                pipe: None,
                retn: MonoType::Int,
            }
            .to_string()
        );
        assert_eq!(
            "(<-:int) -> int",
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
            "(<-a:int) -> int",
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
            "(<-:int, a:int, b:int) -> int",
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
            "(<-:int, ?a:int, ?b:int) -> int",
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
            "(<-:int, a:int, b:int, ?c:int, ?d:int) -> int",
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
            "(a:int, ?b:bool) -> int",
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
            "(<-a:int, b:int, c:int, ?d:bool) -> int",
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
            "forall [] int",
            PolyType {
                vars: Vec::new(),
                cons: TvarKinds::new(),
                expr: MonoType::Int,
            }
            .to_string(),
        );
        assert_eq!(
            "forall [t0] (x:t0) -> t0",
            PolyType {
                vars: vec![Tvar(0)],
                cons: TvarKinds::new(),
                expr: MonoType::Fun(Box::new(Function {
                    req: semantic_map! {
                        String::from("x") => MonoType::Var(Tvar(0)),
                    },
                    opt: MonoTypeMap::new(),
                    pipe: None,
                    retn: MonoType::Var(Tvar(0)),
                })),
            }
            .to_string(),
        );
        assert_eq!(
            "forall [t0, t1] (x:t0, y:t1) -> {x:t0 | y:t1 | {}}",
            PolyType {
                vars: vec![Tvar(0), Tvar(1)],
                cons: TvarKinds::new(),
                expr: MonoType::Fun(Box::new(Function {
                    req: semantic_map! {
                        String::from("x") => MonoType::Var(Tvar(0)),
                        String::from("y") => MonoType::Var(Tvar(1)),
                    },
                    opt: MonoTypeMap::new(),
                    pipe: None,
                    retn: MonoType::Row(Box::new(Row::Extension {
                        head: Property {
                            k: String::from("x"),
                            v: MonoType::Var(Tvar(0)),
                        },
                        tail: MonoType::Row(Box::new(Row::Extension {
                            head: Property {
                                k: String::from("y"),
                                v: MonoType::Var(Tvar(1)),
                            },
                            tail: MonoType::Row(Box::new(Row::Empty)),
                        })),
                    })),
                })),
            }
            .to_string(),
        );
        assert_eq!(
            "forall [t0] where t0:Addable (a:t0, b:t0) -> t0",
            PolyType {
                vars: vec![Tvar(0)],
                cons: semantic_map! {Tvar(0) => vec![Kind::Addable]},
                expr: MonoType::Fun(Box::new(Function {
                    req: semantic_map! {
                        String::from("a") => MonoType::Var(Tvar(0)),
                        String::from("b") => MonoType::Var(Tvar(0)),
                    },
                    opt: MonoTypeMap::new(),
                    pipe: None,
                    retn: MonoType::Var(Tvar(0)),
                })),
            }
            .to_string(),
        );
        assert_eq!(
            "forall [t0, t1] where t0:Addable, t1:Divisible (x:t0, y:t1) -> {x:t0 | y:t1 | {}}",
            PolyType {
                vars: vec![Tvar(0), Tvar(1)],
                cons: semantic_map! {
                    Tvar(0) => vec![Kind::Addable],
                    Tvar(1) => vec![Kind::Divisible],
                },
                expr: MonoType::Fun(Box::new(Function {
                    req: semantic_map! {
                        String::from("x") => MonoType::Var(Tvar(0)),
                        String::from("y") => MonoType::Var(Tvar(1)),
                    },
                    opt: MonoTypeMap::new(),
                    pipe: None,
                    retn: MonoType::Row(Box::new(Row::Extension {
                        head: Property {
                            k: String::from("x"),
                            v: MonoType::Var(Tvar(0)),
                        },
                        tail: MonoType::Row(Box::new(Row::Extension {
                            head: Property {
                                k: String::from("y"),
                                v: MonoType::Var(Tvar(1)),
                            },
                            tail: MonoType::Row(Box::new(Row::Empty)),
                        })),
                    })),
                })),
            }
            .to_string(),
        );
        assert_eq!(
            "forall [t0, t1] where t0:Comparable + Equatable, t1:Addable + Divisible (x:t0, y:t1) -> {x:t0 | y:t1 | {}}",
            PolyType {
                vars: vec![Tvar(0), Tvar(1)],
                cons: semantic_map! {
                    Tvar(0) => vec![Kind::Comparable, Kind::Equatable],
                    Tvar(1) => vec![Kind::Addable, Kind::Divisible],
                },
                expr: MonoType::Fun(Box::new(Function {
                    req: semantic_map! {
                        String::from("x") => MonoType::Var(Tvar(0)),
                        String::from("y") => MonoType::Var(Tvar(1)),
                    },
                    opt: MonoTypeMap::new(),
                    pipe: None,
                    retn: MonoType::Row(Box::new(Row::Extension {
                        head: Property {
                            k: String::from("x"),
                            v: MonoType::Var(Tvar(0)),
                        },
                        tail: MonoType::Row(Box::new(Row::Extension {
                            head: Property {
                                k: String::from("y"),
                                v: MonoType::Var(Tvar(1)),
                            },
                            tail: MonoType::Row(Box::new(Row::Empty)),
                        })),
                    })),
                })),
            }
            .to_string(),
        );
    }

    #[test]
    fn compare_records() {
        assert_eq!(
            // {a:int | b:string | t0}
            MonoType::Row(Box::new(Row::Extension {
                head: Property {
                    k: String::from("a"),
                    v: MonoType::Int,
                },
                tail: MonoType::Row(Box::new(Row::Extension {
                    head: Property {
                        k: String::from("b"),
                        v: MonoType::String,
                    },
                    tail: MonoType::Var(Tvar(0)),
                })),
            })),
            // {b:string | a:int | t0}
            MonoType::Row(Box::new(Row::Extension {
                head: Property {
                    k: String::from("b"),
                    v: MonoType::String,
                },
                tail: MonoType::Row(Box::new(Row::Extension {
                    head: Property {
                        k: String::from("a"),
                        v: MonoType::Int,
                    },
                    tail: MonoType::Var(Tvar(0)),
                })),
            })),
        );
        assert_eq!(
            // {a:int | b:string | b:int | c:float | t0}
            MonoType::Row(Box::new(Row::Extension {
                head: Property {
                    k: String::from("a"),
                    v: MonoType::Int,
                },
                tail: MonoType::Row(Box::new(Row::Extension {
                    head: Property {
                        k: String::from("b"),
                        v: MonoType::String,
                    },
                    tail: MonoType::Row(Box::new(Row::Extension {
                        head: Property {
                            k: String::from("b"),
                            v: MonoType::Int,
                        },
                        tail: MonoType::Row(Box::new(Row::Extension {
                            head: Property {
                                k: String::from("c"),
                                v: MonoType::Float,
                            },
                            tail: MonoType::Var(Tvar(0)),
                        })),
                    })),
                })),
            })),
            // {c:float | b:string | b:int | a:int | t0}
            MonoType::Row(Box::new(Row::Extension {
                head: Property {
                    k: String::from("c"),
                    v: MonoType::Float,
                },
                tail: MonoType::Row(Box::new(Row::Extension {
                    head: Property {
                        k: String::from("b"),
                        v: MonoType::String,
                    },
                    tail: MonoType::Row(Box::new(Row::Extension {
                        head: Property {
                            k: String::from("b"),
                            v: MonoType::Int,
                        },
                        tail: MonoType::Row(Box::new(Row::Extension {
                            head: Property {
                                k: String::from("a"),
                                v: MonoType::Int,
                            },
                            tail: MonoType::Var(Tvar(0)),
                        })),
                    })),
                })),
            })),
        );
        assert_ne!(
            // {a:int | b:string | b:int | c:float | t0}
            MonoType::Row(Box::new(Row::Extension {
                head: Property {
                    k: String::from("a"),
                    v: MonoType::Int,
                },
                tail: MonoType::Row(Box::new(Row::Extension {
                    head: Property {
                        k: String::from("b"),
                        v: MonoType::String,
                    },
                    tail: MonoType::Row(Box::new(Row::Extension {
                        head: Property {
                            k: String::from("b"),
                            v: MonoType::Int,
                        },
                        tail: MonoType::Row(Box::new(Row::Extension {
                            head: Property {
                                k: String::from("c"),
                                v: MonoType::Float,
                            },
                            tail: MonoType::Var(Tvar(0)),
                        })),
                    })),
                })),
            })),
            // {a:int | b:int | b:string | c:float | t0}
            MonoType::Row(Box::new(Row::Extension {
                head: Property {
                    k: String::from("a"),
                    v: MonoType::Int,
                },
                tail: MonoType::Row(Box::new(Row::Extension {
                    head: Property {
                        k: String::from("b"),
                        v: MonoType::Int,
                    },
                    tail: MonoType::Row(Box::new(Row::Extension {
                        head: Property {
                            k: String::from("b"),
                            v: MonoType::String,
                        },
                        tail: MonoType::Row(Box::new(Row::Extension {
                            head: Property {
                                k: String::from("c"),
                                v: MonoType::Float,
                            },
                            tail: MonoType::Var(Tvar(0)),
                        })),
                    })),
                })),
            })),
        );
        assert_ne!(
            // {a:int | b:string | {}}
            MonoType::Row(Box::new(Row::Extension {
                head: Property {
                    k: String::from("a"),
                    v: MonoType::Int,
                },
                tail: MonoType::Row(Box::new(Row::Extension {
                    head: Property {
                        k: String::from("b"),
                        v: MonoType::String,
                    },
                    tail: MonoType::Row(Box::new(Row::Empty)),
                })),
            })),
            // {b:int | a:int | {}}
            MonoType::Row(Box::new(Row::Extension {
                head: Property {
                    k: String::from("b"),
                    v: MonoType::Int,
                },
                tail: MonoType::Row(Box::new(Row::Extension {
                    head: Property {
                        k: String::from("a"),
                        v: MonoType::Int,
                    },
                    tail: MonoType::Row(Box::new(Row::Empty)),
                })),
            })),
        );
        assert_ne!(
            // {a:int | {}}
            MonoType::Row(Box::new(Row::Extension {
                head: Property {
                    k: String::from("a"),
                    v: MonoType::Int,
                },
                tail: MonoType::Row(Box::new(Row::Empty)),
            })),
            // {a:int | t0}
            MonoType::Row(Box::new(Row::Extension {
                head: Property {
                    k: String::from("a"),
                    v: MonoType::Int,
                },
                tail: MonoType::Var(Tvar(0)),
            })),
        );
        assert_ne!(
            // {a:int | t0}
            MonoType::Row(Box::new(Row::Extension {
                head: Property {
                    k: String::from("a"),
                    v: MonoType::Int,
                },
                tail: MonoType::Var(Tvar(0)),
            })),
            // {a:int | t1}
            MonoType::Row(Box::new(Row::Extension {
                head: Property {
                    k: String::from("a"),
                    v: MonoType::Int,
                },
                tail: MonoType::Var(Tvar(1)),
            })),
        );
    }

    #[test]
    fn unify_ints() {
        let sub = MonoType::Int
            .unify(
                MonoType::Int,
                &mut TvarKinds::new(),
                &mut Fresher::default(),
            )
            .unwrap();
        assert_eq!(sub, Substitution::empty());
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
        ];
        for c in allowable_cons {
            let sub = MonoType::Int.constrain(c, &mut TvarKinds::new());
            assert_eq!(Ok(Substitution::empty()), sub);
        }

        let sub = MonoType::Int.constrain(Kind::Row, &mut TvarKinds::new());
        assert_eq!(Err(Error::CannotConstrain(MonoType::Int, Kind::Row)), sub);
    }
    #[test]
    fn constrain_rows() {
        let sub = Row::Empty.constrain(Kind::Row, &mut TvarKinds::new());
        assert_eq!(Ok(Substitution::empty()), sub);

        let unallowable_cons = vec![
            Kind::Addable,
            Kind::Subtractable,
            Kind::Divisible,
            Kind::Numeric,
            Kind::Comparable,
            Kind::Nullable,
        ];
        for c in unallowable_cons {
            let sub = Row::Empty.constrain(c, &mut TvarKinds::new());
            assert_eq!(
                Err(Error::CannotConstrain(
                    MonoType::Row(Box::new(Row::Empty)),
                    c
                )),
                sub
            );
        }
    }
    #[test]
    fn unify_error() {
        let err = MonoType::Int
            .unify(
                MonoType::String,
                &mut TvarKinds::new(),
                &mut Fresher::default(),
            )
            .unwrap_err();
        assert_eq!(err.to_string(), String::from("int != string"),);
    }
    #[test]
    fn unify_tvars() {
        let sub = MonoType::Var(Tvar(0))
            .unify(
                MonoType::Var(Tvar(1)),
                &mut TvarKinds::new(),
                &mut Fresher::default(),
            )
            .unwrap();
        assert_eq!(
            sub,
            Substitution::from(semantic_map! {Tvar(0) => MonoType::Var(Tvar(1))}),
        );
    }
    #[test]
    fn unify_constrained_tvars() {
        let mut cons = semantic_map! {Tvar(0) => vec![Kind::Addable, Kind::Divisible]};
        let sub = MonoType::Var(Tvar(0))
            .unify(MonoType::Var(Tvar(1)), &mut cons, &mut Fresher::default())
            .unwrap();
        assert_eq!(
            sub,
            Substitution::from(semantic_map! {Tvar(0) => MonoType::Var(Tvar(1))})
        );
        assert_eq!(
            cons,
            semantic_map! {Tvar(1) => vec![Kind::Addable, Kind::Divisible]},
        );
    }
    #[test]
    fn cannot_unify_functions() {
        // g-required and g-optional arguments do not contain a f-required argument (and viceversa).
        let f = polytype(
            "forall [t0, t1] where t0: Addable, t1: Divisible (a: t0, b: t0, ?c: t1) -> t0",
        );
        let g = polytype("forall [t2] where t2: Addable (d: t2, ?e: t2) -> t2");
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
            let res = f
                .clone()
                .unify(*g.clone(), &mut cons, &mut Fresher::default());
            assert!(res.is_err());
            let res = g
                .clone()
                .unify(*f.clone(), &mut cons, &mut Fresher::default());
            assert!(res.is_err());
        } else {
            panic!("the monotypes under examination are not functions");
        }
        // f has a pipe argument, but g does not (and viceversa).
        let f =
            polytype("forall [t0, t1] where t0: Addable, t1: Divisible (<-pip:t0, a: t1) -> t0");
        let g = polytype("forall [t2] where t2: Addable (a: t2) -> t2");
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
            let res = f
                .clone()
                .unify(*g.clone(), &mut cons, &mut Fresher::default());
            assert!(res.is_err());
            let res = g
                .clone()
                .unify(*f.clone(), &mut cons, &mut Fresher::default());
            assert!(res.is_err());
        } else {
            panic!("the monotypes under examination are not functions");
        }
    }
    #[test]
    fn unify_function_with_function_call() {
        let fn_type = polytype(
            "forall [t0, t1] where t0: Addable, t1: Divisible (a: t0, b: t0, ?c: t1) -> t0",
        );
        // (a: int, b: int) -> int
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
            let sub = f
                .unify(call_type, &mut cons, &mut Fresher::default())
                .unwrap();
            assert_eq!(
                sub,
                Substitution::from(semantic_map! {Tvar(0) => MonoType::Int})
            );
            // the constraint on t0 gets removed.
            assert_eq!(cons, semantic_map! {Tvar(1) => vec![Kind::Divisible]});
        } else {
            panic!("the monotype under examination is not a function");
        }
    }
    #[test]
    fn unify_functions() {
        let f = polytype(
            "forall [t0, t1] where t0: Addable, t1: Divisible (a: t0, b: t0, ?c: t1) -> t0",
        );
        let g = polytype("forall [t2] where t2: Addable (a: t2, ?b: t2, c: float) -> t2");
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
            let sub = f.unify(*g, &mut cons, &mut Fresher::default()).unwrap();
            assert_eq!(
                sub,
                Substitution::from(semantic_map! {
                    Tvar(0) => MonoType::Var(Tvar(2)),
                    Tvar(1) => MonoType::Float,
                })
            );
            // t0 is equal to t2 and t2 is Addable, so we only need one constraint on t2;
            // t1 ended up being a float, so we do not need any kind constraint on it.
            assert_eq!(cons, semantic_map! {Tvar(2) => vec![Kind::Addable]});
        } else {
            panic!("the monotypes under examination are not functions");
        }
    }
    #[test]
    fn unify_higher_order_functions() {
        let f = polytype(
            "forall [t0, t1] where t0: Addable, t1: Divisible (a: t0, b: t0, ?c: (a: t0) -> t1) -> (d:  string) -> t0",
        );
        let g = polytype("forall [] (a: int, b: int, c: (a: int) -> float) -> (d: string) -> int");
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
            let sub = f.unify(*g, &mut cons, &mut Fresher::default()).unwrap();
            assert_eq!(
                sub,
                Substitution::from(semantic_map! {
                    Tvar(0) => MonoType::Int,
                    Tvar(1) => MonoType::Float,
                })
            );
            // we know everything about tvars, there is no constraint.
            assert_eq!(cons, semantic_map! {});
        } else {
            panic!("the monotypes under examination are not functions");
        }
    }
}

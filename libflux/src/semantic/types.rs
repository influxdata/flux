use crate::semantic::{
    fresh::Fresher,
    sub::{Substitutable, Substitution},
};
use std::{
    cmp,
    collections::{BTreeMap, BTreeSet, HashMap},
    fmt,
};

#[derive(Debug, Clone, PartialEq)]
pub struct PolyType {
    pub vars: Vec<Tvar>,
    pub cons: HashMap<Tvar, Vec<Kind>>,
    pub expr: MonoType,
}

impl fmt::Display for PolyType {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        let vars = &self
            .vars
            .iter()
            .map(|x| x.to_string())
            .collect::<Vec<_>>()
            .join(", ");
        if self.cons.len() == 0 {
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

impl PolyType {
    // Normalize a polytype's free variables by replacing them with
    // new fresh variables starting from t0.
    pub fn normalize(self, f: &mut Fresher) -> Self {
        let mut sub = HashMap::new();
        for tv in &self.vars {
            sub.insert(*tv, f.fresh());
        }

        let mut vars = Vec::new();
        for tv in &self.vars {
            vars.push(*sub.get(tv).unwrap());
        }

        let mut cons = HashMap::new();
        for (tv, kinds) in &self.cons {
            cons.insert(*sub.get(tv).unwrap(), kinds.to_owned());
        }

        let sub: Substitution = sub
            .into_iter()
            .map(|(a, b)| (a, MonoType::Var(b)))
            .collect::<HashMap<Tvar, MonoType>>()
            .into();

        PolyType {
            vars: vars,
            cons: cons,
            expr: self.expr.apply(&sub),
        }
    }
    fn display_constraints(cons: &HashMap<Tvar, Vec<Kind>>) -> String {
        cons.iter()
            // A BTree produces a sorted iterator for
            // deterministic display output
            .collect::<BTreeMap<_, _>>()
            .iter()
            .map(|(&&tv, &kinds)| format!("{}:{}", tv, PolyType::display_kinds(kinds)))
            .collect::<Vec<_>>()
            .join(", ")
    }
    fn display_kinds(kinds: &Vec<Kind>) -> String {
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

#[derive(Debug)]
pub struct Error {
    msg: String,
}

impl fmt::Display for Error {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        f.write_str(&self.msg)
    }
}

impl Error {
    // An error can occur when the unification of two types
    // contradicts what we have already inferred about the types
    // in our program.
    fn cannot_unify(t: MonoType, with: MonoType) -> Error {
        Error {
            msg: format!("cannot unify {} with {}", t, with),
        }
    }
    // An error can occur if we constrain a type with a kind to
    // which it does not belong.
    fn cannot_constrain(t: MonoType, with: Kind) -> Error {
        Error {
            msg: format!("{} is not of kind {}", t, with,),
        }
    }
    // An error can occur if we attempt to unify a type variable
    // with a monotype that contains that same type variable.
    fn occurs_check(tv: Tvar, t: MonoType) -> Error {
        Error {
            msg: format!("type variable {} occurs in {}", tv, t),
        }
    }
}

// Kind represents a class or family of types
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub enum Kind {
    Addable,
    Subtractable,
    Divisible,
    Comparable,
    Equatable,
    Nullable,
}

impl fmt::Display for Kind {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match self {
            Kind::Addable => f.write_str("Addable"),
            Kind::Subtractable => f.write_str("Subtractable"),
            Kind::Divisible => f.write_str("Divisible"),
            Kind::Comparable => f.write_str("Comparable"),
            Kind::Equatable => f.write_str("Equatable"),
            Kind::Nullable => f.write_str("Nullable"),
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

// TvarKinds is a map from type variables to their constraining kinds.
type TvarKinds = HashMap<Tvar, Vec<Kind>>;

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
    Var(Tvar),
    Arr(Box<Array>),
    Row(Box<Row>),
    Fun(Box<Function>),
}

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
            | MonoType::Regexp => self,
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
            | MonoType::Regexp => Vec::new(),
            MonoType::Var(tvr) => vec![*tvr],
            MonoType::Arr(arr) => arr.free_vars(),
            MonoType::Row(obj) => obj.free_vars(),
            MonoType::Fun(fun) => fun.free_vars(),
        }
    }
}

impl MonoType {
    pub fn unify(self, with: Self, cons: &mut TvarKinds) -> Result<Substitution, Error> {
        match (self, with) {
            (MonoType::Bool, MonoType::Bool)
            | (MonoType::Int, MonoType::Int)
            | (MonoType::Uint, MonoType::Uint)
            | (MonoType::Float, MonoType::Float)
            | (MonoType::String, MonoType::String)
            | (MonoType::Duration, MonoType::Duration)
            | (MonoType::Time, MonoType::Time)
            | (MonoType::Regexp, MonoType::Regexp) => Ok(Substitution::empty()),
            (MonoType::Var(tv), t) => tv.unify(t, cons),
            (t, MonoType::Var(tv)) => tv.unify(t, cons),
            (MonoType::Arr(t), MonoType::Arr(s)) => t.unify(*s, cons),
            (MonoType::Row(t), MonoType::Row(s)) => t.unify(*s, cons),
            (MonoType::Fun(t), MonoType::Fun(s)) => t.unify(*s, cons),
            (t, with) => Err(Error::cannot_unify(t, with)),
        }
    }

    pub fn constrain(self, with: Kind, cons: &mut TvarKinds) -> Result<Substitution, Error> {
        match self {
            MonoType::Bool => match with {
                Kind::Equatable | Kind::Nullable => Ok(Substitution::empty()),
                _ => Err(Error::cannot_constrain(self, with)),
            },
            MonoType::Int => match with {
                Kind::Addable
                | Kind::Subtractable
                | Kind::Divisible
                | Kind::Comparable
                | Kind::Equatable
                | Kind::Nullable => Ok(Substitution::empty()),
            },
            MonoType::Uint => match with {
                Kind::Addable
                | Kind::Divisible
                | Kind::Comparable
                | Kind::Equatable
                | Kind::Nullable => Ok(Substitution::empty()),
                _ => Err(Error::cannot_constrain(self, with)),
            },
            MonoType::Float => match with {
                Kind::Addable
                | Kind::Subtractable
                | Kind::Divisible
                | Kind::Comparable
                | Kind::Equatable
                | Kind::Nullable => Ok(Substitution::empty()),
            },
            MonoType::String => match with {
                Kind::Addable | Kind::Comparable | Kind::Equatable | Kind::Nullable => {
                    Ok(Substitution::empty())
                }
                _ => Err(Error::cannot_constrain(self, with)),
            },
            MonoType::Duration => match with {
                Kind::Comparable | Kind::Equatable | Kind::Nullable => Ok(Substitution::empty()),
                _ => Err(Error::cannot_constrain(self, with)),
            },
            MonoType::Time => match with {
                Kind::Comparable | Kind::Equatable | Kind::Nullable => Ok(Substitution::empty()),
                _ => Err(Error::cannot_constrain(self, with)),
            },
            MonoType::Regexp => Err(Error::cannot_constrain(self, with)),
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
            | MonoType::Regexp => false,
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

impl fmt::Display for Tvar {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "t{}", self.0)
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
                    Err(Error::occurs_check(self, with))
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
            cons.remove(&self).unwrap_or(Vec::new()),
            cons.remove(&tv).unwrap_or(Vec::new()),
        );
        cons.insert(tv, kinds);
        Ok(Substitution::from(
            maplit::hashmap! {self => MonoType::Var(tv)},
        ))
    }

    fn unify_with_type(self, t: MonoType, cons: &mut TvarKinds) -> Result<Substitution, Error> {
        let sub = Substitution::from(maplit::hashmap! {self => t.clone()});
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

impl Array {
    fn unify(self, with: Self, cons: &mut TvarKinds) -> Result<Substitution, Error> {
        self.0.unify(with.0, cons)
    }

    fn constrain(self, with: Kind, _: &mut TvarKinds) -> Result<Substitution, Error> {
        Err(Error::cannot_constrain(MonoType::Arr(Box::new(self)), with))
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
        self.display(f)?;
        f.write_str("}")
    }
}

impl cmp::PartialEq for Row {
    fn eq(&self, other: &Self) -> bool {
        let mut l = HashMap::new();
        let mut r = HashMap::new();
        self.flatten(&mut l) == other.flatten(&mut r) && l == r
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

impl Row {
    fn unify(self, _: Self, _: &mut TvarKinds) -> Result<Substitution, Error> {
        unimplemented!();
    }

    fn constrain(self, with: Kind, _: &mut TvarKinds) -> Result<Substitution, Error> {
        Err(Error::cannot_constrain(MonoType::Row(Box::new(self)), with))
    }

    fn contains(&self, tv: Tvar) -> bool {
        match self {
            Row::Empty => false,
            Row::Extension { head, tail } => head.v.contains(tv) && tail.contains(tv),
        }
    }

    // Display a row type in flattened format
    fn display(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match self {
            Row::Empty => f.write_str("{}"),
            Row::Extension { head, tail } => {
                write!(f, "{} | ", head)?;
                match tail {
                    MonoType::Var(tvr) => write!(f, "{}", tvr),
                    MonoType::Row(obj) => obj.display(f),
                    _ => Err(fmt::Error),
                }
            }
        }
    }

    // Flatten a record type into a hashmap of property names and types
    fn flatten(&self, props: &mut HashMap<String, MonoType>) -> Option<Tvar> {
        match self {
            Row::Empty => None,
            Row::Extension { head, tail } => {
                props.insert(head.k.clone(), head.v.clone());
                match tail {
                    MonoType::Row(obj) => obj.flatten(props),
                    MonoType::Var(tvr) => Some(*tvr),
                    _ => panic!("tail of row must be either a row variable or another row"),
                }
            }
        }
    }
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

// Function represents a function type.
//
// A function type is defined by as set of required arguments,
// a set of optional arguments, an optional pipe argument, and
// a required return type.
//
#[derive(Debug, Clone, PartialEq)]
pub struct Function {
    pub req: HashMap<String, MonoType>,
    pub opt: HashMap<String, MonoType>,
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

impl<T: Substitutable> Substitutable for HashMap<String, T> {
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

impl Function {
    fn unify(self, _: Self, _: &mut TvarKinds) -> Result<Substitution, Error> {
        unimplemented!();
    }

    fn constrain(self, with: Kind, _: &mut TvarKinds) -> Result<Substitution, Error> {
        Err(Error::cannot_constrain(MonoType::Fun(Box::new(self)), with))
    }

    fn contains(&self, tv: Tvar) -> bool {
        if let Some(pipe) = &self.pipe {
            self.req.values().fold(false, |ok, t| ok || t.contains(tv))
                || self.opt.values().fold(false, |ok, t| ok || t.contains(tv))
                || pipe.v.contains(tv)
                || self.retn.contains(tv)
        } else {
            self.req.values().fold(false, |ok, t| ok || t.contains(tv))
                || self.opt.values().fold(false, |ok, t| ok || t.contains(tv))
                || self.retn.contains(tv)
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

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
                req: HashMap::new(),
                opt: HashMap::new(),
                pipe: None,
                retn: MonoType::Int,
            }
            .to_string()
        );
        assert_eq!(
            "(<-:int) -> int",
            Function {
                req: HashMap::new(),
                opt: HashMap::new(),
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
                req: HashMap::new(),
                opt: HashMap::new(),
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
                req: maplit::hashmap! {
                    String::from("a") => MonoType::Int,
                    String::from("b") => MonoType::Int,
                },
                opt: HashMap::new(),
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
                req: HashMap::new(),
                opt: maplit::hashmap! {
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
                req: maplit::hashmap! {
                    String::from("a") => MonoType::Int,
                    String::from("b") => MonoType::Int,
                },
                opt: maplit::hashmap! {
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
                req: maplit::hashmap! {
                    String::from("a") => MonoType::Int,
                },
                opt: maplit::hashmap! {
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
                req: maplit::hashmap! {
                    String::from("b") => MonoType::Int,
                    String::from("c") => MonoType::Int,
                },
                opt: maplit::hashmap! {
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
                cons: HashMap::new(),
                expr: MonoType::Int,
            }
            .to_string(),
        );
        assert_eq!(
            "forall [t0] (x:t0) -> t0",
            PolyType {
                vars: vec![Tvar(0)],
                cons: HashMap::new(),
                expr: MonoType::Fun(Box::new(Function {
                    req: maplit::hashmap! {
                        String::from("x") => MonoType::Var(Tvar(0)),
                    },
                    opt: HashMap::new(),
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
                cons: HashMap::new(),
                expr: MonoType::Fun(Box::new(Function {
                    req: maplit::hashmap! {
                        String::from("x") => MonoType::Var(Tvar(0)),
                        String::from("y") => MonoType::Var(Tvar(1)),
                    },
                    opt: HashMap::new(),
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
                cons: maplit::hashmap! {Tvar(0) => vec![Kind::Addable]},
                expr: MonoType::Fun(Box::new(Function {
                    req: maplit::hashmap! {
                        String::from("a") => MonoType::Var(Tvar(0)),
                        String::from("b") => MonoType::Var(Tvar(0)),
                    },
                    opt: HashMap::new(),
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
                cons: maplit::hashmap! {
                    Tvar(0) => vec![Kind::Addable],
                    Tvar(1) => vec![Kind::Divisible],
                },
                expr: MonoType::Fun(Box::new(Function {
                    req: maplit::hashmap! {
                        String::from("x") => MonoType::Var(Tvar(0)),
                        String::from("y") => MonoType::Var(Tvar(1)),
                    },
                    opt: HashMap::new(),
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
                cons: maplit::hashmap! {
                    Tvar(0) => vec![Kind::Comparable, Kind::Equatable],
                    Tvar(1) => vec![Kind::Addable, Kind::Divisible],
                },
                expr: MonoType::Fun(Box::new(Function {
                    req: maplit::hashmap! {
                        String::from("x") => MonoType::Var(Tvar(0)),
                        String::from("y") => MonoType::Var(Tvar(1)),
                    },
                    opt: HashMap::new(),
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
    // Ensure any two permutations of the same record are equal
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
    }

    #[test]
    fn unify_ints() {
        let sub = MonoType::Int
            .unify(MonoType::Int, &mut HashMap::new())
            .unwrap();
        assert_eq!(sub, Substitution::empty());
    }
    #[test]
    fn unify_error() {
        let err = MonoType::Int
            .unify(MonoType::String, &mut HashMap::new())
            .unwrap_err();
        assert_eq!(
            err.to_string(),
            String::from("cannot unify int with string"),
        );
    }
    #[test]
    fn unify_tvars() {
        let sub = MonoType::Var(Tvar(0))
            .unify(MonoType::Var(Tvar(1)), &mut HashMap::new())
            .unwrap();
        assert_eq!(
            sub,
            Substitution::from(maplit::hashmap! {Tvar(0) => MonoType::Var(Tvar(1))}),
        );
    }
    #[test]
    fn unify_constrained_tvars() {
        let mut cons = maplit::hashmap! {Tvar(0) => vec![Kind::Addable, Kind::Divisible]};
        let sub = MonoType::Var(Tvar(0))
            .unify(MonoType::Var(Tvar(1)), &mut cons)
            .unwrap();
        assert_eq!(
            sub,
            Substitution::from(maplit::hashmap! {Tvar(0) => MonoType::Var(Tvar(1))})
        );
        assert_eq!(
            cons,
            maplit::hashmap! {Tvar(1) => vec![Kind::Addable, Kind::Divisible]},
        );
    }
}

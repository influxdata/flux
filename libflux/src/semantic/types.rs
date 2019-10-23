use crate::semantic::sub::{Subst, Substitutable};
use std::{
    cmp,
    collections::{BTreeMap, BTreeSet, HashMap, HashSet},
    fmt, result,
};

// PolyType represents a generic parametrized type.
//
// TODO:
//     Do not derive PartialEq implementation.
//     Instead provide a custom implementation
//     that instantiates both polytypes with the
//     same type variables.
//
//     Note this requires a substitution, so remove
//     this derivation once substitutions are defined.
//
#[derive(Debug, Clone, PartialEq)]
pub struct PolyType {
    pub free: Vec<Tvar>,
    pub cons: Option<TvarKinds>,
    pub expr: MonoType,
}

impl fmt::Display for PolyType {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        let vars = self.free.iter().map(|x| x.to_string()).collect::<Vec<_>>();
        match &self.cons {
            Some(constraints) => write!(
                f,
                "forall [{}] where {} {}",
                vars.join(", "),
                constraints,
                self.expr,
            ),
            None => write!(f, "forall [{}] {}", vars.join(", "), self.expr),
        }
    }
}

// TvarKinds maps a type variable to the kinds to which it must belong.
//
// Note that during inference we might infer that a type variable is of
// a particular kind (type class) without inferring an exact monotype
// for said type variable.
//
#[derive(Debug, Clone, PartialEq)]
pub struct TvarKinds(pub HashMap<Tvar, HashSet<Kind>>);

impl fmt::Display for TvarKinds {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        f.write_str(
            &self
                .0
                .iter()
                // A BTree produces a sorted iterator for
                // deterministic display output
                .collect::<BTreeMap<_, _>>()
                .iter()
                .map(|(&&tv, &kinds)| display_kinds(tv, kinds))
                .collect::<Vec<_>>()
                .join(", "),
        )
    }
}

fn display_kinds(tv: Tvar, kinds: &HashSet<Kind>) -> String {
    format!(
        "{}:{}",
        tv,
        kinds
            .iter()
            // Sort kinds with BTree
            .collect::<BTreeSet<_>>()
            .iter()
            .map(|x| x.to_string())
            .collect::<Vec<_>>()
            .join(" + "),
    )
}

impl TvarKinds {
    fn update(self, tv: Tvar, with: HashSet<Kind>) -> Self {
        let mut constraints = self;
        if let Some(kinds) = constraints.0.get(&tv) {
            let new: HashSet<Kind> = kinds.union(&with).map(|&kind| kind).collect();
            constraints.0.insert(tv, new);
            constraints
        } else {
            constraints.0.insert(tv, with);
            constraints
        }
    }
}

pub type Result = result::Result<(Subst, TvarKinds), Error>;

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
    fn cannot_constrain(t: MonoType, with: HashSet<Kind>) -> Error {
        Error {
            msg: format!(
                "{} is not of kind {}",
                t,
                with.iter()
                    .map(|kind| kind.to_string())
                    .collect::<Vec<String>>()
                    .join(" | ")
            ),
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
    fn apply(self, sub: &Subst) -> Self {
        match self {
            MonoType::Bool => MonoType::Bool,
            MonoType::Int => MonoType::Int,
            MonoType::Uint => MonoType::Uint,
            MonoType::Float => MonoType::Float,
            MonoType::String => MonoType::String,
            MonoType::Duration => MonoType::Duration,
            MonoType::Time => MonoType::Time,
            MonoType::Regexp => MonoType::Regexp,
            MonoType::Var(tvr) => tvr.apply(sub),
            MonoType::Arr(arr) => MonoType::Arr(Box::new(arr.apply(sub))),
            MonoType::Row(obj) => MonoType::Row(Box::new(obj.apply(sub))),
            MonoType::Fun(fun) => MonoType::Fun(Box::new(fun.apply(sub))),
        }
    }
}

impl MonoType {
    fn unify(self, with: Self, cons: TvarKinds) -> Result {
        match (self, with) {
            (MonoType::Bool, MonoType::Bool)
            | (MonoType::Int, MonoType::Int)
            | (MonoType::Uint, MonoType::Uint)
            | (MonoType::Float, MonoType::Float)
            | (MonoType::String, MonoType::String)
            | (MonoType::Duration, MonoType::Duration)
            | (MonoType::Time, MonoType::Time)
            | (MonoType::Regexp, MonoType::Regexp) => Ok((Subst::empty(), cons)),
            (MonoType::Var(tv), t) => tv.unify(t, cons),
            (t, MonoType::Var(tv)) => tv.unify(t, cons),
            (MonoType::Arr(t), MonoType::Arr(s)) => t.unify(*s, cons),
            (MonoType::Row(t), MonoType::Row(s)) => t.unify(*s, cons),
            (MonoType::Fun(t), MonoType::Fun(s)) => t.unify(*s, cons),
            (t, with) => Err(Error::cannot_unify(t, with)),
        }
    }

    fn constrain(self, with: HashSet<Kind>, cons: TvarKinds) -> Result {
        match self {
            MonoType::Bool => {
                self.satisfies(maplit::hashset![Kind::Equatable, Kind::Nullable,])(with, cons)
            }
            MonoType::Int => self.satisfies(maplit::hashset![
                Kind::Addable,
                Kind::Subtractable,
                Kind::Divisible,
                Kind::Comparable,
                Kind::Equatable,
                Kind::Nullable
            ])(with, cons),
            MonoType::Uint => self.satisfies(maplit::hashset![
                Kind::Addable,
                Kind::Divisible,
                Kind::Comparable,
                Kind::Equatable,
                Kind::Nullable
            ])(with, cons),
            MonoType::Float => self.satisfies(maplit::hashset![
                Kind::Addable,
                Kind::Subtractable,
                Kind::Divisible,
                Kind::Comparable,
                Kind::Equatable,
                Kind::Nullable
            ])(with, cons),
            MonoType::String => self.satisfies(maplit::hashset![
                Kind::Addable,
                Kind::Comparable,
                Kind::Equatable,
                Kind::Nullable
            ])(with, cons),
            MonoType::Duration => self.satisfies(maplit::hashset![
                Kind::Comparable,
                Kind::Equatable,
                Kind::Nullable,
            ])(with, cons),
            MonoType::Time => self.satisfies(maplit::hashset![
                Kind::Comparable,
                Kind::Equatable,
                Kind::Nullable,
            ])(with, cons),
            MonoType::Regexp => Err(Error::cannot_constrain(self, with)),
            MonoType::Var(tvr) => tvr.constrain(with, cons),
            MonoType::Arr(arr) => arr.constrain(with, cons),
            MonoType::Row(obj) => obj.constrain(with, cons),
            MonoType::Fun(fun) => fun.constrain(with, cons),
        }
    }

    // Returns a closure that specifies the possible type classes (kinds) to which a monotype belongs
    fn satisfies(self, kinds: HashSet<Kind>) -> impl FnOnce(HashSet<Kind>, TvarKinds) -> Result {
        move |k, constraints| {
            if k.is_subset(&kinds) {
                Ok((Subst::empty(), constraints))
            } else {
                Err(Error::cannot_constrain(
                    self,
                    k.difference(&kinds).map(|&kind| kind).collect(),
                ))
            }
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
pub struct Tvar(pub i64);

impl fmt::Display for Tvar {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "t{}", self.0)
    }
}

// Fresher returns a fresh type variable with incrementing id.
pub struct Fresher(i64);

impl Fresher {
    pub fn new() -> Fresher {
        Fresher(0)
    }

    pub fn fresh(&mut self) -> Tvar {
        self.0 += 1;
        Tvar(self.0)
    }
}

impl Tvar {
    fn apply(self, sub: &Subst) -> MonoType {
        match sub.lookup(self) {
            Some(t) => t.clone(),
            None => MonoType::Var(self),
        }
    }

    fn unify(self, with: MonoType, cons: TvarKinds) -> Result {
        match with {
            MonoType::Var(tv) => {
                if self == tv {
                    // The empty substitution will always
                    // unify a type variable with itself.
                    Ok((Subst::empty(), cons))
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

    fn unify_with_tvar(self, tv: Tvar, cons: TvarKinds) -> Result {
        let mut cons = cons;
        // Gather the kind constraints for both type variables
        let kinds: HashSet<Kind> = cons
            .0
            .get(&self)
            .unwrap_or(&HashSet::new())
            .union(cons.0.get(&tv).unwrap_or(&HashSet::new()))
            .map(|&kind| kind)
            .collect();
        let sub = Subst::init(maplit::hashmap! {self => MonoType::Var(tv)});
        if kinds.len() > 0 {
            // Update the kind constraints
            cons.0.remove(&self);
            cons.0.insert(tv, kinds);
        };
        Ok((sub, cons))
    }

    fn unify_with_type(self, t: MonoType, cons: TvarKinds) -> Result {
        let mut cons = cons;
        let sub = Subst::init(maplit::hashmap! {self => t.clone()});
        if let Some(kinds) = cons.0.remove(&self) {
            // The monotype must satisfy all the constraints of
            // the type variable.
            let (s, cons) = t.constrain(kinds, cons)?;
            Ok((sub.merge(s), cons))
        } else {
            Ok((sub, cons))
        }
    }

    fn constrain(self, with: HashSet<Kind>, cons: TvarKinds) -> Result {
        Ok((Subst::empty(), cons.update(self, with)))
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
    fn apply(self, sub: &Subst) -> Self {
        Array(self.0.apply(sub))
    }
}

impl Array {
    fn unify(self, with: Self, cons: TvarKinds) -> Result {
        self.0.unify(with.0, cons)
    }

    fn constrain(self, with: HashSet<Kind>, _: TvarKinds) -> Result {
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
    fn apply(self, sub: &Subst) -> Self {
        match self {
            Row::Empty => Row::Empty,
            Row::Extension { head, tail } => Row::Extension {
                head: head.apply(sub),
                tail: tail.apply(sub),
            },
        }
    }
}

impl Row {
    fn unify(self, _: Self, _: TvarKinds) -> Result {
        unimplemented!();
    }

    fn constrain(self, with: HashSet<Kind>, _: TvarKinds) -> Result {
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
    fn apply(self, sub: &Subst) -> Self {
        Property {
            k: self.k,
            v: self.v.apply(sub),
        }
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

impl Substitutable for HashMap<String, MonoType> {
    fn apply(self, sub: &Subst) -> Self {
        self.into_iter().map(|(k, v)| (k, v.apply(sub))).collect()
    }
}

impl Substitutable for Function {
    fn apply(self, sub: &Subst) -> Self {
        Function {
            req: self.req.apply(sub),
            opt: self.opt.apply(sub),
            pipe: match self.pipe {
                None => None,
                Some(p) => Some(p.apply(sub)),
            },
            retn: self.retn.apply(sub),
        }
    }
}

impl Function {
    fn unify(self, _: Self, _: TvarKinds) -> Result {
        unimplemented!();
    }

    fn constrain(self, with: HashSet<Kind>, _: TvarKinds) -> Result {
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
                free: Vec::new(),
                cons: None,
                expr: MonoType::Int,
            }
            .to_string(),
        );
        assert_eq!(
            "forall [t0] (x:t0) -> t0",
            PolyType {
                free: vec![Tvar(0)],
                cons: None,
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
                free: vec![Tvar(0), Tvar(1)],
                cons: None,
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
                free: vec![Tvar(0)],
                cons: Some(TvarKinds(
                    maplit::hashmap! {Tvar(0) => maplit::hashset![Kind::Addable]}
                )),
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
                free: vec![Tvar(0), Tvar(1)],
                cons: Some(TvarKinds(maplit::hashmap! {
                    Tvar(0) => maplit::hashset![Kind::Addable],
                    Tvar(1) => maplit::hashset![Kind::Divisible],
                })),
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
                free: vec![Tvar(0), Tvar(1)],
                cons: Some(TvarKinds(maplit::hashmap! {
                    Tvar(0) => maplit::hashset![Kind::Comparable, Kind::Equatable],
                    Tvar(1) => maplit::hashset![Kind::Addable, Kind::Divisible],
                })),
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
        let (sub, cons) = MonoType::Int
            .unify(MonoType::Int, TvarKinds(HashMap::new()))
            .unwrap();
        assert_eq!(sub, Subst::empty());
        assert_eq!(cons, TvarKinds(HashMap::new()));
    }
    #[test]
    fn unify_error() {
        let err = MonoType::Int
            .unify(MonoType::String, TvarKinds(HashMap::new()))
            .unwrap_err();
        assert_eq!(
            err.to_string(),
            String::from("cannot unify int with string"),
        );
    }
    #[test]
    fn unify_tvars() {
        let (sub, cons) = MonoType::Var(Tvar(0))
            .unify(MonoType::Var(Tvar(1)), TvarKinds(HashMap::new()))
            .unwrap();
        assert_eq!(
            sub,
            Subst::init(maplit::hashmap! {Tvar(0) => MonoType::Var(Tvar(1))}),
        );
        assert_eq!(cons, TvarKinds(HashMap::new()));
    }
    #[test]
    fn unify_constrained_tvars() {
        let (sub, cons) = MonoType::Var(Tvar(0))
            .unify(
                MonoType::Var(Tvar(1)),
                TvarKinds(
                    maplit::hashmap! {Tvar(0) => maplit::hashset![Kind::Addable, Kind::Divisible]},
                ),
            )
            .unwrap();
        assert_eq!(
            sub,
            Subst::init(maplit::hashmap! {Tvar(0) => MonoType::Var(Tvar(1))})
        );
        assert_eq!(
            cons,
            TvarKinds(
                maplit::hashmap! {Tvar(1) => maplit::hashset![Kind::Addable, Kind::Divisible]}
            ),
        );
    }
}

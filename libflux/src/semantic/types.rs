use std::{cmp, collections::HashMap, fmt};

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
    pub bnds: Option<HashMap<Tvar, Kind>>,
    pub expr: MonoType,
}

impl fmt::Display for PolyType {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match &self.bnds {
            Some(bnds) => {
                let mut bounds = Vec::new();
                for tv in &self.free {
                    if let Some(kind) = bnds.get(tv) {
                        bounds.push(BoundTvar {
                            tv: *tv,
                            kind: *kind,
                        })
                    }
                }
                write!(
                    f,
                    "forall [{}] where {} {}",
                    DisplayList {
                        values: &self.free,
                        delim: ", "
                    },
                    DisplayList {
                        values: &bounds,
                        delim: ", "
                    },
                    self.expr,
                )
            }
            None => write!(
                f,
                "forall [{}] {}",
                DisplayList {
                    values: &self.free,
                    delim: ", "
                },
                self.expr
            ),
        }
    }
}

// Kind represents a class or family of types
#[derive(Debug, PartialEq, Clone, Copy)]
pub enum Kind {
    Addable,
    Subtractable,
    Divisible,
    Comparable,
    Nullable,
}

impl fmt::Display for Kind {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match self {
            Kind::Addable => f.write_str("Addable"),
            Kind::Subtractable => f.write_str("Subtractable"),
            Kind::Divisible => f.write_str("Divisible"),
            Kind::Comparable => f.write_str("Comparable"),
            Kind::Nullable => f.write_str("Nullable"),
        }
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

// Tvar stands for type variable.
// A type variable holds an unknown type.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub struct Tvar(pub i64);

impl fmt::Display for Tvar {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "t{}", self.0)
    }
}

// Array is a homogeneous list type
#[derive(Debug, Clone, PartialEq)]
pub struct Array(MonoType);

impl fmt::Display for Array {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "[{}]", self.0)
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
    Var(Tvar),
    Extension { head: Property, tail: Box<Row> },
}

impl fmt::Display for Row {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        if let Err(e) = f.write_str("{") {
            return Err(e);
        }
        if let Err(e) = self.display(f) {
            return Err(e);
        }
        return f.write_str("}");
    }
}

impl cmp::PartialEq for Row {
    fn eq(&self, other: &Self) -> bool {
        let mut l = HashMap::new();
        let mut r = HashMap::new();
        self.flatten(&mut l) == other.flatten(&mut r) && l == r
    }
}

impl Row {
    // Records are implemented as a sequence or list of extenstions.
    // This function extends a record by adding a new property to the
    // head of the list.
    fn extend(self, head: Property) -> Self {
        Row::Extension {
            head: head,
            tail: Box::new(self),
        }
    }

    // Display a row type in flattened format
    fn display(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match self {
            Row::Empty => f.write_str("{}"),
            Row::Var(tv) => write!(f, "{}", tv),
            Row::Extension { head: h, tail: t } => {
                if let Err(e) = write!(f, "{} | ", h) {
                    return Err(e);
                }
                t.display(f)
            }
        }
    }

    // Flatten a record type into a hashmap of property names and types
    fn flatten(&self, props: &mut HashMap<String, MonoType>) -> Option<Tvar> {
        match self {
            Row::Empty => None,
            Row::Var(tv) => Some(*tv),
            Row::Extension { head: h, tail: t } => {
                props.insert(h.k.clone(), h.v.clone());
                t.flatten(props)
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
        let mut req: Vec<Property> = Vec::new();
        for (k, v) in &self.req {
            req.push(Property {
                k: k.clone(),
                v: v.clone(),
            });
        }
        req.sort_unstable_by(|a, b| a.k.cmp(&b.k));

        let mut opt: Vec<Property> = Vec::new();
        for (k, v) in &self.opt {
            opt.push(Property {
                k: String::from("?") + &k,
                v: v.clone(),
            });
        }
        opt.sort_unstable_by(|a, b| a.k.cmp(&b.k));

        let mut args: Vec<Property> = Vec::new();
        if let Some(pipe) = &self.pipe {
            if pipe.k == "<-" {
                args.push(pipe.clone());
            } else {
                args.push(Property {
                    k: String::from("<-") + &pipe.k,
                    v: pipe.v.clone(),
                });
            }
        }

        args.append(&mut req);
        args.append(&mut opt);
        write!(
            f,
            "({}) -> {}",
            DisplayList {
                values: &args,
                delim: ", "
            },
            self.retn
        )
    }
}

// BoundTvar represents a constrained type variable.
// Used solely for displaying the generic constraints of a polytype.
#[derive(Debug)]
struct BoundTvar {
    tv: Tvar,
    kind: Kind,
}

impl fmt::Display for BoundTvar {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "{}:{}", self.tv, self.kind)
    }
}

// DisplayList is a list of elements each of which can be displayed
#[derive(Debug, Clone, PartialEq)]
struct DisplayList<'a, T> {
    values: &'a Vec<T>,
    delim: &'static str,
}

impl<'a, T: fmt::Display> fmt::Display for DisplayList<'a, T> {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        if self.values.is_empty() {
            return Ok(());
        }
        let size = self.values.len();
        let list = &self.values[..size - 1];
        for v in list {
            if let Err(e) = write!(f, "{}{}", v, self.delim) {
                return Err(e);
            }
        }
        self.values[size - 1].fmt(f)
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
            Row::Var(Tvar(0))
                .extend(Property {
                    k: String::from("b"),
                    v: MonoType::String,
                })
                .extend(Property {
                    k: String::from("a"),
                    v: MonoType::Int,
                })
                .to_string()
        );
        assert_eq!(
            "{a:int | b:string | {}}",
            Row::Empty
                .extend(Property {
                    k: String::from("b"),
                    v: MonoType::String,
                })
                .extend(Property {
                    k: String::from("a"),
                    v: MonoType::Int,
                })
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
                bnds: None,
                expr: MonoType::Int,
            }
            .to_string(),
        );
        assert_eq!(
            "forall [t0] (x:t0) -> t0",
            PolyType {
                free: vec![Tvar(0)],
                bnds: None,
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
                bnds: None,
                expr: MonoType::Fun(Box::new(Function {
                    req: maplit::hashmap! {
                        String::from("x") => MonoType::Var(Tvar(0)),
                        String::from("y") => MonoType::Var(Tvar(1)),
                    },
                    opt: HashMap::new(),
                    pipe: None,
                    retn: MonoType::Row(Box::new(
                        Row::Empty
                            .extend(Property {
                                k: String::from("y"),
                                v: MonoType::Var(Tvar(1)),
                            })
                            .extend(Property {
                                k: String::from("x"),
                                v: MonoType::Var(Tvar(0)),
                            })
                    )),
                })),
            }
            .to_string(),
        );
        assert_eq!(
            "forall [t0] where t0:Addable (a:t0, b:t0) -> t0",
            PolyType {
                free: vec![Tvar(0)],
                bnds: Some(maplit::hashmap! {
                    Tvar(0) => Kind::Addable,
                }),
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
                bnds: Some(maplit::hashmap! {
                    Tvar(0) => Kind::Addable,
                    Tvar(1) => Kind::Divisible,
                }),
                expr: MonoType::Fun(Box::new(Function {
                    req: maplit::hashmap! {
                        String::from("x") => MonoType::Var(Tvar(0)),
                        String::from("y") => MonoType::Var(Tvar(1)),
                    },
                    opt: HashMap::new(),
                    pipe: None,
                    retn: MonoType::Row(Box::new(
                        Row::Empty
                            .extend(Property {
                                k: String::from("y"),
                                v: MonoType::Var(Tvar(1)),
                            })
                            .extend(Property {
                                k: String::from("x"),
                                v: MonoType::Var(Tvar(0)),
                            })
                    )),
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
            MonoType::Row(Box::new(
                Row::Var(Tvar(0))
                    .extend(Property {
                        k: String::from("b"),
                        v: MonoType::String,
                    })
                    .extend(Property {
                        k: String::from("a"),
                        v: MonoType::Int,
                    })
            )),
            // {b:string | a:int | t0}
            MonoType::Row(Box::new(
                Row::Var(Tvar(0))
                    .extend(Property {
                        k: String::from("a"),
                        v: MonoType::Int,
                    })
                    .extend(Property {
                        k: String::from("b"),
                        v: MonoType::String,
                    })
            )),
        );
        assert_ne!(
            // {a:int | b:string | {}}
            MonoType::Row(Box::new(
                Row::Empty
                    .extend(Property {
                        k: String::from("b"),
                        v: MonoType::String,
                    })
                    .extend(Property {
                        k: String::from("a"),
                        v: MonoType::Int,
                    })
            )),
            // {b:int | a:int | {}}
            MonoType::Row(Box::new(
                Row::Empty
                    .extend(Property {
                        k: String::from("a"),
                        v: MonoType::Int,
                    })
                    .extend(Property {
                        k: String::from("b"),
                        v: MonoType::Int,
                    })
            )),
        );
    }
}

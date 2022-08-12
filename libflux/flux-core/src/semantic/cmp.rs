//! Utility to compare Flux types with each other.

use std::cmp::max;

use crate::semantic::types::{collect_record, MonoType, PolyType};

/// Represents a difference between two types.
#[derive(PartialEq, Eq, PartialOrd, Ord, Clone, Debug, Copy)]
pub enum TypeDiff {
    /// Represents a backwards incomptible change to the types
    Major = 2,
    /// Represents a backwards comptible change to type
    Minor = 1,
    /// Represents no change to the types
    Patch = 0,
}

/// Report the difference between the old and new polytypes.
pub fn diff(old: &PolyType, new: &PolyType) -> TypeDiff {
    println!("{} {}", old, new);
    if old.vars != new.vars {
        return TypeDiff::Major;
    }
    if old.cons != new.cons {
        return TypeDiff::Major;
    }
    diff_monotype(&old.expr, &new.expr)
}

fn diff_monotype(old: &MonoType, new: &MonoType) -> TypeDiff {
    match (old, new) {
        (MonoType::Error, MonoType::Error) => TypeDiff::Patch,
        (MonoType::Builtin(o), MonoType::Builtin(n)) => {
            if o == n {
                TypeDiff::Patch
            } else {
                TypeDiff::Major
            }
        }
        (MonoType::Label(o), MonoType::Label(n)) => {
            if o == n {
                TypeDiff::Patch
            } else {
                TypeDiff::Major
            }
        }
        (MonoType::Var(o), MonoType::Var(n)) => {
            if o == n {
                TypeDiff::Patch
            } else {
                TypeDiff::Major
            }
        }
        (MonoType::BoundVar(o), MonoType::BoundVar(n)) => {
            if o == n {
                TypeDiff::Patch
            } else {
                TypeDiff::Major
            }
        }
        (MonoType::Collection(old), MonoType::Collection(new)) => {
            if old.collection != new.collection {
                TypeDiff::Major
            } else {
                diff_monotype(&old.arg, &new.arg)
            }
        }
        (MonoType::Record(old), MonoType::Record(new)) => {
            let mut diff = TypeDiff::Patch;
            let (ofields, otail) = collect_record(old);
            let (nfields, ntail) = collect_record(new);
            for (key, ofield) in ofields.iter() {
                if let Some(nfield) = nfields.get(key) {
                    if ofield.len() != nfield.len() {
                        return diff;
                    }
                    for (i, of) in ofield.iter().enumerate() {
                        let nf = nfield[i];
                        diff = max(diff, diff_monotype(&of, &nf));
                        if diff == TypeDiff::Major {
                            return diff;
                        }
                    }
                } else {
                    return TypeDiff::Major;
                }
            }
            match (otail, ntail) {
                (Some(old), Some(new)) => {
                    diff = max(diff, diff_monotype(old, new));
                    if diff == TypeDiff::Major {
                        return diff;
                    }
                }
                (None, None) => {}
                (_, _) => {
                    return TypeDiff::Major;
                }
            };
            for (key, _) in nfields.iter() {
                if ofields.get(key).is_none() {
                    // Adding a new field to a record is a major change
                    return TypeDiff::Major;
                }
            }
            diff
        }
        (MonoType::Dict(ref old), MonoType::Dict(ref new)) => max(
            diff_monotype(&old.key, &new.key),
            diff_monotype(&old.val, &new.val),
        ),
        (MonoType::Fun(old), MonoType::Fun(new)) => {
            let mut diff = TypeDiff::Patch;
            // Check for missing old required args
            for (ok, ov) in &old.req {
                if let Some((_, nv)) = new.req.iter().find(|(nk, _)| *nk == ok) {
                    diff = max(diff, diff_monotype(ov, nv));
                    if diff == TypeDiff::Major {
                        return diff;
                    }
                } else {
                    return TypeDiff::Major;
                }
            }
            // Check for new required args
            for (nk, _) in &new.req {
                if old.req.iter().find(|(ok, _)| *ok == nk).is_none() {
                    return TypeDiff::Major;
                }
            }
            // Check for missing old optional args
            for (ok, ov) in &old.opt {
                if let Some((_, nv)) = new.opt.iter().find(|(nk, _)| *nk == ok) {
                    match (ov.default.as_ref(), nv.default.as_ref()) {
                        (Some(old), Some(new)) => {
                            diff = max(diff, diff_monotype(&old, &new));
                            if diff == TypeDiff::Major {
                                return diff;
                            }
                        }
                        (None, None) => {}
                        (_, _) => {
                            return TypeDiff::Major;
                        }
                    }
                    diff = max(diff, diff_monotype(&ov.typ, &nv.typ));
                    if diff == TypeDiff::Major {
                        return diff;
                    }
                } else {
                    return TypeDiff::Major;
                }
            }
            // Check for new optional args
            for (nk, _) in &new.opt {
                if old.opt.iter().find(|(ok, _)| *ok == nk).is_none() {
                    diff = max(diff, TypeDiff::Minor);
                }
            }
            match (old.pipe.as_ref(), new.pipe.as_ref()) {
                (Some(old), Some(new)) => {
                    if old.k != new.k {
                        return TypeDiff::Major;
                    }
                    diff = max(diff, diff_monotype(&old.v, &new.v));
                    if diff == TypeDiff::Major {
                        return diff;
                    }
                }
                (None, None) => {}
                (_, _) => {
                    return TypeDiff::Major;
                }
            };
            max(diff, diff_monotype(&old.retn, &new.retn))
        }
        // If the core monotype is different its a major change
        (_, _) => TypeDiff::Major,
    }
}

#[cfg(test)]
mod tests {
    use super::{diff, TypeDiff};
    use crate::{
        parser,
        semantic::{convert::convert_polytype, types::PolyType, AnalyzerConfig},
    };

    fn str_to_type(typ: &str) -> PolyType {
        let type_expr = parser::Parser::new(typ).parse_type_expression();
        convert_polytype(&type_expr, &AnalyzerConfig::default()).unwrap()
    }
    macro_rules! test {
        ($o:expr, $n:expr, $d:expr) => {
            let old = str_to_type($o);
            let new = str_to_type($n);

            assert_eq!($d, diff(&old, &new));
        };
    }

    #[test]
    fn test_basic() {
        test!("int", "int", TypeDiff::Patch);
        test!("int", "float", TypeDiff::Major);
    }
    #[test]
    fn test_record() {
        test!("{a:int}", "{a:int}", TypeDiff::Patch);
        test!("{a:int}", "{a:float}", TypeDiff::Major);
        test!("{a:int,b:int}", "{a:int}", TypeDiff::Major);
        test!("{a:int,b:int}", "{a:int,b:int}", TypeDiff::Patch);
        test!("{a:int,b:int | C}", "{a:int,b:int | C}", TypeDiff::Patch);
        test!("{a:int,b:int}", "{a:int,b:int | C}", TypeDiff::Major);
    }

    #[test]
    fn test_function() {
        test!("() => int", "() => int", TypeDiff::Patch);
        test!("() => int", "() => float", TypeDiff::Major);
        test!("(a:int) => int", "(a:int) => int", TypeDiff::Patch);
        test!("(<-a:int) => int", "(<-a:int) => int", TypeDiff::Patch);
        test!(
            "(<-a:int,b:float) => int",
            "(<-a:int,b:float) => int",
            TypeDiff::Patch
        );
        test!(
            "(<-a:int,b:float) => int",
            "(a:int,<-b:float) => int",
            TypeDiff::Major
        );
        test!("(a:int) => int", "(a:int,?b:int) => int", TypeDiff::Minor);
        test!("(a:int) => int", "() => int", TypeDiff::Major);
        test!("(a:int) => int", "(a:int, b:float) => int", TypeDiff::Major);
        test!("(a:int,?b:float) => int", "(a:int, b:float) => int", TypeDiff::Major);
    }

    #[test]
    fn test_constraints() {
        test!("A where A: Record", "A where A: Record", TypeDiff::Patch);
        test!("A where A: Record", "A where A: Addable", TypeDiff::Major);
    }

    #[test]
    fn test_collection() {
        test!("[int]", "[int]", TypeDiff::Patch);
        test!("[int]", "[float]", TypeDiff::Major);

        test!("stream[int]", "stream[int]", TypeDiff::Patch);
        test!("stream[int]", "stream[float]", TypeDiff::Major);

        test!("vector[int]", "vector[int]", TypeDiff::Patch);
        test!("vector[int]", "vector[float]", TypeDiff::Major);

        test!("stream[int]", "vector[float]", TypeDiff::Major);
    }

    #[test]
    fn test_dict() {
        test!("[int:int]", "[int:int]", TypeDiff::Patch);
        test!("[int:int]", "[int:float]", TypeDiff::Major);
    }
}

use crate::semantic::types::{Kind, MonoType};
use std::ops;

// Type constraints are produced during type inference and come
// in two flavors.
//
// A kind constraint asserts that a particular type is of a
// particular kind or family of types.
//
// An equality contraint asserts that two types are equivalent
// and will be unified at some point.
//
#[derive(Debug, PartialEq)]
enum Constraint {
    Kind(MonoType, Kind),
    Equal(MonoType, MonoType),
}

#[derive(Debug, PartialEq)]
struct Constraints(Vec<Constraint>);

// Constraints can be added using the '+' operator
impl ops::Add for Constraints {
    type Output = Constraints;

    fn add(self, cons: Constraints) -> Self::Output {
        Constraints(self.0.into_iter().chain(cons.0.into_iter()).collect())
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::semantic::types::Tvar;

    #[test]
    fn add_constraints() {
        let c0 = Constraints(vec![
            Constraint::Equal(MonoType::Var(Tvar(0)), MonoType::Var(Tvar(1))),
            Constraint::Kind(MonoType::Var(Tvar(1)), Kind::Addable),
        ]);
        let c1 = Constraints(vec![
            Constraint::Equal(MonoType::Var(Tvar(2)), MonoType::Var(Tvar(3))),
            Constraint::Kind(MonoType::Var(Tvar(3)), Kind::Divisible),
        ]);
        assert_eq!(
            c0 + c1,
            Constraints(vec![
                Constraint::Equal(MonoType::Var(Tvar(0)), MonoType::Var(Tvar(1))),
                Constraint::Kind(MonoType::Var(Tvar(1)), Kind::Addable),
                Constraint::Equal(MonoType::Var(Tvar(2)), MonoType::Var(Tvar(3))),
                Constraint::Kind(MonoType::Var(Tvar(3)), Kind::Divisible),
            ])
        );
    }
}

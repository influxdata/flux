use crate::semantic::env::Environment;
use crate::semantic::fresh::Fresher;
use crate::semantic::sub::{Substitutable, Substitution};
use crate::semantic::types::{minus, Error, Kind, MonoType, PolyType, Tvar};
use std::collections::HashMap;
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
pub enum Constraint {
    Kind(MonoType, Kind),
    Equal(MonoType, MonoType),
}

#[derive(Debug, PartialEq)]
pub struct Constraints(Vec<Constraint>);

impl Constraints {
    pub fn empty() -> Constraints {
        Constraints(Vec::new())
    }
}

// Constraints can be added using the '+' operator
impl ops::Add for Constraints {
    type Output = Constraints;

    fn add(mut self, mut cons: Constraints) -> Self::Output {
        self.0.append(&mut cons.0);
        self
    }
}

impl From<Vec<Constraint>> for Constraints {
    fn from(constraints: Vec<Constraint>) -> Constraints {
        Constraints(constraints)
    }
}

impl From<Constraints> for Vec<Constraint> {
    fn from(constraints: Constraints) -> Vec<Constraint> {
        constraints.0
    }
}

// Solve a set of type constraints
pub fn solve(
    cons: &Constraints,
    with: &mut HashMap<Tvar, Vec<Kind>>,
) -> Result<Substitution, Error> {
    cons.0
        .iter()
        .try_fold(Substitution::empty(), |sub, constraint| match constraint {
            Constraint::Kind(t, kind) => {
                // Apply the current substitution to the type, then constrain
                let s = t.clone().apply(&sub).constrain(*kind, with)?;
                Ok(sub.merge(s))
            }
            Constraint::Equal(t, other) => {
                // Apply the current substitution to the constraint, then unify
                let s = t.clone().apply(&sub).unify(other.clone(), with)?;
                Ok(sub.merge(s))
            }
        })
}

// Create a parametric type from a monotype by universally quantifying
// all of its free type variables.
//
// A type variable is free in a monotype if it is **free** in the global
// type environment. Equivalently a type variable is free and can be
// quantified if has not already been quantified another type in the
// type environment.
//
pub fn generalize(env: &Environment, with: &HashMap<Tvar, Vec<Kind>>, t: MonoType) -> PolyType {
    let vars = minus(&env.free_vars(), t.free_vars());
    let mut cons = HashMap::new();
    for (tv, kinds) in with {
        if vars.contains(tv) {
            cons.insert(*tv, kinds.to_owned());
        }
    }
    PolyType {
        vars: vars,
        cons: cons,
        expr: t,
    }
}

// Instantiate a new monotype from a polytype by assigning all universally
// quantified type variables new fresh variables that do not exist in the
// current type environment.
//
// Instantiation is what allows for polymorphic function specialization
// based on the context in which a function is called.
pub fn instantiate(poly: PolyType, f: &mut Fresher) -> (MonoType, Constraints) {
    // Substitute fresh type variables for all quantified variables
    let sub: Substitution = poly
        .vars
        .into_iter()
        .map(|tv| (tv, MonoType::Var(f.fresh())))
        .collect::<HashMap<Tvar, MonoType>>()
        .into();
    // Generate constraints for the new fresh type variables
    let constraints = poly
        .cons
        .into_iter()
        .fold(Constraints::empty(), |cons, (tv, kinds)| {
            cons + kinds
                .into_iter()
                .map(|kind| Constraint::Kind(sub.apply(tv), kind))
                .collect::<Vec<Constraint>>()
                .into()
        });
    // Instantiate monotype using new fresh type variables
    (poly.expr.apply(&sub), constraints)
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

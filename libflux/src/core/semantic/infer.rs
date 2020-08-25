use crate::ast::SourceLocation;
use crate::semantic::env::Environment;
use crate::semantic::fresh::Fresher;
use crate::semantic::sub::{Substitutable, Substitution};
use crate::semantic::types;
use crate::semantic::types::{minus, Kind, MonoType, PolyType, SubstitutionMap, TvarKinds};
use std::fmt;
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
// A constraint is composed of an expected type, the actual type
// that was found, and the source location of the actual type.
//
#[derive(Debug, PartialEq)]
pub enum Constraint {
    Kind {
        exp: Kind,
        act: MonoType,
        loc: SourceLocation,
    },
    Equal {
        exp: MonoType,
        act: MonoType,
        loc: SourceLocation,
    },
}

#[derive(Debug, PartialEq)]
pub struct Constraints(Vec<Constraint>);

impl Constraints {
    pub fn empty() -> Constraints {
        Constraints(Vec::new())
    }

    pub fn add(&mut self, cons: Constraint) {
        self.0.push(cons);
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

impl From<Constraint> for Constraints {
    fn from(constraint: Constraint) -> Constraints {
        Constraints::from(vec![constraint])
    }
}

#[derive(Debug, PartialEq)]
pub struct Error {
    loc: SourceLocation,
    err: types::Error,
}

impl fmt::Display for Error {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "type error {}: {}", self.loc, self.err)
    }
}

// Solve a set of type constraints
pub fn solve(
    cons: &Constraints,
    with: &mut TvarKinds,
    fresher: &mut Fresher,
) -> Result<Substitution, Error> {
    cons.0
        .iter()
        .try_fold(Substitution::empty(), |sub, constraint| match constraint {
            Constraint::Kind { exp, act, loc } => {
                // Apply the current substitution to the type, then constrain
                let s = match act.clone().apply(&sub).constrain(*exp, with) {
                    Err(e) => Err(Error {
                        loc: loc.clone(),
                        err: e,
                    }),
                    Ok(s) => Ok(s),
                }?;
                Ok(sub.merge(s))
            }
            Constraint::Equal { exp, act, loc } => {
                // Apply the current substitution to the constraint, then unify
                let exp = exp.clone().apply(&sub);
                let act = act.clone().apply(&sub);
                let s = match exp.unify(act, with, fresher) {
                    Err(e) => Err(Error {
                        loc: loc.clone(),
                        err: e,
                    }),
                    Ok(s) => Ok(s),
                }?;
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
pub fn generalize(env: &Environment, with: &TvarKinds, t: MonoType) -> PolyType {
    let vars = minus(&env.free_vars(), t.free_vars());
    let mut cons = TvarKinds::new();
    for (tv, kinds) in with {
        if vars.contains(tv) {
            cons.insert(*tv, kinds.to_owned());
        }
    }
    PolyType {
        vars,
        cons,
        expr: t,
    }
}

// Instantiate a new monotype from a polytype by assigning all universally
// quantified type variables new fresh variables that do not exist in the
// current type environment.
//
// Instantiation is what allows for polymorphic function specialization
// based on the context in which a function is called.
pub fn instantiate(
    poly: PolyType,
    f: &mut Fresher,
    loc: SourceLocation,
) -> (MonoType, Constraints) {
    // Substitute fresh type variables for all quantified variables
    let sub: Substitution = poly
        .vars
        .into_iter()
        .map(|tv| (tv, MonoType::Var(f.fresh())))
        .collect::<SubstitutionMap>()
        .into();
    // Generate constraints for the new fresh type variables
    let constraints = poly
        .cons
        .into_iter()
        .fold(Constraints::empty(), |cons, (tv, kinds)| {
            cons + kinds
                .into_iter()
                .map(|kind| Constraint::Kind {
                    exp: kind,
                    act: sub.apply(tv),
                    loc: loc.clone(),
                })
                .collect::<Vec<Constraint>>()
                .into()
        });
    // Instantiate monotype using new fresh type variables
    (poly.expr.apply(&sub), constraints)
}

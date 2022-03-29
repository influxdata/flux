use std::{cell::RefCell, ops};

use derive_more::Display;

use crate::{
    ast::SourceLocation,
    errors::{Errors, Located},
    semantic::{
        env::Environment,
        sub::{Substitutable, Substituter, Substitution},
        types::{self, Kind, MonoType, PolyType, SubstitutionMap, Tvar, TvarKinds},
    },
};

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
#[must_use = "Constraints must be solved"]
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
#[must_use = "Constraints must be solved"]
pub struct Constraints(Vec<Constraint>);

impl std::ops::Deref for Constraints {
    type Target = [Constraint];
    fn deref(&self) -> &[Constraint] {
        &self.0
    }
}

impl AsRef<[Constraint]> for Constraints {
    fn as_ref(&self) -> &[Constraint] {
        &self.0
    }
}

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

    fn add(mut self, cons: Constraints) -> Self::Output {
        self += cons;
        self
    }
}

impl ops::AddAssign for Constraints {
    fn add_assign(&mut self, mut cons: Constraints) {
        self.0.append(&mut cons.0);
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

#[derive(Debug, Display, PartialEq)]
#[display(fmt = "type error {}: {}", loc, err)]
pub struct Error {
    pub loc: SourceLocation,
    pub err: types::Error,
}

impl std::error::Error for Error {}

impl Substitutable for Error {
    fn walk(&self, sub: &dyn Substituter) -> Option<Self> {
        self.err.visit(sub).map(|err| Error {
            loc: self.loc.clone(),
            err,
        })
    }
}

// Solve a set of type constraints
pub fn solve(
    cons: &[Constraint],
    sub: &mut Substitution,
) -> Result<(), Errors<Located<types::Error>>> {
    let mut errors = Errors::new();
    for constraint in cons {
        match constraint {
            Constraint::Kind { exp, act, loc } => {
                // Apply the current substitution to the type, then constrain
                if let Err(err) = constrain(*exp, act, loc, sub) {
                    errors.push(err);
                }
            }
            Constraint::Equal { exp, act, loc } => {
                // Apply the current substitution to the constraint, then unify
                if let Err(err) = equal(exp, act, loc, sub) {
                    errors.extend(err.error.into_iter().map(|error| Located {
                        location: loc.clone(),
                        error,
                    }));
                }
            }
        }
    }
    if errors.has_errors() {
        Err(errors)
    } else {
        Ok(())
    }
}

pub fn constrain(
    exp: Kind,
    act: &MonoType,
    loc: &SourceLocation,
    sub: &mut Substitution,
) -> Result<(), Located<types::Error>> {
    log::debug!("Constraint::Kind {:?}: {} => {}", loc.source, exp, act);
    act.apply_cow(sub)
        .constrain(exp, sub.cons())
        .map_err(|error| Located {
            location: loc.clone(),
            error,
        })
}

pub fn equal(
    exp: &MonoType,
    act: &MonoType,
    loc: &SourceLocation,
    sub: &mut Substitution,
) -> Result<MonoType, Located<Errors<types::Error>>> {
    log::debug!("Constraint::Equal {:?}: {} <===> {}", loc.source, exp, act);
    exp.try_unify(act, sub).map_err(|error| {
        log::debug!("Unify error: {} <=> {}", exp, act);

        Located {
            location: loc.clone(),
            error,
        }
    })
}

/// Generalizes `t` without modifying the substitution.
pub(crate) fn temporary_generalize(
    env: &Environment,
    sub: &mut Substitution,
    t: MonoType,
) -> PolyType {
    struct Generalize {
        env_free_vars: Vec<Tvar>,
        vars: RefCell<Vec<(Tvar, Tvar)>>,
    }

    impl Substituter for Generalize {
        fn try_apply_bound(&self, var: Tvar) -> Option<MonoType> {
            let mut vars = self.vars.borrow_mut();
            if vars.iter().all(|(_, v)| *v != var) {
                vars.push((var, var));
            }
            None
        }
        fn try_apply(&self, var: Tvar) -> Option<MonoType> {
            if !self.env_free_vars.contains(&var) {
                let mut vars = self.vars.borrow_mut();
                match vars.iter().find(|(v, _)| *v == var) {
                    Some((_, new_var)) => Some(MonoType::BoundVar(*new_var)),
                    None => {
                        let new_var = Tvar(vars.len() as u64);
                        vars.push((var, new_var));
                        Some(MonoType::BoundVar(new_var))
                    }
                }
            } else {
                None
            }
        }
    }

    let generalize = Generalize {
        env_free_vars: env.free_vars(),
        vars: Default::default(),
    };
    let t = t.apply(&generalize);

    let vars = generalize.vars.into_inner();

    let mut cons = TvarKinds::new();
    for (tv, bound_tv) in &vars {
        if let Some(kinds) = sub.cons().get(tv) {
            cons.insert(*bound_tv, kinds.to_owned());
        }
    }
    PolyType {
        vars: vars.into_iter().map(|(_, tv)| tv).collect(),
        cons,
        expr: t,
    }
}

// Create a parametric type from a monotype by universally quantifying
// all of its free type variables.
//
// A type variable is free in a monotype if it is **free** in the global
// type environment. Equivalently a type variable is free and can be
// quantified if has not already been quantified another type in the
// type environment.
//
pub fn generalize(env: &Environment, sub: &mut Substitution, t: MonoType) -> PolyType {
    struct Generalize<'a> {
        env_free_vars: Vec<Tvar>,
        sub: &'a Substitution,
        vars: RefCell<Vec<(Tvar, Tvar)>>,
    }

    impl Substituter for Generalize<'_> {
        fn try_apply(&self, var: Tvar) -> Option<MonoType> {
            if !self.env_free_vars.contains(&var) {
                if (var.0 as usize) < self.sub.len() {
                    if let Some(new_var) = self.sub.try_apply(var) {
                        return Some(new_var);
                    }
                }

                let mut vars = self.vars.borrow_mut();
                let new_var = Tvar(vars.len() as u64);
                vars.push((var, new_var));
                let new_type = MonoType::BoundVar(new_var);
                if var.0 as usize > self.sub.len() {
                    self.sub.mk_fresh(var.0 as usize - self.sub.len() + 1);
                }
                self.sub.union_type(var, new_type.clone()).ok()?;
                Some(new_type)
            } else {
                None
            }
        }
    }

    let generalize = Generalize {
        env_free_vars: env.free_vars(),
        sub,
        vars: Default::default(),
    };
    let t = t.apply(&generalize);

    let vars = generalize.vars.into_inner();

    let mut cons = TvarKinds::new();
    for (tv, bound_tv) in &vars {
        if let Some(kinds) = sub.cons().get(tv) {
            cons.insert(*bound_tv, kinds.to_owned());
        }
    }
    PolyType {
        vars: vars.into_iter().map(|(_, tv)| tv).collect(),
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
    sub: &mut Substitution,
    loc: SourceLocation,
) -> (MonoType, Constraints) {
    // Substitute fresh type variables for all quantified variables
    let sub: SubstitutionMap = poly
        .vars
        .into_iter()
        .map(|tv| (tv, MonoType::Var(sub.fresh())))
        .collect();
    // Generate constraints for the new fresh type variables
    let constraints = poly
        .cons
        .into_iter()
        .fold(Constraints::empty(), |cons, (tv, kinds)| {
            cons + kinds
                .into_iter()
                .map(|kind| Constraint::Kind {
                    exp: kind,
                    act: sub.get(&tv).unwrap().clone(),
                    loc: loc.clone(),
                })
                .collect::<Vec<Constraint>>()
                .into()
        });

    // Equivalent to `SubstitutionMap` but instantiates bound variables instead of free variables
    struct InstantiationMap(SubstitutionMap);

    impl Substituter for InstantiationMap {
        fn try_apply(&self, _var: Tvar) -> Option<MonoType> {
            None
        }
        fn try_apply_bound(&self, var: Tvar) -> Option<MonoType> {
            self.0.get(&var).cloned()
        }
    }

    // Instantiate monotype using new fresh type variables
    (poly.expr.apply(&InstantiationMap(sub)), constraints)
}

// The following clippy lint is ignored due to taking a `Self` type as the
// first parameter in Environment::new, which clippy wrongly identifies as the
// parameter name, used in a place that shouldn't take a `self` parameter.
// See https://github.com/rust-lang/rust-clippy/issues/3414
#![allow(clippy::wrong_self_convention, clippy::unknown_clippy_lints)]

use crate::semantic::sub::{Substitutable, Substitution};
use crate::semantic::types::{union, PolyType, Tvar};
use std::collections::HashMap;

// A type environment maps program identifiers to their polymorphic types.
//
// A type environment is implemented as a stack of frames where each
// frame holds the bindings for the identifiers declared in a particular
// lexical block.
//
#[derive(Debug, Clone, PartialEq)]
pub struct Environment {
    pub parent: Option<Box<Environment>>,
    pub values: HashMap<String, PolyType>,
}

impl Substitutable for Environment {
    fn apply(self, sub: &Substitution) -> Self {
        Environment {
            parent: match self.parent {
                Some(env) => Some(Box::new(env.apply(sub))),
                None => None,
            },
            values: self.values.apply(sub),
        }
    }
    fn free_vars(&self) -> Vec<Tvar> {
        match &self.parent {
            Some(env) => union(env.free_vars(), self.values.free_vars()),
            None => self.values.free_vars(),
        }
    }
}

// Derive a type environment from a hash map
impl From<HashMap<String, PolyType>> for Environment {
    fn from(bindings: HashMap<String, PolyType>) -> Environment {
        Environment {
            parent: None,
            values: bindings,
        }
    }
}

impl Environment {
    pub fn empty() -> Environment {
        Environment {
            parent: None,
            values: HashMap::new(),
        }
    }
    pub fn new(from: Self) -> Environment {
        Environment {
            parent: Some(Box::new(from)),
            values: HashMap::new(),
        }
    }
    pub fn lookup(&self, v: &str) -> Option<&PolyType> {
        if let Some(t) = self.values.get(v) {
            Some(t)
        } else if let Some(env) = &self.parent {
            env.lookup(v)
        } else {
            None
        }
    }
    // Add a new variable binding to the current stack frame
    pub fn add(&mut self, name: String, t: PolyType) {
        self.values.insert(name, t);
    }
    // Remove a variable binding from the current stack frame
    pub fn remove(&mut self, name: &str) {
        self.values.remove(name);
    }
    // A type environment is a stack where each frame corresponds to a lexical
    // block inside a Flux program.
    //
    // After inferring the types in each lexical block, the frame at the top
    // of the stack will be popped, returning the type environment for the
    // enclosing block.
    //
    // Note that 'pop' must be paired with a corresponding call to 'new'. In
    // particular when inferring the type of a function expression, a new
    // frame must be added to the top of the stack by calling 'new'. Then the
    // bindings for the function arguments must be added to the new frame. At
    // that point the type of the function body is inferred and the last frame
    // is popped from the stack and returned to the calling function.
    //
    // It is invalid to call pop on a type environment with only one stack
    // frame. This will result in a panic.
    //
    pub fn pop(self) -> Environment {
        match self.parent {
            Some(env) => *env,
            None => panic!("cannot pop final stack frame from type environment"),
        }
    }
}

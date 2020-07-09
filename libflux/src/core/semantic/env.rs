use crate::semantic::import::Importer;
use crate::semantic::sub::{Substitutable, Substitution};
use crate::semantic::types::{union, PolyType, PolyTypeMap, Tvar};

// A type environment maps program identifiers to their polymorphic types.
//
// A type environment is implemented as a stack of frames where each
// frame holds the bindings for the identifiers declared in a particular
// lexical block.
//
#[derive(Debug, Clone, PartialEq)]
pub struct Environment {
    pub parent: Option<Box<Environment>>,
    pub values: PolyTypeMap,
    pub readwrite: bool,
}

impl Substitutable for Environment {
    fn apply(self, sub: &Substitution) -> Self {
        match (self.readwrite, self.parent) {
            // This is a performance optimization where false implies
            // this is the top-level of the type environment and apply
            // is a no-op.
            (false, None) => Environment {
                parent: None,
                values: self.values,
                readwrite: false,
            },
            (false, parent) => Environment {
                parent,
                values: self.values,
                readwrite: false,
            },
            // Even though this is the top-level of the type environment
            // and apply should be a no-op, readwrite is set to true so
            // we apply anyway.
            (true, None) => Environment {
                parent: None,
                values: self.values.apply(sub),
                readwrite: true,
            },
            (true, Some(env)) => Environment {
                parent: Some(Box::new(env.apply(sub))),
                values: self.values.apply(sub),
                readwrite: true,
            },
        }
    }
    fn free_vars(&self) -> Vec<Tvar> {
        match (self.readwrite, &self.parent) {
            (false, None) | (false, _) => Vec::new(),
            (true, None) => self.values.free_vars(),
            (true, Some(env)) => union(env.free_vars(), self.values.free_vars()),
        }
    }
}

impl Importer for Environment {
    fn import(&self, name: &str) -> Option<PolyType> {
        match self.lookup(name) {
            Some(pty) => Some(pty.clone()),
            None => None,
        }
    }
}

// Derive a type environment from a hash map
impl From<PolyTypeMap> for Environment {
    fn from(bindings: PolyTypeMap) -> Environment {
        Environment {
            parent: None,
            values: bindings,
            readwrite: false,
        }
    }
}

impl Environment {
    pub fn empty(readwrite: bool) -> Environment {
        Environment {
            parent: None,
            values: PolyTypeMap::new(),
            readwrite,
        }
    }
    // The following clippy lint is ignored due to taking a `Self` type as the
    // first parameter, which clippy wrongly identifies as the parameter name,
    // used in a place that shouldn't take a `self` parameter.
    // See https://github.com/rust-lang/rust-clippy/issues/3414
    #[allow(clippy::wrong_self_convention)]
    pub fn new(from: Self) -> Environment {
        Environment {
            parent: Some(Box::new(from)),
            values: PolyTypeMap::new(),
            readwrite: true,
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

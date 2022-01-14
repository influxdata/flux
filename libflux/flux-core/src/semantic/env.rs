//! Type environments.
use std::{fmt, mem};

use crate::semantic::{
    nodes::Symbol,
    sub::{apply2, Substitutable, Substituter},
    types::{PolyType, PolyTypeHashMap, PolyTypeMap, Tvar},
    PackageExports,
};

/// A type environment maps program identifiers to their polymorphic types.
///
/// Type environments are implemented as a stack where each
/// frame holds the bindings for the identifiers declared in a particular
/// lexical block.
#[derive(Debug, Clone, PartialEq)]
pub struct Environment<'a> {
    /// An external environment if one is provided
    pub external: Option<&'a PackageExports>,
    /// An optional parent environment.
    pub parent: Option<Box<Environment<'a>>>,
    /// Values in the environment.
    pub values: PolyTypeHashMap<Symbol>,
    /// Read/write permissions flag.
    pub readwrite: bool,
}

impl fmt::Display for Environment<'_> {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        let mut debug_map = f.debug_map();
        self.fmt_display(&mut debug_map);
        debug_map.finish()
    }
}

impl Substitutable for Environment<'_> {
    fn apply_ref(&self, sub: &dyn Substituter) -> Option<Self> {
        match (self.readwrite, &self.parent) {
            // This is a performance optimization where false implies
            // this is the top-level of the type environment and apply
            // is a no-op.
            (false, None) | (false, Some(_)) => None,
            // Even though this is the top-level of the type environment
            // and apply should be a no-op, readwrite is set to true so
            // we apply anyway.
            (true, None) => self.values.apply_ref(sub).map(|values| Environment {
                external: self.external,
                parent: None,
                values,
                readwrite: true,
            }),
            (true, Some(env)) => {
                apply2(&**env, &self.values, sub).map(|(parent, values)| Environment {
                    external: self.external,
                    parent: Some(Box::new(parent)),
                    values,
                    readwrite: true,
                })
            }
        }
    }
    fn free_vars(&self, vars: &mut Vec<Tvar>) {
        match (self.readwrite, &self.parent) {
            (false, None) | (false, _) => (),
            (true, None) => self.values.free_vars(vars),
            (true, Some(env)) => {
                env.free_vars(vars);
                self.values.free_vars(vars);
            }
        }
    }
}

// Derive a type environment from a hash map
impl From<PolyTypeHashMap<Symbol>> for Environment<'_> {
    fn from(bindings: PolyTypeHashMap<Symbol>) -> Self {
        Environment {
            external: None,
            parent: None,
            values: bindings,
            readwrite: false,
        }
    }
}

impl From<PolyTypeMap> for Environment<'_> {
    fn from(bindings: PolyTypeMap) -> Self {
        Environment::from(
            bindings
                .into_iter()
                .map(|(k, v)| (Symbol::from(k), v))
                .collect::<PolyTypeHashMap<Symbol>>(),
        )
    }
}

impl<'env> From<&'env PackageExports> for Environment<'env> {
    fn from(external: &'env PackageExports) -> Self {
        let mut env = Environment::empty(true);
        env.external = Some(external);
        env
    }
}

impl Default for Environment<'_> {
    fn default() -> Self {
        Environment::empty(false)
    }
}

impl Environment<'_> {
    /// Create an empty environment with no parent.
    pub fn empty(readwrite: bool) -> Self {
        Environment {
            external: None,
            parent: None,
            values: Default::default(),
            readwrite,
        }
    }

    /// Return a new environment from the current one.
    // The following clippy lint is ignored due to taking a `Self` type as the
    // first parameter, which clippy wrongly identifies as the parameter name,
    // used in a place that shouldn't take a `self` parameter.
    // See https://github.com/rust-lang/rust-clippy/issues/3414
    #[allow(clippy::wrong_self_convention)]
    pub fn new(from: Self) -> Self {
        Environment {
            external: None,
            parent: Some(Box::new(from)),
            values: Default::default(),
            readwrite: true,
        }
    }

    /// Return a new environment from the current one.
    pub fn enter_scope(&mut self) {
        let parent = mem::replace(self, Environment::empty(true));
        self.parent = Some(Box::new(parent));
    }

    /// Check whether a `PolyType` `t` given by a
    /// string identifier is in the environment. Also checks parent environments.
    /// If the type is present, returns a pointer to `t`; otherwise, returns `None`.
    pub fn lookup(&self, v: &Symbol) -> Option<&PolyType> {
        if let Some(t) = self.values.get(v) {
            Some(t)
        } else if let Some(t) = self.external.as_ref().and_then(|env| env.lookup(v)) {
            Some(t)
        } else if let Some(t) = self.parent.as_ref().and_then(|env| env.lookup(v)) {
            Some(t)
        } else {
            None
        }
    }

    /// Check whether a `PolyType` `t` given by a
    /// string identifier is in the environment. Also checks parent environments.
    /// If the type is present, returns a pointer to `t`; otherwise, returns `None`.
    pub fn lookup_str(&self, v: &str) -> Option<&PolyType> {
        // TODO Avoid iteration here
        if let Some((_, t)) = self
            .values
            .iter_by(|l, r| l.name().cmp(r.name()))
            .find(|(symbol, _)| *symbol == v)
        {
            Some(t)
        } else if let Some(t) = self.external.as_ref().and_then(|env| env.lookup(v)) {
            Some(t)
        } else if let Some(t) = self.parent.as_ref().and_then(|env| env.lookup_str(v)) {
            Some(t)
        } else {
            None
        }
    }

    pub(crate) fn lookup_symbol(&self, v: &str) -> Option<&Symbol> {
        // TODO Avoid iteration here
        if let Some(t) = self
            .values
            .keys_by(|l, r| l.name().cmp(r.name()))
            .find(|symbol| *symbol == v)
        {
            Some(t)
        } else if let Some(env) = &self.parent {
            env.lookup_symbol(v)
        } else {
            None
        }
    }

    /// Add a new variable binding to the current stack frame.
    pub fn add(&mut self, name: Symbol, t: PolyType) {
        self.values.insert(name, t);
    }

    /// Remove a variable binding from the current stack frame.
    pub fn remove(&mut self, name: &Symbol) {
        self.values.remove(name);
    }

    pub(crate) fn exit_scope(&mut self) -> Self {
        match self.parent.take() {
            Some(mut env) => {
                mem::swap(self, &mut env);
                *env
            }
            None => panic!("cannot pop final stack frame from type environment"),
        }
    }

    /// Copy all the variable bindings from another [`Environment`] to the current environment.
    /// This does not change the current environment's `parent` or `readwrite` flag.
    #[cfg(test)]
    pub fn copy_bindings_from(&mut self, other: &Environment) {
        for (name, t) in other.values.iter_by(|l, r| l.name().cmp(r.name())) {
            self.add(name.clone(), t.clone());
        }
    }

    /// Returns a string keyed `PolyTypeMap`
    pub fn string_values(self) -> PolyTypeMap {
        self.values
            .into_iter_by(|l, r| l.name().cmp(r.name()))
            .map(|(k, v)| (k.to_string(), v))
            .collect()
    }

    fn fmt_display(&self, f: &mut fmt::DebugMap<'_, '_>) {
        f.entries(
            self.values
                .iter_by(|l, r| l.name().cmp(r.name()))
                .map(|(k, v)| (k, v.to_string())),
        );
        if let Some(parent) = &self.parent {
            parent.fmt_display(f);
        }
        if let Some(external) = &self.external {
            f.entries(external.iter().map(|(k, v)| (k, v.to_string())));
        }
    }
}

use crate::ast::{walk, PropertyKey, SourceLocation};
use std::fmt;

// check() inspects an AST node and returns a list of found AST errors plus
// any errors existed before ast.check() is performed.
pub fn check(node: walk::Node) -> Vec<Error> {
    let mut errors = vec![];
    walk::walk(
        &walk::create_visitor(&mut |n| {
            // collect any errors we found prior to ast.check().
            for err in n.base().errors.iter() {
                errors.push(Error {
                    location: n.base().location.clone(),
                    message: err.clone(),
                });
            }

            match *n {
                walk::Node::BadStmt(n) => errors.push(Error {
                    location: n.base.location.clone(),
                    message: format!("invalid statement: {}", n.text),
                }),
                walk::Node::ObjectExpr(n) => {
                    let mut has_implicit = false;
                    let mut has_explicit = false;
                    for p in n.properties.iter() {
                        if p.base.errors.is_empty() {
                            match p.value {
                                None => {
                                    has_implicit = true;
                                    if let PropertyKey::StringLit(s) = &p.key {
                                        errors.push(Error {
                                            location: n.base.location.clone(),
                                            message: format!(
                                                "string literal key {} must have a value",
                                                s.value
                                            ),
                                        })
                                    }
                                }
                                Some(_) => {
                                    has_explicit = true;
                                }
                            }
                        }
                    }
                    if has_implicit && has_explicit {
                        errors.push(Error {
                            location: n.base.location.clone(),
                            message: String::from("cannot mix implicit and explicit properties"),
                        })
                    }
                }
                _ => {}
            }
        }),
        node,
    );
    errors
}

#[derive(Debug, PartialEq)] // derive std::fmt::Debug on AppError
pub struct Error {
    pub location: SourceLocation,
    pub message: String,
}

impl fmt::Display for Error {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "error at {}: {}", self.location, self.message)
    }
}

#[cfg(test)]
mod tests;

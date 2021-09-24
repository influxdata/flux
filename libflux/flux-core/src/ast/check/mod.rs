//! Checking the AST.

use std::fmt;
use thiserror::Error;

use crate::ast::{walk, PropertyKey, SourceLocation};

/// Inspects an AST node and returns a list of found AST errors plus
/// any errors existed before `ast.check()` is performed.
pub fn check(node: walk::Node) -> Result<(), Errors> {
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
                walk::Node::BadExpr(n) => errors.push(Error {
                    location: n.base.location.clone(),
                    message: format!("invalid expression: {}", n.text),
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
    if errors.is_empty() {
        Ok(())
    } else {
        Err(Errors(errors))
    }
}

/// An error that can be returned while checking the AST.
#[derive(Error, Debug, PartialEq)]
#[error("error at {}: {}", location, message)]
pub struct Error {
    /// Location of the error in source code.
    pub location: SourceLocation,
    /// Error message.
    pub message: String,
}

/// Set of any errors encountered when checking the AST.
#[derive(Debug, PartialEq)]
pub struct Errors(pub Vec<Error>);

impl fmt::Display for Errors {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(
            f,
            "{}",
            self.0
                .iter()
                .map(|err| err.to_string())
                .collect::<Vec<_>>()
                .join("\n")
        )
    }
}
impl std::error::Error for Errors {}

#[cfg(test)]
mod tests;

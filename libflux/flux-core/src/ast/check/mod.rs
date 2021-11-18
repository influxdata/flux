//! Checking the AST.

use codespan_reporting::diagnostic;
use thiserror::Error;

use crate::{
    ast::{walk, PropertyKey},
    errors::{located, Errors, Located},
};

/// Inspects an AST node and returns a list of found AST errors plus
/// any errors existed before `ast.check()` is performed.
pub fn check(node: walk::Node) -> Result<(), Errors<Error>> {
    let mut errors = Errors::new();
    walk::walk(
        &mut |n: walk::Node| {
            // collect any errors we found prior to ast.check().
            for err in n.base().errors.iter() {
                errors.push(located(
                    n.base().location.clone(),
                    ErrorKind {
                        message: err.clone(),
                    },
                ));
            }

            match n {
                walk::Node::BadStmt(n) => errors.push(located(
                    n.base.location.clone(),
                    ErrorKind {
                        message: format!("invalid statement: {}", n.text),
                    },
                )),
                walk::Node::BadExpr(n) if !n.text.is_empty() => errors.push(located(
                    n.base.location.clone(),
                    ErrorKind {
                        message: format!("invalid expression: {}", n.text),
                    },
                )),
                walk::Node::ObjectExpr(n) => {
                    let mut has_implicit = false;
                    let mut has_explicit = false;
                    for p in n.properties.iter() {
                        if p.base.errors.is_empty() {
                            match p.value {
                                None => {
                                    has_implicit = true;
                                    if let PropertyKey::StringLit(s) = &p.key {
                                        errors.push(located(
                                            n.base.location.clone(),
                                            ErrorKind {
                                                message: format!(
                                                    "string literal key {} must have a value",
                                                    s.value
                                                ),
                                            },
                                        ))
                                    }
                                }
                                Some(_) => {
                                    has_explicit = true;
                                }
                            }
                        }
                    }
                    if has_implicit && has_explicit {
                        errors.push(located(
                            n.base.location.clone(),
                            ErrorKind {
                                message: String::from(
                                    "cannot mix implicit and explicit properties",
                                ),
                            },
                        ))
                    }
                }
                _ => {}
            }
        },
        node,
    );
    if errors.is_empty() {
        Ok(())
    } else {
        Err(errors)
    }
}

/// An error that can be returned while checking the AST.
pub type Error = Located<ErrorKind>;

/// An error that can be returned while checking the AST.
#[derive(Error, Debug, PartialEq)]
#[error("{}", message)]
pub struct ErrorKind {
    /// Error message.
    pub message: String,
}

impl Error {
    pub(crate) fn as_diagnostic(
        &self,
        source: &dyn crate::semantic::Source,
    ) -> diagnostic::Diagnostic<()> {
        diagnostic::Diagnostic::error()
            .with_message(self.error.to_string())
            .with_labels(vec![diagnostic::Label::primary(
                (),
                source.codespan_range(&self.location),
            )])
    }
}

#[cfg(test)]
mod tests;

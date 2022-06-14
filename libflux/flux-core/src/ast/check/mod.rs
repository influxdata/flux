//! Checking the AST.

use codespan_reporting::diagnostic;
use thiserror::Error;

use crate::{
    ast::{walk, PropertyKey},
    errors::{located, AsDiagnostic, Errors, Located},
};

/// Inspects an AST node and returns a list of found AST errors plus
/// any errors existed before `ast.check()` is performed.
pub fn check(node: walk::Node) -> Result<(), Errors<Error>> {
    const MAX_DEPTH: u32 = 1000;

    #[derive(Default)]
    struct Check {
        depth: u32,
        errors: Errors<Error>,
    }

    impl<'a> walk::Visitor<'a> for Check {
        fn visit(&mut self, n: walk::Node<'a>) -> bool {
            self.depth += 1;

            let errors = &mut self.errors;

            if self.depth > MAX_DEPTH {
                errors.push(located(n.base().location.clone(), ErrorKind::NestedToDeep));

                return false;
            }

            // collect any errors we found prior to ast.check().
            for err in n.base().errors.iter() {
                errors.push(located(
                    n.base().location.clone(),
                    ErrorKind::Message {
                        message: err.clone(),
                    },
                ));
            }

            match n {
                walk::Node::BadStmt(n) => errors.push(located(
                    n.base.location.clone(),
                    ErrorKind::Message {
                        message: format!("invalid statement: {}", n.text),
                    },
                )),
                walk::Node::BadExpr(n) if !n.text.is_empty() => errors.push(located(
                    n.base.location.clone(),
                    ErrorKind::Message {
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
                                            ErrorKind::Message {
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
                            ErrorKind::Message {
                                message: String::from(
                                    "cannot mix implicit and explicit properties",
                                ),
                            },
                        ))
                    }
                }
                _ => {}
            }

            true
        }

        fn done(&mut self, _: walk::Node<'a>) {
            self.depth -= 1;
        }
    }

    let mut check = Check::default();
    walk::walk(&mut check, node);
    let errors = check.errors;

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
#[allow(missing_docs)]
pub enum ErrorKind {
    #[error("Program is nested to deep")]
    NestedToDeep,
    #[error("{message}")]
    Message { message: String },
}

impl ErrorKind {
    pub(crate) fn is_fatal(&self) -> bool {
        matches!(self, Self::NestedToDeep)
    }
}

impl AsDiagnostic for ErrorKind {
    fn as_diagnostic(&self, _source: &dyn crate::semantic::Source) -> diagnostic::Diagnostic<()> {
        diagnostic::Diagnostic::error().with_message(self.to_string())
    }
}

#[cfg(test)]
mod tests;

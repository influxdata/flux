use crate::ast::walk;
use crate::ast::SourceLocation;
use std::fmt;

// Check counts
pub fn check(node: walk::Node) -> i32 {
    let mut count = 0;
    walk::walk(
        &walk::create_visitor(&mut |n| {
            if has_error(&n) {
                count += 1
            }
        }),
        node,
    );
    count
}

pub fn error(node: walk::Node) -> Option<Error> {
    None
}

fn has_error(node: &walk::Node) -> bool {
    node.base().errors.len() > 0
}

#[derive(Debug)] // derive std::fmt::Debug on AppError
pub struct Error {
    location: SourceLocation,
    message: String,
}

impl fmt::Display for Error {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "error at {}: {}", self.location, self.message)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    #[test]
    fn test_has_error() {
        assert_eq!(true, false);
    }
}

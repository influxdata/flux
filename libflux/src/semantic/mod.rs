mod analyze;
pub use analyze::analyze;

mod env;
mod fresh;
mod infer;

// TODO(jsternberg): Once more work is done on the infer methods,
// this should be removed and the warnings fixed.
#[allow(warnings)]
pub mod nodes;

mod sub;
mod types;
pub mod walk;

#[cfg(test)]
mod parser;

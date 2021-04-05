#![cfg_attr(feature = "strict", deny(warnings, missing_docs))]

//! The flux crate handles the parsing and semantic analysis of flux source
//! code.
extern crate chrono;
extern crate derive_more;
extern crate fnv;
#[macro_use]
extern crate serde_derive;
extern crate serde_aux;

#[allow(clippy::unnecessary_wraps)]
pub mod ast;
pub mod formatter;
pub mod parser;
pub mod scanner;
#[allow(clippy::unnecessary_wraps)]
pub mod semantic;

use std::error;
use std::hash::BuildHasherDefault;

use derive_more::Display;
use fnv::FnvHasher;

pub use ast::DEFAULT_PACKAGE_NAME;

type DefaultHasher = BuildHasherDefault<FnvHasher>;

/// An error that can occur due to problems in ast generation or semantic
/// analysis.
#[derive(Debug, Display, Clone)]
#[display(fmt = "{}", msg)]
pub struct Error {
    /// Message.
    pub msg: String,
}

impl error::Error for Error {
    fn source(&self) -> Option<&(dyn error::Error + 'static)> {
        None
    }
}

impl From<String> for Error {
    fn from(msg: String) -> Self {
        Error { msg }
    }
}

impl From<&str> for Error {
    fn from(msg: &str) -> Self {
        Error {
            msg: String::from(msg),
        }
    }
}

impl From<semantic::nodes::Error> for Error {
    fn from(sn_err: semantic::nodes::Error) -> Self {
        Error {
            msg: sn_err.to_string(),
        }
    }
}

impl From<semantic::check::Error> for Error {
    fn from(err: semantic::check::Error) -> Self {
        Error {
            msg: format!("{}", err),
        }
    }
}

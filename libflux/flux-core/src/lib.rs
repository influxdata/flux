#![cfg_attr(feature = "strict", deny(warnings, missing_docs))]

//! This crate performs parsing and semantic analysis of Flux source
//! code. It forms the core of the compiler for the [Flux language].
//! It is made up of five modules. Four of these handle the analysis
//! of Flux code during compilation:
//!
//! - [`scanner`] produces tokens from plain source code;
//! - [`parser`] forms the abstract syntax tree (AST);
//! - [`ast`] defines the AST data structures and provides functions for its analysis; and
//! - [`semantic`] performs semantic analysis, including type inference,
//!   producing a semantic graph.
//!
//! In addition, the [`formatter`] module provides functions for code formatting utilities.
//!
//! [Flux language]: https://github.com/influxdata/flux

extern crate chrono;
extern crate derive_more;
extern crate fnv;
#[macro_use]
extern crate serde_derive;
extern crate serde_aux;

pub mod ast;
pub mod formatter;
pub mod parser;
pub mod scanner;
pub mod semantic;

use std::error;
use std::hash::BuildHasherDefault;

use derive_more::Display;
use fnv::FnvHasher;

pub use ast::DEFAULT_PACKAGE_NAME;

type DefaultHasher = BuildHasherDefault<FnvHasher>;

/// An error that can occur due to problems in AST generation or semantic
/// analysis.
#[derive(Debug, Display, Clone)]
#[display(fmt = "{}", msg)]
pub struct Error {
    /// Contents of the error message.
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

/// merge_packages takes an input package and an output package, checks that the package
/// clauses match and merges the files from the input package into the output package. If
/// package clauses fail validation then an option with an Error is returned.
pub fn merge_packages(out_pkg: &mut ast::Package, in_pkg: &mut ast::Package) -> Option<Error> {
    let out_pkg_name = if let Some(pc) = &out_pkg.files[0].package {
        &pc.name.name
    } else {
        DEFAULT_PACKAGE_NAME
    };

    // Check that all input files have a package clause that matches the output package.
    for file in &in_pkg.files {
        match file.package.as_ref() {
            Some(pc) => {
                let in_pkg_name = &pc.name.name;
                if in_pkg_name != out_pkg_name {
                    return Some(Error::from(format!(
                        r#"error at {}: file is in package "{}", but other files are in package "{}""#,
                        pc.base.location, in_pkg_name, out_pkg_name
                    )));
                }
            }
            None => {
                if out_pkg_name != DEFAULT_PACKAGE_NAME {
                    return Some(Error::from(format!(
                        r#"error at {}: file is in default package "{}", but other files are in package "{}""#,
                        file.base.location, DEFAULT_PACKAGE_NAME, out_pkg_name
                    )));
                }
            }
        };
    }
    out_pkg.files.append(&mut in_pkg.files);
    None
}

#![cfg_attr(feature = "strict", deny(warnings, missing_docs))]
#![allow(
    clippy::needless_update, //
    clippy::identity_op, // Warns on `1 * YEARS` which seems clearer than `YEARS`
    clippy::neg_multiply, // Warns on `-1 * YEARS` which seems clearer than `-YEARS`
)]

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

#[macro_use]
extern crate serde_derive;

#[macro_use]
#[cfg(test)]
extern crate pretty_assertions;

// Only include the doc module if the feature is enabled.
// The code has lots of dependencies we do not want as part of the crate by default.
#[cfg(feature = "doc")]
pub mod doc;

pub mod ast;
pub mod errors;
pub mod formatter;
pub mod parser;
pub mod scanner;
pub mod semantic;

mod db;
mod map;

use std::hash::BuildHasherDefault;

use anyhow::{bail, Result};
pub use ast::DEFAULT_PACKAGE_NAME;
use fnv::FnvHasher;

type DefaultHasher = BuildHasherDefault<FnvHasher>;

/// merge_packages takes an input package and an output package, checks that the package
/// clauses match and merges the files from the input package into the output package. If
/// package clauses fail validation then an option with an Error is returned.
pub fn merge_packages(out_pkg: &mut ast::Package, in_pkg: &mut ast::Package) -> Result<()> {
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
                    bail!(
                        r#"error at {}: file is in package "{}", but other files are in package "{}""#,
                        pc.base.location,
                        in_pkg_name,
                        out_pkg_name
                    );
                }
            }
            None => {
                if out_pkg_name != DEFAULT_PACKAGE_NAME {
                    bail!(
                        r#"error at {}: file is in default package "{}", but other files are in package "{}""#,
                        file.base.location,
                        DEFAULT_PACKAGE_NAME,
                        out_pkg_name
                    );
                }
            }
        };
    }
    out_pkg.files.append(&mut in_pkg.files);
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::merge_packages;
    use crate::ast;

    #[test]
    fn ok_merge_multi_file() {
        let in_script = "package foo\na = 1\n";
        let out_script = "package foo\nb = 2\n";

        let in_file = crate::parser::parse_string("test".to_string(), in_script);
        let out_file = crate::parser::parse_string("test".to_string(), out_script);
        let mut in_pkg = ast::Package {
            base: Default::default(),
            path: "./test".to_string(),
            package: "foo".to_string(),
            files: vec![in_file.clone()],
        };
        let mut out_pkg = ast::Package {
            base: Default::default(),
            path: "./test".to_string(),
            package: "foo".to_string(),
            files: vec![out_file.clone()],
        };
        merge_packages(&mut out_pkg, &mut in_pkg).unwrap();
        let got = out_pkg.files;
        let want = vec![out_file, in_file];
        assert_eq!(want, got);
    }

    #[test]
    fn ok_merge_one_default_pkg() {
        // Make sure we can merge one file with default "main"
        // and on explicit
        let has_clause_script = "package main\nx = 32";
        let no_clause_script = "y = 32";
        let has_clause_file =
            crate::parser::parse_string("has_clause.flux".to_string(), has_clause_script);
        let no_clause_file =
            crate::parser::parse_string("no_clause.flux".to_string(), no_clause_script);
        {
            let mut out_pkg: ast::Package = has_clause_file.clone().into();
            let mut in_pkg: ast::Package = no_clause_file.clone().into();
            merge_packages(&mut out_pkg, &mut in_pkg).unwrap();
            let got = out_pkg.files;
            let want = vec![has_clause_file.clone(), no_clause_file.clone()];
            assert_eq!(want, got);
        }
        {
            // Same as previous test, but reverse order
            let mut out_pkg: ast::Package = no_clause_file.clone().into();
            let mut in_pkg: ast::Package = has_clause_file.clone().into();
            merge_packages(&mut out_pkg, &mut in_pkg).unwrap();
            let got = out_pkg.files;
            let want = vec![no_clause_file, has_clause_file];
            assert_eq!(want, got);
        }
    }

    #[test]
    fn ok_no_in_pkg() {
        let out_script = "package foo\nb = 2\n";

        let out_file = crate::parser::parse_string("test".to_string(), out_script);
        let mut in_pkg = ast::Package {
            base: Default::default(),
            path: "./test".to_string(),
            package: "foo".to_string(),
            files: vec![],
        };
        let mut out_pkg = ast::Package {
            base: Default::default(),
            path: "./test".to_string(),
            package: "foo".to_string(),
            files: vec![out_file.clone()],
        };
        merge_packages(&mut out_pkg, &mut in_pkg).unwrap();
        let got = out_pkg.files;
        let want = vec![out_file];
        assert_eq!(want, got);
    }

    #[test]
    fn err_no_out_pkg_clause() {
        let in_script = "package foo\na = 1\n";
        let out_script = "";

        let in_file = crate::parser::parse_string("test_in.flux".to_string(), in_script);
        let out_file = crate::parser::parse_string("test_out.flux".to_string(), out_script);
        let mut in_pkg = ast::Package {
            base: Default::default(),
            path: "./test".to_string(),
            package: "foo".to_string(),
            files: vec![in_file],
        };
        let mut out_pkg = ast::Package {
            base: Default::default(),
            path: "./test".to_string(),
            package: "foo".to_string(),
            files: vec![out_file],
        };
        let got_err = merge_packages(&mut out_pkg, &mut in_pkg)
            .unwrap_err()
            .to_string();
        let want_err = r#"error at test_in.flux@1:1-1:12: file is in package "foo", but other files are in package "main""#;
        assert_eq!(got_err, want_err);
    }

    #[test]
    fn err_no_in_pkg_clause() {
        let in_script = "a = 1000\n";
        let out_script = "package foo\nb = 100\n";

        let in_file = crate::parser::parse_string("test_in.flux".to_string(), in_script);
        let out_file = crate::parser::parse_string("test_out.flux".to_string(), out_script);
        let mut in_pkg = ast::Package {
            base: Default::default(),
            path: "./test".to_string(),
            package: "foo".to_string(),
            files: vec![in_file],
        };
        let mut out_pkg = ast::Package {
            base: Default::default(),
            path: "./test".to_string(),
            package: "foo".to_string(),
            files: vec![out_file],
        };
        let got_err = merge_packages(&mut out_pkg, &mut in_pkg)
            .unwrap_err()
            .to_string();
        let want_err = r#"error at test_in.flux@1:1-1:9: file is in default package "main", but other files are in package "foo""#;
        assert_eq!(got_err, want_err);
    }

    #[test]
    fn ok_no_pkg_clauses() {
        let in_script = "a = 100\n";
        let out_script = "b = a * a\n";
        let in_file = crate::parser::parse_string("test".to_string(), in_script);
        let out_file = crate::parser::parse_string("test".to_string(), out_script);
        let mut in_pkg = ast::Package {
            base: Default::default(),
            path: "./test".to_string(),
            package: "foo".to_string(),
            files: vec![in_file],
        };
        let mut out_pkg = ast::Package {
            base: Default::default(),
            path: "./test".to_string(),
            package: "foo".to_string(),
            files: vec![out_file],
        };
        merge_packages(&mut out_pkg, &mut in_pkg).unwrap();
        assert_eq!(2, out_pkg.files.len());
    }
}

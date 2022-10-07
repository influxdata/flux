#![cfg_attr(feature = "strict", deny(warnings, missing_docs))]

//! This module provides the public facing API for Flux's Go runtime, including formatting,
//! parsing, and standard library analysis.
use std::sync::Arc;

use anyhow::anyhow;
use fluxcore::semantic::env::Environment;
use fluxcore::semantic::flatbuffers::semantic_generated::fbsemantic as fb;
use fluxcore::semantic::import::Packages;
use fluxcore::semantic::{Analyzer, AnalyzerConfig, PackageExports};
use fluxcore::{Database, Flux};
use once_cell::sync::Lazy;
use thiserror::Error;

pub use fluxcore::{ast, formatter, semantic, *};

#[macro_use]
#[cfg(test)]
extern crate pretty_assertions;

#[cfg(feature = "cffi")]
mod cffi;
#[cfg(feature = "cffi")]
pub use cffi::*;

/// Error type for flux
#[derive(Error, Debug)]
pub enum Error {
    /// Semantic error
    #[error(transparent)]
    Semantic(#[from] semantic::FileErrors),

    /// Options error
    #[error("Invalid compilation options: {0}")]
    InvalidOptions(String),

    /// Other errors that do not have a dedicated variant
    #[error(transparent)]
    Other(#[from] anyhow::Error),
}

/// Result type for flux
pub type Result<T, E = Error> = std::result::Result<T, E>;

/// Prelude are the names and types of values that are inscope in all Flux scripts.
pub fn prelude() -> Option<PackageExports> {
    let _ = env_logger::try_init();

    let buf = include_bytes!(concat!(env!("OUT_DIR"), "/prelude.data"));

    flatbuffers::root::<fb::TypeEnvironment>(buf)
        .unwrap_or_else(|err| panic!("{}", err))
        .into()
}

static PRELUDE: Lazy<Option<Arc<PackageExports>>> = Lazy::new(|| prelude().map(Arc::new));

/// Imports is a map of import path to types of packages.
pub fn imports() -> Option<Packages> {
    let _ = env_logger::try_init();

    let buf = include_bytes!(concat!(env!("OUT_DIR"), "/stdlib.data"));
    flatbuffers::root::<fb::Packages>(buf)
        .unwrap_or_else(|err| panic!("{}", err))
        .into()
}

static IMPORTS: Lazy<Option<Packages>> = Lazy::new(imports);

/// Creates a new analyzer that can semantically analyze Flux source code.
///
/// The analyzer is aware of the stdlib and prelude.
pub fn new_semantic_analyzer(config: AnalyzerConfig) -> Result<Analyzer<'static, Database>> {
    let env = PRELUDE.as_ref().ok_or_else(|| anyhow!("missing prelude"))?;

    let db = new_db()?;

    Ok(Analyzer::new(Environment::from(&**env), db, config))
}

fn new_db() -> Result<Database> {
    let mut db = fluxcore::Database::default();

    let imports = IMPORTS
        .as_ref()
        .ok_or_else(|| anyhow!("missing stdlib imports"))?;
    db.set_precompiled_packages(Some(&imports));

    Ok(db)
}

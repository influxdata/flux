#![cfg_attr(feature = "strict", deny(warnings, missing_docs))]

//! This module provides the public facing API for Flux's Go runtime, including formatting,
//! parsing, and standard library analysis.

use anyhow::anyhow;
use fluxcore::semantic::env::Environment;
use fluxcore::semantic::flatbuffers::semantic_generated::fbsemantic as fb;
use fluxcore::semantic::import::Packages;
use fluxcore::semantic::{Analyzer, AnalyzerConfig, PackageExports};
use once_cell::sync::Lazy;
use thiserror::Error;

pub use fluxcore::{ast, formatter, semantic, *};

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

    /// Other errors that do not have a dedicated variant
    #[error(transparent)]
    Other(#[from] anyhow::Error),
}

/// Result type for flux
pub type Result<T, E = Error> = std::result::Result<T, E>;

/// Prelude are the names and types of values that are inscope in all Flux scripts.
pub fn prelude() -> Option<PackageExports> {
    let buf = include_bytes!(concat!(env!("OUT_DIR"), "/prelude.data"));
    flatbuffers::root::<fb::TypeEnvironment>(buf)
        .unwrap()
        .into()
}

static PRELUDE: Lazy<Option<PackageExports>> = Lazy::new(prelude);

/// Imports is a map of import path to types of packages.
pub fn imports() -> Option<Packages> {
    let buf = include_bytes!(concat!(env!("OUT_DIR"), "/stdlib.data"));
    flatbuffers::root::<fb::Packages>(buf).unwrap().into()
}

static IMPORTS: Lazy<Option<Packages>> = Lazy::new(imports);

/// Creates a new analyzer that can semantically analyze Flux source code.
///
/// The analyzer is aware of the stdlib and prelude.
pub fn new_semantic_analyzer(
    config: AnalyzerConfig,
) -> Result<Analyzer<'static, &'static Packages>> {
    let env = PRELUDE.as_ref().ok_or_else(|| anyhow!("missing prelude"))?;

    let importer = IMPORTS
        .as_ref()
        .ok_or_else(|| anyhow!("missing stdlib imports"))?;

    Ok(Analyzer::new(Environment::from(env), importer, config))
}

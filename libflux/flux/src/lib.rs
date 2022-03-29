#![cfg_attr(feature = "strict", deny(warnings, missing_docs))]

//! This module provides the public facing API for Flux's Go runtime, including formatting,
//! parsing, and standard library analysis.

extern crate fluxcore;
extern crate serde_aux;

extern crate serde_derive;

#[cfg(test)]
#[macro_use]
extern crate pretty_assertions;

pub use fluxcore::{ast, formatter, semantic, *};

#[cfg(feature = "cffi")]
mod cffi;
#[cfg(feature = "cffi")]
pub use cffi::*;

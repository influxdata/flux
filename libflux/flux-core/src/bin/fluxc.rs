use std::path::{Path, PathBuf};

use anyhow::Result;
use clap::Parser;
use fluxcore::semantic::bootstrap;

#[derive(Debug, Parser)]
#[clap(author, version, about = "compile the Flux source code", long_about = None)]
enum FluxC {
    /// Dump JSON encoding of documentation from Flux source code.
    Stdlib {
        /// Directory containing Flux source code.
        #[clap(short, long, parse(from_os_str))]
        srcdir: PathBuf,
        /// Output directory for compiled Flux files.
        #[clap(short, long, parse(from_os_str))]
        outdir: PathBuf,
    },
}

fn main() -> Result<()> {
    let app = FluxC::parse();
    match app {
        FluxC::Stdlib { srcdir, outdir } => stdlib(&srcdir, &outdir)?,
    };
    Ok(())
}

fn stdlib(srcdir: &Path, outdir: &Path) -> Result<()> {
    bootstrap::compile_stdlib(srcdir, outdir)?;
    Ok(())
}

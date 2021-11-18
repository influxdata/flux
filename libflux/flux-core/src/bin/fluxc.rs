use std::path::{Path, PathBuf};

use anyhow::Result;
use fluxcore::semantic::bootstrap;
use structopt::StructOpt;

#[derive(Debug, StructOpt)]
#[structopt(about = "compile the Flux source code")]
enum FluxC {
    /// Dump JSON encoding of documentation from Flux source code.
    Stdlib {
        /// Directory containing Flux source code.
        #[structopt(short, long, parse(from_os_str))]
        srcdir: PathBuf,
        /// Output directory for compiled Flux files.
        #[structopt(short, long, parse(from_os_str))]
        outdir: PathBuf,
    },
}

fn main() -> Result<()> {
    let app = FluxC::from_args();
    match app {
        FluxC::Stdlib { srcdir, outdir } => stdlib(&srcdir, &outdir)?,
    };
    Ok(())
}

fn stdlib(srcdir: &Path, outdir: &Path) -> Result<()> {
    bootstrap::compile_stdlib(srcdir, outdir)?;
    Ok(())
}

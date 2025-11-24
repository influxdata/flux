use std::{
    collections::BinaryHeap,
    fs::File,
    io::{self, Write},
    path::{Path, PathBuf},
    process::Command,
};

use anyhow::{bail, Context, Result};
use clap::{Parser, Subcommand};
use rayon::prelude::*;

use fluxcore::{
    doc::{self, example},
    semantic::bootstrap,
    DatabaseBuilder, Flux, FluxBase,
};

/// Generate and validate Flux source code documentation
#[derive(Debug, Parser)]
#[command(about = "generate and validate Flux source code documentation")]
struct Cli {
    #[command(subcommand)]
    command: FluxDoc,
}

#[derive(Debug, Subcommand)]
enum FluxDoc {
    /// Dump JSON encoding of documentation from Flux source code
    Dump {
        /// Path to flux command, must be the cmd from internal/cmd/flux
        #[arg(long, value_name = "PATH")]
        flux_cmd_path: Option<PathBuf>,
        /// Directory containing Flux source code
        #[arg(short, long, value_name = "DIR")]
        dir: PathBuf,
        /// Output file, stdout if not present
        #[arg(short, long, value_name = "FILE")]
        output: Option<PathBuf>,
        /// Whether to structure the documentation as nested packages
        #[arg(short, long)]
        nested: bool,
        /// Whether to omit full descriptions and keep only the short form docs
        #[arg(long)]
        short: bool,
    },
    /// Check Flux source code for documentation linting errors
    Lint {
        /// Path to flux command, must be the cmd from internal/cmd/flux
        #[arg(long, value_name = "PATH")]
        flux_cmd_path: Option<PathBuf>,
        /// Directory containing Flux source code
        #[arg(short, long, value_name = "DIR")]
        dir: PathBuf,
        /// Limit the number of diagnostics to report. Default 10. 0 means no limit
        #[arg(short, long)]
        limit: Option<i64>,
    },
}

fn main() -> Result<()> {
    env_logger::init();

    let cli = Cli::parse();
    match cli.command {
        FluxDoc::Dump {
            flux_cmd_path,
            dir,
            output,
            nested,
            short,
        } => dump(
            flux_cmd_path.as_deref(),
            &dir,
            output.as_deref(),
            nested,
            short,
        )?,
        FluxDoc::Lint {
            flux_cmd_path,
            dir,
            limit,
        } => lint(flux_cmd_path.as_deref(), &dir, limit)?,
    };
    Ok(())
}

const DEFAULT_FLUX_CMD_PATH: &str = "flux";

fn resolve_default_paths(flux_cmd_path: Option<&Path>) -> &Path {
    let flux_cmd_path = match flux_cmd_path {
        Some(flux_cmd_path) => flux_cmd_path,
        None => Path::new(DEFAULT_FLUX_CMD_PATH),
    };
    flux_cmd_path
}

fn dump(
    flux_cmd_path: Option<&Path>,
    dir: &Path,
    output: Option<&Path>,
    nested: bool,
    short: bool,
) -> Result<()> {
    let flux_cmd_path = resolve_default_paths(flux_cmd_path);
    let f = match output {
        Some(p) => Box::new(File::create(p).context(format!("creating output file {:?}", p))?)
            as Box<dyn io::Write>,
        None => Box::new(io::stdout()),
    };

    let (mut docs, diagnostics) = parse_docs(dir).context("parsing source code")?;
    if !diagnostics.is_empty() {
        bail!(
            "found {} diagnostics when building documentation:\n{}",
            diagnostics.len(),
            diagnostics
                .iter()
                .map(|d| format!("{}", d))
                .collect::<Vec<String>>()
                .join("\n"),
        );
    }
    if short {
        for d in docs.iter_mut() {
            doc::shorten(d);
        }
    } else {
        // Evaluate examples only if not in short mode as shorten removes them.
        let executor = CLIExecutor {
            path: flux_cmd_path,
        };
        for d in docs.iter_mut() {
            for result in example::evaluate_package_examples(d, &executor) {
                result?;
            }
        }
    }
    if nested {
        let nested_docs = doc::nest_docs(docs);
        serde_json::to_writer(f, &nested_docs).context("encoding nested json")?;
    } else {
        serde_json::to_writer(f, &docs).context("encoding json")?;
    }

    Ok(())
}

fn lint(flux_cmd_path: Option<&Path>, dir: &Path, limit: Option<i64>) -> Result<()> {
    let flux_cmd_path = resolve_default_paths(flux_cmd_path);
    let limit = match limit {
        Some(0) => i64::MAX,
        Some(limit) => limit,
        None => 10,
    };
    let (mut docs, mut diagnostics) = parse_docs(dir)?;
    let mut pass = true;
    if !diagnostics.is_empty() {
        let rest = diagnostics.len() as i64 - limit;
        println!("Found {} diagnostics", diagnostics.len());
        diagnostics.truncate(limit as usize);
        for d in diagnostics {
            println!("{}", d);
        }
        if rest > 0 {
            println!("Hiding the remaining {} diagnostics", rest);
        }
        pass = false;
    }
    // Evaluate doc examples
    let executor = CLIExecutor {
        path: flux_cmd_path,
    };

    let tests = docs
        .par_iter_mut()
        .enumerate()
        .map(|(i, d)| (i, example::evaluate_package_examples(d, &executor)));

    let mut test_count = 0;
    consume_sequentially(tests, |results| {
        for result in results {
            test_count += 1;
            match result {
                Ok((name, duration)) => {
                    eprintln!("OK ... {}, took {}ms", name, duration.as_millis())
                }
                Err(e) => {
                    eprintln!("Error {:?}\n", e);
                    pass = false;
                }
            }
        }
    });

    if pass {
        eprintln!("Finished running {} tests", test_count);
        Ok(())
    } else {
        bail!("docs do not pass lint")
    }
}

/// Iterates through `iter` in parallel however each item is passed to `consume` in the same order
/// that they were produced by the iterator as long as the iterator uses `enumerate` to supply
/// indices. Will deadlock if the indicies passed are not a complete 0..N sequence.
fn consume_sequentially<T>(
    iter: impl IndexedParallelIterator<Item = (usize, T)>,
    mut consume: impl FnMut(T) + Send,
) where
    T: Send,
{
    let (sender, receiver) = std::sync::mpsc::sync_channel::<Element<T>>(12);

    struct Element<T>(usize, T);

    impl<T> PartialEq for Element<T> {
        fn eq(&self, other: &Self) -> bool {
            self.0 == other.0
        }
    }

    impl<T> Eq for Element<T> {}

    impl<T> PartialOrd for Element<T> {
        fn partial_cmp(&self, other: &Self) -> Option<std::cmp::Ordering> {
            Some(self.cmp(other))
        }
    }

    impl<T> Ord for Element<T> {
        fn cmp(&self, other: &Self) -> std::cmp::Ordering {
            other.0.cmp(&self.0)
        }
    }

    let mut heap = BinaryHeap::new();
    let mut current = 0;

    rayon::join(
        move || {
            iter.for_each(|(i, t)| {
                let _ = sender.send(Element(i, t));
            });
        },
        move || {
            for t in receiver {
                heap.push(t);
                while heap.peek().map(|e| e.0) == Some(current) {
                    current += 1;
                    let Element(_, t) = heap.pop().unwrap();
                    consume(t);
                }
            }
        },
    );
}

/// Parse documentation for the specified directory.
fn parse_docs(dir: &Path) -> Result<(Vec<doc::PackageDoc>, doc::Diagnostics)> {
    let db = DatabaseBuilder::default()
        .filesystem_roots(vec![dir.into()])
        .build();

    let mut package_names = bootstrap::parse_dir(dir)?;
    package_names.sort();
    let mut docs = Vec::with_capacity(package_names.len());
    let mut diagnostics = Vec::new();
    for pkgpath in package_names {
        let ast_pkg = db.ast_package(pkgpath.clone()).map_err(|err| {
            let mut errors = db.package_errors();
            errors.push(err);
            errors
        })?;
        let (pkgtypes, _) = db.semantic_package(pkgpath.clone()).map_err(|err| {
            let mut errors = db.package_errors();
            errors.push(err.error);
            errors
        })?;

        let (doc, mut diags) = doc::parse_package_doc_comments(&ast_pkg, &pkgpath, &pkgtypes)
            .context(format!("generating docs for \"{}\"", &pkgpath))?;
        diagnostics.append(&mut diags);
        docs.push(doc);
    }
    Ok((docs, diagnostics))
}

struct CLIExecutor<'a> {
    path: &'a Path,
}

impl example::Executor for CLIExecutor<'_> {
    fn execute(&self, code: &str) -> Result<String> {
        let tmpfile = tempfile::NamedTempFile::new()?;
        write!(tmpfile.reopen()?, "{}", code)?;

        let mut cmd = Command::new(self.path);
        cmd.arg("--format")
            .arg("csv")
            .arg(tmpfile.path())
            .arg("--features")
            .arg(r#"{"labelPolymorphism": true}"#);
        log::debug!("Executing {:?}", cmd);
        let output = cmd
            .output()
            .with_context(|| format!("Unable to execute `{}`", self.path.display()))?;

        if output.status.success() {
            Ok(String::from_utf8(output.stdout)?)
        } else {
            let stderr = String::from_utf8(output.stderr)?;
            // Find error in output
            for line in stderr.lines() {
                if let Some(msg) = line.strip_prefix("Error: ") {
                    bail!("{}", msg)
                }
            }
            // we didn't find a specific error message, report the entire stderr
            bail!("stderr: {}", stderr)
        }
    }
}

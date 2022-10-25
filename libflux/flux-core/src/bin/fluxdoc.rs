use std::{
    collections::BinaryHeap,
    fs::File,
    io::{self, Write},
    path::{Path, PathBuf},
    process::Command,
};

use anyhow::{bail, Context, Result};
use rayon::prelude::*;
use structopt::StructOpt;

use fluxcore::{
    doc::{self, example},
    semantic::bootstrap,
    DatabaseBuilder, Flux, FluxBase,
};

#[derive(Debug, StructOpt)]
#[structopt(about = "generate and validate Flux source code documentation")]
enum FluxDoc {
    /// Dump JSON encoding of documentation from Flux source code.
    Dump {
        /// Directory containing Flux source code.
        #[structopt(short, long, parse(from_os_str))]
        stdlib_dir: Option<PathBuf>,
        /// Path to flux command, must be the cmd from internal/cmd/flux
        #[structopt(long, parse(from_os_str))]
        flux_cmd_path: Option<PathBuf>,
        /// Directory containing Flux source code.
        #[structopt(short, long, parse(from_os_str))]
        dir: PathBuf,
        /// Output file, stdout if not present.
        #[structopt(short, long, parse(from_os_str))]
        output: Option<PathBuf>,
        /// Whether to structure the documentation as nested pacakges.
        #[structopt(short, long)]
        nested: bool,
        /// Whether to omit full descriptions and keep only the short form docs.
        #[structopt(long)]
        short: bool,
    },
    /// Check Flux source code for documentation linting errors
    Lint {
        /// Directory containing Flux source code.
        #[structopt(short, long, parse(from_os_str))]
        stdlib_dir: Option<PathBuf>,
        /// Path to flux command, must be the cmd from internal/cmd/flux
        #[structopt(long, parse(from_os_str))]
        flux_cmd_path: Option<PathBuf>,
        /// Directory containing Flux source code.
        #[structopt(short, long, parse(from_os_str))]
        dir: PathBuf,
        /// Limit the number of diagnostics to report. Default 10. 0 means no limit.
        #[structopt(short, long)]
        limit: Option<i64>,
    },
}

fn main() -> Result<()> {
    env_logger::init();

    let app = FluxDoc::from_args();
    match app {
        FluxDoc::Dump {
            stdlib_dir,
            flux_cmd_path,
            dir,
            output,
            nested,
            short,
        } => dump(
            stdlib_dir.as_deref(),
            flux_cmd_path.as_deref(),
            &dir,
            output.as_deref(),
            nested,
            short,
        )?,
        FluxDoc::Lint {
            stdlib_dir,
            flux_cmd_path,
            dir,
            limit,
        } => lint(stdlib_dir.as_deref(), flux_cmd_path.as_deref(), &dir, limit)?,
    };
    Ok(())
}

const DEFAULT_STDLIB_PATH: &str = "./stdlib-compiled";
const DEFAULT_FLUX_CMD_PATH: &str = "flux";

fn resolve_default_paths<'a>(
    stdlib_dir: Option<&'a Path>,
    flux_cmd_path: Option<&'a Path>,
) -> (&'a Path, &'a Path) {
    let stdlib_dir = match stdlib_dir {
        Some(stdlib_dir) => stdlib_dir,
        None => Path::new(DEFAULT_STDLIB_PATH),
    };
    let flux_cmd_path = match flux_cmd_path {
        Some(flux_cmd_path) => flux_cmd_path,
        None => Path::new(DEFAULT_FLUX_CMD_PATH),
    };
    (stdlib_dir, flux_cmd_path)
}

fn dump(
    stdlib_dir: Option<&Path>,
    flux_cmd_path: Option<&Path>,
    dir: &Path,
    output: Option<&Path>,
    nested: bool,
    short: bool,
) -> Result<()> {
    let (stdlib_dir, flux_cmd_path) = resolve_default_paths(stdlib_dir, flux_cmd_path);
    let f = match output {
        Some(p) => Box::new(File::create(p).context(format!("creating output file {:?}", p))?)
            as Box<dyn io::Write>,
        None => Box::new(io::stdout()),
    };

    let (mut docs, diagnostics) = parse_docs(stdlib_dir, dir).context("parsing source code")?;
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

fn lint(
    stdlib_dir: Option<&Path>,
    flux_cmd_path: Option<&Path>,
    dir: &Path,
    limit: Option<i64>,
) -> Result<()> {
    let (stdlib_dir, flux_cmd_path) = resolve_default_paths(stdlib_dir, flux_cmd_path);
    let limit = match limit {
        Some(limit) if limit == 0 => i64::MAX,
        Some(limit) => limit,
        None => 10,
    };
    let (mut docs, mut diagnostics) = parse_docs(stdlib_dir, dir)?;
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
                Ok(name) => eprintln!("OK ... {}", name),
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
            other.0.partial_cmp(&self.0)
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
fn parse_docs(stdlib_dir: &Path, dir: &Path) -> Result<(Vec<doc::PackageDoc>, doc::Diagnostics)> {
    let db = DatabaseBuilder::default()
        // We resolve paths in stdlib_dir first, then `dir` which mimicks the previous behavior
        // most closely
        .filesystem_roots(vec![stdlib_dir.into(), dir.into()])
        .build();

    let package_names = bootstrap::parse_dir(dir)?;
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

impl<'a> example::Executor for CLIExecutor<'a> {
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

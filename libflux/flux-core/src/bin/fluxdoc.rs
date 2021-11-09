use std::fs::File;
use std::io::{self, Write};
use std::path::{Path, PathBuf};
use std::process::Command;

use anyhow::{bail, Context, Result};
use structopt::StructOpt;

use fluxcore::{
    doc,
    doc::example,
    semantic::{bootstrap, Analyzer},
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
        /// Honor the exception list.
        #[structopt(long)]
        allow_exceptions: bool,
    },
    /// Check Flux source code for documentation linting errors
    Lint {
        /// Directory containing Flux source code.
        #[structopt(short, long, parse(from_os_str))]
        stdlib_dir: Option<PathBuf>,
        /// Directory containing Flux source code.
        #[structopt(short, long, parse(from_os_str))]
        dir: PathBuf,
        /// Limit the number of diagnostics to report. Default 10. 0 means no limit.
        #[structopt(short, long)]
        limit: Option<i64>,
        /// Honor the exception list.
        #[structopt(long)]
        allow_exceptions: bool,
    },
}

fn main() -> Result<()> {
    let app = FluxDoc::from_args();
    match app {
        FluxDoc::Dump {
            stdlib_dir,
            dir,
            output,
            nested,
            short,
            allow_exceptions,
            flux_cmd_path,
        } => dump(
            stdlib_dir.as_deref(),
            flux_cmd_path.as_deref(),
            &dir,
            output.as_deref(),
            nested,
            short,
            allow_exceptions,
        )?,
        FluxDoc::Lint {
            stdlib_dir,
            dir,
            limit,
            allow_exceptions,
        } => lint(stdlib_dir.as_deref(), &dir, limit, allow_exceptions)?,
    };
    Ok(())
}

const DEFAULT_STDLIB_PATH: &str = "./stdlib-compiled";
const DEFAULT_FLUX_CMD_PATH: &str = "flux";

fn dump(
    stdlib_dir: Option<&Path>,
    flux_cmd_path: Option<&Path>,
    dir: &Path,
    output: Option<&Path>,
    nested: bool,
    short: bool,
    allow_exceptions: bool,
) -> Result<()> {
    let stdlib_dir = match stdlib_dir {
        Some(stdlib_dir) => stdlib_dir,
        None => Path::new(DEFAULT_STDLIB_PATH),
    };
    let flux_cmd_path = match flux_cmd_path {
        Some(flux_cmd_path) => flux_cmd_path,
        None => Path::new(DEFAULT_FLUX_CMD_PATH),
    };
    let f = match output {
        Some(p) => Box::new(File::create(p).context(format!("creating output file {:?}", p))?)
            as Box<dyn io::Write>,
        None => Box::new(io::stdout()),
    };

    let exceptions = if allow_exceptions {
        &EXCEPTIONS[..]
    } else {
        &[]
    };
    let (mut docs, diagnostics) =
        parse_docs(stdlib_dir, dir, exceptions).context("parsing source code")?;
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
        let mut executor = CLIExecutor {
            path: flux_cmd_path,
        };
        for d in docs.iter_mut() {
            example::evaluate_package_examples(d, &mut executor)?;
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
    dir: &Path,
    limit: Option<i64>,
    allow_exceptions: bool,
) -> Result<()> {
    let stdlib_dir = match stdlib_dir {
        Some(stdlib_dir) => stdlib_dir,
        None => Path::new(DEFAULT_STDLIB_PATH),
    };
    let limit = match limit {
        Some(limit) if limit == 0 => i64::MAX,
        Some(limit) => limit,
        None => 10,
    };
    let exceptions = if allow_exceptions {
        &EXCEPTIONS[..]
    } else {
        &[]
    };
    let (_, mut diagnostics) = parse_docs(stdlib_dir, dir, exceptions)?;
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
        bail!("docs do not pass lint");
    }
    Ok(())
}

/// Parse documentation for the specified directory.
fn parse_docs(
    stdlib_dir: &Path,
    dir: &Path,
    //TODO(nathanielc): Remove exceptions once the EXCEPTIONS list is empty
    exceptions: &[&str],
) -> Result<(Vec<doc::PackageDoc>, doc::Diagnostics)> {
    let (prelude, stdlib_importer) = bootstrap::stdlib(stdlib_dir)?;
    let mut analyzer = Analyzer::new_with_defaults(prelude, stdlib_importer);
    let ast_packages = bootstrap::parse_dir(dir)?;
    let mut docs = Vec::with_capacity(ast_packages.len());
    let mut diagnostics = Vec::new();
    for (pkgpath, ast_pkg) in ast_packages {
        let (pkgtypes, _) = analyzer.analyze_ast(ast_pkg.clone())?;
        let (doc, mut diags) = doc::parse_package_doc_comments(&ast_pkg, &pkgpath, &pkgtypes)
            .context(format!("generating docs for \"{}\"", &pkgpath))?;
        if !exceptions.contains(&pkgpath.as_str()) {
            diagnostics.append(&mut diags);
        }
        docs.push(doc);
    }
    Ok((docs, diagnostics))
}

struct CLIExecutor<'a> {
    path: &'a Path,
}

impl<'a> example::Executor for CLIExecutor<'a> {
    fn execute(&mut self, code: &str) -> Result<String> {
        let tmpfile = tempfile::NamedTempFile::new()?;
        write!(tmpfile.reopen()?, "{}", code)?;
        let output = Command::new(self.path)
            .arg("--format")
            .arg("csv")
            .arg(tmpfile.path())
            .output()?;

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

// HACK: Any package in this list will not report any errors from documentation diagnostics
// The intent is that as we get each package passing the documentation diagnostics check
// that we remove them from the list and do not add any packages to this list.
// This way we can incrementally improve the documentation while also ensuring we are keeping a
// high standard going forward.
//
// See https://github.com/influxdata/flux/issues/4141 for tacking removing of this list.
const EXCEPTIONS: &[&str] = &[
    "contrib",
    "contrib/RohanSreerama5",
    "contrib/RohanSreerama5/images",
    "contrib/RohanSreerama5/naiveBayesClassifier",
    "contrib/anaisdg",
    "contrib/anaisdg/anomalydetection",
    "contrib/anaisdg/statsmodels",
    "contrib/bonitoo-io",
    "contrib/bonitoo-io/alerta",
    "contrib/bonitoo-io/hex",
    "contrib/bonitoo-io/servicenow",
    "contrib/bonitoo-io/tickscript",
    "contrib/bonitoo-io/zenoss",
    "contrib/chobbs",
    "contrib/chobbs/discord",
    "contrib/jsternberg",
    "contrib/jsternberg/aggregate",
    "contrib/jsternberg/influxdb",
    "contrib/jsternberg/math",
    "contrib/jsternberg/rows",
    "contrib/rhajek",
    "contrib/rhajek/bigpanda",
    "contrib/sranka",
    "contrib/sranka/opsgenie",
    "contrib/sranka/sensu",
    "contrib/sranka/teams",
    "contrib/sranka/telegram",
    "contrib/sranka/webexteams",
    "contrib/tomhollingworth",
    "contrib/tomhollingworth/events",
    "experimental",
    "experimental/aggregate",
    "experimental/array",
    "experimental/bigtable",
    "experimental/csv",
    "experimental/geo",
    "experimental/http",
    "experimental/influxdb",
    "experimental/json",
    "experimental/mqtt",
    "experimental/oee",
    "experimental/prometheus",
    "experimental/query",
    "experimental/record",
    "experimental/table",
    "experimental/usage",
    "influxdata",
    "influxdata/influxdb",
    "influxdata/influxdb/internal",
    "influxdata/influxdb/internal/testutil",
    "influxdata/influxdb/monitor",
    "influxdata/influxdb/sample",
    "influxdata/influxdb/schema",
    "influxdata/influxdb/secrets",
    "influxdata/influxdb/tasks",
    "influxdata/influxdb/v1",
    "internal",
    "internal/boolean",
    "internal/debug",
    "internal/gen",
    "internal/influxql",
    "internal/promql",
    "internal/testutil",
    "interpolate",
    "json",
    "kafka",
    "math",
    "pagerduty",
    "planner",
    "profiler",
    "pushbullet",
    "regexp",
    "runtime",
    "sampledata",
    "slack",
    "socket",
    "sql",
    "strings",
    "system",
    "testing",
    "testing/chronograf",
    "testing/expect",
    "testing/influxql",
    "testing/kapacitor",
    "testing/pandas",
    "testing/prometheus",
    "testing/promql",
    "testing/usage",
    "timezone",
    "universe",
    "universe/holt_winters",
];

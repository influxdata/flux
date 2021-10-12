use std::fs::File;
use std::io;
use std::path::{Path, PathBuf};

use anyhow::{bail, Context, Result};
use structopt::StructOpt;

use fluxcore::semantic::{bootstrap, doc, Analyzer};

#[derive(Debug, StructOpt)]
#[structopt(about = "generate and validate Flux source code documentation")]
enum FluxDoc {
    /// Dump JSON encoding of documentation from Flux source code.
    Dump {
        /// Directory containing Flux source code.
        #[structopt(short, long, parse(from_os_str))]
        stdlib_dir: Option<PathBuf>,
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
        /// Directory containing Flux source code.
        #[structopt(short, long, parse(from_os_str))]
        dir: PathBuf,
        /// Limit the number of diagnostics to report. Default 10.
        #[structopt(short, long)]
        limit: Option<i32>,
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
        } => dump(
            stdlib_dir.as_deref(),
            &dir,
            output.as_deref(),
            nested,
            short,
        )?,
        FluxDoc::Lint {
            stdlib_dir,
            dir,
            limit,
        } => lint(stdlib_dir.as_deref(), &dir, limit)?,
    };
    Ok(())
}

const DEFAULT_STDLIB_PATH: &str = "./stdlib-compiled";

fn dump(
    stdlib_dir: Option<&Path>,
    dir: &Path,
    output: Option<&Path>,
    nested: bool,
    short: bool,
) -> Result<()> {
    let stdlib_dir = match stdlib_dir {
        Some(stdlib_dir) => stdlib_dir,
        None => Path::new(DEFAULT_STDLIB_PATH),
    };
    let f = match output {
        Some(p) => Box::new(File::create(p).context(format!("creating output file {:?}", p))?)
            as Box<dyn io::Write>,
        None => Box::new(io::stdout()),
    };

    let (mut docs, diagnostics) =
        parse_docs(stdlib_dir, dir, &EXCEPTIONS[..]).context("parsing source code")?;
    if !diagnostics.is_empty() {
        bail!(
            "found {} diagnostics when building documentation: {}",
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
    }
    if nested {
        let nested_docs = doc::nest_docs(docs);
        serde_json::to_writer(f, &nested_docs).context("encoding nested json")?;
    } else {
        serde_json::to_writer(f, &docs).context("encoding json")?;
    }

    Ok(())
}

fn lint(stdlib_dir: Option<&Path>, dir: &Path, limit: Option<i32>) -> Result<()> {
    let stdlib_dir = match stdlib_dir {
        Some(stdlib_dir) => stdlib_dir,
        None => Path::new(DEFAULT_STDLIB_PATH),
    };
    let limit = match limit {
        Some(limit) => limit as usize,
        None => 10,
    };
    let (_, mut diagnostics) = parse_docs(stdlib_dir, dir, &[])?;
    if !diagnostics.is_empty() {
        let rest = diagnostics.len() as i64 - limit as i64;
        println!("Found {} diagnostics", diagnostics.len());
        diagnostics.truncate(limit);
        for d in diagnostics {
            println!("{}", d);
        }
        if rest > 0 {
            println!("Hiding the remaining {} diagnostics", rest);
        }
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
    let mut analyzer = Analyzer::new(prelude, stdlib_importer);
    let ast_packages = bootstrap::parse_dir(dir)?;
    let mut docs = Vec::with_capacity(ast_packages.len());
    let mut diagnostics = Vec::new();
    for (pkgpath, ast_pkg) in ast_packages {
        let (pkgtypes, _) = analyzer.analyze_ast(ast_pkg.clone())?;
        let (doc, mut diags) = doc::parse_package_doc_comments(&ast_pkg, &pkgpath, &pkgtypes)?;
        if !exceptions.contains(&pkgpath.as_str()) {
            diagnostics.append(&mut diags);
        }
        docs.push(doc);
    }
    Ok((docs, diagnostics))
}

// HACK: Any package in this list will not report any errors from documentation diagnostics
// The intent is that as we get each package passing the documentation diagnostics check
// that we remove them from the list and do not add any packages to this list.
// This way we can incrementally improve the documentation while also ensuring we are keeping a
// high standard going forward.
//
// See https://github.com/influxdata/flux/issues/4141 for tacking removing of this list.
const EXCEPTIONS: [&str; 96] = [
    "array",
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
    "contrib/bonitoo-io/tickscript",
    "contrib/bonitoo-io/victorops",
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
    "csv",
    "date",
    "dict",
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
    "generate",
    "http",
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
    "universe",
    "universe/holt_winters",
];

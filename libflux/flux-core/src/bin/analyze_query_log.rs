use std::{io, path::PathBuf};

use anyhow::{anyhow, Context as _, Error, Result};
use rayon::prelude::*;
use serde::Deserialize;
use structopt::StructOpt;

use fluxcore::semantic::{self, import::Packages, Analyzer};

#[derive(Debug, StructOpt)]
#[structopt(about = "analyze a query log database")]
struct AnalyzeQueryLog {
    #[structopt(long, help = "How many sources to skip")]
    skip: Option<usize>,
    #[structopt(
        long,
        min_values = 1,
        use_delimiter = true,
        help = "Which new features to compare against"
    )]
    new_features: Vec<semantic::Feature>,
    #[structopt(
        long,
        help = "Report differences when the script fails to compile under both sets of features"
    )]
    report_already_failing_scripts: bool,
    database: PathBuf,

    #[structopt(long, help = "Prints the source code for each script in the input")]
    print_sources: bool,
}

trait Source {
    fn read(&mut self) -> Result<Box<dyn Iterator<Item = Result<String>> + '_>>;
}

impl Source for rusqlite::Statement<'_> {
    fn read(&mut self) -> Result<Box<dyn Iterator<Item = Result<String>> + '_>> {
        Ok(Box::new(
            self.query_map([], |row| row.get(0))?
                .map(|e| e.map_err(Error::from)),
        ))
    }
}

impl<R> Source for csv::Reader<R>
where
    R: io::Read + Send,
{
    fn read(&mut self) -> Result<Box<dyn Iterator<Item = Result<String>> + '_>> {
        let headers = self.headers()?;
        let i = headers
            .iter()
            .position(|field| field == "_value")
            .ok_or_else(|| anyhow!("Missing _value field"))?;

        #[derive(Deserialize, Debug)]
        struct QueryLogValue {
            request: QueryLogRequest,
        }
        #[derive(Deserialize, Debug)]
        struct QueryLogRequest {
            compiler: QueryLogCompiler,
        }
        #[derive(Deserialize, Debug)]
        struct QueryLogCompiler {
            query: String,
        }

        Ok(Box::new(self.records().map(move |record| {
            let record = record?;
            let s = record.get(i).unwrap();
            let value = serde_json::from_str::<QueryLogValue>(s)?;
            Ok::<_, Error>(value.request.compiler.query)
        })))
    }
}

fn main() -> Result<()> {
    env_logger::init();

    let app = AnalyzeQueryLog::from_args();

    let new_config = semantic::AnalyzerConfig {
        features: app.new_features,
    };

    let stdlib_path = PathBuf::from("../stdlib");

    let (prelude, imports, _sem_pkgs) =
        semantic::bootstrap::infer_stdlib_dir(&stdlib_path, semantic::AnalyzerConfig::default())?;

    let (new_prelude, new_imports, _sem_pkgs) =
        semantic::bootstrap::infer_stdlib_dir(&stdlib_path, new_config.clone())?;

    let analysis: Box<dyn Analysis> = if app.print_sources {
        Box::new(PrintSources {})
    } else {
        Box::new(FeatureDiff {
            prelude,
            imports,
            new_prelude,
            new_imports,
            new_config,
            report_already_failing_scripts: app.report_already_failing_scripts,
        })
    };

    let mut connection;

    let sources: Box<dyn FnOnce() -> Box<dyn Source> + Send> =
        match app.database.extension().and_then(|e| e.to_str()) {
            Some("flux") => {
                let source = std::fs::read_to_string(&app.database)?;

                analysis.analyze(0, &source)?;

                return Ok(());
            }
            Some("csv") => {
                let input = std::fs::read_to_string(&app.database)
                    .with_context(|| format!("`{}` could not be read", app.database.display()))?;

                // The flux csv format has extra headers which we remove before parsing the csv
                let mut first = true;
                let input = input
                    .lines()
                    .filter(|line| {
                        if line.starts_with(",result") {
                            if first {
                                first = false;
                                true
                            } else {
                                false
                            }
                        } else {
                            true
                        }
                    })
                    .collect::<Vec<_>>()
                    .join("\r\n");

                let reader = csv::ReaderBuilder::new()
                    .comment(Some(b'#'))
                    .flexible(true)
                    .from_reader(std::io::Cursor::new(input.into_bytes()));

                Box::new(move || Box::new(reader))
            }
            _ => {
                connection = rusqlite::Connection::open(&app.database)?;
                let connection = &mut connection;
                Box::new(move || {
                    Box::new(
                        connection
                            .prepare("SELECT source FROM query limit 100000")
                            .unwrap(),
                    )
                })
            }
        };

    let (tx, rx) = crossbeam_channel::bounded(128);

    let (final_tx, final_rx) = crossbeam_channel::bounded(128);

    let mut count = 0;

    let (r, r2, ()) = join3(
        move || {
            for (i, result) in sources().read()?.enumerate() {
                if let Some(skip) = app.skip {
                    if i < skip {
                        continue;
                    }
                }

                let source: String = result?;
                tx.send((i, source))?;
            }

            Ok::<_, Error>(())
        },
        move || {
            rx.into_iter()
                .par_bridge()
                .try_for_each(|(i, source): (usize, String)| {
                    // eprintln!("{}", source);

                    analysis.analyze(i, &source)?;

                    final_tx.send(())?;

                    Ok::<_, Error>(())
                })
        },
        || {
            for _ in final_rx {
                count += 1;

                if count % 100 == 0 {
                    eprintln!("Checked {} queries", count);
                }
            }
        },
    );

    r?;
    r2?;

    eprintln!("Done! Checked {} queries", count);

    Ok(())
}

trait Analysis: Send + Sync {
    fn analyze(&self, i: usize, source: &str) -> Result<()>;
}

struct FeatureDiff {
    prelude: semantic::PackageExports,
    imports: Packages,
    new_prelude: semantic::PackageExports,
    new_imports: Packages,
    new_config: semantic::AnalyzerConfig,
    report_already_failing_scripts: bool,
}

impl Analysis for FeatureDiff {
    fn analyze(&self, i: usize, source: &str) -> Result<()> {
        let analyzer = || {
            Analyzer::new(
                (&self.prelude).into(),
                &self.imports,
                semantic::AnalyzerConfig::default(),
            )
        };

        let current_result = match std::panic::catch_unwind(|| {
            analyzer().analyze_source("".into(), "".into(), source)
        }) {
            Ok(x) => x,
            Err(_) => panic!("Panic at source {}: {}", i, source),
        };

        let new_analyzer = || {
            Analyzer::new(
                (&self.new_prelude).into(),
                &self.new_imports,
                self.new_config.clone(),
            )
        };
        let new_result = match std::panic::catch_unwind(|| {
            new_analyzer().analyze_source("".into(), "".into(), source)
        }) {
            Ok(x) => x,
            Err(_) => panic!("Panic at source {}: {}", i, source),
        };

        match (current_result, new_result) {
            (Ok(_), Ok(_)) => (),
            (Err(err), Ok(_)) => {
                eprintln!("### {}", i);
                eprintln!("{}", source);

                eprintln!(
                    "Missing errors when the features are enabled: {}",
                    err.error.pretty(source)
                );
                eprintln!("-------------------------------");
            }
            (Ok(_), Err(err)) => {
                eprintln!("### {}", i);
                eprintln!("{}", source);

                eprintln!(
                    "New errors when the features are enabled: {}",
                    err.error.pretty(source)
                );
                eprintln!("-------------------------------");
            }
            (Err(current_err), Err(new_err)) => {
                if self.report_already_failing_scripts {
                    let current_err = current_err.error.pretty(source);
                    let new_err = new_err.error.pretty(source);
                    if current_err != new_err {
                        eprintln!("{}", source);

                        eprintln!(
                            "Different when the new features are enabled:\n{}",
                            pretty_assertions::StrComparison::new(&current_err, &new_err,)
                        );
                        eprintln!("-------------------------------");
                    }
                }
            }
        }

        Ok(())
    }
}

struct PrintSources {}

impl Analysis for PrintSources {
    fn analyze(&self, i: usize, source: &str) -> Result<()> {
        eprintln!("### {}", i);
        eprintln!("{}", source);

        let mut s = String::new();
        std::io::stdin().read_line(&mut s)?;

        Ok(())
    }
}

fn join3<A, B, C>(
    a: impl FnOnce() -> A + Send,
    b: impl FnOnce() -> B + Send,
    c: impl FnOnce() -> C + Send,
) -> (A, B, C)
where
    A: Send,
    B: Send,
    C: Send,
{
    let (a, (b, c)) = rayon::join(a, || rayon::join(b, c));
    (a, b, c)
}

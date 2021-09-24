extern crate fluxcore;

use std::{env, fs, io, io::Write, path};

use deflate::deflate_bytes;
use fluxcore::semantic::bootstrap;
use fluxcore::semantic::bootstrap::stdlib_docs;
use fluxcore::semantic::env::Environment;
use fluxcore::semantic::flatbuffers::types as fb;
use fluxcore::semantic::sub::Substitutable;

#[derive(Debug)]
struct Error {
    msg: String,
}

impl From<env::VarError> for Error {
    fn from(err: env::VarError) -> Error {
        Error {
            msg: err.to_string(),
        }
    }
}

impl From<io::Error> for Error {
    fn from(err: io::Error) -> Error {
        Error {
            msg: format!("{:?}", err),
        }
    }
}

impl From<bootstrap::Error> for Error {
    fn from(err: bootstrap::Error) -> Error {
        Error { msg: err.msg }
    }
}

fn serialize<'a, T, S, F>(ty: T, f: F, path: &path::Path) -> Result<(), Error>
where
    F: Fn(&mut flatbuffers::FlatBufferBuilder<'a>, T) -> flatbuffers::WIPOffset<S>,
{
    let mut builder = flatbuffers::FlatBufferBuilder::new();
    let buf = fb::serialize(&mut builder, ty, f);
    let mut file = fs::File::create(path)?;
    file.write_all(buf)?;
    Ok(())
}

fn main() -> Result<(), Error> {
    let dir = path::PathBuf::from(env::var("OUT_DIR")?);

    let std_lib_values = bootstrap::infer_stdlib()?;
    let (pre, lib, libmap, files, file_map) = (
        std_lib_values.prelude,
        std_lib_values.importer,
        std_lib_values.importermap,
        std_lib_values.rerun_if_changed,
        std_lib_values.files,
    );
    for f in files.iter() {
        println!("cargo:rerun-if-changed={}", f);
    }

    // Validate there aren't any free type variables in the environment
    for (name, ty) in &pre {
        if !ty.free_vars().is_empty() {
            return Err(Error {
                msg: format!("found free variables in type of {}: {}", name, ty),
            });
        }
    }
    for (name, ty) in &lib {
        if !ty.free_vars().is_empty() {
            return Err(Error {
                msg: format!("found free variables in type of package {}: {}", name, ty),
            });
        }
    }
    // Any package in this list will not report any errors from documentation diagnostics
    // The intent is that as we get each package passing the documentation diagnostics check
    // that we remove them from the list and do not add any packages to this list.
    // This way we can incrementally improve the documentation while also ensuring we are keeping a
    // high standard going forward.
    let exceptions: Vec<&str> = vec![
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
    let new_docs = stdlib_docs(&libmap, &file_map, exceptions).unwrap();
    let json_docs = serde_json::to_vec(&new_docs).unwrap();
    let comp_docs = deflate_bytes(&json_docs);
    let path = dir.join("docs.json");
    let mut file = fs::File::create(path)?;
    file.write_all(&comp_docs)?;

    let path = dir.join("prelude.data");
    serialize(Environment::from(pre), fb::build_env, &path)?;

    let path = dir.join("stdlib.data");
    serialize(Environment::from(lib), fb::build_env, &path)?;

    Ok(())
}

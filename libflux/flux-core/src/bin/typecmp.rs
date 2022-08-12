use std::{cmp::max, fs, path::PathBuf};

use anyhow::{anyhow, Result};
use fluxcore::{
    doc::{Doc, PackageDoc},
    parser,
    semantic::{
        cmp::{self, TypeDiff},
        convert::convert_polytype,
        types::PolyType,
        AnalyzerConfig,
    },
};
use structopt::StructOpt;

#[derive(Debug, StructOpt)]
#[structopt(about = "compare Flux types")]
enum TypeCmp {
    /// Diff an old and new Flux types
    Diff {
        /// Old Flux type
        #[structopt(short, long, parse(from_os_str))]
        old: PathBuf,
        /// New Flux type
        #[structopt(short, long, parse(from_os_str))]
        new: PathBuf,
        /// Versbose
        #[structopt(short, long)]
        verbose: bool,
    },
}
fn main() -> Result<()> {
    let app = TypeCmp::from_args();
    match app {
        TypeCmp::Diff { old, new, verbose } => diff(&old, &new, verbose)?,
    };
    Ok(())
}

struct Diagnostic {
    pkg: String,
    diff: TypeDiff,
    new: Vec<String>,
    diffs: Vec<(String, Diff)>,
}

struct Diff {
    old: PolyType,
    new: PolyType,
    diff: TypeDiff,
}

fn diff(old: &PathBuf, new: &PathBuf, verbose: bool) -> Result<()> {
    let old_contents = fs::read_to_string(old)?;
    let old: Vec<PackageDoc> = serde_json::from_str(&old_contents)?;

    let new_contents = fs::read_to_string(new)?;
    let new: Vec<PackageDoc> = serde_json::from_str(&new_contents)?;

    let mut diff = TypeDiff::Patch;
    let mut diags = Vec::new();
    // Check for breaking changes
    for opkg in &old {
        let mut diag = Diagnostic {
            pkg: opkg.path.clone(),
            diff: TypeDiff::Patch,
            new: Vec::new(),
            diffs: Vec::new(),
        };
        if let Some(npkg) = new.iter().find(|p| p.path == opkg.path) {
            for (name, omember) in &opkg.members {
                if let Some((_, nmember)) = &npkg.members.iter().find(|(k, _)| *k == name) {
                    let old_type = type_of_doc(omember)?;
                    let new_type = type_of_doc(*nmember)?;
                    let d = cmp::diff(&old_type, &new_type);
                    if verbose && d != TypeDiff::Patch {
                        diag.diffs.push((
                            name.to_owned(),
                            Diff {
                                old: old_type.clone(),
                                new: new_type.clone(),
                                diff: d,
                            },
                        ));
                    }
                    diag.diff = max(diag.diff, d);
                }
            }
            for (name, _) in &npkg.members {
                if !opkg.members.contains_key(name) {
                    diag.diff = max(diag.diff, TypeDiff::Minor);
                    diag.new.push(name.to_owned());
                }
            }
        }
        diff = max(diff, diag.diff);
        diags.push(diag);
    }
    let mut new_pkgs = Vec::new();
    // Check for new packages
    for npkg in &new {
        if old.iter().find(|p| p.path == npkg.path).is_none() {
            diff = max(diff, TypeDiff::Minor);
            new_pkgs.push(npkg.path.to_owned());
        }
    }
    if verbose {
        for diag in diags {
            if diag.diff != TypeDiff::Patch {
                println!("{} has a {:?} change", diag.pkg, diag.diff);
                for (name, d) in diag.diffs {
                    let old = format!("{}", d.old).replace("\n", "\n\t\t");
                    let new = format!("{}", d.new).replace("\n", "\n\t\t");
                    println!(
                        "\t{}.{} has a {:?} change:\n\t\told: {}\n\t\tvs\n\t\tnew: {}",
                        diag.pkg, name, d.diff, old, new
                    );
                }
                for name in diag.new {
                    println!("\t{} has new member {}", diag.pkg, name);
                }
            }
        }
        for pkg in new_pkgs {
            println!("{} is a new package", pkg);
        }
    }
    println!("{:?}", diff);

    Ok(())
}

fn type_of_doc(doc: &Doc) -> Result<PolyType> {
    match doc {
        Doc::Package(_) => Err(anyhow!("unexpected package")),
        Doc::Value(v) => str_to_type(v.flux_type.as_str()),
        Doc::Function(f) => str_to_type(f.flux_type.as_str()),
    }
}

fn str_to_type(typ: &str) -> Result<PolyType> {
    let type_expr = parser::Parser::new(typ).parse_type_expression();
    Ok(convert_polytype(&type_expr, &AnalyzerConfig::default())?)
}

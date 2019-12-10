use std::collections::{HashMap, HashSet};
use std::fs;
use std::io;
use std::path::PathBuf;

use crate::ast;
use crate::parser;
use crate::semantic::analyze::analyze_file;
use crate::semantic::env::Environment;
use crate::semantic::fresh::Fresher;
use crate::semantic::import::Importer;
use crate::semantic::infer;
use crate::semantic::infer::Constraints;
use crate::semantic::nodes;
use crate::semantic::nodes::infer_file;
use crate::semantic::sub::Substitutable;
use crate::semantic::types;
use crate::semantic::types::{MonoType, PolyType, Property, Row};

use walkdir::WalkDir;

const PRELUDE: [&str; 2] = ["universe", "influxdata/influxdb"];

type Error = String;

impl From<nodes::Error> for Error {
    fn from(err: nodes::Error) -> Error {
        err.to_string()
    }
}

impl From<types::Error> for Error {
    fn from(err: types::Error) -> Error {
        err.to_string()
    }
}

#[allow(dead_code)]
// Recursively parse all flux files within a directory.
fn parse_flux_files(path: &str) -> io::Result<Vec<ast::File>> {
    let mut files = Vec::new();
    let entries = WalkDir::new(PathBuf::from(path))
        .into_iter()
        .filter_map(|r| r.ok())
        .filter(|r| r.path().is_file());

    for entry in entries {
        if let Some(path) = entry.path().to_str() {
            if path.ends_with(".flux") && !path.ends_with("_test.flux") {
                files.push(parser::parse_string(
                    path.rsplitn(2, "/stdlib/").collect::<Vec<&str>>()[0],
                    &fs::read_to_string(path)?,
                ));
            }
        }
    }
    Ok(files)
}

#[allow(dead_code)]
// Associates an import path with each file
fn file_map(files: Vec<ast::File>) -> HashMap<String, ast::File> {
    files.into_iter().fold(HashMap::new(), |mut acc, file| {
        let name = file.name.rsplitn(2, '/').collect::<Vec<&str>>()[1].to_string();
        acc.insert(name, file);
        acc
    })
}

fn imports(file: &ast::File) -> Vec<&str> {
    let mut dependencies = Vec::new();
    for import in &file.imports {
        dependencies.push(&import.path.value[..]);
    }
    dependencies
}

// Determines the dependencies of a package. That is, all packages
// that must be evaluated before the package in question. Each
// dependency is added to the `deps` vector in evaluation order.
#[allow(clippy::type_complexity)]
fn dependencies<'a>(
    name: &'a str,
    pkgs: &'a HashMap<String, ast::File>,
    mut deps: Vec<&'a str>,
    mut seen: HashSet<&'a str>,
    mut done: HashSet<&'a str>,
) -> Result<(Vec<&'a str>, HashSet<&'a str>, HashSet<&'a str>), Error> {
    if seen.contains(name) && !done.contains(name) {
        Err(format!(r#"package "{}" depends on itself"#, name))
    } else {
        seen.insert(name);
        match pkgs.get(name) {
            None => Err(format!(r#"package "{}" not found"#, name)),
            Some(file) => {
                for name in imports(file) {
                    let (x, y, z) = dependencies(name, pkgs, deps, seen, done)?;
                    deps = x;
                    seen = y;
                    done = z;
                    if !deps.contains(&name) {
                        deps.push(name);
                    }
                }
                done.insert(name);
                Ok((deps, seen, done))
            }
        }
    }
}

// Constructs a polytype, or more specifically a generic row type, from a hash map
pub fn build_polytype(from: HashMap<String, PolyType>, f: &mut Fresher) -> Result<PolyType, Error> {
    let (r, cons) = build_row(from, f);
    let mut kinds = HashMap::new();
    let sub = infer::solve(&cons, &mut kinds, f)?;
    Ok(infer::generalize(
        &Environment::empty(),
        &kinds,
        MonoType::Row(Box::new(r)).apply(&sub),
    ))
}

fn build_row(from: HashMap<String, PolyType>, f: &mut Fresher) -> (Row, Constraints) {
    let mut r = Row::Empty;
    let mut cons = Constraints::empty();

    for (name, poly) in from {
        let (ty, constraints) = infer::instantiate(poly.clone(), f);
        r = Row::Extension {
            head: Property { k: name, v: ty },
            tail: MonoType::Row(Box::new(r)),
        };
        cons = cons + constraints;
    }
    (r, cons)
}

#[allow(dead_code)]
#[allow(clippy::type_complexity)]
fn infer_prelude<I: Importer>(
    f: &mut Fresher,
    files: &HashMap<String, ast::File>,
    builtin: &HashMap<&str, I>,
) -> Result<(HashMap<String, PolyType>, HashMap<String, PolyType>), Error> {
    let mut prelude = HashMap::new();
    let mut imports = HashMap::new();
    for name in &PRELUDE {
        let (types, importer) = infer_pkg(name, f, files, builtin, HashMap::new(), imports)?;
        for (k, v) in types {
            prelude.insert(k, v);
        }
        imports = importer;
    }
    Ok((prelude, imports))
}

// Infer the types in a package(file), returning a hash map containing
// the inferred types along with a possibly updated map of package imports.
//
#[allow(clippy::type_complexity)]
fn infer_pkg<I: Importer>(
    name: &str,                         // name of package to infer
    f: &mut Fresher,                    // type variable fresher
    files: &HashMap<String, ast::File>, // files available for inference
    builtin: &HashMap<&str, I>,         // builtin types
    prelude: HashMap<String, PolyType>, // prelude types
    imports: HashMap<String, PolyType>, // types available for import
) -> Result<
    (
        HashMap<String, PolyType>, // inferred types
        HashMap<String, PolyType>, // types available for import (possibly updated)
    ),
    Error,
> {
    // Determine the order in which we must infer dependencies
    let (deps, _, _) = dependencies(name, files, Vec::new(), HashSet::new(), HashSet::new())?;

    let mut imports = imports;

    // Infer all dependencies
    for pkg in deps {
        if imports.import(pkg).is_none() {
            let file = files.get(pkg);
            if file.is_none() {
                return Err(format!(r#"package "{}" not found"#, pkg));
            }
            let file = file.unwrap().to_owned();

            let env = if let Some(builtins) = builtin.get(pkg) {
                infer_file(
                    &mut analyze_file(file, f)?,
                    Environment::new(prelude.clone().into()),
                    f,
                    &imports,
                    builtins,
                )?
                .0
            } else {
                infer_file(
                    &mut analyze_file(file, f)?,
                    Environment::new(prelude.clone().into()),
                    f,
                    &imports,
                    &None,
                )?
                .0
            };

            imports.insert(pkg.to_string(), build_polytype(env.values, f)?);
        }
    }

    let file = files.get(name);
    if file.is_none() {
        return Err(format!("package '{}' not found", name));
    }
    let file = file.unwrap().to_owned();

    let env = if let Some(builtins) = builtin.get(name) {
        infer_file(
            &mut analyze_file(file, f)?,
            Environment::new(prelude.clone().into()),
            f,
            &imports,
            builtins,
        )?
        .0
    } else {
        infer_file(
            &mut analyze_file(file, f)?,
            Environment::new(prelude.clone().into()),
            f,
            &imports,
            &None,
        )?
        .0
    };

    Ok((env.values, imports))
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::parser::parse_string;
    use crate::semantic::env::Environment;
    use crate::semantic::parser::parse;

    #[test]
    fn infer_program() {
        let a = r#"
            f = (x) => x
        "#;
        let b = r#"
            import "a"

            builtin x

            y = a.f(x: x)
        "#;
        let c = r#"
            import "b"

            z = b.y
        "#;
        let files = maplit::hashmap! {
            String::from("a") => parse_string("a.flux", a),
            String::from("b") => parse_string("b.flux", b),
            String::from("c") => parse_string("c.flux", c),
        };
        let builtins = maplit::hashmap! {
            "b" => Environment::from(maplit::hashmap! {
                String::from("x") => parse("forall [] int").unwrap(),
            }),
        };
        let (types, imports) = infer_pkg(
            "c",
            &mut Fresher::from(1),
            &files,
            &builtins,
            HashMap::new(),
            HashMap::new(),
        )
        .unwrap();

        let want = maplit::hashmap! {
            String::from("z") => parse("forall [] int").unwrap(),
        };
        assert_eq!(want, types);

        let want = maplit::hashmap! {
            String::from("a") => parse("forall [t0] {f: (x: t0) -> t0}").unwrap(),
            String::from("b") => parse("forall [] {x: int | y: int}").unwrap(),
        };
        assert_eq!(want, imports);
    }

    #[test]
    fn prelude_dependencies() {
        let files = file_map(parse_flux_files("../../../stdlib").unwrap());

        let r = PRELUDE.iter().try_fold(
            (Vec::new(), HashSet::new(), HashSet::new()),
            |(deps, seen, done), name| dependencies(name, &files, deps, seen, done),
        );

        let names = r.unwrap().0;

        assert_eq!(vec!["system", "date", "math", "strings", "regexp"], names,);
    }

    #[test]
    fn cyclic_dependency() {
        let a = r#"
            import "b"
        "#;
        let b = r#"
            import "a"
        "#;
        let files = maplit::hashmap! {
            String::from("a") => parse_string("a.flux", a),
            String::from("b") => parse_string("b.flux", b),
        };
        assert_eq!(
            Err(r#"package "b" depends on itself"#.to_string()),
            dependencies("b", &files, Vec::new(), HashSet::new(), HashSet::new(),),
        );
    }
}

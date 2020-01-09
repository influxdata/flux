use std::collections::{HashMap, HashSet};
use std::fs;
use std::io;
use std::path::PathBuf;

use crate::ast;
use crate::parser;
use crate::semantic::builtins::builtins;
use crate::semantic::convert::convert_file;
use crate::semantic::env::Environment;
use crate::semantic::fresh::Fresher;
use crate::semantic::import::Importer;
use crate::semantic::infer;
use crate::semantic::infer::Constraints;
use crate::semantic::nodes;
use crate::semantic::nodes::infer_file;
use crate::semantic::parser::parse;
use crate::semantic::sub::Substitutable;
use crate::semantic::types;
use crate::semantic::types::{MaxTvar, MonoType, PolyType, Property, Row, Tvar};

use walkdir::WalkDir;

const PRELUDE: [&str; 2] = ["universe", "influxdata/influxdb"];

#[derive(Debug, PartialEq)]
pub struct Error {
    pub msg: String,
}

impl From<io::Error> for Error {
    fn from(err: io::Error) -> Error {
        Error {
            msg: format!("{:?}", err),
        }
    }
}

impl From<nodes::Error> for Error {
    fn from(err: nodes::Error) -> Error {
        Error {
            msg: err.to_string(),
        }
    }
}

impl From<types::Error> for Error {
    fn from(err: types::Error) -> Error {
        Error {
            msg: err.to_string(),
        }
    }
}

impl From<String> for Error {
    fn from(msg: String) -> Error {
        Error { msg }
    }
}

impl From<&str> for Error {
    fn from(msg: &str) -> Error {
        Error {
            msg: msg.to_string(),
        }
    }
}

#[allow(clippy::type_complexity)]
// Infer the types of the standard library returning two importers, one for the prelude
// and one for the standard library, as well as a type variable fresher.
pub fn infer_stdlib() -> Result<
    (
        HashMap<String, PolyType>,
        HashMap<String, PolyType>,
        Fresher,
    ),
    Error,
> {
    let (builtins, mut f) = builtin_types()?;

    let files = file_map(parse_flux_files("../../stdlib")?);

    let (prelude, importer) = infer_pre(&mut f, &files, &builtins)?;
    let importer = infer_std(&mut f, &files, &builtins, prelude.clone(), importer)?;

    Ok((prelude, importer, f))
}

#[allow(clippy::type_complexity)]
fn builtin_types() -> Result<(HashMap<String, HashMap<String, PolyType>>, Fresher), Error> {
    let mut tv = Tvar(0);
    let mut ty = HashMap::new();
    for (mut path, expr) in builtins().iter() {
        let name = path.pop().unwrap();
        let expr = parse(expr)?;

        let tvar = expr.max_tvar();
        if tvar > tv {
            tv = tvar;
        }

        ty.entry(path.join("/"))
            .or_insert_with(HashMap::new)
            .insert(name.to_string(), expr);
    }
    Ok((ty, Fresher::from(tv.0 + 1)))
}

#[allow(clippy::type_complexity)]
fn infer_pre<I: Importer>(
    f: &mut Fresher,
    files: &HashMap<String, ast::File>,
    builtin: &HashMap<String, I>,
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

#[allow(clippy::type_complexity)]
fn infer_std<I: Importer>(
    f: &mut Fresher,
    files: &HashMap<String, ast::File>,
    builtin: &HashMap<String, I>,
    prelude: HashMap<String, PolyType>,
    mut imports: HashMap<String, PolyType>,
) -> Result<HashMap<String, PolyType>, Error> {
    for (path, _) in files.iter() {
        if imports.contains_key(path) {
            continue;
        }
        let (types, mut importer) = infer_pkg(path, f, files, builtin, prelude.clone(), imports)?;
        importer.insert(path.to_string(), build_polytype(types, f)?);
        imports = importer;
    }
    Ok(imports)
}

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
        Err(Error {
            msg: format!(r#"package "{}" depends on itself"#, name),
        })
    } else {
        seen.insert(name);
        match pkgs.get(name) {
            None => Err(Error {
                msg: format!(r#"package "{}" not found"#, name),
            }),
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
pub fn build_polytype<S: ::std::hash::BuildHasher>(
    from: HashMap<String, PolyType, S>,
    f: &mut Fresher,
) -> Result<PolyType, Error> {
    let (r, cons) = build_row(from, f);
    let mut kinds = HashMap::new();
    let sub = infer::solve(&cons, &mut kinds, f)?;
    Ok(infer::generalize(
        &Environment::empty(),
        &kinds,
        MonoType::Row(Box::new(r)).apply(&sub),
    ))
}

fn build_row<S: ::std::hash::BuildHasher>(
    from: HashMap<String, PolyType, S>,
    f: &mut Fresher,
) -> (Row, Constraints) {
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

// Infer the types in a package(file), returning a hash map containing
// the inferred types along with a possibly updated map of package imports.
//
#[allow(clippy::type_complexity)]
fn infer_pkg<I: Importer>(
    name: &str,                         // name of package to infer
    f: &mut Fresher,                    // type variable fresher
    files: &HashMap<String, ast::File>, // files available for inference
    builtin: &HashMap<String, I>,       // builtin types
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
                return Err(Error {
                    msg: format!(r#"package "{}" not found"#, pkg),
                });
            }
            let file = file.unwrap().to_owned();

            let env = if let Some(builtins) = builtin.get(pkg) {
                infer_file(
                    &mut convert_file(file, f)?,
                    Environment::new(prelude.clone().into()),
                    f,
                    &imports,
                    builtins,
                )?
                .0
            } else {
                infer_file(
                    &mut convert_file(file, f)?,
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
        return Err(Error {
            msg: format!("package '{}' not found", name),
        });
    }
    let file = file.unwrap().to_owned();

    let env = if let Some(builtins) = builtin.get(name) {
        infer_file(
            &mut convert_file(file, f)?,
            Environment::new(prelude.into()),
            f,
            &imports,
            builtins,
        )?
        .0
    } else {
        infer_file(
            &mut convert_file(file, f)?,
            Environment::new(prelude.into()),
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
    fn infer_program() -> Result<(), Error> {
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
            String::from("b") => Environment::from(maplit::hashmap! {
                String::from("x") => parse("forall [] int")?,
            }),
        };
        let (types, imports) = infer_pkg(
            "c",
            &mut Fresher::from(1),
            &files,
            &builtins,
            HashMap::new(),
            HashMap::new(),
        )?;

        let want = maplit::hashmap! {
            String::from("z") => parse("forall [] int")?,
        };
        if want != types {
            return Err(Error {
                msg: format!(
                    "unexpected inference result:\n\nwant: {:?}\n\ngot: {:?}",
                    want, types
                ),
            });
        }

        let want = maplit::hashmap! {
            String::from("a") => parse("forall [t0] {f: (x: t0) -> t0}")?,
            String::from("b") => parse("forall [] {x: int | y: int}")?,
        };
        if want != imports {
            return Err(Error {
                msg: format!(
                    "unexpected type importer:\n\nwant: {:?}\n\ngot: {:?}",
                    want, types
                ),
            });
        }

        Ok(())
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
        let got_err = dependencies("b", &files, Vec::new(), HashSet::new(), HashSet::new())
            .expect_err("expected cyclic dependency error");

        assert_eq!(
            Error {
                msg: r#"package "b" depends on itself"#.to_string()
            },
            got_err
        );
    }
}
